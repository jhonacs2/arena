// Package sim simula una carrera.
//
// Implementa docs/contract/race-simulation.md. La otra implementación de la
// misma especificación está en scripts/lib/race-sim.mjs y alimenta el fixture
// y el mock del frontend.
//
// Que las dos produzcan la misma carrera es lo que permite VERIFICAR el punto
// 5 de la definición de terminado — "se ve igual contra el mock y contra el
// backend real" — en vez de prometerlo. El test golden de este paquete
// reproduce fixtures/race-ticks.jsonl tick por tick.
//
// Determinístico: mismos (raceID, runIndex, caballos) → misma carrera. Sin
// rand, sin reloj.
package sim

import (
	"math"
	"sort"
	"strconv"

	"github.com/talentodh/hipodromo/internal/contract"
)

const (
	TickHz           = 10
	CountdownSeconds = 60

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
//
// La implementación JavaScript recorre unidades UTF-16 con charCodeAt. Para
// los ids del contrato —todos ASCII— eso coincide byte a byte. Si algún día
// un id lleva un carácter no ASCII, las dos implementaciones divergen: por eso
// los ids son `race_005` y no `carrera_ñandú`.
func fnv1a(s string) uint32 {
	h := uint32(0x811c9dc5)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 0x01000193
	}
	return h
}

// mix32 es la mezcla final: sin ella, ids parecidos dan valores parecidos.
func mix32(h uint32) uint32 {
	h ^= h >> 15
	h *= 0x2545f491
	h ^= h >> 13
	return h
}

// rnd devuelve un valor en [0, 1) derivado de la clave completa.
func rnd(raceID string, runIndex int, horseID, salt string) float64 {
	key := raceID + "/" + strconv.Itoa(runIndex) + "/" + horseID + "/" + salt
	return float64(mix32(fnv1a(key))) / 4294967296.0
}

// ── Preparación ───────────────────────────────────────────────────────────

// Runner es un caballo con sus parámetros de carrera ya resueltos.
type Runner struct {
	HorseID    string
	Name       string
	Number     int
	Odds       float64
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

// Race es una carrera completa ya resuelta.
type Race struct {
	RaceID   string
	RunIndex int
	Duration float64
	Runners  []Runner
	Ticks    []Tick
	Podium   []contract.PodiumEntry
}

// Prepare resuelve tiempo de llegada, estilo y fase de cada caballo.
func Prepare(raceID string, runIndex int, horses []contract.Horse) []Runner {
	// La cuota más baja es el favorito. El empate se rompe por número de
	// partida, para que el resultado no dependa del orden del array.
	byOdds := make([]contract.Horse, len(horses))
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
		styleIndex := int(rnd(raceID, runIndex, h.ID, "s") * float64(len(Shapes)))
		if styleIndex >= len(Shapes) {
			styleIndex = len(Shapes) - 1
		}
		runners[i] = Runner{
			HorseID:    h.ID,
			Name:       h.Name,
			Number:     h.Number,
			Odds:       h.Odds,
			FinishTime: BaseDuration + OddsSpread*skill[h.ID] + Jitter*(rnd(raceID, runIndex, h.ID, "t")-0.5),
			Shape:      Shapes[styleIndex],
			Style:      StyleNames[styleIndex],
			Phase:      rnd(raceID, runIndex, h.ID, "p") * 2 * math.Pi,
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
// declaraba terminada sin que nadie cruzara. Pasaba en 1 de cada 4 corridas.
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
func Simulate(raceID string, runIndex int, horses []contract.Horse) Race {
	runners := Prepare(raceID, runIndex, horses)
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

		positions := make([]Position, len(raw))
		for j, p := range raw {
			positions[j] = Position{HorseID: p.HorseID, Progress: round3(p.Progress), Place: place[p.HorseID]}
		}
		ticks = append(ticks, Tick{T: t, Positions: positions})
	}

	return Race{
		RaceID:   raceID,
		RunIndex: runIndex,
		Duration: duration,
		Runners:  runners,
		Ticks:    ticks,
		Podium:   podium(runners, ticks),
	}
}

// podium son los tres primeros del último tick. Los demás quedan con progress
// menor a 1, que es lo que pasa en una carrera de verdad.
func podium(runners []Runner, ticks []Tick) []contract.PodiumEntry {
	if len(ticks) == 0 {
		return []contract.PodiumEntry{}
	}
	last := ticks[len(ticks)-1].Positions
	ordered := make([]Position, len(last))
	copy(ordered, last)
	sort.SliceStable(ordered, func(a, b int) bool { return ordered[a].Place < ordered[b].Place })

	byID := make(map[string]Runner, len(runners))
	for _, r := range runners {
		byID[r.HorseID] = r
	}

	n := min(3, len(ordered))
	out := make([]contract.PodiumEntry, 0, n)
	for i := 0; i < n; i++ {
		r := byID[ordered[i].HorseID]
		out = append(out, contract.PodiumEntry{
			Place:     i + 1,
			HorseID:   r.HorseID,
			HorseName: r.Name,
			Number:    r.Number,
			Odds:      r.Odds,
		})
	}
	return out
}
