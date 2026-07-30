package accounts

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/talentodh/arena/internal/contract"
	"github.com/talentodh/arena/internal/ledger"
)

// uniqueViolation es el SQLSTATE 23505. Se usa para traducir el índice único de
// `lower(username)` a USERNAME_TAKEN sin tener que consultar antes: preguntar
// «¿está libre?» y después insertar es una carrera de datos, insertar y mirar el
// error no lo es.
const (
	uniqueViolation     = "23505"
	foreignKeyViolation = "23503"
)

func isUniqueViolation(err error) bool { return hasSQLState(err, uniqueViolation) }

// isForeignKeyViolation atrapa un id de usuario que no existe. La FK es la que
// se da cuenta, y traducirla es lo que evita un 500 cuando el instructor pega
// mal un uuid.
func isForeignKeyViolation(err error) bool { return hasSQLState(err, foreignKeyViolation) }

func hasSQLState(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}

// ByID devuelve el usuario, o ErrNotFound.
func (s *Store) ByID(ctx context.Context, userID string) (User, error) {
	user, err := scanUser(s.Pool.QueryRow(ctx,
		`select `+userColumns+` from users where id = $1`, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("leyendo el usuario %s: %w", userID, err)
	}
	return user, nil
}

// Authenticate valida usuario y contraseña.
//
// Recibe la verificación como función para no arrastrar el paquete auth hasta
// acá: el store no tiene por qué saber cómo se hashea una contraseña.
//
// Devuelve el MISMO error si el usuario no existe y si la contraseña está mal.
// Distinguirlos le diría a cualquiera qué usuarios existen, y con nombres de
// usuario cortos como los de una clase eso es media lista de asistencia.
func (s *Store) Authenticate(ctx context.Context, username string, verify func(hash string) bool) (User, error) {
	var user User
	var hash string

	// Búsqueda por `lower(username)`: es el índice único, y nadie va a recordar
	// si se registró como «Ana» o «ana».
	err := s.Pool.QueryRow(ctx,
		`select `+userColumns+`, password_hash from users where lower(username) = lower($1)`,
		username,
	).Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName,
		&user.Role, &user.Balance, &user.CreatedAt, &hash)

	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, contract.Errorf(contract.CodeInvalidCredentials)
	}
	if err != nil {
		return User{}, fmt.Errorf("buscando el usuario %q: %w", username, err)
	}
	if !verify(hash) {
		return User{}, contract.Errorf(contract.CodeInvalidCredentials)
	}
	return user, nil
}

// RedeemInput es un canje. La contraseña llega ya hasheada: el hash es asunto
// del paquete auth y el store no debería poder ver una contraseña en claro.
type RedeemInput struct {
	Code         string
	FirstName    string
	LastName     string
	Username     string
	PasswordHash string
}

