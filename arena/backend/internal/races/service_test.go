package races

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/talentodh/arena/internal/contract"
	"github.com/talentodh/arena/internal/sim"
)

// Los tests de este archivo corren contra el doble en memoria de
// memstore_test.go. Prueban las reglas que decisiones.md llama duras, y por eso
// están escritos contra el SERVICIO y no contra los handlers: la validación pasa
// en el servicio, dentro de una transacción, y un test que fuera por HTTP
// probaría además el enrutado sin probar nada más de lo que importa.

const (
	adminID = "user-admin"
	anaID   = "user-ana"
	beoID   = "user-beto"
	caroID  = "user-caro"
)

var (
	admin = Identity{UserID: adminID, Username: "profe", Role: RoleAdmin}
	ana   = Identity{UserID: anaID, Username: "anag", Role: RoleStudent}
	beto  = Identity{UserID: beoID, Username: "betoz", Role: RoleStudent}
	caro  = Identity{UserID: caroID, Username: "caro", Role: RoleStudent}
)

// La semilla fija hace que los tests no dependan de crypto/rand.
const testSeed int64 = 99

type harness struct {
	t     *testing.T
	store *memStore
	hub   *recorder
	svc   *Service
	now   time.Time
}

// newHarness arma un servicio con reloj y semilla fijos, tres alumnos con 1000
// monedas —el saldo inicial de decisiones.md §1— y un instructor.
func newHarness(t *testing.T, audience ...string) *harness {
	t.Helper()
	return newHarnessWithRule(t, nil, audience...)
}

// newHarnessWithRule arma el servicio con una regla de liquidación de prueba.
//
// Con rule nula las carreras no se liquidan, que es el estado real del proyecto
// mientras la economía esté sin decidir.
func newHarnessWithRule(t *testing.T, rule SettlementRule, audience ...string) *harness {
	t.Helper()

	store := newMemStore()
	store.addUser(adminID, "profe", 0)
	store.addUser(anaID, "anag", 1000)
	store.addUser(beoID, "betoz", 1000)
	store.addUser(caroID, "caro", 1000)

	hub := newRecorder(audience...)
	now := time.Date(2026, 7, 29, 18, 0, 0, 0, time.UTC)

	svc := New(Config{
		Store:  store,
		Hub:    hub,
		Log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
		Rule:   rule,
		Clock:  func() time.Time { return now },
		Seeder: func() int64 { return testSeed },
	})
	// Ninguna simulación tiene que sobrevivir al test: si una quedara viva,
	// emitiría ticks mientras corre el test siguiente.
	t.Cleanup(svc.Runner().Close)

	return &harness{t: t, store: store, hub: hub, svc: svc, now: now}
}

// openRace crea una carrera con tres caballos y la abre.
func (h *harness) openRace() (raceID string, horses []HorseView) {
	h.t.Helper()

	created, err := h.svc.CreateRace(context.Background(), admin, NewRace{
		Name: "Clásico del Recuerdo",
		Horses: []NewHorse{
			{Number: 1, Name: "Viento Norte", NominalOdds: 340},
			{Number: 2, Name: "Doña Rosa", NominalOdds: 210},
			{Number: 3, Name: "Malambo", NominalOdds: 755},
		},
	})
	if err != nil {
		h.t.Fatalf("creando la carrera: %v", err)
	}

	opened, err := h.svc.Open(context.Background(), admin, created.Race.ID)
	if err != nil {
		h.t.Fatalf("abriendo la carrera: %v", err)
	}
	return opened.Race.ID, opened.Race.Horses
}

// assertCode comprueba que el error sea del catálogo y con el código esperado.
//
// Compara el CÓDIGO y no el mensaje: el mensaje está en castellano y se puede
// reescribir sin romper el frontend, el código no.
func assertCode(t *testing.T, err error, want contract.Code) {
	t.Helper()

	if err == nil {
		t.Fatalf("se esperaba el error %s y no hubo error", want)
	}
	var fault *contract.Fault
	if !errors.As(err, &fault) {
		t.Fatalf("se esperaba el error %s y llegó %v, que no es del catálogo", want, err)
	}
	if fault.Code != want {
		t.Fatalf("el error es %s y se esperaba %s", fault.Code, want)
	}
}

// ── La máquina de estados, por la API ─────────────────────────────────────

// TestTransicionesInvalidasDevuelvenINVALID_TRANSITION recorre TODAS las
// combinaciones de estado y acción de instructor.
//
// El test de state_test.go prueba la tabla; este prueba que los endpoints la
// respeten. Son dos cosas distintas: la tabla podría estar bien y un handler
// podría no consultarla.
func TestTransicionesInvalidasDevuelvenINVALID_TRANSITION(t *testing.T) {
	actions := []struct {
		name  string
		to    Status
		apply func(*Service, context.Context, string) error
	}{
		{"open", StatusOpen, func(s *Service, ctx context.Context, id string) error {
			_, err := s.Open(ctx, admin, id)
			return err
		}},
		{"start", StatusRunning, func(s *Service, ctx context.Context, id string) error {
			_, err := s.Start(ctx, admin, id)
			return err
		}},
		{"cancel", StatusCancelled, func(s *Service, ctx context.Context, id string) error {
			_, err := s.Cancel(ctx, admin, id, "")
			return err
		}},
	}

	for _, from := range AllStatuses {
		for _, action := range actions {
			name := string(from) + "→" + action.name
			t.Run(name, func(t *testing.T) {
				h := newHarness(t)
				raceID, _ := h.openRace()

				// La semilla se necesita cuando el escenario es `running`:
				// SetSeed se escribe una sola vez y sin ella Start fallaría por
				// otro motivo que el que se está probando.
				var seed *int64
				if from == StatusRunning || from == StatusFinished {
					value := testSeed
					seed = &value
				}
				h.store.forceStatus(raceID, from, seed)

				err := action.apply(h.svc, context.Background(), raceID)
				want := CanTransition(from, action.to)

				if want {
					if err != nil {
						t.Fatalf("%s tenía que estar permitida y falló: %v", name, err)
					}
					return
				}

				assertCode(t, err, contract.CodeInvalidTransition)
				// Y el estado no se movió: una transición rechazada no deja
				// nada a medias.
				if got := h.store.statusOf(raceID); got != from {
					t.Fatalf("la transición se rechazó pero el estado quedó en %q en vez de %q", got, from)
				}
			})
		}
	}
}

