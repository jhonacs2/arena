package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/talentodh/arena/internal/accounts"
	"github.com/talentodh/arena/internal/api"
	"github.com/talentodh/arena/internal/auth"
	"github.com/talentodh/arena/internal/ledger"
	"github.com/talentodh/arena/internal/testdb"
)

const testSecret = "secreto-de-prueba"

type harness struct {
	t        *testing.T
	server   *api.Server
	handler  http.Handler
	accounts *accounts.Store
	ledger   *ledger.Store
}

func newHarness(t *testing.T) *harness {
	t.Helper()

	pool := testdb.Pool(t)
	accountStore := accounts.New(pool)
	ledgerStore := ledger.New(pool)

	server := &api.Server{
		Accounts:       accountStore,
		Ledger:         ledgerStore,
		Signer:         auth.NewSigner(testSecret),
		Log:            slog.New(slog.NewTextHandler(io.Discard, nil)),
		AllowedOrigins: []string{"http://localhost:4200"},
		Cookie:         api.DefaultCookieOptions(),
		Clock:          time.Now,
	}

	return &harness{
		t: t, server: server, handler: server.Handler(),
		accounts: accountStore, ledger: ledgerStore,
	}
}

// do manda una petición. Un token vacío es una petición anónima.
func (h *harness) do(method, path, token string, body any) *httptest.ResponseRecorder {
	h.t.Helper()

	var payload io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			h.t.Fatal(err)
		}
		payload = bytes.NewReader(raw)
	}

	req := httptest.NewRequest(method, path, payload)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)
	return rec
}

// errorCode saca el `code` del sobre de error de una respuesta.
func errorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("la respuesta no es el sobre de error: %s", rec.Body.String())
	}
	if body.Error.Message == "" {
		t.Error("el sobre de error vino sin message, y el message se le muestra al usuario")
	}
	return body.Error.Code
}

func (h *harness) token(userID string) string {
	h.t.Helper()

	token, err := h.server.Signer.Sign(userID, time.Now())
	if err != nil {
		h.t.Fatal(err)
	}
	return token
}

// student crea un alumno con su código canjeado y devuelve el usuario y su token.
func (h *harness) student(username string) (accounts.User, string) {
	h.t.Helper()

	admin := h.admin("profe-" + username)

	// Un código legítimo, generado por el instructor: el alumno se crea por el
	// mismo camino que en producción y no con un insert a mano.
	codes, err := h.accounts.CreateCodes(h.t.Context(), admin.ID, 1, 1000, "")
	if err != nil {
		h.t.Fatal(err)
	}

	hash, err := auth.HashPassword("caballo-de-batalla")
	if err != nil {
		h.t.Fatal(err)
	}
	user, err := h.accounts.Redeem(h.t.Context(), accounts.RedeemInput{
		Code: codes[0], FirstName: "Ana", LastName: "Gómez",
		Username: username, PasswordHash: hash,
	})
	if err != nil {
		h.t.Fatal(err)
	}
	return user, h.token(user.ID)
}

func (h *harness) admin(username string) accounts.User {
	h.t.Helper()

	user, err := h.accounts.EnsureAdmin(h.t.Context(), username, "Jhonatan", "Soto", "x")
	if err != nil {
		h.t.Fatal(err)
	}
	return user
}

// ── El rol, que es la razón de ser de todo esto ───────────────────────────

// adminRoutes son TODAS las rutas de instructor de este paquete. Si se agrega
// una, se agrega acá: es la lista que el test recorre.
var adminRoutes = []struct{ method, path string }{
	{http.MethodPost, "/api/admin/codes"},
	{http.MethodGet, "/api/admin/codes"},
	{http.MethodGet, "/api/admin/scores"},
	{http.MethodPost, "/api/admin/users/11111111-1111-1111-1111-111111111111/gift"},
	{http.MethodPost, "/api/admin/users/11111111-1111-1111-1111-111111111111/grant-points"},
}

