// Package store guarda todo el estado del servidor en memoria, con una copia
// en disco para que no se pierda al reiniciar.
//
// No hay base de datos a propósito: el dataset entero pesa menos de 100 KB y
// una app de enseñanza no debería necesitar levantar un Postgres para dar la
// primera clase. La copia en disco es un JSON; si la ruta no se puede
// escribir, el servidor avisa una vez y sigue funcionando en memoria.
package store

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/talentodh/hipodromo/internal/auth"
	"github.com/talentodh/hipodromo/internal/contract"
	"github.com/talentodh/hipodromo/internal/seed"
)

// DevPassword es la contraseña de todas las cuentas del dataset. Está
// documentada en docs/contract/README.md; son cuentas de prueba.
const DevPassword = "Carrera123!"

// User es un usuario con su hash de contraseña, que nunca sale de este paquete.
type User struct {
	contract.User
	PasswordHash string `json:"passwordHash"`
}

// RefreshToken es de un solo uso: al canjearlo se borra y se emite otro.
type RefreshToken struct {
	UserID    string    `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// VerificationToken caduca a las 24 h y también es de un solo uso.
type VerificationToken struct {
	UserID    string    `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Store es el estado completo. Un solo candado: el volumen de escritura de
// esta app no justifica nada más fino, y un candado se razona.
type Store struct {
	mu sync.RWMutex

	races     map[string]*contract.Race
	raceOrder []string
	results   map[string]*contract.RaceResult

	users        map[string]*User
	usersByEmail map[string]string // email en minúsculas → userID

	bets     map[string]*contract.Bet
	betOrder []string

	refreshTokens map[string]RefreshToken
	verifyTokens  map[string]VerificationToken

	// runIndex es la corrida actual de cada carrera. Alimenta el simulador,
	// que es determinístico por (raceID, runIndex).
	runIndex map[string]int

	seq          int
	snapshotPath string
	persistErr   error
}

// Options configura el store.
type Options struct {
	// SnapshotPath es dónde se guarda la copia. Vacío desactiva la persistencia.
	SnapshotPath string
	// Reset ignora la copia existente y arranca desde el dataset limpio.
	// Es lo que hay que usar para volver al estado conocido de la demo.
	Reset bool
}

// New carga el dataset y, si corresponde, le aplica la copia guardada encima.
func New(data *seed.Data, opts Options) (*Store, error) {
	// Todas las cuentas del dataset comparten contraseña, así que se deriva
	// una sola vez: 12 × PBKDF2 con 210 000 iteraciones serían más de un
	// segundo de arranque para nada.
	devHash, err := auth.HashPassword(DevPassword)
	if err != nil {
		return nil, fmt.Errorf("hasheando la contraseña de desarrollo: %w", err)
	}

	s := &Store{
		races:         make(map[string]*contract.Race, len(data.Races)),
		results:       make(map[string]*contract.RaceResult, len(data.Results)),
		users:         make(map[string]*User, len(data.Users)),
		usersByEmail:  make(map[string]string, len(data.Users)),
		bets:          make(map[string]*contract.Bet, len(data.Bets)),
		refreshTokens: make(map[string]RefreshToken),
		verifyTokens:  make(map[string]VerificationToken),
		runIndex:      make(map[string]int),
		snapshotPath:  opts.SnapshotPath,
	}

	for i := range data.Races {
		race := data.Races[i]
		s.races[race.ID] = &race
		s.raceOrder = append(s.raceOrder, race.ID)
	}
	for i := range data.Results {
		result := data.Results[i]
		s.results[result.RaceID] = &result
	}
	for _, u := range data.Users {
		user := &User{User: u.User, PasswordHash: devHash}
		s.users[user.ID] = user
		s.usersByEmail[strings.ToLower(user.Email)] = user.ID
	}
	for i := range data.Bets {
		bet := data.Bets[i]
		s.bets[bet.ID] = &bet
		s.betOrder = append(s.betOrder, bet.ID)
	}
	s.seq = len(data.Users) + len(data.Bets)

	if opts.SnapshotPath != "" && !opts.Reset {
		if err := s.loadSnapshot(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// PersistError devuelve el error de la última escritura fallida, si hubo.
// El servidor lo consulta al arrancar para avisar una sola vez.
func (s *Store) PersistError() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.persistErr
}

// ── Carreras ──────────────────────────────────────────────────────────────

// Races devuelve las carreras ordenadas para mostrar: primero las que están
// corriendo, después las que vienen por hora de largada, y al final las
// terminadas de la más reciente a la más vieja.
func (s *Store) Races(status contract.RaceStatus) []contract.Race {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]contract.Race, 0, len(s.raceOrder))
	for _, id := range s.raceOrder {
		race := s.races[id]
		if status != "" && race.Status != status {
			continue
		}
		out = append(out, *race)
	}

	rank := map[contract.RaceStatus]int{contract.StatusLive: 0, contract.StatusUpcoming: 1, contract.StatusFinished: 2}
	sort.SliceStable(out, func(i, j int) bool {
		if rank[out[i].Status] != rank[out[j].Status] {
			return rank[out[i].Status] < rank[out[j].Status]
		}
		if out[i].Status == contract.StatusFinished {
			return out[i].StartsAt > out[j].StartsAt // más reciente primero
		}
		return out[i].StartsAt < out[j].StartsAt
	})
	return out
}

func (s *Store) Race(id string) (contract.Race, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	race, ok := s.races[id]
	if !ok {
		return contract.Race{}, false
	}
	return *race, true
}

// ScheduleRace deja una carrera lista para largar: la marca `upcoming`, le
// pone hora y devuelve el índice de corrida que le toca.
func (s *Store) ScheduleRace(id string, startsAt time.Time) (int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	race, ok := s.races[id]
	if !ok {
		return 0, false
	}
	race.Status = contract.StatusUpcoming
	race.StartsAt = seed.Format(startsAt)
	run := s.runIndex[id]
	s.persistLocked()
	return run, true
}

func (s *Store) SetRaceStatus(id string, status contract.RaceStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if race, ok := s.races[id]; ok {
		race.Status = status
		s.persistLocked()
	}
}

// Result devuelve el resultado de una carrera con los pagos del usuario
// indicado. Sin sesión (`userID` vacío) los pagos van vacíos: el podio es
// público, lo que cobró cada uno no.
func (s *Store) Result(raceID, userID string) (contract.RaceResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result, ok := s.results[raceID]
	if !ok {
		return contract.RaceResult{}, false
	}

	out := *result
	out.Payouts = []contract.Payout{}
	if userID != "" {
		for _, id := range s.betOrder {
			bet := s.bets[id]
			if bet.RaceID != raceID || bet.UserID != userID || bet.Status == contract.BetPending {
				continue
			}
			out.Payouts = append(out.Payouts, contract.Payout{
				BetID: bet.ID, HorseID: bet.HorseID, Stake: bet.Amount, Amount: bet.Payout,
			})
		}
	}
	return out, true
}

// ── Usuarios ──────────────────────────────────────────────────────────────

func (s *Store) UserByID(id string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return User{}, false
	}
	return *u, true
}

