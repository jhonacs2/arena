// Package races es el dominio de las carreras: la máquina de estados, las
// apuestas, la simulación en vivo y la liquidación.
//
// Implementa docs/contract/decisiones.md §3 y §4 y docs/contract/api.md. Sigue
// la forma del backend del hipódromo (project/backend) a propósito: los dos
// backends se leen igual.
//
// El paquete no sabe de Postgres ni de WebSocket. Habla con la base por la
// interfaz Store y con el socket por la interfaz Broadcaster, las dos
// declaradas acá — es el consumidor el que declara lo que necesita. La
// implementación de Store para Postgres está en internal/racesdb.
package races

import (
	"errors"
	"time"
)

// ErrNotFound lo devuelve el Store cuando la fila no existe. El servicio lo
// traduce al código del contrato; el Store no sabe de HTTP.
var ErrNotFound = errors.New("no existe")

// ── Estados ───────────────────────────────────────────────────────────────

// Status es el estado de una carrera. Coincide con el enum race_status del
// esquema.
type Status string

const (
	StatusDraft     Status = "draft"
	StatusOpen      Status = "open"
	StatusRunning   Status = "running"
	StatusFinished  Status = "finished"
	StatusCancelled Status = "cancelled"
)

// BetStatus es el estado de una apuesta. Coincide con el enum bet_status.
type BetStatus string

const (
	BetStatusPlaced   BetStatus = "placed"
	BetStatusWon      BetStatus = "won"
	BetStatusLost     BetStatus = "lost"
	BetStatusRefunded BetStatus = "refunded"
)

// Reason es el motivo de un movimiento del ledger. Coincide con el enum
// ledger_reason. Acá solo se usan los tres de una apuesta: el resto (regalos,
// canje del código, ajustes) los mueve el paquete que administra el ledger.
type Reason string

const (
	ReasonBetPlaced   Reason = "bet_placed"
	ReasonBetWon      Reason = "bet_won"
	ReasonBetRefunded Reason = "bet_refunded"
)

// ── Filas ─────────────────────────────────────────────────────────────────

// Race es una carrera. Los punteros son las columnas que empiezan nulas y se
// llenan al pasar de estado: no hay un cero que signifique «todavía no».
type Race struct {
	ID          string
	Name        string
	Status      Status
	ScheduledAt *time.Time

	CreatedBy  string
	CreatedAt  time.Time
	OpenedAt   *time.Time
	StartedAt  *time.Time
	FinishedAt *time.Time

	// Seed es la semilla de la simulación. Se fija al largar y con ella la
	// carrera es reproducible: si alguien reclama el resultado, se vuelve a
	// correr igual. Nula mientras no haya largado.
	Seed *int64
}

// Horse es un caballo. NominalOdds es la cuota nominal ×100: 340 es 3.40.
//
// Se llama nominal porque alimenta al simulador —es su parámetro de fuerza— y le
// indica al alumno quién es favorito. Si además determina el pago es la decisión
// de economía que está abierta; ver el comentario de internal/contract/money.go.
type Horse struct {
	ID          string
	RaceID      string
	Number      int
	Name        string
	NominalOdds int
}

// Bet es una apuesta. Username viene del join: los eventos del socket lo
// necesitan y pedirlo por separado sería una consulta por apuesta.
//
// NO tiene un pago potencial ni una cuota congelada. Con cuotas fijas habría que
// congelar la cuota acá; con pari-mutuel no existe un número al momento de la
// apuesta que tenga sentido guardar. La economía está sin decidir, así que la
// estructura no afirma ninguna de las dos: Payout se llena AL LIQUIDAR y lo
// calcula SettlementRule.
type Bet struct {
	ID       string
	RaceID   string
	UserID   string
	Username string
	HorseID  string

	Amount int64
	Status BetStatus
	Payout *int64

	CreatedAt time.Time
	SettledAt *time.Time
}

// Settlement es la liquidación de una carrera: una fila por carrera, escrita una
// sola vez.
//
// Guarda el pool y el pool ganador además de los pagos porque son lo que permite
// **reconstruir** por qué cada uno cobró lo que cobró. Con pari-mutuel el pago de
// una apuesta no se puede recalcular desde la apuesta sola: depende de cómo apostó
// el resto, y si mañana alguien reclama la nota, `amount * pool / winningPool` es
// la cuenta que hay que poder mostrar.
//
// `PaidOut` tiene que ser exactamente igual a `Pool`. Lo impone un CHECK en el
// esquema, no solo el código: es la conservación de monedas del curso.
type Settlement struct {
	RaceID        string
	WinnerHorseID string

	Pool        int64 // todo lo apostado en la carrera
	WinningPool int64 // lo apostado al ganador
	PaidOut     int64 // suma de los pagos efectivos
	Refunded    bool  // nadie acertó: se devolvió cada apuesta íntegra

	SettledAt time.Time
}

// Participant es un alumno en la sala.
type Participant struct {
	UserID    string
	Username  string
	FirstName string
	LastName  string
	JoinedAt  time.Time
}

// Summary es una carrera en el listado, con los agregados ya resueltos.
//
// MyBet viene resuelto acá y no en una llamada aparte: sin eso el frontend
// haría una consulta por carrera.
type Summary struct {
	Race
	HorseCount       int
	ParticipantCount int
	MyBet            *Bet
}

// ── Entradas ──────────────────────────────────────────────────────────────

// NewRace es lo que hace falta para crear una carrera. Nace en `draft`.
type NewRace struct {
	Name        string
	ScheduledAt *time.Time
	CreatedBy   string
	Horses      []NewHorse
}

// NewHorse es un caballo a agregar. Odds es ×100.
type NewHorse struct {
	Number int
	Name   string
	// NominalOdds se llama igual que en `horses` y que en el JSON a propósito.
	// Se llamó `Odds` mientras la economía era de cuota fija, y esa diferencia de
	// una palabra hizo que el endpoint recibiera `odds` y devolviera `nominalOdds`
	// durante semanas, sin que ningún test lo notara: los de handler mandaban lo
	// que el handler leía.
	NominalOdds int
}

// RacePatch son los campos editables de una carrera en `draft`. Nil es «no lo
// toques»: sin punteros no se podría distinguir de «ponelo en cero».
type RacePatch struct {
	Name        *string
	ScheduledAt *time.Time
}

// Movement es un movimiento del ledger. Delta positivo acredita, negativo
// descuenta.
type Movement struct {
	UserID string
	Delta  int64
	Reason Reason
	// RefID es a qué apunta el movimiento: el id de la apuesta. Es lo que le
	// permite al panel del instructor armar el rastro.
	RefID string
	// CreatedBy es el instructor, cuando el movimiento lo generó una acción
	// suya. Vacío cuando lo generó el sistema al liquidar.
	CreatedBy string
}

// Identity es quién hace la petición. La provee el paquete de autenticación por
// la interfaz Authenticator.
type Identity struct {
	UserID   string
	Username string
	Role     string
}

// IsAdmin es el chequeo de rol de decisiones.md §4. Se hace en el SERVIDOR, en
// cada handler de /admin/: un alumno que edite su token no obtiene nada.
func (i Identity) IsAdmin() bool { return i.Role == RoleAdmin }

const (
	RoleStudent = "student"
	RoleAdmin   = "admin"
)
