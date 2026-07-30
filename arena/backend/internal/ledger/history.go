package ledger

import (
	"context"
	"fmt"
	"time"
)

// Transaction es un movimiento como lo ve el alumno en su historial.
//
// Es lo que le permite entender por qué tiene la nota que tiene, así que lleva
// el nombre de la carrera resuelto: sin eso, el frontend vería `bet_placed` con
// un uuid al lado.
type Transaction struct {
	ID           int64     `json:"id"`
	Delta        int64     `json:"delta"`
	Reason       string    `json:"reason"`
	BalanceAfter int64     `json:"balanceAfter"`
	CreatedAt    time.Time `json:"createdAt"`

	// RaceName solo viene en los movimientos de apuesta.
	RaceName string `json:"raceName,omitempty"`
	// Note es la nota del instructor en un regalo o un ajuste.
	Note string `json:"note,omitempty"`
}

// Topes del historial. El default alcanza para una cursada entera y el máximo
// existe para que `?limit=999999` no se lleve la base puesta.
const (
	DefaultHistoryLimit = 100
	MaxHistoryLimit     = 500
)

// Transactions devuelve el historial del alumno, **más nuevo primero**.
//
// El orden es por `id` y no por `created_at`: dos movimientos de la misma
// transacción comparten el `now()` de Postgres al milisegundo, y ordenar por
// fecha los mostraría en cualquier orden. El `id` es una secuencia y no empata.
func (s *Store) Transactions(ctx context.Context, userID string, limit, offset int) ([]Transaction, error) {
	if limit <= 0 {
		limit = DefaultHistoryLimit
	}
	limit = min(limit, MaxHistoryLimit)
	offset = max(offset, 0)

	// El join va contra `bets.id::text` y no al revés: `ref_id` es texto libre y
	// en un canje guarda el código (`AVBD-1234`), que no es un uuid. Castear
	// ref_id a uuid haría fallar la consulta entera por esas filas.
	rows, err := s.Pool.Query(ctx, `
		select t.id,
		       t.delta,
		       t.reason::text,
		       t.balance_after,
		       t.created_at,
		       coalesce(r.name, '') as race_name,
		       case when t.reason in ('gift', 'adjustment') then coalesce(t.ref_id, '') else '' end as note
		from coin_transactions t
		left join bets b
		       on t.reason in ('bet_placed', 'bet_won', 'bet_refunded')
		      and b.id::text = t.ref_id
		left join races r on r.id = b.race_id
		where t.user_id = $1
		order by t.id desc
		limit $2 offset $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("leyendo el historial de %s: %w", userID, err)
	}
	defer rows.Close()

	// Se devuelve un slice vacío y no nil: `{"items": []}` es lo que el frontend
	// espera, y `null` lo obligaría a chequear.
	items := []Transaction{}
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.ID, &t.Delta, &t.Reason, &t.BalanceAfter, &t.CreatedAt, &t.RaceName, &t.Note); err != nil {
			return nil, fmt.Errorf("leyendo un movimiento de %s: %w", userID, err)
		}
		items = append(items, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("recorriendo el historial de %s: %w", userID, err)
	}
	return items, nil
}

// Gift acredita o descuenta monedas por decisión del instructor.
//
// Un `coins` positivo es un regalo (`gift`) y uno negativo es una corrección
// (`adjustment`). Son motivos distintos en el enum porque el panel agrupa por
// eso: «cuánto regalé» y «cuánto corregí» no son la misma pregunta.
//
// Queda `created_by` con el instructor, así que el rastro de quién lo hizo no se
// pierde. Es la contracara de que las monedas sean nota.
func (s *Store) Gift(ctx context.Context, adminID, userID string, coins int64, note string) (int64, error) {
	reason := ReasonGift
	if coins < 0 {
		reason = ReasonAdjustment
	}

	// La nota va en `ref_id`. No es lo que la columna documenta —«el id de la
	// apuesta, o el código canjeado»— y es el único lugar que hay:
	// `coin_transactions` no tiene columna de nota y el esquema es del contrato.
	// Si molesta, es una columna nueva en schema.sql y una línea acá.
	return s.Move(ctx, Movement{
		UserID:    userID,
		Delta:     coins,
		Reason:    reason,
		RefID:     note,
		CreatedBy: adminID,
	})
}