func (s *Store) UserByEmail(email string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.usersByEmail[strings.ToLower(strings.TrimSpace(email))]
	if !ok {
		return User{}, false
	}
	return *s.users[id], true
}

// CreateUser registra una cuenta nueva con el saldo inicial del contrato.
func (s *Store) CreateUser(email, displayName, passwordHash string) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := strings.ToLower(strings.TrimSpace(email))
	if _, exists := s.usersByEmail[key]; exists {
		return User{}, contract.Errorf(contract.CodeEmailAlreadyRegistered)
	}

	s.seq++
	user := &User{
		User: contract.User{
			ID:            fmt.Sprintf("usr_%03d", s.seq),
			Email:         key,
			DisplayName:   strings.TrimSpace(displayName),
			Balance:       contract.SignupBalance,
			EmailVerified: false,
		},
		PasswordHash: passwordHash,
	}
	s.users[user.ID] = user
	s.usersByEmail[key] = user.ID
	s.persistLocked()
	return *user, nil
}

func (s *Store) MarkVerified(userID string) (contract.User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[userID]
	if !ok {
		return contract.User{}, false
	}
	user.EmailVerified = true
	s.persistLocked()
	return user.User, true
}

// ── Tokens ────────────────────────────────────────────────────────────────

func (s *Store) SaveRefreshToken(token, userID string, expiresAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshTokens[token] = RefreshToken{UserID: userID, ExpiresAt: expiresAt}
	s.persistLocked()
}

