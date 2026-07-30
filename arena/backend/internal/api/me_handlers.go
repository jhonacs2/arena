package api

import (
	"net/http"

	"github.com/talentodh/arena/internal/accounts"
	"github.com/talentodh/arena/internal/ledger"
)

type meResponse struct {
	User    accounts.User   `json:"user"`
	Balance int64           `json:"balance"`
	Points  accounts.Points `json:"points"`
}

// handleMe es el saldo y la nota del alumno.
//
// El saldo sale de `users.balance` y los puntos de la vista `user_scores`, que es
// la única que sabe la fórmula `floor(monedas/100) + regalados`. No se calcula en
// Go: si la fórmula viviera en dos lados, un día se separarían y la nota dependería
// de por dónde se la mire.
func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user, err := requireUser(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	points, err := s.Accounts.PointsFor(r.Context(), user.ID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, meResponse{User: user, Balance: user.Balance, Points: points})
}

type transactionsResponse struct {
	Items []ledger.Transaction `json:"items"`
}

// handleMyTransactions es el historial del ledger del alumno, más nuevo primero.
//
// Es lo que le permite entender por qué tiene la nota que tiene, y es la razón
// por la que el ledger es append-only: si se pudiera editar, esta pantalla no
// probaría nada.
//
// Solo el propio: no hay forma de pedir el de otro. El del resto lo ve el
// instructor por /api/admin/scores.
func (s *Server) handleMyTransactions(w http.ResponseWriter, r *http.Request) {
	user, err := requireUser(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	items, err := s.Ledger.Transactions(r.Context(), user.ID,
		intQuery(r, "limit", ledger.DefaultHistoryLimit),
		intQuery(r, "offset", 0))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, transactionsResponse{Items: items})
}
