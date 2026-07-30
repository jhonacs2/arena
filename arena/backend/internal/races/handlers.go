package races

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/talentodh/arena/internal/contract"
	"github.com/talentodh/arena/internal/ws"
)

// Handlers es la capa HTTP de las carreras. No tiene lógica: lee el cuerpo,
// llama al servicio y serializa. Todo lo que se valida se valida en el servicio,
// dentro de una transacción.
type Handlers struct {
	svc *Service
	hub *ws.Hub
	log *slog.Logger

	identify      func(*http.Request) (Identity, error)
	identifyToken func(context.Context, string) (Identity, error)
}

// HandlersConfig son las dependencias de la capa HTTP.
//
// Identify e IdentifyToken las provee el paquete de autenticación. Se piden como
// funciones y no como un middleware para que el chequeo de ROL quede de este
// lado: cada handler de /admin/ llama al servicio, que empieza con requireAdmin.
// Así no hay forma de agregar una ruta de instructor y olvidarse del chequeo.
type HandlersConfig struct {
	Service *Service
	Hub     *ws.Hub
	Log     *slog.Logger

	// Identify saca la identidad de la petición HTTP (Authorization: Bearer).
	// Devuelve un error con código UNAUTHENTICATED si no hay sesión válida.
	Identify func(*http.Request) (Identity, error)

	// IdentifyToken valida el token del handshake del WebSocket. El navegador no
	// permite mandar headers en el handshake, así que el token viene por query o
	// en el primer mensaje.
	IdentifyToken func(ctx context.Context, token string) (Identity, error)
}

func NewHandlers(cfg HandlersConfig) *Handlers {
	log := cfg.Log
	if log == nil {
		log = slog.Default()
	}
	return &Handlers{
		svc:           cfg.Service,
		hub:           cfg.Hub,
		log:           log,
		identify:      cfg.Identify,
		identifyToken: cfg.IdentifyToken,
	}
}

// Register registra las rutas de carreras, apuestas y socket en el router.
//
// Router de la biblioteca estándar: desde Go 1.22 http.ServeMux entiende método y
// comodines, así que un router externo solo agregaría una dependencia.
//
// El prefijo /api ya está en las rutas porque es lo que dice el contrato.
func (h *Handlers) Register(mux *http.ServeMux) {
	// Instructor. Todas verifican el rol EN EL SERVIDOR.
	mux.HandleFunc("POST /api/admin/races", h.createRace)
	mux.HandleFunc("PATCH /api/admin/races/{id}", h.patchRace)
	mux.HandleFunc("POST /api/admin/races/{id}/horses", h.addHorses)
	mux.HandleFunc("POST /api/admin/races/{id}/open", h.open)
	mux.HandleFunc("POST /api/admin/races/{id}/start", h.start)
	mux.HandleFunc("POST /api/admin/races/{id}/cancel", h.cancel)

	// Alumno.
	mux.HandleFunc("GET /api/races", h.list)
	mux.HandleFunc("GET /api/races/{id}", h.detail)
	mux.HandleFunc("POST /api/races/{id}/join", h.join)
	mux.HandleFunc("POST /api/races/{id}/bet", h.bet)

	// Socket.
	mux.Handle("GET /api/ws", h.socket())
}

// caller resuelve quién hace la petición. Todos los endpoints de carreras exigen
// sesión: no hay nada público acá.
func (h *Handlers) caller(r *http.Request) (Identity, error) {
	if h.identify == nil {
		return Identity{}, contract.Errorf(contract.CodeUnauthenticated)
	}
	id, err := h.identify(r)
	if err != nil {
		return Identity{}, err
	}
	if id.UserID == "" {
		return Identity{}, contract.Errorf(contract.CodeUnauthenticated)
	}
	return id, nil
}

// ── Instructor ────────────────────────────────────────────────────────────

type createRaceRequest struct {
	Name        string           `json:"name"`
	ScheduledAt *time.Time       `json:"scheduledAt"`
	Horses      []newHorseFields `json:"horses"`
}

type newHorseFields struct {
	Number int    `json:"number"`
	Name   string `json:"name"`
	Odds   int    `json:"odds"` // ×100: 340 es 3.40
}

func toNewHorses(in []newHorseFields) []NewHorse {
	out := make([]NewHorse, len(in))
	for i, h := range in {
		out[i] = NewHorse{Number: h.Number, Name: h.Name, Odds: h.Odds}
	}
	return out
}

