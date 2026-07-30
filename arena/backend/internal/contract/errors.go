// Package contract es el sobre de error y los tipos del cable.
//
// Implementa docs/contract/api.md. El catálogo de códigos es CERRADO: agregar
// uno acá obliga a agregarlo también en api.md, y verify-arena.mjs compara las
// dos listas.
package contract

import "net/http"

// APIError es el sobre uniforme de error de api.md.
//
//	{ "error": { "code": "…", "message": "…", "details": {} } }
//
// El frontend hace switch sobre `code`, nunca sobre `message`.
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
	CodeValidationFailed    Code = "VALIDATION_FAILED"
	CodeCodeNotFound        Code = "CODE_NOT_FOUND"
	CodeCodeAlreadyRedeemed Code = "CODE_ALREADY_REDEEMED"
	CodeUsernameTaken       Code = "USERNAME_TAKEN"
	CodeInvalidCredentials  Code = "INVALID_CREDENTIALS"
	CodeUnauthenticated     Code = "UNAUTHENTICATED"
	CodeForbidden           Code = "FORBIDDEN"
	CodeUserNotFound        Code = "USER_NOT_FOUND"
	CodeRateLimited         Code = "RATE_LIMITED"
	CodeRaceNotFound        Code = "RACE_NOT_FOUND"
	CodeRaceNotOpen         Code = "RACE_NOT_OPEN"
	CodeBetAlreadyPlaced    Code = "BET_ALREADY_PLACED"
	CodeInsufficientBalance Code = "INSUFFICIENT_BALANCE"
	CodeInvalidTransition   Code = "INVALID_TRANSITION"
	CodeInternal            Code = "INTERNAL"
)

// catalog fija el estado HTTP y el mensaje en castellano de cada código.
//
// El mensaje se muestra TAL CUAL al usuario, así que se escribe como se quiere
// leer en pantalla. Los `code` son para el frontend y van en inglés; ver
// arena/CLAUDE.md §3.
var catalog = map[Code]struct {
	Status  int
	Message string
}{
	CodeValidationFailed:    {http.StatusBadRequest, "Revisá los campos marcados."},
	CodeCodeNotFound:        {http.StatusNotFound, "Ese código no existe. Revisá cómo lo escribiste."},
	CodeCodeAlreadyRedeemed: {http.StatusConflict, "Ese código ya fue usado."},
	CodeUsernameTaken:       {http.StatusConflict, "Ese usuario ya está ocupado."},
	CodeInvalidCredentials:  {http.StatusUnauthorized, "Usuario o contraseña incorrectos."},
	CodeUnauthenticated:     {http.StatusUnauthorized, "Iniciá sesión para continuar."},
	CodeForbidden:           {http.StatusForbidden, "No tenés acceso a esto."},

	// Los dos de abajo NO están en la tabla de api.md. Se agregaron porque hacía
	// falta y quedan anotados como pendientes de agregar al contrato:
	//
	//   USER_NOT_FOUND — `POST /api/admin/users/:id/gift` con un id inexistente.
	//     Sin este código habría que devolver 500 por un error de tipeo del
	//     instructor, o mentirle con un 200.
	//   RATE_LIMITED — `/auth/login` y `/auth/check-code` están abiertos a
	//     internet. Un login sin tope es fuerza bruta gratis, y devolver
	//     INVALID_CREDENTIALS cuando en realidad se frenó por límite le mentiría
	//     al alumno que sí puso bien la contraseña.
	CodeUserNotFound: {http.StatusNotFound, "No encontramos a ese usuario."},
	CodeRateLimited:  {http.StatusTooManyRequests, "Demasiados intentos. Esperá un momento."},

	CodeRaceNotFound:        {http.StatusNotFound, "No encontramos esa carrera."},
	CodeRaceNotOpen:         {http.StatusConflict, "Esta carrera no está tomando apuestas."},
	CodeBetAlreadyPlaced:    {http.StatusConflict, "Ya apostaste en esta carrera."},
	CodeInsufficientBalance: {http.StatusConflict, "No te alcanza el saldo para esta apuesta."},
	CodeInvalidTransition:   {http.StatusConflict, "La carrera no puede pasar a ese estado."},
	CodeInternal:            {http.StatusInternalServerError, "Algo salió mal de nuestro lado."},
}

// Fault es un error del dominio que sabe cómo convertirse en respuesta HTTP.
type Fault struct {
	Code    Code
	Details map[string]any
}

func (f *Fault) Error() string { return string(f.Code) }

// Status y Message salen del catálogo. Un código fuera de catálogo se degrada a
// 500 en vez de responder con estado 0 — que es peor de depurar.
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
// docs/contract/api.md.
func KnownCodes() []Code {
	out := make([]Code, 0, len(catalog))
	for c := range catalog {
		out = append(out, c)
	}
	return out
}