// **El test que justifica el middleware de rol.** Las monedas son nota: el que se
// cuela en /admin/ no gana una partida, se pone una calificación.
//
// Un alumno con un token perfectamente válido —firmado por nosotros, sin vencer—
// recibe 403 en todas las rutas de instructor. El rol no viaja en el token: se
// lee de `users.role` en cada petición, así que no hay nada que editar del lado
// del alumno que sirva de algo.
func TestAdminRechazaAUnAlumno(t *testing.T) {
	h := newHarness(t)
	_, token := h.student("anag")

	for _, route := range adminRoutes {
		rec := h.do(route.method, route.path, token, map[string]any{})

		if rec.Code != http.StatusForbidden {
			t.Errorf("%s %s: estado %d, se esperaba 403", route.method, route.path, rec.Code)
			continue
		}
		if got := errorCode(t, rec); got != "FORBIDDEN" {
			t.Errorf("%s %s: code %q, se esperaba FORBIDDEN", route.method, route.path, got)
		}
	}
}

func TestAdminRechazaSinSesion(t *testing.T) {
	h := newHarness(t)

	for _, route := range adminRoutes {
		rec := h.do(route.method, route.path, "", map[string]any{})

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: estado %d, se esperaba 401", route.method, route.path, rec.Code)
			continue
		}
		if got := errorCode(t, rec); got != "UNAUTHENTICATED" {
			t.Errorf("%s %s: code %q, se esperaba UNAUTHENTICATED", route.method, route.path, got)
		}
	}
}

