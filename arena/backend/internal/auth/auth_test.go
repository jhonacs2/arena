package auth

import (
	"strings"
	"testing"
	"time"
)

func TestHashPasswordVerifica(t *testing.T) {
	hash, err := HashPassword("caballo-de-batalla")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !strings.HasPrefix(hash, "pbkdf2$sha256$") {
		t.Fatalf("el formato del hash cambió: %q", hash)
	}
	if !VerifyPassword("caballo-de-batalla", hash) {
		t.Error("la contraseña correcta no verifica")
	}
	if VerifyPassword("caballo-de-batall", hash) {
		t.Error("una contraseña incorrecta verifica")
	}
}

func TestHashPasswordNuncaRepiteElSalt(t *testing.T) {
	uno, err := HashPassword("misma-contraseña")
	if err != nil {
		t.Fatal(err)
	}
	dos, err := HashPassword("misma-contraseña")
	if err != nil {
		t.Fatal(err)
	}
	if uno == dos {
		t.Error("dos hashes de la misma contraseña son iguales: falta el salt")
	}
}

func TestVerifyPasswordRechazaBasura(t *testing.T) {
	for _, encoded := range []string{
		"", "$", "pbkdf2$sha256$1$a$b$c", "bcrypt$sha256$1$YQ$Yg",
		"pbkdf2$sha512$1$YQ$Yg", "pbkdf2$sha256$0$YQ$Yg", "pbkdf2$sha256$x$YQ$Yg",
	} {
		if VerifyPassword("cualquiera", encoded) {
			t.Errorf("verificó contra un hash inválido: %q", encoded)
		}
	}
}

func TestSignYParse(t *testing.T) {
	signer := NewSigner("secreto-de-prueba")
	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)

	token, err := signer.Sign("11111111-1111-1111-1111-111111111111", now)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	claims, err := signer.Parse(token, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.Subject != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("sub = %q", claims.Subject)
	}
	if want := now.Add(AccessTokenTTL).Unix(); claims.ExpiresAt != want {
		t.Errorf("exp = %d, se esperaba %d", claims.ExpiresAt, want)
	}
}

func TestParseRechazaTokenVencido(t *testing.T) {
	signer := NewSigner("secreto-de-prueba")
	now := time.Now()

	token, err := signer.Sign("usuario", now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := signer.Parse(token, now.Add(AccessTokenTTL+time.Second)); err != ErrExpiredToken {
		t.Errorf("err = %v, se esperaba ErrExpiredToken", err)
	}
}

func TestParseRechazaFirmaAjena(t *testing.T) {
	now := time.Now()
	token, err := NewSigner("secreto-verdadero").Sign("usuario", now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewSigner("otro-secreto").Parse(token, now); err != ErrInvalidToken {
		t.Errorf("err = %v, se esperaba ErrInvalidToken", err)
	}
}

// El caso que importa: alguien cambia el payload para hacerse pasar por otro.
// La firma se verifica antes de leer el contenido, así que no llega a nada.
func TestParseRechazaPayloadManipulado(t *testing.T) {
	signer := NewSigner("secreto-de-prueba")
	now := time.Now()

	token, err := signer.Sign("alumno", now)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(token, ".")
	forged := parts[0] + "." + b64.EncodeToString([]byte(
		`{"sub":"instructor","iat":0,"exp":99999999999}`)) + "." + parts[2]

	if _, err := signer.Parse(forged, now); err != ErrInvalidToken {
		t.Errorf("err = %v, se esperaba ErrInvalidToken", err)
	}
}

func TestParseRechazaFormaInvalida(t *testing.T) {
	signer := NewSigner("secreto-de-prueba")
	for _, token := range []string{"", "a", "a.b", "a.b.c.d", "...", "a.b.c"} {
		if _, err := signer.Parse(token, time.Now()); err == nil {
			t.Errorf("aceptó %q", token)
		}
	}
}

func TestNewOpaqueTokenEsUnicoYLlevaPrefijo(t *testing.T) {
	vistos := map[string]bool{}
	for range 100 {
		token, err := NewOpaqueToken("rt")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(token, "rt_") {
			t.Fatalf("sin prefijo: %q", token)
		}
		if vistos[token] {
			t.Fatalf("token repetido: %q", token)
		}
		vistos[token] = true
	}
}

func TestHashTokenEsEstableYNoDevuelveElToken(t *testing.T) {
	token := "rt_abc123"
	hash := HashToken(token)
	if hash != HashToken(token) {
		t.Error("HashToken no es estable")
	}
	if strings.Contains(hash, token) {
		t.Error("el hash contiene el token")
	}
	if len(hash) != 64 {
		t.Errorf("len(hash) = %d, se esperaba 64 (sha256 en hexa)", len(hash))
	}
}
