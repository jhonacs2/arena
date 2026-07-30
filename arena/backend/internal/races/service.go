package races

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/talentodh/arena/internal/contract"
	"github.com/talentodh/arena/internal/sim"
)

// Service es el dominio de las carreras. Los handlers no tienen lógica: leen el
// cuerpo, llaman acá y serializan.
//
// TODA validación pasa por este paquete, dentro de una transacción. Las monedas
// son nota y donde hay nota hay incentivo para hacer trampa: un botón
// deshabilitado en el frontend es cortesía, no un control — ver arena/CLAUDE.md
// §4.
type Service struct {
	store Store
	hub   Broadcaster
	log   *slog.Logger

	clock  func() time.Time
	seeder func() int64

	// rule es la economía. Puede ser nil: en ese caso las carreras corren y se
	// pueden cancelar, pero NO se liquidan. Ver SettlementRule.
	rule SettlementRule

	runner *Runner
}

// Config son las dependencias del servicio. Clock y Seeder existen para los
// tests; en producción se dejan nulos y toman su valor real.
type Config struct {
	Store Store
	Hub   Broadcaster
	Log   *slog.Logger

	// Rule es la regla de liquidación. Este repositorio NO trae ninguna: la
	// economía está sin decidir. Dejarla nula es un estado válido y explícito —
	// las carreras corren, se apuesta y se cancela, y al terminar se registra que
	// no se puede liquidar.
	Rule SettlementRule

	Clock  func() time.Time
	Seeder func() int64
}

func New(cfg Config) *Service {
	s := &Service{
		store:  cfg.Store,
		hub:    cfg.Hub,
		log:    cfg.Log,
		clock:  cfg.Clock,
		seeder: cfg.Seeder,
		rule:   cfg.Rule,
	}
	if s.log == nil {
		s.log = slog.Default()
	}
	if s.clock == nil {
		s.clock = time.Now
	}
	if s.seeder == nil {
		s.seeder = randomSeed
	}
	// El runner necesita liquidar y el servicio necesita largar: se cierra el
	// ciclo con la clausura, no con una dependencia mutua entre paquetes.
	s.runner = newRunner(s.log, s.clock, cfg.Hub, s.finish)
	return s
}

// Runner es el driver de las simulaciones en vivo. Lo expone para que el
// cableado pueda apagarlo al cerrar el servidor.
func (s *Service) Runner() *Runner { return s.runner }

// randomSeed es la semilla de una carrera.
//
// crypto/rand y no math/rand: la semilla decide el resultado, y el resultado es
// nota. Un generador que alguien pueda predecir desde afuera sería un alumno que
// sabe quién gana antes de apostar.
func randomSeed() int64 {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		// rand.Read no falla en las plataformas que soportamos; si algún día
		// fallara, es mejor una carrera con semilla del reloj que ninguna.
		return time.Now().UnixNano()
	}
	// El bit de signo se apaga: la columna es bigint y una semilla negativa
	// entraría igual, pero un número sin signo es más fácil de dictar por
	// teléfono cuando alguien reclama un resultado.
	return int64(binary.BigEndian.Uint64(buf[:]) >> 1)
}

// ── Guardas ───────────────────────────────────────────────────────────────

// requireAdmin es el chequeo de rol de decisiones.md §4. Todos los métodos de
// instructor empiezan con esto: el rol se verifica EN EL SERVIDOR, en cada
// endpoint. Un alumno que edite su token no obtiene nada.
func requireAdmin(id Identity) error {
	if !id.IsAdmin() {
		return contract.Errorf(contract.CodeForbidden)
	}
	return nil
}

// notFound traduce el ErrNotFound del store al código del contrato. Cualquier
// otro error sube tal cual: es un bug nuestro y termina en un INTERNAL.
func notFound(err error) error {
	if errors.Is(err, ErrNotFound) {
		return contract.Errorf(contract.CodeRaceNotFound)
	}
	return err
}

// ── Instructor: armar la carrera ──────────────────────────────────────────

