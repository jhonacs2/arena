// Package sim simula una carrera.
//
// Portado de project/backend/internal/sim: mismas constantes, misma curva de
// esfuerzo, misma ondulación. Las dos diferencias son de contrato, no de
// modelo:
//
//   - la clave aleatoria es la SEMILLA de la carrera (races.seed) y no el par
//     (raceID, runIndex). En Arena la carrera la larga el instructor una vez, y
//     la semilla se guarda al largar para que el resultado sea reproducible: si
//     alguien reclama, se vuelve a correr igual;
//   - las cuotas entran como entero ×100, porque acá los montos son nota y no
//     hay float en ninguna parte — ver arena/CLAUDE.md §5.
//
// Determinístico: misma semilla y mismos caballos → misma carrera. Sin rand,
// sin reloj. El test golden de este paquete congela una corrida completa.
package sim

import (
	"math"
	"sort"
	"strconv"
)

const (
	TickHz = 10

	BaseDuration = 42.0 // segundos que tarda el favorito nominal
	OddsSpread   = 4.5  // cuánto más tarda, de base, el de cuota más alta
	Jitter       = 5.0  // ventana aleatoria; es lo que permite el batacazo
)

// Shapes es el exponente de la curva de esfuerzo por estilo.
var Shapes = [3]float64{0.82, 1.00, 1.22}

// StyleNames indexa igual que Shapes.
var StyleNames = [3]string{"front", "even", "closer"}

const (
	wobbleA1        = 0.0045
	wobbleF1        = 0.9
	wobbleA2        = 0.0022
	wobbleF2        = 2.3
	wobblePhaseMult = 2.7
)

// ── Aleatoriedad determinística ───────────────────────────────────────────

// fnv1a de 32 bits sobre los bytes de s.
func fnv1a(s string) uint32 {
	h := uint32(0x811c9dc5)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 0x01000193
	}
	return h
}

// mix32 es la mezcla final: sin ella, claves parecidas dan valores parecidos —
// y los ids de los caballos son UUID, que comparten mucho prefijo de formato.
func mix32(h uint32) uint32 {
	h ^= h >> 15
	h *= 0x2545f491
	h ^= h >> 13
	return h
}

// rnd devuelve un valor en [0, 1) derivado de la clave completa.
func rnd(seed int64, horseID, salt string) float64 {
	key := strconv.FormatInt(seed, 10) + "/" + horseID + "/" + salt
	return float64(mix32(fnv1a(key))) / 4294967296.0
}

// ── Entrada ───────────────────────────────────────────────────────────────

// Horse es lo que el simulador necesita de un caballo. Odds es la cuota ×100.
//
// El paquete no depende del dominio de carreras: recibe esta estructura mínima
// y devuelve ticks. Así el golden no se rompe cuando la tabla horses gana una
// columna.
type Horse struct {
	ID     string
	Name   string
	Number int
	Odds   int
}

// Runner es un caballo con sus parámetros de carrera ya resueltos.
type Runner struct {
	HorseID    string
	Name       string
	Number     int
	Odds       int
	FinishTime float64
	Shape      float64
	Style      string
	Phase      float64
}

// Position es la posición de un caballo en un instante.
type Position struct {
	HorseID  string  `json:"horseId"`
	Progress float64 `json:"progress"`
	Place    int     `json:"place"`
}

// Tick es un instante de la carrera. El orden de Positions es el de los
// caballos de la carrera, NO por puesto: ordenar es tarea del cliente.
type Tick struct {
	T         float64    `json:"t"`
	Positions []Position `json:"positions"`
}

// Result es el puesto final de un caballo.
//
// Están TODOS los caballos, no solo los tres primeros: la tabla race_results
// tiene unique (race_id, position) y se llena completa. El ganador es Position 1.
type Result struct {
	HorseID   string `json:"horseId"`
	HorseName string `json:"horseName"`
	Number    int    `json:"number"`
	Odds      int    `json:"odds"`
	Position  int    `json:"position"`
}

// Race es una carrera completa ya resuelta.
type Race struct {
	Seed     int64
	Duration float64
	Runners  []Runner
	Ticks    []Tick
	Results  []Result
}

// Winner es el caballo que llegó primero.
func (r Race) Winner() (Result, bool) {
	for _, res := range r.Results {
		if res.Position == 1 {
			return res, true
		}
	}
	return Result{}, false
}

// ── Preparación ───────────────────────────────────────────────────────────