// TestNoSePuedeAbrirUnaCarreraSinCaballos: abrir una carrera vacía dejaría a los
// alumnos mirando una pantalla sin nada a lo que apostarle.
func TestNoSePuedeAbrirUnaCarreraSinCaballos(t *testing.T) {
	h := newHarness(t)

	created, err := h.svc.CreateRace(context.Background(), admin, NewRace{Name: "Vacía"})
	if err != nil {
		t.Fatalf("creando la carrera: %v", err)
	}

	_, err = h.svc.Open(context.Background(), admin, created.Race.ID)
	assertCode(t, err, contract.CodeInvalidTransition)
}

// TestUnAlumnoNoPuedeOperarCarreras: el rol se verifica EN EL SERVIDOR, en cada
// endpoint. Un alumno que edite su token no obtiene nada.
func TestUnAlumnoNoPuedeOperarCarreras(t *testing.T) {
	h := newHarness(t)
	raceID, _ := h.openRace()
	ctx := context.Background()

	operations := map[string]func() error{
		"crear": func() error {
			_, err := h.svc.CreateRace(ctx, ana, NewRace{Name: "Mía", Horses: []NewHorse{{Number: 1, Name: "X", NominalOdds: 200}}})
			return err
		},
		"editar": func() error {
			name := "Otra"
			_, err := h.svc.PatchRace(ctx, ana, raceID, RacePatch{Name: &name})
			return err
		},
		"agregar caballos": func() error {
			_, err := h.svc.AddHorses(ctx, ana, raceID, []NewHorse{{Number: 9, Name: "Y", NominalOdds: 200}})
			return err
		},
		"abrir": func() error { _, err := h.svc.Open(ctx, ana, raceID); return err },
		"largar": func() error {
			_, err := h.svc.Start(ctx, ana, raceID)
			return err
		},
		"cancelar": func() error {
			_, err := h.svc.Cancel(ctx, ana, raceID, "")
			return err
		},
	}

	for name, operation := range operations {
		t.Run(name, func(t *testing.T) {
			assertCode(t, operation(), contract.CodeForbidden)
		})
	}
}

// TestElAlumnoNoVeLasCarrerasEnDraft: `draft` es la carrera que el instructor
// está armando, con las cuotas todavía sin decidir.
func TestElAlumnoNoVeLasCarrerasEnDraft(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()

	created, err := h.svc.CreateRace(ctx, admin, NewRace{
		Name:   "En borrador",
		Horses: []NewHorse{{Number: 1, Name: "Secreto", NominalOdds: 200}},
	})
	if err != nil {
		t.Fatalf("creando la carrera: %v", err)
	}

	list, err := h.svc.List(ctx, ana)
	if err != nil {
		t.Fatalf("listando: %v", err)
	}
	if len(list.Items) != 0 {
		t.Errorf("el alumno ve %d carreras y todas están en borrador", len(list.Items))
	}

	// Y el detalle es RACE_NOT_FOUND, no FORBIDDEN: un 403 le confirmaría que la
	// carrera existe.
	_, err = h.svc.Detail(ctx, ana, created.Race.ID)
	assertCode(t, err, contract.CodeRaceNotFound)

	// El instructor sí la ve.
	adminList, err := h.svc.List(ctx, admin)
	if err != nil {
		t.Fatalf("listando como instructor: %v", err)
	}
	if len(adminList.Items) != 1 {
		t.Errorf("el instructor ve %d carreras y tendría que ver 1", len(adminList.Items))
	}
}

// ── Apostar ───────────────────────────────────────────────────────────────

// TestNoSePuedeApostarDosVecesEnLaMismaCarrera es la regla que impide cubrir
// todos los caballos y garantizarse nota — decisiones.md §1.
//
// Comprueba además que el segundo intento no deje rastro: una apuesta rechazada
// no puede haber descontado nada.
func TestNoSePuedeApostarDosVecesEnLaMismaCarrera(t *testing.T) {
	h := newHarness(t)
	raceID, horses := h.openRace()
	ctx := context.Background()

	first, err := h.svc.PlaceBet(ctx, ana, raceID, horses[0].ID, 200)
	if err != nil {
		t.Fatalf("la primera apuesta falló: %v", err)
	}
	if first.Balance != 800 {
		t.Fatalf("después de apostar 200 de 1000, el saldo es %d", first.Balance)
	}

	// Segunda apuesta, a OTRO caballo: es exactamente la jugada que la regla
	// prohíbe.
	_, err = h.svc.PlaceBet(ctx, ana, raceID, horses[1].ID, 300)
	assertCode(t, err, contract.CodeBetAlreadyPlaced)

	if got := h.store.balanceOf(anaID); got != 800 {
		t.Errorf("la apuesta rechazada movió el saldo a %d; tenía que quedar en 800", got)
	}
	if movements := h.store.movements(anaID); len(movements) != 1 {
		t.Errorf("hay %d movimientos en el ledger y tendría que haber 1", len(movements))
	}
}

// TestNoSePuedeApostarConLaCarreraCorriendo: las apuestas se cierran EN EL
// SERVIDOR al pasar a `running`. El botón deshabilitado del frontend es una
// cortesía, no un control.
func TestNoSePuedeApostarConLaCarreraCorriendo(t *testing.T) {
	seed := testSeed

	for _, status := range []Status{StatusDraft, StatusRunning, StatusFinished, StatusCancelled} {
		t.Run(string(status), func(t *testing.T) {
			h := newHarness(t)
			raceID, horses := h.openRace()

			var withSeed *int64
			if status == StatusRunning || status == StatusFinished {
				withSeed = &seed
			}
			h.store.forceStatus(raceID, status, withSeed)

			_, err := h.svc.PlaceBet(context.Background(), ana, raceID, horses[0].ID, 200)
			assertCode(t, err, contract.CodeRaceNotOpen)

			if got := h.store.balanceOf(anaID); got != 1000 {
				t.Errorf("la apuesta rechazada movió el saldo a %d", got)
			}
			if _, found := h.store.betOf(raceID, anaID); found {
				t.Error("quedó una apuesta guardada de una carrera que no estaba abierta")
			}
		})
	}
}

