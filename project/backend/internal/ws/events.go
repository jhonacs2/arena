package ws

import "github.com/talentodh/hipodromo/internal/contract"

// Los eventos del servidor al cliente, tal cual están en
// docs/contract/ws-events.md. Las etiquetas JSON son normativas: hay una
// interfaz TypeScript con exactamente estos nombres en core/ws/ws-events.model.ts.

type Countdown struct {
	Type        string `json:"type"`
	RaceID      string `json:"raceId"`
	SecondsLeft int    `json:"secondsLeft"`
}

func NewCountdown(raceID string, secondsLeft int) Countdown {
	return Countdown{Type: "race.countdown", RaceID: raceID, SecondsLeft: secondsLeft}
}

type Started struct {
	Type      string `json:"type"`
	RaceID    string `json:"raceId"`
	StartedAt string `json:"startedAt"`
}

func NewStarted(raceID, startedAt string) Started {
	return Started{Type: "race.started", RaceID: raceID, StartedAt: startedAt}
}

// Position se repite acá en vez de reusar sim.Position para que el paquete ws
// no dependa del simulador: el hub transporta eventos, no sabe de carreras.
type Position struct {
	HorseID  string  `json:"horseId"`
	Progress float64 `json:"progress"`
	Place    int     `json:"place"`
}

type Tick struct {
	Type      string     `json:"type"`
	RaceID    string     `json:"raceId"`
	T         float64    `json:"t"`
	Positions []Position `json:"positions"`
}

func NewTick(raceID string, t float64, positions []Position) Tick {
	return Tick{Type: "race.tick", RaceID: raceID, T: t, Positions: positions}
}

// FinishedPayout es más chico que contract.Payout a propósito: el evento del
// socket manda solo lo que el cliente necesita para animar el cobro. El detalle
// completo está en GET /races/:id/results.
type FinishedPayout struct {
	BetID  string `json:"betId"`
	Amount int    `json:"amount"`
}

type Finished struct {
	Type   string   `json:"type"`
	RaceID string   `json:"raceId"`
	Podium []string `json:"podium"`
	// Solo las apuestas del usuario que recibe el evento. Por eso Finished se
	// arma una vez por destinatario y no se difunde igual para todos.
	Payouts []FinishedPayout `json:"payouts"`
}

func NewFinished(raceID string, podium []string, payouts []FinishedPayout) Finished {
	if payouts == nil {
		payouts = []FinishedPayout{}
	}
	return Finished{Type: "race.finished", RaceID: raceID, Podium: podium, Payouts: payouts}
}

type BalanceUpdated struct {
	Type    string `json:"type"`
	Balance int    `json:"balance"`
}

func NewBalanceUpdated(balance int) BalanceUpdated {
	return BalanceUpdated{Type: "balance.updated", Balance: balance}
}

type LeaderboardUpdated struct {
	Type    string                      `json:"type"`
	Entries []contract.LeaderboardEntry `json:"entries"`
}

func NewLeaderboardUpdated(entries []contract.LeaderboardEntry) LeaderboardUpdated {
	if entries == nil {
		entries = []contract.LeaderboardEntry{}
	}
	return LeaderboardUpdated{Type: "leaderboard.updated", Entries: entries}
}

type Pong struct {
	Type string `json:"type"`
}

func NewPong() Pong { return Pong{Type: "pong"} }

// clientMessage es lo que manda el cliente: subscribe, unsubscribe o ping.
type clientMessage struct {
	Type   string `json:"type"`
	RaceID string `json:"raceId"`
}
