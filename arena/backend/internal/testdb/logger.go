package testdb

import (
	"io"
	"log/slog"
)

// quietLogger descarta la salida: en los tests, «esquema al día» impreso una vez
// por test tapa el resultado que sí importa.
func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
}
