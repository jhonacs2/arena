package contract

import "net/http"

// APIError es el sobre uniforme de error de docs/contract/error-codes.md.
//
//	{ "error": { "code": "…", "message": "…", "details": {} } }
//
// El frontend hace switch sobre `code`, nunca sobre `message`. Por eso el
// catálogo es cerrado: agregar un código acá obliga a agregarlo también en
// error-codes.md, y `verify.mjs` compara las dos listas.
type APIError struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details"`
}

// Code es un código del catálogo cerrado.
type Code string

const (
	CodeValidationFailed         Code = "VALIDATION_FAILED"
	CodeInvalidCredentials       Code = "INVALID_CREDENTIALS"
	CodeEmailAlreadyRegistered   Code = "EMAIL_ALREADY_REGISTERED"
	CodeEmailNotVerified         Code = "EMAIL_NOT_VERIFIED"
	CodeInvalidVerificationToken Code = "INVALID_VERIFICATION_TOKEN"
	CodeVerificationTokenExpired Code = "VERIFICATION_TOKEN_EXPIRED"
	CodeAlreadyVerified          Code = "ALREADY_VERIFIED"
	CodeUnauthenticated          Code = "UNAUTHENTICATED"
	CodeInvalidRefreshToken      Code = "INVALID_REFRESH_TOKEN"
	CodeForbidden                Code = "FORBIDDEN"
	CodeNotFound                 Code = "NOT_FOUND"
	CodeRaceAlreadyStarted       Code = "RACE_ALREADY_STARTED"
	CodeHorseNotInRace           Code = "HORSE_NOT_IN_RACE"
	CodeInsufficientBalance      Code = "INSUFFICIENT_BALANCE"
	CodeBetAmountOutOfRange      Code = "BET_AMOUNT_OUT_OF_RANGE"
	CodeResultsNotAvailable      Code = "RESULTS_NOT_AVAILABLE"
	CodeRateLimited              Code = "RATE_LIMITED"
	CodeInternal                 Code = "INTERNAL"
)

// catalog fija el estado HTTP y el mensaje en español de cada código.
// El mensaje se muestra tal cual al usuario, así que se escribe como se
// quiere leer en pantalla.
var catalog = map[Code]struct {
	Status  int
	Message string
}{
	CodeValidationFailed:         {http.StatusUnprocessableEntity, "Revisá los campos marcados."},
	CodeInvalidCredentials:       {http.StatusUnauthorized, "Correo o contraseña incorrectos."},
	CodeEmailAlreadyRegistered:   {http.StatusConflict, "Ese correo ya tiene una cuenta."},
	CodeEmailNotVerified:         {http.StatusForbidden, "Verificá tu correo antes de apostar."},
	CodeInvalidVerificationToken: {http.StatusBadRequest, "El enlace de verificación no es válido."},
	CodeVerificationTokenExpired: {http.StatusGone, "El enlace de verificación venció. Pedí uno nuevo."},
	CodeAlreadyVerified:          {http.StatusConflict, "Esta cuenta ya está verificada."},
	CodeUnauthenticated:          {http.StatusUnauthorized, "Iniciá sesión para continuar."},
	CodeInvalidRefreshToken:      {http.StatusUnauthorized, "Tu sesión expiró. Iniciá sesión de nuevo."},
	CodeForbidden:                {http.StatusForbidden, "No tenés acceso a este recurso."},
	CodeNotFound:                 {http.StatusNotFound, "No encontramos lo que buscabas."},
	CodeRaceAlreadyStarted:       {http.StatusConflict, "La carrera ya arrancó. No se puede apostar."},
	CodeHorseNotInRace:           {http.StatusUnprocessableEntity, "Ese caballo no corre en esta carrera."},
	CodeInsufficientBalance:      {http.StatusConflict, "No te alcanza el saldo para esta apuesta."},
	CodeBetAmountOutOfRange:      {http.StatusUnprocessableEntity, "El monto tiene que estar entre 10 y 5000."},
	CodeResultsNotAvailable:      {http.StatusConflict, "Esta carrera todavía no terminó."},
	CodeRateLimited:              {http.StatusTooManyRequests, "Demasiados intentos. Esperá un momento."},
	CodeInternal:                 {http.StatusInternalServerError, "Algo salió mal de nuestro lado."},
}

// Fault es un error del dominio que sabe cómo convertirse en respuesta HTTP.
type Fault struct {
	Code    Code
	Details map[string]any
}

func (f *Fault) Error() string { return string(f.Code) }

// Status y Message salen del catálogo. Un código fuera de catálogo se degrada
// a 500 en vez de responder con estado 0 — que es peor de depurar.
func (f *Fault) Status() int {
	if e, ok := catalog[f.Code]; ok {
		return e.Status
	}
	return http.StatusInternalServerError
}

func (f *Fault) Message() string {
	if e, ok := catalog[f.Code]; ok {
		return e.Message
	}
	return catalog[CodeInternal].Message
}

// Body arma el sobre. `details` siempre existe: `{}` cuando no hay nada que
// agregar, nunca `null`, porque el frontend lo lee sin chequear.
func (f *Fault) Body() APIError {
	details := f.Details
	if details == nil {
		details = map[string]any{}
	}
	return APIError{Error: ErrorBody{
		Code:    string(f.Code),
		Message: f.Message(),
		Details: details,
	}}
}

// Errorf crea un Fault sin detalles.
func Errorf(code Code) *Fault { return &Fault{Code: code} }

// ErrorWith crea un Fault con detalles.
//
//	contract.ErrorWith(contract.CodeInsufficientBalance, map[string]any{
//		"balance": 980, "amount": 5000,
//	})
func ErrorWith(code Code, details map[string]any) *Fault {
	return &Fault{Code: code, Details: details}
}

// FieldErrors arma un VALIDATION_FAILED con errores por campo.
func FieldErrors(fields map[string]string) *Fault {
	as := make(map[string]any, len(fields))
	for k, v := range fields {
		as[k] = v
	}
	return &Fault{Code: CodeValidationFailed, Details: map[string]any{"fields": as}}
}

// KnownCodes es el catálogo completo. Lo usa el test que lo compara contra
// docs/contract/error-codes.md.
func KnownCodes() []Code {
	out := make([]Code, 0, len(catalog))
	for c := range catalog {
		out = append(out, c)
	}
	return out
}
