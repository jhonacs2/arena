// Package racesdb es la implementación Postgres de races.Store.
//
// Es el único paquete del dominio de carreras que conoce pgx. El dominio
// (internal/races) habla con la interfaz, así que sus tests corren sin base:
// probar «cancelar devuelve exactamente lo apostado» no puede costar levantar
// Postgres, o ese test no se corre en cada cambio.
//
// Dos reglas de escritura, las dos por el mismo motivo:
//
//   - todos los uuid se leen con `::text` y se pasan con `$n::uuid`. Así no hay
//     que adivinar cómo mapea pgx el tipo uuid, y un id basura del cliente da un
//     «no existe» en vez de un 500;
//   - los montos son bigint y se leen en int64. Nunca float — ver
//     arena/CLAUDE.md §5.
package racesdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/talentodh/arena/internal/races"
	"github.com/talentodh/arena/internal/sim"
)

// Pool es lo que este paquete necesita de la conexión. Lo satisface
// *pgxpool.Pool, así que el cableado le pasa el pool y nada más.
type Pool interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Movement es un movimiento del ledger, en los términos del ledger.
type Movement struct {
	UserID string
	Delta  int64
	Reason string
	RefID  string
	// CreatedBy es el instructor cuando la acción fue suya; vacío cuando el
	// movimiento lo generó el sistema al liquidar una carrera.
	CreatedBy string
}

// Ledger es lo que este paquete necesita del ledger de monedas.
//
// La declara el consumidor y la implementa el paquete del ledger. Recibe la
// `pgx.Tx` en curso, y eso NO es un detalle: es lo que hace que el descuento de
// una apuesta y la apuesta misma caigan juntos o no caigan. Si el ledger abriera
// su propia transacción, existiría un instante en el que a alguien le falta el
// dinero y no tiene la apuesta.
//
// El ledger es append-only y mantiene users.balance en la misma transacción.
// Devuelve error si el saldo quedaría negativo: no confía en su llamador, y hace
// bien.
type Ledger interface {
	MoveTx(ctx context.Context, tx pgx.Tx, m Movement) (balanceAfter int64, err error)
}

// LedgerFunc adapta una función a la interfaz Ledger.
//
// Existe porque `ledger.Movement` y `racesdb.Movement` tienen la MISMA forma pero
// son tipos distintos, y una interfaz de Go no acepta un método cuyo parámetro
// difiere. La conversión de struct es de una línea, y tenerla explícita en el
// cableado es mejor que hacer que uno de los dos paquetes importe al otro solo
// para compartir cuatro campos:
//
//	racesdb.New(pool, racesdb.LedgerFunc(
//		func(ctx context.Context, tx pgx.Tx, m racesdb.Movement) (int64, error) {
//			return ledgerStore.MoveTx(ctx, tx, ledger.Movement(m))
//		}))
type LedgerFunc func(ctx context.Context, tx pgx.Tx, m Movement) (int64, error)

func (f LedgerFunc) MoveTx(ctx context.Context, tx pgx.Tx, m Movement) (int64, error) {
	return f(ctx, tx, m)
}

// Comprobación en tiempo de compilación. Si al dominio le agregan una operación,
// el build falla ACÁ, con el nombre del método que falta, en vez de fallar en el
// cableado con un mensaje sobre una interfaz.
var (
	_ races.Store = (*Store)(nil)
	_ races.Tx    = (*queries)(nil)
)

// Store implementa races.Store.
type Store struct {
	pool   Pool
	ledger Ledger
}

func New(pool Pool, ledger Ledger) *Store {
	return &Store{pool: pool, ledger: ledger}
}