// Redeem canjea el código y crea la cuenta. **Es UNA transacción**: usuario +
// código marcado + monedas acreditadas, o nada (decisiones.md §2).
//
// Un código a medio canjear —usuario creado sin monedas, o código quemado sin
// usuario— es el peor estado posible: el alumno no puede registrarse y el código
// ya no sirve, y arreglarlo a mano es entrar a la base en medio de una clase.
//
// **Un código, un uso**, y eso lo sostiene el `for update` de abajo, no una
// consulta previa. Si dos personas mandan el mismo código en el mismo instante,
// la segunda transacción se queda esperando el lock de la fila; cuando la
// primera confirma, la segunda despierta, vuelve a leer la fila —en READ
// COMMITTED, `for update` re-evalúa contra la versión ya confirmada— y ve el
// código canjeado. Una gana y la otra recibe CODE_ALREADY_REDEEMED.
//
// Un `select` sin lock y después un `insert` no daría eso: las dos leerían el
// código libre y las dos seguirían.
func (s *Store) Redeem(ctx context.Context, in RedeemInput) (User, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return User{}, fmt.Errorf("abriendo la transacción del canje: %w", err)
	}
	// Rollback sobre una transacción ya confirmada no hace nada, así que el defer
	// es seguro y cubre todos los returns de error de abajo.
	defer tx.Rollback(ctx)

	// 1. Tomar el código y bloquear su fila.
	var coins int64
	var redeemedBy *string
	err = tx.QueryRow(ctx,
		`select coins_granted, redeemed_by::text from invite_codes where code = $1 for update`,
		in.Code,
	).Scan(&coins, &redeemedBy)

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		// Inexistente y ya canjeado son errores DISTINTOS a propósito: el alumno
		// tiene que poder distinguir «lo escribí mal» de «ya me registré».
		return User{}, contract.Errorf(contract.CodeCodeNotFound)
	case err != nil:
		return User{}, fmt.Errorf("leyendo el código %s: %w", in.Code, err)
	case redeemedBy != nil:
		return User{}, contract.Errorf(contract.CodeCodeAlreadyRedeemed)
	}

	// 2. Crear el usuario. El saldo arranca en 0 y lo mueve el ledger: el saldo
	// inicial también es un movimiento y tiene que estar en el historial.
	user, err := scanUser(tx.QueryRow(ctx, `
		insert into users (username, first_name, last_name, password_hash)
		values ($1, $2, $3, $4)
		returning `+userColumns,
		in.Username, in.FirstName, in.LastName, in.PasswordHash))

	switch {
	case isUniqueViolation(err):
		// El único índice único de `users` es lower(username). Al abortar, el
		// código queda SIN canjear: es la parte que hace que valga la transacción.
		return User{}, contract.Errorf(contract.CodeUsernameTaken)
	case err != nil:
		return User{}, fmt.Errorf("creando el usuario %q: %w", in.Username, err)
	}

	// 3. Quemar el código. El `and redeemed_by is null` es la segunda red: si por
	// lo que sea el lock de arriba no hubiera alcanzado, acá se cuentan las filas
	// afectadas y cero significa que ganó otro.
	tag, err := tx.Exec(ctx, `
		update invite_codes
		set redeemed_by = $2, redeemed_at = now()
		where code = $1 and redeemed_by is null`,
		in.Code, user.ID)
	if err != nil {
		return User{}, fmt.Errorf("marcando el código %s: %w", in.Code, err)
	}
	if tag.RowsAffected() != 1 {
		return User{}, contract.Errorf(contract.CodeCodeAlreadyRedeemed)
	}

	// 4. Acreditar, en la MISMA transacción. `ref_id` es el código, así que el
	// primer movimiento del historial dice de dónde salieron las monedas.
	balance, err := ledger.Move(ctx, tx, ledger.Movement{
		UserID: user.ID,
		Delta:  coins,
		Reason: ledger.ReasonCodeRedeemed,
		RefID:  in.Code,
	})
	if err != nil {
		return User{}, err
	}
	user.Balance = balance

	if err := tx.Commit(ctx); err != nil {
		return User{}, fmt.Errorf("confirmando el canje de %s: %w", in.Code, err)
	}
	return user, nil
}

// EnsureAdmin crea o actualiza al instructor.
//
// Hace falta porque `invite_codes.created_by` es NOT NULL y referencia a un
// usuario: sin un admin en la base no hay forma de generar el primer código, y
// no hay registro abierto por el que llegue uno. Se resuelve con variables de
// entorno en el arranque en vez de un `insert` a mano en producción.
//
// Idempotente y **reescribe la contraseña**: es también la forma de recuperar el
// acceso si el instructor la olvida a mitad de la cursada.
func (s *Store) EnsureAdmin(ctx context.Context, username, firstName, lastName, passwordHash string) (User, error) {
	user, err := scanUser(s.Pool.QueryRow(ctx, `
		insert into users (username, first_name, last_name, password_hash, role)
		values ($1, $2, $3, $4, 'admin')
		on conflict (lower(username)) do update
		set password_hash = excluded.password_hash,
		    first_name    = excluded.first_name,
		    last_name     = excluded.last_name,
		    role          = 'admin'
		returning `+userColumns,
		username, firstName, lastName, passwordHash))
	if err != nil {
		return User{}, fmt.Errorf("asegurando el instructor %q: %w", username, err)
	}
	return user, nil
}
