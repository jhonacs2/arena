package ledger_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/talentodh/arena/internal/accounts"
	"github.com/talentodh/arena/internal/contract"
	"github.com/talentodh/arena/internal/ledger"
	"github.com/talentodh/arena/internal/testdb"
)

func TestMoveAcreditaYDescuenta(t *testing.T) {
	pool := testdb.Pool(t)
	store := ledger.New(pool)
	ctx := context.Background()

	admin := testdb.InsertUser(t, pool, "profe", accounts.RoleAdmin)
	ana := testdb.InsertUser(t, pool, "ana", accounts.RoleStudent)

	// Las 1000 iniciales.
	balance, err := store.Move(ctx, ledger.Movement{
		UserID: ana, Delta: 1000, Reason: ledger.ReasonCodeRedeemed, RefID: "AVBD-2345",
	})
	if err != nil {
		t.Fatalf("acreditando: %v", err)
	}
	if balance != 1000 {
		t.Errorf("saldo = %d, se esperaba 1000", balance)
	}

	// Apuesta de 200.
	balance, err = store.Move(ctx, ledger.Movement{
		UserID: ana, Delta: -200, Reason: ledger.ReasonBetPlaced, RefID: "una-apuesta",
	})
	if err != nil {
		t.Fatalf("descontando: %v", err)
	}
	if balance != 800 {
		t.Errorf("saldo = %d, se esperaba 800", balance)
	}

	// Cobra 534 (el pago pari-mutuel del ejemplo de decisiones.md).
	balance, err = store.Move(ctx, ledger.Movement{
		UserID: ana, Delta: 534, Reason: ledger.ReasonBetWon, RefID: "una-apuesta",
	})
	if err != nil {
		t.Fatalf("pagando: %v", err)
	}
	if balance != 1334 {
		t.Errorf("saldo = %d, se esperaba 1334", balance)
	}

	// Regalo del instructor, con rastro de quién lo hizo.
	balance, err = store.Gift(ctx, admin, ana, 300, "participación en clase")
	if err != nil {
		t.Fatalf("regalando: %v", err)
	}
	if balance != 1634 {
		t.Errorf("saldo = %d, se esperaba 1634", balance)
	}

	if got := testdb.Balance(t, pool, ana); got != 1634 {
		t.Errorf("users.balance = %d, se esperaba 1634", got)
	}
	if got := testdb.CountTransactions(t, pool, ana); got != 4 {
		t.Errorf("movimientos = %d, se esperaban 4", got)
	}
	testdb.Reconcile(t, pool)
}

