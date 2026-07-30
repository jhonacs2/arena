package api

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/talentodh/arena/internal/accounts"
	"github.com/talentodh/arena/internal/auth"
	"github.com/talentodh/arena/internal/contract"
)

type ctxKey int

const userKey ctxKey = iota

// withAuth lee el Bearer, valida la firma y **carga el usuario de la base**.
//
// Cargarlo de la base y no del token es la decisión que importa: el rol vive en
// `users.role` y se lee en cada petición, así que bajar a alguien de admin surte
// efecto en la petición siguiente y no cuando venza su token. El JWT solo dice
// quién es; qué puede hacer lo dice la base (decisiones.md §4).
//
// NO rechaza la petición: dejar pasar al anónimo es lo que permite que `/health`
// y las rutas públicas convivan en el mismo mux. Exigir sesión es tarea de cada
// handler, con requireUser.
func (s *Server) withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !found || token == "" {
			next.ServeHTTP(w, r)
			return
		}

		user, err := s.userFromToken(r.Context(), token)
		if err != nil {
			next.ServeHTTP(w, r) // token inválido o vencido = anónimo
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, user)))
	})
}

// userFromToken valida el token y devuelve el usuario que dice ser.
func (s *Server) userFromToken(ctx context.Context, token string) (accounts.User, error) {
	claims, err := s.Signer.Parse(token, s.now())
	if err != nil {
		return accounts.User{}, contract.Errorf(contract.CodeUnauthenticated)
	}

	user, err := s.Accounts.ByID(ctx, claims.Subject)
	if err != nil {
		// Token bien firmado de un usuario que ya no está. Pasa si se borró la
		// cuenta; no es un error del servidor.
		if errors.Is(err, accounts.ErrNotFound) {
			return accounts.User{}, contract.Errorf(contract.CodeUnauthenticated)
		}
		return accounts.User{}, err
	}
	return user, nil
}

// currentUser devuelve el usuario de la sesión, o false si la petición es
// anónima.
func currentUser(r *http.Request) (accounts.User, bool) {
	user, ok := r.Context().Value(userKey).(accounts.User)
	return user, ok
}

// requireUser devuelve el usuario o un error UNAUTHENTICATED.
func requireUser(r *http.Request) (accounts.User, error) {
	if user, ok := currentUser(r); ok {
		return user, nil
	}
	return accounts.User{}, contract.Errorf(contract.CodeUnauthenticated)
}

// requireAdmin exige rol admin. El chequeo es contra el rol que salió de la
// BASE, no del token.
func requireAdmin(r *http.Request) (accounts.User, error) {
	user, err := requireUser(r)
	if err != nil {
		return accounts.User{}, err
	}
	if !user.IsAdmin() {
		return accounts.User{}, contract.Errorf(contract.CodeForbidden)
	}
	return user, nil
}

