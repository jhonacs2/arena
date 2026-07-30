package accounts_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/talentodh/arena/internal/accounts"
	"github.com/talentodh/arena/internal/auth"
	"github.com/talentodh/arena/internal/contract"
	"github.com/talentodh/arena/internal/ledger"
	"github.com/talentodh/arena/internal/testdb"
)

// faultCode saca el código del contrato de un error, o "" si no es un Fault.
func faultCode(err error) contract.Code {
	var fault *contract.Fault
	if errors.As(err, &fault) {
		return fault.Code
	}
	return ""
}

func input(code, username string) accounts.RedeemInput {
	return accounts.RedeemInput{
		Code: code, FirstName: "Ana", LastName: "Gómez",
		Username: username, PasswordHash: "pbkdf2$sha256$1$YQ$Yg",
	}
}

// ── El canje ──────────────────────────────────────────────────────────────

func TestRedeemCreaUsuarioQuemaElCodigoYAcredita(t *testing.T) {
	pool := testdb.Pool(t)
	store := accounts.New(pool)
	ctx := context.Background()

	admin := testdb.InsertUser(t, pool, "profe", accounts.RoleAdmin)
	testdb.InsertCode(t, pool, "AVBD-2345", 1000, admin)

	user, err := store.Redeem(ctx, input("AVBD-2345", "anag"))
	if err != nil {
		t.Fatalf("Redeem: %v", err)
	}

	if user.Balance != 1000 {
		t.Errorf("saldo = %d, se esperaba 1000", user.Balance)
	}
	if user.Role != accounts.RoleStudent {
		t.Errorf("rol = %q; un canje crea alumnos, no instructores", user.Role)
	}

	// El código quedó marcado, y con quién lo usó.
	var redeemedBy string
	var redeemedAt *time.Time
	if err := pool.QueryRow(ctx,
		`select redeemed_by::text, redeemed_at from invite_codes where code = 'AVBD-2345'`,
	).Scan(&redeemedBy, &redeemedAt); err != nil {
		t.Fatal(err)
	}
	if redeemedBy != user.ID {
		t.Errorf("redeemed_by = %q, se esperaba %q", redeemedBy, user.ID)
	}
	if redeemedAt == nil {
		t.Error("redeemed_at quedó en null: el CHECK canje_consistente tendría que impedirlo")
	}

	// Las monedas iniciales están en el ledger, no puestas a mano en el saldo.
	if got := testdb.CountTransactions(t, pool, user.ID); got != 1 {
		t.Errorf("movimientos = %d, se esperaba 1 (el canje)", got)
	}
	var reason, refID string
	if err := pool.QueryRow(ctx,
		`select reason::text, ref_id from coin_transactions where user_id = $1`, user.ID,
	).Scan(&reason, &refID); err != nil {
		t.Fatal(err)
	}
	if reason != ledger.ReasonCodeRedeemed {
		t.Errorf("reason = %q, se esperaba %q", reason, ledger.ReasonCodeRedeemed)
	}
	if refID != "AVBD-2345" {
		t.Errorf("ref_id = %q; el primer movimiento tiene que decir de qué código salió", refID)
	}

	testdb.Reconcile(t, pool)
}

