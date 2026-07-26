package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/talentodh/hipodromo/internal/api"
	"github.com/talentodh/hipodromo/internal/auth"
	"github.com/talentodh/hipodromo/internal/contract"
	"github.com/talentodh/hipodromo/internal/seed"
	"github.com/talentodh/hipodromo/internal/store"
	"github.com/talentodh/hipodromo/internal/ws"
)

// ── Andamiaje ─────────────────────────────────────────────────────────────

// El reloj de los tests está clavado en el ancla del dataset. Así el estado es
// el conocido: 4 carreras terminadas, race_005 en vivo, 3 por venir.
var testNow = seed.Anchor

type harness struct {
	t       *testing.T
	server  *api.Server
	handler http.Handler
	store   *store.Store
}

// silentMailer registra los envíos sin imprimir nada.
type silentMailer struct{ sent []string }

func (m *silentMailer) SendVerification(_ context.Context, to, _, token string) error {
	m.sent = append(m.sent, to+"|"+token)
	return nil
}

func newHarness(t *testing.T) (*harness, *silentMailer) {
	t.Helper()

	data, err := seed.Load(testNow)
	if err != nil {
		t.Fatalf("cargando el dataset: %v", err)
	}
	// Sin SnapshotPath: cada test arranca limpio y no toca el disco.
	st, err := store.New(data, store.Options{})
	if err != nil {
		t.Fatalf("creando el store: %v", err)
	}

	mailer := &silentMailer{}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	signer := auth.NewSigner("test-secret")

	server := &api.Server{
		Store: st, Signer: signer, Hub: ws.NewHub(log, signer), Mail: mailer, Log: log,
		FrontURL: "http://localhost:4200",
		Clock:    func() time.Time { return testNow },
	}
	return &harness{t: t, server: server, handler: server.Handler(), store: st}, mailer
}

func (h *harness) do(method, path, token string, body any) *httptest.ResponseRecorder {
	h.t.Helper()

	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			h.t.Fatalf("serializando el cuerpo: %v", err)
		}
		reader = bytes.NewReader(raw)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	res := httptest.NewRecorder()
	h.handler.ServeHTTP(res, req)
	return res
}

// login devuelve el access token de una cuenta del dataset.
func (h *harness) login(email string) string {
	h.t.Helper()

	res := h.do(http.MethodPost, api.BasePath+"/auth/login", "",
		map[string]string{"email": email, "password": store.DevPassword})
	if res.Code != http.StatusOK {
		h.t.Fatalf("login de %s: estado %d, cuerpo %s", email, res.Code, res.Body)
	}

	var tokens contract.AuthTokens
	decode(h.t, res, &tokens)
	return tokens.AccessToken
}

func decode(t *testing.T, res *httptest.ResponseRecorder, into any) {
	t.Helper()
	if err := json.Unmarshal(res.Body.Bytes(), into); err != nil {
		t.Fatalf("parseando la respuesta: %v\ncuerpo: %s", err, res.Body)
	}
}

// expectError verifica estado y código del catálogo.
func expectError(t *testing.T, res *httptest.ResponseRecorder, status int, code contract.Code) {
	t.Helper()

	if res.Code != status {
		t.Errorf("estado %d, se esperaba %d — cuerpo: %s", res.Code, status, res.Body)
	}
	var body contract.APIError
	decode(t, res, &body)

	if body.Error.Code != string(code) {
		t.Errorf("code %q, se esperaba %q", body.Error.Code, code)
	}
	if body.Error.Message == "" {
		t.Error("el mensaje de error está vacío; se muestra al usuario tal cual")
	}
	if body.Error.Details == nil {
		t.Error("details es nil; el contrato dice que siempre existe, aunque sea {}")
	}
}

// ── Golden contra docs/contract/samples/ ──────────────────────────────────

