package api

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/talentodh/arena/internal/accounts"
	"github.com/talentodh/arena/internal/contract"
)

// Todos los handlers de este archivo están detrás del portón de withAdminGate,
// que rechaza /api/admin/* si el rol no es admin. Igual llaman a requireAdmin:
// el portón protege de que alguien agregue una ruta y se olvide el chequeo, y el
// chequeo del handler protege de que alguien mueva la ruta fuera del prefijo.
// Con las monedas valiendo nota, dos candados no son de más.

// uuidPattern es la forma canónica de un uuid. Se valida antes de mandarlo a
// Postgres: un id mal pegado daría un error de tipo 22P02 que saldría como 500,
// y lo que pasó en realidad es que el instructor copió mal.
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// pathUserID lee el {id} de la ruta y lo valida.
func pathUserID(r *http.Request) (string, error) {
	id := r.PathValue("id")
	if !uuidPattern.MatchString(id) {
		return "", contract.Errorf(contract.CodeUserNotFound)
	}
	return id, nil
}

// ── POST /api/admin/codes ─────────────────────────────────────────────────

type createCodesRequest struct {
	Count        int    `json:"count"`
	CoinsGranted int64  `json:"coinsGranted"`
	Note         string `json:"note"`
}

type createCodesResponse struct {
	Codes []string `json:"codes"`
}

func (s *Server) handleCreateCodes(w http.ResponseWriter, r *http.Request) {
	admin, err := requireAdmin(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	var body createCodesRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, r, err)
		return
	}
	if body.Count == 0 {
		body.Count = 1
	}
	if body.CoinsGranted == 0 {
		body.CoinsGranted = accounts.DefaultCoins
	}

	codes, err := s.Accounts.CreateCodes(r.Context(), admin.ID,
		body.Count, body.CoinsGranted, strings.TrimSpace(body.Note))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, createCodesResponse{Codes: codes})
}

// ── GET /api/admin/codes ──────────────────────────────────────────────────

type listCodesResponse struct {
	Items []accounts.Code `json:"items"`
}

func (s *Server) handleListCodes(w http.ResponseWriter, r *http.Request) {
	if _, err := requireAdmin(r); err != nil {
		writeError(w, r, err)
		return
	}

	items, err := s.Accounts.ListCodes(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, listCodesResponse{Items: items})
}

// ── GET /api/admin/scores ─────────────────────────────────────────────────

type scoresResponse struct {
	Items []accounts.Score `json:"items"`
}

// handleScores es el panel de nota. Sale entero de la vista `user_scores`.
//
// Es también la pantalla que reemplaza a la recarga automática: no hay recarga,
// así que el instructor mira esta lista para ver a quién se le acabaron las
// monedas sin que tenga que levantar la mano (decisiones.md §1).
func (s *Server) handleScores(w http.ResponseWriter, r *http.Request) {
	if _, err := requireAdmin(r); err != nil {
		writeError(w, r, err)
		return
	}

	items, err := s.Accounts.Scores(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, scoresResponse{Items: items})
}

// ── POST /api/admin/users/{id}/gift ───────────────────────────────────────

type giftRequest struct {
	Coins int64  `json:"coins"`
	Note  string `json:"note"`
}

type giftResponse struct {
	Balance int64           `json:"balance"`
	Points  accounts.Points `json:"points"`
}

// handleGift regala o descuenta MONEDAS.
//
// Monedas, no puntos: le da con qué seguir jugando, y se pueden volver a perder.
// Para subirle la nota directo está /grant-points, que es otra cosa
// (decisiones.md §1).
//
// Acepta `coins` negativo para un ajuste. Va al ledger con `created_by`, así que
// queda el rastro de quién lo hizo.
func (s *Server) handleGift(w http.ResponseWriter, r *http.Request) {
	admin, err := requireAdmin(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	userID, err := pathUserID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	var body giftRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, r, err)
		return
	}
	if body.Coins == 0 {
		writeError(w, r, contract.FieldErrors(map[string]string{
			"coins": "Un regalo de cero monedas no cambia nada.",
		}))
		return
	}

	balance, err := s.Ledger.Gift(r.Context(), admin.ID, userID, body.Coins, strings.TrimSpace(body.Note))
	if err != nil {
		writeError(w, r, err)
		return
	}

	points, err := s.Accounts.PointsFor(r.Context(), userID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, giftResponse{Balance: balance, Points: points})
}

// ── POST /api/admin/users/{id}/grant-points ───────────────────────────────

type grantPointsRequest struct {
	// Points en centésimas de punto: 250 son 2,5 puntos. Entero en el cable por
	// el mismo motivo que las monedas — ver arena/CLAUDE.md §5.
	Points int64  `json:"points"`
	Reason string `json:"reason"`
}

type grantPointsResponse struct {
	Grant  accounts.PointGrant `json:"grant"`
	Points accounts.Points     `json:"points"`
}

// handleGrantPoints regala PUNTOS: le sube la nota directo.
//
// **No está en api.md.** Se agregó porque decisiones.md §1 lo pide explícitamente
// —«el instructor regala monedas o puntos, y no son lo mismo»— y el esquema ya
// tiene la tabla `point_grants` para eso. Sin este endpoint la tabla queda
// muerta y la mitad de esa decisión no se puede ejercer. Queda anotado como
// pendiente de agregar al contrato.
//
// Un punto regalado no pasa por el juego y no se puede perder apostando: por eso
// vive en `point_grants` y no en el ledger de monedas.
func (s *Server) handleGrantPoints(w http.ResponseWriter, r *http.Request) {
	admin, err := requireAdmin(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	userID, err := pathUserID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	var body grantPointsRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, r, err)
		return
	}

	grant, err := s.Accounts.GrantPoints(r.Context(), admin.ID, userID,
		accounts.Points(body.Points), strings.TrimSpace(body.Reason))
	if err != nil {
		writeError(w, r, err)
		return
	}

	points, err := s.Accounts.PointsFor(r.Context(), userID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, grantPointsResponse{Grant: grant, Points: points})
}
