package races

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/talentodh/arena/internal/sim"
)

// memStore es un Store en memoria para los tests.
//
// No es un mock que registre llamadas: es una implementación que se comporta como
// la base, incluidas las tres restricciones del esquema que importan para las
// reglas que se prueban acá:
//
//   - `unique (race_id, user_id)` en bets — una apuesta por carrera;
//   - `unique (race_id, number)` en horses;
//   - `check (balance >= 0)` en users — el saldo tiene piso en 0.
//
// Y hace rollback de verdad: si fn devuelve error, el estado vuelve a como estaba.
// Sin eso, un test podría pasar con la mitad de una transacción aplicada, que es
// justamente el estado que no queremos que exista.
type memStore struct {
	mu   sync.Mutex
	data *memData
}

type memUser struct {
	id        string
	username  string
	firstName string
	lastName  string
	balance   int64
}

type memData struct {
	races        map[string]*Race
	horses       map[string]*Horse
	participants map[string]map[string]time.Time
	bets         map[string]*Bet
	results      map[string][]sim.Result
	settlements  map[string]Settlement
	users        map[string]*memUser
	ledger       []Movement
	seq          int
}

func newMemStore() *memStore {
	return &memStore{data: &memData{
		races:        map[string]*Race{},
		horses:       map[string]*Horse{},
		participants: map[string]map[string]time.Time{},
		bets:         map[string]*Bet{},
		results:      map[string][]sim.Result{},
		settlements:  map[string]Settlement{},
		users:        map[string]*memUser{},
	}}
}

func (d *memData) clone() *memData {
	out := &memData{
		races:        make(map[string]*Race, len(d.races)),
		horses:       make(map[string]*Horse, len(d.horses)),
		participants: make(map[string]map[string]time.Time, len(d.participants)),
		bets:         make(map[string]*Bet, len(d.bets)),
		results:      make(map[string][]sim.Result, len(d.results)),
		settlements:  make(map[string]Settlement, len(d.settlements)),
		users:        make(map[string]*memUser, len(d.users)),
		ledger:       append([]Movement(nil), d.ledger...),
		seq:          d.seq,
	}
	for k, v := range d.races {
		copied := *v
		out.races[k] = &copied
	}
	for k, v := range d.horses {
		copied := *v
		out.horses[k] = &copied
	}
	for k, v := range d.bets {
		copied := *v
		out.bets[k] = &copied
	}
	for k, v := range d.users {
		copied := *v
		out.users[k] = &copied
	}
	for k, v := range d.participants {
		inner := make(map[string]time.Time, len(v))
		for u, at := range v {
			inner[u] = at
		}
		out.participants[k] = inner
	}
	for k, v := range d.results {
		out.results[k] = append([]sim.Result(nil), v...)
	}
	for k, v := range d.settlements {
		out.settlements[k] = v
	}
	return out
}

func (d *memData) nextID(prefix string) string {
	d.seq++
	return fmt.Sprintf("%s-%04d", prefix, d.seq)
}

// InTx corre fn contra una copia y la publica solo si no hubo error. Es el mismo
// contrato que la transacción de Postgres: todo o nada.
func (s *memStore) InTx(ctx context.Context, fn func(Tx) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	working := s.data.clone()
	if err := fn(&memTx{data: working}); err != nil {
		return err
	}
	s.data = working
	return nil
}

// ── Helpers de armado para los tests ──────────────────────────────────────

func (s *memStore) addUser(id, username string, balance int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.users[id] = &memUser{
		id: id, username: username,
		firstName: "Nombre", lastName: "Apellido",
		balance: balance,
	}
}

func (s *memStore) balanceOf(userID string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if u, ok := s.data.users[userID]; ok {
		return u.balance
	}
	return -1
}

func (s *memStore) betOf(raceID, userID string) (Bet, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, bet := range s.data.bets {
		if bet.RaceID == raceID && bet.UserID == userID {
			return *bet, true
		}
	}
	return Bet{}, false
}

// movements son los movimientos del ledger de un usuario, en orden.
func (s *memStore) movements(userID string) []Movement {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []Movement{}
	for _, m := range s.data.ledger {
		if m.UserID == userID {
			out = append(out, m)
		}
	}
	return out
}

