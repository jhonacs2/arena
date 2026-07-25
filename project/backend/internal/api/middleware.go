package api

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/talentodh/hipodromo/internal/contract"
)

type ctxKey int

const userIDKey ctxKey = iota

// userID devuelve el usuario de la sesión, o vacío si la petición es anónima.
func userID(r *http.Request) string {
	id, _ := r.Context().Value(userIDKey).(string)
	return id
}

// requireUser devuelve el usuario o un error UNAUTHENTICATED.
func requireUser(r *http.Request) (string, error) {
	if id := userID(r); id != "" {
		return id, nil
	}
	return "", contract.Errorf(contract.CodeUnauthenticated)
}

// withAuth lee el Bearer y, si es válido, deja el usuario en el contexto.
//
// NO rechaza la petición: hay rutas que funcionan con y sin sesión
// —`/races/:id/results` devuelve el podio a cualquiera y los pagos solo al
// dueño—. Exigir sesión es tarea de cada handler, con requireUser.
func (s *Server) withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		token, found := strings.CutPrefix(header, "Bearer ")
		if !found || token == "" {
			next.ServeHTTP(w, r)
			return
		}

		claims, err := s.Signer.Parse(token, s.now())
		if err != nil {
			next.ServeHTTP(w, r) // token inválido = anónimo
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userIDKey, claims.Subject)))
	})
}

// withCORS habilita el navegador del frontend.
func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// `*` solo cuando no hay lista configurada — es el modo desarrollo.
		// Con lista, se refleja el origen concreto: `*` es incompatible con
		// credenciales y esconde errores de configuración.
		switch {
		case len(s.AllowedOrigins) == 0:
			w.Header().Set("Access-Control-Allow-Origin", "*")
		case origin != "" && slices.Contains(s.AllowedOrigins, origin):
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// statusRecorder captura el estado para el log de acceso.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Los ticks del socket no se loguean: 10 por segundo ahogarían todo lo demás.
		if strings.HasPrefix(r.URL.Path, "/ws") {
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
		s.Log.Log(r.Context(), level, "petición",
			"método", r.Method, "ruta", r.URL.Path, "estado", rec.status,
			"ms", time.Since(start).Milliseconds(), "usuario", userID(r))
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
// fuerza bruta contra el login y abuso del reenvío de correo.
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
// segundos hay que esperar, que es lo que el contrato pone en `details`.
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

// clientIP obtiene la IP mirando primero X-Forwarded-For, porque en producción
// hay un proxy adelante.
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
