package races

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/talentodh/arena/internal/contract"
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
		// La cabecera ya salió: no se puede cambiar el estado. Solo queda dejar
		// rastro.
		slog.Error("no se pudo escribir la respuesta", "error", err)
	}
}

// writeError traduce cualquier error al sobre uniforme del contrato.
//
// Un error que no sea *contract.Fault es un bug nuestro: se loguea completo y al
// cliente le llega un INTERNAL genérico. NUNCA se filtra un mensaje interno en la
// respuesta — un error de Postgres puede traer el nombre de una columna o el
// contenido de una fila.
func (h *Handlers) writeError(w http.ResponseWriter, r *http.Request, err error) {
	var fault *contract.Fault
	if !errors.As(err, &fault) {
		h.log.Error("error no controlado",
			"error", err, "método", r.Method, "ruta", r.URL.Path)
		fault = contract.Errorf(contract.CodeInternal)
	}
	writeJSON(w, fault.Status(), fault.Body())
}

// decodeJSON lee el cuerpo con un tope de tamaño y rechazando campos de más.
//
// DisallowUnknownFields es deliberado: si el frontend manda `monto` donde va
// `amount`, es mejor un 400 explícito que apostar cero monedas en silencio.
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