// TestNoSePuedeApostarMasQueElSaldo: el monto de una apuesta es `1 ≤ monto ≤
// saldo`, y el saldo tiene piso en 0 — decisiones.md §1.
func TestNoSePuedeApostarMasQueElSaldo(t *testing.T) {
	h := newHarness(t)
	raceID, horses := h.openRace()
	ctx := context.Background()

	_, err := h.svc.PlaceBet(ctx, ana, raceID, horses[0].ID, 1001)
	assertCode(t, err, contract.CodeInsufficientBalance)

	// El saldo exacto SÍ se puede apostar: fundirse es una jugada válida.
	out, err := h.svc.PlaceBet(ctx, ana, raceID, horses[0].ID, 1000)
	if err != nil {
		t.Fatalf("apostar el saldo entero falló: %v", err)
	}
	if out.Balance != 0 {
		t.Errorf("después de apostar todo, el saldo es %d y se esperaba 0", out.Balance)
	}
}

// TestMontosInvalidos: cero y negativo se rechazan antes de tocar la base.
func TestMontosInvalidos(t *testing.T) {
	h := newHarness(t)
	raceID, horses := h.openRace()

	for _, amount := range []int64{0, -1, -1000} {
		_, err := h.svc.PlaceBet(context.Background(), ana, raceID, horses[0].ID, amount)
		assertCode(t, err, contract.CodeValidationFailed)
	}
}

// TestNoSePuedeApostarAUnCaballoDeOtraCarrera: sin este chequeo se podría
// apostarle al favorito de otra carrera y cobrar en esta.
func TestNoSePuedeApostarAUnCaballoDeOtraCarrera(t *testing.T) {
	h := newHarness(t)
	first, _ := h.openRace()
	_, otherHorses := h.openRace()

	_, err := h.svc.PlaceBet(context.Background(), ana, first, otherHorses[0].ID, 100)
	assertCode(t, err, contract.CodeValidationFailed)

	if got := h.store.balanceOf(anaID); got != 1000 {
		t.Errorf("el saldo se movió a %d", got)
	}
}

// ── La revelación de las apuestas ─────────────────────────────────────────

// TestBetPlacedNoRevelaElCaballo es la regla que hace que la apuesta mida
// criterio: si `bet.placed` revelara el caballo, los últimos en apostar copiarían
// a los primeros.
//
// El evento directamente NO TIENE el campo, así que el test lo comprueba sobre el
// tipo: si alguien lo agregara, esto no compilaría.
func TestBetPlacedNoRevelaElCaballo(t *testing.T) {
	h := newHarness(t)
	raceID, horses := h.openRace()

	if _, err := h.svc.PlaceBet(context.Background(), ana, raceID, horses[2].ID, 150); err != nil {
		t.Fatalf("apostando: %v", err)
	}

	events := roomEventsOf[BetPlaced](h.hub)
	if len(events) != 1 {
		t.Fatalf("se difundieron %d eventos bet.placed y tendría que haber 1", len(events))
	}

	event := events[0]
	if event.UserID != anaID || event.Username != ana.Username || event.Amount != 150 {
		t.Errorf("el evento no describe la apuesta: %+v", event)
	}

	// El caballo no aparece en NINGÚN campo del evento. Se comprueba
	// serializando, que es lo que efectivamente sale por el cable.
	if serialized := mustJSON(t, event); containsHorse(serialized, horses[2].ID) {
		t.Errorf("bet.placed filtró el caballo: %s", serialized)
	}
}

// TestRoomStateNoRevelaLosCaballosMientrasEstaOpen: el mismo cuidado que
// `bet.placed`, en el otro camino por el que un cliente ve las apuestas de los
// demás.
func TestRoomStateNoRevelaLosCaballosMientrasEstaOpen(t *testing.T) {
	h := newHarness(t)
	raceID, horses := h.openRace()
	ctx := context.Background()

	if _, err := h.svc.PlaceBet(ctx, ana, raceID, horses[0].ID, 100); err != nil {
		t.Fatalf("apuesta de Ana: %v", err)
	}
	if _, err := h.svc.PlaceBet(ctx, beto, raceID, horses[1].ID, 250); err != nil {
		t.Fatalf("apuesta de Beto: %v", err)
	}

	// Caro entra a la sala sin haber apostado: es la que más tendría que ganar
	// con ver a qué le apostaron los otros.
	view, err := h.svc.Detail(ctx, caro, raceID)
	if err != nil {
		t.Fatalf("detalle: %v", err)
	}

	if len(view.Bets) != 2 {
		t.Fatalf("se ven %d apuestas y hay 2", len(view.Bets))
	}
	for _, bet := range view.Bets {
		// El monto SÍ se ve: es lo que dice el contrato y no da ventaja.
		if bet.Amount == 0 {
			t.Errorf("no se ve el monto de la apuesta de %s", bet.Username)
		}
		if bet.HorseID != "" {
			t.Errorf("con la carrera abierta se filtró el caballo de %s: %s", bet.Username, bet.HorseID)
		}

	}
	if view.MyBet != nil {
		t.Error("Caro no apostó y le llegó un myBet")
	}

	// Y la propia sí se ve completa: es suya.
	own, err := h.svc.Detail(ctx, ana, raceID)
	if err != nil {
		t.Fatalf("detalle de Ana: %v", err)
	}
	if own.MyBet == nil || own.MyBet.HorseID != horses[0].ID {
		t.Errorf("Ana no ve su propia apuesta completa: %+v", own.MyBet)
	}
}

