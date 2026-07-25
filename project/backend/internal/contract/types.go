// Package contract son los tipos del contrato de docs/contract/openapi.yaml.
//
// Las etiquetas JSON son normativas: el frontend Angular tiene una interfaz
// TypeScript con exactamente estos nombres. Si acá cambia un `json:"…"`, se
// rompe el otro lado en silencio — por eso los tests golden de este paquete
// comparan contra los mismos samples que consume el frontend.
package contract

// ── Usuario y sesión ──────────────────────────────────────────────────────

type User struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	DisplayName   string `json:"displayName"`
	Balance       int    `json:"balance"`
	EmailVerified bool   `json:"emailVerified"`
}

type AuthTokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	User         User   `json:"user"`
}

type RefreshedTokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type VerifiedUser struct {
	User User `json:"user"`
}

// ── Carreras ──────────────────────────────────────────────────────────────

type RaceStatus string

const (
	StatusUpcoming RaceStatus = "upcoming"
	StatusLive     RaceStatus = "live"
	StatusFinished RaceStatus = "finished"
)

func (s RaceStatus) Valid() bool {
	return s == StatusUpcoming || s == StatusLive || s == StatusFinished
}

type Horse struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Number int     `json:"number"`
	Odds   float64 `json:"odds"`
}

type Race struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	StartsAt string     `json:"startsAt"`
	Status   RaceStatus `json:"status"`
	// El listado también trae los caballos. Está decidido en
	// docs/contract/README.md: si no, el tipo del frontend tendría que ser
	// opcional y el alumno pelearía con `undefined` desde S1.
	Horses []Horse `json:"horses"`
}

// Favourite devuelve el caballo de menor cuota. Empate por número de partida.
func (r Race) Favourite() (Horse, bool) {
	if len(r.Horses) == 0 {
		return Horse{}, false
	}
	best := r.Horses[0]
	for _, h := range r.Horses[1:] {
		if h.Odds < best.Odds || (h.Odds == best.Odds && h.Number < best.Number) {
			best = h
		}
	}
	return best, true
}

func (r Race) Horse(id string) (Horse, bool) {
	for _, h := range r.Horses {
		if h.ID == id {
			return h, true
		}
	}
	return Horse{}, false
}

type PodiumEntry struct {
	Place     int     `json:"place"`
	HorseID   string  `json:"horseId"`
	HorseName string  `json:"horseName"`
	Number    int     `json:"number"`
	Odds      float64 `json:"odds"`
}

type Payout struct {
	BetID   string `json:"betId"`
	HorseID string `json:"horseId"`
	Stake   int    `json:"stake"`
	Amount  int    `json:"amount"`
}

type RaceResult struct {
	RaceID     string        `json:"raceId"`
	FinishedAt string        `json:"finishedAt"`
	Podium     []PodiumEntry `json:"podium"`
	// Solo las apuestas del usuario autenticado. Sin sesión va vacío — el
	// podio es público, los pagos no.
	Payouts []Payout `json:"payouts"`
}

// ── Apuestas ──────────────────────────────────────────────────────────────

type BetStatus string

const (
	BetPending BetStatus = "pending"
	BetWon     BetStatus = "won"
	BetLost    BetStatus = "lost"
)

// Bet denormaliza RaceName y HorseName a propósito: el historial se pinta sin
// ir a buscar la carrera ni el caballo.
type Bet struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	RaceID    string    `json:"raceId"`
	RaceName  string    `json:"raceName"`
	HorseID   string    `json:"horseId"`
	HorseName string    `json:"horseName"`
	Amount    int       `json:"amount"`
	Odds      float64   `json:"odds"`
	Status    BetStatus `json:"status"`
	Payout    int       `json:"payout"`
	PlacedAt  string    `json:"placedAt"`
}

type BetCreated struct {
	Bet     Bet `json:"bet"`
	Balance int `json:"balance"`
}

type LeaderboardEntry struct {
	Rank        int    `json:"rank"`
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
	Profit      int    `json:"profit"`
	Bets        int    `json:"bets"`
	Wins        int    `json:"wins"`
}

type Leaderboard struct {
	Period  string             `json:"period"`
	Entries []LeaderboardEntry `json:"entries"`
}

// ── Paginación ────────────────────────────────────────────────────────────

type Page[T any] struct {
	Items []T `json:"items"`
	Page  int `json:"page"`
	Size  int `json:"size"`
	// Total de elementos que matchean el filtro, no de la página.
	Total int `json:"total"`
}

// NewPage arma una página aplicando el recorte. Items nunca es nil: un `null`
// en vez de `[]` rompe el `@for` del frontend.
func NewPage[T any](all []T, page, size int) Page[T] {
	total := len(all)
	start := (page - 1) * size
	if start > total {
		start = total
	}
	end := start + size
	if end > total {
		end = total
	}
	items := make([]T, 0, end-start)
	items = append(items, all[start:end]...)
	return Page[T]{Items: items, Page: page, Size: size, Total: total}
}

// ── Límites del dominio ───────────────────────────────────────────────────

const (
	MinBetAmount   = 10
	MaxBetAmount   = 5000
	SignupBalance  = 1000 // saldo virtual que se otorga al registrarse
	LeaderboardTop = 20
)