// CreateRace crea una carrera en `draft`. No la ven los alumnos hasta que se
// abra.
func (s *Service) CreateRace(ctx context.Context, id Identity, in NewRace) (AdminRaceResponse, error) {
	if err := requireAdmin(id); err != nil {
		return AdminRaceResponse{}, err
	}

	in.Name = strings.TrimSpace(in.Name)
	in.CreatedBy = id.UserID
	if err := validateNewRace(in); err != nil {
		return AdminRaceResponse{}, err
	}

	var out AdminRaceResponse
	err := s.store.InTx(ctx, func(tx Tx) error {
		race, err := tx.CreateRace(ctx, in)
		if err != nil {
			return err
		}
		horses, err := tx.Horses(ctx, race.ID)
		if err != nil {
			return err
		}
		out = AdminRaceResponse{Race: adminRaceView(race, horses)}
		return nil
	})
	return out, err
}

// PatchRace edita nombre y horario. Solo en `draft`: una vez que los alumnos la
// vieron y apostaron, mover algo sería mover la nota de alguien.
func (s *Service) PatchRace(ctx context.Context, id Identity, raceID string, patch RacePatch) (AdminRaceResponse, error) {
	if err := requireAdmin(id); err != nil {
		return AdminRaceResponse{}, err
	}
	if patch.Name != nil {
		trimmed := strings.TrimSpace(*patch.Name)
		if trimmed == "" {
			return AdminRaceResponse{}, contract.FieldErrors(map[string]string{
				"name": "La carrera necesita un nombre.",
			})
		}
		patch.Name = &trimmed
	}

	var out AdminRaceResponse
	err := s.store.InTx(ctx, func(tx Tx) error {
		race, err := tx.RaceForUpdate(ctx, raceID)
		if err != nil {
			return notFound(err)
		}
		if !race.Status.Editable() {
			return contract.ErrorWith(contract.CodeInvalidTransition, map[string]any{
				"status": string(race.Status),
				"reason": "Solo se puede editar una carrera en borrador.",
			})
		}

		updated, err := tx.UpdateRace(ctx, raceID, patch)
		if err != nil {
			return err
		}
		horses, err := tx.Horses(ctx, raceID)
		if err != nil {
			return err
		}
		out = AdminRaceResponse{Race: adminRaceView(updated, horses)}
		return nil
	})
	return out, err
}

// AddHorses agrega caballos a una carrera en `draft`.
func (s *Service) AddHorses(ctx context.Context, id Identity, raceID string, in []NewHorse) (AdminRaceResponse, error) {
	if err := requireAdmin(id); err != nil {
		return AdminRaceResponse{}, err
	}
	if len(in) == 0 {
		return AdminRaceResponse{}, contract.FieldErrors(map[string]string{
			"horses": "Mandá al menos un caballo.",
		})
	}

	var out AdminRaceResponse
	err := s.store.InTx(ctx, func(tx Tx) error {
		race, err := tx.RaceForUpdate(ctx, raceID)
		if err != nil {
			return notFound(err)
		}
		if !race.Status.Editable() {
			return contract.ErrorWith(contract.CodeInvalidTransition, map[string]any{
				"status": string(race.Status),
				"reason": "Solo se pueden agregar caballos a una carrera en borrador.",
			})
		}

		// Los números ya usados entran a la validación: el esquema tiene
		// unique (race_id, number) y un 500 por violar la restricción no le
		// dice nada al instructor.
		existing, err := tx.Horses(ctx, raceID)
		if err != nil {
			return err
		}
		used := make(map[int]bool, len(existing))
		for _, h := range existing {
			used[h.Number] = true
		}
		if err := validateHorses(in, used); err != nil {
			return err
		}

		if _, err := tx.AddHorses(ctx, raceID, in); err != nil {
			return err
		}
		horses, err := tx.Horses(ctx, raceID)
		if err != nil {
			return err
		}
		out = AdminRaceResponse{Race: adminRaceView(race, horses)}
		return nil
	})
	return out, err
}

// ── Instructor: las transiciones ──────────────────────────────────────────