// ConsumeRefreshToken canjea un refresh token: lo borra y devuelve su dueño.
// Es de un solo uso, así que reusar uno ya canjeado falla — y eso es lo que
// detecta un token robado.
func (s *Store) ConsumeRefreshToken(token string, now time.Time) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.refreshTokens[token]
	if !ok {
		return "", false
	}
	delete(s.refreshTokens, token)
	s.persistLocked()

	if now.After(entry.ExpiresAt) {
		return "", false
	}
	return entry.UserID, true
}

// RevokeUserTokens invalida todos los refresh tokens de un usuario. Es lo que
// hace el logout.
func (s *Store) RevokeUserTokens(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for token, entry := range s.refreshTokens {
		if entry.UserID == userID {
			delete(s.refreshTokens, token)
		}
	}
	s.persistLocked()
}

func (s *Store) SaveVerificationToken(token, userID string, expiresAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.verifyTokens[token] = VerificationToken{UserID: userID, ExpiresAt: expiresAt}
	s.persistLocked()
}

// ConsumeVerificationToken distingue "no existe" de "venció": el contrato
// tiene un código distinto para cada caso, y al usuario le sirve saber si
// tiene que pedir uno nuevo o si el enlace estaba mal.
func (s *Store) ConsumeVerificationToken(token string, now time.Time) (userID string, expired bool, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.verifyTokens[token]
	if !exists {
		return "", false, false
	}
	delete(s.verifyTokens, token)
	s.persistLocked()

	if now.After(entry.ExpiresAt) {
		return "", true, false
	}
	return entry.UserID, false, true
}

// ── Apuestas ──────────────────────────────────────────────────────────────

// PlaceBet valida y registra una apuesta. Descontar el saldo y crear la
// apuesta pasa bajo el mismo candado: no puede quedar una sin lo otro.
func (s *Store) PlaceBet(userID, raceID, horseID string, amount int, now time.Time) (contract.Bet, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return contract.Bet{}, 0, contract.Errorf(contract.CodeUnauthenticated)
	}
	if !user.EmailVerified {
		return contract.Bet{}, 0, contract.Errorf(contract.CodeEmailNotVerified)
	}

	race, ok := s.races[raceID]
	if !ok {
		return contract.Bet{}, 0, contract.ErrorWith(contract.CodeNotFound,
			map[string]any{"resource": "race", "id": raceID})
	}
	if race.Status != contract.StatusUpcoming {
		return contract.Bet{}, 0, contract.ErrorWith(contract.CodeRaceAlreadyStarted,
			map[string]any{"raceId": raceID, "status": string(race.Status)})
	}

	horse, ok := race.Horse(horseID)
	if !ok {
		return contract.Bet{}, 0, contract.ErrorWith(contract.CodeHorseNotInRace,
			map[string]any{"raceId": raceID, "horseId": horseID})
	}
	if amount < contract.MinBetAmount || amount > contract.MaxBetAmount {
		return contract.Bet{}, 0, contract.ErrorWith(contract.CodeBetAmountOutOfRange,
			map[string]any{"min": contract.MinBetAmount, "max": contract.MaxBetAmount})
	}
	if user.Balance < amount {
		return contract.Bet{}, 0, contract.ErrorWith(contract.CodeInsufficientBalance,
			map[string]any{"balance": user.Balance, "amount": amount})
	}

	s.seq++
	bet := &contract.Bet{
		ID:        fmt.Sprintf("bet_%03d", s.seq),
		UserID:    userID,
		RaceID:    race.ID,
		RaceName:  race.Name,
		HorseID:   horse.ID,
		HorseName: horse.Name,
		Amount:    amount,
		Odds:      horse.Odds, // congelada al apostar
		Status:    contract.BetPending,
		Payout:    0,
		PlacedAt:  seed.Format(now),
	}
	s.bets[bet.ID] = bet
	s.betOrder = append(s.betOrder, bet.ID)
	user.Balance -= amount
	s.persistLocked()

	return *bet, user.Balance, nil
}

// BetsByUser devuelve el historial, de la más reciente a la más vieja.
func (s *Store) BetsByUser(userID string) []contract.Bet {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]contract.Bet, 0, 8)
	for _, id := range s.betOrder {
		if bet := s.bets[id]; bet.UserID == userID {
			out = append(out, *bet)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].PlacedAt > out[j].PlacedAt })
	return out
}