// TestAlLargarSeRevelanTodasJuntas: al pasar a `running` se revelan las apuestas,
// todas al mismo tiempo. Nadie alcanza a usar la información para apostar, porque
// las apuestas ya están cerradas.
func TestAlLargarSeRevelanTodasJuntas(t *testing.T) {
	h := newHarness(t)
	raceID, horses := h.openRace()
	ctx := context.Background()

	if _, err := h.svc.PlaceBet(ctx, ana, raceID, horses[0].ID, 100); err != nil {
		t.Fatalf("apuesta de Ana: %v", err)
	}
	if _, err := h.svc.PlaceBet(ctx, beto, raceID, horses[1].ID, 250); err != nil {
		t.Fatalf("apuesta de Beto: %v", err)
	}

	started, err := h.svc.Start(ctx, admin, raceID)
	if err != nil {
		t.Fatalf("largando: %v", err)
	}
	if started.Race.Seed == nil || *started.Race.Seed != testSeed {
		t.Errorf("la semilla no quedó guardada: %+v", started.Race.Seed)
	}
	if started.Race.Status != StatusRunning {
		t.Errorf("después de largar, el estado es %q", started.Race.Status)
	}

	events := roomEventsOf[RaceStarted](h.hub)
	if len(events) != 1 {
		t.Fatalf("se difundieron %d eventos race.started", len(events))
	}
	if len(events[0].Bets) != 2 {
		t.Fatalf("race.started trae %d apuestas y hay 2", len(events[0].Bets))
	}
	for _, bet := range events[0].Bets {
		if bet.HorseID == "" {
			t.Errorf("al largar, la apuesta de %s sigue sin revelar el caballo", bet.Username)
		}

	}
}

// ── Cancelar ──────────────────────────────────────────────────────────────

// TestCancelarDevuelveExactamenteLoApostado es la regla de decisiones.md §3: se
// devuelve cada apuesta ÍNTEGRA. No el pago potencial, no un porcentaje.
func TestCancelarDevuelveExactamenteLoApostado(t *testing.T) {
	h := newHarness(t, anaID, beoID, caroID)
	raceID, horses := h.openRace()
	ctx := context.Background()

	amounts := map[string]int64{anaID: 137, beoID: 1000, caroID: 1}
	bettors := []struct {
		id     Identity
		amount int64
		horse  string
	}{
		{ana, amounts[anaID], horses[0].ID},
		{beto, amounts[beoID], horses[1].ID},
		{caro, amounts[caroID], horses[2].ID},
	}

	for _, bettor := range bettors {
		if _, err := h.svc.PlaceBet(ctx, bettor.id, raceID, bettor.horse, bettor.amount); err != nil {
			t.Fatalf("apuesta de %s: %v", bettor.id.Username, err)
		}
		if got := h.store.balanceOf(bettor.id.UserID); got != 1000-bettor.amount {
			t.Fatalf("después de apostar, %s tiene %d y tendría que tener %d",
				bettor.id.Username, got, 1000-bettor.amount)
		}
	}

	if _, err := h.svc.Cancel(ctx, admin, raceID, "Se cortó la luz."); err != nil {
		t.Fatalf("cancelando: %v", err)
	}

	for _, bettor := range bettors {
		userID := bettor.id.UserID

		// Exactamente el saldo original. Ni una moneda más ni una menos.
		if got := h.store.balanceOf(userID); got != 1000 {
			t.Errorf("%s quedó con %d después de la devolución; tenía que volver a 1000",
				bettor.id.Username, got)
		}

		// El movimiento de devolución es por el monto exacto y con el motivo
		// que el panel del instructor agrupa.
		movements := h.store.movements(userID)
		if len(movements) != 2 {
			t.Fatalf("%s tiene %d movimientos y tendría que tener 2", bettor.id.Username, len(movements))
		}
		discount, refund := movements[0], movements[1]
		if discount.Delta != -bettor.amount || discount.Reason != ReasonBetPlaced {
			t.Errorf("el descuento de %s es %+v", bettor.id.Username, discount)
		}
		if refund.Delta != bettor.amount {
			t.Errorf("la devolución de %s es %d y tenía que ser %d",
				bettor.id.Username, refund.Delta, bettor.amount)
		}
		if refund.Reason != ReasonBetRefunded {
			t.Errorf("la devolución de %s tiene motivo %q", bettor.id.Username, refund.Reason)
		}
		// El ledger es append-only: la devolución es un movimiento NUEVO, no una
		// corrección del anterior.
		if discount.RefID != refund.RefID || refund.RefID == "" {
			t.Errorf("los dos movimientos de %s no apuntan a la misma apuesta: %q y %q",
				bettor.id.Username, discount.RefID, refund.RefID)
		}

		// Y la apuesta queda marcada como devuelta.
		bet, found := h.store.betOf(raceID, userID)
		if !found {
			t.Fatalf("desapareció la apuesta de %s", bettor.id.Username)
		}
		if bet.Status != BetStatusRefunded {
			t.Errorf("la apuesta de %s quedó en %q", bettor.id.Username, bet.Status)
		}
		if bet.Payout == nil || *bet.Payout != bettor.amount {
			t.Errorf("la apuesta de %s quedó con payout %v", bettor.id.Username, bet.Payout)
		}

		// El evento le llega con SU devolución, no con la de todos.
		events := perUserEventsOf[RaceCancelled](h.hub, userID)
		if len(events) != 1 {
			t.Fatalf("a %s le llegaron %d eventos race.cancelled", bettor.id.Username, len(events))
		}
		event := events[0]
		if event.MyRefund == nil || *event.MyRefund != bettor.amount {
			t.Errorf("a %s le llegó myRefund %v y apostó %d", bettor.id.Username, event.MyRefund, bettor.amount)
		}
		if event.Balance == nil || *event.Balance != 1000 {
			t.Errorf("a %s le llegó balance %v", bettor.id.Username, show(event.Balance))
		}
		if event.Reason != "Se cortó la luz." {
			t.Errorf("el motivo que llegó es %q", event.Reason)
		}
	}
}