// shapeOf reduce un JSON a su estructura de claves y tipos, sin los valores.
// Los valores cambian con el rebase de fechas; los NOMBRES DE CAMPO son el
// contrato y son lo que rompe al frontend en silencio.
func shapeOf(value any) string {
	switch v := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(v))
		for k := range v {
			if strings.HasPrefix(k, "_") { // _request y _note son anotaciones del sample
				continue
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)

		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, k+":"+shapeOf(v[k]))
		}
		return "{" + strings.Join(parts, ",") + "}"

	case []any:
		if len(v) == 0 {
			return "[]"
		}
		return "[" + shapeOf(v[0]) + "]"

	case float64:
		return "number"
	case string:
		return "string"
	case bool:
		return "bool"
	default:
		return "null"
	}
}

func shapeFromSample(t *testing.T, name string) string {
	t.Helper()
	raw, err := seed.Sample(name)
	if err != nil {
		t.Fatalf("leyendo el sample %s: %v", name, err)
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatalf("parseando el sample %s: %v", name, err)
	}
	return shapeOf(value)
}

func shapeFromResponse(t *testing.T, res *httptest.ResponseRecorder) string {
	t.Helper()
	var value any
	if err := json.Unmarshal(res.Body.Bytes(), &value); err != nil {
		t.Fatalf("parseando la respuesta: %v\ncuerpo: %s", err, res.Body)
	}
	return shapeOf(value)
}

// TestResponsesMatchContractSamples es EL test de contrato de este paquete.
// El frontend tiene interfaces TypeScript con exactamente estos nombres de
// campo; si acá cambia uno, el otro lado se rompe sin que nadie se entere.
func TestResponsesMatchContractSamples(t *testing.T) {
	h, _ := newHarness(t)
	token := h.login("ana@hipodromo.test")

	casos := []struct {
		nombre string
		sample string
		hacer  func() *httptest.ResponseRecorder
	}{
		{"login", "auth.login.200.json", func() *httptest.ResponseRecorder {
			return h.do(http.MethodPost, api.BasePath+"/auth/login", "",
				map[string]string{"email": "ana@hipodromo.test", "password": store.DevPassword})
		}},
		{"me", "me.200.json", func() *httptest.ResponseRecorder {
			return h.do(http.MethodGet, api.BasePath+"/me", token, nil)
		}},
		{"listado de carreras", "races.list.200.json", func() *httptest.ResponseRecorder {
			return h.do(http.MethodGet, api.BasePath+"/races?status=upcoming&page=1&size=2", "", nil)
		}},
		{"detalle de carrera", "races.detail.200.json", func() *httptest.ResponseRecorder {
			return h.do(http.MethodGet, api.BasePath+"/races/race_005", "", nil)
		}},
		{"resultados", "races.results.200.json", func() *httptest.ResponseRecorder {
			return h.do(http.MethodGet, api.BasePath+"/races/race_003/results", h.login("bruno@hipodromo.test"), nil)
		}},
		{"mis apuestas", "bets.me.200.json", func() *httptest.ResponseRecorder {
			return h.do(http.MethodGet, api.BasePath+"/bets/me?page=1&size=3", token, nil)
		}},
		{"leaderboard", "leaderboard.all.200.json", func() *httptest.ResponseRecorder {
			return h.do(http.MethodGet, api.BasePath+"/leaderboard?period=all", "", nil)
		}},
		{"crear apuesta", "bets.create.201.json", func() *httptest.ResponseRecorder {
			return h.do(http.MethodPost, api.BasePath+"/bets", token,
				map[string]any{"raceId": "race_006", "horseId": "hrs_036", "amount": 200})
		}},
		{"saldo insuficiente", "error.insufficient-balance.409.json", func() *httptest.ResponseRecorder {
			return h.do(http.MethodPost, api.BasePath+"/bets", h.login("hugo@hipodromo.test"),
				map[string]any{"raceId": "race_006", "horseId": "hrs_036", "amount": 5000})
		}},
		{"validación", "error.validation.422.json", func() *httptest.ResponseRecorder {
			return h.do(http.MethodPost, api.BasePath+"/auth/register", "",
				map[string]string{"email": "no-es-un-correo", "password": "corta", "displayName": "A"})
		}},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			res := caso.hacer()
			got := shapeFromResponse(t, res)
			want := shapeFromSample(t, caso.sample)
			if got != want {
				t.Errorf("la forma de la respuesta se separó de %s\n  servidor: %s\n  sample:   %s",
					caso.sample, got, want)
			}
		})
	}
}

