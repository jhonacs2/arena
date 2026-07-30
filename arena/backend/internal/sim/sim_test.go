package sim_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/talentodh/arena/internal/sim"
)

// -update reescribe el golden. Se corre a mano cuando se cambia el modelo A
// PROPÓSITO, y el diff del archivo entra al commit para que se vea qué cambió:
//
//	go test ./internal/sim -run TestGolden -update
var update = flag.Bool("update", false, "reescribe el golden de la simulación")

// La corrida congelada. La semilla es arbitraria pero fija: lo que importa es
// que la MISMA semilla dé siempre la misma carrera, porque eso es lo que
// permite volver a correr una carrera si un alumno reclama el resultado.
const goldenSeed int64 = 4815162342

// goldenHorses son ids con forma de UUID a propósito: los ids reales salen de
// gen_random_uuid() y comparten mucho prefijo de formato. Si el mezclador de
// bits no alcanzara, se vería acá y no en clase.
var goldenHorses = []sim.Horse{
	{ID: "0f1a2b3c-0000-4000-8000-000000000001", Name: "Viento Norte", Number: 1, Odds: 340},
	{ID: "0f1a2b3c-0000-4000-8000-000000000002", Name: "Doña Rosa", Number: 2, Odds: 520},
	{ID: "0f1a2b3c-0000-4000-8000-000000000003", Name: "Malambo", Number: 3, Odds: 210},
	{ID: "0f1a2b3c-0000-4000-8000-000000000004", Name: "Cimarrón", Number: 4, Odds: 780},
	{ID: "0f1a2b3c-0000-4000-8000-000000000005", Name: "Tarde Gris", Number: 5, Odds: 1150},
	{ID: "0f1a2b3c-0000-4000-8000-000000000006", Name: "Pampa", Number: 6, Odds: 430},
}

// golden es lo que se congela: la carrera entera, tick por tick, más los
// puestos finales.
type golden struct {
	Seed     int64        `json:"seed"`
	Duration float64      `json:"duration"`
	Ticks    []sim.Tick   `json:"ticks"`
	Results  []sim.Result `json:"results"`
}

func goldenPath() string { return filepath.Join("testdata", "race.golden.json") }

// TestGolden es EL test de este paquete: si el modelo cambia sin que nadie lo
// haya querido, acá se ve. Una carrera que cambia de resultado en silencio es
// una nota que cambia en silencio.
func TestGolden(t *testing.T) {
	race := sim.Simulate(goldenSeed, goldenHorses)
	got := golden{Seed: race.Seed, Duration: race.Duration, Ticks: race.Ticks, Results: race.Results}

	encoded, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		t.Fatalf("serializando la carrera: %v", err)
	}
	encoded = append(encoded, '\n')

	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatalf("creando testdata: %v", err)
		}
		if err := os.WriteFile(goldenPath(), encoded, 0o644); err != nil {
			t.Fatalf("escribiendo el golden: %v", err)
		}
		t.Logf("golden reescrito: %d ticks, duración %.1fs", len(race.Ticks), race.Duration)
		return
	}

	want, err := os.ReadFile(goldenPath())
	if err != nil {
		t.Fatalf("leyendo el golden (corré `go test ./internal/sim -run TestGolden -update`): %v", err)
	}

	if bytes.Equal(bytes.ReplaceAll(want, []byte("\r\n"), []byte("\n")), encoded) {
		return
	}

	// Comparar los bytes da un diff enorme e ilegible. Cuando no coinciden, se
	// parsea y se informa la PRIMERA divergencia: es la que explica el resto.
	var expected golden
	if err := json.Unmarshal(want, &expected); err != nil {
		t.Fatalf("el golden no es un JSON válido: %v", err)
	}
	reportFirstDivergence(t, expected, got)
	t.Fatal("la simulación ya no reproduce el golden")
}

func reportFirstDivergence(t *testing.T, want, got golden) {
	t.Helper()

	if want.Duration != got.Duration {
		t.Errorf("duración: %v, el golden dice %v", got.Duration, want.Duration)
	}
	if len(want.Ticks) != len(got.Ticks) {
		t.Errorf("cantidad de ticks: %d, el golden tiene %d", len(got.Ticks), len(want.Ticks))
	}

	for i := 0; i < min(len(want.Ticks), len(got.Ticks)); i++ {
		w, g := want.Ticks[i], got.Ticks[i]
		if w.T != g.T {
			t.Errorf("tick %d: t = %v, el golden dice %v", i, g.T, w.T)
			return
		}
		for j := range w.Positions {
			wp, gp := w.Positions[j], g.Positions[j]
			// El orden del array también es parte del contrato: sale en el orden
			// de los caballos de la carrera, no por puesto.
			if wp.HorseID != gp.HorseID || wp.Progress != gp.Progress || wp.Place != gp.Place {
				t.Errorf("tick %d (t=%v) posición %d: %+v, el golden dice %+v", i, w.T, j, gp, wp)
				return
			}
		}
	}

	for i := range want.Results {
		if i >= len(got.Results) || want.Results[i] != got.Results[i] {
			t.Errorf("resultado %d: %+v, el golden dice %+v", i, got.Results[i], want.Results[i])
			return
		}
	}
}