// **El test que más importa del paquete.**
//
// Dos personas mandan el mismo código en el mismo instante: exactamente una gana
// y la otra recibe CODE_ALREADY_REDEEMED. Ni las dos, ni ninguna.
//
// Es lo que un `select` seguido de un `insert` no puede dar: las dos leerían el
// código libre y las dos seguirían. Lo que lo sostiene es el `for update` del
// canje, que hace esperar a la segunda transacción hasta que la primera confirme.
func TestRedeemConcurrenteGanaExactamenteUno(t *testing.T) {
	pool := testdb.Pool(t)
	store := accounts.New(pool)
	ctx := context.Background()

	admin := testdb.InsertUser(t, pool, "profe", accounts.RoleAdmin)
	testdb.InsertCode(t, pool, "AVBD-2345", 1000, admin)

	// Una barrera para que las dos goroutines lleguen al canje juntas. Sin ella,
	// la primera podría terminar antes de que la segunda arranque y el test no
	// probaría nada aunque el código estuviera mal.
	var listos sync.WaitGroup
	listos.Add(2)
	largada := make(chan struct{})

	type resultado struct {
		user accounts.User
		err  error
	}
	resultados := make([]resultado, 2)

	var wg sync.WaitGroup
	for i, username := range []string{"anag", "brunod"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			listos.Done()
			<-largada

			user, err := store.Redeem(ctx, input("AVBD-2345", username))
			resultados[i] = resultado{user: user, err: err}
		}()
	}

	listos.Wait()
	close(largada)
	wg.Wait()

	var ganadores, perdedores int
	var ganador accounts.User
	for _, res := range resultados {
		switch {
		case res.err == nil:
			ganadores++
			ganador = res.user
		case faultCode(res.err) == contract.CodeCodeAlreadyRedeemed:
			perdedores++
		default:
			t.Errorf("error inesperado en el canje: %v", res.err)
		}
	}

	if ganadores != 1 || perdedores != 1 {
		t.Fatalf("ganaron %d y perdieron %d; se esperaba exactamente uno de cada", ganadores, perdedores)
	}

	// Y del lado de la base: un solo usuario, un solo movimiento, un solo canje.
	var users, movimientos int
	if err := pool.QueryRow(ctx, `select count(*) from users where role = 'student'`).Scan(&users); err != nil {
		t.Fatal(err)
	}
	if users != 1 {
		t.Errorf("quedaron %d alumnos creados; el que perdió no tenía que quedar", users)
	}
	if err := pool.QueryRow(ctx, `select count(*) from coin_transactions`).Scan(&movimientos); err != nil {
		t.Fatal(err)
	}
	if movimientos != 1 {
		t.Errorf("hay %d movimientos; las 1000 monedas se acreditaron más de una vez", movimientos)
	}

	var redeemedBy string
	if err := pool.QueryRow(ctx,
		`select redeemed_by::text from invite_codes where code = 'AVBD-2345'`).Scan(&redeemedBy); err != nil {
		t.Fatal(err)
	}
	if redeemedBy != ganador.ID {
		t.Errorf("el código quedó a nombre de %q y ganó %q", redeemedBy, ganador.ID)
	}

	testdb.Reconcile(t, pool)
}

// Diez a la vez, para que no sea una casualidad de dos.
func TestRedeemConcurrenteConDiezIntentos(t *testing.T) {
	pool := testdb.Pool(t)
	store := accounts.New(pool)
	ctx := context.Background()

	admin := testdb.InsertUser(t, pool, "profe", accounts.RoleAdmin)
	testdb.InsertCode(t, pool, "KMPR-8827", 1000, admin)

	const intentos = 10
	largada := make(chan struct{})
	var listos, wg sync.WaitGroup
	listos.Add(intentos)

	var mu sync.Mutex
	var ganadores, yaCanjeado int

	for i := range intentos {
		wg.Add(1)
		go func() {
			defer wg.Done()
			listos.Done()
			<-largada

			_, err := store.Redeem(ctx, input("KMPR-8827", "alumno"+string(rune('a'+i))))

			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				ganadores++
			case faultCode(err) == contract.CodeCodeAlreadyRedeemed:
				yaCanjeado++
			default:
				t.Errorf("error inesperado: %v", err)
			}
		}()
	}

	listos.Wait()
	close(largada)
	wg.Wait()

	if ganadores != 1 || yaCanjeado != intentos-1 {
		t.Errorf("ganaron %d y rebotaron %d; se esperaba 1 y %d", ganadores, yaCanjeado, intentos-1)
	}
	testdb.Reconcile(t, pool)
}

// El caso que justifica que el canje sea UNA transacción: si el usuario no se
// puede crear, el código **no se quema**. Un código quemado sin usuario detrás
// deja al alumno sin poder registrarse y sin nada que se pueda arreglar salvo a
// mano en la base, en medio de una clase.
func TestRedeemConUsuarioOcupadoNoQuemaElCodigo(t *testing.T) {
	pool := testdb.Pool(t)
	store := accounts.New(pool)
	ctx := context.Background()

	admin := testdb.InsertUser(t, pool, "profe", accounts.RoleAdmin)
	testdb.InsertUser(t, pool, "anag", accounts.RoleStudent)
	testdb.InsertCode(t, pool, "AVBD-2345", 1000, admin)

	_, err := store.Redeem(ctx, input("AVBD-2345", "anag"))
	if got := faultCode(err); got != contract.CodeUsernameTaken {
		t.Fatalf("código de error = %q, se esperaba USERNAME_TAKEN (err: %v)", got, err)
	}

	var redeemedBy *string
	if err := pool.QueryRow(ctx,
		`select redeemed_by::text from invite_codes where code = 'AVBD-2345'`).Scan(&redeemedBy); err != nil {
		t.Fatal(err)
	}
	if redeemedBy != nil {
		t.Error("el código quedó canjeado aunque el usuario no se creó")
	}

	// Y el código sigue sirviendo con otro nombre.
	user, err := store.Redeem(ctx, input("AVBD-2345", "anagomez"))
	if err != nil {
		t.Fatalf("el código tendría que seguir libre: %v", err)
	}
	if user.Balance != 1000 {
		t.Errorf("saldo = %d, se esperaba 1000", user.Balance)
	}
	testdb.Reconcile(t, pool)
}