// Open pasa la carrera de `draft` a `open`. Desde acá los alumnos la ven y
// pueden apostar.
func (s *Service) Open(ctx context.Context, id Identity, raceID string) (AdminRaceResponse, error) {
	if err := requireAdmin(id); err != nil {
		return AdminRaceResponse{}, err
	}

	var out AdminRaceResponse
	err := s.store.InTx(ctx, func(tx Tx) error {
		race, horses, err := s.lockAndCheck(ctx, tx, raceID, StatusOpen)
		if err != nil {
			return err
		}
		// Abrir una carrera sin caballos dejaría a los alumnos mirando una
		// pantalla vacía y sin nada a lo que apostarle.
		if len(horses) == 0 {
			return contract.ErrorWith(contract.CodeInvalidTransition, map[string]any{
				"reason": "La carrera no tiene caballos.",
			})
		}

		now := s.clock()
		if err := tx.SetStatus(ctx, raceID, StatusOpen, now); err != nil {
			return err
		}
		race.Status = StatusOpen
		race.OpenedAt = &now
		out = AdminRaceResponse{Race: adminRaceView(race, horses)}
		return nil
	})
	if err != nil {
		return AdminRaceResponse{}, err
	}

	s.log.Info("carrera abierta", "carrera", raceID, "instructor", id.UserID)
	return out, nil
}

// Start pasa la carrera de `open` a `running`: CIERRA LAS APUESTAS en el
// servidor, fija la semilla y arranca la simulación.
//
// La semilla se guarda antes de simular. Con ella la carrera es reproducible: si
// alguien reclama el resultado, se vuelve a correr igual.
func (s *Service) Start(ctx context.Context, id Identity, raceID string) (AdminRaceResponse, error) {
	if err := requireAdmin(id); err != nil {
		return AdminRaceResponse{}, err
	}

	var (
		out       AdminRaceResponse
		startedAt time.Time
		seed      int64
		horses    []Horse
		revealed  []PublicBetView
	)
	err := s.store.InTx(ctx, func(tx Tx) error {
		race, hs, err := s.lockAndCheck(ctx, tx, raceID, StatusRunning)
		if err != nil {
			return err
		}
		if len(hs) == 0 {
			return contract.ErrorWith(contract.CodeInvalidTransition, map[string]any{
				"reason": "La carrera no tiene caballos.",
			})
		}

		startedAt = s.clock()
		seed = s.seeder()
		if err := tx.SetSeed(ctx, raceID, seed); err != nil {
			return err
		}
		if err := tx.SetStatus(ctx, raceID, StatusRunning, startedAt); err != nil {
			return err
		}

		// Las apuestas se revelan TODAS JUNTAS acá, no antes: mientras la
		// carrera estaba `open`, `bet.placed` no llevaba el caballo.
		bets, err := tx.Bets(ctx, raceID)
		if err != nil {
			return err
		}
		revealed = betViews(StatusRunning, bets)

		horses = hs
		race.Status = StatusRunning
		race.StartedAt = &startedAt
		race.Seed = &seed
		out = AdminRaceResponse{Race: adminRaceView(race, horses)}
		return nil
	})
	if err != nil {
		return AdminRaceResponse{}, err
	}

	// Recién con la transacción confirmada se difunde y se larga: si se hiciera
	// adentro y la transacción se revirtiera, los clientes verían correr una
	// carrera que en la base sigue en `open`.
	s.hub.ToRoom(raceID, NewRaceStarted(raceID, startedAt, revealed))

	result := sim.Simulate(seed, toSimHorses(horses))
	s.runner.Start(raceID, result, startedAt)

	s.log.Info("largó",
		"carrera", raceID, "semilla", seed,
		"duración", result.Duration, "apuestas", len(revealed))
	return out, nil
}

