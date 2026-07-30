// Package invite genera y normaliza los códigos de invitación.
//
// Formato `AAAA-9999` (decisiones.md §2). El alfabeto NO tiene caracteres
// ambiguos: el código se dicta en voz alta o se copia de un chat, y `AVBD-1O34`
// es una llamada de soporte garantizada.
package invite

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
)

// Letters son las 22 letras que quedan al sacar I, L, O y U.
//
//	I se confunde con 1 y con L
//	L se confunde con I y con 1
//	O se confunde con 0
//	U se confunde con V dicha en voz alta
const Letters = "ABCDEFGHJKMNPQRSTVWXYZ"

// Digits son los 8 dígitos que quedan al sacar 0 (por la O) y 1 (por la I y la L).
const Digits = "23456789"

// Longitudes del formato AAAA-9999.
const (
	letterCount = 4
	digitCount  = 4
	// Length es el largo del código canónico, con el guion.
	Length = letterCount + 1 + digitCount
)

// Generate devuelve un código nuevo, por ejemplo `AVBD-2345`.
//
// El espacio es 22⁴ × 8⁴ = 959 512 576 combinaciones. No se adivina a mano, y
// de todos modos /auth/check-code está limitado por IP.
func Generate() (string, error) {
	var sb strings.Builder
	sb.Grow(Length)

	for range letterCount {
		c, err := pick(Letters)
		if err != nil {
			return "", err
		}
		sb.WriteByte(c)
	}
	sb.WriteByte('-')
	for range digitCount {
		c, err := pick(Digits)
		if err != nil {
			return "", err
		}
		sb.WriteByte(c)
	}
	return sb.String(), nil
}

// pick elige un carácter del alfabeto sin sesgo.
//
// El rechazo importa: `random % 22` sobre un byte le da a las primeras seis
// letras un 4,7 % más de probabilidad que al resto. Acá no cambia nada
// práctico, pero es la clase de atajo que después se copia a un sorteo que sí
// importa.
func pick(alphabet string) (byte, error) {
	n := len(alphabet)

	// El límite se calcula en int y NO en byte. Con los 8 dígitos, `256 % 8` es 0
	// y `256 - 0` no entra en un byte: se desborda a 0, todos los valores quedan
	// rechazados y la generación falla siempre. Costó un test entender por qué los
	// códigos no salían.
	limit := 256 - (256 % n)

	var buf [1]byte
	for range 64 {
		if _, err := rand.Read(buf[:]); err != nil {
			return 0, fmt.Errorf("generando el código: %w", err)
		}
		if int(buf[0]) < limit {
			return alphabet[int(buf[0])%n], nil
		}
	}
	// 64 rechazos seguidos tienen probabilidad ~1e-40: si pasa, algo anda mal
	// en la fuente de aleatoriedad y es mejor fallar que devolver un sesgo.
	return 0, errors.New("no se pudo generar el código sin sesgo")
}

// ErrMalformed es un código que no tiene la forma AAAA-9999. El handler lo
// traduce a VALIDATION_FAILED: es distinto de «no existe».
var ErrMalformed = errors.New("el código no tiene el formato AAAA-9999")

// Normalize lleva lo que escribió el alumno a la forma canónica.
//
// Acepta minúsculas, espacios y el código sin guion: `avbd 1234` y `AVBD1234`
// son el mismo código que `AVBD-1234`. Lo que NO hace es corregir caracteres
// ambiguos —un `AVBD-1O34` no se convierte en nada— porque adivinar qué quiso
// escribir alguien es peor que decirle que ese código no existe.
//
// Un código con formato válido pero con letras que el generador nunca usa
// (`AVBI-1234`) pasa por acá y muere en la base con CODE_NOT_FOUND, que es
// exactamente lo que hay que mostrarle.
func Normalize(raw string) (string, error) {
	var sb strings.Builder
	sb.Grow(Length)
	for _, r := range strings.ToUpper(strings.TrimSpace(raw)) {
		switch {
		case r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			sb.WriteRune(r)
		case r == '-' || r == ' ' || r == '_':
			// separadores: se descartan y se reinserta el guion canónico
		default:
			return "", ErrMalformed
		}
	}

	body := sb.String()
	if len(body) != letterCount+digitCount {
		return "", ErrMalformed
	}
	for i := range letterCount {
		if body[i] < 'A' || body[i] > 'Z' {
			return "", ErrMalformed
		}
	}
	for i := letterCount; i < len(body); i++ {
		if body[i] < '0' || body[i] > '9' {
			return "", ErrMalformed
		}
	}
	return body[:letterCount] + "-" + body[letterCount:], nil
}
