package races

// La máquina de estados de decisiones.md §3:
//
//	draft ──▶ open ──▶ running ──▶ finished
//	  │         │         │
//	  └─────────┴─────────┴──▶ cancelled
//
// Es una tabla y no una cadena de `if`: con la tabla, agregar un estado obliga a
// decidir explícitamente a dónde puede ir, y el test recorre la grilla completa
// de 5×5 y verifica que todo lo que no esté acá se rechace.
var transitions = map[Status][]Status{
	StatusDraft:     {StatusOpen, StatusCancelled},
	StatusOpen:      {StatusRunning, StatusCancelled},
	StatusRunning:   {StatusFinished, StatusCancelled},
	StatusFinished:  {},
	StatusCancelled: {},
}

// AllStatuses es la grilla completa. La usa el test que recorre todas las
// combinaciones posibles.
var AllStatuses = []Status{StatusDraft, StatusOpen, StatusRunning, StatusFinished, StatusCancelled}

// CanTransition dice si el paso está permitido. Un estado desconocido no
// transiciona a ninguna parte: si alguien inventa un valor, el resultado es «no»
// y no un pánico.
func CanTransition(from, to Status) bool {
	for _, allowed := range transitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

// Visible dice si un alumno puede ver una carrera en este estado.
//
// `draft` NO se ve: es la carrera que el instructor todavía está armando. Es la
// única regla de visibilidad del contrato, y está acá y no repetida en cada
// consulta para que haya un solo lugar donde pueda estar mal.
func (s Status) Visible() bool {
	return s == StatusOpen || s == StatusRunning || s == StatusFinished
}

// AcceptsBets dice si se puede apostar. Solo `open`.
//
// Las apuestas se cierran EN EL SERVIDOR al pasar a `running`: el botón
// deshabilitado del frontend es una cortesía, no un control.
func (s Status) AcceptsBets() bool { return s == StatusOpen }

// RevealsBets dice si ya se puede mostrar a qué caballo apostó cada uno.
//
// Mientras la carrera está `open` NO se revela: si se revelara, los últimos en
// apostar copiarían a los primeros y la apuesta dejaría de medir criterio. Al
// pasar a `running` se revelan todas juntas.
func (s Status) RevealsBets() bool { return s != StatusOpen && s != StatusDraft }

// Editable dice si el instructor todavía puede cambiarle el nombre, la fecha o
// los caballos. Solo en `draft`: una vez que los alumnos la vieron y apostaron,
// mover una cuota sería mover la nota de alguien.
func (s Status) Editable() bool { return s == StatusDraft }
