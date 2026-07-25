// Package api es la capa HTTP: rutas, middleware y handlers.
//
// Router de la biblioteca estándar. Desde Go 1.22 `http.ServeMux` entiende
// método y comodines —`GET /api/v1/races/{id}`—, así que un router externo
// solo agregaría una dependencia y una sintaxis más que explicar.
package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/talentodh/hipodromo/internal/auth"
	"github.com/talentodh/hipodromo/internal/mail"
	"github.com/talentodh/hipodromo/internal/store"
	"github.com/talentodh/hipodromo/internal/ws"
)

// BasePath es el prefijo de toda la API, del contrato.
const BasePath = "/api/v1"

type Server struct {
	Store          *store.Store
	Signer         *auth.Signer
	Hub            *ws.Hub
	Mail           mail.Sender
	Log            *slog.Logger
	FrontURL       string
	AllowedOrigins []string
	Clock          func() time.Time

	loginLimiter  *rateLimiter
	resendLimiter *rateLimiter
}

func (s *Server) now() time.Time {
	if s.Clock != nil {
		return s.Clock()
	}
	return time.Now()
}

// Handler arma el router con toda la cadena de middleware.
func (s *Server) Handler() http.Handler {
	// Los límites protegen las dos puertas que se pueden golpear desde afuera
	// sin sesión: adivinar contraseñas y usar el reenvío de correo para spamear.
	s.loginLimiter = newRateLimiter(10, time.Minute, s.now)
	s.resendLimiter = newRateLimiter(3, time.Hour, s.now)

	mux := http.NewServeMux()

	// Auth
	mux.HandleFunc("POST "+BasePath+"/auth/register", s.handleRegister)
	mux.HandleFunc("POST "+BasePath+"/auth/verify", s.handleVerify)
	mux.HandleFunc("POST "+BasePath+"/auth/resend-verification", s.handleResendVerification)
	mux.HandleFunc("POST "+BasePath+"/auth/login", s.handleLogin)
	mux.HandleFunc("POST "+BasePath+"/auth/refresh", s.handleRefresh)
	mux.HandleFunc("POST "+BasePath+"/auth/logout", s.handleLogout)
	mux.HandleFunc("GET "+BasePath+"/me", s.handleMe)

	// Carreras
	mux.HandleFunc("GET "+BasePath+"/races", s.handleRaces)
	mux.HandleFunc("GET "+BasePath+"/races/{id}", s.handleRace)
	mux.HandleFunc("GET "+BasePath+"/races/{id}/results", s.handleRaceResults)

	// Apuestas
	mux.HandleFunc("POST "+BasePath+"/bets", s.handleCreateBet)
	mux.HandleFunc("GET "+BasePath+"/bets/me", s.handleMyBets)
	mux.HandleFunc("GET "+BasePath+"/leaderboard", s.handleLeaderboard)

	// Operación
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.Handle("GET /ws", s.Hub.Handler())

	// De afuera hacia adentro: recuperación → log → CORS → autenticación.
	// La recuperación va primero para que atrape también los pánicos del resto.
	return s.withRecovery(s.withLogging(s.withCORS(s.withAuth(mux))))
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":      "ok",
		"time":        s.now().UTC().Format(time.RFC3339),
		"connections": s.Hub.Connections(),
	})
}
