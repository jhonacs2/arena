// Package ws es el hub de WebSocket: una conexión por cliente, multiplexada por
// sala.
//
// Implementa `GET /api/ws?raceId=…` de docs/contract/api.md. Sigue la forma del
// hub del hipódromo (project/backend/internal/ws) a propósito.
//
// El hub TRANSPORTA eventos y no sabe de carreras: los tipos de los eventos
// viven en internal/races. Lo único que sabe es a qué sala pertenece cada
// conexión y cómo mandarle bytes.
//
// El token va por query string o en el primer mensaje, y no por header, porque la
// API WebSocket del navegador no permite mandar headers en el handshake.
package ws

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const (
	// writeWait es cuánto se espera por un envío antes de dar la conexión por
	// perdida. Con ticks a 10 Hz, un cliente que no lee tiene que caer rápido o
	// se acumula memoria.
	writeWait = 5 * time.Second

	// readLimit acota el mensaje del cliente. Los únicos mensajes válidos son
	// dos objetos JSON diminutos.
	readLimit = 4 << 10

	// sendBuffer: un cliente lento tolera un segundo y medio de ticks antes de
	// que se lo desconecte. Más buffer sería mostrarle una carrera vieja.
	sendBuffer = 16

	// maxConnectionsPerUser: abrir una pestaña por carrera es normal; veinte es
	// una fuga.
	maxConnectionsPerUser = 8

	// handshakeWait es cuánto se espera el token cuando no vino por query. Un
	// socket abierto sin autenticar no puede quedarse abierto.
	handshakeWait = 5 * time.Second
)

// Códigos de cierre. 4001 le dice al cliente que refresque el token y reconecte
// una vez, en lugar de entrar en backoff contra un servidor que está bien.
const (
	CloseUnauthorized = websocket.StatusCode(4001)
	CloseTooMany      = websocket.StatusCode(4029)
	CloseBadRequest   = websocket.StatusCode(4000)
)

// Identity es quién está del otro lado. La resuelve el paquete de autenticación
// por la función Authenticate.
//
// Lleva el rol porque el hub multiplexa por sala y quién puede entrar a una sala
// depende del rol: una carrera en `draft` la ve el instructor y no el alumno. El
// hub no interpreta el rol, solo lo pasa al JoinHook.
type Identity struct {
	UserID   string
	Username string
	Role     string
}

// Authenticate valida el token del handshake. La provee el cableado: el hub no
// sabe de JWT.
//
// Recibe contexto porque validar un token puede tocar la base —el usuario puede
// haber sido dado de baja después de que se emitió— y esa consulta tiene que
// morir con la conexión.
type Authenticate func(ctx context.Context, token string) (Identity, error)

// JoinHook se llama con la conexión ya establecida y admitida en la sala.
//
// Devuelve dos eventos: `self` va solo a quien se acaba de conectar
// (`room.state`, que incluye su propia apuesta y por eso se arma por
// destinatario) y `room` se difunde al resto (`room.joined`). Cualquiera de los
// dos puede ser nil.
type JoinHook func(ctx context.Context, raceID string, id Identity) (self any, room any, err error)

type client struct {
	id     Identity
	raceID string
	conn   *websocket.Conn
	send   chan []byte
}

// Hub reparte eventos entre los clientes conectados.
type Hub struct {
	log *slog.Logger

	mu      sync.RWMutex
	rooms   map[string]map[*client]bool
	byUser  map[string]map[*client]bool
	clients int
}

func NewHub(log *slog.Logger) *Hub {
	if log == nil {
		log = slog.Default()
	}
	return &Hub{
		log:    log,
		rooms:  map[string]map[*client]bool{},
		byUser: map[string]map[*client]bool{},
	}
}

// Handler es `GET /api/ws?raceId=…`.
func (h *Hub) Handler(authenticate Authenticate, onJoin JoinHook) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raceID := r.URL.Query().Get("raceId")
		if raceID == "" {
			http.Error(w, "falta raceId", http.StatusBadRequest)
			return
		}

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			// El origen lo controla el middleware CORS del router; acá se acepta
			// cualquiera porque el token es lo que autentica al cliente.
			InsecureSkipVerify: true,
		})
		if err != nil {
			h.log.Debug("no se pudo aceptar el WebSocket", "error", err)
			return
		}
		conn.SetReadLimit(readLimit)

		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		token := r.URL.Query().Get("token")
		if token == "" {
			// No vino por query: se espera el mensaje de handshake, con plazo.
			// Un socket sin autenticar no se queda abierto.
			token = readToken(ctx, conn)
		}

		id, err := authenticate(ctx, token)
		if err != nil || id.UserID == "" {
			conn.Close(CloseUnauthorized, "token inválido o vencido")
			return
		}

		h.serve(ctx, cancel, conn, raceID, id, onJoin)
	}
}

// tokenMessage es el handshake: `{ "type": "auth", "token": "…" }`.
type tokenMessage struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

func readToken(ctx context.Context, conn *websocket.Conn) string {
	ctx, cancel := context.WithTimeout(ctx, handshakeWait)
	defer cancel()

	_, raw, err := conn.Read(ctx)
	if err != nil {
		return ""
	}
	var msg tokenMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		return ""
	}
	return msg.Token
}