// withAdminGate rechaza todo lo que empiece con /api/admin/ si el rol no es
// admin, ANTES de llegar al handler.
//
// Los handlers igual llaman a requireAdmin, y la redundancia es a propósito: el
// riesgo real no es que este chequeo esté mal escrito, es que alguien agregue
// mañana un endpoint de admin y se olvide de poner el chequeo. Un portón por
// prefijo no se puede olvidar — cubre también las rutas de admin del paquete de
// carreras, que se registran en este mismo mux.
//
// Y las monedas son nota (arena/CLAUDE.md §4): acá el que se cuela no gana una
// partida, se pone una calificación.
func (s *Server) withAdminGate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, adminPrefix) {
			next.ServeHTTP(w, r)
			return
		}
		if _, err := requireAdmin(r); err != nil {
			writeError(w, r, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ── Identidad para los otros paquetes ─────────────────────────────────────

// Identity resuelve quién hace la petición, para los handlers de otros paquetes.
//
// Devuelve UNAUTHENTICATED si no hay sesión. **No chequea rol**: expone el rol y
// deja el gate a quien la llama —además del portón por prefijo de acá arriba.
//
// `auth.Identity` tiene la misma forma que `races.Identity` y `ws.Identity`, así
// que del otro lado alcanza una conversión de struct.
func (s *Server) Identity(r *http.Request) (auth.Identity, error) {
	user, err := requireUser(r)
	if err != nil {
		return auth.Identity{}, err
	}
	return auth.Identity{UserID: user.ID, Username: user.Username, Role: user.Role}, nil
}

// IdentityFromToken resuelve la identidad desde un token suelto.
//
// Es lo que necesita el WebSocket: ahí el token no viene en una cabecera
// —`WebSocket` del navegador no deja poner cabeceras— sino en el query o en el
// primer mensaje.
func (s *Server) IdentityFromToken(ctx context.Context, token string) (auth.Identity, error) {
	user, err := s.userFromToken(ctx, token)
	if err != nil {
		return auth.Identity{}, err
	}
	return auth.Identity{UserID: user.ID, Username: user.Username, Role: user.Role}, nil
}

// ── CORS ──────────────────────────────────────────────────────────────────

// withCORS habilita el navegador del frontend.
//
// **Con credenciales**, porque el refresh viaja en una cookie HttpOnly: eso
// obliga a reflejar el origen concreto y prohíbe `*`. Un `Access-Control-Allow-
// Origin: *` junto con `Allow-Credentials: true` es una combinación que el
// navegador rechaza directamente.
func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Con lista configurada, solo los de la lista. Sin lista es modo
		// desarrollo y se refleja el origen que vino, que sigue sin ser `*`.
		if origin != "" && (len(s.AllowedOrigins) == 0 || slices.Contains(s.AllowedOrigins, origin)) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Add("Vary", "Origin")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ── Log y recuperación ────────────────────────────────────────────────────

// statusRecorder captura el estado para el log de acceso.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Unwrap deja pasar el http.Hijacker que necesita el WebSocket. Sin esto, un
// envoltorio del ResponseWriter rompe la actualización a socket.
func (r *statusRecorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// El socket no se loguea: los ticks de carrera van a 10 Hz y ahogarían
		// todo lo demás.
		if strings.HasSuffix(r.URL.Path, "/ws") {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		level := slog.LevelInfo
		if rec.status >= 500 {
			level = slog.LevelError
		}

		usuario := ""
		if user, ok := currentUser(r); ok {
			usuario = user.Username
		}
		s.Log.Log(r.Context(), level, "petición",
			"método", r.Method, "ruta", r.URL.Path, "estado", rec.status,
			"ms", time.Since(start).Milliseconds(), "usuario", usuario)
	})
}

// withRecovery evita que un pánico en un handler tire el servidor entero —
// incluida la carrera en vivo, que corre en otra goroutine.
func (s *Server) withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				s.Log.Error("pánico en un handler",
					"pánico", recovered, "método", r.Method, "ruta", r.URL.Path)
				writeJSON(w, http.StatusInternalServerError, contract.Errorf(contract.CodeInternal).Body())
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ── Límite de intentos ────────────────────────────────────────────────────

// rateLimiter es una ventana fija en memoria. Alcanza para lo que protege:
// fuerza bruta contra el login y contra los códigos de invitación.
type rateLimiter struct {
	mu      sync.Mutex
	hits    map[string][]time.Time
	limit   int
	window  time.Duration
	nowFunc func() time.Time
}

func newRateLimiter(limit int, window time.Duration, now func() time.Time) *rateLimiter {
	return &rateLimiter{hits: map[string][]time.Time{}, limit: limit, window: window, nowFunc: now}
}

// allow registra un intento y dice si está permitido. Devuelve además cuántos
// segundos hay que esperar, que es lo que va en `details`.
func (l *rateLimiter) allow(key string) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.nowFunc()
	cutoff := now.Add(-l.window)

	kept := l.hits[key][:0]
	for _, at := range l.hits[key] {
		if at.After(cutoff) {
			kept = append(kept, at)
		}
	}
	l.hits[key] = kept

	if len(kept) >= l.limit {
		retry := int(kept[0].Add(l.window).Sub(now).Seconds()) + 1
		return false, retry
	}
	l.hits[key] = append(l.hits[key], now)
	return true, 0
}

// check aplica el límite y devuelve el error del contrato si se pasó.
func (l *rateLimiter) check(r *http.Request) error {
	if allowed, retry := l.allow(clientIP(r)); !allowed {
		return contract.ErrorWith(contract.CodeRateLimited,
			map[string]any{"retryAfterSeconds": retry})
	}
	return nil
}

// clientIP obtiene la IP mirando primero X-Forwarded-For, porque en producción
// hay un túnel de Cloudflare adelante y RemoteAddr sería siempre el mismo.
func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		if first, _, found := strings.Cut(forwarded, ","); found {
			return strings.TrimSpace(first)
		}
		return strings.TrimSpace(forwarded)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
