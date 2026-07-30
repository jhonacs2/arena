package races

import (
	"time"

	"github.com/talentodh/arena/internal/sim"
)

// Las formas que salen al cable, tal cual están en docs/contract/api.md. Las
// etiquetas JSON son normativas: el frontend tiene interfaces con exactamente
// estos nombres.
//
// Son tipos aparte de las filas del dominio a propósito. Si el handler
// serializara la fila, agregar una columna la publicaría sin que nadie lo
// decida — y una de esas columnas es a qué caballo apostó cada uno.

// ── Fechas ────────────────────────────────────────────────────────────────

// stamp formatea una fecha opcional. Nula queda `null`, no cadena vacía: el
// frontend distingue «todavía no largó» de «largó en el instante cero».
func stamp(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}

// ── Caballos ──────────────────────────────────────────────────────────────

type HorseView struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
	Name   string `json:"name"`
	// NominalOdds es ×100: 340 es 3.40.
	NominalOdds int `json:"nominalOdds"`
}

func horseViews(horses []Horse) []HorseView {
	out := make([]HorseView, 0, len(horses))
	for _, h := range horses {
		out = append(out, HorseView{ID: h.ID, Number: h.Number, Name: h.Name, NominalOdds: h.NominalOdds})
	}
	return out
}

// ── Apuestas ──────────────────────────────────────────────────────────────

// PublicBetView es una apuesta como la ve EL RESTO de la sala.
//
// HorseID y Status llevan `omitempty` y se llenan solo cuando el estado de la
// carrera lo permite: ver betViews, que es el único lugar donde se decide.
type PublicBetView struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Amount   int64  `json:"amount"`

	HorseID string `json:"horseId,omitempty"`
	Status  string `json:"status,omitempty"`
}

// MyBetView es la apuesta de quien recibe la respuesta o el evento. Trae el
// detalle completo, incluido el pago: es SUYA.
//
// NO trae un pago potencial. Con pari-mutuel no existe —depende de cómo apueste
// el resto— y con cuotas fijas saldría de una cuota congelada que hoy no se
// guarda. Payout es nulo hasta que la carrera se liquida.
type MyBetView struct {
	ID        string  `json:"id"`
	HorseID   string  `json:"horseId"`
	Amount    int64   `json:"amount"`
	Status    string  `json:"status"`
	Payout    *int64  `json:"payout"`
	CreatedAt string  `json:"createdAt"`
	SettledAt *string `json:"settledAt"`
}

func myBetView(b Bet) MyBetView {
	return MyBetView{
		ID:        b.ID,
		HorseID:   b.HorseID,
		Amount:    b.Amount,
		Status:    string(b.Status),
		Payout:    b.Payout,
		CreatedAt: b.CreatedAt.UTC().Format(time.RFC3339),
		SettledAt: stamp(b.SettledAt),
	}
}

func myBetViewPtr(b Bet, found bool) *MyBetView {
	if !found {
		return nil
	}
	v := myBetView(b)
	return &v
}

// betViews arma la lista de apuestas que ve el resto de la sala.
//
// ES EL ÚNICO LUGAR donde se decide si se revela el caballo, y por eso es una
// función y no tres bloques repetidos: mientras la carrera está `open` NO se
// incluye, porque si se incluyera los últimos en apostar copiarían a los
// primeros y la apuesta dejaría de medir criterio. Al pasar a `running` se
// revelan todas juntas.
//
// El pago NUNCA sale por acá, ni con la carrera terminada: cuánto cobró cada uno
// es de cada uno. El propio va por MyBetView, que se arma por destinatario.
func betViews(status Status, bets []Bet) []PublicBetView {
	reveal := status.RevealsBets()

	out := make([]PublicBetView, 0, len(bets))
	for _, b := range bets {
		view := PublicBetView{UserID: b.UserID, Username: b.Username, Amount: b.Amount}
		if reveal {
			view.HorseID = b.HorseID
			view.Status = string(b.Status)
		}
		out = append(out, view)
	}
	return out
}

// ── Sala ──────────────────────────────────────────────────────────────────

