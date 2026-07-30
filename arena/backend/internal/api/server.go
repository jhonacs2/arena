// Package api es la capa HTTP: rutas, middleware y handlers de autenticación,
// saldo y administración.
//
// Router de la biblioteca estándar. Desde Go 1.22 `http.ServeMux` entiende
// método y comodines —`POST /api/admin/users/{id}/gift`—, así que un router
// externo solo agregaría una dependencia.
//
// Las rutas de carreras, salas y WebSocket **no se registran acá**: viven en su
// propio paquete y se enganchan con ExtraRoutes, así las dos mitades del backend
// no se pisan el archivo del router.
package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/talentodh/arena/internal/accounts"
	"github.com/talentodh/arena/internal/auth"
	"github.com/talentodh/arena/internal/ledger"
)

// BasePath es el prefijo de toda la API (api.md).
const BasePath = "/api"

// adminPrefix es el prefijo que exige rol admin. Lo usa el middleware, no cada
// handler: ver withAdminGate.
const adminPrefix = BasePath + "/admin/"

type Server struct {
	Accounts *accounts.Store
	Ledger   *ledger.Store
	Signer   *auth.Signer
	Log      *slog.Logger

	// AllowedOrigins son los orígenes del frontend. Vacío = modo desarrollo.
	AllowedOrigins []string

	Cookie CookieOptions

	// Clock permite fijar el tiempo en los tests. En producción es time.Now.
	Clock func() time.Time

	// ExtraRoutes son registradores de rutas de otros paquetes. Es el gancho por
	// el que entran las carreras, las salas y el socket:
	//
	//	server.ExtraRoutes = append(server.ExtraRoutes, raceHandlers.Register)
	//
	// Se registran en el MISMO mux, así que pasan por la misma cadena de
	// middleware: recuperación, log, CORS, identidad y el portón de /admin/.
	ExtraRoutes []func(*http.ServeMux)

	loginLimiter *rateLimiter
	codeLimiter  *rateLimiter
}

func (s *Server) now() time.Time {
	if s.Clock != nil {
		return s.Clock()
	}
	return time.Now()
}

// Handler arma el router con toda la cadena de middleware.
func (s *Server) Handler() http.Handler {
	// Las dos puertas que se pueden golpear desde afuera sin sesión: adivinar
	// contraseñas y adivinar códigos de invitación.
	//
	// Los topes son GENEROSOS porque una clase entera sale por la misma IP: 30
	// alumnos detrás del NAT del aula comparten contador, y un límite de 10 por
	// minuto los dejaría afuera a la mitad justo en el peor momento. Frenan la
	// fuerza bruta de un script, no a un aula.
	s.loginLimiter = newRateLimiter(60, time.Minute, s.now)
	s.codeLimiter = newRateLimiter(120, time.Minute, s.now)

	mux := http.NewServeMux()

	// Público
	mux.HandleFunc("POST "+BasePath+"/auth/check-code", s.handleCheckCode)
	mux.HandleFunc("POST "+BasePath+"/auth/redeem", s.handleRedeem)
	mux.HandleFunc("POST "+BasePath+"/auth/login", s.handleLogin)
	mux.HandleFunc("POST "+BasePath+"/auth/refresh", s.handleRefresh)
	mux.HandleFunc("POST "+BasePath+"/auth/logout", s.handleLogout)

	// Alumno
	mux.HandleFunc("GET "+BasePath+"/me", s.handleMe)
	mux.HandleFunc("GET "+BasePath+"/me/transactions", s.handleMyTransactions)

	// Instructor
	mux.HandleFunc("POST "+BasePath+"/admin/codes", s.handleCreateCodes)
	mux.HandleFunc("GET "+BasePath+"/admin/codes", s.handleListCodes)
	mux.HandleFunc("GET "+BasePath+"/admin/scores", s.handleScores)
	mux.HandleFunc("POST "+BasePath+"/admin/users/{id}/gift", s.handleGift)
	mux.HandleFunc("POST "+BasePath+"/admin/users/{id}/grant-points", s.handleGrantPoints)

	// Carreras, salas y socket, desde su propio paquete.
	for _, register := range s.ExtraRoutes {
		register(mux)
	}

	mux.HandleFunc("GET /health", s.handleHealth)

	// De afuera hacia adentro: recuperación → log → CORS → identidad → portón de
	// admin. La recuperación va primero para que atrape también los pánicos del
	// resto de la cadena.
	return s.withRecovery(s.withLogging(s.withCORS(s.withAuth(s.withAdminGate(mux)))))
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   s.now().UTC().Format(time.RFC3339),
	})
}
