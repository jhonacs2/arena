// Package accounts es la identidad: usuarios, códigos de invitación, sesiones y
// la vista de nota.
//
// Todo lo que toca monedas pasa por internal/ledger. Acá se abre la transacción
// del canje y se llama al ledger adentro, para que crear el usuario, quemar el
// código y acreditar las 1000 monedas sean una sola cosa.
package accounts

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Roles de `users.role`. El rol se valida SIEMPRE en el servidor y se lee de la
// base, no del token: ver decisiones.md §4.
const (
	RoleStudent = "student"
	RoleAdmin   = "admin"
)

// ErrNotFound es «no hay tal fila». No se traduce a un código HTTP acá porque
// depende de quién pregunte: un usuario que no existe es 401 si es el dueño del
// token y 404 si es un id que escribió el instructor.
var ErrNotFound = errors.New("no encontrado")

type Store struct{ Pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{Pool: pool} }

// User es un usuario como sale al cable. `balance` y `role` no van en el objeto
// `user` del contrato —el saldo viaja al lado, no adentro— así que se marcan
// fuera del JSON y el handler los pone donde corresponde.
type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Role      string `json:"role"`

	Balance   int64     `json:"-"`
	CreatedAt time.Time `json:"-"`
}

func (u User) IsAdmin() bool { return u.Role == RoleAdmin }

// userColumns es la proyección de User. En un solo lugar para que agregar una
// columna no deje una consulta escaneando de menos.
//
// Los `::text` no son decoración: `id` es uuid y `role` es un enum, y castearlos
// en la consulta evita depender de cómo pgx resuelve un OID que no registró.
const userColumns = `id::text, username, first_name, last_name, role::text, balance, created_at`

// scanner es lo que devuelven QueryRow y Rows.Scan por igual.
type scanner interface{ Scan(dest ...any) error }

func scanUser(row scanner) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Username, &u.FirstName, &u.LastName, &u.Role, &u.Balance, &u.CreatedAt)
	return u, err
}

// ── Puntos ────────────────────────────────────────────────────────────────

// Points es una nota, expresada en **centésimas de punto**: 1750 son 17,5
// puntos.
//
// Entera por el mismo motivo que las monedas (arena/CLAUDE.md §5): un float que
// redondea mal acá se lleva medio punto de alguien. Los puntos sí tienen
// decimales —`point_grants.points` es `numeric(6,2)` y un regalo de 2,5 puntos
// es normal— pero eso no obliga a usar coma flotante: se guarda el entero de
// centésimas y se formatea al serializar.
type Points int64

// MarshalJSON emite un número JSON, no una cadena: `17.5`, no `"17.5"`.
//
// Sin ceros de relleno, porque el frontend lo muestra tal cual y `10` se lee
// mejor que `10.00` en una tabla de notas.
func (p Points) MarshalJSON() ([]byte, error) { return []byte(p.String()), nil }

func (p Points) String() string {
	sign, n := "", int64(p)
	if n < 0 {
		sign, n = "-", -n
	}
	whole, cents := n/100, n%100

	switch {
	case cents == 0:
		return sign + strconv.FormatInt(whole, 10)
	case cents%10 == 0:
		return fmt.Sprintf("%s%d.%d", sign, whole, cents/10)
	default:
		return fmt.Sprintf("%s%d.%02d", sign, whole, cents)
	}
}
