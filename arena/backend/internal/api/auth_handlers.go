package api

import (
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/talentodh/arena/internal/accounts"
	"github.com/talentodh/arena/internal/auth"
	"github.com/talentodh/arena/internal/contract"
	"github.com/talentodh/arena/internal/invite"
)

// CookieOptions es cómo se emite la cookie del refresh.
//
// Se configura por entorno porque el desarrollo y la producción no se parecen:
// en producción el frontend está en Cloudflare Pages y llama a `/api` en su
// propio dominio (SameSite=Lax alcanza); en desarrollo es localhost:4200 contra
// localhost:8080, que para el navegador son sitios distintos y necesita
// SameSite=None con Secure.
type CookieOptions struct {
	Name     string
	Path     string
	Domain   string
	Secure   bool
	SameSite http.SameSite
}

// DefaultCookieOptions es el default de producción.
//
// El Path acotado a /api/auth no es cosmético: la cookie solo la necesitan
// refresh y logout, así que no viaja en las otras doscientas peticiones de una
// clase. Menos superficie y menos bytes.
func DefaultCookieOptions() CookieOptions {
	return CookieOptions{
		Name:     "arena_refresh",
		Path:     BasePath + "/auth",
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}
}

// sessionResponse es lo que devuelven redeem, login y refresh (api.md).
//
// El refresh NO va en el cuerpo: va en la cookie HttpOnly, que es todo el punto
// de que sea HttpOnly. Si viajara en el JSON, cualquier XSS podría leerlo.
type sessionResponse struct {
	AccessToken string          `json:"accessToken"`
	User        accounts.User   `json:"user"`
	Balance     int64           `json:"balance"`
	Points      accounts.Points `json:"points"`
}

// ── POST /api/auth/check-code ─────────────────────────────────────────────

type checkCodeRequest struct {
	Code string `json:"code"`
}

type checkCodeResponse struct {
	Valid        bool  `json:"valid"`
	CoinsGranted int64 `json:"coinsGranted"`
}

// handleCheckCode valida el código SIN canjearlo, para habilitar el resto del
// formulario de registro.
func (s *Server) handleCheckCode(w http.ResponseWriter, r *http.Request) {
	if err := s.codeLimiter.check(r); err != nil {
		writeError(w, r, err)
		return
	}

	var body checkCodeRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, r, err)
		return
	}

	code, err := invite.Normalize(body.Code)
	if err != nil {
		writeError(w, r, contract.FieldErrors(map[string]string{
			"code": "El código tiene cuatro letras, un guion y cuatro números: AVBD-1234.",
		}))
		return
	}

	coins, err := s.Accounts.CheckCode(r.Context(), code)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, checkCodeResponse{Valid: true, CoinsGranted: coins})
}

// ── POST /api/auth/redeem ─────────────────────────────────────────────────

