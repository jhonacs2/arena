package racesdb_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/talentodh/arena/internal/ledger"
	"github.com/talentodh/arena/internal/races"
	"github.com/talentodh/arena/internal/racesdb"
	"github.com/talentodh/arena/internal/sim"
	"github.com/talentodh/arena/internal/testdb"
)

// Estos tests corren contra Postgres de verdad. Los de internal/races prueban las
// REGLAS contra un doble en memoria; estos prueban que el SQL exista y haga lo que
// dice, que es lo único que un doble no puede probar:
//
//   - que las consultas compilen contra el esquema real;
//   - que `select … for update` serialice de verdad dos apuestas simultáneas;
//   - que `unique (race_id, user_id)` rechace la segunda apuesta aunque el
//     chequeo de Go fallara;
//   - que el descuento del ledger y la apuesta se confirmen o se reviertan JUNTOS.
//
// Sin ARENA_TEST_DATABASE_URL se saltean — ver internal/testdb.

// store arma el Store con el ledger de verdad. El adaptador es el mismo que va en
// el cableado, así que este test también prueba que el pegamento funcione.
func store(pool *pgxpool.Pool) *racesdb.Store {
	ledgerStore := &ledger.Store{Pool: pool}
	return racesdb.New(pool, racesdb.LedgerFunc(
		func(ctx context.Context, tx pgx.Tx, m racesdb.Movement) (int64, error) {
			return ledgerStore.MoveTx(ctx, tx, ledger.Movement(m))
		}))
}

// fund acredita monedas con el ledger, no con un UPDATE a mano: un saldo puesto a
// mano sin movimiento detrás es exactamente el estado que reconcile.sql busca.
func fund(t *testing.T, pool *pgxpool.Pool, userID string, coins int64) {
	t.Helper()

	ledgerStore := &ledger.Store{Pool: pool}
	if _, err := ledgerStore.Move(context.Background(), ledger.Movement{
		UserID: userID, Delta: coins, Reason: "code_redeemed", RefID: "TEST-9999",
	}); err != nil {
		t.Fatalf("acreditando %d monedas a %s: %v", coins, userID, err)
	}
}

// scenario arma una carrera abierta con tres caballos y devuelve todo lo que hace
// falta para apostar.
type scenario struct {
	pool   *pgxpool.Pool
	store  *racesdb.Store
	admin  string
	raceID string
	horses []races.Horse
}

func setup(t *testing.T) *scenario {
	t.Helper()

	pool := testdb.Pool(t)
	st := store(pool)
	admin := testdb.InsertUser(t, pool, "profe", "admin")
	ctx := context.Background()

	var (
		raceID string
		horses []races.Horse
	)
	err := st.InTx(ctx, func(tx races.Tx) error {
		race, err := tx.CreateRace(ctx, races.NewRace{
			Name:      "Clásico del Recuerdo",
			CreatedBy: admin,
			Horses: []races.NewHorse{
				{Number: 1, Name: "Viento Norte", NominalOdds: 340},
				{Number: 2, Name: "Doña Rosa", NominalOdds: 210},
				{Number: 3, Name: "Malambo", NominalOdds: 755},
			},
		})
		if err != nil {
			return err
		}
		if err := tx.SetStatus(ctx, race.ID, races.StatusOpen, time.Now()); err != nil {
			return err
		}
		raceID = race.ID
		horses, err = tx.Horses(ctx, race.ID)
		return err
	})
	if err != nil {
		t.Fatalf("armando el escenario: %v", err)
	}
	if len(horses) != 3 {
		t.Fatalf("se crearon %d caballos y tendrían que ser 3", len(horses))
	}

	return &scenario{pool: pool, store: st, admin: admin, raceID: raceID, horses: horses}
}