type ParticipantView struct {
	UserID    string `json:"userId"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	JoinedAt  string `json:"joinedAt"`
}

func participantViews(ps []Participant) []ParticipantView {
	out := make([]ParticipantView, 0, len(ps))
	for _, p := range ps {
		out = append(out, ParticipantView{
			UserID:    p.UserID,
			Username:  p.Username,
			FirstName: p.FirstName,
			LastName:  p.LastName,
			JoinedAt:  p.JoinedAt.UTC().Format(time.RFC3339),
		})
	}
	return out
}

// ── Carreras ──────────────────────────────────────────────────────────────

// RaceView es una carrera en el listado.
type RaceView struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Status           Status     `json:"status"`
	ScheduledAt      *string    `json:"scheduledAt"`
	HorseCount       int        `json:"horseCount"`
	ParticipantCount int        `json:"participantCount"`
	MyBet            *MyBetView `json:"myBet"`
}

func raceView(s Summary) RaceView {
	view := RaceView{
		ID:               s.ID,
		Name:             s.Name,
		Status:           s.Status,
		ScheduledAt:      stamp(s.ScheduledAt),
		HorseCount:       s.HorseCount,
		ParticipantCount: s.ParticipantCount,
	}
	if s.MyBet != nil {
		view.MyBet = myBetViewPtr(*s.MyBet, true)
	}
	return view
}

// RaceListResponse es GET /api/races.
type RaceListResponse struct {
	Items []RaceView `json:"items"`
}

// RaceDetailView es GET /api/races/:id y también la carga de `room.state`: el
// detalle que se pide por HTTP y el que llega al conectarse al socket son la
// misma cosa, así el cliente tiene un solo modelo.
type RaceDetailView struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Status      Status  `json:"status"`
	ScheduledAt *string `json:"scheduledAt"`
	OpenedAt    *string `json:"openedAt"`
	StartedAt   *string `json:"startedAt"`
	FinishedAt  *string `json:"finishedAt"`

	Horses       []HorseView       `json:"horses"`
	Participants []ParticipantView `json:"participants"`

	// Bets son las apuestas del resto. Sin el caballo mientras la carrera está
	// `open` — ver betViews.
	Bets []PublicBetView `json:"bets"`

	// MyBet es la propia, con detalle completo. Nula si no apostó.
	MyBet *MyBetView `json:"myBet"`

	// Results están solo si la carrera terminó.
	Results []ResultView `json:"results"`
}

type ResultView struct {
	HorseID     string `json:"horseId"`
	HorseName   string `json:"horseName"`
	Number      int    `json:"number"`
	NominalOdds int    `json:"nominalOdds"`
	Position    int    `json:"position"`
}

func resultViews(rs []sim.Result) []ResultView {
	out := make([]ResultView, 0, len(rs))
	for _, r := range rs {
		out = append(out, ResultView{
			HorseID:     r.HorseID,
			HorseName:   r.HorseName,
			Number:      r.Number,
			NominalOdds: r.Odds,
			Position:    r.Position,
		})
	}
	return out
}

// AdminRaceResponse es lo que devuelven POST /api/admin/races y las
// transiciones: la carrera con sus caballos.
type AdminRaceResponse struct {
	Race AdminRaceView `json:"race"`
}

type AdminRaceView struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Status      Status      `json:"status"`
	ScheduledAt *string     `json:"scheduledAt"`
	OpenedAt    *string     `json:"openedAt"`
	StartedAt   *string     `json:"startedAt"`
	FinishedAt  *string     `json:"finishedAt"`
	Seed        *int64      `json:"seed"`
	Horses      []HorseView `json:"horses"`
}

func adminRaceView(r Race, horses []Horse) AdminRaceView {
	return AdminRaceView{
		ID:          r.ID,
		Name:        r.Name,
		Status:      r.Status,
		ScheduledAt: stamp(r.ScheduledAt),
		OpenedAt:    stamp(r.OpenedAt),
		StartedAt:   stamp(r.StartedAt),
		FinishedAt:  stamp(r.FinishedAt),
		Seed:        r.Seed,
		Horses:      horseViews(horses),
	}
}

// ── Apostar ───────────────────────────────────────────────────────────────

// BetResponse es POST /api/races/:id/bet.
//
// Devuelve el saldo resultante, que es lo que el frontend necesita para actualizar
// el widget sin volver a pedir nada. NO devuelve los puntos: la fórmula tiene un
// piso en discusión y suma los puntos regalados, que este paquete no conoce.
// Salen de `GET /api/me`.
type BetResponse struct {
	Bet     MyBetView `json:"bet"`
	Balance int64     `json:"balance"`
}

func betResponse(b Bet, balance int64) BetResponse {
	return BetResponse{Bet: myBetView(b), Balance: balance}
}

// JoinResponse es POST /api/races/:id/join.
type JoinResponse struct {
	Race RaceDetailView `json:"race"`
}
