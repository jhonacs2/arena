// Package ledger es la única puerta por la que se mueven monedas.
//
// Dos reglas que no se negocian (arena/CLAUDE.md §4):
//
//   - **Todo movimiento escribe en `coin_transactions` Y actualiza
//     `users.balance` en la MISMA transacción.** `users.balance` es una caché;
//     la verdad es la suma de los deltas. Si las dos escrituras no son atómicas,
//     la caché y la verdad se separan y alguien termina con una nota que el
//     ledger no explica. Lo comprueba `arena/scripts/reconcile.sql`.
//   - **El ledger es append-only**, y lo impone un trigger de la base. Un error
//     se compensa con otro movimiento; no se edita la historia de la nota de
//     alguien.
//
// Los montos son `int64`. Nunca float — ver arena/CLAUDE.md §5.
package ledger

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/talentodh/arena/internal/contract"
)

// Los motivos del enum `ledger_reason` del esquema. Constantes sin tipo propio
// a propósito: el paquete de carreras tiene su propio `races.Reason` y así una
// conversión a string alcanza para cruzar el límite.
const (
	ReasonCodeRedeemed = "code_redeemed" // las 1000 monedas iniciales
	ReasonGift         = "gift"          // regalo del instructor
	ReasonBetPlaced    = "bet_placed"    // se descuenta al apostar
	ReasonBetWon       = "bet_won"       // se acredita el pago
	ReasonBetRefunded  = "bet_refunded"  // carrera cancelada o sin aciertos
	ReasonAdjustment   = "adjustment"    // corrección manual del instructor
)

// Movement es un movimiento del ledger. Delta positivo acredita, negativo
// descuenta. Nunca cero: lo prohíbe el CHECK delta_no_cero del esquema.
//
// La forma coincide campo por campo con `races.Movement`, así que el paquete de
// carreras cruza el límite con una conversión y no con un mapeo a mano.
type Movement struct {
	UserID string
	Delta  int64
	// Reason es uno de los valores del enum ledger_reason.
	Reason string
	// RefID es a qué apunta el movimiento: el id de la apuesta, el código
	// canjeado, o la nota del instructor si fue un regalo.
	RefID string
	// CreatedBy es el instructor, cuando el movimiento salió de una acción suya.
	// Vacío cuando lo generó el sistema al liquidar una carrera.
	CreatedBy string
}

// Querier es lo mínimo que necesita Move. Lo satisfacen `pgx.Tx` y
// `*pgxpool.Pool`, así que el mismo código sirve dentro de una transacción
// ajena y fuera de toda transacción.
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Move registra un movimiento y devuelve el saldo que dejó.
//
// Recibe el ejecutor a propósito: el paquete de carreras le pasa la `pgx.Tx` en
// la que está insertando la apuesta, y así el descuento y la apuesta se
// confirman juntos o no se confirman.
//
// El orden de las dos escrituras NO es casual: primero el UPDATE del saldo y
// después el INSERT del movimiento. El UPDATE toma el lock de la fila del
// usuario, así que dos movimientos simultáneos del mismo alumno se serializan
// ahí, y el que entra segundo saca su `id` de la secuencia después del primero.
// Al revés, los `id` podrían quedar en un orden distinto al de los saldos y la
// cadena de `balance_after` que verifica reconcile.sql se rompería sin que nada
// falle.
func Move(ctx context.Context, db Querier, m Movement) (int64, error) {
	if m.Delta == 0 {
		// El CHECK delta_no_cero lo rechazaría igual; acá el error dice qué pasó.
		return 0, fmt.Errorf("movimiento de cero monedas para %s (%s)", m.UserID, m.Reason)
	}

	// El `balance + $2 >= 0` es el piso del saldo de decisiones.md §1 aplicado en
	// la misma sentencia que descuenta: no hay ventana entre leer el saldo y
	// gastarlo. El CHECK balance_no_negativo del esquema es la red de atrás.
	var balance int64
	err := db.QueryRow(ctx, `
		update users
		set balance = balance + $2
		where id = $1 and balance + $2 >= 0
		returning balance`,
		m.UserID, m.Delta,
	).Scan(&balance)

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		// Cero filas por una de dos razones. Se distinguen porque el instructor
		// necesita saber si escribió mal el id o si el alumno no tiene saldo.
		var exists bool
		if err := db.QueryRow(ctx, `select true from users where id = $1`, m.UserID).Scan(&exists); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return 0, contract.Errorf(contract.CodeUserNotFound)
			}
			return 0, fmt.Errorf("verificando el usuario %s: %w", m.UserID, err)
		}
		return 0, contract.Errorf(contract.CodeInsufficientBalance)
	case err != nil:
		return 0, fmt.Errorf("actualizando el saldo de %s: %w", m.UserID, err)
	}

	if _, err := db.Exec(ctx, `
		insert into coin_transactions (user_id, delta, reason, ref_id, balance_after, created_by)
		values ($1, $2, $3::ledger_reason, $4, $5, $6)`,
		m.UserID, m.Delta, m.Reason, nullable(m.RefID), balance, nullable(m.CreatedBy),
	); err != nil {
		return 0, fmt.Errorf("escribiendo el movimiento de %s: %w", m.UserID, err)
	}
	return balance, nil
}

// IsInsufficientBalance dice si el movimiento no entró porque el saldo no
// alcanzaba. Es lo que le permite al llamador distinguir «no le alcanza» —que
// es una respuesta 409 al alumno— de «la base se cayó».
func IsInsufficientBalance(err error) bool {
	var fault *contract.Fault
	return errors.As(err, &fault) && fault.Code == contract.CodeInsufficientBalance
}

// nullable manda NULL en vez de cadena vacía. `ref_id` y `created_by` son
// nulables en el esquema, y un `created_by` en cadena vacía no es un uuid.
func nullable(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

// ── Store ─────────────────────────────────────────────────────────────────

// Store es el ledger con su propio pool, para los movimientos que no comparten
// transacción con nadie: un regalo del instructor, el canje de un código.
type Store struct{ Pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{Pool: pool} }

// Move registra un movimiento en su propia transacción.
//
// Una transacción para una sola sentencia doble parece de más y no lo es: son
// DOS escrituras —el saldo y el movimiento— y tienen que ser atómicas.
func (s *Store) Move(ctx context.Context, m Movement) (int64, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("abriendo la transacción del movimiento: %w", err)
	}
	defer tx.Rollback(ctx)

	balance, err := Move(ctx, tx, m)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("confirmando el movimiento: %w", err)
	}
	return balance, nil
}

// MoveTx registra un movimiento dentro de una transacción ajena. Es el mismo
// Move de paquete, expuesto como método para poder satisfacer una interfaz.
func (s *Store) MoveTx(ctx context.Context, tx pgx.Tx, m Movement) (int64, error) {
	return Move(ctx, tx, m)
}

// Balance devuelve el saldo cacheado del usuario.
func (s *Store) Balance(ctx context.Context, userID string) (int64, error) {
	var balance int64
	err := s.Pool.QueryRow(ctx, `select balance from users where id = $1`, userID).Scan(&balance)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, contract.Errorf(contract.CodeUserNotFound)
	}
	if err != nil {
		return 0, fmt.Errorf("leyendo el saldo de %s: %w", userID, err)
	}
	return balance, nil
}