// InTx abre una transacción, corre fn y confirma. Si fn devuelve error, revierte.
//
// El rollback va en un defer y no en la rama de error: si fn entra en pánico, la
// transacción igual se cierra y la conexión vuelve al pool. Sin eso, un pánico
// en un handler dejaría una transacción abierta tomando bloqueos de fila, y la
// próxima largada se quedaría esperando para siempre.
func (s *Store) InTx(ctx context.Context, fn func(races.Tx) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("abriendo la transacción: %w", err)
	}
	defer func() {
		// Después de un Commit exitoso, Rollback devuelve ErrTxClosed y no hace
		// nada: es seguro llamarlo siempre.
		_ = tx.Rollback(context.WithoutCancel(ctx))
	}()

	if err := fn(&queries{tx: tx, ledger: s.ledger}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("confirmando la transacción: %w", err)
	}
	return nil
}

// queries implementa races.Tx sobre una transacción abierta.
type queries struct {
	tx     pgx.Tx
	ledger Ledger
}

// ── Errores e ids ─────────────────────────────────────────────────────────

// invalidUUID es el código de Postgres para «esto no parece un uuid» (22P02).
const invalidUUID = "22P02"

// notFound traduce el error de pgx al ErrNotFound del dominio.
//
// Un id con forma inválida también es «no existe»: el cliente puede mandar
// cualquier cosa en la ruta, y un 500 por eso sería ruido en los logs y un mensaje
// inútil para el alumno.
func notFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return races.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == invalidUUID {
		return races.ErrNotFound
	}
	return err
}

// checkIDs devuelve ErrNotFound si alguno de los ids no tiene forma de uuid, SIN
// consultar la base.
//
// Que el chequeo esté acá y no solo en notFound no es una optimización, es una
// corrección. Un uuid mal formado hace que Postgres ABORTE LA TRANSACCIÓN ENTERA
// (SQLSTATE 25P02): a partir de ese error, toda consulta siguiente en la misma
// transacción falla con «current transaction is aborted», y hasta el COMMIT se
// convierte en un ROLLBACK. Una petición con un id basura en la ruta pasaría de un
// 404 limpio a un 500 y arrastraría con ella cualquier otra lectura de la misma
// transacción.
//
// Lo descubrió el test de integración contra Postgres de verdad. Con el doble en
// memoria no se ve: en memoria un id basura es simplemente una clave que no está.
func checkIDs(ids ...string) error {
	for _, id := range ids {
		if !looksLikeUUID(id) {
			return races.ErrNotFound
		}
	}
	return nil
}

// looksLikeUUID valida la FORMA: 8-4-4-4-12 hexadecimal.
//
// No valida versión ni variante a propósito. Lo único que hace falta es que
// Postgres pueda convertirlo sin abortar la transacción, y para eso alcanza la
// forma. Escrito a mano en vez de con una dependencia: son doce líneas y una
// dependencia más es una decisión más grande que esto.
func looksLikeUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
			continue
		}
		isHex := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
		if !isHex {
			return false
		}
	}
	return true
}

// ── Carreras ──────────────────────────────────────────────────────────────

// raceColumns es la lista de columnas de una carrera, siempre en el mismo orden.
// Está una sola vez para que scanRace no pueda desalinearse con el SELECT.
const raceColumns = `
	id::text, name, status::text, scheduled_at,
	created_by::text, created_at, opened_at, started_at, finished_at, seed`

func scanRace(row pgx.Row) (races.Race, error) {
	var race races.Race
	var status string
	err := row.Scan(
		&race.ID, &race.Name, &status, &race.ScheduledAt,
		&race.CreatedBy, &race.CreatedAt, &race.OpenedAt, &race.StartedAt, &race.FinishedAt, &race.Seed,
	)
	if err != nil {
		return races.Race{}, notFound(err)
	}
	race.Status = races.Status(status)
	return race, nil
}

func (q *queries) CreateRace(ctx context.Context, in races.NewRace) (races.Race, error) {
	if err := checkIDs(in.CreatedBy); err != nil {
		return races.Race{}, err
	}

	// Nace en `draft`: el default de la columna. No se pasa el estado a propósito
	// — el único camino a otro estado es SetStatus, que pasa por la máquina.
	race, err := scanRace(q.tx.QueryRow(ctx, `
		insert into races (name, scheduled_at, created_by)
		values ($1, $2, $3::uuid)
		returning `+raceColumns,
		in.Name, in.ScheduledAt, in.CreatedBy))
	if err != nil {
		return races.Race{}, err
	}

	if len(in.Horses) > 0 {
		if _, err := q.AddHorses(ctx, race.ID, in.Horses); err != nil {
			return races.Race{}, err
		}
	}
	return race, nil
}