// Un token firmado con OTRO secreto no sirve: es el caso de alguien que se
// fabrica un token con `role: admin` adentro.
func TestAdminRechazaTokenDeOtroSecreto(t *testing.T) {
	h := newHarness(t)
	admin := h.admin("profe")

	falso, err := auth.NewSigner("no-es-el-secreto").Sign(admin.ID, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	rec := h.do(http.MethodGet, "/api/admin/scores", falso, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("estado %d, se esperaba 401", rec.Code)
	}
}

// Y el chequeo es contra la BASE: un token emitido cuando alguien era admin deja
// de servir en cuanto se le baja el rol, sin esperar los 15 minutos del JWT.
func TestBajarElRolInvalidaElAccesoDeInmediato(t *testing.T) {
	h := newHarness(t)
	admin := h.admin("profe")
	token := h.token(admin.ID)

	if rec := h.do(http.MethodGet, "/api/admin/scores", token, nil); rec.Code != http.StatusOK {
		t.Fatalf("el admin no pudo entrar: %d %s", rec.Code, rec.Body.String())
	}

	pool := h.ledger.Pool
	if _, err := pool.Exec(t.Context(),
		`update users set role = 'student' where id = $1`, admin.ID); err != nil {
		t.Fatal(err)
	}

	rec := h.do(http.MethodGet, "/api/admin/scores", token, nil)
	if rec.Code != http.StatusForbidden {
		t.Errorf("estado %d con el MISMO token después de bajarle el rol, se esperaba 403", rec.Code)
	}
}

func TestAdminConRolAdminSiPuede(t *testing.T) {
	h := newHarness(t)
	admin := h.admin("profe")
	token := h.token(admin.ID)

	rec := h.do(http.MethodPost, "/api/admin/codes", token,
		map[string]any{"count": 3, "coinsGranted": 1000, "note": "grupo del martes"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("estado %d: %s", rec.Code, rec.Body.String())
	}

	var body struct{ Codes []string }
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Codes) != 3 {
		t.Fatalf("codes = %v, se esperaban 3", body.Codes)
	}
	for _, code := range body.Codes {
		if strings.ContainsAny(code, "ILOU") {
			t.Errorf("%s tiene una letra ambigua", code)
		}
	}

	if rec := h.do(http.MethodGet, "/api/admin/codes", token, nil); rec.Code != http.StatusOK {
		t.Errorf("GET /admin/codes: estado %d", rec.Code)
	}
	if rec := h.do(http.MethodGet, "/api/admin/scores", token, nil); rec.Code != http.StatusOK {
		t.Errorf("GET /admin/scores: estado %d", rec.Code)
	}
}

// ── El registro de punta a punta ──────────────────────────────────────────

func TestFlujoCompletoDeRegistro(t *testing.T) {
	h := newHarness(t)
	admin := h.admin("profe")
	adminToken := h.token(admin.ID)

	// El instructor genera un código.
	rec := h.do(http.MethodPost, "/api/admin/codes", adminToken, map[string]any{"count": 1})
	if rec.Code != http.StatusCreated {
		t.Fatalf("creando el código: %d %s", rec.Code, rec.Body.String())
	}
	var creados struct{ Codes []string }
	if err := json.Unmarshal(rec.Body.Bytes(), &creados); err != nil {
		t.Fatal(err)
	}
	code := creados.Codes[0]

	// El alumno lo valida sin canjearlo. En minúsculas y sin guion, como lo
	// escribiría alguien que lo copió de un chat.
	sinGuion := strings.ToLower(strings.ReplaceAll(code, "-", ""))
	rec = h.do(http.MethodPost, "/api/auth/check-code", "", map[string]any{"code": sinGuion})
	if rec.Code != http.StatusOK {
		t.Fatalf("check-code: %d %s", rec.Code, rec.Body.String())
	}
	var check struct {
		Valid        bool  `json:"valid"`
		CoinsGranted int64 `json:"coinsGranted"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &check); err != nil {
		t.Fatal(err)
	}
	if !check.Valid || check.CoinsGranted != 1000 {
		t.Errorf("check-code devolvió %+v", check)
	}

	// Y lo canjea.
	rec = h.do(http.MethodPost, "/api/auth/redeem", "", map[string]any{
		"code": code, "firstName": "Ana", "lastName": "Gómez",
		"username": "anag", "password": "caballo-de-batalla",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("redeem: %d %s", rec.Code, rec.Body.String())
	}

	var sesion struct {
		AccessToken string `json:"accessToken"`
		User        struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			Role     string `json:"role"`
		} `json:"user"`
		Balance int64           `json:"balance"`
		Points  json.RawMessage `json:"points"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &sesion); err != nil {
		t.Fatal(err)
	}
	if sesion.AccessToken == "" || sesion.Balance != 1000 {
		t.Errorf("sesión inesperada: %+v", sesion)
	}
	if string(sesion.Points) != "10" {
		t.Errorf("points = %s, se esperaba 10", sesion.Points)
	}
	if sesion.User.Role != accounts.RoleStudent {
		t.Errorf("rol = %q", sesion.User.Role)
	}

	// El refresh viajó en una cookie HttpOnly y NO en el cuerpo.
	cookie := refreshCookie(t, rec)
	if !cookie.HttpOnly {
		t.Error("la cookie del refresh no es HttpOnly")
	}
	if strings.Contains(rec.Body.String(), cookie.Value) {
		t.Error("el refresh también salió en el cuerpo: la cookie HttpOnly deja de servir para nada")
	}

	// El mismo código no se puede canjear dos veces.
	rec = h.do(http.MethodPost, "/api/auth/redeem", "", map[string]any{
		"code": code, "firstName": "Bruno", "lastName": "Díaz",
		"username": "brunod", "password": "caballo-de-batalla",
	})
	if rec.Code != http.StatusConflict || errorCode(t, rec) != "CODE_ALREADY_REDEEMED" {
		t.Errorf("segundo canje: %d %s", rec.Code, rec.Body.String())
	}

	// /api/me con el token de la sesión.
	rec = h.do(http.MethodGet, "/api/me", sesion.AccessToken, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("/api/me: %d %s", rec.Code, rec.Body.String())
	}
	var me struct {
		Balance int64           `json:"balance"`
		Points  json.RawMessage `json:"points"`
		User    struct {
			Username string `json:"username"`
		} `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &me); err != nil {
		t.Fatal(err)
	}
	if me.Balance != 1000 || me.User.Username != "anag" || string(me.Points) != "10" {
		t.Errorf("/api/me devolvió %+v (points %s)", me, me.Points)
	}

	// El historial ya tiene el canje.
	rec = h.do(http.MethodGet, "/api/me/transactions", sesion.AccessToken, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("/api/me/transactions: %d %s", rec.Code, rec.Body.String())
	}
	var historial struct {
		Items []struct {
			Delta        int64  `json:"delta"`
			Reason       string `json:"reason"`
			BalanceAfter int64  `json:"balanceAfter"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &historial); err != nil {
		t.Fatal(err)
	}
	if len(historial.Items) != 1 {
		t.Fatalf("el historial tiene %d movimientos, se esperaba 1", len(historial.Items))
	}
	if historial.Items[0].Reason != ledger.ReasonCodeRedeemed || historial.Items[0].Delta != 1000 {
		t.Errorf("primer movimiento: %+v", historial.Items[0])
	}

	// Login con el usuario recién creado, en otra caja.
	rec = h.do(http.MethodPost, "/api/auth/login", "",
		map[string]any{"username": "AnaG", "password": "caballo-de-batalla"})
	if rec.Code != http.StatusOK {
		t.Errorf("login: %d %s", rec.Code, rec.Body.String())
	}
	rec = h.do(http.MethodPost, "/api/auth/login", "",
		map[string]any{"username": "anag", "password": "otra-cosa"})
	if rec.Code != http.StatusUnauthorized || errorCode(t, rec) != "INVALID_CREDENTIALS" {
		t.Errorf("login con contraseña mal: %d %s", rec.Code, rec.Body.String())
	}
}

func TestCheckCodeDistingueLosDosErrores(t *testing.T) {
	h := newHarness(t)

	// Formato mal: es un error de tipeo, no un código que no existe.
	rec := h.do(http.MethodPost, "/api/auth/check-code", "", map[string]any{"code": "nada"})
	if rec.Code != http.StatusBadRequest || errorCode(t, rec) != "VALIDATION_FAILED" {
		t.Errorf("formato inválido: %d %s", rec.Code, rec.Body.String())
	}

	// Formato bien, código inexistente.
	rec = h.do(http.MethodPost, "/api/auth/check-code", "", map[string]any{"code": "ZZZZ-9999"})
	if rec.Code != http.StatusNotFound || errorCode(t, rec) != "CODE_NOT_FOUND" {
		t.Errorf("código inexistente: %d %s", rec.Code, rec.Body.String())
	}
}

func TestRedeemValidaLosCampos(t *testing.T) {
	h := newHarness(t)
	admin := h.admin("profe")
	codes, err := h.accounts.CreateCodes(t.Context(), admin.ID, 1, 1000, "")
	if err != nil {
		t.Fatal(err)
	}

	casos := []struct {
		nombre string
		body   map[string]any
	}{
		{"sin nombre", map[string]any{"code": codes[0], "firstName": "", "lastName": "Gómez", "username": "anag", "password": "caballo-de-batalla"}},
		{"contraseña corta", map[string]any{"code": codes[0], "firstName": "Ana", "lastName": "Gómez", "username": "anag", "password": "corta"}},
		{"usuario con espacios", map[string]any{"code": codes[0], "firstName": "Ana", "lastName": "Gómez", "username": "ana g", "password": "caballo-de-batalla"}},
		{"usuario corto", map[string]any{"code": codes[0], "firstName": "Ana", "lastName": "Gómez", "username": "an", "password": "caballo-de-batalla"}},
	}

	for _, caso := range casos {
		rec := h.do(http.MethodPost, "/api/auth/redeem", "", caso.body)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: estado %d, se esperaba 400", caso.nombre, rec.Code)
			continue
		}
		if got := errorCode(t, rec); got != "VALIDATION_FAILED" {
			t.Errorf("%s: code %q", caso.nombre, got)
		}
	}

	// Y ninguno de esos intentos quemó el código.
	if _, err := h.accounts.CheckCode(t.Context(), codes[0]); err != nil {
		t.Errorf("el código se quemó con un intento inválido: %v", err)
	}
}