// Cancel pasa la carrera a `cancelled` desde `draft`, `open` o `running`, y
// DEVUELVE CADA APUESTA ÍNTEGRA al saldo, en una transacción.
func (s *Service) Cancel(ctx context.Context, id Identity, raceID, reason string) (AdminRaceResponse, error) {
	if err := requireAdmin(id); err != nil {
		return AdminRaceResponse{}, err
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "La carrera fue cancelada por el instructor."
	}

	type refund struct {
		userID  string
		amount  int64
		balance int64
	}

	var (
		out     AdminRaceResponse
		refunds []refund
	)
	err := s.store.InTx(ctx, func(tx Tx) error {
		race, horses, err := s.lockAndCheck(ctx, tx, raceID, StatusCancelled)
		if err != nil {
			return err
		}

		now := s.clock()
		if err := tx.SetStatus(ctx, raceID, StatusCancelled, now); err != nil {
			return err
		}

		bets, err := tx.Bets(ctx, raceID)
		if err != nil {
			return err
		}
		refunds = refunds[:0]
		for _, bet := range bets {
			// Solo las que están `placed`. Una ya liquidada no se toca: el
			// ledger es append-only y volver a moverla duplicaría el pago.
			if bet.Status != BetStatusPlaced {
				continue
			}

			// ÍNTEGRA: se devuelve exactamente bet.Amount. No el pago
			// potencial, no un porcentaje.
			balance, err := tx.Move(ctx, Movement{
				UserID:    bet.UserID,
				Delta:     bet.Amount,
				Reason:    ReasonBetRefunded,
				RefID:     bet.ID,
				CreatedBy: id.UserID,
			})
			if err != nil {
				return err
			}
			if err := tx.SettleBet(ctx, bet.ID, BetStatusRefunded, bet.Amount, now); err != nil {
				return err
			}
			refunds = append(refunds, refund{userID: bet.UserID, amount: bet.Amount, balance: balance})
		}

		race.Status = StatusCancelled
		out = AdminRaceResponse{Race: adminRaceView(race, horses)}
		return nil
	})
	if err != nil {
		return AdminRaceResponse{}, err
	}

	// Si estaba corriendo, se corta la simulación. Va después del commit: el
	// runner liquida bajo el mismo bloqueo de fila, así que si llegó primero ve
	// `cancelled` y aborta solo.
	s.runner.Stop(raceID)

	byUser := make(map[string]refund, len(refunds))
	for _, r := range refunds {
		byUser[r.userID] = r
	}
	s.hub.ToRoomPerUser(raceID, func(userID string) any {
		event := RaceCancelled{Type: EventRaceCancelled, RaceID: raceID, Reason: reason}
		if r, ok := byUser[userID]; ok {
			amount := r.amount
			event.MyRefund = &amount
			event.Balance = balanceOf(r.balance, true)
		}
		return event
	})

	s.log.Info("carrera cancelada",
		"carrera", raceID, "instructor", id.UserID, "devoluciones", len(refunds))
	return out, nil
}

// lockAndCheck bloquea la fila y valida la transición.
//
// Es el único camino a un cambio de estado: con la fila bloqueada, dos clics
// simultáneos en «largar» no pueden leer los dos el mismo `open`. El segundo
// espera, ve `running` y recibe INVALID_TRANSITION.
func (s *Service) lockAndCheck(ctx context.Context, tx Tx, raceID string, to Status) (Race, []Horse, error) {
	race, err := tx.RaceForUpdate(ctx, raceID)
	if err != nil {
		return Race{}, nil, notFound(err)
	}
	if !CanTransition(race.Status, to) {
		return Race{}, nil, contract.ErrorWith(contract.CodeInvalidTransition, map[string]any{
			"from": string(race.Status),
			"to":   string(to),
		})
	}

	horses, err := tx.Horses(ctx, raceID)
	if err != nil {
		return Race{}, nil, err
	}
	return race, horses, nil
}

func toSimHorses(horses []Horse) []sim.Horse {
	out := make([]sim.Horse, len(horses))
	for i, h := range horses {
		out[i] = sim.Horse{ID: h.ID, Name: h.Name, Number: h.Number, Odds: h.NominalOdds}
	}
	return out
}

// ── Alumno ────────────────────────────────────────────────────────────────

