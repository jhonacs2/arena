package testdb

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Estos ayudantes escriben SQL a mano en vez de usar internal/accounts a
// propósito: si el armado del escenario pasara por el mismo código que se está
// probando, un bug en ese código haría pasar el test. Además evita el ciclo de
// importación entre un paquete y el test del paquete que lo usa.

// InsertUser crea un usuario con saldo 0.
//
// Saldo 0 y no un saldo de regalo: el saldo se mueve con el ledger, y un saldo
// puesto a mano sin movimiento detrás es exactamente el estado que
// reconcile.sql busca. Los tests que necesitan saldo lo acreditan con Move.
func InsertUser(t *testing.T, pool *pgxpool.Pool, username, role string) string {
	t.Helper()

	var id string
	err := pool.QueryRow(context.Background(), `
		insert into users (username, first_name, last_name, password_hash, role)
		values ($1, $2, 'Prueba', 'x', $3::user_role)
		returning id::text`, username, username, role).Scan(&id)
	if err != nil {
		t.Fatalf("creando el usuario %q: %v", username, err)
	}
	return id
}

// InsertCode crea un código de invitación sin canjear.
func InsertCode(t *testing.T, pool *pgxpool.Pool, code string, coins int64, createdBy string) {
	t.Helper()

	_, err := pool.Exec(context.Background(), `
		insert into invite_codes (code, coins_granted, created_by) values ($1, $2, $3)`,
		code, coins, createdBy)
	if err != nil {
		t.Fatalf("creando el código %s: %v", code, err)
	}
}

// Balance lee el saldo cacheado.
func Balance(t *testing.T, pool *pgxpool.Pool, userID string) int64 {
	t.Helper()

	var balance int64
	if err := pool.QueryRow(context.Background(),
		`select balance from users where id = $1`, userID).Scan(&balance); err != nil {
		t.Fatalf("leyendo el saldo de %s: %v", userID, err)
	}
	return balance
}

// Reconcile es arena/scripts/reconcile.sql como aserción.
//
// Comprueba las dos igualdades que sostienen que la nota de alguien sea
// explicable:
//
//  1. `users.balance` = suma de `coin_transactions.delta`. La columna es una
//     caché; si se separa de la suma, hay monedas que nadie puede rastrear.
//  2. Cada `balance_after` es el anterior más su `delta`. Encuentra el
//     movimiento exacto donde se rompió la cadena, que es lo que hace falta para
//     escribir la compensación.
//
// Se llama al final de todos los tests que mueven monedas. Es la aserción más
// importante del paquete: un test puede verificar que el saldo final es el
// esperado y aun así haber dejado el ledger inconsistente.
func Reconcile(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	rows, err := pool.Query(ctx, `
		with ledger as (
			select user_id, sum(delta)::bigint as ledger_balance
			from coin_transactions group by user_id
		)
		select u.username, u.balance, coalesce(l.ledger_balance, 0)
		from users u
		left join ledger l on l.user_id = u.id
		where u.balance <> coalesce(l.ledger_balance, 0)`)
	if err != nil {
		t.Fatalf("reconciliando el saldo contra el ledger: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var username string
		var cached, fromLedger int64
		if err := rows.Scan(&username, &cached, &fromLedger); err != nil {
			t.Fatalf("leyendo la reconciliación: %v", err)
		}
		t.Errorf("reconciliación: %s tiene saldo %d pero el ledger suma %d", username, cached, fromLedger)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("recorriendo la reconciliación: %v", err)
	}

	chain, err := pool.Query(ctx, `
		with chained as (
			select id, user_id, delta, balance_after,
			       lag(balance_after, 1, 0::bigint) over (partition by user_id order by id) as previous
			from coin_transactions
		)
		select id, user_id::text, previous, delta, balance_after
		from chained
		where balance_after <> previous + delta`)
	if err != nil {
		t.Fatalf("verificando la cadena de balance_after: %v", err)
	}
	defer chain.Close()

	for chain.Next() {
		var id, previous, delta, after int64
		var userID string
		if err := chain.Scan(&id, &userID, &previous, &delta, &after); err != nil {
			t.Fatalf("leyendo la cadena: %v", err)
		}
		t.Errorf("cadena rota en el movimiento %d de %s: %d + %d debería dar %d, dio %d",
			id, userID, previous, delta, previous+delta, after)
	}
	if err := chain.Err(); err != nil {
		t.Fatalf("recorriendo la cadena: %v", err)
	}
}

// CountTransactions cuenta los movimientos de un usuario.
func CountTransactions(t *testing.T, pool *pgxpool.Pool, userID string) int {
	t.Helper()

	var count int
	if err := pool.QueryRow(context.Background(),
		`select count(*) from coin_transactions where user_id = $1`, userID).Scan(&count); err != nil {
		t.Fatalf("contando los movimientos de %s: %v", userID, err)
	}
	return count
}