// TestLeaderboardMatchesGolden compara el cálculo del servidor contra
// docs/contract/seed/leaderboard.json, que es la expectativa escrita a mano.
func TestLeaderboardMatchesGolden(t *testing.T) {
	h, _ := newHarness(t)

	raw, err := seed.SeedFile("leaderboard.json")
	if err != nil {
		t.Fatalf("leyendo el golden: %v", err)
	}
	var golden struct {
		All   []contract.LeaderboardEntry `json:"all"`
		Daily []contract.LeaderboardEntry `json:"daily"`
	}
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatalf("parseando el golden: %v", err)
	}

	for _, caso := range []struct {
		period string
		want   []contract.LeaderboardEntry
	}{{"all", golden.All}, {"daily", golden.Daily}} {
		t.Run(caso.period, func(t *testing.T) {
			res := h.do(http.MethodGet, api.BasePath+"/leaderboard?period="+caso.period, "", nil)
			if res.Code != http.StatusOK {
				t.Fatalf("estado %d", res.Code)
			}
			var got contract.Leaderboard
			decode(t, res, &got)

			if len(got.Entries) != len(caso.want) {
				t.Fatalf("%d entradas, el golden tiene %d", len(got.Entries), len(caso.want))
			}
			for i := range caso.want {
				if got.Entries[i] != caso.want[i] {
					t.Errorf("puesto %d:\n  servidor %+v\n  golden   %+v", i+1, got.Entries[i], caso.want[i])
				}
			}
		})
	}
}

// ── Carreras ──────────────────────────────────────────────────────────────

func TestRaceListing(t *testing.T) {
	h, _ := newHarness(t)

	t.Run("ordena en vivo, por venir y terminadas", func(t *testing.T) {
		res := h.do(http.MethodGet, api.BasePath+"/races?size=50", "", nil)
		var page contract.Page[contract.Race]
		decode(t, res, &page)

		if page.Total != 8 {
			t.Fatalf("total %d, el dataset tiene 8 carreras", page.Total)
		}
		rank := map[contract.RaceStatus]int{contract.StatusLive: 0, contract.StatusUpcoming: 1, contract.StatusFinished: 2}
		for i := 1; i < len(page.Items); i++ {
			if rank[page.Items[i-1].Status] > rank[page.Items[i].Status] {
				t.Fatalf("%s (%s) quedó antes que %s (%s)",
					page.Items[i-1].ID, page.Items[i-1].Status, page.Items[i].ID, page.Items[i].Status)
			}
		}
	})

	t.Run("el listado trae los caballos", func(t *testing.T) {
		res := h.do(http.MethodGet, api.BasePath+"/races?size=50", "", nil)
		var page contract.Page[contract.Race]
		decode(t, res, &page)

		for _, race := range page.Items {
			if len(race.Horses) < 2 {
				t.Errorf("%s vino con %d caballos; el contrato exige al menos 2 en el listado",
					race.ID, len(race.Horses))
			}
		}
	})

	t.Run("items nunca es null", func(t *testing.T) {
		// Un `null` en vez de `[]` rompe el @for del frontend.
		res := h.do(http.MethodGet, api.BasePath+"/races?page=99", "", nil)
		if !strings.Contains(res.Body.String(), `"items":[]`) {
			t.Errorf("una página vacía devolvió %s", res.Body)
		}
	})

	t.Run("status inválido", func(t *testing.T) {
		res := h.do(http.MethodGet, api.BasePath+"/races?status=galopando", "", nil)
		expectError(t, res, http.StatusUnprocessableEntity, contract.CodeValidationFailed)
	})

	t.Run("carrera inexistente", func(t *testing.T) {
		res := h.do(http.MethodGet, api.BasePath+"/races/race_099", "", nil)
		expectError(t, res, http.StatusNotFound, contract.CodeNotFound)
	})
}

