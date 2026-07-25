// Package ws es el hub de WebSocket: una conexión por cliente, multiplexada
// por sala.
//
// El token va por query string y no por header. No es una decisión de diseño:
// la API WebSocket del navegador no permite mandar headers en el handshake, y
// por eso está documentado en docs/contract/ws-events.md — es una de las cosas
// que el alumno pregunta en S10.
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
	"github.com/talentodh/hipodromo/internal/auth"
)

const (
	// writeWait es cuánto se espera por un envío antes de dar la conexión por
	// perdida. Con ticks a 10 Hz, un cliente que no lee tiene que caer rápido
	// o se acumula memoria.
	writeWait = 5 * time.Second

	// readLimit acota el mensaje del cliente. Los únicos mensajes válidos son
	// tres objetos JSON diminutos.
	readLimit = 4 << 10

	// sendBuffer: un cliente lento tolera un segundo de ticks antes de que se
	// lo desconecte. Más buffer sería mostrarle una carrera vieja.
	sendBuffer = 16

	// maxConnectionsPerUser: abrir una pestaña por carrera es normal; veinte
	// es una fuga.
	maxConnectionsPerUser = 8
)

// Códigos de cierre del contrato.
const (
	CloseUnauthorized = websocket.StatusCode(4001)
	CloseTooMany      = websocket.StatusCode(4029)
)

type client struct {
	userID string
	conn   *websocket.Conn
	send   chan []byte

	mu    sync.Mutex
	rooms map[string]bool
}

func (c *client) inRoom(raceID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.rooms[raceID]
}

// Hub reparte eventos entre los clientes conectados.
type Hub struct {
	log    *slog.Logger
	signer *auth.Signer

	mu      sync.RWMutex
	clients map[*client]bool
	byUser  map[string]map[*client]bool
}

func NewHub(log *slog.Logger, signer *auth.Signer) *Hub {
	return &Hub{
		log:     log,
		signer:  signer,
		clients: map[*client]bool{},
		byUser:  map[string]map[*client]bool{},
	}
}

// Handler acepta la conexión, valida el token y atiende al cliente.
func (h *Hub) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "falta el token", http.StatusUnauthorized)
			return
		}

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			// El origen lo controla el middleware CORS del router; acá se
			// acepta cualquiera porque el token ya autentica al cliente.
			InsecureSkipVerify: true,
		})
		if err != nil {
			h.log.Debug("no se pudo aceptar el WebSocket", "error", err)
			return
		}

		claims, err := h.signer.Parse(token, time.Now())
		if err != nil {
			// Cierre 4001: el cliente sabe que tiene que refrescar el token y
			// reconectar una vez, en lugar de entrar en backoff.
			conn.Close(CloseUnauthorized, "token inválido o vencido")
			return
		}

		h.serve(r.Context(), conn, claims.Subject)
	}
}

func (h *Hub) serve(ctx context.Context, conn *websocket.Conn, userID string) {
	conn.SetReadLimit(readLimit)

	c := &client{userID: userID, conn: conn, send: make(chan []byte, sendBuffer), rooms: map[string]bool{}}

	if !h.add(c) {
		conn.Close(CloseTooMany, "demasiadas conexiones abiertas")
		return
	}
	defer h.remove(c)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go h.writeLoop(ctx, c, cancel)
	h.readLoop(ctx, c)
}

func (h *Hub) add(c *client) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.byUser[c.userID]) >= maxConnectionsPerUser {
		return false
	}
	h.clients[c] = true
	if h.byUser[c.userID] == nil {
		h.byUser[c.userID] = map[*client]bool{}
	}
	h.byUser[c.userID][c] = true
	return true
}

func (h *Hub) remove(c *client) {
	h.mu.Lock()
	if h.clients[c] {
		delete(h.clients, c)
		delete(h.byUser[c.userID], c)
		if len(h.byUser[c.userID]) == 0 {
			delete(h.byUser, c.userID)
		}
		close(c.send)
	}
	h.mu.Unlock()
	c.conn.CloseNow()
}

func (h *Hub) readLoop(ctx context.Context, c *client) {
	for {
		_, raw, err := c.conn.Read(ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) && websocket.CloseStatus(err) == -1 {
				h.log.Debug("lectura del socket cortada", "usuario", c.userID, "error", err)
			}
			return
		}

		var msg clientMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue // un mensaje mal formado se ignora, no tira la conexión
		}

		switch msg.Type {
		case "subscribe":
			if msg.RaceID != "" {
				c.mu.Lock()
				c.rooms[msg.RaceID] = true
				c.mu.Unlock()
			}
		case "unsubscribe":
			c.mu.Lock()
			delete(c.rooms, msg.RaceID)
			c.mu.Unlock()
		case "ping":
			h.enqueue(c, NewPong())
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
		h.log.Debug("cliente lento, se lo desconecta", "usuario", c.userID)
		go h.remove(c)
	}
}

// ── Difusión ──────────────────────────────────────────────────────────────

// ToRoom manda un evento a todos los suscriptos a una carrera.
func (h *Hub) ToRoom(raceID string, event any) {
	h.mu.RLock()
	targets := make([]*client, 0, len(h.clients))
	for c := range h.clients {
		if c.inRoom(raceID) {
			targets = append(targets, c)
		}
	}
	h.mu.RUnlock()

	for _, c := range targets {
		h.enqueue(c, event)
	}
}

// ToUser manda un evento a todas las conexiones de un usuario. El saldo se
// actualiza en todas sus pestañas, no solo en la que apostó.
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

// ToRoomPerUser manda a cada suscripto un evento armado para él. Lo usa
// `race.finished`, cuyo campo `payouts` trae solo las apuestas de quien lo
// recibe: difundir el mismo objeto a todos filtraría lo que cobró cada uno.
func (h *Hub) ToRoomPerUser(raceID string, build func(userID string) any) {
	h.mu.RLock()
	targets := make([]*client, 0, len(h.clients))
	for c := range h.clients {
		if c.inRoom(raceID) {
			targets = append(targets, c)
		}
	}
	h.mu.RUnlock()

	for _, c := range targets {
		h.enqueue(c, build(c.userID))
	}
}

// ToAll difunde a todos los conectados. Lo usa `leaderboard.updated`, que es
// público.
func (h *Hub) ToAll(event any) {
	h.mu.RLock()
	targets := make([]*client, 0, len(h.clients))
	for c := range h.clients {
		targets = append(targets, c)
	}
	h.mu.RUnlock()

	for _, c := range targets {
		h.enqueue(c, event)
	}
}

// Connections es cuántas conexiones hay abiertas. Lo expone /health.
func (h *Hub) Connections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