// El login es case-insensitive, así que el nombre ocupado también lo es.
func TestRedeemRechazaUsuarioOcupadoConOtraCaja(t *testing.T) {
	pool := testdb.Pool(t)
	store := accounts.New(pool)
	ctx := context.Background()

	admin := testdb.InsertUser(t, pool, "profe", accounts.RoleAdmin)
	testdb.InsertUser(t, pool, "anag", accounts.RoleStudent)
	testdb.InsertCode(t, pool, "AVBD-2345", 1000, admin)

	_, err := store.Redeem(ctx, input("AVBD-2345", "AnaG"))
	if got := faultCode(err); got != contract.CodeUsernameTaken {
		t.Errorf("código de error = %q, se esperaba USERNAME_TAKEN", got)
	}
}

func TestRedeemDistingueCodigoInexistenteDeCanjeado(t *testing.T) {
	pool := testdb.Pool(t)
	store := accounts.New(pool)
	ctx := context.Background()

	admin := testdb.InsertUser(t, pool, "profe", accounts.RoleAdmin)
	testdb.InsertCode(t, pool, "AVBD-2345", 1000, admin)

	if _, err := store.Redeem(ctx, input("ZZZZ-9999", "otro")); faultCode(err) != contract.CodeCodeNotFound {
		t.Errorf("un código inexistente dio %q, se esperaba CODE_NOT_FOUND", faultCode(err))
	}
	if _, err := store.Redeem(ctx, input("AVBD-2345", "anag")); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Redeem(ctx, input("AVBD-2345", "brunod")); faultCode(err) != contract.CodeCodeAlreadyRedeemed {
		t.Errorf("un código canjeado dio %q, se esperaba CODE_ALREADY_REDEEMED", faultCode(err))
	}
}

// El código puede acreditar algo distinto de 1000: es una columna del código, no
// una constante del backend.
func TestRedeemAcreditaLoQueDiceElCodigo(t *testing.T) {
	pool := testdb.Pool(t)
	store := accounts.New(pool)
	ctx := context.Background()

	admin := testdb.InsertUser(t, pool, "profe", accounts.RoleAdmin)
	testdb.InsertCode(t, pool, "AVBD-2345", 2500, admin)

	user, err := store.Redeem(ctx, input("AVBD-2345", "anag"))
	if err != nil {
		t.Fatal(err)
	}
	if user.Balance != 2500 {
		t.Errorf("saldo = %d, se esperaba 2500", user.Balance)
	}
	testdb.Reconcile(t, pool)
}

// ── check-code ────────────────────────────────────────────────────────────

func TestCheckCodeNoCanjea(t *testing.T) {
	pool := testdb.Pool(t)
	store := accounts.New(pool)
	ctx := context.Background()

	admin := testdb.InsertUser(t, pool, "profe", accounts.RoleAdmin)
	testdb.InsertCode(t, pool, "AVBD-2345", 1000, admin)

	coins, err := store.CheckCode(ctx, "AVBD-2345")
	if err != nil {
		t.Fatalf("CheckCode: %v", err)
	}
	if coins != 1000 {
		t.Errorf("coins = %d, se esperaba 1000", coins)
	}

	// Mirarlo no lo gasta: se puede seguir canjeando.
	if _, err := store.Redeem(ctx, input("AVBD-2345", "anag")); err != nil {
		t.Fatalf("el código tendría que seguir libre después de CheckCode: %v", err)
	}
	if _, err := store.CheckCode(ctx, "AVBD-2345"); faultCode(err) != contract.CodeCodeAlreadyRedeemed {
		t.Errorf("después del canje, CheckCode dio %q", faultCode(err))
	}
	if _, err := store.CheckCode(ctx, "ZZZZ-9999"); faultCode(err) != contract.CodeCodeNotFound {
		t.Errorf("un código inexistente dio %q", faultCode(err))
	}
}

