package invite

import (
	"regexp"
	"strings"
	"testing"
)

// El formato que exige el CHECK code_formato de schema.sql.
var schemaFormat = regexp.MustCompile(`^[A-Z]{4}-[0-9]{4}$`)

func TestElAlfabetoNoTieneCaracteresAmbiguos(t *testing.T) {
	for _, prohibida := range "ILOU" {
		if strings.ContainsRune(Letters, prohibida) {
			t.Errorf("Letters contiene %q, que se confunde al dictarla", prohibida)
		}
	}
	for _, prohibido := range "01" {
		if strings.ContainsRune(Digits, prohibido) {
			t.Errorf("Digits contiene %q, que se confunde al dictarlo", prohibido)
		}
	}
	if len(Letters) != 22 {
		t.Errorf("len(Letters) = %d, se esperaban 22 (26 menos I, L, O, U)", len(Letters))
	}
	if len(Digits) != 8 {
		t.Errorf("len(Digits) = %d, se esperaban 8 (10 menos 0 y 1)", len(Digits))
	}
}

func TestGenerateRespetaElFormatoYElAlfabeto(t *testing.T) {
	for range 500 {
		code, err := Generate()
		if err != nil {
			t.Fatalf("Generate: %v", err)
		}
		if !schemaFormat.MatchString(code) {
			t.Fatalf("%q no cumple el CHECK code_formato del esquema", code)
		}
		for i := range letterCount {
			if !strings.ContainsRune(Letters, rune(code[i])) {
				t.Fatalf("%q usa la letra ambigua %q", code, code[i])
			}
		}
		for i := letterCount + 1; i < len(code); i++ {
			if !strings.ContainsRune(Digits, rune(code[i])) {
				t.Fatalf("%q usa el dígito ambiguo %q", code, code[i])
			}
		}
	}
}

func TestGenerateNoRepiteEnSeguida(t *testing.T) {
	vistos := map[string]bool{}
	for range 1000 {
		code, err := Generate()
		if err != nil {
			t.Fatal(err)
		}
		if vistos[code] {
			t.Fatalf("código repetido en mil intentos: %q", code)
		}
		vistos[code] = true
	}
}

func TestNormalize(t *testing.T) {
	casos := []struct {
		entrada string
		quiere  string
	}{
		{"AVBD-1234", "AVBD-1234"},
		{"avbd-1234", "AVBD-1234"},
		{"AVBD1234", "AVBD-1234"},
		{"  avbd 1234  ", "AVBD-1234"},
		{"avbd_1234", "AVBD-1234"},
		// El guion en el lugar equivocado también se recupera: los separadores se
		// descartan y el canónico se rearma. Lo que no se adivina son los
		// caracteres.
		{"AVB-D1234", "AVBD-1234"},
		// Formato válido con letras que el generador nunca produce: pasa, y
		// muere después con CODE_NOT_FOUND.
		{"AVBI-1034", "AVBI-1034"},
	}
	for _, caso := range casos {
		got, err := Normalize(caso.entrada)
		if err != nil {
			t.Errorf("Normalize(%q): %v", caso.entrada, err)
			continue
		}
		if got != caso.quiere {
			t.Errorf("Normalize(%q) = %q, se esperaba %q", caso.entrada, got, caso.quiere)
		}
		if !schemaFormat.MatchString(got) {
			t.Errorf("Normalize(%q) = %q, que no cumple el CHECK del esquema", caso.entrada, got)
		}
	}
}

func TestNormalizeRechazaLoQueNoEsUnCodigo(t *testing.T) {
	for _, entrada := range []string{
		"", "AVBD", "1234", "AVBD-123", "AVBD-12345",
		"1234-AVBD", "AVBD-12A4", "AVB!-1234", "AVBDÑ1234", "AVBD-1234-5678",
	} {
		if got, err := Normalize(entrada); err == nil {
			t.Errorf("Normalize(%q) = %q, se esperaba ErrMalformed", entrada, got)
		}
	}
}