func (s *memStore) statusOf(raceID string) Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	if race, ok := s.data.races[raceID]; ok {
		return race.Status
	}
	return ""
}

// forceStatus pone un estado sin pasar por la máquina. Lo usan los tests para
// armar el escenario: para probar que apostar con la carrera en `running` falla,
// hay que poder poner una carrera en `running` sin correr la simulación de 42
// segundos.
func (s *memStore) forceStatus(raceID string, status Status, seed *int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if race, ok := s.data.races[raceID]; ok {
		race.Status = status
		race.Seed = seed
	}
}

// ── memTx ─────────────────────────────────────────────────────────────────

type memTx struct {
	data *memData
}

func (t *memTx) CreateRace(_ context.Context, in NewRace) (Race, error) {
	race := &Race{
		ID:          t.data.nextID("race"),
		Name:        in.Name,
		Status:      StatusDraft,
		ScheduledAt: in.ScheduledAt,
		CreatedBy:   in.CreatedBy,
		CreatedAt:   time.Unix(0, 0).UTC(),
	}
	t.data.races[race.ID] = race

	if len(in.Horses) > 0 {
		if _, err := t.AddHorses(context.Background(), race.ID, in.Horses); err != nil {
			return Race{}, err
		}
	}
	return *race, nil
}

func (t *memTx) UpdateRace(_ context.Context, raceID string, patch RacePatch) (Race, error) {
	race, ok := t.data.races[raceID]
	if !ok {
		return Race{}, ErrNotFound
	}
	if patch.Name != nil {
		race.Name = *patch.Name
	}
	if patch.ScheduledAt != nil {
		race.ScheduledAt = patch.ScheduledAt
	}
	return *race, nil
}

func (t *memTx) Race(_ context.Context, raceID string) (Race, error) {
	race, ok := t.data.races[raceID]
	if !ok {
		return Race{}, ErrNotFound
	}
	return *race, nil
}

func (t *memTx) RaceForUpdate(ctx context.Context, raceID string) (Race, error) {
	// En memoria no hace falta bloquear: InTx ya serializa con un mutex, que es
	// más estricto que `for update`.
	return t.Race(ctx, raceID)
}

func (t *memTx) SetStatus(_ context.Context, raceID string, status Status, at time.Time) error {
	race, ok := t.data.races[raceID]
	if !ok {
		return ErrNotFound
	}
	race.Status = status
	switch status {
	case StatusOpen:
		race.OpenedAt = &at
	case StatusRunning:
		race.StartedAt = &at
	case StatusFinished, StatusCancelled:
		race.FinishedAt = &at
	}
	return nil
}

func (t *memTx) SetSeed(_ context.Context, raceID string, seed int64) error {
	race, ok := t.data.races[raceID]
	if !ok {
		return ErrNotFound
	}
	// La semilla se escribe una sola vez, igual que en Postgres.
	if race.Seed != nil {
		return fmt.Errorf("la carrera %s ya tenía semilla", raceID)
	}
	race.Seed = &seed
	return nil
}