func (q *queries) UpdateRace(ctx context.Context, raceID string, patch races.RacePatch) (races.Race, error) {
	if err := checkIDs(raceID); err != nil {
		return races.Race{}, err
	}

	// coalesce con el parámetro nulo deja la columna como estaba: es lo que hace
	// que «no mandé el campo» y «lo mandé vacío» sean cosas distintas.
	return scanRace(q.tx.QueryRow(ctx, `
		update races
		set name         = coalesce($2, name),
		    scheduled_at = coalesce($3, scheduled_at)
		where id = $1::uuid
		returning `+raceColumns,
		raceID, patch.Name, patch.ScheduledAt))
}

func (q *queries) Race(ctx context.Context, raceID string) (races.Race, error) {
	if err := checkIDs(raceID); err != nil {
		return races.Race{}, err
	}

	return scanRace(q.tx.QueryRow(ctx, `select `+raceColumns+` from races where id = $1::uuid`, raceID))
}

// RaceForUpdate bloquea la fila hasta el fin de la transacción.
//
// `for update` es lo que serializa las transiciones y las apuestas: dos clics en
// «largar» no pueden leer los dos el mismo `open`, y una apuesta no puede colarse
// entre el SELECT y el UPDATE de la largada. El segundo en llegar espera, vuelve a
// leer y ve el estado nuevo.
func (q *queries) RaceForUpdate(ctx context.Context, raceID string) (races.Race, error) {
	if err := checkIDs(raceID); err != nil {
		return races.Race{}, err
	}

	return scanRace(q.tx.QueryRow(ctx, `select `+raceColumns+` from races where id = $1::uuid for update`, raceID))
}

// SetStatus cambia el estado y sella la marca de tiempo que corresponda.
//
// Los booleanos se calculan en Go y viajan como parámetros en vez de comparar el
// estado dentro del SQL: con un solo parámetro de tipo enum usado también en un
// `case when`, Postgres tiene que inferir dos tipos para el mismo parámetro y es
// una fuente de errores que no aporta nada.
//
// `cancelled` también sella finished_at: el esquema no tiene cancelled_at, y una
// carrera cancelada también terminó.
func (q *queries) SetStatus(ctx context.Context, raceID string, status races.Status, at time.Time) error {
	if err := checkIDs(raceID); err != nil {
		return err
	}

	opened := status == races.StatusOpen
	started := status == races.StatusRunning
	ended := status == races.StatusFinished || status == races.StatusCancelled

	_, err := q.tx.Exec(ctx, `
		update races
		set status      = $2::race_status,
		    opened_at   = case when $4 then $3 else opened_at end,
		    started_at  = case when $5 then $3 else started_at end,
		    finished_at = case when $6 then $3 else finished_at end
		where id = $1::uuid`,
		raceID, string(status), at, opened, started, ended)
	return err
}

// SetSeed guarda la semilla. `where seed is null` la hace escribible UNA sola
// vez: la semilla es la prueba de que el resultado es reproducible, y
// sobrescribirla sería cambiar el resultado de una carrera ya corrida.
func (q *queries) SetSeed(ctx context.Context, raceID string, seed int64) error {
	if err := checkIDs(raceID); err != nil {
		return err
	}

	tag, err := q.tx.Exec(ctx, `update races set seed = $2 where id = $1::uuid and seed is null`, raceID, seed)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("la carrera %s ya tenía semilla", raceID)
	}
	return nil
}

