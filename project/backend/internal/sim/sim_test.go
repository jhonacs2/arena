package sim_test

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/talentodh/hipodromo/internal/contract"
	"github.com/talentodh/hipodromo/internal/seed"
	"github.com/talentodh/hipodromo/internal/sim"
)

// El fixture es la grabación de race_005 corrida 164 — ver
// docs/contract/race-simulation.md.
const (
	fixtureRaceID   = "race_005"
	fixtureRunIndex = 164
)

// tolerance: math.Sin y math.Pow son IEEE-754 en Go y en JavaScript, pero no
// está garantizado que coincidan al último bit. Con el redondeo a 3 decimales
// una diferencia de 1 ULP no debería cambiar nada; si alguna vez cambia, este
// test es el que avisa.
const tolerance = 1e-9

type fixtureEvent struct {
	Type      string         `json:"type"`
	RaceID    string         `json:"raceId"`
	T         float64        `json:"t"`
	Positions []sim.Position `json:"positions"`
	Podium    []string       `json:"podium"`
}

func loadFixture(t *testing.T) []fixtureEvent {
	t.Helper()
	raw, err := seed.Fixture("race-ticks.jsonl")
	if err != nil {
		t.Fatalf("leyendo el fixture: %v", err)
	}
	var events []fixtureEvent
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		var e fixtureEvent
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatalf("parseando el fixture: %v", err)
		}
		events = append(events, e)
	}
	return events
}

func loadHorses(t *testing.T, raceID string) []contract.Horse {
	t.Helper()
	raw, err := seed.SeedFile("races.json")
	if err != nil {
		t.Fatalf("leyendo races.json: %v", err)
	}
	var races []contract.Race
	if err := json.Unmarshal(raw, &races); err != nil {
		t.Fatalf("parseando races.json: %v", err)
	}
	for _, r := range races {
		if r.ID == raceID {
			return r.Horses
		}
	}
	t.Fatalf("la carrera %s no está en el seed", raceID)
	return nil
}

// TestSimulateReproducesFixture es EL test de este paquete: prueba que la
// implementación Go y la JavaScript son la misma especificación.
func TestSimulateReproducesFixture(t *testing.T) {
	events := loadFixture(t)
	horses := loadHorses(t, fixtureRaceID)
	race := sim.Simulate(fixtureRaceID, fixtureRunIndex, horses)

	var expected []fixtureEvent
	for _, e := range events {
		if e.Type == "race.tick" {
			expected = append(expected, e)
		}
	}

	if len(race.Ticks) != len(expected) {
		t.Fatalf("cantidad de ticks: Go generó %d, el fixture tiene %d", len(race.Ticks), len(expected))
	}

	for i, want := range expected {
		got := race.Ticks[i]

		if math.Abs(got.T-want.T) > tolerance {
			t.Fatalf("tick %d: t = %v, el fixture dice %v", i, got.T, want.T)
		}
		if len(got.Positions) != len(want.Positions) {
			t.Fatalf("tick %d: %d posiciones, el fixture tiene %d", i, len(got.Positions), len(want.Positions))
		}

		for j := range want.Positions {
			w, g := want.Positions[j], got.Positions[j]

			// El orden del array también es parte del contrato: sale en el
			// orden de los caballos de la carrera, no por puesto.
			if g.HorseID != w.HorseID {
				t.Fatalf("tick %d pos %d: horseId %q, el fixture dice %q", i, j, g.HorseID, w.HorseID)
			}
			if math.Abs(g.Progress-w.Progress) > tolerance {
				t.Errorf("tick %d (t=%v) %s: progress %v, el fixture dice %v (Δ %.2e)",
					i, want.T, w.HorseID, g.Progress, w.Progress, math.Abs(g.Progress-w.Progress))
			}
			// Los puestos tienen que coincidir EXACTAMENTE: son enteros, no
			// hay tolerancia posible y son lo que ve el alumno en pantalla.
			if g.Place != w.Place {
				t.Errorf("tick %d (t=%v) %s: puesto %d, el fixture dice %d", i, want.T, w.HorseID, g.Place, w.Place)
			}
		}
		if t.Failed() {
			t.Fatalf("divergencia en el tick %d — no sigo comparando", i)
		}
	}
}