// TestDeterministic: la misma semilla siempre da lo mismo, y semillas distintas
// dan carreras distintas. Sin lo primero el golden no serviría y no se podría
// reproducir una carrera reclamada; sin lo segundo todas las carreras serían la
// misma.
func TestDeterministic(t *testing.T) {
	a := sim.Simulate(goldenSeed, goldenHorses)
	b := sim.Simulate(goldenSeed, goldenHorses)

	winnerA, _ := a.Winner()
	winnerB, _ := b.Winner()
	if a.Duration != b.Duration || winnerA.HorseID != winnerB.HorseID {
		t.Error("la misma semilla dio dos carreras distintas")
	}

	c := sim.Simulate(goldenSeed+1, goldenHorses)
	winnerC, _ := c.Winner()
	if a.Duration == c.Duration && winnerA.HorseID == winnerC.HorseID {
		t.Error("dos semillas distintas dieron exactamente la misma carrera")
	}
}

// TestResultsCubrenTodosLosCaballos: race_results tiene unique (race_id,
// position) y se llena con TODOS los caballos. Si el simulador repitiera un
// puesto o se dejara uno afuera, el insert de la liquidación explotaría en
// producción y la carrera quedaría sin liquidar.
func TestResultsCubrenTodosLosCaballos(t *testing.T) {
	for seed := int64(0); seed < 50; seed++ {
		race := sim.Simulate(seed, goldenHorses)

		if len(race.Results) != len(goldenHorses) {
			t.Fatalf("semilla %d: %d resultados para %d caballos", seed, len(race.Results), len(goldenHorses))
		}

		positions := map[int]bool{}
		horses := map[string]bool{}
		for _, r := range race.Results {
			if positions[r.Position] {
				t.Fatalf("semilla %d: el puesto %d está repetido", seed, r.Position)
			}
			if horses[r.HorseID] {
				t.Fatalf("semilla %d: %s aparece dos veces", seed, r.HorseID)
			}
			positions[r.Position] = true
			horses[r.HorseID] = true
		}
		for p := 1; p <= len(goldenHorses); p++ {
			if !positions[p] {
				t.Fatalf("semilla %d: falta el puesto %d", seed, p)
			}
		}
		if _, ok := race.Winner(); !ok {
			t.Fatalf("semilla %d: no hay ganador", seed)
		}
	}
}

// TestProgressIsMonotonic: un caballo nunca retrocede y nunca pasa de 1. Si esto
// falla, en pantalla se ve un caballo yendo para atrás.
func TestProgressIsMonotonic(t *testing.T) {
	for seed := int64(0); seed < 25; seed++ {
		race := sim.Simulate(seed, goldenHorses)
		last := map[string]float64{}

		for _, tick := range race.Ticks {
			seen := map[int]bool{}
			for _, p := range tick.Positions {
				if p.Progress < last[p.HorseID] {
					t.Fatalf("semilla %d, t=%v: %s retrocedió de %v a %v", seed, tick.T, p.HorseID, last[p.HorseID], p.Progress)
				}
				if p.Progress < 0 || p.Progress > 1 {
					t.Fatalf("semilla %d, t=%v: %s tiene progress %v", seed, tick.T, p.HorseID, p.Progress)
				}
				last[p.HorseID] = p.Progress

				if seen[p.Place] {
					t.Fatalf("semilla %d, t=%v: el puesto %d está repetido", seed, tick.T, p.Place)
				}
				seen[p.Place] = true
			}
		}

		if !hasWinnerAtGoal(race.Ticks[len(race.Ticks)-1].Positions) {
			t.Errorf("semilla %d: al terminar, nadie llegó a progress 1", seed)
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

// TestFavouriteWinsAboutHalf: con las constantes actuales el favorito gana
// alrededor de la mitad. Suficiente para que apostarle sea razonable, y no
// tanto como para que sea aburrido. Si alguien toca OddsSpread o Jitter, este
// test dice si se rompió el equilibrio.
func TestFavouriteWinsAboutHalf(t *testing.T) {
	favourite := goldenHorses[0]
	for _, h := range goldenHorses[1:] {
		if h.Odds < favourite.Odds || (h.Odds == favourite.Odds && h.Number < favourite.Number) {
			favourite = h
		}
	}

	const runs = 200
	wins := 0
	for seed := int64(0); seed < runs; seed++ {
		if winner, _ := sim.Simulate(seed, goldenHorses).Winner(); winner.HorseID == favourite.ID {
			wins++
		}
	}

	rate := float64(wins) / runs
	if rate < 0.35 || rate > 0.65 {
		t.Errorf("el favorito ganó %.0f%% de las corridas; se espera entre 35%% y 65%%", rate*100)
	}
	t.Logf("el favorito (%s @%d) ganó %d de %d corridas (%.0f%%)", favourite.Name, favourite.Odds, wins, runs, rate*100)
}

// TestUnSoloCaballo: no es una carrera interesante, pero el instructor puede
// abrir una con un caballo por error y no tiene que tirar el servidor abajo.
func TestUnSoloCaballo(t *testing.T) {
	race := sim.Simulate(7, goldenHorses[:1])
	if len(race.Results) != 1 || race.Results[0].Position != 1 {
		t.Fatalf("con un caballo, los resultados son %+v", race.Results)
	}
	if math.Abs(race.Duration-sim.BaseDuration) > sim.Jitter {
		t.Errorf("duración %v fuera de la ventana esperada", race.Duration)
	}
}

// TestSinCaballos: una carrera vacía devuelve una carrera vacía, no un pánico.
func TestSinCaballos(t *testing.T) {
	race := sim.Simulate(7, nil)
	if len(race.Ticks) != 0 || len(race.Results) != 0 {
		t.Fatalf("una carrera sin caballos devolvió %d ticks y %d resultados", len(race.Ticks), len(race.Results))
	}
	if _, ok := race.Winner(); ok {
		t.Error("una carrera sin caballos tiene ganador")
	}
}