// TestElCicloCompletoContraPostgres recorre crear, abrir, apostar, largar y
// liquidar. Es el test que dice que todas las consultas existen y encajan con el
// esquema.
func TestElCicloCompletoContraPostgres(t *testing.T) {
	s := setup(t)
	ctx := context.Background()

	ana := testdb.InsertUser(t, s.pool, "anag", "student")
	fund(t, s.pool, ana, 1000)

	// Apostar: la apuesta y el descuento en la misma transacción.
	var placed races.Bet
	err := s.store.InTx(ctx, func(tx races.Tx) error {
		race, err := tx.RaceForUpdate(ctx, s.raceID)
		if err != nil {
			return err
		}
		if !race.Status.AcceptsBets() {
			t.Fatalf("la carrera quedó en %q", race.Status)
		}

		balance, err := tx.Balance(ctx, ana)
		if err != nil {
			return err
		}
		if balance != 1000 {
			t.Fatalf("el saldo es %d y se esperaba 1000", balance)
		}

		placed, err = tx.InsertBet(ctx, races.Bet{
			RaceID: s.raceID, UserID: ana, HorseID: s.horses[1].ID,
			Amount: 700, Status: races.BetStatusPlaced,
		})
		if err != nil {
			return err
		}

		after, err := tx.Move(ctx, races.Movement{
			UserID: ana, Delta: -700, Reason: races.ReasonBetPlaced, RefID: placed.ID,
		})
		if err != nil {
			return err
		}
		if after != 300 {
			t.Errorf("después de apostar 700, el saldo es %d", after)
		}

		_, err = tx.AddParticipant(ctx, s.raceID, ana)
		return err
	})
	if err != nil {
		t.Fatalf("apostando: %v", err)
	}

	// El username vino del join: los eventos del socket lo necesitan y pedirlo por
	// separado sería una consulta por apuesta.
	if placed.Username != "anag" {
		t.Errorf("la apuesta guardada es %+v", placed)
	}
	if testdb.Balance(t, s.pool, ana) != 300 {
		t.Errorf("el saldo en la base es %d", testdb.Balance(t, s.pool, ana))
	}

	// Largar: semilla y estado.
	const seed int64 = 4815162342
	err = s.store.InTx(ctx, func(tx races.Tx) error {
		if err := tx.SetSeed(ctx, s.raceID, seed); err != nil {
			return err
		}
		return tx.SetStatus(ctx, s.raceID, races.StatusRunning, time.Now())
	})
	if err != nil {
		t.Fatalf("largando: %v", err)
	}

	err = s.store.InTx(ctx, func(tx races.Tx) error {
		race, err := tx.Race(ctx, s.raceID)
		if err != nil {
			return err
		}
		if race.Status != races.StatusRunning {
			t.Errorf("el estado es %q", race.Status)
		}
		if race.Seed == nil || *race.Seed != seed {
			t.Errorf("la semilla guardada es %v", race.Seed)
		}
		if race.StartedAt == nil {
			t.Error("no se selló started_at")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("leyendo la carrera: %v", err)
	}

	// La semilla se escribe UNA sola vez.
	err = s.store.InTx(ctx, func(tx races.Tx) error {
		return tx.SetSeed(ctx, s.raceID, seed+1)
	})
	if err == nil {
		t.Error("se pudo sobrescribir la semilla de una carrera ya largada")
	}

	// Liquidar: resultados, estado y pago.
	//
	// El MONTO es arbitrario a propósito: la economía está sin decidir y este test
	// prueba que el SQL de la liquidación funcione, no cuánto se paga. Ver
	// races.SettlementRule.
	const payout = int64(1470)

	finished := time.Now()
	err = s.store.InTx(ctx, func(tx races.Tx) error {
		results := resultsWinner(s.horses, 1)
		if err := tx.InsertResults(ctx, s.raceID, results); err != nil {
			return err
		}
		if err := tx.SetStatus(ctx, s.raceID, races.StatusFinished, finished); err != nil {
			return err
		}

		after, err := tx.Move(ctx, races.Movement{
			UserID: ana, Delta: payout, Reason: races.ReasonBetWon, RefID: placed.ID,
		})
		if err != nil {
			return err
		}
		if after != 1770 {
			t.Errorf("después de cobrar, el saldo es %d y se esperaba 1770", after)
		}
		return tx.SettleBet(ctx, placed.ID, races.BetStatusWon, payout, finished)
	})
	if err != nil {
		t.Fatalf("liquidando: %v", err)
	}

	// Y todo quedó como corresponde.
	err = s.store.InTx(ctx, func(tx races.Tx) error {
		results, err := tx.Results(ctx, s.raceID)
		if err != nil {
			return err
		}
		if len(results) != 3 || results[0].Position != 1 || results[0].HorseID != s.horses[1].ID {
			t.Errorf("los resultados guardados son %+v", results)
		}

		bet, found, err := tx.BetByUser(ctx, s.raceID, ana)
		if err != nil {
			return err
		}
		if !found || bet.Status != races.BetStatusWon || bet.Payout == nil || *bet.Payout != 1470 {
			t.Errorf("la apuesta liquidada es %+v", bet)
		}
		if bet.SettledAt == nil {
			t.Error("no se selló settled_at")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("verificando: %v", err)
	}

	// La liquidación no se puede repetir: es la guarda contra el pago duplicado.
	err = s.store.InTx(ctx, func(tx races.Tx) error {
		return tx.SettleBet(ctx, placed.ID, races.BetStatusWon, payout, finished)
	})
	if err == nil {
		t.Error("se pudo liquidar dos veces la misma apuesta")
	}

	// El ledger cierra contra users.balance.
	testdb.Reconcile(t, s.pool)
}

// TestLaRestriccionDeLaBaseRechazaLaSegundaApuesta.
//
// El servicio ya chequea «una apuesta por carrera» antes de insertar. Este test
// prueba el CINTURÓN DE SEGURIDAD: si ese chequeo fallara —por un bug, o por dos
// peticiones que se cruzan— la base tiene que rechazarla igual. Es la regla que
// impide cubrir todos los caballos y garantizarse nota, así que tiene dos
// candados y no uno.
func TestLaRestriccionDeLaBaseRechazaLaSegundaApuesta(t *testing.T) {
	s := setup(t)
	ctx := context.Background()

	ana := testdb.InsertUser(t, s.pool, "anag", "student")
	fund(t, s.pool, ana, 1000)

	insert := func(horseID string, amount int64) error {
		return s.store.InTx(ctx, func(tx races.Tx) error {
			bet, err := tx.InsertBet(ctx, races.Bet{
				RaceID: s.raceID, UserID: ana, HorseID: horseID,
				Amount: amount, Status: races.BetStatusPlaced,
			})
			if err != nil {
				return err
			}
			_, err = tx.Move(ctx, races.Movement{
				UserID: ana, Delta: -amount, Reason: races.ReasonBetPlaced, RefID: bet.ID,
			})
			return err
		})
	}

	if err := insert(s.horses[0].ID, 100); err != nil {
		t.Fatalf("la primera apuesta falló: %v", err)
	}
	if err := insert(s.horses[1].ID, 200); err == nil {
		t.Fatal("la base aceptó una segunda apuesta del mismo alumno en la misma carrera")
	}

	// Y la transacción de la segunda se revirtió ENTERA: no quedó un descuento
	// sin apuesta.
	if got := testdb.Balance(t, s.pool, ana); got != 900 {
		t.Errorf("el saldo es %d; la apuesta rechazada tenía que dejarlo en 900", got)
	}
	testdb.Reconcile(t, s.pool)
}

// TestElPisoDelSaldoLoImponeLaBase: el `check (balance >= 0)`. El servicio valida
// `amount <= balance`, pero el ledger no confía en su llamador y la base tampoco.
func TestElPisoDelSaldoLoImponeLaBase(t *testing.T) {
	s := setup(t)
	ctx := context.Background()

	ana := testdb.InsertUser(t, s.pool, "anag", "student")
	fund(t, s.pool, ana, 100)

	err := s.store.InTx(ctx, func(tx races.Tx) error {
		bet, err := tx.InsertBet(ctx, races.Bet{
			RaceID: s.raceID, UserID: ana, HorseID: s.horses[0].ID,
			Amount: 500, Status: races.BetStatusPlaced,
		})
		if err != nil {
			return err
		}
		_, err = tx.Move(ctx, races.Movement{
			UserID: ana, Delta: -500, Reason: races.ReasonBetPlaced, RefID: bet.ID,
		})
		return err
	})
	if err == nil {
		t.Fatal("se pudo descontar más de lo que había en el saldo")
	}

	if got := testdb.Balance(t, s.pool, ana); got != 100 {
		t.Errorf("el saldo es %d y tenía que quedar en 100", got)
	}
	// Y no quedó la apuesta: la transacción se revirtió entera.
	err = s.store.InTx(ctx, func(tx races.Tx) error {
		if _, found, err := tx.BetByUser(ctx, s.raceID, ana); err != nil {
			return err
		} else if found {
			t.Error("quedó una apuesta cuyo descuento se revirtió")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("verificando: %v", err)
	}
	testdb.Reconcile(t, s.pool)
}

// TestVisibleRacesNoDevuelveDraft es la regla de visibilidad en el WHERE, no
// filtrada en Go: no puede haber un camino por el que una carrera en borrador
// llegue al alumno.
func TestVisibleRacesNoDevuelveDraft(t *testing.T) {
	s := setup(t)
	ctx := context.Background()

	ana := testdb.InsertUser(t, s.pool, "anag", "student")
	fund(t, s.pool, ana, 1000)

	// Una segunda carrera que se queda en `draft`.
	err := s.store.InTx(ctx, func(tx races.Tx) error {
		_, err := tx.CreateRace(ctx, races.NewRace{
			Name: "En borrador", CreatedBy: s.admin,
			Horses: []races.NewHorse{{Number: 1, Name: "Secreto", NominalOdds: 200}},
		})
		return err
	})
	if err != nil {
		t.Fatalf("creando la carrera en borrador: %v", err)
	}

	err = s.store.InTx(ctx, func(tx races.Tx) error {
		visible, err := tx.VisibleRaces(ctx, ana, false)
		if err != nil {
			return err
		}
		if len(visible) != 1 {
			t.Fatalf("el alumno ve %d carreras y tendría que ver 1", len(visible))
		}
		if visible[0].Status == races.StatusDraft {
			t.Error("el alumno ve una carrera en borrador")
		}
		if visible[0].HorseCount != 3 {
			t.Errorf("horseCount es %d y se esperaba 3", visible[0].HorseCount)
		}
		if visible[0].MyBet != nil {
			t.Errorf("sin apostar, myBet es %+v", visible[0].MyBet)
		}

		forAdmin, err := tx.VisibleRaces(ctx, s.admin, true)
		if err != nil {
			return err
		}
		if len(forAdmin) != 2 {
			t.Errorf("el instructor ve %d carreras y tendría que ver 2", len(forAdmin))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("listando: %v", err)
	}

	// Con apuesta, myBet viene resuelto en el listado.
	err = s.store.InTx(ctx, func(tx races.Tx) error {
		bet, err := tx.InsertBet(ctx, races.Bet{
			RaceID: s.raceID, UserID: ana, HorseID: s.horses[2].ID,
			Amount: 250, Status: races.BetStatusPlaced,
		})
		if err != nil {
			return err
		}
		_, err = tx.Move(ctx, races.Movement{
			UserID: ana, Delta: -250, Reason: races.ReasonBetPlaced, RefID: bet.ID,
		})
		return err
	})
	if err != nil {
		t.Fatalf("apostando: %v", err)
	}

	err = s.store.InTx(ctx, func(tx races.Tx) error {
		visible, err := tx.VisibleRaces(ctx, ana, false)
		if err != nil {
			return err
		}
		if visible[0].MyBet == nil {
			t.Fatal("myBet no vino resuelto en el listado")
		}
		if visible[0].MyBet.Amount != 250 {
			t.Errorf("myBet del listado es %+v", visible[0].MyBet)
		}
		if visible[0].ParticipantCount != 0 {
			t.Errorf("participantCount es %d", visible[0].ParticipantCount)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("listando con apuesta: %v", err)
	}
	testdb.Reconcile(t, s.pool)
}

// TestUnIdBasuraEsNoExiste, Y NO ENVENENA LA TRANSACCIÓN.
//
// El cliente puede mandar cualquier cosa en la ruta. Un «no existe» es la
// respuesta correcta.
//
// La segunda mitad del test es la que importa y la que encontró un bug de verdad:
// si el id basura llega a Postgres, Postgres ABORTA LA TRANSACCIÓN ENTERA (SQLSTATE
// 25P02) y a partir de ahí toda consulta siguiente falla, incluido el COMMIT. Una
// petición con un id mal formado pasaría de un 404 limpio a un 500 y se llevaría
// con ella cualquier otra lectura de la misma transacción. Por eso la forma del id
// se valida en Go ANTES de consultar.
func TestUnIdBasuraEsNoExiste(t *testing.T) {
	s := setup(t)
	ctx := context.Background()

	basura := []string{
		"no-soy-un-uuid",
		"",
		"12345",
		// La longitud correcta pero con un carácter que no es hexadecimal: es el
		// caso que una validación por largo dejaría pasar.
		"0f1a2b3c-0000-4000-8000-00000000000z",
		"0f1a2b3c00004000800000000000000000000",
	}

	err := s.store.InTx(ctx, func(tx races.Tx) error {
		for _, id := range basura {
			if _, err := tx.Race(ctx, id); !errors.Is(err, races.ErrNotFound) {
				t.Errorf("Race(%q) devolvió %v", id, err)
			}
			if _, err := tx.Horse(ctx, id); !errors.Is(err, races.ErrNotFound) {
				t.Errorf("Horse(%q) devolvió %v", id, err)
			}
		}

		// Un uuid con forma válida que no existe: también es «no existe», y este
		// sí llega a Postgres.
		if _, err := tx.Race(ctx, "00000000-0000-4000-8000-000000000000"); !errors.Is(err, races.ErrNotFound) {
			t.Errorf("Race con un uuid inexistente devolvió %v", err)
		}

		// Y DESPUÉS de todo eso la transacción sigue viva: se puede leer la
		// carrera que sí existe.
		race, err := tx.Race(ctx, s.raceID)
		if err != nil {
			return err
		}
		if race.ID != s.raceID {
			t.Errorf("después de los ids basura, la carrera leída es %q", race.ID)
		}
		return nil
	})
	// Y el COMMIT tiene que funcionar. Si la transacción hubiera quedado abortada,
	// esto devolvería «commit unexpectedly resulted in rollback».
	if err != nil {
		t.Fatalf("la transacción quedó envenenada por un id basura: %v", err)
	}
}

// TestAddParticipantEsIdempotente: recargar la página no le avisa a toda la sala.
func TestAddParticipantEsIdempotente(t *testing.T) {
	s := setup(t)
	ctx := context.Background()

	ana := testdb.InsertUser(t, s.pool, "anag", "student")

	err := s.store.InTx(ctx, func(tx races.Tx) error {
		added, err := tx.AddParticipant(ctx, s.raceID, ana)
		if err != nil {
			return err
		}
		if !added {
			t.Error("la primera vez tenía que agregarlo")
		}

		added, err = tx.AddParticipant(ctx, s.raceID, ana)
		if err != nil {
			return err
		}
		if added {
			t.Error("la segunda vez dijo que lo agregó de nuevo")
		}

		participants, err := tx.Participants(ctx, s.raceID)
		if err != nil {
			return err
		}
		if len(participants) != 1 || participants[0].Username != "anag" {
			t.Errorf("los participantes son %+v", participants)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("entrando a la sala: %v", err)
	}
}

// TestForUpdateSerializaDosApuestas.
//
// Es el test del `select … for update`. Dos transacciones intentan apostar por el
// mismo alumno al mismo tiempo; la segunda tiene que esperar a que la primera
// termine, y después ver el mundo que la primera dejó. Sin el bloqueo, las dos
// leerían el mismo saldo y lo gastarían dos veces.
func TestForUpdateSerializaDosApuestas(t *testing.T) {
	s := setup(t)
	ctx := context.Background()

	ana := testdb.InsertUser(t, s.pool, "anag", "student")
	fund(t, s.pool, ana, 1000)

	// La primera transacción se queda con el bloqueo y avisa.
	locked := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 2)

	go func() {
		done <- s.store.InTx(ctx, func(tx races.Tx) error {
			if _, err := tx.RaceForUpdate(ctx, s.raceID); err != nil {
				return err
			}
			close(locked)
			<-release // se mantiene el bloqueo hasta que el test lo suelte
			return nil
		})
	}()

	<-locked

	// La segunda intenta tomar el mismo bloqueo con un plazo corto: tiene que
	// vencer, porque la primera lo tiene.
	blocked, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()

	err := s.store.InTx(blocked, func(tx races.Tx) error {
		_, err := tx.RaceForUpdate(blocked, s.raceID)
		return err
	})
	if err == nil {
		t.Error("la segunda transacción tomó el bloqueo mientras la primera lo tenía")
	}

	close(release)
	if err := <-done; err != nil {
		t.Fatalf("la primera transacción falló: %v", err)
	}

	// Y una vez suelto, se puede tomar sin problema.
	if err := s.store.InTx(ctx, func(tx races.Tx) error {
		_, err := tx.RaceForUpdate(ctx, s.raceID)
		return err
	}); err != nil {
		t.Fatalf("tomando el bloqueo ya liberado: %v", err)
	}
}

// resultsWinner arma los puestos finales con el caballo elegido en primer lugar.
func resultsWinner(horses []races.Horse, winnerIndex int) []sim.Result {
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