func (t *memTx) VisibleRaces(ctx context.Context, userID string, includeDraft bool) ([]Summary, error) {
	out := []Summary{}
	for _, race := range t.data.races {
		if race.Status == StatusDraft && !includeDraft {
			continue
		}

		summary := Summary{Race: *race}
		for _, h := range t.data.horses {
			if h.RaceID == race.ID {
				summary.HorseCount++
			}
		}
		summary.ParticipantCount = len(t.data.participants[race.ID])
		if bet, found, _ := t.BetByUser(ctx, race.ID, userID); found {
			summary.MyBet = &bet
		}
		out = append(out, summary)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (t *memTx) RunningRaces(_ context.Context) ([]Race, error) {
	out := []Race{}
	for _, race := range t.data.races {
		if race.Status == StatusRunning {
			out = append(out, *race)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (t *memTx) AddHorses(_ context.Context, raceID string, in []NewHorse) ([]Horse, error) {
	if _, ok := t.data.races[raceID]; !ok {
		return nil, ErrNotFound
	}

	used := map[int]bool{}
	for _, h := range t.data.horses {
		if h.RaceID == raceID {
			used[h.Number] = true
		}
	}

	out := []Horse{}
	for _, h := range in {
		// unique (race_id, number).
		if used[h.Number] {
			return nil, fmt.Errorf("el número %d ya está usado en la carrera %s", h.Number, raceID)
		}
		used[h.Number] = true

		horse := &Horse{
			ID:          t.data.nextID("horse"),
			RaceID:      raceID,
			Number:      h.Number,
			Name:        h.Name,
			NominalOdds: h.Odds,
		}
		t.data.horses[horse.ID] = horse
		out = append(out, *horse)
	}
	return out, nil
}

func (t *memTx) Horses(_ context.Context, raceID string) ([]Horse, error) {
	out := []Horse{}
	for _, h := range t.data.horses {
		if h.RaceID == raceID {
			out = append(out, *h)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Number < out[j].Number })
	return out, nil
}

func (t *memTx) Horse(_ context.Context, horseID string) (Horse, error) {
	horse, ok := t.data.horses[horseID]
	if !ok {
		return Horse{}, ErrNotFound
	}
	return *horse, nil
}

func (t *memTx) AddParticipant(_ context.Context, raceID, userID string) (bool, error) {
	if t.data.participants[raceID] == nil {
		t.data.participants[raceID] = map[string]time.Time{}
	}
	if _, already := t.data.participants[raceID][userID]; already {
		return false, nil
	}
	t.data.participants[raceID][userID] = time.Unix(0, 0).UTC()
	return true, nil
}

func (t *memTx) Participants(_ context.Context, raceID string) ([]Participant, error) {
	out := []Participant{}
	for userID, at := range t.data.participants[raceID] {
		user, ok := t.data.users[userID]
		if !ok {
			continue
		}
		out = append(out, Participant{
			UserID: userID, Username: user.username,
			FirstName: user.firstName, LastName: user.lastName, JoinedAt: at,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Username < out[j].Username })
	return out, nil
}

func (t *memTx) Bets(_ context.Context, raceID string) ([]Bet, error) {
	out := []Bet{}
	for _, bet := range t.data.bets {
		if bet.RaceID == raceID {
			out = append(out, *bet)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (t *memTx) BetByUser(_ context.Context, raceID, userID string) (Bet, bool, error) {
	for _, bet := range t.data.bets {
		if bet.RaceID == raceID && bet.UserID == userID {
			return *bet, true, nil
		}
	}
	return Bet{}, false, nil
}

func (t *memTx) InsertBet(_ context.Context, bet Bet) (Bet, error) {
	// unique (race_id, user_id): la restricción que impide cubrir todos los
	// caballos y garantizarse nota.
	for _, existing := range t.data.bets {
		if existing.RaceID == bet.RaceID && existing.UserID == bet.UserID {
			return Bet{}, errors.New("ya hay una apuesta de ese usuario en esa carrera")
		}
	}

	stored := bet
	stored.ID = t.data.nextID("bet")
	if stored.Status == "" {
		stored.Status = BetStatusPlaced
	}
	t.data.bets[stored.ID] = &stored
	return stored, nil
}

func (t *memTx) SettleBet(_ context.Context, betID string, status BetStatus, payout int64, at time.Time) error {
	bet, ok := t.data.bets[betID]
	if !ok {
		return ErrNotFound
	}
	if bet.Status != BetStatusPlaced {
		return fmt.Errorf("la apuesta %s ya estaba liquidada", betID)
	}
	bet.Status = status
	bet.Payout = &payout
	bet.SettledAt = &at
	return nil
}

func (t *memTx) InsertResults(_ context.Context, raceID string, results []sim.Result) error {
	if _, already := t.data.results[raceID]; already {
		return fmt.Errorf("la carrera %s ya tenía resultados", raceID)
	}
	t.data.results[raceID] = append([]sim.Result(nil), results...)
	return nil
}

func (t *memTx) Results(_ context.Context, raceID string) ([]sim.Result, error) {
	return append([]sim.Result(nil), t.data.results[raceID]...), nil
}

// InsertSettlement reproduce la clave primaria de race_settlements: liquidar dos
// veces la misma carrera falla. Sin esto, el doble sería más permisivo que Postgres
// y un test podría pasar acá y romper en producción.
func (t *memTx) InsertSettlement(_ context.Context, s Settlement) error {
	if _, exists := t.data.settlements[s.RaceID]; exists {
		return fmt.Errorf("la carrera %s ya tiene liquidación", s.RaceID)
	}
	// Y los dos CHECK del esquema, por el mismo motivo.
	if s.PaidOut != s.Pool {
		return fmt.Errorf("liquidación que no conserva: %d pagado de un pool de %d", s.PaidOut, s.Pool)
	}
	if s.Refunded != (s.WinningPool == 0) {
		return fmt.Errorf("refunded=%v con un pool ganador de %d", s.Refunded, s.WinningPool)
	}
	t.data.settlements[s.RaceID] = s
	return nil
}

func (t *memTx) Settlement(_ context.Context, raceID string) (Settlement, bool, error) {
	s, ok := t.data.settlements[raceID]
	return s, ok, nil
}

func (t *memTx) Balance(_ context.Context, userID string) (int64, error) {
	user, ok := t.data.users[userID]
	if !ok {
		return 0, ErrNotFound
	}
	return user.balance, nil
}

// Move es el doble del ledger.
//
// Hace lo mismo que hará el de verdad en lo único que este paquete necesita: es
// append-only, mantiene el saldo en la misma transacción, y RECHAZA el movimiento
// si el saldo quedaría negativo. Ese rechazo es el `check (balance >= 0)` del
// esquema, y está acá para que el test pueda comprobar que el servicio no lo trata
// como imposible.
func (t *memTx) Move(_ context.Context, m Movement) (int64, error) {
	user, ok := t.data.users[m.UserID]
	if !ok {
		return 0, ErrNotFound
	}
	if m.Delta == 0 {
		return 0, errors.New("un movimiento de cero no es un movimiento")
	}
	if user.balance+m.Delta < 0 {
		return 0, fmt.Errorf("el saldo quedaría negativo: %d %+d", user.balance, m.Delta)
	}

	user.balance += m.Delta
	t.data.ledger = append(t.data.ledger, m)
	return user.balance, nil
}

func (t *memTx) User(_ context.Context, userID string) (Participant, error) {
	user, ok := t.data.users[userID]
	if !ok {
		return Participant{}, ErrNotFound
	}
	return Participant{
		UserID: user.id, Username: user.username,
		FirstName: user.firstName, LastName: user.lastName,
	}, nil
}

// Comprobación en tiempo de compilación: el doble sigue implementando lo mismo
// que la implementación de verdad.
var (
	_ Store = (*memStore)(nil)
	_ Tx    = (*memTx)(nil)
)

// ── Hub falso ─────────────────────────────────────────────────────────────

// recorder es un Broadcaster que guarda lo que se difundió. Es lo que permite
// verificar que `bet.placed` no lleva el caballo y que `race.finished` se arma por
// destinatario, en vez de confiar en que el código lo hace.
type recorder struct {
	mu sync.Mutex

	// room son los eventos difundidos a toda la sala.
	room []any
	// perUser son los eventos armados por destinatario, indexados por usuario.
	perUser map[string][]any
	// toUser son los eventos dirigidos a un usuario.
	toUser map[string][]any

	// audience es a quiénes se les arma un evento por destinatario. Lo fija el
	// test, porque en memoria no hay conexiones abiertas.
	audience []string
}

func newRecorder(audience ...string) *recorder {
	return &recorder{
		perUser:  map[string][]any{},
		toUser:   map[string][]any{},
		audience: audience,
	}
}

func (r *recorder) ToRoom(_ string, event any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.room = append(r.room, event)
}

func (r *recorder) ToUser(userID string, event any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.toUser[userID] = append(r.toUser[userID], event)
}

func (r *recorder) ToRoomPerUser(_ string, build func(userID string) any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, userID := range r.audience {
		r.perUser[userID] = append(r.perUser[userID], build(userID))
	}
}

// roomEvents devuelve los eventos difundidos de un tipo concreto.
func roomEventsOf[T any](r *recorder) []T {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := []T{}
	for _, event := range r.room {
		if typed, ok := event.(T); ok {
			out = append(out, typed)
		}
	}
	return out
}

func perUserEventsOf[T any](r *recorder, userID string) []T {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := []T{}
	for _, event := range r.perUser[userID] {
		if typed, ok := event.(T); ok {
			out = append(out, typed)
		}
	}
	return out
}

var _ Broadcaster = (*recorder)(nil)