// Settlement es lo que le pasó a un usuario cuando terminó una carrera.
type Settlement struct {
	UserID  string
	Balance int
	Payouts []contract.Payout
}

// SettleRace liquida todas las apuestas pendientes de una carrera y guarda el
// resultado. Devuelve una entrada por usuario afectado, que el hub usa para
// mandarle `balance.updated` a cada uno.
func (s *Store) SettleRace(raceID string, podium []contract.PodiumEntry, finishedAt time.Time) []Settlement {
	s.mu.Lock()
	defer s.mu.Unlock()

	race, ok := s.races[raceID]
	if !ok || len(podium) == 0 {
		return nil
	}
	winner := podium[0].HorseID

	byUser := map[string]*Settlement{}
	for _, id := range s.betOrder {
		bet := s.bets[id]
		if bet.RaceID != raceID || bet.Status != contract.BetPending {
			continue
		}

		if bet.HorseID == winner {
			bet.Status = contract.BetWon
			bet.Payout = int(math.Round(float64(bet.Amount) * bet.Odds))
		} else {
			bet.Status = contract.BetLost
			bet.Payout = 0
		}

		entry, seen := byUser[bet.UserID]
		if !seen {
			entry = &Settlement{UserID: bet.UserID}
			byUser[bet.UserID] = entry
		}
		entry.Payouts = append(entry.Payouts, contract.Payout{
			BetID: bet.ID, HorseID: bet.HorseID, Stake: bet.Amount, Amount: bet.Payout,
		})

		if user, exists := s.users[bet.UserID]; exists {
			user.Balance += bet.Payout
			entry.Balance = user.Balance
		}
	}

	race.Status = contract.StatusFinished
	s.results[raceID] = &contract.RaceResult{
		RaceID:     raceID,
		FinishedAt: seed.Format(finishedAt),
		Podium:     podium,
		Payouts:    []contract.Payout{},
	}
	s.runIndex[raceID]++
	s.persistLocked()

	out := make([]Settlement, 0, len(byUser))
	for _, entry := range byUser {
		out = append(out, *entry)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UserID < out[j].UserID })
	return out
}

// ── Leaderboard ───────────────────────────────────────────────────────────

// Leaderboard calcula el ranking desde las apuestas liquidadas. No se guarda
// nada precalculado: el golden de docs/contract/seed/leaderboard.json existe
// justamente para verificar este cálculo.
//
// Solo entran usuarios con al menos una apuesta liquidada en el período.
// Orden: profit desc, wins desc, displayName asc — un orden total, para que el
// ranking no baile entre requests.
func (s *Store) Leaderboard(period string, now time.Time) []contract.LeaderboardEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	today := now.UTC().Format("2006-01-02")

	type agg struct{ profit, bets, wins int }
	acc := map[string]*agg{}

	for _, id := range s.betOrder {
		bet := s.bets[id]
		if bet.Status == contract.BetPending {
			continue
		}
		result, ok := s.results[bet.RaceID]
		if !ok {
			continue
		}
		if period == "daily" && !strings.HasPrefix(result.FinishedAt, today) {
			continue
		}

		entry, seen := acc[bet.UserID]
		if !seen {
			entry = &agg{}
			acc[bet.UserID] = entry
		}
		entry.profit += bet.Payout - bet.Amount
		entry.bets++
		if bet.Status == contract.BetWon {
			entry.wins++
		}
	}

	out := make([]contract.LeaderboardEntry, 0, len(acc))
	for userID, e := range acc {
		user, ok := s.users[userID]
		if !ok {
			continue
		}
		out = append(out, contract.LeaderboardEntry{
			UserID: userID, DisplayName: user.DisplayName,
			Profit: e.profit, Bets: e.bets, Wins: e.wins,
		})
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Profit != out[j].Profit {
			return out[i].Profit > out[j].Profit
		}
		if out[i].Wins != out[j].Wins {
			return out[i].Wins > out[j].Wins
		}
		return compareES(out[i].DisplayName, out[j].DisplayName)
	})

	if len(out) > contract.LeaderboardTop {
		out = out[:contract.LeaderboardTop]
	}
	for i := range out {
		out[i].Rank = i + 1
	}
	return out
}