func (q *queries) RunningRaces(ctx context.Context) ([]races.Race, error) {
	rows, err := q.tx.Query(ctx, `select `+raceColumns+` from races where status = 'running' order by started_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []races.Race{}
	for rows.Next() {
		race, err := scanRace(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, race)
	}
	return out, rows.Err()
}

// VisibleRaces son las carreras del listado, con los agregados y la apuesta del
// que pregunta ya resueltos.
//
// `status <> 'draft'` salvo que sea el instructor: es la regla de visibilidad de
// decisiones.md §4, y está en el WHERE y no filtrada en Go para que no haya un
// camino en el que una carrera en borrador llegue al alumno.
func (q *queries) VisibleRaces(ctx context.Context, userID string, includeDraft bool) ([]races.Summary, error) {
	if err := checkIDs(userID); err != nil {
		return nil, err
	}

	rows, err := q.tx.Query(ctx, `
		select r.id::text, r.name, r.status::text, r.scheduled_at,
		       r.created_by::text, r.created_at, r.opened_at, r.started_at, r.finished_at, r.seed,
		       (select count(*) from horses h where h.race_id = r.id),
		       (select count(*) from race_participants p where p.race_id = r.id),
		       b.id::text, b.horse_id::text, b.amount,
		       b.status::text, b.payout, b.created_at, b.settled_at
		from races r
		left join bets b on b.race_id = r.id and b.user_id = $1::uuid
		where ($2 or r.status <> 'draft')
		order by coalesce(r.scheduled_at, r.created_at) desc, r.id`,
		userID, includeDraft)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []races.Summary{}
	for rows.Next() {
		var (
			summary   races.Summary
			status    string
			betID     *string
			horseID   *string
			amount    *int64
			betStatus *string
			payout    *int64
			createdAt *time.Time
			settledAt *time.Time
		)
		if err := rows.Scan(
			&summary.ID, &summary.Name, &status, &summary.ScheduledAt,
			&summary.CreatedBy, &summary.CreatedAt, &summary.OpenedAt, &summary.StartedAt, &summary.FinishedAt, &summary.Seed,
			&summary.HorseCount, &summary.ParticipantCount,
			&betID, &horseID, &amount, &betStatus, &payout, &createdAt, &settledAt,
		); err != nil {
			return nil, err
		}
		summary.Status = races.Status(status)

		if betID != nil {
			summary.MyBet = &races.Bet{
				ID:        *betID,
				RaceID:    summary.ID,
				UserID:    userID,
				HorseID:   deref(horseID),
				Amount:    derefInt64(amount),
				Status:    races.BetStatus(deref(betStatus)),
				Payout:    payout,
				CreatedAt: derefTime(createdAt),
				SettledAt: settledAt,
			}
		}
		out = append(out, summary)
	}
	return out, rows.Err()
}

// ── Caballos ──────────────────────────────────────────────────────────────

// AddHorses inserta los caballos en un solo viaje.
//
// unnest en vez de un INSERT por caballo: una carrera de doce caballos serían
// doce viajes de ida y vuelta, y el instructor arma las carreras en vivo.
func (q *queries) AddHorses(ctx context.Context, raceID string, in []races.NewHorse) ([]races.Horse, error) {
	if err := checkIDs(raceID); err != nil {
		return nil, err
	}

	numbers := make([]int32, len(in))
	names := make([]string, len(in))
	odds := make([]int32, len(in))
	for i, h := range in {
		numbers[i] = int32(h.Number)
		names[i] = h.Name
		odds[i] = int32(h.Odds)
	}

	rows, err := q.tx.Query(ctx, `
		insert into horses (race_id, number, name, nominal_odds)
		select $1::uuid, t.number, t.name, t.nominal_odds
		from unnest($2::int[], $3::text[], $4::int[]) as t(number, name, nominal_odds)
		returning id::text, race_id::text, number, name, nominal_odds`,
		raceID, numbers, names, odds)
	if err != nil {
		return nil, notFound(err)
	}
	defer rows.Close()

	return scanHorses(rows)
}

func (q *queries) Horses(ctx context.Context, raceID string) ([]races.Horse, error) {
	if err := checkIDs(raceID); err != nil {
		return nil, err
	}

	rows, err := q.tx.Query(ctx, `
		select id::text, race_id::text, number, name, nominal_odds
		from horses where race_id = $1::uuid order by number`, raceID)
	if err != nil {
		return nil, notFound(err)
	}
	defer rows.Close()

	return scanHorses(rows)
}

func (q *queries) Horse(ctx context.Context, horseID string) (races.Horse, error) {
	if err := checkIDs(horseID); err != nil {
		return races.Horse{}, err
	}

	var horse races.Horse
	err := q.tx.QueryRow(ctx, `
		select id::text, race_id::text, number, name, nominal_odds
		from horses where id = $1::uuid`, horseID).
		Scan(&horse.ID, &horse.RaceID, &horse.Number, &horse.Name, &horse.NominalOdds)
	if err != nil {
		return races.Horse{}, notFound(err)
	}
	return horse, nil
}

func scanHorses(rows pgx.Rows) ([]races.Horse, error) {
	out := []races.Horse{}
	for rows.Next() {
		var h races.Horse
		if err := rows.Scan(&h.ID, &h.RaceID, &h.Number, &h.Name, &h.NominalOdds); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// ── Sala ──────────────────────────────────────────────────────────────────

// AddParticipant es idempotente: `on conflict do nothing`. Devuelve si esta
// llamada lo agregó de verdad, que es lo que decide si se difunde `room.joined`
// — recargar la página no tiene que avisarle a toda la sala.
func (q *queries) AddParticipant(ctx context.Context, raceID, userID string) (bool, error) {
	if err := checkIDs(raceID, userID); err != nil {
		return false, err
	}

	tag, err := q.tx.Exec(ctx, `
		insert into race_participants (race_id, user_id)
		values ($1::uuid, $2::uuid)
		on conflict (race_id, user_id) do nothing`, raceID, userID)
	if err != nil {
		return false, notFound(err)
	}
	return tag.RowsAffected() > 0, nil
}

func (q *queries) Participants(ctx context.Context, raceID string) ([]races.Participant, error) {
	if err := checkIDs(raceID); err != nil {
		return nil, err
	}

	rows, err := q.tx.Query(ctx, `
		select u.id::text, u.username, u.first_name, u.last_name, p.joined_at
		from race_participants p
		join users u on u.id = p.user_id
		where p.race_id = $1::uuid
		order by p.joined_at, u.username`, raceID)
	if err != nil {
		return nil, notFound(err)
	}
	defer rows.Close()

	out := []races.Participant{}
	for rows.Next() {
		var p races.Participant
		if err := rows.Scan(&p.UserID, &p.Username, &p.FirstName, &p.LastName, &p.JoinedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (q *queries) User(ctx context.Context, userID string) (races.Participant, error) {
	if err := checkIDs(userID); err != nil {
		return races.Participant{}, err
	}

	var p races.Participant
	err := q.tx.QueryRow(ctx, `
		select id::text, username, first_name, last_name, created_at
		from users where id = $1::uuid`, userID).
		Scan(&p.UserID, &p.Username, &p.FirstName, &p.LastName, &p.JoinedAt)
	if err != nil {
		return races.Participant{}, notFound(err)
	}
	return p, nil
}

// ── Apuestas ──────────────────────────────────────────────────────────────

// betColumns NO incluye una cuota congelada: la columna no existe en el esquema
// y si la economía termina siendo de cuotas fijas habrá que agregarla acá y en la
// migración a la vez. Ver races.SettlementRule.
const betColumns = `
	b.id::text, b.race_id::text, b.user_id::text, u.username, b.horse_id::text,
	b.amount, b.status::text, b.payout, b.created_at, b.settled_at`

func scanBet(row pgx.Row) (races.Bet, error) {
	var bet races.Bet
	var status string
	err := row.Scan(
		&bet.ID, &bet.RaceID, &bet.UserID, &bet.Username, &bet.HorseID,
		&bet.Amount, &status, &bet.Payout, &bet.CreatedAt, &bet.SettledAt,
	)
	if err != nil {
		return races.Bet{}, notFound(err)
	}
	bet.Status = races.BetStatus(status)
	return bet, nil
}

func (q *queries) Bets(ctx context.Context, raceID string) ([]races.Bet, error) {
	if err := checkIDs(raceID); err != nil {
		return nil, err
	}

	rows, err := q.tx.Query(ctx, `
		select `+betColumns+`
		from bets b join users u on u.id = b.user_id
		where b.race_id = $1::uuid
		order by b.created_at, b.id`, raceID)
	if err != nil {
		return nil, notFound(err)
	}
	defer rows.Close()

	out := []races.Bet{}
	for rows.Next() {
		bet, err := scanBet(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, bet)
	}
	return out, rows.Err()
}

func (q *queries) BetByUser(ctx context.Context, raceID, userID string) (races.Bet, bool, error) {
	if err := checkIDs(raceID, userID); err != nil {
		return races.Bet{}, false, err
	}

	bet, err := scanBet(q.tx.QueryRow(ctx, `
		select `+betColumns+`
		from bets b join users u on u.id = b.user_id
		where b.race_id = $1::uuid and b.user_id = $2::uuid`, raceID, userID))
	if errors.Is(err, races.ErrNotFound) {
		return races.Bet{}, false, nil
	}
	if err != nil {
		return races.Bet{}, false, err
	}
	return bet, true, nil
}

// InsertBet inserta la apuesta y devuelve la fila guardada.
//
// El id lo genera la base. Hace falta para el `ref_id` del movimiento del ledger:
// es lo que le permite al instructor seguir un descuento hasta la apuesta que lo
// causó, y a un alumno entender por qué tiene la nota que tiene.
func (q *queries) InsertBet(ctx context.Context, bet races.Bet) (races.Bet, error) {
	if err := checkIDs(bet.RaceID, bet.UserID, bet.HorseID); err != nil {
		return races.Bet{}, err
	}

	return scanBet(q.tx.QueryRow(ctx, `
		with inserted as (
			insert into bets (race_id, user_id, horse_id, amount, status)
			values ($1::uuid, $2::uuid, $3::uuid, $4, $5::bet_status)
			returning *
		)
		select `+betColumns+`
		from inserted b join users u on u.id = b.user_id`,
		bet.RaceID, bet.UserID, bet.HorseID, bet.Amount, string(bet.Status)))
}

// SettleBet marca la apuesta y le pone el pago.
//
// `where status = 'placed'` es la guarda contra la doble liquidación: si por
// cualquier camino se liquidara dos veces, la segunda no afecta ninguna fila. El
// ledger es append-only y un pago duplicado no se puede borrar.
func (q *queries) SettleBet(ctx context.Context, betID string, status races.BetStatus, payout int64, at time.Time) error {
	if err := checkIDs(betID); err != nil {
		return err
	}

	tag, err := q.tx.Exec(ctx, `
		update bets set status = $2::bet_status, payout = $3, settled_at = $4
		where id = $1::uuid and status = 'placed'`,
		betID, string(status), payout, at)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("la apuesta %s ya estaba liquidada", betID)
	}
	return nil
}

// ── Resultados ────────────────────────────────────────────────────────────

func (q *queries) InsertResults(ctx context.Context, raceID string, results []sim.Result) error {
	if err := checkIDs(raceID); err != nil {
		return err
	}

	horseIDs := make([]string, len(results))
	positions := make([]int32, len(results))
	for i, r := range results {
		horseIDs[i] = r.HorseID
		positions[i] = int32(r.Position)
	}

	_, err := q.tx.Exec(ctx, `
		insert into race_results (race_id, horse_id, position)
		select $1::uuid, t.horse_id::uuid, t.position
		from unnest($2::text[], $3::int[]) as t(horse_id, position)`,
		raceID, horseIDs, positions)
	return err
}

// InsertSettlement escribe la liquidación. Una fila por carrera: la clave primaria
// es `race_id`, así que un segundo intento de liquidar la misma carrera choca acá.
//
// Esa colisión es una defensa, no un estorbo. El esquema además exige
// `paid_out = pool` y que la devolución solo ocurra sin aciertos, así que una
// liquidación que no conserve monedas no entra ni con el código equivocado.
func (q *queries) InsertSettlement(ctx context.Context, s races.Settlement) error {
	if err := checkIDs(s.RaceID, s.WinnerHorseID); err != nil {
		return err
	}

	_, err := q.tx.Exec(ctx, `
		insert into race_settlements
			(race_id, winner_id, pool, winning_pool, paid_out, refunded, settled_at)
		values ($1::uuid, $2::uuid, $3, $4, $5, $6, $7)`,
		s.RaceID, s.WinnerHorseID, s.Pool, s.WinningPool, s.PaidOut, s.Refunded, s.SettledAt)
	return err
}

// Settlement lee la liquidación de una carrera. La usa el panel del instructor para
// mostrar de dónde salió cada pago: con pari-mutuel el pago de una apuesta no se
// puede recalcular desde la apuesta sola.
func (q *queries) Settlement(ctx context.Context, raceID string) (races.Settlement, bool, error) {
	if err := checkIDs(raceID); err != nil {
		return races.Settlement{}, false, err
	}

	var s races.Settlement
	err := q.tx.QueryRow(ctx, `
		select race_id::text, winner_id::text, pool, winning_pool, paid_out, refunded, settled_at
		from race_settlements where race_id = $1::uuid`, raceID,
	).Scan(&s.RaceID, &s.WinnerHorseID, &s.Pool, &s.WinningPool, &s.PaidOut, &s.Refunded, &s.SettledAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return races.Settlement{}, false, nil
	}
	if err != nil {
		return races.Settlement{}, false, err
	}
	return s, true, nil
}

// Results son los puestos finales, con el nombre y la cuota del caballo para que
// el cliente no tenga que cruzarlos.
func (q *queries) Results(ctx context.Context, raceID string) ([]sim.Result, error) {
	if err := checkIDs(raceID); err != nil {
		return nil, err
	}

	rows, err := q.tx.Query(ctx, `
		select h.id::text, h.name, h.number, h.nominal_odds, r.position
		from race_results r join horses h on h.id = r.horse_id
		where r.race_id = $1::uuid
		order by r.position`, raceID)
	if err != nil {
		return nil, notFound(err)
	}
	defer rows.Close()

	out := []sim.Result{}
	for rows.Next() {
		var r sim.Result
		if err := rows.Scan(&r.HorseID, &r.HorseName, &r.Number, &r.Odds, &r.Position); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ── Saldo ─────────────────────────────────────────────────────────────────

// Balance lee el saldo bloqueando la fila del usuario.
//
// `for update` acá y no un SELECT suelto: entre leer el saldo y descontar la
// apuesta no puede entrar otro movimiento. Sin el bloqueo, dos apuestas
// simultáneas del mismo alumno en dos carreras distintas podrían leer las dos el
// mismo saldo y gastarlo dos veces.
func (q *queries) Balance(ctx context.Context, userID string) (int64, error) {
	if err := checkIDs(userID); err != nil {
		return 0, err
	}

	var balance int64
	err := q.tx.QueryRow(ctx, `select balance from users where id = $1::uuid for update`, userID).Scan(&balance)
	if err != nil {
		return 0, notFound(err)
	}
	return balance, nil
}

// Move delega en el ledger, pasándole la transacción en curso.
func (q *queries) Move(ctx context.Context, m races.Movement) (int64, error) {
	if err := checkIDs(m.UserID); err != nil {
		return 0, err
	}

	if q.ledger == nil {
		return 0, errors.New("no hay ledger cableado: racesdb.New se llamó sin implementación")
	}
	return q.ledger.MoveTx(ctx, q.tx, Movement{
		UserID:    m.UserID,
		Delta:     m.Delta,
		Reason:    string(m.Reason),
		RefID:     m.RefID,
		CreatedBy: m.CreatedBy,
	})
}

// ── Auxiliares ────────────────────────────────────────────────────────────

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefInt64(n *int64) int64 {
	if n == nil {
		return 0
	}
	return *n
}

func derefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
