// Hipódromo — monolito Go que sirve la app ancla del módulo Angular.
//
// Un solo binario, una sola dependencia externa (github.com/coder/websocket),
// sin base de datos y con el dataset embebido. `go build` produce un ejecutable
// que corre solo.
//
//	go run .                    arranca en :8080 con el emisor de correo de log
//	RESET=1 go run .            arranca desde el dataset limpio
//	go test ./...               tests, incluidos los golden contra el contrato
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/talentodh/hipodromo/internal/api"
	"github.com/talentodh/hipodromo/internal/auth"
	"github.com/talentodh/hipodromo/internal/mail"
	"github.com/talentodh/hipodromo/internal/program"
	"github.com/talentodh/hipodromo/internal/seed"
	"github.com/talentodh/hipodromo/internal/store"
	"github.com/talentodh/hipodromo/internal/ws"
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

	// El dataset se reubica en el tiempo tomando este instante como referencia
	// (docs/contract/README.md, regla de rebase).
	data, err := seed.Load(time.Now())
	if err != nil {
		return err
	}

	st, err := store.New(data, store.Options{SnapshotPath: cfg.SnapshotPath, Reset: cfg.Reset})
	if err != nil {
		return err
	}
	if err := st.PersistError(); err != nil {
		log.Warn("no se puede escribir la copia en disco, el estado vive solo en memoria",
			"ruta", cfg.SnapshotPath, "error", err)
	}

	signer := auth.NewSigner(cfg.JWTSecret)
	hub := ws.NewHub(log, signer)

	var sender mail.Sender
	if cfg.ResendAPIKey != "" {
		sender = mail.NewResendSender(cfg.ResendAPIKey, cfg.MailFrom, cfg.FrontURL, log)
		log.Info("correo por Resend", "desde", cfg.MailFrom)
	} else {
		// No es un stub incompleto: imprime el enlace de verificación completo,
		// así se puede verificar una cuenta en clase sin configurar nada.
		sender = &mail.LogSender{Log: log, FrontURL: cfg.FrontURL}
		log.Info("correo en modo desarrollo: los enlaces de verificación salen por el log")
	}

	server := &api.Server{
		Store: st, Signer: signer, Hub: hub, Mail: sender, Log: log,
		FrontURL: cfg.FrontURL, AllowedOrigins: cfg.AllowedOrigins,
		Clock: time.Now,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// El calendario corre en su propia goroutine: larga carreras solo, sin
	// esperar a que nadie pida nada.
	go program.New(st, hub, log, data.Program).Run(ctx)

	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: server.Handler(),
		// Sin WriteTimeout: el WebSocket es una conexión larga y cualquier
		// tope la cortaría a mitad de carrera.
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Info("escuchando",
			"puerto", cfg.Port, "api", api.BasePath, "front", cfg.FrontURL,
			"carreras", len(data.Program), "ciclo", seed.CycleLength)
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

// ── Configuración ─────────────────────────────────────────────────────────

const defaultJWTSecret = "hipodromo-desarrollo-no-usar-en-produccion"

type config struct {
	Port           string
	JWTSecret      string
	FrontURL       string
	AllowedOrigins []string
	ResendAPIKey   string
	MailFrom       string
	SnapshotPath   string
	Reset          bool
	LogLevel       slog.Level
}

func loadConfig() config {
	cfg := config{
		Port:         env("PORT", "8080"),
		JWTSecret:    env("JWT_SECRET", defaultJWTSecret),
		FrontURL:     env("FRONT_URL", "http://localhost:4200"),
		ResendAPIKey: env("RESEND_API_KEY", ""),
		MailFrom:     env("MAIL_FROM", "Hipódromo <no-reply@hipodromo.test>"),
		SnapshotPath: env("SNAPSHOT_PATH", "data/snapshot.json"),
		Reset:        env("RESET", "") != "",
		LogLevel:     slog.LevelInfo,
	}

	if raw := env("ALLOWED_ORIGINS", ""); raw != "" {
		for _, origin := range strings.Split(raw, ",") {
			if trimmed := strings.TrimSpace(origin); trimmed != "" {
				cfg.AllowedOrigins = append(cfg.AllowedOrigins, trimmed)
			}
		}
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

func newLogger(level slog.Level) *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
