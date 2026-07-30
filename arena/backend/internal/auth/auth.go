// Package auth resuelve contraseñas, JWT de acceso y tokens opacos.
//
// Sin dependencias externas: PBKDF2 y HMAC-SHA256 están en la biblioteca
// estándar desde Go 1.24. Un JWT firmado con HS256 son unas cuarenta líneas y
// se lee entero. Es el mismo paquete que el backend del hipódromo
// (project/backend/internal/auth) a propósito: los dos backends se leen igual.
package auth

import (
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Duraciones del contrato (docs/contract/api.md): access de 15 minutos, refresh
// de un solo uso que vive un mes.
const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 30 * 24 * time.Hour
)

// pbkdf2Iterations: 210 000 es la recomendación de OWASP para PBKDF2-SHA256.
const pbkdf2Iterations = 210_000

// Longitudes de contraseña. El tope existe porque una contraseña de 1 MB es un
// PBKDF2 de 1 MB, o sea una denegación de servicio gratis.
const (
	MinPasswordLength = 8
	MaxPasswordLength = 72
)

var (
	ErrInvalidToken = errors.New("token inválido")
	ErrExpiredToken = errors.New("token vencido")
)

// Identity es quién hace la petición. El rol viaja acá porque se lee de la
// BASE en cada petición, no del token: si el instructor baja a alguien de
// admin, no hay que esperar quince minutos a que venza un JWT.
//
// El orden y los tipos de los campos son parte del acuerdo con el paquete de
// carreras: races.Identity y ws.Identity tienen la misma forma, así que una
// conversión de struct alcanza para cruzar el límite.
type Identity struct {
	UserID   string
	Username string
	Role     string
}

// ── Contraseñas ───────────────────────────────────────────────────────────

// HashPassword devuelve `pbkdf2$sha256$iteraciones$salt$hash`, todo en base64
// sin relleno. El formato lleva las iteraciones adentro para poder subirlas
// más adelante sin invalidar las contraseñas ya guardadas.
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generando salt: %w", err)
	}
	sum, err := pbkdf2.Key(sha256.New, password, salt, pbkdf2Iterations, 32)
	if err != nil {
		return "", fmt.Errorf("derivando la clave: %w", err)
	}
	enc := base64.RawStdEncoding
	return fmt.Sprintf("pbkdf2$sha256$%d$%s$%s", pbkdf2Iterations, enc.EncodeToString(salt), enc.EncodeToString(sum)), nil
}

// VerifyPassword compara en tiempo constante. Un `==` acá filtra información
// por el tiempo de respuesta.
func VerifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 5 || parts[0] != "pbkdf2" || parts[1] != "sha256" {
		return false
	}
	iterations, err := strconv.Atoi(parts[2])
	if err != nil || iterations <= 0 {
		return false
	}
	enc := base64.RawStdEncoding
	salt, err := enc.DecodeString(parts[3])
	if err != nil {
		return false
	}
	want, err := enc.DecodeString(parts[4])
	if err != nil {
		return false
	}
	got, err := pbkdf2.Key(sha256.New, password, salt, iterations, len(want))
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(got, want) == 1
}

// ── Tokens opacos ─────────────────────────────────────────────────────────

// NewOpaqueToken genera un token aleatorio con prefijo. El refresh es opaco a
// propósito: no lleva datos, se busca en la base y se puede invalidar de a uno.
// Un JWT no se puede revocar, y un refresh de un solo uso necesita justamente
// eso.
func NewOpaqueToken(prefix string) (string, error) {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generando token: %w", err)
	}
	return prefix + "_" + hex.EncodeToString(raw), nil
}

// HashToken es lo que se guarda en `refresh_tokens.token_hash`.
//
// La base guarda el hash y no el token: un volcado de la base no tiene que
// alcanzar para hacerse pasar por nadie. SHA-256 sin salt y sin estirar está
// bien acá y no lo estaría para una contraseña — el token tiene 192 bits de
// entropía, no se adivina con un diccionario.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// ── JWT de acceso ─────────────────────────────────────────────────────────

type Claims struct {
	Subject   string `json:"sub"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

// Signer firma y valida los access tokens.
//
// Las claims llevan solo el usuario. El ROL NO va en el token: se lee de la
// base en cada petición. Ver arena/CLAUDE.md §4 — «un alumno que edite su
// token no obtiene nada», y tampoco lo obtiene uno al que le sacaron el rol.
type Signer struct{ secret []byte }

func NewSigner(secret string) *Signer { return &Signer{secret: []byte(secret)} }

var b64 = base64.RawURLEncoding

// Sign emite un JWT HS256 para el usuario indicado.
func (s *Signer) Sign(userID string, now time.Time) (string, error) {
	header, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	claims, err := json.Marshal(Claims{
		Subject:   userID,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(AccessTokenTTL).Unix(),
	})
	if err != nil {
		return "", err
	}
	body := b64.EncodeToString(header) + "." + b64.EncodeToString(claims)
	return body + "." + b64.EncodeToString(s.sign(body)), nil
}

// Parse valida firma y vencimiento, y devuelve las claims.
func (s *Signer) Parse(token string, now time.Time) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	// La firma se verifica ANTES de mirar el contenido: nunca se confía en un
	// payload sin validar, ni siquiera para leerlo.
	want := s.sign(parts[0] + "." + parts[1])
	got, err := b64.DecodeString(parts[2])
	if err != nil || !hmac.Equal(got, want) {
		return nil, ErrInvalidToken
	}

	raw, err := b64.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var claims Claims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return nil, ErrInvalidToken
	}
	if claims.Subject == "" {
		return nil, ErrInvalidToken
	}
	if now.Unix() >= claims.ExpiresAt {
		return nil, ErrExpiredToken
	}
	return &claims, nil
}

func (s *Signer) sign(body string) []byte {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(body))
	return mac.Sum(nil)
}
