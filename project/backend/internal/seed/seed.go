// Package seed carga el dataset de docs/contract/seed/ y lo reubica en el
// tiempo para que la app siempre tenga algo que mostrar.
//
// Los archivos están embebidos en el binario: `project/` se publica al alumno
// sin `docs/`, así que el backend lleva su propia copia sincronizada por
// scripts/sync-contract.mjs. Un `go build` produce un ejecutable que corre
// solo, sin archivos al lado.
package seed

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/talentodh/hipodromo/internal/contract"
)

//go:embed data/seed/*.json data/samples/*.json data/fixtures/*.jsonl
var files embed.FS

// Anchor es el instante al que está anclado el seed: la largada de race_005.
// Todo el dataset se define relativo a este momento.
var Anchor = time.Date(2026, 3, 14, 12, 0, 0, 0, time.UTC)

// CycleLength es cada cuánto vuelve a empezar el programa de carreras. Con
// 120 minutos, dentro de una clase de 2 horas se ve una progresión normal y
// nunca se repite una carrera.
const CycleLength = 120 * time.Minute

// Data es el dataset ya cargado y reubicado en el tiempo.
type Data struct {
	Races   []contract.Race
	Users   []SeedUser
	Bets    []contract.Bet
	Results []contract.RaceResult

	// Program son las carreras que el servidor larga en vivo, con su
	// desplazamiento respecto del arranque del ciclo.
	Program []ProgramEntry
}

// SeedUser es un usuario del dataset. La contraseña de desarrollo es la misma
// para todos y está en docs/contract/README.md.
type SeedUser struct {
	contract.User
}

// ProgramEntry es una carrera del programa y cuándo larga dentro del ciclo.
type ProgramEntry struct {
	RaceID string
	Offset time.Duration
}

// Load lee el dataset embebido y lo reubica tomando `now` como referencia.
//
// Regla de rebase (docs/contract/README.md): a todas las fechas se les suma
// `now - Anchor`. Un `startsAt` de marzo de 2026 en una clase de agosto se ve
// viejo; con el desplazamiento, las carreras terminadas siempre acaban de
// terminar y las que vienen siempre están por venir.
func Load(now time.Time) (*Data, error) {
	var races []contract.Race
	var users []SeedUser
	var bets []contract.Bet
	var results []contract.RaceResult

	for _, item := range []struct {
		file string
		into any
	}{
		{"data/seed/races.json", &races},
		{"data/seed/users.json", &users},
		{"data/seed/bets.json", &bets},
		{"data/seed/results.json", &results},
	} {
		raw, err := files.ReadFile(item.file)
		if err != nil {
			return nil, fmt.Errorf("leyendo %s: %w", item.file, err)
		}
		if err := json.Unmarshal(raw, item.into); err != nil {
			return nil, fmt.Errorf("parseando %s: %w", item.file, err)
		}
	}

	offset := now.Sub(Anchor)

	for i := range races {
		shifted, err := shift(races[i].StartsAt, offset)
		if err != nil {
			return nil, fmt.Errorf("carrera %s: %w", races[i].ID, err)
		}
		races[i].StartsAt = shifted
	}
	for i := range bets {
		shifted, err := shift(bets[i].PlacedAt, offset)
		if err != nil {
			return nil, fmt.Errorf("apuesta %s: %w", bets[i].ID, err)
		}
		bets[i].PlacedAt = shifted
	}
	for i := range results {
		shifted, err := shift(results[i].FinishedAt, offset)
		if err != nil {
			return nil, fmt.Errorf("resultado %s: %w", results[i].RaceID, err)
		}
		results[i].FinishedAt = shifted
		// El campo Payouts se llena por usuario en cada request. En el
		// dataset viene vacío, no nulo: un `null` rompería el @for.
		if results[i].Payouts == nil {
			results[i].Payouts = []contract.Payout{}
		}
	}

	program, err := buildProgram(races)
	if err != nil {
		return nil, err
	}

	return &Data{Races: races, Users: users, Bets: bets, Results: results, Program: program}, nil
}

// buildProgram arma el calendario de las carreras que se corren en vivo. El
// desplazamiento de cada una sale del propio seed: race_005 larga en el ancla,
// race_006 ocho minutos después, y así. El programa no inventa horarios, los
// deriva del contrato.
func buildProgram(races []contract.Race) ([]ProgramEntry, error) {
	var program []ProgramEntry
	for _, race := range races {
		if race.Status == contract.StatusFinished {
			continue
		}
		// StartsAt ya está desplazado, así que se recupera el offset original
		// volviendo a parsear el valor del seed.
		raw, err := originalStartsAt(race.ID)
		if err != nil {
			return nil, err
		}
		program = append(program, ProgramEntry{RaceID: race.ID, Offset: raw.Sub(Anchor)})
	}
	sort.Slice(program, func(i, j int) bool { return program[i].Offset < program[j].Offset })

	if len(program) == 0 {
		return nil, fmt.Errorf("el seed no tiene ninguna carrera para correr en vivo")
	}
	if last := program[len(program)-1].Offset; last >= CycleLength {
		return nil, fmt.Errorf("la última carrera larga en %v, fuera del ciclo de %v", last, CycleLength)
	}
	return program, nil
}

// originalStartsAt vuelve al archivo embebido para leer la fecha sin desplazar.
func originalStartsAt(raceID string) (time.Time, error) {
	raw, err := files.ReadFile("data/seed/races.json")
	if err != nil {
		return time.Time{}, err
	}
	var races []contract.Race
	if err := json.Unmarshal(raw, &races); err != nil {
		return time.Time{}, err
	}
	for _, r := range races {
		if r.ID == raceID {
			return time.Parse(time.RFC3339, r.StartsAt)
		}
	}
	return time.Time{}, fmt.Errorf("carrera %s no está en el seed", raceID)
}

func shift(iso string, offset time.Duration) (string, error) {
	parsed, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return "", fmt.Errorf("fecha inválida %q: %w", iso, err)
	}
	return Format(parsed.Add(offset)), nil
}

// Format es el formato de fecha del contrato: ISO 8601 en UTC, sin fracción.
// Todas las fechas que salen de la API pasan por acá.
func Format(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05Z")
}

// Sample devuelve el contenido de un archivo de docs/contract/samples/.
// Lo usan los tests golden.
func Sample(name string) ([]byte, error) { return files.ReadFile("data/samples/" + name) }

// Fixture devuelve el contenido de un archivo de docs/contract/fixtures/.
func Fixture(name string) ([]byte, error) { return files.ReadFile("data/fixtures/" + name) }

// SeedFile devuelve el contenido crudo de un archivo del dataset.
func SeedFile(name string) ([]byte, error) { return files.ReadFile("data/seed/" + name) }
