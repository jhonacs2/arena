package api

import (
	"net/http"

	"github.com/talentodh/hipodromo/internal/contract"
	"github.com/talentodh/hipodromo/internal/ws"
)

type createBetRequest struct {
	RaceID  string `json:"raceId"`
	HorseID string `json:"horseId"`
	Amount  int    `json:"amount"`
}

func (s *Server) handleCreateBet(w http.ResponseWriter, r *http.Request) {
	id, err := requireUser(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	var body createBetRequest
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, r, err)
		return
	}

	fields := map[string]string{}
	if body.RaceID == "" {
		fields["raceId"] = "Elegí una carrera."
	}
	if body.HorseID == "" {
		fields["horseId"] = "Elegí un caballo."
	}
	if len(fields) > 0 {
		writeError(w, r, contract.FieldErrors(fields))
		return
	}

	// El store valida el resto y descuenta el saldo bajo el mismo candado:
	// no puede quedar una apuesta creada sin su descuento.
	bet, balance, err := s.Store.PlaceBet(id, body.RaceID, body.HorseID, body.Amount, s.now())
	if err != nil {
		writeError(w, r, err)
		return
	}

	// El saldo también sale por el socket: si el usuario tiene otra pestaña
	// abierta, el widget de saldo se actualiza ahí también.
	s.Hub.ToUser(id, ws.NewBalanceUpdated(balance))

	writeJSON(w, http.StatusCreated, contract.BetCreated{Bet: bet, Balance: balance})
}

func (s *Server) handleMyBets(w http.ResponseWriter, r *http.Request) {
	id, err := requireUser(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	page, size := pagination(r)
	writeJSON(w, http.StatusOK, contract.NewPage(s.Store.BetsByUser(id), page, size))
}

func (s *Server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "all"
	}
	if period != "all" && period != "daily" {
		writeError(w, r, contract.FieldErrors(map[string]string{
			"period": "Los valores válidos son daily o all.",
		}))
		return
	}

	writeJSON(w, http.StatusOK, contract.Leaderboard{
		Period:  period,
		Entries: s.Store.Leaderboard(period, s.now()),
	})
}
