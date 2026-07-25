package api

import (
	"net/http"

	"github.com/talentodh/hipodromo/internal/contract"
)

func (s *Server) handleRaces(w http.ResponseWriter, r *http.Request) {
	status := contract.RaceStatus(r.URL.Query().Get("status"))
	if status != "" && !status.Valid() {
		writeError(w, r, contract.FieldErrors(map[string]string{
			"status": "Los valores válidos son upcoming, live o finished.",
		}))
		return
	}

	page, size := pagination(r)
	writeJSON(w, http.StatusOK, contract.NewPage(s.Store.Races(status), page, size))
}

func (s *Server) handleRace(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	race, found := s.Store.Race(id)
	if !found {
		writeError(w, r, contract.ErrorWith(contract.CodeNotFound,
			map[string]any{"resource": "race", "id": id}))
		return
	}
	writeJSON(w, http.StatusOK, race)
}

func (s *Server) handleRaceResults(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	race, found := s.Store.Race(id)
	if !found {
		writeError(w, r, contract.ErrorWith(contract.CodeNotFound,
			map[string]any{"resource": "race", "id": id}))
		return
	}
	if race.Status != contract.StatusFinished {
		writeError(w, r, contract.ErrorWith(contract.CodeResultsNotAvailable,
			map[string]any{"status": string(race.Status)}))
		return
	}

	// Sin sesión devuelve el podio con payouts vacío: el podio es público, lo
	// que cobró cada uno no. Por eso esta ruta no exige autenticación.
	result, found := s.Store.Result(id, userID(r))
	if !found {
		writeError(w, r, contract.ErrorWith(contract.CodeResultsNotAvailable,
			map[string]any{"status": string(race.Status)}))
		return
	}
	writeJSON(w, http.StatusOK, result)
}
