package api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/talentodh/hipodromo/internal/contract"
)

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
// Un error que no sea *contract.Fault es un bug nuestro: se loguea completo y
// al cliente le llega un INTERNAL genérico. Nunca se filtra un mensaje interno
// en la respuesta.
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
// DisallowUnknownFields es deliberado: si el frontend manda `amount` donde va
// `monto`, es mejor un 422 explícito que un cero silencioso. En clase, ese
// error se lee y se arregla en treinta segundos.
func decodeJSON(w http.ResponseWriter, r *http.Request, into any) error {
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

// pagination lee `page` y `size` con los valores por defecto del contrato.
// Un valor inválido cae al default en vez de fallar: paginar mal no debería
// romper una pantalla.
func pagination(r *http.Request) (page, size int) {
	page, size = 1, 20

	if raw := r.URL.Query().Get("page"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 1 {
			page = n
		}
	}
	if raw := r.URL.Query().Get("size"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n >= 1 {
			size = min(n, 50)
		}
	}
	return page, size
}