type redeemRequest struct {
	Code      string `json:"code"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

func (s *Server) handleRedeem(w http.ResponseWriter, r *http.Request) {
	if err := s.codeLimiter.check(r); err != nil {
		writeError(w, r, err)
		return
	}

	var body redeemRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, r, err)
		return
	}

	fields := map[string]string{}

	code, codeErr := invite.Normalize(body.Code)
	if codeErr != nil {
		fields["code"] = "El código tiene cuatro letras, un guion y cuatro números: AVBD-1234."
	}
	firstName := strings.TrimSpace(body.FirstName)
	if !validName(firstName) {
		fields["firstName"] = "Escribí tu nombre (entre 2 y 60 caracteres)."
	}
	lastName := strings.TrimSpace(body.LastName)
	if !validName(lastName) {
		fields["lastName"] = "Escribí tu apellido (entre 2 y 60 caracteres)."
	}
	username := strings.TrimSpace(body.Username)
	if !validUsername(username) {
		fields["username"] = "El usuario va de 3 a 24 caracteres, con letras, números, punto, guion o guion bajo."
	}
	if n := utf8.RuneCountInString(body.Password); n < auth.MinPasswordLength || n > auth.MaxPasswordLength {
		fields["password"] = "La contraseña necesita al menos 8 caracteres."
	}
	if len(fields) > 0 {
		writeError(w, r, contract.FieldErrors(fields))
		return
	}

	hash, err := auth.HashPassword(body.Password)
	if err != nil {
		writeError(w, r, err)
		return
	}

	user, err := s.Accounts.Redeem(r.Context(), accounts.RedeemInput{
		Code:         code,
		FirstName:    firstName,
		LastName:     lastName,
		Username:     username,
		PasswordHash: hash,
	})
	if err != nil {
		writeError(w, r, err)
		return
	}

	s.respondWithSession(w, r, user, http.StatusCreated)
}

// ── POST /api/auth/login ──────────────────────────────────────────────────

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := s.loginLimiter.check(r); err != nil {
		writeError(w, r, err)
		return
	}

	var body loginRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, r, err)
		return
	}

	user, err := s.Accounts.Authenticate(r.Context(), strings.TrimSpace(body.Username),
		func(hash string) bool { return auth.VerifyPassword(body.Password, hash) })
	if err != nil {
		writeError(w, r, err)
		return
	}

	s.respondWithSession(w, r, user, http.StatusOK)
}

// ── POST /api/auth/refresh ────────────────────────────────────────────────

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(s.Cookie.Name)
	if err != nil || cookie.Value == "" {
		writeError(w, r, contract.Errorf(contract.CodeUnauthenticated))
		return
	}

	userID, err := s.Accounts.ConsumeRefreshToken(r.Context(), auth.HashToken(cookie.Value), s.now())
	switch {
	case errors.Is(err, accounts.ErrTokenReused):
		// Un token de un solo uso que aparece dos veces: alguien tiene una copia.
		// Ya se revocaron todas las sesiones del dueño; acá solo queda dejar
		// rastro, porque es lo único que va a existir si alguien pregunta después.
		s.Log.Warn("refresh token reusado, se revocaron las sesiones", "usuario", userID)
		s.clearRefreshCookie(w)
		writeError(w, r, contract.Errorf(contract.CodeUnauthenticated))
		return
	case errors.Is(err, accounts.ErrNotFound):
		s.clearRefreshCookie(w)
		writeError(w, r, contract.Errorf(contract.CodeUnauthenticated))
		return
	case err != nil:
		writeError(w, r, err)
		return
	}

	user, err := s.Accounts.ByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, accounts.ErrNotFound) {
			s.clearRefreshCookie(w)
			writeError(w, r, contract.Errorf(contract.CodeUnauthenticated))
			return
		}
		writeError(w, r, err)
		return
	}

	s.respondWithSession(w, r, user, http.StatusOK)
}

// ── POST /api/auth/logout ─────────────────────────────────────────────────

// handleLogout corta TODAS las sesiones del usuario, no solo la del navegador
// que llamó.
//
// Es lo que se espera de un «cerrar sesión» en una app donde las monedas son
// nota: si alguien la usó desde la máquina del aula, cerrar sesión tiene que
// dejarla cerrada ahí también.
//
// Responde 204 aun sin sesión: un logout no tiene por qué fallar nunca, y
// devolver 401 solo lograría que el frontend tenga que manejar un caso que no le
// aporta nada.
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(s.Cookie.Name); err == nil && cookie.Value != "" {
		if userID, err := s.Accounts.ConsumeRefreshToken(r.Context(), auth.HashToken(cookie.Value), s.now()); err == nil {
			if err := s.Accounts.RevokeRefreshTokens(r.Context(), userID); err != nil {
				s.Log.Error("no se pudieron revocar las sesiones", "error", err, "usuario", userID)
			}
		}
	}
	if user, ok := currentUser(r); ok {
		if err := s.Accounts.RevokeRefreshTokens(r.Context(), user.ID); err != nil {
			s.Log.Error("no se pudieron revocar las sesiones", "error", err, "usuario", user.ID)
		}
	}

	s.clearRefreshCookie(w)
	writeJSON(w, http.StatusNoContent, nil)
}

// ── Auxiliares ────────────────────────────────────────────────────────────

// respondWithSession emite el par de tokens y contesta con el sobre de sesión.
func (s *Server) respondWithSession(w http.ResponseWriter, r *http.Request, user accounts.User, status int) {
	now := s.now()

	access, err := s.Signer.Sign(user.ID, now)
	if err != nil {
		writeError(w, r, err)
		return
	}

	refresh, err := auth.NewOpaqueToken("rt")
	if err != nil {
		writeError(w, r, err)
		return
	}
	expiresAt := now.Add(auth.RefreshTokenTTL)
	if err := s.Accounts.SaveRefreshToken(r.Context(), auth.HashToken(refresh), user.ID, now, expiresAt); err != nil {
		writeError(w, r, err)
		return
	}

	points, err := s.Accounts.PointsFor(r.Context(), user.ID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	s.setRefreshCookie(w, refresh, expiresAt)
	writeJSON(w, status, sessionResponse{
		AccessToken: access,
		User:        user,
		Balance:     user.Balance,
		Points:      points,
	})
}

func (s *Server) setRefreshCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:   s.Cookie.Name,
		Value:  token,
		Path:   s.Cookie.Path,
		Domain: s.Cookie.Domain,
		// HttpOnly siempre: el JavaScript del frontend no tiene por qué poder
		// leerlo, y así un XSS no se lleva la sesión.
		HttpOnly: true,
		Secure:   s.Cookie.Secure,
		SameSite: s.Cookie.SameSite,
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
	})
}

func (s *Server) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.Cookie.Name,
		Value:    "",
		Path:     s.Cookie.Path,
		Domain:   s.Cookie.Domain,
		HttpOnly: true,
		Secure:   s.Cookie.Secure,
		SameSite: s.Cookie.SameSite,
		MaxAge:   -1,
	})
}

func validName(value string) bool {
	n := utf8.RuneCountInString(value)
	return n >= 2 && n <= 60
}

// validUsername acota el usuario a lo que se puede dictar y escribir sin dudas.
// Sin espacios ni acentos: el login es case-insensitive pero no
// «acento-insensitive», y «Gómez» contra «Gomez» es una llamada de soporte.
func validUsername(value string) bool {
	n := len(value)
	if n < 3 || n > 24 {
		return false
	}
	for _, r := range value {
		switch {
		case r > unicode.MaxASCII:
			return false
		case unicode.IsLetter(r), unicode.IsDigit(r):
		case r == '.', r == '-', r == '_':
		default:
			return false
		}
	}
	return true
}