func TestResults(t *testing.T) {
	h, _ := newHarness(t)

	t.Run("carrera que no terminó", func(t *testing.T) {
		res := h.do(http.MethodGet, api.BasePath+"/races/race_006/results", "", nil)
		expectError(t, res, http.StatusConflict, contract.CodeResultsNotAvailable)
	})

	t.Run("sin sesión el podio es público y los pagos no", func(t *testing.T) {
		res := h.do(http.MethodGet, api.BasePath+"/races/race_003/results", "", nil)
		var result contract.RaceResult
		decode(t, res, &result)

		if len(result.Podium) != 3 {
			t.Errorf("el podio tiene %d puestos", len(result.Podium))
		}
		if len(result.Payouts) != 0 {
			t.Errorf("sin sesión llegaron %d pagos; no debería llegar ninguno", len(result.Payouts))
		}
	})

	t.Run("con sesión llegan solo los pagos propios", func(t *testing.T) {
		res := h.do(http.MethodGet, api.BasePath+"/races/race_003/results", h.login("bruno@hipodromo.test"), nil)
		var result contract.RaceResult
		decode(t, res, &result)

		if len(result.Payouts) != 1 {
			t.Fatalf("%d pagos, se esperaba 1", len(result.Payouts))
		}
		// bet_008: 120 a Guaraní @11.00, que ganó.
		if got := result.Payouts[0]; got.BetID != "bet_008" || got.Stake != 120 || got.Amount != 1320 {
			t.Errorf("pago %+v; se esperaba bet_008 con stake 120 y amount 1320", got)
		}
	})
}

// ── Autenticación ─────────────────────────────────────────────────────────

func TestLogin(t *testing.T) {
	h, _ := newHarness(t)

	t.Run("contraseña incorrecta", func(t *testing.T) {
		res := h.do(http.MethodPost, api.BasePath+"/auth/login", "",
			map[string]string{"email": "ana@hipodromo.test", "password": "cualquiera"})
		expectError(t, res, http.StatusUnauthorized, contract.CodeInvalidCredentials)
	})

	t.Run("cuenta inexistente da el mismo error", func(t *testing.T) {
		// Distinguirlos le diría a cualquiera qué correos están registrados.
		res := h.do(http.MethodPost, api.BasePath+"/auth/login", "",
			map[string]string{"email": "nadie@hipodromo.test", "password": store.DevPassword})
		expectError(t, res, http.StatusUnauthorized, contract.CodeInvalidCredentials)
	})
}

func TestMe(t *testing.T) {
	h, _ := newHarness(t)

	t.Run("sin token", func(t *testing.T) {
		res := h.do(http.MethodGet, api.BasePath+"/me", "", nil)
		expectError(t, res, http.StatusUnauthorized, contract.CodeUnauthenticated)
	})

	t.Run("token con firma inválida", func(t *testing.T) {
		res := h.do(http.MethodGet, api.BasePath+"/me", "esto.no.es-un-jwt", nil)
		expectError(t, res, http.StatusUnauthorized, contract.CodeUnauthenticated)
	})

	t.Run("con sesión", func(t *testing.T) {
		res := h.do(http.MethodGet, api.BasePath+"/me", h.login("ana@hipodromo.test"), nil)
		var user contract.User
		decode(t, res, &user)
		if user.ID != "usr_001" || user.Balance != 5000 {
			t.Errorf("%+v; se esperaba usr_001 con saldo 5000", user)
		}
	})
}