// ── Códigos del instructor ────────────────────────────────────────────────

func TestCreateCodesGeneraUnLoteSinRepetir(t *testing.T) {
	pool := testdb.Pool(t)
	store := accounts.New(pool)
	ctx := context.Background()

	admin := testdb.InsertUser(t, pool, "profe", accounts.RoleAdmin)

	codes, err := store.CreateCodes(ctx, admin, 25, 1000, "grupo del martes")
	if err != nil {
		t.Fatalf("CreateCodes: %v", err)
	}
	if len(codes) != 25 {
		t.Fatalf("len(codes) = %d, se esperaban 25", len(codes))
	}

	vistos := map[string]bool{}
	for _, code := range codes {
		if vistos[code] {
			t.Errorf("código repetido en el lote: %s", code)
		}
		vistos[code] = true
	}

	items, err := store.ListCodes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 25 {
		t.Errorf("ListCodes devolvió %d, se esperaban 25", len(items))
	}
	for _, item := range items {
		if item.Redeemed {
			t.Errorf("%s aparece canjeado recién creado", item.Code)
		}
		if item.Note != "grupo del martes" {
			t.Errorf("nota = %q", item.Note)
		}
		if item.CoinsGranted != 1000 {
			t.Errorf("coinsGranted = %d", item.CoinsGranted)
		}
	}
}

func TestCreateCodesRechazaLotesAbsurdos(t *testing.T) {
	pool := testdb.Pool(t)
	store := accounts.New(pool)
	ctx := context.Background()

	admin := testdb.InsertUser(t, pool, "profe", accounts.RoleAdmin)

	for _, count := range []int{0, -1, accounts.MaxCodesPerBatch + 1} {
		if _, err := store.CreateCodes(ctx, admin, count, 1000, ""); faultCode(err) != contract.CodeValidationFailed {
			t.Errorf("count=%d dio %q, se esperaba VALIDATION_FAILED", count, faultCode(err))
		}
	}
	if _, err := store.CreateCodes(ctx, admin, 1, 0, ""); faultCode(err) != contract.CodeValidationFailed {
		t.Errorf("coins=0 dio %q, se esperaba VALIDATION_FAILED", faultCode(err))
	}
}

func TestListCodesMuestraQuienLoUso(t *testing.T) {
	pool := testdb.Pool(t)
	store := accounts.New(pool)
	ctx := context.Background()

	admin := testdb.InsertUser(t, pool, "profe", accounts.RoleAdmin)
	testdb.InsertCode(t, pool, "AVBD-2345", 1000, admin)
	testdb.InsertCode(t, pool, "KMPR-8827", 1000, admin)

	user, err := store.Redeem(ctx, input("AVBD-2345", "anag"))
	if err != nil {
		t.Fatal(err)
	}

	items, err := store.ListCodes(ctx)
	if err != nil {
		t.Fatal(err)
	}

	porCodigo := map[string]accounts.Code{}
	for _, item := range items {
		porCodigo[item.Code] = item
	}

	usado := porCodigo["AVBD-2345"]
	if !usado.Redeemed || usado.RedeemedBy != "anag" || usado.RedeemedByID != user.ID {
		t.Errorf("el código usado no dice quién lo usó: %+v", usado)
	}
	if usado.RedeemedAt == nil {
		t.Error("falta redeemedAt en el código usado")
	}
	if libre := porCodigo["KMPR-8827"]; libre.Redeemed || libre.RedeemedBy != "" {
		t.Errorf("el código libre aparece usado: %+v", libre)
	}

	// Los sin canjear van primero: es lo que se busca para dictarle uno a quien
	// llegó tarde.
	if items[0].Redeemed {
		t.Error("los códigos libres tendrían que aparecer primero")
	}
}

// ── Login ─────────────────────────────────────────────────────────────────

