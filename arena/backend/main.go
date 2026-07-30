// Arena — backend de la app en vivo del módulo.
//
// Un solo binario. La base es Postgres (Supabase) y el esquema sale de
// arena/docs/contract/schema.sql, que se aplica en cada arranque.
//
//	DATABASE_URL=… go run .        arranca en :8080 y aplica el esquema
//	go test ./...                  tests; los de base necesitan ARENA_TEST_DATABASE_URL
//
// El backend NO expone puerto en producción: `cloudflared` corre como servicio en
// el VPS y abre la conexión hacia afuera (arena/CLAUDE.md §6).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/talentodh/arena/internal/accounts"
	"github.com/talentodh/arena/internal/api"
	"github.com/talentodh/arena/internal/auth"
	"github.com/talentodh/arena/internal/db"
	"github.com/talentodh/arena/internal/ledger"
	"github.com/talentodh/arena/internal/races"
	"github.com/talentodh/arena/internal/racesdb"
	"github.com/talentodh/arena/internal/ws"
)

func main() {
	if err := run(); err != nil {
		slog.Error("el servidor se detuvo", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := loadConfig()
	log := newLogger(cfg.LogLevel)

	if cfg.JWTSecret == defaultJWTSecret {
		log.Warn("JWT_SECRET no está configurado y se está usando el de desarrollo — no desplegar así")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Open(ctx, cfg.DatabaseURL, db.Options{
		MaxConns:       cfg.DBMaxConns,
		SimpleProtocol: cfg.DBSimpleProtocol,
	})
	if err != nil {
		return err
	}
	defer pool.Close()

	if cfg.Migrate {
		if err := db.Migrate(ctx, pool, log); err != nil {
			return err
		}
	} else {
		log.Warn("MIGRATE=0: no se aplicó el esquema")
	}

	accountStore := accounts.New(pool)
	ledgerStore := ledger.New(pool)

	if err := ensureAdmin(ctx, accountStore, cfg, log); err != nil {
		return err
	}

	// Los refresh vencidos no sirven para nada y ocupan lugar. Se limpian en el
	// arranque y no con un cron: el proceso se reinicia con cada despliegue, que
	// es más seguido que cualquier cron que pudiéramos poner.
	if removed, err := accountStore.PurgeRefreshTokens(ctx, time.Now()); err != nil {
		log.Warn("no se pudieron limpiar los refresh vencidos", "error", err)
	} else if removed > 0 {
		log.Info("refresh vencidos borrados", "cantidad", removed)
	}

	server := &api.Server{
		Accounts:       accountStore,
		Ledger:         ledgerStore,
		Signer:         auth.NewSigner(cfg.JWTSecret),
		Log:            log,
		AllowedOrigins: cfg.AllowedOrigins,
		Cookie:         cfg.Cookie,
		Clock:          time.Now,
	}

	// ── Carreras, salas y WebSocket ──────────────────────────────────────────
	//
	// El paquete de carreras expone un único registrador con sus rutas más
	// GET /api/ws, y queda dentro de la misma cadena de middleware del server
	// —incluido el portón de rol de /api/admin/.
	hub := ws.NewHub(log)

	raceStore := racesdb.New(pool, racesdb.LedgerFunc(
		func(ctx context.Context, tx pgx.Tx, m racesdb.Movement) (int64, error) {
			return ledgerStore.MoveTx(ctx, tx, ledger.Movement(m))
		}))

	raceService, raceHandlers := races.NewModule(races.Deps{
		Store: raceStore,
		Hub:   hub,
		Log:   log,

		// `races.Identity` y `auth.Identity` tienen la misma forma, así que alcanza
		// la conversión: ninguno de los dos paquetes tiene que importar al otro.
		Identity: func(r *http.Request) (races.Identity, error) {
			id, err := server.Identity(r)
			return races.Identity(id), err
		},
		IdentityFromToken: func(ctx context.Context, token string) (races.Identity, error) {
			id, err := server.IdentityFromToken(ctx, token)
			return races.Identity(id), err
		},

		// La economía. Es pari-mutuel por decisión del usuario, citada literal en
		// `docs/contract/decisiones.md` §0 — el pool se reparte entre quienes
		// aciertan y el piso de 10 puntos vive en la vista `user_scores`.
		//
		// Si esto quedara nulo, las carreras correrían y NO se liquidarían: es la
		// red que dejó el diseño de SettlementRule para que nunca se pague con una
		// regla que nadie eligió.
		Rule: races.PariMutuel{},
	})
	server.ExtraRoutes = append(server.ExtraRoutes, raceHandlers.Register)

	// Las carreras que quedaron en `running` porque el proceso se cayó en el medio.
	// Se retoman ANTES de escuchar: si alguien reconecta a una sala, la simulación
	// ya está corriendo y recibe ticks en vez de una carrera congelada.
	if err := raceService.Resume(ctx); err != nil {
		log.Error("no se pudieron retomar las carreras en curso", "error", err)
	}
	defer raceService.Runner().Close()

	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: server.Handler(),
		// Sin WriteTimeout: el WebSocket de la sala es una conexión larga y
		// cualquier tope la cortaría a mitad de carrera.
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Info("escuchando", "puerto", cfg.Port, "api", api.BasePath)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("el servidor HTTP falló", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Info("apagando")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return httpServer.Shutdown(shutdownCtx)
}

// ensureAdmin crea al instructor si hay credenciales configuradas.
//
// Hace falta porque `invite_codes.created_by` referencia a un usuario y no hay
// registro abierto: sin un admin en la base, nadie puede generar el primer código
// y la app queda cerrada para todos, incluido el instructor.
func ensureAdmin(ctx context.Context, store *accounts.Store, cfg config, log *slog.Logger) error {
	if cfg.AdminUsername == "" || cfg.AdminPassword == "" {
		log.Warn("ADMIN_USERNAME/ADMIN_PASSWORD sin configurar: no se creó el instructor y no se van a poder generar códigos")
		return nil
	}
	if len(cfg.AdminPassword) < auth.MinPasswordLength {
		return errors.New("ADMIN_PASSWORD necesita al menos 8 caracteres")
	}

	hash, err := auth.HashPassword(cfg.AdminPassword)
	if err != nil {
		return err
	}
	admin, err := store.EnsureAdmin(ctx, cfg.AdminUsername, cfg.AdminFirstName, cfg.AdminLastName, hash)
	if err != nil {
		return err
	}
	log.Info("instructor listo", "usuario", admin.Username, "id", admin.ID)
	return nil
}

// ── Configuración ─────────────────────────────────────────────────────────

const defaultJWTSecret = "arena-desarrollo-no-usar-en-produccion"

type config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string

	DBMaxConns       int32
	DBSimpleProtocol bool
	Migrate          bool

	AllowedOrigins []string
	Cookie         api.CookieOptions

	AdminUsername  string
	AdminPassword  string
	AdminFirstName string
	AdminLastName  string

	LogLevel slog.Level
}

func loadConfig() config {
	cfg := config{
		Port:        env("PORT", "8080"),
		DatabaseURL: env("DATABASE_URL", ""),
		JWTSecret:   env("JWT_SECRET", defaultJWTSecret),

		DBMaxConns: int32(envInt("DB_MAX_CONNS", 10)),
		// El pooler de transacciones de Supabase (puerto 6543) no soporta
		// sentencias preparadas. Con conexión directa o pooler de sesión se deja
		// apagado, que es más rápido.
		DBSimpleProtocol: envBool("DB_SIMPLE_PROTOCOL", false),
		Migrate:          envBool("MIGRATE", true),

		Cookie: api.DefaultCookieOptions(),

		AdminUsername:  env("ADMIN_USERNAME", ""),
		AdminPassword:  env("ADMIN_PASSWORD", ""),
		AdminFirstName: env("ADMIN_FIRST_NAME", "Instructor"),
		AdminLastName:  env("ADMIN_LAST_NAME", "Arena"),

		LogLevel: slog.LevelInfo,
	}

	if raw := env("ALLOWED_ORIGINS", ""); raw != "" {
		for _, origin := range strings.Split(raw, ",") {
			if trimmed := strings.TrimSpace(origin); trimmed != "" {
				cfg.AllowedOrigins = append(cfg.AllowedOrigins, trimmed)
			}
		}
	}

	cfg.Cookie.Domain = env("COOKIE_DOMAIN", "")
	cfg.Cookie.Secure = envBool("COOKIE_SECURE", true)
	switch strings.ToLower(env("COOKIE_SAMESITE", "none")) {
	case "lax":
		cfg.Cookie.SameSite = http.SameSiteLaxMode
	case "strict":
		cfg.Cookie.SameSite = http.SameSiteStrictMode
	default:
		// None es el default porque en desarrollo el frontend está en otro puerto
		// y para el navegador eso ya es otro sitio. En producción, con `/api` en el
		// mismo dominio, conviene poner COOKIE_SAMESITE=lax.
		cfg.Cookie.SameSite = http.SameSiteNoneMode
	}

	if strings.EqualFold(env("LOG_LEVEL", ""), "debug") {
		cfg.LogLevel = slog.LevelDebug
	}
	return cfg
}

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if n, err := strconv.Atoi(env(key, "")); err == nil && n > 0 {
		return n
	}
	return fallback
}

// envBool acepta 1/true/yes/on y su contrario. Un valor que no se entiende cae
// al default: una variable mal escrita no debería cambiar el comportamiento en
// silencio ni impedir el arranque.
func envBool(key string, fallback bool) bool {
	switch strings.ToLower(env(key, "")) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func newLogger(level slog.Level) *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
