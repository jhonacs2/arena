package races

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/talentodh/arena/internal/sim"
)

// Runner corre las simulaciones en vivo: una goroutine por carrera que emite los
// ticks a 10 Hz y, al llegar, liquida.
//
// La simulación es AUTORITATIVA DEL SERVIDOR. El cliente dibuja lo que recibe;
// no calcula nada. Si el cliente calculara, un alumno podría hacerse ganar en su
// pantalla — y en Arena la pantalla es la nota.
type Runner struct {
	log   *slog.Logger
	clock func() time.Time
	hub   Broadcaster

	// finish liquida la carrera. Es una clausura y no una dependencia para que
	// el runner no tenga que conocer el servicio que lo creó.
	finish func(ctx context.Context, raceID string, result sim.Race) error

	mu      sync.Mutex
	running map[string]context.CancelFunc
	wg      sync.WaitGroup
	closed  bool
}

func newRunner(log *slog.Logger, clock func() time.Time, hub Broadcaster, finish func(context.Context, string, sim.Race) error) *Runner {
	return &Runner{
		log:     log,
		clock:   clock,
		hub:     hub,
		finish:  finish,
		running: map[string]context.CancelFunc{},
	}
}

// Start larga la carrera en una goroutine propia.
//
// startedAt puede estar en el pasado: es lo que pasa cuando el servidor se
// reinició con una carrera corriendo. En ese caso los ticks que ya vencieron se
// SALTEAN en vez de dispararse todos de golpe, y la carrera se retoma en el
// punto donde debería estar. Es la razón por la que la semilla se guarda.
func (r *Runner) Start(raceID string, result sim.Race, startedAt time.Time) {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return
	}
	// Una carrera se larga una sola vez. La transición está protegida por el
	// bloqueo de fila, pero si por algún camino se llamara dos veces, la segunda
	// no duplica los ticks.
	if _, already := r.running[raceID]; already {
		r.mu.Unlock()
		r.log.Warn("se pidió largar una carrera que ya está corriendo", "carrera", raceID)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	r.running[raceID] = cancel
	r.wg.Add(1)
	r.mu.Unlock()

	go func() {
		defer r.wg.Done()
		defer r.forget(raceID)

		if !r.emitTicks(ctx, raceID, startedAt, result) {
			// Cancelada, o el servidor se está apagando. NO se liquida: la
			// carrera queda en `running` y la retoma Resume en el próximo
			// arranque, o ya la puso en `cancelled` quien la canceló.
			return
		}

		if err := r.finish(ctx, raceID, result); err != nil {
			r.log.Error("no se pudo liquidar la carrera", "carrera", raceID, "error", err)
		}
	}()
}

// Stop corta la simulación de una carrera. Lo llama Cancel.
func (r *Runner) Stop(raceID string) {
	r.mu.Lock()
	cancel, ok := r.running[raceID]
	r.mu.Unlock()

	if ok {
		cancel()
	}
}

// Running dice si la carrera está corriendo en este proceso.
func (r *Runner) Running(raceID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.running[raceID]
	return ok
}

func (r *Runner) forget(raceID string) {
	r.mu.Lock()
	if cancel, ok := r.running[raceID]; ok {
		cancel()
		delete(r.running, raceID)
	}
	r.mu.Unlock()
}

// Close corta todas las carreras en vivo y espera a que las goroutines salgan.
// Lo llama el cierre ordenado del servidor.
func (r *Runner) Close() {
	r.mu.Lock()
	r.closed = true
	for _, cancel := range r.running {
		cancel()
	}
	r.mu.Unlock()

	r.wg.Wait()
}

// emitTicks manda los ticks a 10 Hz. Devuelve false si se canceló.
//
// Cada tick se agenda contra un instante ABSOLUTO desde la largada, no con un
// ticker relativo: a 10 Hz durante 40 segundos, un error de pocos milisegundos
// por tick se acumularía hasta desfasar la carrera casi un segundo.
func (r *Runner) emitTicks(ctx context.Context, raceID string, startedAt time.Time, result sim.Race) bool {
	skipped := 0

	for _, tick := range result.Ticks {
		target := startedAt.Add(time.Duration(tick.T * float64(time.Second)))
		wait := target.Sub(r.clock())

		// Ya venció: es una carrera retomada tras un reinicio. Se saltea sin
		// emitir, porque mandar cuarenta segundos de ticks en un milisegundo se
		// vería como un teletransporte.
		if wait <= 0 {
			skipped++
			continue
		}

		if !sleep(ctx, wait) {
			return false
		}
		r.hub.ToRoom(raceID, NewRaceTick(raceID, tick))
	}

	if skipped > 0 {
		r.log.Info("carrera retomada", "carrera", raceID, "ticks salteados", skipped)
	}
	return true
}

// sleep espera respetando la cancelación del contexto. Devuelve false si se
// canceló.
func sleep(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return ctx.Err() == nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
