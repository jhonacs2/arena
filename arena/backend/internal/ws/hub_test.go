package ws

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

// El hub se prueba con un servidor y conexiones de verdad. Un doble no serviría:
// lo que puede salir mal acá es el reparto entre salas y el aislamiento por
// destinatario, y las dos cosas dependen de las conexiones reales.

const dialWait = 3 * time.Second

func quietHub() *Hub {
	return NewHub(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

// tokenIsUser trata el token como el nombre del usuario. Alcanza para probar el
// hub, que no sabe de JWT. El token "vencido" se rechaza: es el caso que prueba
// el cierre 4001.
func tokenIsUser(_ context.Context, token string) (Identity, error) {
	if token == "" || token == "vencido" {
		return Identity{}, errors.New("token inválido o vencido")
	}
	return Identity{UserID: token, Username: token, Role: "student"}, nil
}

// serve levanta un servidor con el hub y devuelve la URL de ws://.
func serve(t *testing.T, hub *Hub, onJoin JoinHook) string {
	t.Helper()

	server := httptest.NewServer(hub.Handler(tokenIsUser, onJoin))
	t.Cleanup(server.Close)
	return "ws" + strings.TrimPrefix(server.URL, "http")
}

// dial se conecta a una sala con un token.
func dial(t *testing.T, base, raceID, token string) *websocket.Conn {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), dialWait)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, base+"?raceId="+raceID+"&token="+token, nil)
	if err != nil {
		t.Fatalf("conectando a la sala %s como %s: %v", raceID, token, err)
	}
	t.Cleanup(func() { conn.CloseNow() })
	return conn
}

// readEvent lee un evento y devuelve el objeto ya parseado.
func readEvent(t *testing.T, conn *websocket.Conn) map[string]any {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), dialWait)
	defer cancel()

	_, raw, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("leyendo un evento: %v", err)
	}
	var event map[string]any
	if err := json.Unmarshal(raw, &event); err != nil {
		t.Fatalf("el evento no es JSON: %v", err)
	}
	return event
}

// waitFor espera hasta que se cumpla la condición: el alta de un cliente la hace
// otra goroutine y no hay forma de saber cuándo terminó.
func waitFor(t *testing.T, condition func() bool, what string) {
	t.Helper()

	deadline := time.Now().Add(dialWait)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("se agotó la espera: %s", what)
}

// TestSinRaceIdNoHayConexion: el hub multiplexa por sala, así que sin sala no hay
// nada que mandar.
func TestSinRaceIdNoHayConexion(t *testing.T) {
	server := httptest.NewServer(quietHub().Handler(tokenIsUser, nil))
	defer server.Close()

	response, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("pidiendo sin raceId: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("el estado es %d y se esperaba 400", response.StatusCode)
	}
}

// TestTokenInvalidoCierraCon4001: el 4001 le dice al cliente que refresque el
// token y reconecte UNA vez, en lugar de entrar en backoff contra un servidor que
// está bien.
func TestTokenInvalidoCierraCon4001(t *testing.T) {
	base := serve(t, quietHub(), nil)
	conn := dial(t, base, "race-1", "vencido")

	ctx, cancel := context.WithTimeout(context.Background(), dialWait)
	defer cancel()

	_, _, err := conn.Read(ctx)
	if got := websocket.CloseStatus(err); got != websocket.StatusCode(4001) {
		t.Errorf("el código de cierre es %d y se esperaba 4001", got)
	}
}

// TestSinTokenNiHandshakeSeCierra: un socket abierto sin autenticar no se puede
// quedar abierto. Se espera el handshake con plazo y después se cierra.
func TestSinTokenNiHandshakeSeCierra(t *testing.T) {
	base := serve(t, quietHub(), nil)

	ctx, cancel := context.WithTimeout(context.Background(), handshakeWait+dialWait)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, base+"?raceId=race-1", nil)
	if err != nil {
		// El servidor puede cortar antes de que el handshake HTTP termine.
		// También es un rechazo válido.
		return
	}
	defer conn.CloseNow()

	if _, _, err := conn.Read(ctx); err == nil {
		t.Error("la conexión sin token siguió viva")
	}
}