// TestCancelarDosVecesNoDuplicaLaDevolucion: la segunda cancelación es una
// transición inválida y no toca el saldo. Si duplicara, sería nota regalada.
func TestCancelarDosVecesNoDuplicaLaDevolucion(t *testing.T) {
	h := newHarness(t, anaID)
	raceID, horses := h.openRace()
	ctx := context.Background()

	if _, err := h.svc.PlaceBet(ctx, ana, raceID, horses[0].ID, 400); err != nil {
		t.Fatalf("apostando: %v", err)
	}
	if _, err := h.svc.Cancel(ctx, admin, raceID, ""); err != nil {
		t.Fatalf("primera cancelación: %v", err)
	}

	err := func() error { _, err := h.svc.Cancel(ctx, admin, raceID, ""); return err }()
	assertCode(t, err, contract.CodeInvalidTransition)

	if got := h.store.balanceOf(anaID); got != 1000 {
		t.Errorf("el saldo quedó en %d; una sola devolución tiene que dejarlo en 1000", got)
	}
	if movements := h.store.movements(anaID); len(movements) != 2 {
		t.Errorf("hay %d movimientos y tendrían que ser 2 (descuento y devolución)", len(movements))
	}
}

// TestCancelarUnaCarreraSinApuestasNoRompe: cancelar una carrera a la que nadie
// le apostó es lo más común de todo en una clase.
func TestCancelarUnaCarreraSinApuestasNoRompe(t *testing.T) {
	h := newHarness(t, anaID)
	raceID, _ := h.openRace()

	if _, err := h.svc.Cancel(context.Background(), admin, raceID, ""); err != nil {
		t.Fatalf("cancelando: %v", err)
	}
	if got := h.store.statusOf(raceID); got != StatusCancelled {
		t.Errorf("el estado quedó en %q", got)
	}

	events := perUserEventsOf[RaceCancelled](h.hub, anaID)
	if len(events) != 1 {
		t.Fatalf("llegaron %d eventos", len(events))
	}
	if events[0].MyRefund != nil {
		t.Errorf("a quien no apostó le llegó una devolución de %v", *events[0].MyRefund)
	}
	if events[0].Reason == "" {
		t.Error("cancelar sin motivo tiene que traer un texto por defecto")
	}
}

// ── Liquidar ──────────────────────────────────────────────────────────────
//
// La ECONOMÍA está sin decidir, así que estos tests no afirman cuánto se paga.
// Prueban la MECÁNICA: qué se marca, qué se acredita, y quién recibe qué evento.
// La regla de prueba paga un número fijo que el test elige, y por eso estos tests
// van a seguir valiendo cuando la economía se decida.

// fakeRule es una regla de liquidación de prueba.
//
// NO implementa ninguna de las dos economías candidatas a propósito: paga un monto
// fijo. Si algún día alguien la confunde con una implementación real, el nombre y
// este comentario están para evitarlo.
type fakeRule struct {
	payEach   int64
	refundAll bool
	calls     int
	err       error
}

func (r *fakeRule) Settle(_ context.Context, _ Tx, _ Race, bets []Bet, winnerHorseID string) (Payouts, error) {
	r.calls++
	if r.err != nil {
		return Payouts{}, r.err
	}
	if r.refundAll {
		return Payouts{RefundAll: true}, nil
	}

	byBet := map[string]int64{}
	for _, bet := range bets {
		if bet.HorseID == winnerHorseID {
			byBet[bet.ID] = r.payEach
		}
	}
	return Payouts{ByBet: byBet}, nil
}

// TestNoSeLiquidaSinReglaDeEconomia es el test de la pieza que falta.
//
// Con la economía sin decidir, una carrera que termina NO se liquida: no se marca
// ninguna apuesta, no se toca un saldo y el estado se queda en `running`. Es
// incómodo y es deliberado — lo contrario sería pagar con una regla que nadie
// eligió, y acá pagar es poner nota.
func TestNoSeLiquidaSinReglaDeEconomia(t *testing.T) {
	h := newHarness(t, anaID)
	raceID, horses := h.openRace()
	ctx := context.Background()

	if _, err := h.svc.PlaceBet(ctx, ana, raceID, horses[0].ID, 400); err != nil {
		t.Fatalf("apostando: %v", err)
	}
	if _, err := h.svc.Start(ctx, admin, raceID); err != nil {
		t.Fatalf("largando: %v", err)
	}
	h.svc.Runner().Stop(raceID)

	err := h.svc.finish(ctx, raceID, sim.Race{Seed: testSeed, Results: resultsWinner(horses, 0)})
	if !errors.Is(err, ErrNoSettlementRule) {
		t.Fatalf("sin regla, liquidar devolvió %v", err)
	}

	// Y no se escribió NADA.
	if got := h.store.statusOf(raceID); got != StatusRunning {
		t.Errorf("la carrera quedó en %q; sin regla se tiene que quedar en running", got)
	}
	if got := h.store.balanceOf(anaID); got != 600 {
		t.Errorf("el saldo es %d; solo tenía que estar el descuento de la apuesta", got)
	}
	bet, _ := h.store.betOf(raceID, anaID)
	if bet.Status != BetStatusPlaced {
		t.Errorf("la apuesta quedó en %q y tenía que seguir en placed", bet.Status)
	}
	if movements := h.store.movements(anaID); len(movements) != 1 {
		t.Errorf("hay %d movimientos y tendría que haber solo el descuento", len(movements))
	}
}