func (h *Handlers) createRace(w http.ResponseWriter, r *http.Request) {
	id, err := h.caller(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	var body createRaceRequest
	if err := decodeJSON(r, &body); err != nil {
		h.writeError(w, r, err)
		return
	}

	out, err := h.svc.CreateRace(r.Context(), id, NewRace{
		Name:        body.Name,
		ScheduledAt: body.ScheduledAt,
		Horses:      toNewHorses(body.Horses),
	})
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

type patchRaceRequest struct {
	Name        *string    `json:"name"`
	ScheduledAt *time.Time `json:"scheduledAt"`
}

func (h *Handlers) patchRace(w http.ResponseWriter, r *http.Request) {
	id, err := h.caller(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	var body patchRaceRequest
	if err := decodeJSON(r, &body); err != nil {
		h.writeError(w, r, err)
		return
	}

	out, err := h.svc.PatchRace(r.Context(), id, r.PathValue("id"), RacePatch{
		Name:        body.Name,
		ScheduledAt: body.ScheduledAt,
	})
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

type addHorsesRequest struct {
	Horses []newHorseFields `json:"horses"`
}

func (h *Handlers) addHorses(w http.ResponseWriter, r *http.Request) {
	id, err := h.caller(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	var body addHorsesRequest
	if err := decodeJSON(r, &body); err != nil {
		h.writeError(w, r, err)
		return
	}

	out, err := h.svc.AddHorses(r.Context(), id, r.PathValue("id"), toNewHorses(body.Horses))
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (h *Handlers) open(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, h.svc.Open)
}

func (h *Handlers) start(w http.ResponseWriter, r *http.Request) {
	h.transition(w, r, h.svc.Start)
}

// transition es el cuerpo compartido de /open y /start: las dos leen el id de la
// ruta, no tienen cuerpo, y devuelven la carrera.
func (h *Handlers) transition(w http.ResponseWriter, r *http.Request, apply func(ctx context.Context, id Identity, raceID string) (AdminRaceResponse, error)) {
	id, err := h.caller(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	out, err := apply(r.Context(), id, r.PathValue("id"))
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

type cancelRequest struct {
	// Reason se le muestra a los alumnos en `race.cancelled`, así que va en
	// castellano. Opcional: si no viene, se usa un texto por defecto.
	Reason string `json:"reason"`
}

func (h *Handlers) cancel(w http.ResponseWriter, r *http.Request) {
	id, err := h.caller(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	// El cuerpo es opcional: cancelar sin explicación tiene que funcionar.
	var body cancelRequest
	if r.ContentLength > 0 {
		if err := decodeJSON(r, &body); err != nil {
			h.writeError(w, r, err)
			return
		}
	}

	out, err := h.svc.Cancel(r.Context(), id, r.PathValue("id"), body.Reason)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// ── Alumno ────────────────────────────────────────────────────────────────

func (h *Handlers) list(w http.ResponseWriter, r *http.Request) {
	id, err := h.caller(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	out, err := h.svc.List(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handlers) detail(w http.ResponseWriter, r *http.Request) {
	id, err := h.caller(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	out, err := h.svc.Detail(r.Context(), id, r.PathValue("id"))
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handlers) join(w http.ResponseWriter, r *http.Request) {
	id, err := h.caller(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	out, err := h.svc.Join(r.Context(), id, r.PathValue("id"))
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, JoinResponse{Race: out})
}

type betRequest struct {
	HorseID string `json:"horseId"`
	// Amount es int64: los montos son enteros y nunca float. Ver
	// arena/CLAUDE.md §5.
	Amount int64 `json:"amount"`
}

func (h *Handlers) bet(w http.ResponseWriter, r *http.Request) {
	id, err := h.caller(r)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	var body betRequest
	if err := decodeJSON(r, &body); err != nil {
		h.writeError(w, r, err)
		return
	}

	out, err := h.svc.PlaceBet(r.Context(), id, r.PathValue("id"), body.HorseID, body.Amount)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// ── Socket ────────────────────────────────────────────────────────────────

// socket es `GET /api/ws?raceId=…`.
//
// Adapta las dos funciones que el hub necesita: validar el token y armar los
// eventos de entrada a la sala. El hub no sabe de carreras y el servicio no sabe
// de sockets; el pegamento está acá y es corto a propósito.
func (h *Handlers) socket() http.Handler {
	authenticate := func(ctx context.Context, token string) (ws.Identity, error) {
		if h.identifyToken == nil {
			return ws.Identity{}, contract.Errorf(contract.CodeUnauthenticated)
		}
		id, err := h.identifyToken(ctx, token)
		if err != nil {
			return ws.Identity{}, err
		}
		return ws.Identity{UserID: id.UserID, Username: id.Username, Role: id.Role}, nil
	}

	onJoin := func(ctx context.Context, raceID string, id ws.Identity) (any, any, error) {
		return h.svc.JoinRoom(ctx, raceID, Identity{
			UserID:   id.UserID,
			Username: id.Username,
			Role:     id.Role,
		})
	}

	return h.hub.Handler(authenticate, onJoin)
}
