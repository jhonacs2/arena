package api

import (
	"context"
	"net/http"
	netmail "net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/talentodh/hipodromo/internal/auth"
	"github.com/talentodh/hipodromo/internal/contract"
)

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var body registerRequest
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, r, err)
		return
	}

	fields := map[string]string{}
	if !validEmail(body.Email) {
		fields["email"] = "Ingresá un correo válido."
	}
	if utf8.RuneCountInString(body.Password) < 8 {
		fields["password"] = "La contraseña necesita al menos 8 caracteres."
	}
	if utf8.RuneCountInString(body.Password) > 72 {
		fields["password"] = "La contraseña no puede tener más de 72 caracteres."
	}
	name := strings.TrimSpace(body.DisplayName)
	if utf8.RuneCountInString(name) < 2 {
		fields["displayName"] = "El nombre necesita al menos 2 caracteres."
	}
	if utf8.RuneCountInString(name) > 40 {
		fields["displayName"] = "El nombre no puede tener más de 40 caracteres."
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

	user, err := s.Store.CreateUser(body.Email, name, hash)
	if err != nil {
		writeError(w, r, err)
		return
	}

	s.sendVerification(r, user.ID, user.Email, user.DisplayName)
	writeJSON(w, http.StatusCreated, user.User)
}

type verifyRequest struct {
	Token string `json:"token"`
}

func (s *Server) handleVerify(w http.ResponseWriter, r *http.Request) {
	var body verifyRequest
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, r, err)
		return
	}
	if strings.TrimSpace(body.Token) == "" {
		writeError(w, r, contract.Errorf(contract.CodeInvalidVerificationToken))
		return
	}

	userID, expired, ok := s.Store.ConsumeVerificationToken(body.Token, s.now())
	if expired {
		writeError(w, r, contract.Errorf(contract.CodeVerificationTokenExpired))
		return
	}
	if !ok {
		writeError(w, r, contract.Errorf(contract.CodeInvalidVerificationToken))
		return
	}

	existing, found := s.Store.UserByID(userID)
	if !found {
		writeError(w, r, contract.Errorf(contract.CodeInvalidVerificationToken))
		return
	}
	if existing.EmailVerified {
		writeError(w, r, contract.Errorf(contract.CodeAlreadyVerified))
		return
	}

	user, _ := s.Store.MarkVerified(userID)
	writeJSON(w, http.StatusOK, contract.VerifiedUser{User: user})
}

type resendRequest struct {
	Email string `json:"email"`
}

func (s *Server) handleResendVerification(w http.ResponseWriter, r *http.Request) {
	var body resendRequest
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, r, err)
		return
	}

	if allowed, retry := s.resendLimiter.allow(strings.ToLower(body.Email)); !allowed {
		writeError(w, r, contract.ErrorWith(contract.CodeRateLimited,
			map[string]any{"retryAfterSeconds": retry}))
		return
	}

	// 202 aunque el correo no exista. Responder 404 convertiría este endpoint
	// en un buscador de cuentas registradas.
	user, found := s.Store.UserByEmail(body.Email)
	if found && !user.EmailVerified {
		s.sendVerification(r, user.ID, user.Email, user.DisplayName)
	}
	writeJSON(w, http.StatusAccepted, nil)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body loginRequest
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, r, err)
		return
	}

	if allowed, retry := s.loginLimiter.allow(clientIP(r)); !allowed {
		writeError(w, r, contract.ErrorWith(contract.CodeRateLimited,
			map[string]any{"retryAfterSeconds": retry}))
		return
	}

	user, found := s.Store.UserByEmail(body.Email)
	// Mismo error para "no existe" y "contraseña mal": distinguirlos le diría
	// a cualquiera qué correos están registrados.
	if !found || !auth.VerifyPassword(body.Password, user.PasswordHash) {
		writeError(w, r, contract.Errorf(contract.CodeInvalidCredentials))
		return
	}

	tokens, err := s.issueTokens(user.ID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, contract.AuthTokens{
		AccessToken: tokens.AccessToken, RefreshToken: tokens.RefreshToken, User: user.User,
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var body refreshRequest
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, r, err)
		return
	}

	// Canjear invalida el token: es de un solo uso. Reusar uno ya canjeado
	// falla, y eso es lo que delata un token robado.
	userID, ok := s.Store.ConsumeRefreshToken(body.RefreshToken, s.now())
	if !ok {
		writeError(w, r, contract.Errorf(contract.CodeInvalidRefreshToken))
		return
	}

	tokens, err := s.issueTokens(userID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	id, err := requireUser(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	s.Store.RevokeUserTokens(id)
	writeJSON(w, http.StatusNoContent, nil)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	id, err := requireUser(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	user, found := s.Store.UserByID(id)
	if !found {
		writeError(w, r, contract.Errorf(contract.CodeUnauthenticated))
		return
	}
	writeJSON(w, http.StatusOK, user.User)
}

// ── Auxiliares ────────────────────────────────────────────────────────────

func (s *Server) issueTokens(userID string) (contract.RefreshedTokens, error) {
	now := s.now()

	access, err := s.Signer.Sign(userID, now)
	if err != nil {
		return contract.RefreshedTokens{}, err
	}
	refresh, err := auth.NewOpaqueToken("rt")
	if err != nil {
		return contract.RefreshedTokens{}, err
	}
	s.Store.SaveRefreshToken(refresh, userID, now.Add(auth.RefreshTokenTTL))

	return contract.RefreshedTokens{AccessToken: access, RefreshToken: refresh}, nil
}

// sendVerification emite el token y manda el correo sin bloquear la respuesta.
// Registrarse no debería tardar lo que tarda un proveedor de correo en
// contestar, y si el envío falla la cuenta igual quedó creada: se reintenta
// desde /auth/resend-verification.
func (s *Server) sendVerification(r *http.Request, userID, email, displayName string) {
	token, err := auth.NewOpaqueToken("vt")
	if err != nil {
		s.Log.Error("no se pudo generar el token de verificación", "error", err, "usuario", userID)
		return
	}
	s.Store.SaveVerificationToken(token, userID, s.now().Add(auth.VerificationTokenTTL))

	go func() {
		// Contexto propio: el de la petición muere al responder.
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := s.Mail.SendVerification(ctx, email, displayName, token); err != nil {
			s.Log.Error("no se pudo enviar el correo de verificación", "error", err, "para", email)
		}
	}()
}

func validEmail(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 254 || !strings.Contains(value, "@") {
		return false
	}
	_, err := netmail.ParseAddress(value)
	return err == nil
}