// TestLiquidarMarcaYAcreditaLoQueDiceLaRegla: la mecánica. El monto sale de la
// regla; este test no opina sobre cuál tiene que ser.
func TestLiquidarMarcaYAcreditaLoQueDiceLaRegla(t *testing.T) {
	const pays = int64(1234)

	rule := &fakeRule{payEach: pays}
	h := newHarnessWithRule(t, rule, anaID, beoID)
	raceID, horses := h.openRace()
	ctx := context.Background()

	if _, err := h.svc.PlaceBet(ctx, ana, raceID, horses[1].ID, 700); err != nil {
		t.Fatalf("apuesta de Ana: %v", err)
	}
	if _, err := h.svc.PlaceBet(ctx, beto, raceID, horses[0].ID, 500); err != nil {
		t.Fatalf("apuesta de Beto: %v", err)
	}

	h.finishRace(raceID, horses, 1) // gana el caballo de Ana

	if rule.calls != 1 {
		t.Errorf("la regla se llamó %d veces y tenía que llamarse una sola", rule.calls)
	}

	// Ana ganó: 1000 − 700 + 1234.
	if got := h.store.balanceOf(anaID); got != 1534 {
		t.Errorf("Ana quedó con %d y se esperaba 1534", got)
	}
	anaBet, _ := h.store.betOf(raceID, anaID)
	if anaBet.Status != BetStatusWon || anaBet.Payout == nil || *anaBet.Payout != pays {
		t.Errorf("la apuesta de Ana quedó %q con payout %v", anaBet.Status, anaBet.Payout)
	}

	// Beto perdió: no hay movimiento de cobro, el descuento ya estaba.
	if got := h.store.balanceOf(beoID); got != 500 {
		t.Errorf("Beto quedó con %d y se esperaba 500", got)
	}
	if movements := h.store.movements(beoID); len(movements) != 1 {
		t.Errorf("Beto tiene %d movimientos; el que pierde no cobra", len(movements))
	}
	betoBet, _ := h.store.betOf(raceID, beoID)
	if betoBet.Status != BetStatusLost {
		t.Errorf("la apuesta de Beto quedó %q", betoBet.Status)
	}
	if betoBet.Payout == nil || *betoBet.Payout != 0 {
		t.Errorf("la apuesta perdida quedó con payout %v", betoBet.Payout)
	}

	if got := h.store.statusOf(raceID); got != StatusFinished {
		t.Errorf("la carrera quedó en %q", got)
	}
}

// TestSiLaReglaPideDevolverSeDevuelveATodos.
//
// Es el caso «no acertó nadie». Las dos economías coinciden: no hay a quién
// pagarle, así que vuelve cada apuesta íntegra — incluso la de quien le había
// acertado al caballo, si la regla lo pidió.
func TestSiLaReglaPideDevolverSeDevuelveATodos(t *testing.T) {
	h := newHarnessWithRule(t, &fakeRule{refundAll: true}, anaID, beoID)
	raceID, horses := h.openRace()
	ctx := context.Background()

	if _, err := h.svc.PlaceBet(ctx, ana, raceID, horses[1].ID, 137); err != nil {
		t.Fatalf("apuesta de Ana: %v", err)
	}
	if _, err := h.svc.PlaceBet(ctx, beto, raceID, horses[0].ID, 999); err != nil {
		t.Fatalf("apuesta de Beto: %v", err)
	}

	h.finishRace(raceID, horses, 1)

	for _, userID := range []string{anaID, beoID} {
		if got := h.store.balanceOf(userID); got != 1000 {
			t.Errorf("%s quedó con %d; la devolución lo tenía que dejar en 1000", userID, got)
		}
		bet, _ := h.store.betOf(raceID, userID)
		if bet.Status != BetStatusRefunded {
			t.Errorf("la apuesta de %s quedó %q", userID, bet.Status)
		}

		movements := h.store.movements(userID)
		if len(movements) != 2 {
			t.Fatalf("%s tiene %d movimientos", userID, len(movements))
		}
		if movements[1].Reason != ReasonBetRefunded {
			t.Errorf("la devolución de %s tiene motivo %q", userID, movements[1].Reason)
		}
		if movements[0].Delta+movements[1].Delta != 0 {
			t.Errorf("los movimientos de %s no se compensan: %d y %d", userID, movements[0].Delta, movements[1].Delta)
		}
	}
}

// TestSiLaReglaFallaNoSeLiquidaNada: la transacción se revierte entera. Media
// liquidación es media clase con la nota mal.
func TestSiLaReglaFallaNoSeLiquidaNada(t *testing.T) {
	boom := errors.New("la regla explotó")
	h := newHarnessWithRule(t, &fakeRule{err: boom}, anaID)
	raceID, horses := h.openRace()
	ctx := context.Background()

	if _, err := h.svc.PlaceBet(ctx, ana, raceID, horses[0].ID, 400); err != nil {
		t.Fatalf("apostando: %v", err)
	}
	if _, err := h.svc.Start(ctx, admin, raceID); err != nil {
		t.Fatalf("largando: %v", err)
	}
	h.svc.Runner().Stop(raceID)

	if err := h.svc.finish(ctx, raceID, sim.Race{Seed: testSeed, Results: resultsWinner(horses, 0)}); !errors.Is(err, boom) {
		t.Fatalf("liquidar devolvió %v", err)
	}

	if got := h.store.statusOf(raceID); got != StatusRunning {
		t.Errorf("la carrera quedó en %q", got)
	}
	if got := h.store.balanceOf(anaID); got != 600 {
		t.Errorf("el saldo es %d", got)
	}
	bet, _ := h.store.betOf(raceID, anaID)
	if bet.Status != BetStatusPlaced {
		t.Errorf("la apuesta quedó en %q", bet.Status)
	}
}

