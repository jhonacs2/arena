package api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/talentodh/arena/internal/contract"
)

// maxBodyBytes: ningún cuerpo legítimo de esta API pasa de unos cientos de
// bytes. El tope está para que nadie pueda hacer que el servidor lea 2 GB en
// memoria mandando un JSON gigante.
const maxBodyBytes = 64 << 10

// writeJSON escribe una respuesta con estado y cuerpo.
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		// La cabecera ya salió: no se puede cambiar el estado. Solo queda dejar rastro.
		slog.Error("no se pudo escribir la respuesta", "error", err)
	}
}

// writeError traduce cualquier error al sobre uniforme del contrato.
//
// Un error que no sea *contract.Fault es un bug nuestro: se loguea completo y al
// cliente le llega un INTERNAL genérico. **Nunca se filtra un mensaje interno en
// la respuesta** — un error de Postgres con el texto de la consulta adentro le
// cuenta a cualquiera cómo está armada la base.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	var fault *contract.Fault
	if !errors.As(err, &fault) {
		slog.Error("error no controlado",
			"error", err, "método", r.Method, "ruta", r.URL.Path)
		fault = contract.Errorf(contract.CodeInternal)
	}
	writeJSON(w, fault.Status(), fault.Body())
}

// decodeJSON lee el cuerpo con un tope de tamaño y rechazando campos de más.
//
// DisallowUnknownFields es deliberado: si el frontend manda `coins` donde va
// `count`, es mejor un 400 explícito que un cero silencioso que regala 0 monedas
// y nadie entiende por qué.
func decodeJSON(r *http.Request, into any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(io.LimitReader(r.Body, maxBodyBytes))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(into); err != nil {
		return contract.FieldErrors(map[string]string{
			"_": "El cuerpo de la petición no es un JSON válido.",
		})
	}
	return nil
}

// intQuery lee un entero del query string. Un valor inválido cae al default en
// vez de fallar: paginar mal no debería romper una pantalla.
func intQuery(r *http.Request, key string, fallback int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return n
}