func TestRefreshTokenIsSingleUse(t *testing.T) {
	h, _ := newHarness(t)

	res := h.do(http.MethodPost, api.BasePath+"/auth/login", "",
		map[string]string{"email": "ana@hipodromo.test", "password": store.DevPassword})
	var tokens contract.AuthTokens
	decode(t, res, &tokens)

	first := h.do(http.MethodPost, api.BasePath+"/auth/refresh", "",
		map[string]string{"refreshToken": tokens.RefreshToken})
	if first.Code != http.StatusOK {
		t.Fatalf("el primer canje falló: %d %s", first.Code, first.Body)
	}

	// Reusar un token ya canjeado es la señal de que fue robado.
	second := h.do(http.MethodPost, api.BasePath+"/auth/refresh", "",
		map[string]string{"refreshToken": tokens.RefreshToken})
	expectError(t, second, http.StatusUnauthorized, contract.CodeInvalidRefreshToken)
}

func TestRegisterAndVerify(t *testing.T) {
	h, mailer := newHarness(t)

	res := h.do(http.MethodPost, api.BasePath+"/auth/register", "",
		map[string]string{"email": "nuevo@hipodromo.test", "password": "Carrera123!", "displayName": "Nuevo Apostador"})
	if res.Code != http.StatusCreated {
		t.Fatalf("registro: estado %d, cuerpo %s", res.Code, res.Body)
	}

	var created contract.User
	decode(t, res, &created)
	if created.Balance != contract.SignupBalance {
		t.Errorf("saldo inicial %d, se esperaba %d", created.Balance, contract.SignupBalance)
	}
	if created.EmailVerified {
		t.Error("la cuenta quedó verificada sin haber canjeado el token")
	}

	t.Run("el correo repetido se rechaza", func(t *testing.T) {
		res := h.do(http.MethodPost, api.BasePath+"/auth/register", "",
			map[string]string{"email": "nuevo@hipodromo.test", "password": "Carrera123!", "displayName": "Otro"})
		expectError(t, res, http.StatusConflict, contract.CodeEmailAlreadyRegistered)
	})

	t.Run("token inválido", func(t *testing.T) {
		res := h.do(http.MethodPost, api.BasePath+"/auth/verify", "", map[string]string{"token": "vt_inventado"})
		expectError(t, res, http.StatusBadRequest, contract.CodeInvalidVerificationToken)
	})

	// El envío es asíncrono: se espera a que el mailer registre el token.
	token := waitForToken(t, mailer)

	t.Run("canje correcto", func(t *testing.T) {
		res := h.do(http.MethodPost, api.BasePath+"/auth/verify", "", map[string]string{"token": token})
		if res.Code != http.StatusOK {
			t.Fatalf("estado %d, cuerpo %s", res.Code, res.Body)
		}
		var body contract.VerifiedUser
		decode(t, res, &body)
		if !body.User.EmailVerified {
			t.Error("la cuenta sigue sin verificar después del canje")
		}
	})

	t.Run("el token no se puede reusar", func(t *testing.T) {
		res := h.do(http.MethodPost, api.BasePath+"/auth/verify", "", map[string]string{"token": token})
		expectError(t, res, http.StatusBadRequest, contract.CodeInvalidVerificationToken)
	})
}