// List son las carreras visibles: `open`, `running` y `finished`. NUNCA `draft`,
// salvo que quien pregunte sea el instructor.
func (s *Service) List(ctx context.Context, id Identity) (RaceListResponse, error) {
	var out RaceListResponse
	err := s.store.InTx(ctx, func(tx Tx) error {
		summaries, err := tx.VisibleRaces(ctx, id.UserID, id.IsAdmin())
		if err != nil {
			return err
		}
		items := make([]RaceView, 0, len(summaries))
		for _, summary := range summaries {
			items = append(items, raceView(summary))
		}
		out = RaceListResponse{Items: items}
		return nil
	})
	return out, err
}

// Detail es el detalle de una carrera: caballos, sala, apuestas y —si terminó—
// resultados.
func (s *Service) Detail(ctx context.Context, id Identity, raceID string) (RaceDetailView, error) {
	var out RaceDetailView
	err := s.store.InTx(ctx, func(tx Tx) error {
		view, err := s.detail(ctx, tx, raceID, id)
		if err != nil {
			return err
		}
		out = view
		return nil
	})
	return out, err
}

// detail arma el detalle POR DESTINATARIO: `myBet` es de quien pregunta, y las
// apuestas del resto pasan por betViews, que no revela el caballo mientras la
// carrera está `open`.
func (s *Service) detail(ctx context.Context, tx Tx, raceID string, id Identity) (RaceDetailView, error) {
	race, err := tx.Race(ctx, raceID)
	if err != nil {
		return RaceDetailView{}, notFound(err)
	}

	// Una carrera en `draft` no existe para el alumno. RACE_NOT_FOUND y no
	// FORBIDDEN: un 403 le confirmaría que la carrera existe, y la lista de
	// carreras que el instructor está armando no es información suya.
	if !race.Status.Visible() && !id.IsAdmin() {
		return RaceDetailView{}, contract.Errorf(contract.CodeRaceNotFound)
	}

	horses, err := tx.Horses(ctx, raceID)
	if err != nil {
		return RaceDetailView{}, err
	}
	participants, err := tx.Participants(ctx, raceID)
	if err != nil {
		return RaceDetailView{}, err
	}
	bets, err := tx.Bets(ctx, raceID)
	if err != nil {
		return RaceDetailView{}, err
	}

	var results []sim.Result
	if race.Status == StatusFinished {
		results, err = tx.Results(ctx, raceID)
		if err != nil {
			return RaceDetailView{}, err
		}
	}

	// La propia se saca de la lista que ya se leyó: una consulta menos, y no
	// hay forma de que las dos discrepen.
	var mine *MyBetView
	others := make([]Bet, 0, len(bets))
	for _, bet := range bets {
		if bet.UserID == id.UserID {
			mine = myBetViewPtr(bet, true)
			continue
		}
		others = append(others, bet)
	}

	return RaceDetailView{
		ID:           race.ID,
		Name:         race.Name,
		Status:       race.Status,
		ScheduledAt:  stamp(race.ScheduledAt),
		OpenedAt:     stamp(race.OpenedAt),
		StartedAt:    stamp(race.StartedAt),
		FinishedAt:   stamp(race.FinishedAt),
		Horses:       horseViews(horses),
		Participants: participantViews(participants),
		Bets:         betViews(race.Status, others),
		MyBet:        mine,
		Results:      resultViews(results),
	}, nil
}

// Join entra a la sala por HTTP. Idempotente: entrar dos veces no es un error, y
// solo la primera difunde `room.joined`.
func (s *Service) Join(ctx context.Context, id Identity, raceID string) (RaceDetailView, error) {
	view, added, err := s.join(ctx, id, raceID)
	if err != nil {
		return RaceDetailView{}, err
	}
	if added {
		s.hub.ToRoom(raceID, NewRoomJoined(raceID, id.UserID, id.Username, len(view.Participants)))
	}
	return view, nil
}

