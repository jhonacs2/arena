package races

import (
	"context"
	"time"

	"github.com/talentodh/arena/internal/sim"
)

// Store es el acceso a datos del dominio de carreras.
//
// Es una interfaz y no un *pgxpool.Pool por dos razones. La primera es que los
// tests de apuestas y de cancelación corren contra un doble en memoria y no
// necesitan una base: si probar «cancelar devuelve exactamente lo apostado»
// costara levantar Postgres, ese test no se correría en cada cambio. La segunda
// es que el ledger lo implementa otro paquete, y la única forma de que el
// descuento y la apuesta caigan en la MISMA transacción sin que este paquete
// conozca el driver es que la transacción sea un parámetro.
//
// La implementación de producción es racesdb.Store.
type Store interface {
	// InTx corre fn dentro de una transacción y la confirma si fn no devuelve
	// error. Si fn devuelve error, se revierte entera.
	//
	// Todo lo que toca dinero pasa por acá. No puede quedar una apuesta creada
	// sin su descuento en el ledger, ni una carrera cancelada con la mitad de
	// las apuestas devueltas.
	InTx(ctx context.Context, fn func(Tx) error) error
}

// Tx son las operaciones disponibles dentro de una transacción.
//
// Las lecturas sueltas también entran por acá: una transacción de solo lectura
// es barata y tener un solo camino es más fácil de auditar que dos.
type Tx interface {
	// ── Carreras ──────────────────────────────────────────────────────────

	CreateRace(ctx context.Context, in NewRace) (Race, error)
	UpdateRace(ctx context.Context, raceID string, patch RacePatch) (Race, error)

	// Race lee la carrera sin bloquearla. Para leer y mostrar.
	Race(ctx context.Context, raceID string) (Race, error)

	// RaceForUpdate lee la carrera y la bloquea hasta el fin de la transacción
	// (SELECT … FOR UPDATE).
	//
	// TODA transición de estado y TODA apuesta entran por acá. Es lo que impide
	// que dos peticiones simultáneas lean el mismo `open` y las dos escriban:
	// sin el bloqueo, dos clics en «largar» arrancarían la carrera dos veces.
	RaceForUpdate(ctx context.Context, raceID string) (Race, error)

	SetStatus(ctx context.Context, raceID string, status Status, at time.Time) error

	// SetSeed guarda la semilla de la simulación. Se llama una sola vez, al
	// largar.
	SetSeed(ctx context.Context, raceID string, seed int64) error

	// VisibleRaces son las carreras del listado del alumno, con los agregados y
	// su propia apuesta ya resueltos. Con includeDraft, también las que el
	// instructor está armando.
	VisibleRaces(ctx context.Context, userID string, includeDraft bool) ([]Summary, error)

	// RunningRaces son las que quedaron en `running`. Las usa Resume al
	// arrancar el servidor.
	RunningRaces(ctx context.Context) ([]Race, error)

	// ── Caballos ──────────────────────────────────────────────────────────

	AddHorses(ctx context.Context, raceID string, in []NewHorse) ([]Horse, error)
	Horses(ctx context.Context, raceID string) ([]Horse, error)
	Horse(ctx context.Context, horseID string) (Horse, error)

	// ── Sala ──────────────────────────────────────────────────────────────

	// AddParticipant mete al alumno en la sala. Idempotente: entrar dos veces
	// no es un error, y devuelve si esta llamada lo agregó de verdad — es lo
	// que decide si se difunde `room.joined`.
	AddParticipant(ctx context.Context, raceID, userID string) (added bool, err error)
	Participants(ctx context.Context, raceID string) ([]Participant, error)

	// ── Apuestas ──────────────────────────────────────────────────────────

	Bets(ctx context.Context, raceID string) ([]Bet, error)
	BetByUser(ctx context.Context, raceID, userID string) (Bet, bool, error)

	// InsertBet inserta la apuesta y devuelve la fila guardada.
	//
	// Devuelve la fila y no solo un error porque el id lo genera la base
	// (gen_random_uuid) y hace falta para el `ref_id` del movimiento del ledger:
	// es lo que le permite al instructor seguir el rastro de un descuento hasta
	// la apuesta que lo causó.
	InsertBet(ctx context.Context, bet Bet) (Bet, error)

	// SettleBet marca la apuesta como won, lost o refunded y le pone el pago.
	SettleBet(ctx context.Context, betID string, status BetStatus, payout int64, at time.Time) error

	// ── Resultados ────────────────────────────────────────────────────────

	InsertResults(ctx context.Context, raceID string, results []sim.Result) error
	Results(ctx context.Context, raceID string) ([]sim.Result, error)

	// ── Saldo ─────────────────────────────────────────────────────────────

	Balance(ctx context.Context, userID string) (int64, error)

	// Move registra un movimiento en el ledger y devuelve el saldo resultante.
	//
	// Lo delega al paquete del ledger, que es append-only y mantiene
	// users.balance en la misma transacción. Devuelve error si el saldo quedaría
	// negativo: el servicio ya valida `amount <= balance` antes, así que si este
	// error aparece hubo una carrera de datos, y se loguea en vez de tratarse
	// como imposible.
	Move(ctx context.Context, m Movement) (balanceAfter int64, err error)

	User(ctx context.Context, userID string) (Participant, error)
}

// Broadcaster es lo que el dominio necesita del hub de WebSocket. Lo satisface
// *ws.Hub.
//
// Se declara acá para que el dominio no dependa del transporte: el hub reparte
// eventos y no sabe de carreras.
type Broadcaster interface {
	// ToRoom manda el mismo evento a todos los conectados a una carrera.
	ToRoom(raceID string, event any)

	// ToUser manda a todas las conexiones de un usuario. El saldo se actualiza
	// en todas sus pestañas, no solo en la que apostó.
	ToUser(userID string, event any)

	// ToRoomPerUser manda a cada conectado un evento armado PARA ÉL.
	//
	// Lo usan `race.finished` y `race.cancelled`: difundir el mismo objeto
	// filtraría cuánto cobró cada uno.
	ToRoomPerUser(raceID string, build func(userID string) any)
}