func waitForToken(t *testing.T, mailer *silentMailer) string {
	t.Helper()
	for i := 0; i < 100; i++ {
		if len(mailer.sent) > 0 {
			_, token, _ := strings.Cut(mailer.sent[0], "|")
			return token
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("no se envió ningún correo de verificación")
	return ""
}

// ── Apuestas ──────────────────────────────────────────────────────────────

func TestPlaceBet(t *testing.T) {
	h, _ := newHarness(t)
	ana := h.login("ana@hipodromo.test")

	t.Run("sin sesión", func(t *testing.T) {
		res := h.do(http.MethodPost, api.BasePath+"/bets", "",
			map[string]any{"raceId": "race_006", "horseId": "hrs_036", "amount": 100})
		expectError(t, res, http.StatusUnauthorized, contract.CodeUnauthenticated)
	})

	t.Run("descuenta el saldo y congela la cuota", func(t *testing.T) {
		res := h.do(http.MethodPost, api.BasePath+"/bets", ana,
			map[string]any{"raceId": "race_006", "horseId": "hrs_036", "amount": 200})
		if res.Code != http.StatusCreated {
			t.Fatalf("estado %d, cuerpo %s", res.Code, res.Body)
		}

		var created contract.BetCreated
		decode(t, res, &created)

		if created.Balance != 4800 {
			t.Errorf("saldo %d, se esperaba 4800", created.Balance)
		}
		if created.Bet.Odds != 3.40 {
			t.Errorf("cuota %v, la de Candela es 3.40", created.Bet.Odds)
		}
		if created.Bet.Status != contract.BetPending || created.Bet.Payout != 0 {
			t.Errorf("una apuesta nueva salió como %s con payout %d", created.Bet.Status, created.Bet.Payout)
		}
		// Denormalizado: el historial se pinta sin ir a buscar la carrera.
		if created.Bet.RaceName != "Premio Estrella del Sur" || created.Bet.HorseName != "Candela" {
			t.Errorf("nombres denormalizados mal: %q / %q", created.Bet.RaceName, created.Bet.HorseName)
		}
	})

	t.Run("correo sin verificar", func(t *testing.T) {
		res := h.do(http.MethodPost, api.BasePath+"/bets", h.login("caro@hipodromo.test"),
			map[string]any{"raceId": "race_006", "horseId": "hrs_036", "amount": 100})
		expectError(t, res, http.StatusForbidden, contract.CodeEmailNotVerified)
	})

	t.Run("la carrera ya arrancó", func(t *testing.T) {
		res := h.do(http.MethodPost, api.BasePath+"/bets", ana,
			map[string]any{"raceId": "race_005", "horseId": "hrs_029", "amount": 100})
		expectError(t, res, http.StatusConflict, contract.CodeRaceAlreadyStarted)
	})

	t.Run("el caballo no corre en esa carrera", func(t *testing.T) {
		res := h.do(http.MethodPost, api.BasePath+"/bets", ana,
			map[string]any{"raceId": "race_006", "horseId": "hrs_001", "amount": 100})
		expectError(t, res, http.StatusUnprocessableEntity, contract.CodeHorseNotInRace)
	})

	t.Run("monto fuera de rango", func(t *testing.T) {
		for _, amount := range []int{0, 9, 5001} {
			res := h.do(http.MethodPost, api.BasePath+"/bets", ana,
				map[string]any{"raceId": "race_006", "horseId": "hrs_036", "amount": amount})
			expectError(t, res, http.StatusUnprocessableEntity, contract.CodeBetAmountOutOfRange)
		}
	})

	t.Run("saldo insuficiente", func(t *testing.T) {
		res := h.do(http.MethodPost, api.BasePath+"/bets", h.login("hugo@hipodromo.test"),
			map[string]any{"raceId": "race_006", "horseId": "hrs_036", "amount": 5000})
		expectError(t, res, http.StatusConflict, contract.CodeInsufficientBalance)

		var body contract.APIError
		decode(t, res, &body)
		if body.Error.Details["balance"] != float64(980) {
			t.Errorf("details.balance = %v, se esperaba 980", body.Error.Details["balance"])
		}
	})

	t.Run("carrera inexistente", func(t *testing.T) {
		res := h.do(http.MethodPost, api.BasePath+"/bets", ana,
			map[string]any{"raceId": "race_099", "horseId": "hrs_036", "amount": 100})
		expectError(t, res, http.StatusNotFound, contract.CodeNotFound)
	})

	t.Run("campo desconocido en el cuerpo", func(t *testing.T) {
		// Mejor un 422 explícito que un cero silencioso.
		res := h.do(http.MethodPost, api.BasePath+"/bets", ana,
			map[string]any{"raceId": "race_006", "horseId": "hrs_036", "monto": 100})
		expectError(t, res, http.StatusUnprocessableEntity, contract.CodeValidationFailed)
	})
}

func TestBetHistory(t *testing.T) {
	h, _ := newHarness(t)

	res := h.do(http.MethodGet, api.BasePath+"/bets/me?size=50", h.login("ana@hipodromo.test"), nil)
	var page contract.Page[contract.Bet]
	decode(t, res, &page)

	if page.Total != 6 {
		t.Fatalf("total %d, Ana tiene 6 apuestas en el dataset", page.Total)
	}
	for i := 1; i < len(page.Items); i++ {
		if page.Items[i-1].PlacedAt < page.Items[i].PlacedAt {
			t.Fatalf("el historial no viene de la más reciente a la más vieja: %s antes que %s",
				page.Items[i-1].PlacedAt, page.Items[i].PlacedAt)
		}
	}
	for _, bet := range page.Items {
		if bet.UserID != "usr_001" {
			t.Errorf("se filtró una apuesta de %s en el historial de Ana", bet.UserID)
		}
	}
}

// ── Liquidación ───────────────────────────────────────────────────────────

// TestSettlement recorre el ciclo completo: se apuesta, corre la carrera y se
// paga. Es lo que el alumno ve en S10.
func TestSettlement(t *testing.T) {
	h, _ := newHarness(t)
	ana := h.login("ana@hipodromo.test")

	// Ana apuesta 500 a Coirón (hrs_038 @2.30) en race_006.
	res := h.do(http.MethodPost, api.BasePath+"/bets", ana,
		map[string]any{"raceId": "race_006", "horseId": "hrs_038", "amount": 500})
	if res.Code != http.StatusCreated {
		t.Fatalf("apostando: %d %s", res.Code, res.Body)
	}
	var created contract.BetCreated
	decode(t, res, &created)

	// Se fuerza la llegada con Coirón ganando.
	race, _ := h.store.Race("race_006")
	winner, _ := race.Horse("hrs_038")
	podium := []contract.PodiumEntry{
		{Place: 1, HorseID: winner.ID, HorseName: winner.Name, Number: winner.Number, Odds: winner.Odds},
		{Place: 2, HorseID: "hrs_036", HorseName: "Candela", Number: 1, Odds: 3.40},
		{Place: 3, HorseID: "hrs_037", HorseName: "Pehuén", Number: 2, Odds: 5.90},
	}
	settlements := h.store.SettleRace("race_006", podium, testNow.Add(time.Minute))

	if len(settlements) == 0 {
		t.Fatal("la liquidación no devolvió nada")
	}

	t.Run("paga amount por odds", func(t *testing.T) {
		res := h.do(http.MethodGet, api.BasePath+"/bets/me?size=50", ana, nil)
		var page contract.Page[contract.Bet]
		decode(t, res, &page)

		for _, bet := range page.Items {
			if bet.ID != created.Bet.ID {
				continue
			}
			if bet.Status != contract.BetWon {
				t.Fatalf("la apuesta quedó %s", bet.Status)
			}
			if bet.Payout != 1150 { // 500 × 2.30
				t.Errorf("payout %d, se esperaba 1150", bet.Payout)
			}
			return
		}
		t.Fatal("la apuesta no apareció en el historial")
	})

	t.Run("acredita el saldo, incluida la apuesta que ya tenía pendiente", func(t *testing.T) {
		// Ana ya traía bet_006 del dataset: 150 al mismo Coirón. Liquidar la
		// carrera paga las dos, no solo la de este test.
		//   5000 − 500 (apuesta) + 1150 (500 × 2.30) + 345 (150 × 2.30) = 5995
		res := h.do(http.MethodGet, api.BasePath+"/me", ana, nil)
		var user contract.User
		decode(t, res, &user)
		if user.Balance != 5995 {
			t.Errorf("saldo %d, se esperaba 5995", user.Balance)
		}
	})

	t.Run("la carrera queda terminada y con resultado", func(t *testing.T) {
		res := h.do(http.MethodGet, api.BasePath+"/races/race_006/results", ana, nil)
		if res.Code != http.StatusOK {
			t.Fatalf("estado %d, cuerpo %s", res.Code, res.Body)
		}
		var result contract.RaceResult
		decode(t, res, &result)
		if result.Podium[0].HorseID != "hrs_038" {
			t.Errorf("ganador %s", result.Podium[0].HorseID)
		}
	})

	t.Run("ya no se puede apostar", func(t *testing.T) {
		res := h.do(http.MethodPost, api.BasePath+"/bets", ana,
			map[string]any{"raceId": "race_006", "horseId": "hrs_036", "amount": 100})
		expectError(t, res, http.StatusConflict, contract.CodeRaceAlreadyStarted)
	})
}

// ── Límite de intentos ────────────────────────────────────────────────────

func TestLoginRateLimit(t *testing.T) {
	h, _ := newHarness(t)

	var last *httptest.ResponseRecorder
	for i := 0; i < 12; i++ {
		last = h.do(http.MethodPost, api.BasePath+"/auth/login", "",
			map[string]string{"email": "ana@hipodromo.test", "password": "mal"})
	}
	expectError(t, last, http.StatusTooManyRequests, contract.CodeRateLimited)

	var body contract.APIError
	decode(t, last, &body)
	if _, ok := body.Error.Details["retryAfterSeconds"]; !ok {
		t.Error("falta details.retryAfterSeconds, que el contrato exige en RATE_LIMITED")
	}
}

// ── CORS y salud ──────────────────────────────────────────────────────────

func TestCORSPreflight(t *testing.T) {
	h, _ := newHarness(t)

	req := httptest.NewRequest(http.MethodOptions, api.BasePath+"/bets", nil)
	req.Header.Set("Origin", "http://localhost:4200")
	res := httptest.NewRecorder()
	h.handler.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Errorf("estado %d, se esperaba 204", res.Code)
	}
	if !strings.Contains(res.Header().Get("Access-Control-Allow-Headers"), "Authorization") {
		t.Error("el preflight no permite la cabecera Authorization")
	}
}

func TestHealth(t *testing.T) {
	h, _ := newHarness(t)

	res := h.do(http.MethodGet, "/health", "", nil)
	if res.Code != http.StatusOK {
		t.Fatalf("estado %d", res.Code)
	}
	var body map[string]any
	decode(t, res, &body)
	if body["status"] != "ok" {
		t.Errorf("status %v", body["status"])
	}
}

// TestEveryErrorCodeHasStatusAndMessage: un código sin entrada en el catálogo
// se degradaría a 500 y el frontend nunca podría reaccionar a él.
func TestEveryErrorCodeHasStatusAndMessage(t *testing.T) {
	for _, code := range contract.KnownCodes() {
		fault := contract.Errorf(code)
		if fault.Status() < 400 || fault.Status() > 599 {
			t.Errorf("%s: estado %d", code, fault.Status())
		}
		if fault.Message() == "" {
			t.Errorf("%s: sin mensaje", code)
		}
		if body := fault.Body(); body.Error.Details == nil {
			t.Errorf("%s: details quedó nil", code)
		}
	}
	if got := len(contract.KnownCodes()); got != 18 {
		t.Errorf("%d códigos en el catálogo; docs/contract/error-codes.md documenta 18", got)
	}
}