// JoinRoom es lo que corre cuando alguien se conecta al socket de una carrera.
//
// Devuelve el `room.state` para quien entró —armado por destinatario, porque
// incluye su propia apuesta— y el `room.joined` que el hub difunde al resto.
// Se pasa como JoinHook a ws.Hub.Handler.
func (s *Service) JoinRoom(ctx context.Context, raceID string, id Identity) (self any, room any, err error) {
	view, added, err := s.join(ctx, id, raceID)
	if err != nil {
		return nil, nil, err
	}

	self = NewRoomState(raceID, view)
	if added {
		room = NewRoomJoined(raceID, id.UserID, id.Username, len(view.Participants))
	}
	return self, room, nil
}

// join mete al alumno en la sala y devuelve el detalle. No difunde nada: quien
// llama decide, porque por HTTP el evento va a toda la sala y por socket va a
// todos menos al que entró.
func (s *Service) join(ctx context.Context, id Identity, raceID string) (RaceDetailView, bool, error) {
	var (
		out   RaceDetailView
		added bool
	)
	err := s.store.InTx(ctx, func(tx Tx) error {
		race, err := tx.Race(ctx, raceID)
		if err != nil {
			return notFound(err)
		}
		// A la sala de una carrera en `draft` no se entra: para el alumno esa
		// carrera no existe.
		if !race.Status.Visible() && !id.IsAdmin() {
			return contract.Errorf(contract.CodeRaceNotFound)
		}

		// El instructor mira la sala pero no la ocupa: no es un participante y
		// no aparece en la lista de quiénes están apostando.
		if !id.IsAdmin() {
			added, err = tx.AddParticipant(ctx, raceID, id.UserID)
			if err != nil {
				return err
			}
		}

		view, err := s.detail(ctx, tx, raceID, id)
		if err != nil {
			return err
		}
		out = view
		return nil
	})
	if err != nil {
		return RaceDetailView{}, false, err
	}
	return out, added, nil
}

// PlaceBet apuesta.
//
// Todo se valida ACÁ y en una transacción: la carrera está `open`, el caballo es
// de esa carrera, `1 ≤ amount ≤ balance`, y no hay apuesta previa de este alumno
// en esta carrera. La cuota se congela en `odds_at_bet` y nunca se recalcula
// desde la cuota actual del caballo.
func (s *Service) PlaceBet(ctx context.Context, id Identity, raceID, horseID string, amount int64) (BetResponse, error) {
	fields := map[string]string{}
	if horseID == "" {
		fields["horseId"] = "Elegí un caballo."
	}
	if amount < 1 {
		fields["amount"] = "El monto mínimo es 1 moneda."
	}
	if len(fields) > 0 {
		return BetResponse{}, contract.FieldErrors(fields)
	}

	var (
		out      BetResponse
		betCount int
		placed   Bet
	)
	err := s.store.InTx(ctx, func(tx Tx) error {
		// La fila de la carrera se bloquea: sin el bloqueo, una apuesta podría
		// colarse entre el `SELECT status` y el `UPDATE` de la largada.
		race, err := tx.RaceForUpdate(ctx, raceID)
		if err != nil {
			return notFound(err)
		}
		if !race.Status.AcceptsBets() {
			return contract.ErrorWith(contract.CodeRaceNotOpen, map[string]any{
				"status": string(race.Status),
			})
		}

		horse, err := tx.Horse(ctx, horseID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return contract.FieldErrors(map[string]string{
					"horseId": "Ese caballo no corre en esta carrera.",
				})
			}
			return err
		}
		// El caballo tiene que ser DE ESTA CARRERA. Sin este chequeo se podría
		// apostarle al favorito de otra carrera y cobrar acá.
		if horse.RaceID != raceID {
			return contract.FieldErrors(map[string]string{
				"horseId": "Ese caballo no corre en esta carrera.",
			})
		}

		// Una apuesta por carrera y por alumno. Con cuotas fijas, cubrir todos
		// los caballos sería una máquina de imprimir nota — decisiones.md §1.
		// El esquema tiene unique (race_id, user_id); esto es para dar el código
		// del contrato en vez de un 500.
		if _, exists, err := tx.BetByUser(ctx, raceID, id.UserID); err != nil {
			return err
		} else if exists {
			return contract.Errorf(contract.CodeBetAlreadyPlaced)
		}

		balance, err := tx.Balance(ctx, id.UserID)
		if err != nil {
			return err
		}
		if amount > balance {
			return contract.ErrorWith(contract.CodeInsufficientBalance, map[string]any{
				"balance": balance,
				"amount":  amount,
			})
		}

		placed, err = tx.InsertBet(ctx, Bet{
			RaceID:    raceID,
			UserID:    id.UserID,
			Username:  id.Username,
			HorseID:   horseID,
			Amount:    amount,
			Status:    BetStatusPlaced,
			CreatedAt: s.clock(),
		})
		if err != nil {
			return err
		}

		after, err := tx.Move(ctx, Movement{
			UserID: id.UserID,
			Delta:  -amount,
			Reason: ReasonBetPlaced,
			RefID:  placed.ID,
		})
		if err != nil {
			// El ledger no confía en su llamador: rechaza el movimiento si el
			// saldo quedaría negativo. Ya se validó `amount <= balance` arriba,
			// así que llegar acá significa que hubo una carrera de datos —
			// se loguea para que se vea, y al alumno le llega el código que
			// describe lo que pasó.
			s.log.Error("el ledger rechazó el descuento de una apuesta ya validada",
				"carrera", raceID, "usuario", id.UserID, "monto", amount, "error", err)
			return contract.ErrorWith(contract.CodeInsufficientBalance, map[string]any{
				"amount": amount,
			})
		}

		// Apostar mete al alumno en la sala: si apostó, quiere ver la carrera.
		if _, err := tx.AddParticipant(ctx, raceID, id.UserID); err != nil {
			return err
		}

		bets, err := tx.Bets(ctx, raceID)
		if err != nil {
			return err
		}
		betCount = len(bets)

		out = betResponse(placed, after)
		return nil
	})
	if err != nil {
		return BetResponse{}, err
	}

	// `bet.placed` NO lleva el caballo: la carrera está `open` y revelarlo haría
	// que los últimos en apostar copien a los primeros. El tipo del evento
	// directamente no tiene el campo.
	s.hub.ToRoom(raceID, NewBetPlaced(raceID, id.UserID, id.Username, amount, betCount))

	return out, nil
}