// TestTokenPorElPrimerMensaje: el navegador no permite mandar headers en el
// handshake de un WebSocket, así que el token entra por query o por el primer
// mensaje. Las dos formas tienen que funcionar.
func TestTokenPorElPrimerMensaje(t *testing.T) {
	hub := quietHub()
	base := serve(t, hub, func(_ context.Context, raceID string, id Identity) (any, any, error) {
		return map[string]string{"type": "room.state", "userId": id.UserID}, nil, nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), dialWait)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, base+"?raceId=race-1", nil)
	if err != nil {
		t.Fatalf("conectando sin token en la query: %v", err)
	}
	defer conn.CloseNow()

	handshake, _ := json.Marshal(tokenMessage{Type: "auth", Token: "ana"})
	if err := conn.Write(ctx, websocket.MessageText, handshake); err != nil {
		t.Fatalf("mandando el handshake: %v", err)
	}

	event := readEvent(t, conn)
	if event["type"] != "room.state" || event["userId"] != "ana" {
		t.Errorf("el primer evento es %+v", event)
	}
}

// TestToRoomSoloLlegaALaSala: dos carreras corriendo al mismo tiempo es lo normal
// en una clase, y los ticks de una no pueden llegar a la otra.
func TestToRoomSoloLlegaALaSala(t *testing.T) {
	hub := quietHub()
	base := serve(t, hub, nil)

	inRace1 := dial(t, base, "race-1", "ana")
	inRace2 := dial(t, base, "race-2", "beto")
	waitFor(t, func() bool { return hub.Connections() == 2 }, "que se conecten los dos")

	hub.ToRoom("race-1", map[string]string{"type": "race.tick", "raceId": "race-1"})

	if event := readEvent(t, inRace1); event["raceId"] != "race-1" {
		t.Errorf("a la sala 1 le llegó %+v", event)
	}

	// El de la otra sala no recibe nada. Se comprueba con un plazo corto: si
	// llega algo, es que se filtró.
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	if _, raw, err := inRace2.Read(ctx); err == nil {
		t.Errorf("a la sala 2 le llegó un evento de la sala 1: %s", raw)
	}
}

// TestToRoomPerUserArmaUnEventoPorDestinatario es el mecanismo del que dependen
// `race.finished` y `race.cancelled`: difundir el mismo objeto filtraría cuánto
// cobró cada uno.
func TestToRoomPerUserArmaUnEventoPorDestinatario(t *testing.T) {
	hub := quietHub()
	base := serve(t, hub, nil)

	anaConn := dial(t, base, "race-1", "ana")
	betoConn := dial(t, base, "race-1", "beto")
	waitFor(t, func() bool { return hub.RoomSize("race-1") == 2 }, "que se llene la sala")

	// Cada uno recibe SU pago.
	payouts := map[string]int64{"ana": 1470, "beto": 0}
	hub.ToRoomPerUser("race-1", func(userID string) any {
		return map[string]any{"type": "race.finished", "myPayout": payouts[userID]}
	})

	anaEvent := readEvent(t, anaConn)
	betoEvent := readEvent(t, betoConn)

	if anaEvent["myPayout"] != float64(1470) {
		t.Errorf("a Ana le llegó myPayout %v", anaEvent["myPayout"])
	}
	if betoEvent["myPayout"] != float64(0) {
		t.Errorf("a Beto le llegó myPayout %v", betoEvent["myPayout"])
	}
}

// TestRoomJoinedNoVuelveAlQueEntro: el que entra recibe `room.state`, y el
// `room.joined` va al resto. Mandárselo a él también sería avisarle de su propia
// llegada.
func TestRoomJoinedNoVuelveAlQueEntro(t *testing.T) {
	hub := quietHub()
	base := serve(t, hub, func(_ context.Context, _ string, id Identity) (any, any, error) {
		return map[string]string{"type": "room.state", "userId": id.UserID},
			map[string]string{"type": "room.joined", "userId": id.UserID},
			nil
	})

	anaConn := dial(t, base, "race-1", "ana")
	if event := readEvent(t, anaConn); event["type"] != "room.state" {
		t.Fatalf("el primer evento de Ana es %+v", event)
	}
	waitFor(t, func() bool { return hub.RoomSize("race-1") == 1 }, "que entre Ana")

	betoConn := dial(t, base, "race-1", "beto")

	// A Ana le llega que entró Beto.
	if event := readEvent(t, anaConn); event["type"] != "room.joined" || event["userId"] != "beto" {
		t.Errorf("a Ana le llegó %+v", event)
	}
	// Y a Beto le llega su room.state, no su propio room.joined.
	if event := readEvent(t, betoConn); event["type"] != "room.state" || event["userId"] != "beto" {
		t.Errorf("a Beto le llegó %+v", event)
	}
}

