package races

import "testing"

// TestCanTransition recorre LA GRILLA COMPLETA de 5×5 y verifica que solo las
// cinco transiciones del contrato estén permitidas.
//
// Es una tabla exhaustiva y no una lista de casos sueltos a propósito: con una
// lista, agregar un estado no rompe nada y las transiciones nuevas quedan sin
// probar. Con la grilla, cualquier combinación que no esté declarada acá abajo
// tiene que dar false, y si alguien la habilita el test falla.
func TestCanTransition(t *testing.T) {
	// Las únicas permitidas, de decisiones.md §3.
	allowed := map[Status]map[Status]bool{
		StatusDraft:   {StatusOpen: true, StatusCancelled: true},
		StatusOpen:    {StatusRunning: true, StatusCancelled: true},
		StatusRunning: {StatusFinished: true, StatusCancelled: true},
		// Terminal: de acá no se sale. Una carrera liquidada no se puede
		// reabrir, porque eso sería mover la nota de alguien.
		StatusFinished:  {},
		StatusCancelled: {},
	}

	for _, from := range AllStatuses {
		for _, to := range AllStatuses {
			want := allowed[from][to]
			if got := CanTransition(from, to); got != want {
				t.Errorf("CanTransition(%q, %q) = %v; se esperaba %v", from, to, got, want)
			}
		}
	}
}

// TestNingunEstadoTransicionaASiMismo: volver a abrir una carrera ya abierta no es
// una operación válida. Si lo fuera, `open` volvería a sellar opened_at y se
// perdería cuándo se abrió de verdad.
func TestNingunEstadoTransicionaASiMismo(t *testing.T) {
	for _, status := range AllStatuses {
		if CanTransition(status, status) {
			t.Errorf("%q transiciona a sí mismo", status)
		}
	}
}

// TestEstadoDesconocidoNoTransiciona: si por un error de datos aparece un estado
// que no existe, la respuesta es «no» y no un pánico.
func TestEstadoDesconocidoNoTransiciona(t *testing.T) {
	fake := Status("caminando")
	for _, to := range AllStatuses {
		if CanTransition(fake, to) {
			t.Errorf("un estado inventado transiciona a %q", to)
		}
		if CanTransition(to, fake) {
			t.Errorf("%q transiciona a un estado inventado", to)
		}
	}
}

// TestReglasDeEstado fija en un test las cuatro preguntas que el resto del
// paquete le hace al estado. Son las que deciden quién ve qué y quién puede
// apostar, así que se prueban explícitamente y no de rebote.
func TestReglasDeEstado(t *testing.T) {
	cases := []struct {
		status                                  Status
		visible, acceptsBets, revealsBets, edit bool
	}{
		// draft: no se ve, no se apuesta, no se revela nada, se edita.
		{StatusDraft, false, false, false, true},
		// open: se ve y se apuesta, pero NO se revela a qué caballo apostó cada
		// uno — si se revelara, los últimos copiarían a los primeros.
		{StatusOpen, true, true, false, false},
		// running: se ve, ya no se apuesta, y las apuestas se revelan todas
		// juntas.
		{StatusRunning, true, false, true, false},
		{StatusFinished, true, false, true, false},
		// cancelled: no aparece en el listado del alumno; las apuestas ya se
		// devolvieron.
		{StatusCancelled, false, false, true, false},
	}

	for _, c := range cases {
		if got := c.status.Visible(); got != c.visible {
			t.Errorf("%q.Visible() = %v; se esperaba %v", c.status, got, c.visible)
		}
		if got := c.status.AcceptsBets(); got != c.acceptsBets {
			t.Errorf("%q.AcceptsBets() = %v; se esperaba %v", c.status, got, c.acceptsBets)
		}
		if got := c.status.RevealsBets(); got != c.revealsBets {
			t.Errorf("%q.RevealsBets() = %v; se esperaba %v", c.status, got, c.revealsBets)
		}
		if got := c.status.Editable(); got != c.edit {
			t.Errorf("%q.Editable() = %v; se esperaba %v", c.status, got, c.edit)
		}
	}
}

// TestSoloOpenAceptaApuestas es la regla en su forma más corta: exactamente un
// estado acepta apuestas. Si alguna vez alguien agrega otro, este test lo dice.
func TestSoloOpenAceptaApuestas(t *testing.T) {
	accepting := []Status{}
	for _, status := range AllStatuses {
		if status.AcceptsBets() {
			accepting = append(accepting, status)
		}
	}
	if len(accepting) != 1 || accepting[0] != StatusOpen {
		t.Errorf("los estados que aceptan apuestas son %v; tiene que ser solo [open]", accepting)
	}
}