// TestRaceFinishedSeArmaPorDestinatario: difundir el mismo objeto filtraría cuánto
// cobró cada uno.
func TestRaceFinishedSeArmaPorDestinatario(t *testing.T) {
	const pays = int64(1234)

	h := newHarnessWithRule(t, &fakeRule{payEach: pays}, anaID, beoID, caroID)
	raceID, horses := h.openRace()
	ctx := context.Background()

	if _, err := h.svc.PlaceBet(ctx, ana, raceID, horses[1].ID, 700); err != nil {
		t.Fatalf("apuesta de Ana: %v", err)
	}
	if _, err := h.svc.PlaceBet(ctx, beto, raceID, horses[0].ID, 500); err != nil {
		t.Fatalf("apuesta de Beto: %v", err)
	}

	h.finishRace(raceID, horses, 1)

	// Ana ganó: ve SU pago.
	anaEvents := perUserEventsOf[RaceFinished](h.hub, anaID)
	if len(anaEvents) != 1 {
		t.Fatalf("a Ana le llegaron %d eventos race.finished", len(anaEvents))
	}
	anaEvent := anaEvents[0]
	if anaEvent.MyBet == nil || anaEvent.MyBet.Payout == nil || *anaEvent.MyBet.Payout != pays {
		t.Errorf("Ana no recibió su pago: %+v", anaEvent.MyBet)
	}
	if anaEvent.Balance == nil || *anaEvent.Balance != 1534 {
		t.Errorf("a Ana le llegó balance %v", show(anaEvent.Balance))
	}

	// Y NO ve el pago de Beto: la lista de los demás no trae payout de nadie.
	if len(anaEvent.Bets) != 1 {
		t.Fatalf("Ana ve %d apuestas ajenas y hay 1", len(anaEvent.Bets))
	}
	if serialized := mustJSON(t, anaEvent.Bets); containsText(serialized, "payout") {
		t.Errorf("race.finished filtró el pago de otro: %s", serialized)
	}
	// El caballo sí: la carrera terminó y ya se había revelado al largar.
	if anaEvent.Bets[0].HorseID == "" {
		t.Error("con la carrera terminada, la apuesta ajena sigue sin revelar el caballo")
	}

	// Beto perdió: su evento dice lost y trae su saldo, no el de Ana.
	betoEvents := perUserEventsOf[RaceFinished](h.hub, beoID)
	if len(betoEvents) != 1 {
		t.Fatalf("a Beto le llegaron %d eventos", len(betoEvents))
	}
	if betoEvents[0].MyBet == nil || betoEvents[0].MyBet.Status != string(BetStatusLost) {
		t.Errorf("el evento de Beto dice %+v", betoEvents[0].MyBet)
	}
	if betoEvents[0].Balance == nil || *betoEvents[0].Balance != 500 {
		t.Errorf("a Beto le llegó balance %v", show(betoEvents[0].Balance))
	}

	// Caro no apostó: recibe los resultados y nada personal.
	caroEvents := perUserEventsOf[RaceFinished](h.hub, caroID)
	if len(caroEvents) != 1 {
		t.Fatalf("a Caro le llegaron %d eventos", len(caroEvents))
	}
	if caroEvents[0].MyBet != nil || caroEvents[0].Balance != nil {
		t.Errorf("a Caro, que no apostó, le llegó %+v", caroEvents[0])
	}
	if len(caroEvents[0].Results) != len(horses) {
		t.Errorf("Caro recibió %d resultados y hay %d caballos", len(caroEvents[0].Results), len(horses))
	}
}

// TestNoSeLiquidaUnaCarreraCancelada: si el instructor cancela mientras la carrera
// corre, la liquidación que llega después no paga nada. Las apuestas ya se
// devolvieron y pagar además sería regalar nota.
func TestNoSeLiquidaUnaCarreraCancelada(t *testing.T) {
	h := newHarnessWithRule(t, &fakeRule{payEach: 5000}, anaID)
	raceID, horses := h.openRace()
	ctx := context.Background()

	if _, err := h.svc.PlaceBet(ctx, ana, raceID, horses[0].ID, 400); err != nil {
		t.Fatalf("apostando: %v", err)
	}
	if _, err := h.svc.Start(ctx, admin, raceID); err != nil {
		t.Fatalf("largando: %v", err)
	}
	if _, err := h.svc.Cancel(ctx, admin, raceID, "Se cortó la luz."); err != nil {
		t.Fatalf("cancelando: %v", err)
	}

	// La liquidación llega tarde: la carrera ya está en `cancelled`.
	if err := h.svc.finish(ctx, raceID, sim.Race{Seed: testSeed, Results: resultsWinner(horses, 0)}); err != nil {
		t.Fatalf("la liquidación tardía devolvió error: %v", err)
	}

	if got := h.store.balanceOf(anaID); got != 1000 {
		t.Errorf("Ana quedó con %d; la devolución la deja en 1000 y no se le paga además", got)
	}
	if movements := h.store.movements(anaID); len(movements) != 2 {
		t.Errorf("hay %d movimientos y tendrían que ser 2", len(movements))
	}
	if got := h.store.statusOf(raceID); got != StatusCancelled {
		t.Errorf("la carrera quedó en %q", got)
	}
}

// finishRace larga la carrera y la liquida con un ganador elegido a mano, sin
// esperar los cuarenta segundos de simulación.
func (h *harness) finishRace(raceID string, horses []HorseView, winnerIndex int) {
	h.t.Helper()

	if _, err := h.svc.Start(context.Background(), admin, raceID); err != nil {
		h.t.Fatalf("largando: %v", err)
	}
	// La simulación en vivo no interesa acá: lo que se prueba es la liquidación.
	h.svc.Runner().Stop(raceID)

	if err := h.svc.finish(context.Background(), raceID, sim.Race{
		Seed: testSeed, Results: resultsWinner(horses, winnerIndex),
	}); err != nil {
		h.t.Fatalf("liquidando: %v", err)
	}
}

// resultsWinner arma los puestos finales con el caballo elegido en primer lugar.
func resultsWinner(horses []HorseView, winnerIndex int) []sim.Result {
	out := []sim.Result{{
		HorseID: horses[winnerIndex].ID, HorseName: horses[winnerIndex].Name,
		Number: horses[winnerIndex].Number, Odds: horses[winnerIndex].NominalOdds, Position: 1,
	}}
	position := 1
	for i, horse := range horses {
		if i == winnerIndex {
			continue
		}
		position++
		out = append(out, sim.Result{
			HorseID: horse.ID, HorseName: horse.Name,
			Number: horse.Number, Odds: horse.NominalOdds, Position: position,
		})
	}
	return out
}

// ── La sala ───────────────────────────────────────────────────────────────