// TestOnJoinQueFallaCierraLaConexion: si la sala no existe o el alumno no puede
// verla, se cierra con 4000 y no con 4001 — no es un problema de token y
// reconectar no lo va a arreglar.
func TestOnJoinQueFallaCierraLaConexion(t *testing.T) {
	base := serve(t, quietHub(), func(context.Context, string, Identity) (any, any, error) {
		return nil, nil, errors.New("esa carrera no existe")
	})

	conn := dial(t, base, "race-inexistente", "ana")

	ctx, cancel := context.WithTimeout(context.Background(), dialWait)
	defer cancel()

	_, _, err := conn.Read(ctx)
	if got := websocket.CloseStatus(err); got != websocket.StatusCode(4000) {
		t.Errorf("el código de cierre es %d y se esperaba 4000", got)
	}
}

// TestDemasiadasConexionesCierraCon4029: abrir una pestaña por carrera es normal;
// veinte es una fuga.
func TestDemasiadasConexionesCierraCon4029(t *testing.T) {
	hub := quietHub()
	base := serve(t, hub, nil)

	for i := 0; i < maxConnectionsPerUser; i++ {
		dial(t, base, "race-1", "ana")
	}
	waitFor(t, func() bool { return hub.Connections() == maxConnectionsPerUser }, "que se abran todas")

	extra := dial(t, base, "race-1", "ana")

	ctx, cancel := context.WithTimeout(context.Background(), dialWait)
	defer cancel()

	_, _, err := extra.Read(ctx)
	if got := websocket.CloseStatus(err); got != websocket.StatusCode(4029) {
		t.Errorf("el código de cierre es %d y se esperaba 4029", got)
	}

	// Y otro usuario no queda afectado por el límite de Ana.
	other := dial(t, base, "race-1", "beto")
	hub.ToRoom("race-1", map[string]string{"type": "race.tick"})
	if event := readEvent(t, other); event["type"] != "race.tick" {
		t.Errorf("Beto recibió %+v", event)
	}
}

// TestPing responde pong: es el keepalive del cliente, y el único mensaje que
// manda además del handshake. Apostar es un POST, no un mensaje de socket.
func TestPing(t *testing.T) {
	base := serve(t, quietHub(), nil)
	conn := dial(t, base, "race-1", "ana")

	ctx, cancel := context.WithTimeout(context.Background(), dialWait)
	defer cancel()

	if err := conn.Write(ctx, websocket.MessageText, []byte(`{"type":"ping"}`)); err != nil {
		t.Fatalf("mandando ping: %v", err)
	}
	if event := readEvent(t, conn); event["type"] != "pong" {
		t.Errorf("la respuesta al ping es %+v", event)
	}
}

// TestUnMensajeBasuraNoTiraLaConexion: un cliente con un bug no puede sacar a
// nadie de la carrera.
func TestUnMensajeBasuraNoTiraLaConexion(t *testing.T) {
	hub := quietHub()
	base := serve(t, hub, nil)
	conn := dial(t, base, "race-1", "ana")
	waitFor(t, func() bool { return hub.Connections() == 1 }, "que se conecte")

	ctx, cancel := context.WithTimeout(context.Background(), dialWait)
	defer cancel()

	if err := conn.Write(ctx, websocket.MessageText, []byte("esto no es json")); err != nil {
		t.Fatalf("mandando basura: %v", err)
	}

	hub.ToRoom("race-1", map[string]string{"type": "race.tick"})
	if event := readEvent(t, conn); event["type"] != "race.tick" {
		t.Errorf("después de la basura llegó %+v", event)
	}
}

// TestAlDesconectarseSeLibera: sin esto, cada recarga de página dejaría un cliente
// muerto en la sala y con ticks a 10 Hz la memoria crecería toda la clase.
func TestAlDesconectarseSeLibera(t *testing.T) {
	hub := quietHub()
	base := serve(t, hub, nil)

	conn := dial(t, base, "race-1", "ana")
	waitFor(t, func() bool { return hub.Connections() == 1 }, "que se conecte")

	if err := conn.Close(websocket.StatusNormalClosure, "chau"); err != nil {
		t.Fatalf("cerrando: %v", err)
	}
	waitFor(t, func() bool { return hub.Connections() == 0 }, "que se libere la conexión")

	if hub.RoomSize("race-1") != 0 {
		t.Errorf("la sala quedó con %d clientes", hub.RoomSize("race-1"))
	}
}