// El regalo negativo es un `adjustment`, no un `gift`: el panel del instructor
// agrupa por motivo, y «cuánto regalé» y «cuánto corregí» no son la misma
// pregunta.
func TestGiftNegativoEsUnAjusteYGuardaElInstructor(t *testing.T) {
	pool := testdb.Pool(t)
	store := ledger.New(pool)
	ctx := context.Background()

	admin := testdb.InsertUser(t, pool, "profe", accounts.RoleAdmin)
	ana := testdb.InsertUser(t, pool, "ana", accounts.RoleStudent)

	if _, err := store.Move(ctx, ledger.Movement{
		UserID: ana, Delta: 1000, Reason: ledger.ReasonCodeRedeemed,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Gift(ctx, admin, ana, -150, "corrección"); err != nil {
		t.Fatalf("ajustando: %v", err)
	}

	var reason, createdBy, refID string
	err := pool.QueryRow(ctx, `
		select reason::text, created_by::text, ref_id
		from coin_transactions where user_id = $1 order by id desc limit 1`, ana,
	).Scan(&reason, &createdBy, &refID)
	if err != nil {
		t.Fatal(err)
	}

	if reason != ledger.ReasonAdjustment {
		t.Errorf("reason = %q, se esperaba %q", reason, ledger.ReasonAdjustment)
	}
	if createdBy != admin {
		t.Errorf("created_by = %q, se esperaba el instructor %q", createdBy, admin)
	}
	if refID != "corrección" {
		t.Errorf("ref_id = %q: se perdió la nota del ajuste", refID)
	}
	testdb.Reconcile(t, pool)
}

// El piso del saldo (decisiones.md §1). Y cuando el movimiento no entra, **no
// entra nada**: ni el saldo ni la fila del ledger.
func TestMoveNoDejaElSaldoNegativo(t *testing.T) {
	pool := testdb.Pool(t)
	store := ledger.New(pool)
	ctx := context.Background()

	ana := testdb.InsertUser(t, pool, "ana", accounts.RoleStudent)
	if _, err := store.Move(ctx, ledger.Movement{
		UserID: ana, Delta: 1000, Reason: ledger.ReasonCodeRedeemed,
	}); err != nil {
		t.Fatal(err)
	}

	_, err := store.Move(ctx, ledger.Movement{
		UserID: ana, Delta: -1001, Reason: ledger.ReasonBetPlaced,
	})
	if err == nil {
		t.Fatal("se aceptó un descuento mayor al saldo")
	}
	if !ledger.IsInsufficientBalance(err) {
		t.Errorf("err = %v; se esperaba INSUFFICIENT_BALANCE", err)
	}

	var fault *contract.Fault
	if !errors.As(err, &fault) || fault.Code != contract.CodeInsufficientBalance {
		t.Errorf("el error no es el del contrato: %v", err)
	}

	if got := testdb.Balance(t, pool, ana); got != 1000 {
		t.Errorf("saldo = %d: el movimiento rechazado igual tocó el saldo", got)
	}
	if got := testdb.CountTransactions(t, pool, ana); got != 1 {
		t.Errorf("movimientos = %d: quedó una fila de un movimiento que no entró", got)
	}
	testdb.Reconcile(t, pool)
}

// Gastar exactamente todo el saldo sí entra: el piso es 0, no 1.
func TestMovePermiteFundirseHastaCero(t *testing.T) {
	pool := testdb.Pool(t)
	store := ledger.New(pool)
	ctx := context.Background()

	ana := testdb.InsertUser(t, pool, "ana", accounts.RoleStudent)
	if _, err := store.Move(ctx, ledger.Movement{
		UserID: ana, Delta: 1000, Reason: ledger.ReasonCodeRedeemed,
	}); err != nil {
		t.Fatal(err)
	}

	balance, err := store.Move(ctx, ledger.Movement{
		UserID: ana, Delta: -1000, Reason: ledger.ReasonBetPlaced,
	})
	if err != nil {
		t.Fatalf("apostar todo el saldo tendría que entrar: %v", err)
	}
	if balance != 0 {
		t.Errorf("saldo = %d, se esperaba 0", balance)
	}
	testdb.Reconcile(t, pool)
}

func TestMoveRechazaUsuarioInexistenteYDeltaCero(t *testing.T) {
	pool := testdb.Pool(t)
	store := ledger.New(pool)
	ctx := context.Background()

	_, err := store.Move(ctx, ledger.Movement{
		UserID: "99999999-9999-9999-9999-999999999999", Delta: 100, Reason: ledger.ReasonGift,
	})
	var fault *contract.Fault
	if !errors.As(err, &fault) || fault.Code != contract.CodeUserNotFound {
		t.Errorf("err = %v; se esperaba USER_NOT_FOUND", err)
	}

	ana := testdb.InsertUser(t, pool, "ana", accounts.RoleStudent)
	if _, err := store.Move(ctx, ledger.Movement{
		UserID: ana, Delta: 0, Reason: ledger.ReasonGift,
	}); err == nil {
		t.Error("se aceptó un movimiento de cero monedas")
	}
}

// El append-only del esquema, verificado desde acá y no solo en schema.test.sql:
// es la propiedad de la que depende que el historial del alumno pruebe algo.
func TestElLedgerNoSePuedeEditarNiBorrar(t *testing.T) {
	pool := testdb.Pool(t)
	store := ledger.New(pool)
	ctx := context.Background()

	ana := testdb.InsertUser(t, pool, "ana", accounts.RoleStudent)
	if _, err := store.Move(ctx, ledger.Movement{
		UserID: ana, Delta: 1000, Reason: ledger.ReasonCodeRedeemed,
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := pool.Exec(ctx,
		`update coin_transactions set delta = 999999 where user_id = $1`, ana); err == nil {
		t.Error("se pudo editar el ledger")
	}
	if _, err := pool.Exec(ctx,
		`delete from coin_transactions where user_id = $1`, ana); err == nil {
		t.Error("se pudo borrar del ledger")
	}

	if got := testdb.Balance(t, pool, ana); got != 1000 {
		t.Errorf("saldo = %d, se esperaba 1000", got)
	}
	testdb.Reconcile(t, pool)
}

// Muchos descuentos simultáneos del mismo alumno. Ninguno se pierde, ninguno se
// aplica dos veces, y la cadena de `balance_after` queda en orden.
//
// Es el escenario de una clase: veinte alumnos apostando al mismo tiempo es
// concurrencia entre filas distintas y no prueba nada; lo que rompe es el MISMO
// alumno con dos peticiones encimadas.
func TestMoveConcurrenteNoPierdeNiDuplicaMovimientos(t *testing.T) {
	pool := testdb.Pool(t)
	store := ledger.New(pool)
	ctx := context.Background()

	ana := testdb.InsertUser(t, pool, "ana", accounts.RoleStudent)
	if _, err := store.Move(ctx, ledger.Movement{
		UserID: ana, Delta: 1000, Reason: ledger.ReasonCodeRedeemed,
	}); err != nil {
		t.Fatal(err)
	}

	const movimientos = 20
	const monto = 50 // 20 × 50 = 1000: gasta el saldo exacto

	var wg sync.WaitGroup
	errs := make([]error, movimientos)
	for i := range movimientos {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = store.Move(ctx, ledger.Movement{
				UserID: ana, Delta: -monto, Reason: ledger.ReasonBetPlaced,
			})
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("movimiento %d falló: %v", i, err)
		}
	}

	if got := testdb.Balance(t, pool, ana); got != 0 {
		t.Errorf("saldo = %d, se esperaba 0 (1000 − 20 × 50)", got)
	}
	if got := testdb.CountTransactions(t, pool, ana); got != movimientos+1 {
		t.Errorf("movimientos = %d, se esperaban %d", got, movimientos+1)
	}

	// La aserción que importa: si dos movimientos leyeran el saldo antes de que el
	// otro escriba, la suma seguiría dando bien pero la cadena de balance_after
	// quedaría rota.
	testdb.Reconcile(t, pool)
}

// Y con el saldo justo: si 21 goroutines intentan gastar 1000 en tramos de 50,
// exactamente una tiene que quedarse afuera. El piso del saldo no se puede
// saltear por concurrencia.
func TestMoveConcurrenteRespetaElPisoDelSaldo(t *testing.T) {
	pool := testdb.Pool(t)
	store := ledger.New(pool)
	ctx := context.Background()

	ana := testdb.InsertUser(t, pool, "ana", accounts.RoleStudent)
	if _, err := store.Move(ctx, ledger.Movement{
		UserID: ana, Delta: 1000, Reason: ledger.ReasonCodeRedeemed,
	}); err != nil {
		t.Fatal(err)
	}

	const intentos = 25
	const monto = 50 // solo 20 pueden entrar

	var wg sync.WaitGroup
	var mu sync.Mutex
	var entraron, rechazados int

	for range intentos {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := store.Move(ctx, ledger.Movement{
				UserID: ana, Delta: -monto, Reason: ledger.ReasonBetPlaced,
			})

			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				entraron++
			case ledger.IsInsufficientBalance(err):
				rechazados++
			default:
				t.Errorf("error inesperado: %v", err)
			}
		}()
	}
	wg.Wait()

	if entraron != 20 || rechazados != 5 {
		t.Errorf("entraron %d y se rechazaron %d; se esperaban 20 y 5", entraron, rechazados)
	}
	if got := testdb.Balance(t, pool, ana); got != 0 {
		t.Errorf("saldo = %d, se esperaba 0 y nunca negativo", got)
	}
	testdb.Reconcile(t, pool)
}

func TestTransactionsDevuelveElHistorialMasNuevoPrimero(t *testing.T) {
	pool := testdb.Pool(t)
	store := ledger.New(pool)
	ctx := context.Background()

	admin := testdb.InsertUser(t, pool, "profe", accounts.RoleAdmin)
	ana := testdb.InsertUser(t, pool, "ana", accounts.RoleStudent)

	if _, err := store.Move(ctx, ledger.Movement{
		UserID: ana, Delta: 1000, Reason: ledger.ReasonCodeRedeemed, RefID: "AVBD-2345",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Gift(ctx, admin, ana, 300, "explicó @for"); err != nil {
		t.Fatal(err)
	}

	items, err := store.Transactions(ctx, ana, 0, 0)
	if err != nil {
		t.Fatalf("Transactions: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, se esperaban 2", len(items))
	}

	if items[0].Reason != ledger.ReasonGift || items[0].Delta != 300 || items[0].BalanceAfter != 1300 {
		t.Errorf("el primero tendría que ser el regalo más nuevo: %+v", items[0])
	}
	if items[0].Note != "explicó @for" {
		t.Errorf("nota = %q, se esperaba la del regalo", items[0].Note)
	}
	if items[1].Reason != ledger.ReasonCodeRedeemed || items[1].BalanceAfter != 1000 {
		t.Errorf("el segundo tendría que ser el canje: %+v", items[1])
	}
	// El código canjeado NO es una nota del instructor: no se muestra como tal.
	if items[1].Note != "" {
		t.Errorf("nota = %q en un canje; el ref_id de un canje es el código, no una nota", items[1].Note)
	}

	// El historial de otro alumno no se mezcla.
	bruno := testdb.InsertUser(t, pool, "bruno", accounts.RoleStudent)
	otros, err := store.Transactions(ctx, bruno, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(otros) != 0 {
		t.Errorf("el historial de bruno tiene %d movimientos ajenos", len(otros))
	}
}

func TestTransactionsAcotaElLimite(t *testing.T) {
	pool := testdb.Pool(t)
	store := ledger.New(pool)
	ctx := context.Background()

	ana := testdb.InsertUser(t, pool, "ana", accounts.RoleStudent)
	for range 5 {
		if _, err := store.Move(ctx, ledger.Movement{
			UserID: ana, Delta: 10, Reason: ledger.ReasonGift,
		}); err != nil {
			t.Fatal(err)
		}
	}

	items, err := store.Transactions(ctx, ana, 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Errorf("len(items) = %d con limit=2", len(items))
	}

	// Un límite absurdo se recorta al máximo, no revienta ni trae la tabla entera.
	items, err = store.Transactions(ctx, ana, 999999, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 5 {
		t.Errorf("len(items) = %d, se esperaban los 5 movimientos", len(items))
	}
}