// compareES ordena nombres como los ordena el frontend con
// localeCompare(…, 'es'): las tildes no cambian el lugar de la letra.
func compareES(a, b string) bool { return foldES(a) < foldES(b) }

var esFolds = strings.NewReplacer(
	"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u",
	"Á", "a", "É", "e", "Í", "i", "Ó", "o", "Ú", "u", "Ü", "u",
)

func foldES(s string) string { return esFolds.Replace(strings.ToLower(s)) }

// ── Copia en disco ────────────────────────────────────────────────────────

type snapshot struct {
	Version       int                          `json:"version"`
	Seq           int                          `json:"seq"`
	Users         []User                       `json:"users"`
	Bets          []contract.Bet               `json:"bets"`
	Results       []contract.RaceResult        `json:"results"`
	RunIndex      map[string]int               `json:"runIndex"`
	RefreshTokens map[string]RefreshToken      `json:"refreshTokens"`
	VerifyTokens  map[string]VerificationToken `json:"verifyTokens"`
}

const snapshotVersion = 1

// persistLocked escribe la copia. Se llama con el candado tomado.
//
// Escritura completa en cada mutación: con un dataset de este tamaño son
// microsegundos, y evita toda una clase de bugs de escritura parcial que en un
// proyecto de enseñanza no aportan nada.
func (s *Store) persistLocked() {
	if s.snapshotPath == "" {
		return
	}

	snap := snapshot{
		Version: snapshotVersion, Seq: s.seq,
		RunIndex: s.runIndex, RefreshTokens: s.refreshTokens, VerifyTokens: s.verifyTokens,
	}
	for _, id := range sortedKeys(s.users) {
		snap.Users = append(snap.Users, *s.users[id])
	}
	for _, id := range s.betOrder {
		snap.Bets = append(snap.Bets, *s.bets[id])
	}
	for _, id := range sortedKeys(s.results) {
		snap.Results = append(snap.Results, *s.results[id])
	}

	raw, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		s.persistErr = err
		return
	}
	if err := os.MkdirAll(filepath.Dir(s.snapshotPath), 0o755); err != nil {
		s.persistErr = err
		return
	}

	// Escritura atómica: temporal y rename. Un corte a mitad de escritura
	// dejaría un JSON truncado que no se puede cargar al arrancar.
	tmp := s.snapshotPath + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		s.persistErr = err
		return
	}
	if err := os.Rename(tmp, s.snapshotPath); err != nil {
		s.persistErr = err
		return
	}
	s.persistErr = nil
}

func (s *Store) loadSnapshot() error {
	raw, err := os.ReadFile(s.snapshotPath)
	if os.IsNotExist(err) {
		return nil // primer arranque
	}
	if err != nil {
		return fmt.Errorf("leyendo la copia %s: %w", s.snapshotPath, err)
	}

	var snap snapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		return fmt.Errorf("la copia %s está corrupta — borrala o arrancá con RESET=1: %w", s.snapshotPath, err)
	}
	if snap.Version != snapshotVersion {
		return fmt.Errorf("la copia %s es de la versión %d y esta build usa la %d — borrala o arrancá con RESET=1",
			s.snapshotPath, snap.Version, snapshotVersion)
	}

	// La copia pisa el dataset: es el estado más nuevo.
	for i := range snap.Users {
		user := snap.Users[i]
		s.users[user.ID] = &user
		s.usersByEmail[strings.ToLower(user.Email)] = user.ID
	}
	s.bets = make(map[string]*contract.Bet, len(snap.Bets))
	s.betOrder = s.betOrder[:0]
	for i := range snap.Bets {
		bet := snap.Bets[i]
		s.bets[bet.ID] = &bet
		s.betOrder = append(s.betOrder, bet.ID)
	}
	for i := range snap.Results {
		result := snap.Results[i]
		s.results[result.RaceID] = &result
	}
	if snap.RunIndex != nil {
		s.runIndex = snap.RunIndex
	}
	if snap.RefreshTokens != nil {
		s.refreshTokens = snap.RefreshTokens
	}
	if snap.VerifyTokens != nil {
		s.verifyTokens = snap.VerifyTokens
	}
	if snap.Seq > s.seq {
		s.seq = snap.Seq
	}
	return nil
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
