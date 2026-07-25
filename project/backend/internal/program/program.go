// Package program corre el calendario de carreras.
//
// El servidor no espera a que nadie pida nada: larga carreras solo, en ciclo,
// para que la app siempre tenga algo en vivo. Es lo que hace que una clase de
// dos horas vea la progresión completa —por venir, en vivo, terminada— sin que
// el instructor tenga que disparar nada a mano.
//
// Los horarios salen del propio seed: race_005 larga al empezar el ciclo,
// race_006 ocho minutos después, y así. El programa no inventa horarios.
package program

import (
	"context"
	"log/slog"
	"math"
	"time"

	"github.com/talentodh/hipodromo/internal/contract"
	"github.com/talentodh/hipodromo/internal/seed"
	"github.com/talentodh/hipodromo/internal/sim"
	"github.com/talentodh/hipodromo/internal/store"
	"github.com/talentodh/hipodromo/internal/ws"
)

// FirstDelay es cuánto espera el servidor antes de largar la primera carrera.
// Corto a propósito: al arrancar en desarrollo se quiere una carrera en vivo
// enseguida, no dentro de ocho minutos.
const FirstDelay = 30 * time.Second

type Program struct {
	Store   *store.Store
	Hub     *ws.Hub
	Log     *slog.Logger
	Entries []seed.ProgramEntry
	Clock   func() time.Time
}

func New(st *store.Store, hub *ws.Hub, log *slog.Logger, entries []seed.ProgramEntry) *Program {
	return &Program{Store: st, Hub: hub, Log: log, Entries: entries, Clock: time.Now}
}

// Run corre el calendario hasta que se cancele el contexto.
func (p *Program) Run(ctx context.Context) {
	cycleStart := p.Clock().Add(FirstDelay)

	for {
		// Todas las carreras del ciclo se publican de entrada, así el listado
		// muestra el programa completo desde el primer momento.
		for _, entry := range p.Entries {
			p.Store.ScheduleRace(entry.RaceID, cycleStart.Add(entry.Offset))
		}
		p.Log.Info("programa publicado",
			"carreras", len(p.Entries),
			"primera", seed.Format(cycleStart),
			"ciclo", seed.CycleLength)

		for _, entry := range p.Entries {
			if err := p.runRace(ctx, entry.RaceID, cycleStart.Add(entry.Offset)); err != nil {
				return // contexto cancelado: el servidor se está apagando
			}
		}
		cycleStart = cycleStart.Add(seed.CycleLength)
	}
}

// runRace espera la largada, corre la carrera y la liquida.
func (p *Program) runRace(ctx context.Context, raceID string, startsAt time.Time) error {
	race, ok := p.Store.Race(raceID)
	if !ok {
		p.Log.Error("el programa apunta a una carrera que no existe", "carrera", raceID)
		return nil
	}

	if err := p.countdown(ctx, raceID, startsAt); err != nil {
		return err
	}

	runIndex, _ := p.Store.ScheduleRace(raceID, startsAt)
	result := sim.Simulate(raceID, runIndex, race.Horses)

	startedAt := p.Clock()
	p.Store.SetRaceStatus(raceID, contract.StatusLive)
	p.Hub.ToRoom(raceID, ws.NewStarted(raceID, seed.Format(startedAt)))
	p.Log.Info("largó", "carrera", raceID, "corrida", runIndex, "duración", result.Duration)

	if err := p.emitTicks(ctx, raceID, startedAt, result); err != nil {
		return err
	}

	p.settle(raceID, result)
	return nil
}

// countdown emite `race.countdown` a 1 Hz durante el último minuto.
func (p *Program) countdown(ctx context.Context, raceID string, startsAt time.Time) error {
	for {
		remaining := startsAt.Sub(p.Clock())
		if remaining <= 0 {
			return nil
		}

		secondsLeft := int(math.Ceil(remaining.Seconds()))
		if secondsLeft <= sim.CountdownSeconds {
			p.Hub.ToRoom(raceID, ws.NewCountdown(raceID, secondsLeft))
		}

		// Dormir hasta el próximo segundo exacto, no un segundo entero: sin
		// esto la cuenta se desfasa y salta números.
		wait := remaining % time.Second
		if wait == 0 {
			wait = time.Second
		}
		if remaining > sim.CountdownSeconds*time.Second {
			wait = remaining - sim.CountdownSeconds*time.Second
		}

		if err := sleep(ctx, wait); err != nil {
			return err
		}
	}
}

// emitTicks manda los ticks a 10 Hz.
//
// Cada tick se agenda contra un instante ABSOLUTO desde la largada, no con un
// ticker relativo: a 10 Hz durante 40 segundos, un error de pocos milisegundos
// por tick se acumularía hasta desfasar la carrera casi un segundo.
func (p *Program) emitTicks(ctx context.Context, raceID string, startedAt time.Time, result sim.Race) error {
	for _, tick := range result.Ticks {
		target := startedAt.Add(time.Duration(tick.T * float64(time.Second)))
		if err := sleep(ctx, time.Until(target)); err != nil {
			return err
		}

		positions := make([]ws.Position, len(tick.Positions))
		for i, pos := range tick.Positions {
			positions[i] = ws.Position{HorseID: pos.HorseID, Progress: pos.Progress, Place: pos.Place}
		}
		p.Hub.ToRoom(raceID, ws.NewTick(raceID, tick.T, positions))
	}
	return nil
}

// settle liquida las apuestas y avisa a quien corresponda.
func (p *Program) settle(raceID string, result sim.Race) {
	finishedAt := p.Clock()
	settlements := p.Store.SettleRace(raceID, result.Podium, finishedAt)

	podium := make([]string, len(result.Podium))
	for i, entry := range result.Podium {
		podium[i] = entry.HorseID
	}

	// `payouts` de race.finished trae solo las apuestas de quien recibe el
	// evento: difundir el mismo objeto a todos filtraría lo que cobró cada uno.
	byUser := map[string][]ws.FinishedPayout{}
	for _, s := range settlements {
		for _, payout := range s.Payouts {
			byUser[s.UserID] = append(byUser[s.UserID], ws.FinishedPayout{BetID: payout.BetID, Amount: payout.Amount})
		}
	}
	p.Hub.ToRoomPerUser(raceID, func(userID string) any {
		return ws.NewFinished(raceID, podium, byUser[userID])
	})

	for _, s := range settlements {
		p.Hub.ToUser(s.UserID, ws.NewBalanceUpdated(s.Balance))
	}

	p.Hub.ToAll(ws.NewLeaderboardUpdated(p.Store.Leaderboard("all", finishedAt)))

	p.Log.Info("llegó",
		"carrera", raceID,
		"ganador", result.Podium[0].HorseName,
		"cuota", result.Podium[0].Odds,
		"liquidadas", len(settlements))
}

// sleep espera respetando la cancelación del contexto.
func sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