func (h *Hub) serve(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn, raceID string, id Identity, onJoin JoinHook) {
	c := &client{id: id, raceID: raceID, conn: conn, send: make(chan []byte, sendBuffer)}

	if !h.add(c) {
		conn.Close(CloseTooMany, "demasiadas conexiones abiertas")
		return
	}
	defer h.remove(c)

	go h.writeLoop(ctx, c, cancel)

	if onJoin != nil {
		self, room, err := onJoin(ctx, raceID, id)
		if err != nil {
			// La sala no existe o el alumno no puede verla. Se cierra con 4000:
			// no es un problema de token y reconectar no lo va a arreglar.
			h.log.Debug("no se pudo entrar a la sala", "carrera", raceID, "usuario", id.UserID, "error", err)
			conn.Close(CloseBadRequest, "no se pudo entrar a la sala")
			return
		}
		if self != nil {
			h.enqueue(c, self)
		}
		if room != nil {
			h.toRoomExcept(raceID, c, room)
		}
	}

	h.readLoop(ctx, c)
}

func (h *Hub) add(c *client) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.byUser[c.id.UserID]) >= maxConnectionsPerUser {
		return false
	}
	if h.rooms[c.raceID] == nil {
		h.rooms[c.raceID] = map[*client]bool{}
	}
	h.rooms[c.raceID][c] = true

	if h.byUser[c.id.UserID] == nil {
		h.byUser[c.id.UserID] = map[*client]bool{}
	}
	h.byUser[c.id.UserID][c] = true
	h.clients++
	return true
}

func (h *Hub) remove(c *client) {
	h.mu.Lock()
	if h.rooms[c.raceID][c] {
		delete(h.rooms[c.raceID], c)
		if len(h.rooms[c.raceID]) == 0 {
			delete(h.rooms, c.raceID)
		}
		delete(h.byUser[c.id.UserID], c)
		if len(h.byUser[c.id.UserID]) == 0 {
			delete(h.byUser, c.id.UserID)
		}
		h.clients--
		close(c.send)
	}
	h.mu.Unlock()

	c.conn.CloseNow()
}

// readLoop atiende lo poco que manda el cliente.
//
// El cliente no manda nada salvo el handshake y el ping: APOSTAR ES UN POST, no
// un mensaje de socket. Si llegara un `bet` por acá, no habría transacción ni
// validación de rol y sería el agujero más grande de la app.
func (h *Hub) readLoop(ctx context.Context, c *client) {
	for {
		_, raw, err := c.conn.Read(ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) && websocket.CloseStatus(err) == -1 {
				h.log.Debug("lectura del socket cortada", "usuario", c.id.UserID, "error", err)
			}
			return
		}

		var msg struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue // un mensaje mal formado se ignora, no tira la conexión
		}
		if msg.Type == "ping" {
			h.enqueue(c, map[string]string{"type": "pong"})
		}
	}
}

func (h *Hub) writeLoop(ctx context.Context, c *client, cancel context.CancelFunc) {
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		case payload, open := <-c.send:
			if !open {
				return
			}
			writeCtx, done := context.WithTimeout(ctx, writeWait)
			err := c.conn.Write(writeCtx, websocket.MessageText, payload)
			done()
			if err != nil {
				return
			}
		}
	}
}

// enqueue serializa el evento y lo encola. Si el buffer está lleno, el cliente
// no está leyendo: se lo desconecta en lugar de acumular ticks viejos.
func (h *Hub) enqueue(c *client, event any) {
	payload, err := json.Marshal(event)
	if err != nil {
		h.log.Error("no se pudo serializar un evento", "error", err)
		return
	}
	select {
	case c.send <- payload:
	default:
		h.log.Debug("cliente lento, se lo desconecta", "usuario", c.id.UserID)
		go h.remove(c)
	}
}

// ── Difusión ──────────────────────────────────────────────────────────────

func (h *Hub) room(raceID string) []*client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	targets := make([]*client, 0, len(h.rooms[raceID]))
	for c := range h.rooms[raceID] {
		targets = append(targets, c)
	}
	return targets
}

// ToRoom manda el mismo evento a todos los conectados a una carrera.
func (h *Hub) ToRoom(raceID string, event any) {
	for _, c := range h.room(raceID) {
		h.enqueue(c, event)
	}
}

func (h *Hub) toRoomExcept(raceID string, skip *client, event any) {
	for _, c := range h.room(raceID) {
		if c == skip {
			continue
		}
		h.enqueue(c, event)
	}
}

// ToUser manda un evento a todas las conexiones de un usuario, en cualquier sala.
func (h *Hub) ToUser(userID string, event any) {
	h.mu.RLock()
	targets := make([]*client, 0, 4)
	for c := range h.byUser[userID] {
		targets = append(targets, c)
	}
	h.mu.RUnlock()

	for _, c := range targets {
		h.enqueue(c, event)
	}
}

// ToRoomPerUser manda a cada conectado un evento armado PARA ÉL.
//
// Lo usan `race.finished` y `race.cancelled`, que traen el pago o la devolución
// de quien los recibe: difundir el mismo objeto a todos filtraría cuánto cobró
// cada uno.
//
// `build` se llama una vez por conexión. Si un usuario tiene dos pestañas
// abiertas se llama dos veces con el mismo id, y eso está bien: el evento es el
// mismo y las dos pestañas tienen que verlo.
func (h *Hub) ToRoomPerUser(raceID string, build func(userID string) any) {
	for _, c := range h.room(raceID) {
		h.enqueue(c, build(c.id.UserID))
	}
}

// Connections es cuántas conexiones hay abiertas. La expone /health.
func (h *Hub) Connections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients
}

// RoomSize es cuántas conexiones hay en una sala.
func (h *Hub) RoomSize(raceID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms[raceID])
}