func TestAuthenticate(t *testing.T) {
	pool := testdb.Pool(t)
	store := accounts.New(pool)
	ctx := context.Background()

	admin := testdb.InsertUser(t, pool, "profe", accounts.RoleAdmin)
	testdb.InsertCode(t, pool, "AVBD-2345", 1000, admin)

	hash, err := auth.HashPassword("caballo-de-batalla")
	if err != nil {
		t.Fatal(err)
	}
	in := input("AVBD-2345", "anag")
	in.PasswordHash = hash
	if _, err := store.Redeem(ctx, in); err != nil {
		t.Fatal(err)
	}

	verify := func(password string) func(string) bool {
		return func(hash string) bool { return auth.VerifyPassword(password, hash) }
	}

	user, err := store.Authenticate(ctx, "anag", verify("caballo-de-batalla"))
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if user.Username != "anag" || user.Balance != 1000 {
		t.Errorf("usuario inesperado: %+v", user)
	}

	// El login no distingue mayúsculas: nadie recuerda cómo se registró.
	if _, err := store.Authenticate(ctx, "AnaG", verify("caballo-de-batalla")); err != nil {
		t.Errorf("el login tendría que ser case-insensitive: %v", err)
	}

	// Contraseña mal y usuario inexistente dan el MISMO error: distinguirlos sería
	// publicar la lista de quién se registró.
	malPass := faultCode(store2err(store.Authenticate(ctx, "anag", verify("otra"))))
	noExiste := faultCode(store2err(store.Authenticate(ctx, "nadie", verify("otra"))))
	if malPass != contract.CodeInvalidCredentials || noExiste != contract.CodeInvalidCredentials {
		t.Errorf("los errores difieren: contraseña %q, inexistente %q", malPass, noExiste)
	}
}

func store2err(_ accounts.User, err error) error { return err }

// ── Sesiones ──────────────────────────────────────────────────────────────

func TestRefreshTokenEsDeUnSoloUso(t *testing.T) {
	pool := testdb.Pool(t)
	store := accounts.New(pool)
	ctx := context.Background()

	ana := testdb.InsertUser(t, pool, "ana", accounts.RoleStudent)
	now := time.Now()

	hash := auth.HashToken("rt_uno")
	if err := store.SaveRefreshToken(ctx, hash, ana, now, now.Add(auth.RefreshTokenTTL)); err != nil {
		t.Fatalf("SaveRefreshToken: %v", err)
	}

	userID, err := store.ConsumeRefreshToken(ctx, hash, now)
	if err != nil {
		t.Fatalf("primer canje: %v", err)
	}
	if userID != ana {
		t.Errorf("userID = %q, se esperaba %q", userID, ana)
	}

	// El segundo canje del mismo token es reuso.
	if _, err := store.ConsumeRefreshToken(ctx, hash, now); !errors.Is(err, accounts.ErrTokenReused) {
		t.Errorf("err = %v, se esperaba ErrTokenReused", err)
	}
}

