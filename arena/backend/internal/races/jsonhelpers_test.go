package races

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

// Los eventos se verifican SOBRE EL JSON y no sobre los campos de la estructura.
//
// Es a propósito: lo que puede filtrar información es lo que sale por el cable, y
// un campo puede aparecer en el JSON por un tipo embebido, por una etiqueta mal
// puesta o por un `any` que alguien metió en el medio. Mirar los bytes es la única
// forma de estar seguros.

func mustJSON(t *testing.T, value any) string {
	t.Helper()

	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("serializando %T: %v", value, err)
	}
	return string(encoded)
}

// containsHorse busca el id de un caballo en el JSON.
func containsHorse(serialized, horseID string) bool {
	return strings.Contains(serialized, horseID)
}

// containsText busca un texto —el nombre de un campo, normalmente— en el JSON.
func containsText(serialized, text string) bool {
	return strings.Contains(serialized, text)
}

// show imprime un *int64 legible. Sin esto, un fallo de test muestra la dirección
// del puntero y hay que volver a correrlo para entender qué pasó.
func show(n *int64) string {
	if n == nil {
		return "nil"
	}
	return strconv.FormatInt(*n, 10)
}