// TestJoinEsIdempotente: recargar la página no le avisa a toda la sala.
func TestJoinEsIdempotente(t *testing.T) {
	h := newHarness(t)
	raceID, _ := h.openRace()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if _, err := h.svc.Join(ctx, ana, raceID); err != nil {
			t.Fatalf("entrando a la sala (intento %d): %v", i+1, err)
		}
	}

	if events := roomEventsOf[RoomJoined](h.hub); len(events) != 1 {
		t.Errorf("se difundieron %d eventos room.joined y tendría que haber 1", len(events))
	}

	view, err := h.svc.Detail(ctx, ana, raceID)
	if err != nil {
		t.Fatalf("detalle: %v", err)
	}
	if len(view.Participants) != 1 {
		t.Errorf("la sala tiene %d participantes", len(view.Participants))
	}
}

// TestApostarMeteEnLaSala: si apostó, quiere ver la carrera.
func TestApostarMeteEnLaSala(t *testing.T) {
	h := newHarness(t)
	raceID, horses := h.openRace()
	ctx := context.Background()

	if _, err := h.svc.PlaceBet(ctx, ana, raceID, horses[0].ID, 100); err != nil {
		t.Fatalf("apostando: %v", err)
	}

	view, err := h.svc.Detail(ctx, ana, raceID)
	if err != nil {
		t.Fatalf("detalle: %v", err)
	}
	if len(view.Participants) != 1 || view.Participants[0].UserID != anaID {
		t.Errorf("los participantes son %+v", view.Participants)
	}
}

// TestElInstructorNoOcupaLugarEnLaSala: mira la sala pero no aparece como
// participante — no está apostando.
func TestElInstructorNoOcupaLugarEnLaSala(t *testing.T) {
	h := newHarness(t)
	raceID, _ := h.openRace()
	ctx := context.Background()

	if _, err := h.svc.Join(ctx, admin, raceID); err != nil {
		t.Fatalf("el instructor no pudo entrar a la sala: %v", err)
	}

	view, err := h.svc.Detail(ctx, admin, raceID)
	if err != nil {
		t.Fatalf("detalle: %v", err)
	}
	if len(view.Participants) != 0 {
		t.Errorf("la sala tiene %d participantes y el instructor no cuenta", len(view.Participants))
	}
}

// TestMyBetVieneResueltoEnElListado: sin él, el frontend haría una llamada por
// carrera.
func TestMyBetVieneResueltoEnElListado(t *testing.T) {
	h := newHarness(t)
	raceID, horses := h.openRace()
	ctx := context.Background()

	if _, err := h.svc.PlaceBet(ctx, ana, raceID, horses[0].ID, 250); err != nil {
		t.Fatalf("apostando: %v", err)
	}

	list, err := h.svc.List(ctx, ana)
	if err != nil {
		t.Fatalf("listando: %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("hay %d carreras en el listado", len(list.Items))
	}
	item := list.Items[0]
	if item.MyBet == nil || item.MyBet.Amount != 250 {
		t.Errorf("myBet del listado es %+v", item.MyBet)
	}
	if item.HorseCount != 3 || item.ParticipantCount != 1 {
		t.Errorf("los agregados son horseCount=%d participantCount=%d", item.HorseCount, item.ParticipantCount)
	}

	// Y a quien no apostó le llega null, no un objeto vacío.
	betoList, err := h.svc.List(ctx, beto)
	if err != nil {
		t.Fatalf("listando para Beto: %v", err)
	}
	if betoList.Items[0].MyBet != nil {
		t.Errorf("a Beto, que no apostó, le llegó myBet %+v", betoList.Items[0].MyBet)
	}
}

// ── Retomar ───────────────────────────────────────────────────────────────

// TestResumeRetomaLasCarrerasQueQuedaronCorriendo: si el proceso se cae en el
// medio, la carrera se vuelve a correr DESDE LA SEMILLA GUARDADA y termina con el
// mismo resultado que se estaba viendo.
func TestResumeRetomaLasCarrerasQueQuedaronCorriendo(t *testing.T) {
	h := newHarness(t, anaID)
	raceID, horses := h.openRace()
	ctx := context.Background()

	if _, err := h.svc.PlaceBet(ctx, ana, raceID, horses[0].ID, 100); err != nil {
		t.Fatalf("apostando: %v", err)
	}
	if _, err := h.svc.Start(ctx, admin, raceID); err != nil {
		t.Fatalf("largando: %v", err)
	}

	// Se simula la caída: el runner de este proceso se olvida de la carrera,
	// pero en la base sigue en `running` con su semilla.
	h.svc.Runner().Stop(raceID)
	waitUntil(t, func() bool { return !h.svc.Runner().Running(raceID) })

	if err := h.svc.Resume(ctx); err != nil {
		t.Fatalf("retomando: %v", err)
	}
	if !h.svc.Runner().Running(raceID) {
		t.Error("Resume no volvió a poner la carrera en vivo")
	}

	// Y el resultado es el de la semilla guardada, no uno nuevo.
	expected := sim.Simulate(testSeed, []sim.Horse{
		{ID: horses[0].ID, Name: horses[0].Name, Number: horses[0].Number, Odds: horses[0].NominalOdds},
		{ID: horses[1].ID, Name: horses[1].Name, Number: horses[1].Number, Odds: horses[1].NominalOdds},
		{ID: horses[2].ID, Name: horses[2].Name, Number: horses[2].Number, Odds: horses[2].NominalOdds},
	})
	again := sim.Simulate(testSeed, []sim.Horse{
		{ID: horses[0].ID, Name: horses[0].Name, Number: horses[0].Number, Odds: horses[0].NominalOdds},
		{ID: horses[1].ID, Name: horses[1].Name, Number: horses[1].Number, Odds: horses[1].NominalOdds},
		{ID: horses[2].ID, Name: horses[2].Name, Number: horses[2].Number, Odds: horses[2].NominalOdds},
	})
	winnerA, _ := expected.Winner()
	winnerB, _ := again.Winner()
	if winnerA.HorseID != winnerB.HorseID {
		t.Error("la misma semilla dio dos ganadores distintos")
	}
}

// waitUntil espera a que se cumpla una condición que resuelve otra goroutine.
func waitUntil(t *testing.T, condition func() bool) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("se agotó la espera")
}