// ── Validación ────────────────────────────────────────────────────────────

func validateNewRace(in NewRace) error {
	fields := map[string]string{}
	if in.Name == "" {
		fields["name"] = "La carrera necesita un nombre."
	}
	if len(fields) > 0 {
		return contract.FieldErrors(fields)
	}
	// Crear la carrera sin caballos está permitido: el instructor la arma en
	// varios pasos y para eso existe POST /admin/races/:id/horses. Lo que no se
	// puede es ABRIRLA vacía.
	if len(in.Horses) == 0 {
		return nil
	}
	return validateHorses(in.Horses, map[int]bool{})
}

// validateHorses valida los caballos nuevos contra los números ya usados.
//
// Repite los CHECK del esquema a propósito. La base es la que garantiza la
// regla; esto existe para que el instructor lea «la cuota mínima es 1.01» en vez
// de un error de restricción de Postgres.
func validateHorses(horses []NewHorse, used map[int]bool) error {
	fields := map[string]string{}
	seen := map[int]bool{}

	for _, h := range horses {
		if strings.TrimSpace(h.Name) == "" {
			fields["horses.name"] = "Todos los caballos necesitan nombre."
		}
		if h.Number < 1 {
			fields["horses.number"] = "El número de partida arranca en 1."
		}
		if seen[h.Number] || used[h.Number] {
			fields["horses.number"] = "Hay dos caballos con el mismo número de partida."
		}
		seen[h.Number] = true

		// Cuotas ×100: 101 es 1.01. Una cuota de 1.00 sería devolver lo
		// apostado, y eso no es una apuesta.
		if h.Odds < contract.MinNominalOdds {
			fields["horses.nominalOdds"] = "La cuota mínima es 1.01, y va en entero ×100 (340 es 3.40)."
		}
	}

	if len(fields) > 0 {
		return contract.FieldErrors(fields)
	}
	return nil
}
