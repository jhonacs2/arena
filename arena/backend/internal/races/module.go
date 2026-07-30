package races

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/talentodh/arena/internal/ws"
)

// Deps son las dependencias de todo el módulo de carreras.
//
// Existe para que el cableado sea UNA llamada. Sin esto, main tendría que armar el
// servicio, después los handlers, y acordarse de pasarle el servicio a los
// handlers — tres pasos con un orden que importa y ninguna razón para que
// importe.
type Deps struct {
	// Store es el acceso a datos. En producción es racesdb.New(pool, ledger).
	Store Store

	// Hub es el hub de WebSocket. El mismo que usa el resto de la app.
	Hub *ws.Hub

	Log *slog.Logger

	// Identity resuelve quién hace la petición HTTP. La provee el paquete de
	// autenticación:
	//
	//	Identity: func(r *http.Request) (races.Identity, error) {
	//		id, err := server.Identity(r)
	//		return races.Identity(id), err   // misma forma: alcanza la conversión
	//	},
	Identity func(*http.Request) (Identity, error)

	// IdentityFromToken resuelve la identidad desde un token suelto, para el
	// handshake del WebSocket.
	IdentityFromToken func(ctx context.Context, token string) (Identity, error)

	// Rule es la economía: cuánto cobra una apuesta ganadora. Este repositorio NO
	// trae ninguna implementación porque la decisión está abierta — ver
	// races.SettlementRule. Con Rule nula las carreras corren y se cancelan, pero
	// al terminar no se liquidan y queda registrado el motivo.
	Rule SettlementRule

	// Clock y Seeder existen para los tests. En producción se dejan nulos.
	Clock  func() time.Time
	Seeder func() int64
}

// NewModule arma el servicio y los handlers.
//
// Devuelve los dos porque el cableado necesita los dos: los handlers para
// registrar las rutas, y el servicio para retomar las carreras que quedaron
// corriendo (Resume) y para apagar las simulaciones al cerrar
// (Runner().Close()).
//
//	svc, handlers := races.NewModule(races.Deps{…})
//	server.ExtraRoutes = append(server.ExtraRoutes, handlers.Register)
//	if err := svc.Resume(ctx); err != nil { … }
//	defer svc.Runner().Close()
func NewModule(deps Deps) (*Service, *Handlers) {
	service := New(Config{
		Store:  deps.Store,
		Hub:    deps.Hub,
		Log:    deps.Log,
		Rule:   deps.Rule,
		Clock:  deps.Clock,
		Seeder: deps.Seeder,
	})

	handlers := NewHandlers(HandlersConfig{
		Service:       service,
		Hub:           deps.Hub,
		Log:           deps.Log,
		Identify:      deps.Identity,
		IdentifyToken: deps.IdentityFromToken,
	})

	return service, handlers
}