// Un campo de más es un error explícito y no un cero silencioso: si el frontend
// manda `coins` donde va `count`, mejor un 400 que regalar 0 monedas.
func TestUnCampoDeMasEsUnError(t *testing.T) {
	h := newHarness(t)
	admin := h.admin("profe")
	token := h.token(admin.ID)

	rec := h.do(http.MethodPost, "/api/admin/codes", token,
		map[string]any{"count": 1, "cantidad": 5})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("estado %d, se esperaba 400", rec.Code)
	}
}

// ── Refresh y logout ──────────────────────────────────────────────────────

func refreshCookie(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()

	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == api.DefaultCookieOptions().Name {
			return cookie
		}
	}
	t.Fatalf("la respuesta no trajo la cookie del refresh: %v", rec.Header().Values("Set-Cookie"))
	return nil
}

// withCookie manda una petición con la cookie del refresh.
func (h *harness) withCookie(method, path string, cookie *http.Cookie) *httptest.ResponseRecorder {
	h.t.Helper()

	req := httptest.NewRequest(method, path, nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)
	return rec
}

func TestRefreshEsDeUnSoloUso(t *testing.T) {
	h := newHarness(t)
	admin := h.admin("profe")
	codes, err := h.accounts.CreateCodes(t.Context(), admin.ID, 1, 1000, "")
	if err != nil {
		t.Fatal(err)
	}

	rec := h.do(http.MethodPost, "/api/auth/redeem", "", map[string]any{
		"code": codes[0], "firstName": "Ana", "lastName": "Gómez",
		"username": "anag", "password": "caballo-de-batalla",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("redeem: %d %s", rec.Code, rec.Body.String())
	}
	primera := refreshCookie(t, rec)

	// Primer refresh: sale un token nuevo y una cookie nueva.
	rec = h.withCookie(http.MethodPost, "/api/auth/refresh", primera)
	if rec.Code != http.StatusOK {
		t.Fatalf("refresh: %d %s", rec.Code, rec.Body.String())
	}
	segunda := refreshCookie(t, rec)
	if segunda.Value == primera.Value {
		t.Error("el refresh devolvió el mismo token: no es de un solo uso")
	}

	var sesion struct {
		AccessToken string `json:"accessToken"`
		Balance     int64  `json:"balance"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &sesion); err != nil {
		t.Fatal(err)
	}
	if sesion.AccessToken == "" || sesion.Balance != 1000 {
		t.Errorf("el refresh no devolvió una sesión usable: %+v", sesion)
	}

	// Reusar el primero falla, y además tira abajo el segundo: es la respuesta a
	// un token que aparece dos veces.
	rec = h.withCookie(http.MethodPost, "/api/auth/refresh", primera)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("reusar el refresh dio %d, se esperaba 401", rec.Code)
	}
	rec = h.withCookie(http.MethodPost, "/api/auth/refresh", segunda)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("el token de la familia comprometida sigue sirviendo: %d", rec.Code)
	}
}

func TestRefreshSinCookieEs401(t *testing.T) {
	h := newHarness(t)

	rec := h.withCookie(http.MethodPost, "/api/auth/refresh", nil)
	if rec.Code != http.StatusUnauthorized || errorCode(t, rec) != "UNAUTHENTICATED" {
		t.Errorf("estado %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLogoutCortaLaSesionYNoFallaSinSesion(t *testing.T) {
	h := newHarness(t)
	admin := h.admin("profe")
	codes, err := h.accounts.CreateCodes(t.Context(), admin.ID, 1, 1000, "")
	if err != nil {
		t.Fatal(err)
	}

	rec := h.do(http.MethodPost, "/api/auth/redeem", "", map[string]any{
		"code": codes[0], "firstName": "Ana", "lastName": "Gómez",
		"username": "anag", "password": "caballo-de-batalla",
	})
	cookie := refreshCookie(t, rec)

	if rec := h.withCookie(http.MethodPost, "/api/auth/logout", cookie); rec.Code != http.StatusNoContent {
		t.Errorf("logout: %d", rec.Code)
	}
	if rec := h.withCookie(http.MethodPost, "/api/auth/refresh", cookie); rec.Code != http.StatusUnauthorized {
		t.Errorf("el refresh sigue vivo después del logout: %d", rec.Code)
	}
	// Un logout no tiene por qué fallar nunca.
	if rec := h.withCookie(http.MethodPost, "/api/auth/logout", nil); rec.Code != http.StatusNoContent {
		t.Errorf("logout sin sesión: %d", rec.Code)
	}
}

// ── Regalos ───────────────────────────────────────────────────────────────

func TestGiftAcreditaDescuentaYRespetaElPiso(t *testing.T) {
	h := newHarness(t)
	admin := h.admin("profe")
	adminToken := h.token(admin.ID)
	ana, anaToken := h.student("anag")

	gift := func(coins int64) *httptest.ResponseRecorder {
		return h.do(http.MethodPost, "/api/admin/users/"+ana.ID+"/gift", adminToken,
			map[string]any{"coins": coins, "note": "participación en clase"})
	}

	rec := gift(300)
	if rec.Code != http.StatusOK {
		t.Fatalf("regalo: %d %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Balance int64           `json:"balance"`
		Points  json.RawMessage `json:"points"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Balance != 1300 || string(body.Points) != "13" {
		t.Errorf("saldo %d y puntos %s; se esperaban 1300 y 13", body.Balance, body.Points)
	}

	// Un ajuste negativo entra.
	if rec := gift(-300); rec.Code != http.StatusOK {
		t.Errorf("ajuste: %d %s", rec.Code, rec.Body.String())
	}

	// Uno que dejaría el saldo negativo, no.
	rec = gift(-5000)
	if rec.Code != http.StatusConflict || errorCode(t, rec) != "INSUFFICIENT_BALANCE" {
		t.Errorf("un ajuste mayor al saldo dio %d %s", rec.Code, rec.Body.String())
	}

	// Cero no es un regalo.
	if rec := gift(0); rec.Code != http.StatusBadRequest {
		t.Errorf("regalo de 0: %d", rec.Code)
	}

	// Un id que no existe es 404, no 500.
	rec = h.do(http.MethodPost, "/api/admin/users/99999999-9999-9999-9999-999999999999/gift",
		adminToken, map[string]any{"coins": 100})
	if rec.Code != http.StatusNotFound || errorCode(t, rec) != "USER_NOT_FOUND" {
		t.Errorf("usuario inexistente: %d %s", rec.Code, rec.Body.String())
	}

	// Un id que no es un uuid tampoco puede salir como 500.
	rec = h.do(http.MethodPost, "/api/admin/users/no-soy-un-uuid/gift",
		adminToken, map[string]any{"coins": 100})
	if rec.Code != http.StatusNotFound {
		t.Errorf("id inválido: %d %s", rec.Code, rec.Body.String())
	}

	// El alumno ve los movimientos en su historial, con la nota del instructor.
	rec = h.do(http.MethodGet, "/api/me/transactions", anaToken, nil)
	var historial struct {
		Items []struct {
			Reason string `json:"reason"`
			Note   string `json:"note"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &historial); err != nil {
		t.Fatal(err)
	}
	if len(historial.Items) != 3 {
		t.Fatalf("el historial tiene %d movimientos, se esperaban 3", len(historial.Items))
	}
	if historial.Items[0].Reason != ledger.ReasonAdjustment {
		t.Errorf("el más nuevo es %q, se esperaba adjustment", historial.Items[0].Reason)
	}
	if historial.Items[0].Note != "participación en clase" {
		t.Errorf("nota = %q", historial.Items[0].Note)
	}

	testdb.Reconcile(t, h.ledger.Pool)
}

func TestGrantPointsSubeLaNotaSinTocarElSaldo(t *testing.T) {
	h := newHarness(t)
	admin := h.admin("profe")
	adminToken := h.token(admin.ID)
	ana, _ := h.student("anag")

	// 2,5 puntos = 250 centésimas.
	rec := h.do(http.MethodPost, "/api/admin/users/"+ana.ID+"/grant-points", adminToken,
		map[string]any{"points": 250, "reason": "Explicó @for en el code review"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("grant-points: %d %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Points json.RawMessage `json:"points"`
		Grant  struct {
			Points json.RawMessage `json:"points"`
			Reason string          `json:"reason"`
		} `json:"grant"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if string(body.Grant.Points) != "2.5" {
		t.Errorf("grant.points = %s, se esperaba 2.5", body.Grant.Points)
	}
	if string(body.Points) != "12.5" {
		t.Errorf("points = %s, se esperaba 12.5", body.Points)
	}

	// El saldo no se movió: los puntos no son monedas.
	if got := testdb.Balance(t, h.ledger.Pool, ana.ID); got != 1000 {
		t.Errorf("saldo = %d; un regalo de puntos no toca las monedas", got)
	}
	testdb.Reconcile(t, h.ledger.Pool)
}

// ── Varios ────────────────────────────────────────────────────────────────

func TestMeExigeSesion(t *testing.T) {
	h := newHarness(t)

	for _, path := range []string{"/api/me", "/api/me/transactions"} {
		rec := h.do(http.MethodGet, path, "", nil)
		if rec.Code != http.StatusUnauthorized || errorCode(t, rec) != "UNAUTHENTICATED" {
			t.Errorf("%s: %d %s", path, rec.Code, rec.Body.String())
		}
	}
}

// Un alumno no puede ver el historial de otro: no hay endpoint que lo permita, y
// /api/me/transactions siempre resuelve al dueño del token.
func TestElHistorialEsSoloElPropio(t *testing.T) {
	h := newHarness(t)
	_, anaToken := h.student("anag")
	bruno, _ := h.student("brunod")

	if _, err := h.ledger.Gift(t.Context(), h.admin("otro-profe").ID, bruno.ID, 500, "de bruno"); err != nil {
		t.Fatal(err)
	}

	rec := h.do(http.MethodGet, "/api/me/transactions", anaToken, nil)
	var historial struct {
		Items []struct {
			Delta int64 `json:"delta"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &historial); err != nil {
		t.Fatal(err)
	}
	for _, item := range historial.Items {
		if item.Delta == 500 {
			t.Error("el historial de ana tiene el regalo de bruno")
		}
	}
}

func TestHealthNoNecesitaSesion(t *testing.T) {
	h := newHarness(t)

	rec := h.do(http.MethodGet, "/health", "", nil)
	if rec.Code != http.StatusOK {
		t.Errorf("estado %d", rec.Code)
	}
}

// El CORS tiene que reflejar el origen concreto y permitir credenciales: el
// refresh viaja en cookie, y `*` con credenciales el navegador lo rechaza.
func TestCORSReflejaElOrigenYPermiteCredenciales(t *testing.T) {
	h := newHarness(t)

	req := httptest.NewRequest(http.MethodOptions, "/api/me", nil)
	req.Header.Set("Origin", "http://localhost:4200")
	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:4200" {
		t.Errorf("Allow-Origin = %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Allow-Credentials = %q", got)
	}

	// Un origen fuera de la lista no recibe la cabecera.
	req = httptest.NewRequest(http.MethodOptions, "/api/me", nil)
	req.Header.Set("Origin", "https://sitio-ajeno.example")
	rec = httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("un origen ajeno recibió Allow-Origin = %q", got)
	}
}

// El sobre de error tiene `message` en castellano: se le muestra tal cual al
// usuario. Los `code` van en inglés.
func TestElMensajeDeErrorEstaEnCastellano(t *testing.T) {
	h := newHarness(t)

	rec := h.do(http.MethodGet, "/api/me", "", nil)
	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != "UNAUTHENTICATED" {
		t.Errorf("code = %q", body.Error.Code)
	}
	if body.Error.Message != "Iniciá sesión para continuar." {
		t.Errorf("message = %q", body.Error.Message)
	}
}