// Prepare resuelve tiempo de llegada, estilo y fase de cada caballo.
func Prepare(seed int64, horses []Horse) []Runner {
	// La cuota más baja es el favorito. El empate se rompe por número de
	// partida, para que el resultado no dependa del orden del array.
	byOdds := make([]Horse, len(horses))
	copy(byOdds, horses)
	sort.SliceStable(byOdds, func(i, j int) bool {
		if byOdds[i].Odds != byOdds[j].Odds {
			return byOdds[i].Odds < byOdds[j].Odds
		}
		return byOdds[i].Number < byOdds[j].Number
	})

	skill := make(map[string]float64, len(horses))
	for i, h := range byOdds {
		if len(horses) > 1 {
			skill[h.ID] = float64(i) / float64(len(horses)-1)
		}
	}

	runners := make([]Runner, len(horses))
	for i, h := range horses {
		styleIndex := int(rnd(seed, h.ID, "s") * float64(len(Shapes)))
		if styleIndex >= len(Shapes) {
			styleIndex = len(Shapes) - 1
		}
		runners[i] = Runner{
			HorseID:    h.ID,
			Name:       h.Name,
			Number:     h.Number,
			Odds:       h.Odds,
			FinishTime: BaseDuration + OddsSpread*skill[h.ID] + Jitter*(rnd(seed, h.ID, "t")-0.5),
			Shape:      Shapes[styleIndex],
			Style:      StyleNames[styleIndex],
			Phase:      rnd(seed, h.ID, "p") * 2 * math.Pi,
		}
	}
	return runners
}

// ── Progreso ──────────────────────────────────────────────────────────────

func round1(n float64) float64 { return math.Round(n*10) / 10 }
func round3(n float64) float64 { return math.Round(n*1000) / 1000 }

func wobble(phase, t float64) float64 {
	return wobbleA1*math.Sin(phase+t*wobbleF1) + wobbleA2*math.Sin(phase*wobblePhaseMult+t*wobbleF2)
}

// Duration es cuánto dura la carrera: termina cuando cruza el primero, no
// cuando llegan todos.
//
// Redondea hacia ARRIBA a la décima, no al más cercano. Con redondeo al más
// cercano, un tiempo de 42.14 daba una duración de 42.1 y el último tick caía
// antes de la llegada: el ganador quedaba en progress 0.998 y la carrera se
// declaraba terminada sin que nadie cruzara.
func Duration(runners []Runner) float64 {
	best := runners[0].FinishTime
	for _, r := range runners[1:] {
		if r.FinishTime < best {
			best = r.FinishTime
		}
	}
	return math.Ceil(best*10) / 10
}

// Simulate resuelve la carrera completa.
func Simulate(seed int64, horses []Horse) Race {
	if len(horses) == 0 {
		return Race{Seed: seed, Ticks: []Tick{}, Results: []Result{}}
	}

	runners := Prepare(seed, horses)
	duration := Duration(runners)
	total := int(math.Round(duration * TickHz))

	previous := make(map[string]float64, len(runners))
	ticks := make([]Tick, 0, total)

	for i := 1; i <= total; i++ {
		t := round1(float64(i) / TickHz)

		raw := make([]Position, len(runners))
		for j, r := range runners {
			u := math.Min(t/r.FinishTime, 1)
			// La ondulación se apaga cerca del disco: nadie zigzaguea en la llegada.
			noisy := math.Pow(u, r.Shape) + wobble(r.Phase, t)*(1-u)
			progress := math.Min(math.Max(noisy, previous[r.HorseID]), 1)
			previous[r.HorseID] = progress
			raw[j] = Position{HorseID: r.HorseID, Progress: progress}
		}

		// Los puestos se calculan ANTES de redondear, sobre el valor exacto.
		place := places(raw)

		positions := make([]Position, len(raw))
		for j, p := range raw {
			positions[j] = Position{HorseID: p.HorseID, Progress: round3(p.Progress), Place: place[p.HorseID]}
		}
		ticks = append(ticks, Tick{T: t, Positions: positions})
	}

	return Race{
		Seed:     seed,
		Duration: duration,
		Runners:  runners,
		Ticks:    ticks,
		Results:  results(runners, ticks),
	}
}

// places ordena por progreso exacto y devuelve el puesto de cada caballo. El
// empate se rompe por id para que el puesto no dependa del orden del array.
func places(raw []Position) map[string]int {
	ordered := make([]Position, len(raw))
	copy(ordered, raw)
	sort.SliceStable(ordered, func(a, b int) bool {
		if ordered[a].Progress != ordered[b].Progress {
			return ordered[a].Progress > ordered[b].Progress
		}
		return ordered[a].HorseID < ordered[b].HorseID
	})

	place := make(map[string]int, len(ordered))
	for idx, p := range ordered {
		place[p.HorseID] = idx + 1
	}
	return place
}

// results son los puestos finales, tomados del último tick. Los que no llegaron
// quedan con progress menor a 1, que es lo que pasa en una carrera de verdad:
// el puesto igual está definido porque sale del orden de progreso.
func results(runners []Runner, ticks []Tick) []Result {
	if len(ticks) == 0 {
		return []Result{}
	}

	last := ticks[len(ticks)-1].Positions
	ordered := make([]Position, len(last))
	copy(ordered, last)
	sort.SliceStable(ordered, func(a, b int) bool { return ordered[a].Place < ordered[b].Place })

	byID := make(map[string]Runner, len(runners))
	for _, r := range runners {
		byID[r.HorseID] = r
	}

	out := make([]Result, 0, len(ordered))
	for i, p := range ordered {
		r := byID[p.HorseID]
		out = append(out, Result{
			HorseID:   r.HorseID,
			HorseName: r.Name,
			Number:    r.Number,
			Odds:      r.Odds,
			Position:  i + 1,
		})
	}
	return out
}
