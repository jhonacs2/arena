package races

import (
	"time"

	"github.com/talentodh/arena/internal/sim"
)

// Los eventos servidor → cliente de docs/contract/api.md. Los tipos viven en
// este paquete y no en internal/ws porque el hub transporta eventos y no sabe de
// carreras — igual que en el backend del hipódromo.
//
// El cliente no manda nada salvo el handshake: apostar es un POST, no un mensaje
// de socket.

const (
	EventRoomState     = "room.state"
	EventRoomJoined    = "room.joined"
	EventBetPlaced     = "bet.placed"
	EventRaceStarted   = "race.started"
	EventRaceTick      = "race.tick"
	EventRaceFinished  = "race.finished"
	EventRaceCancelled = "race.cancelled"
)

// RoomState llega al conectarse: la sala completa.
//
// La carga es el mismo RaceDetailView que devuelve GET /api/races/:id, así que
// pasa por el mismo betViews y tampoco revela el caballo mientras la carrera
// está `open`. Se arma POR DESTINATARIO, porque incluye `myBet`.
type RoomState struct {
	Type   string         `json:"type"`
	RaceID string         `json:"raceId"`
	Race   RaceDetailView `json:"race"`
}

func NewRoomState(raceID string, race RaceDetailView) RoomState {
	return RoomState{Type: EventRoomState, RaceID: raceID, Race: race}
}

// RoomJoined se difunde cuando alguien entra a la sala.
type RoomJoined struct {
	Type             string `json:"type"`
	RaceID           string `json:"raceId"`
	UserID           string `json:"userId"`
	Username         string `json:"username"`
	ParticipantCount int    `json:"participantCount"`
}

func NewRoomJoined(raceID, userID, username string, count int) RoomJoined {
	return RoomJoined{
		Type: EventRoomJoined, RaceID: raceID,
		UserID: userID, Username: username, ParticipantCount: count,
	}
}

// BetPlaced se difunde cuando alguien apuesta.
//
// NO lleva el caballo. Es la regla de decisiones.md §3 y no es cosmética: si lo
// llevara, los últimos en apostar copiarían a los primeros. El campo no está en
// la estructura, así que no hay forma de filtrarlo por descuido — no alcanza con
// acordarse de no llenarlo.
type BetPlaced struct {
	Type     string `json:"type"`
	RaceID   string `json:"raceId"`
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Amount   int64  `json:"amount"`
	BetCount int    `json:"betCount"`
}

func NewBetPlaced(raceID, userID, username string, amount int64, betCount int) BetPlaced {
	return BetPlaced{
		Type: EventBetPlaced, RaceID: raceID,
		UserID: userID, Username: username, Amount: amount, BetCount: betCount,
	}
}

// RaceStarted se difunde al largar.
//
// Lleva `bets` con TODAS las apuestas ya reveladas: es el momento en que se
// revelan juntas, y este es el único empujón que recibe el cliente entre el
// cierre de apuestas y el primer tick. api.md lista `{ startedAt }` como carga;
// `bets` es el agregado que hace cumplible la regla de la revelación —
// documentado en el reporte de la sesión.
type RaceStarted struct {
	Type      string          `json:"type"`
	RaceID    string          `json:"raceId"`
	StartedAt string          `json:"startedAt"`
	Bets      []PublicBetView `json:"bets"`
}

func NewRaceStarted(raceID string, startedAt time.Time, bets []PublicBetView) RaceStarted {
	return RaceStarted{
		Type: EventRaceStarted, RaceID: raceID,
		StartedAt: startedAt.UTC().Format(time.RFC3339),
		Bets:      bets,
	}
}

// TickPosition se repite acá en vez de reusar sim.Position para que el cable no
// quede atado a la estructura interna del simulador.
type TickPosition struct {
	HorseID  string  `json:"horseId"`
	Progress float64 `json:"progress"`
	Place    int     `json:"place"`
}

// RaceTick sale a 10 Hz mientras la carrera corre.
type RaceTick struct {
	Type      string         `json:"type"`
	RaceID    string         `json:"raceId"`
	T         float64        `json:"t"`
	Positions []TickPosition `json:"positions"`
}

func NewRaceTick(raceID string, tick sim.Tick) RaceTick {
	positions := make([]TickPosition, len(tick.Positions))
	for i, p := range tick.Positions {
		positions[i] = TickPosition{HorseID: p.HorseID, Progress: p.Progress, Place: p.Place}
	}
	return RaceTick{Type: EventRaceTick, RaceID: raceID, T: tick.T, Positions: positions}
}

// RaceFinished se manda al terminar, ARMADO POR DESTINATARIO.
//
// `myBet`, `balance` y `points` son de quien lo recibe. Difundir el mismo objeto
// a todos filtraría cuánto cobró cada uno — mismo motivo que en el backend del
// hipódromo. Por eso sale por Broadcaster.ToRoomPerUser y no por ToRoom.
type RaceFinished struct {
	Type    string       `json:"type"`
	RaceID  string       `json:"raceId"`
	Results []ResultView `json:"results"`

	// Bets son las del resto, ya reveladas (la carrera terminó) pero sin el
	// pago de nadie.
	Bets []PublicBetView `json:"bets"`

	// MyBet es la de quien recibe el evento, liquidada. Nula si no apostó.
	MyBet *MyBetView `json:"myBet"`

	// Balance es el de quien recibe el evento, después de cobrar. Nulo si no
	// apostó y por lo tanto su saldo no cambió.
	//
	// Sin `points`: la fórmula tiene un piso en discusión y suma los puntos
	// regalados, que este paquete no conoce. Salen de `GET /api/me`.
	Balance *int64 `json:"balance"`
}

// RaceCancelled se manda al cancelar, TAMBIÉN POR DESTINATARIO: lleva la
// devolución de quien lo recibe.
type RaceCancelled struct {
	Type   string `json:"type"`
	RaceID string `json:"raceId"`
	Reason string `json:"reason"`

	// MyRefund es lo que se le devolvió a quien recibe el evento. Nulo si no
	// había apostado. La devolución es ÍNTEGRA: exactamente lo apostado.
	MyRefund *int64 `json:"myRefund"`
	Balance  *int64 `json:"balance"`
}

// balanceOf arma el saldo de un destinatario. Devuelve nil cuando su saldo no
// cambió, para que el cliente sepa que no tiene que tocar el widget.
func balanceOf(balance int64, changed bool) *int64 {
	if !changed {
		return nil
	}
	return &balance
}
