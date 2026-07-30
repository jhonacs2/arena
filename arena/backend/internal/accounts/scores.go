package accounts

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/talentodh/arena/internal/contract"
)

// Score es una fila del panel de nota del instructor. Sale de la vista
// `user_scores`, que es la ÚNICA que sabe la fórmula:
//
//	puntos = max(10, floor(monedas / 100)) + puntos regalados
//
// Se lee de la vista y no se recalcula en Go a propósito. Si la fórmula viviera
// en los dos lados, un día se separarían y el número de la nota dependería de
// por dónde se lo mire. `schema.test.sql` verifica la vista; no hay un segundo
// lugar que verificar.
//
// Y ya se cobró el beneficio: la fórmula cambió tres veces durante la construcción
// de este backend —el piso entró, salió y volvió— y como se lee de la vista, nunca
// hubo una línea de Go que tocar.
//
// **El piso de 10** (decisiones.md §0) es lo que el alumno recibe al canjear el
// código, y apostar mal no lo puede tocar: pierde monedas, no calificación. Va
// junto con el pari-mutuel, porque con un pool de suma cero y sin piso la nota que
// gana uno saldría de la de otro. Y los puntos REGALADOS se suman por encima del
// piso, no compiten con él: por eso viven en `point_grants` y no en el ledger.
type Score struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`

	Balance int64 `json:"balance"`

	// PointsFromCoins son los puntos que salen del saldo, y suben y bajan con él.
	PointsFromCoins Points `json:"pointsFromCoins"`
	// PointsGranted son los regalados por el instructor, que no pasan por el juego
	// y por lo tanto no se pueden perder apostando.
	PointsGranted Points `json:"pointsGranted"`
	// Points es la nota: la suma de los dos.
	Points Points `json:"points"`

	BetsPlaced int64 `json:"betsPlaced"`
	BetsWon    int64 `json:"betsWon"`
}

// Los puntos se leen de la vista multiplicados por 100 y como entero: así
// cruzan a Go sin pasar por un float. Ver el tipo Points.
const scoreColumns = `
	id::text,
	username,
	first_name,
	last_name,
	balance,
	(points_from_coins * 100)::bigint,
	(points_granted    * 100)::bigint,
	(points            * 100)::bigint,
	bets_placed,
	bets_won`

// Scores es el panel de nota, de mayor a menor.
//
// La vista ya filtra `role = 'student'`: el instructor no se califica a sí mismo
// y aparecer en su propio ranking con 0 puntos sería ruido.
func (s *Store) Scores(ctx context.Context) ([]Score, error) {
	rows, err := s.Pool.Query(ctx,
		`select `+scoreColumns+` from user_scores order by points desc, balance desc, lower(username)`)
	if err != nil {
		return nil, fmt.Errorf("leyendo el panel de nota: %w", err)
	}
	defer rows.Close()

	items := []Score{}
	for rows.Next() {
		var sc Score
		if err := rows.Scan(&sc.ID, &sc.Username, &sc.FirstName, &sc.LastName, &sc.Balance,
			&sc.PointsFromCoins, &sc.PointsGranted, &sc.Points,
			&sc.BetsPlaced, &sc.BetsWon); err != nil {
			return nil, fmt.Errorf("leyendo una fila del panel: %w", err)
		}
		items = append(items, sc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("recorriendo el panel: %w", err)
	}
	return items, nil
}

// PointsFor son los puntos de un usuario, leídos de la misma vista.
//
// El instructor NO está en `user_scores` —la vista filtra por `role = 'student'`—
// así que para él devuelve 0. Es correcto: el instructor no tiene nota.
func (s *Store) PointsFor(ctx context.Context, userID string) (Points, error) {
	var points Points
	err := s.Pool.QueryRow(ctx,
		`select (points * 100)::bigint from user_scores where id = $1`, userID).Scan(&points)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("leyendo los puntos de %s: %w", userID, err)
	}
	return points, nil
}

// PointGrant es un regalo de puntos ya registrado.
type PointGrant struct {
	ID        int64     `json:"id"`
	UserID    string    `json:"userId"`
	Points    Points    `json:"points"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"createdAt"`
}

// GrantPoints regala puntos, que **no es lo mismo que regalar monedas**
// (decisiones.md §1):
//
//	monedas → le da con qué seguir jugando, y se puede volver a perder
//	puntos  → le sube la nota directo, no pasa por el juego y no se pierde
//
// Por eso viven en tablas separadas y por eso esto no toca el ledger.
//
// `point_grants` es append-only y lo impone un trigger, igual que el ledger: un
// regalo mal hecho se compensa con otro de signo contrario, no se borra.
//
// El monto llega en centésimas de punto: 250 son 2,5 puntos. La conversión a la
// columna `numeric(6,2)` se hace en SQL y es exacta —`numeric` no es binario, no
// hay error de redondeo.
func (s *Store) GrantPoints(ctx context.Context, adminID, userID string, points Points, reason string) (PointGrant, error) {
	if points == 0 {
		return PointGrant{}, contract.FieldErrors(map[string]string{
			"points": "Un regalo de cero puntos no cambia nada.",
		})
	}
	if reason == "" {
		return PointGrant{}, contract.FieldErrors(map[string]string{
			"reason": "Escribí por qué se los das: queda en el registro.",
		})
	}

	var grant PointGrant
	err := s.Pool.QueryRow(ctx, `
		insert into point_grants (user_id, points, reason, granted_by)
		values ($1, $2::numeric / 100, $3, $4)
		returning id, user_id::text, (points * 100)::bigint, reason, created_at`,
		userID, int64(points), reason, adminID,
	).Scan(&grant.ID, &grant.UserID, &grant.Points, &grant.Reason, &grant.CreatedAt)

	// La FK a users es lo que atrapa un id inventado. Se traduce a 404 en vez de
	// dejar salir un 500: el instructor escribió mal el id, no se rompió nada.
	if isForeignKeyViolation(err) {
		return PointGrant{}, contract.Errorf(contract.CodeUserNotFound)
	}
	if err != nil {
		return PointGrant{}, fmt.Errorf("regalando puntos a %s: %w", userID, err)
	}
	return grant, nil
}