// Reusar un token es la señal de que alguien tiene una copia. La respuesta no es
// solo rechazarlo: se cortan TODAS las sesiones del dueño, porque no se sabe cuál
// de las dos copias es la legítima.
func TestRefreshReusadoRevocaTodaLaFamilia(t *testing.T) {
	pool := testdb.Pool(t)
	store := accounts.New(pool)
	ctx := context.Background()

	ana := testdb.InsertUser(t, pool, "ana", accounts.RoleStudent)
	now := time.Now()

	robado := auth.HashToken("rt_robado")
	otra := auth.HashToken("rt_otra-pestaña")
	for _, hash := range []string{robado, otra} {
		if err := store.SaveRefreshToken(ctx, hash, ana, now, now.Add(auth.RefreshTokenTTL)); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := store.ConsumeRefreshToken(ctx, robado, now); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ConsumeRefreshToken(ctx, robado, now); !errors.Is(err, accounts.ErrTokenReused) {
		t.Fatalf("err = %v, se esperaba ErrTokenReused", err)
	}

	// La otra pestaña también quedó afuera.
	if _, err := store.ConsumeRefreshToken(ctx, otra, now); !errors.Is(err, accounts.ErrNotFound) {
		t.Errorf("la otra sesión sigue viva: %v", err)
	}
}

func TestRefreshTokenVencidoNoSirve(t *testing.T) {
	pool := testdb.Pool(t)
	store := accounts.New(pool)
	ctx := context.Background()

	ana := testdb.InsertUser(t, pool, "ana", accounts.RoleStudent)
	now := time.Now()

	hash := auth.HashToken("rt_viejo")
	if err := store.SaveRefreshToken(ctx, hash, ana, now.Add(-2*auth.RefreshTokenTTL), now.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}

	if _, err := store.ConsumeRefreshToken(ctx, hash, now); !errors.Is(err, accounts.ErrNotFound) {
		t.Errorf("err = %v, se esperaba ErrNotFound", err)
	}
}

func TestRevokeYPurgeRefreshTokens(t *testing.T) {
	pool := testdb.Pool(t)
	store := accounts.New(pool)
	ctx := context.Background()

	ana := testdb.InsertUser(t, pool, "ana", accounts.RoleStudent)
	now := time.Now()

	vivo := auth.HashToken("rt_vivo")
	vencido := auth.HashToken("rt_vencido")
	if err := store.SaveRefreshToken(ctx, vivo, ana, now, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveRefreshToken(ctx, vencido, ana, now.Add(-2*time.Hour), now.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}

	removed, err := store.PurgeRefreshTokens(ctx, now)
	if err != nil {
		t.Fatal(err)
	}
	if removed != 1 {
		t.Errorf("se borraron %d, se esperaba 1 (solo el vencido)", removed)
	}
	if _, err := store.ConsumeRefreshToken(ctx, vivo, now); err != nil {
		t.Errorf("se borró el token que seguía vivo: %v", err)
	}

	if err := store.RevokeRefreshTokens(ctx, ana); err != nil {
		t.Fatal(err)
	}
	var quedan int
	if err := pool.QueryRow(ctx, `select count(*) from refresh_tokens where user_id = $1`, ana).Scan(&quedan); err != nil {
		t.Fatal(err)
	}
	if quedan != 0 {
		t.Errorf("quedaron %d sesiones después del logout", quedan)
	}
}

// ── Puntos ────────────────────────────────────────────────────────────────

// La fórmula de decisiones.md §1, leída de la vista `user_scores`:
//
//	puntos = max(10, floor(monedas / 100)) + puntos regalados
//
// **El piso de 10 es la mitad de la economía.** Apostar mal saca monedas —y con
// eso, capacidad de seguir jugando— pero nunca calificación. Va junto con el
// pari-mutuel: como el pool es suma cero, sin piso la nota que gana un alumno
// saldría de la de otro.
//
// Se lee de la vista y no se calcula en Go, así que lo único que hay acá son los
// valores esperados. Si el piso se cayera del esquema, estos tres casos lo dicen.
func TestPointsAplicaElPisoDeDiez(t *testing.T) {
	pool := testdb.Pool(t)
	store := accounts.New(pool)
	ledgerStore := ledger.New(pool)
	ctx := context.Background()

	admin := testdb.InsertUser(t, pool, "profe", accounts.RoleAdmin)
	testdb.InsertCode(t, pool, "AVBD-2345", 1000, admin)

	ana, err := store.Redeem(ctx, input("AVBD-2345", "anag"))
	if err != nil {
		t.Fatal(err)
	}

	casos := []struct {
		gasto  int64
		saldo  int64
		puntos accounts.Points
		porque string
	}{
		{0, 1000, 1000, "1000 monedas → 10 puntos"},
		{500, 500, 1000, "500 monedas → 10 puntos: sin el piso serían 5"},
		{500, 0, 1000, "fundida → 10 puntos. Apostar mal no baja la nota"},
	}

	for _, caso := range casos {
		if caso.gasto > 0 {
			if _, err := ledgerStore.Move(ctx, ledger.Movement{
				UserID: ana.ID, Delta: -caso.gasto, Reason: ledger.ReasonBetPlaced,
			}); err != nil {
				t.Fatal(err)
			}
		}
		if got := testdb.Balance(t, pool, ana.ID); got != caso.saldo {
			t.Fatalf("saldo = %d, se esperaba %d", got, caso.saldo)
		}

		points, err := store.PointsFor(ctx, ana.ID)
		if err != nil {
			t.Fatal(err)
		}
		if points != caso.puntos {
			t.Errorf("%s: puntos = %s, se esperaba %s", caso.porque, points, caso.puntos)
		}
	}

	// Ganar sí sube.
	if _, err := ledgerStore.Move(ctx, ledger.Movement{
		UserID: ana.ID, Delta: 1500, Reason: ledger.ReasonBetWon,
	}); err != nil {
		t.Fatal(err)
	}
	points, err := store.PointsFor(ctx, ana.ID)
	if err != nil {
		t.Fatal(err)
	}
	if points != 1500 {
		t.Errorf("con 1500 monedas los puntos dan %s, se esperaba 15", points)
	}
	testdb.Reconcile(t, pool)
}

// Los puntos regalados se suman aparte y NO pasan por el juego: no se pueden
// perder apostando. Por eso viven en `point_grants` y no en el ledger.
func TestGrantPointsSeSumaYNoSePierdeApostando(t *testing.T) {
	pool := testdb.Pool(t)
	store := accounts.New(pool)
	ledgerStore := ledger.New(pool)
	ctx := context.Background()

	admin := testdb.InsertUser(t, pool, "profe", accounts.RoleAdmin)
	testdb.InsertCode(t, pool, "AVBD-2345", 1000, admin)

	ana, err := store.Redeem(ctx, input("AVBD-2345", "anag"))
	if err != nil {
		t.Fatal(err)
	}

	// 2,5 puntos = 250 centésimas.
	grant, err := store.GrantPoints(ctx, admin, ana.ID, 250, "Explicó @for en el code review")
	if err != nil {
		t.Fatalf("GrantPoints: %v", err)
	}
	if grant.Points != 250 {
		t.Errorf("grant.Points = %s, se esperaba 2.5", grant.Points)
	}

	points, err := store.PointsFor(ctx, ana.ID)
	if err != nil {
		t.Fatal(err)
	}
	if points != 1250 {
		t.Errorf("puntos = %s, se esperaba 12.5 (10 del saldo + 2,5 regalados)", points)
	}

	// Funde el saldo. **Es acá donde se ve que los dos regalos no son lo mismo:**
	//
	//	los 10 puntos del saldo se sostienen en el piso → apostar mal no baja la nota
	//	los 2,5 regalados se suman aparte               → y no pasan por el juego
	//
	// Lo segundo es lo que garantiza que `point_grants` sea otra tabla y no otro
	// motivo del ledger. Si el regalo hubiera entrado como monedas, habría quedado
	// sujeto al piso en lugar de sumarse por encima.
	if _, err := ledgerStore.Move(ctx, ledger.Movement{
		UserID: ana.ID, Delta: -1000, Reason: ledger.ReasonBetPlaced,
	}); err != nil {
		t.Fatal(err)
	}
	points, err = store.PointsFor(ctx, ana.ID)
	if err != nil {
		t.Fatal(err)
	}
	if points != 1250 {
		t.Errorf("después de fundirse los puntos dan %s, se esperaba 12.5: "+
			"el piso sostiene los 10 y el regalo se suma encima", points)
	}

	// point_grants es append-only, igual que el ledger.
	if _, err := pool.Exec(ctx, `update point_grants set points = 100 where id = $1`, grant.ID); err == nil {
		t.Error("se pudo editar un regalo de puntos")
	}
	testdb.Reconcile(t, pool)
}

func TestGrantPointsValida(t *testing.T) {
	pool := testdb.Pool(t)
	store := accounts.New(pool)
	ctx := context.Background()

	admin := testdb.InsertUser(t, pool, "profe", accounts.RoleAdmin)
	ana := testdb.InsertUser(t, pool, "ana", accounts.RoleStudent)

	if _, err := store.GrantPoints(ctx, admin, ana, 0, "nada"); faultCode(err) != contract.CodeValidationFailed {
		t.Errorf("cero puntos dio %q", faultCode(err))
	}
	if _, err := store.GrantPoints(ctx, admin, ana, 100, ""); faultCode(err) != contract.CodeValidationFailed {
		t.Errorf("sin motivo dio %q", faultCode(err))
	}
	if _, err := store.GrantPoints(ctx, admin, "99999999-9999-9999-9999-999999999999", 100, "x"); faultCode(err) != contract.CodeUserNotFound {
		t.Errorf("un usuario inexistente dio %q, se esperaba USER_NOT_FOUND", faultCode(err))
	}
}

// ── Panel de nota ─────────────────────────────────────────────────────────

func TestScoresNoIncluyeAlInstructorYOrdenaPorPuntos(t *testing.T) {
	pool := testdb.Pool(t)
	store := accounts.New(pool)
	ledgerStore := ledger.New(pool)
	ctx := context.Background()

	admin := testdb.InsertUser(t, pool, "profe", accounts.RoleAdmin)
	for i, code := range []string{"AVBD-2345", "KMPR-8827"} {
		testdb.InsertCode(t, pool, code, 1000, admin)
		user, err := store.Redeem(ctx, input(code, []string{"anag", "brunod"}[i]))
		if err != nil {
			t.Fatal(err)
		}
		// Bruno gana 1500 más: tiene que quedar primero.
		if i == 1 {
			if _, err := ledgerStore.Move(ctx, ledger.Movement{
				UserID: user.ID, Delta: 1500, Reason: ledger.ReasonBetWon,
			}); err != nil {
				t.Fatal(err)
			}
		}
	}

	items, err := store.Scores(ctx)
	if err != nil {
		t.Fatalf("Scores: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, se esperaban 2 alumnos (el instructor no va)", len(items))
	}
	for _, item := range items {
		if item.Username == "profe" {
			t.Error("el instructor aparece en su propio panel de nota")
		}
	}
	if items[0].Username != "brunod" {
		t.Errorf("primero = %q, se esperaba brunod (2500 monedas)", items[0].Username)
	}
	if items[0].Points != 2500 || items[0].Balance != 2500 {
		t.Errorf("bruno: puntos %s, saldo %d; se esperaban 25 y 2500", items[0].Points, items[0].Balance)
	}
	if items[1].Points != 1000 {
		t.Errorf("ana: puntos %s, se esperaba 10", items[1].Points)
	}

	// El instructor no tiene nota, y pedírsela no es un error.
	points, err := store.PointsFor(ctx, admin)
	if err != nil {
		t.Fatalf("PointsFor del instructor: %v", err)
	}
	if points != 0 {
		t.Errorf("el instructor tiene %s puntos, se esperaba 0", points)
	}
	testdb.Reconcile(t, pool)
}

// ── EnsureAdmin ───────────────────────────────────────────────────────────

// Sin admin en la base no hay quien genere el primer código, y no hay registro
// abierto por el que llegue uno. Tiene que poder correrse en cada arranque.
func TestEnsureAdminEsIdempotenteYReescribeLaContrasena(t *testing.T) {
	pool := testdb.Pool(t)
	store := accounts.New(pool)
	ctx := context.Background()

	primero, err := store.EnsureAdmin(ctx, "profe", "Jhonatan", "Soto", "hash-uno")
	if err != nil {
		t.Fatalf("EnsureAdmin: %v", err)
	}
	if primero.Role != accounts.RoleAdmin {
		t.Errorf("rol = %q, se esperaba admin", primero.Role)
	}

	segundo, err := store.EnsureAdmin(ctx, "profe", "Jhonatan", "Soto", "hash-dos")
	if err != nil {
		t.Fatalf("segundo EnsureAdmin: %v", err)
	}
	if segundo.ID != primero.ID {
		t.Error("el segundo arranque creó otro instructor en vez de actualizar el que había")
	}

	var hash string
	if err := pool.QueryRow(ctx, `select password_hash from users where id = $1`, primero.ID).Scan(&hash); err != nil {
		t.Fatal(err)
	}
	if hash != "hash-dos" {
		t.Errorf("hash = %q; reescribir la contraseña es la forma de recuperar el acceso", hash)
	}

	var admins int
	if err := pool.QueryRow(ctx, `select count(*) from users where role = 'admin'`).Scan(&admins); err != nil {
		t.Fatal(err)
	}
	if admins != 1 {
		t.Errorf("hay %d instructores, se esperaba 1", admins)
	}
}

// Un alumno que ya existe con ese nombre pasa a admin. Es intencional: es cómo se
// asciende a alguien sin entrar a la base.
func TestEnsureAdminAsciendeUnUsuarioExistente(t *testing.T) {
	pool := testdb.Pool(t)
	store := accounts.New(pool)
	ctx := context.Background()

	alumno := testdb.InsertUser(t, pool, "profe", accounts.RoleStudent)
	admin, err := store.EnsureAdmin(ctx, "profe", "Jhonatan", "Soto", "hash")
	if err != nil {
		t.Fatal(err)
	}
	if admin.ID != alumno {
		t.Error("se creó otro usuario en vez de ascender al que estaba")
	}
	if admin.Role != accounts.RoleAdmin {
		t.Errorf("rol = %q, se esperaba admin", admin.Role)
	}
}
