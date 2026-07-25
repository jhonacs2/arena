// Package mail manda el correo de verificación.
//
// Dos implementaciones: una que escribe en el log y otra que llama a Resend.
// La de log NO es un stub incompleto — imprime el enlace completo, así que en
// desarrollo y en clase se puede verificar una cuenta sin configurar nada ni
// tener acceso a una casilla real. Es lo que hace que la pantalla `/verificar`
// se pueda demostrar en vivo.
package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Sender manda el correo de verificación a una dirección.
type Sender interface {
	SendVerification(ctx context.Context, to, displayName, token string) error
}

// VerificationLink arma el enlace que va en el correo. El front lee el token
// del query y hace el POST — está en docs/contract/openapi.yaml.
func VerificationLink(frontURL, token string) string {
	return strings.TrimRight(frontURL, "/") + "/verificar?token=" + url.QueryEscape(token)
}

// ── Emisor de desarrollo ──────────────────────────────────────────────────

// LogSender imprime el enlace en el log en lugar de mandar un correo.
type LogSender struct {
	Log      *slog.Logger
	FrontURL string
}

func (s *LogSender) SendVerification(_ context.Context, to, displayName, token string) error {
	link := VerificationLink(s.FrontURL, token)
	// Se imprime bien visible: en clase esto se copia y se pega en el navegador.
	s.Log.Info("correo de verificación (modo desarrollo, no se envió nada)",
		"para", to, "nombre", displayName, "enlace", link)
	fmt.Printf("\n  ┌─ VERIFICAR CUENTA ─────────────────────────────────────────\n"+
		"  │  %s\n  │  %s\n"+
		"  └────────────────────────────────────────────────────────────\n\n", to, link)
	return nil
}

// ── Emisor real ───────────────────────────────────────────────────────────

// ResendSender manda el correo por la API de Resend.
type ResendSender struct {
	APIKey   string
	From     string
	FrontURL string
	Log      *slog.Logger
	Client   *http.Client
}

func NewResendSender(apiKey, from, frontURL string, log *slog.Logger) *ResendSender {
	return &ResendSender{
		APIKey: apiKey, From: from, FrontURL: frontURL, Log: log,
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *ResendSender) SendVerification(ctx context.Context, to, displayName, token string) error {
	link := VerificationLink(s.FrontURL, token)

	payload, err := json.Marshal(map[string]any{
		"from":    s.From,
		"to":      []string{to},
		"subject": "Verificá tu cuenta del Hipódromo",
		"html":    verificationHTML(displayName, link),
		"text": fmt.Sprintf("Hola %s,\n\nVerificá tu cuenta entrando acá:\n%s\n\n"+
			"El enlace vence en 24 horas.\n", displayName, link),
	})
	if err != nil {
		return fmt.Errorf("armando el correo: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("llamando a Resend: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		body := make([]byte, 512)
		n, _ := res.Body.Read(body)
		return fmt.Errorf("Resend devolvió %d: %s", res.StatusCode, body[:n])
	}
	s.Log.Info("correo de verificación enviado", "para", to)
	return nil
}

// verificationHTML usa la misma dirección visual que la app: borde sólido de
// 3 px, sombra dura sin blur, cero gradientes. Todo en línea, porque los
// clientes de correo descartan las hojas de estilo externas.
func verificationHTML(displayName, link string) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="es"><body style="margin:0;padding:32px;background:#f7f7f2;font-family:'Segoe UI',system-ui,sans-serif;color:#14201a">
  <table role="presentation" style="max-width:520px;margin:0 auto;border-collapse:collapse">
    <tr><td style="background:#f7f7f2;border:3px solid #14201a;box-shadow:4px 4px 0 #14201a;padding:32px">
      <h1 style="margin:0 0 16px;font-size:28px;line-height:1.1">Bienvenido al Hipódromo</h1>
      <p style="margin:0 0 12px;font-size:16px;line-height:1.5">Hola %s,</p>
      <p style="margin:0 0 24px;font-size:16px;line-height:1.5">
        Verificá tu cuenta para empezar a apostar. Tenés 1000 de saldo virtual esperándote.
      </p>
      <a href="%s" style="display:inline-block;background:#e3a04a;color:#14201a;text-decoration:none;
        font-weight:700;font-size:16px;padding:14px 28px;border:3px solid #14201a;box-shadow:4px 4px 0 #14201a">
        Verificar mi cuenta
      </a>
      <p style="margin:28px 0 0;font-size:13px;line-height:1.5;color:#5c6660">
        El enlace vence en 24 horas. Si no funciona, copiá y pegá esta dirección:<br>
        <span style="word-break:break-all">%s</span>
      </p>
    </td></tr>
    <tr><td style="padding:16px 4px;font-size:12px;color:#5c6660">
      Saldo virtual. No se mueve dinero real en ningún momento.
    </td></tr>
  </table>
</body></html>`, displayName, link, link)
}