func TestPodiumMatchesFixture(t *testing.T) {
	events := loadFixture(t)
	horses := loadHorses(t, fixtureRaceID)
	race := sim.Simulate(fixtureRaceID, fixtureRunIndex, horses)

	var want []string
	for _, e := range events {
		if e.Type == "race.finished" {
			want = e.Podium
		}
	}
	if want == nil {
		t.Fatal("el fixture no tiene un evento race.finished")
	}

	if len(race.Podium) != len(want) {
		t.Fatalf("podio de %d, el fixture tiene %d", len(race.Podium), len(want))
	}
	for i, id := range want {
		if race.Podium[i].HorseID != id {
			t.Errorf("puesto %d: %s, el fixture dice %s", i+1, race.Podium[i].HorseID, id)
		}
		if race.Podium[i].Place != i+1 {
			t.Errorf("puesto %d: el campo Place dice %d", i+1, race.Podium[i].Place)
		}
	}
}

// TestDeterminista: la misma corrida siempre da lo mismo, y corridas distintas
// dan carreras distintas. Sin lo primero el fixture no serviría; sin lo segundo
// todas las carreras del programa serían la misma.
func TestDeterminista(t *testing.T) {
	horses := loadHorses(t, fixtureRaceID)

	a := sim.Simulate(fixtureRaceID, 7, horses)
	b := sim.Simulate(fixtureRaceID, 7, horses)
	if a.Duration != b.Duration || a.Podium[0].HorseID != b.Podium[0].HorseID {
		t.Error("la misma corrida dio dos carreras distintas")
	}

	c := sim.Simulate(fixtureRaceID, 8, horses)
	if a.Duration == c.Duration && a.Podium[0].HorseID == c.Podium[0].HorseID {
		t.Error("dos corridas distintas dieron exactamente la misma carrera")
	}
}

// TestProgresoMonotono: un caballo nunca retrocede y nunca pasa de 1. Si esto
// falla, en pantalla se ve un caballo yendo para atrás.
func TestProgresoMonotono(t *testing.T) {
	horses := loadHorses(t, fixtureRaceID)

	for run := 0; run < 25; run++ {
		race := sim.Simulate(fixtureRaceID, run, horses)
		last := map[string]float64{}

		for _, tick := range race.Ticks {
			seen := map[int]bool{}
			for _, p := range tick.Positions {
				if p.Progress < last[p.HorseID] {
					t.Fatalf("corrida %d, t=%v: %s retrocedió de %v a %v", run, tick.T, p.HorseID, last[p.HorseID], p.Progress)
				}
				if p.Progress < 0 || p.Progress > 1 {
					t.Fatalf("corrida %d, t=%v: %s tiene progress %v", run, tick.T, p.HorseID, p.Progress)
				}
				last[p.HorseID] = p.Progress

				if seen[p.Place] {
					t.Fatalf("corrida %d, t=%v: el puesto %d está repetido", run, tick.T, p.Place)
				}
				seen[p.Place] = true
			}
		}

		if got := race.Ticks[len(race.Ticks)-1]; !hasWinnerAtGoal(got.Positions) {
			t.Errorf("corrida %d: al terminar, nadie llegó a progress 1", run)
		}
	}
}

func hasWinnerAtGoal(positions []sim.Position) bool {
	for _, p := range positions {
		if p.Place == 1 && p.Progress >= 1 {
			return true
		}
	}
	return false
}

// TestFavoritoGanaLaMayoria: con las constantes actuales el favorito gana
// alrededor de la mitad. Suficiente para que apostarle sea razonable, y no
// tanto como para que sea aburrido. Si alguien toca ODDS_SPREAD o JITTER, este
// test dice si se rompió el equilibrio.
func TestFavoritoGanaLaMayoria(t *testing.T) {
	horses := loadHorses(t, fixtureRaceID)
	favourite := contract.Race{Horses: horses}
	fav, _ := favourite.Favourite()

	const runs = 200
	wins := 0
	for run := 0; run < runs; run++ {
		if sim.Simulate(fixtureRaceID, run, horses).Podium[0].HorseID == fav.ID {
			wins++
		}
	}

	rate := float64(wins) / runs
	if rate < 0.35 || rate > 0.65 {
		t.Errorf("el favorito ganó %.0f%% de las corridas; se espera entre 35%% y 65%%", rate*100)
	}
	t.Logf("el favorito (%s @%.2f) ganó %d de %d corridas (%.0f%%)", fav.Name, fav.Odds, wins, runs, rate*100)
}
