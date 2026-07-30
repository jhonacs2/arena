package races

import (
	"context"
	"errors"
	"time"

	"github.com/talentodh/arena/internal/sim"
)

// ── La economía, que está sin decidir ─────────────────────────────────────

// ErrNoSettlementRule lo devuelve finish cuando no hay regla de liquidación
// cableada. Es un error y no un valor por defecto A PROPÓSITO: ver SettlementRule.
var ErrNoSettlementRule = errors.New("no hay regla de liquidación cableada: la economía está sin decidir")

// SettlementRule decide cuánto cobra cada apuesta de una carrera terminada.
//
// **ESTE REPOSITORIO NO LA IMPLEMENTA, Y ES DELIBERADO.** La economía de Arena es
// una decisión abierta, y las dos candidatas dan números distintos para la misma
// carrera:
//
//	cuotas fijas   pago = amount * oddsAtBet / 100      paga «la casa»; la masa de
//	                                                    monedas del curso crece
//	pari-mutuel    pago = amount * pool / winningPool    suma cero; los alumnos se
//	                                                    reparten entre ellos
//
// La diferencia no es de implementación: cambia el esquema —cuotas fijas necesita
// una cuota congelada al apostar, pari-mutuel necesita repartir el resto de la
// división para no perder monedas— y cambia qué significa la nota. Poner una de
// las dos «por ahora» sería lo peor posible: compilaría, pasaría los tests, y
// nadie volvería a mirarla hasta que treinta alumnos tuvieran la nota calculada
// con una regla que nadie eligió.
//
// Así que el resto del paquete está completo y esta pieza falta en UN solo lugar,
// con nombre. Sin ella una carrera no se liquida: se queda en `running`, se
// registra el error y no se escribe nada. Es incómodo, y esa es la idea.
//
// Recibe la transacción en curso para poder escribir sus propios registros —el
// pool y el pool ganador, si la economía elegida los tiene— dentro de la misma
// transacción que los pagos. La liquidación se calcula UNA sola vez y queda
// escrita: recalcular con datos que cambiaron es cómo se paga dos veces.
type SettlementRule interface {
	Settle(ctx context.Context, tx Tx, race Race, bets []Bet, winnerHorseID string) (Payouts, error)
}

// Payouts es lo que decidió la regla.
type Payouts struct {
	// ByBet es el pago de cada apuesta GANADORA, indexado por id de apuesta. Las
	// que no estén acá pierden y su pago es 0.
	ByBet map[string]int64

	// RefundAll pide devolver cada apuesta ÍNTEGRA en lugar de pagar.
	//
	// Es el caso «no acertó nadie»: no hay a quién pagarle. Las dos economías
	// coinciden en esto, así que la mecánica de la devolución vive en este paquete
	// y lo único que aporta la regla es la decisión.
	RefundAll bool
}

// ── La liquidación ────────────────────────────────────────────────────────

// settlement es lo que le pasó a un alumno. Se junta dentro de la transacción y se
// usa después para armar el evento de cada uno.
type settlement struct {
	bet     Bet
	balance int64
}

// finish liquida la carrera: `running → finished`.
//
// Todo en UNA transacción: los resultados, el estado, cada apuesta a won, lost o
// refunded, y cada pago acreditado. Una liquidación a medias sería la mitad de un
// curso con la nota mal.
//
// La llama el Runner cuando cruza el último tick. No es un endpoint: nadie de
// afuera puede pedir que una carrera termine.
func (s *Service) finish(ctx context.Context, raceID string, result sim.Race) error {
	winner, ok := result.Winner()
	if !ok {
		s.log.Error("la simulación terminó sin ganador", "carrera", raceID)
		return nil
	}

	// Sin regla no se liquida NADA: no se marca ninguna apuesta, no se escribe el
	// estado y no se toca un saldo. La carrera queda en `running` y la retoma
	// Resume cuando la economía esté cableada.
	if s.rule == nil {
		s.log.Error("no se liquida: falta la regla de liquidación",
			"carrera", raceID, "ganador", winner.HorseName,
			"detalle", "ver races.SettlementRule — la economía está sin decidir")
		return ErrNoSettlementRule
	}

	var (
		settled  []settlement
		finished time.Time
		results  []ResultView
		revealed []PublicBetView
		skipped  bool
	)

	err := s.store.InTx(ctx, func(tx Tx) error {
		race, err := tx.RaceForUpdate(ctx, raceID)
		if err != nil {
			return notFound(err)
		}

		// Si ya no está en `running`, alguien la canceló mientras corría. El
		// bloqueo de fila hace que una de las dos llegue primero, y esta es la que
		// llegó segunda: no liquida, porque las apuestas ya se devolvieron.
		if !CanTransition(race.Status, StatusFinished) {
			s.log.Info("no se liquida: la carrera ya no está corriendo",
				"carrera", raceID, "estado", race.Status)
			skipped = true
			return nil
		}

		finished = s.clock()
		if err := tx.InsertResults(ctx, raceID, result.Results); err != nil {
			return err
		}
		if err := tx.SetStatus(ctx, raceID, StatusFinished, finished); err != nil {
			return err
		}

		bets, err := tx.Bets(ctx, raceID)
		if err != nil {
			return err
		}

		// La economía decide acá, y SOLO acá. Escribe sus propios registros en esta
		// misma transacción.
		payouts, err := s.rule.Settle(ctx, tx, race, bets, winner.HorseID)
		if err != nil {
			return err
		}

		for _, bet := range bets {
			// Una que no esté `placed` ya se liquidó o se devolvió. El ledger es
			// append-only: volver a moverla duplicaría el pago.
			if bet.Status != BetStatusPlaced {
				continue
			}

			status, amount := outcome(bet, payouts, winner.HorseID)

			// Perder no mueve el saldo: el descuento ya se hizo al apostar.
			if amount == 0 {
				if err := tx.SettleBet(ctx, bet.ID, status, 0, finished); err != nil {
					return err
				}
				balance, err := tx.Balance(ctx, bet.UserID)
				if err != nil {
					return err
				}
				settled = append(settled, settlement{
					bet: settledBet(bet, status, 0, finished), balance: balance,
				})
				continue
			}

			balance, err := tx.Move(ctx, Movement{
				UserID: bet.UserID,
				Delta:  amount,
				Reason: reasonFor(status),
				RefID:  bet.ID,
			})
			if err != nil {
				return err
			}
			if err := tx.SettleBet(ctx, bet.ID, status, amount, finished); err != nil {
				return err
			}
			settled = append(settled, settlement{
				bet: settledBet(bet, status, amount, finished), balance: balance,
			})
		}

		results = resultViews(result.Results)
		// La carrera terminó, así que las apuestas se muestran con el caballo. El
		// PAGO de cada uno no sale por acá: betViews no lo incluye nunca.
		revealed = betViews(StatusFinished, betsOf(settled))
		return nil
	})
	if err != nil {
		return err
	}
	if skipped {
		return nil
	}

	// `race.finished` SE ARMA POR DESTINATARIO. Difundir el mismo objeto filtraría
	// cuánto cobró cada uno.
	byUser := make(map[string]settlement, len(settled))
	for _, item := range settled {
		byUser[item.bet.UserID] = item
	}
	s.hub.ToRoomPerUser(raceID, func(userID string) any {
		event := RaceFinished{
			Type:    EventRaceFinished,
			RaceID:  raceID,
			Results: results,
		}

		mine, hasBet := byUser[userID]
		// Las de los demás: se saca la propia de la lista para no repetirla.
		event.Bets = withoutUser(revealed, userID)
		if hasBet {
			view := myBetView(mine.bet)
			event.MyBet = &view
			event.Balance = balanceOf(mine.balance, true)
		}
		return event
	})

	won := 0
	for _, item := range settled {
		if item.bet.Status == BetStatusWon {
			won++
		}
	}
	s.log.Info("llegó",
		"carrera", raceID,
		"ganador", winner.HorseName,
		"liquidadas", len(settled),
		"ganadoras", won)
	return nil
}

// outcome decide en qué queda una apuesta y cuánto cobra.
//
// La devolución tiene prioridad: si la regla pidió devolver, se devuelve a TODOS,
// incluso a quien le había acertado al caballo. Es el caso «no acertó nadie», y
// mezclarlo con un pago sería pagarle dos veces al mismo.
func outcome(bet Bet, payouts Payouts, winnerHorseID string) (BetStatus, int64) {
	if payouts.RefundAll {
		return BetStatusRefunded, bet.Amount
	}
	if amount, ok := payouts.ByBet[bet.ID]; ok && amount > 0 {
		return BetStatusWon, amount
	}
	// Una apuesta al ganador con pago 0 sigue siendo GANADORA: puede pasar si la
	// economía trunca a cero un monto diminuto. Se marca won con pago 0 y no lost,
	// porque acertó.
	if bet.HorseID == winnerHorseID {
		return BetStatusWon, 0
	}
	return BetStatusLost, 0
}

func reasonFor(status BetStatus) Reason {
	if status == BetStatusRefunded {
		return ReasonBetRefunded
	}
	return ReasonBetWon
}

func settledBet(bet Bet, status BetStatus, payout int64, at time.Time) Bet {
	bet.Status = status
	bet.Payout = &payout
	bet.SettledAt = &at
	return bet
}

// ── Retomar ───────────────────────────────────────────────────────────────

// Resume retoma las carreras que quedaron en `running`.
//
// Existe porque el proceso se puede caer en el medio de una carrera, y una carrera
// en `running` para siempre es un grupo de apuestas que nunca se liquidan. Se
// vuelve a correr desde la SEMILLA GUARDADA, así que el resultado es exactamente el
// que se estaba viendo antes del reinicio.
//
// La llama el cableado al arrancar el servidor, después de conectar a la base.
func (s *Service) Resume(ctx context.Context) error {
	var pending []Race
	err := s.store.InTx(ctx, func(tx Tx) error {
		found, err := tx.RunningRaces(ctx)
		if err != nil {
			return err
		}
		pending = found
		return nil
	})
	if err != nil {
		return err
	}

	for _, race := range pending {
		if race.Seed == nil || race.StartedAt == nil {
			// No debería pasar: la semilla y la largada se escriben en la misma
			// transacción que el estado. Si pasa, se avisa y se deja quieta en vez
			// de inventar un resultado.
			s.log.Error("carrera en running sin semilla o sin largada; no se retoma",
				"carrera", race.ID)
			continue
		}

		var horses []Horse
		err := s.store.InTx(ctx, func(tx Tx) error {
			found, err := tx.Horses(ctx, race.ID)
			if err != nil {
				return err
			}
			horses = found
			return nil
		})
		if err != nil {
			return err
		}

		result := sim.Simulate(*race.Seed, toSimHorses(horses))
		s.runner.Start(race.ID, result, *race.StartedAt)
		s.log.Info("se retoma una carrera que quedó corriendo",
			"carrera", race.ID, "semilla", *race.Seed, "largó", race.StartedAt.UTC())
	}
	return nil
}

func betsOf(items []settlement) []Bet {
	out := make([]Bet, 0, len(items))
	for _, item := range items {
		out = append(out, item.bet)
	}
	return out
}

func withoutUser(views []PublicBetView, userID string) []PublicBetView {
	out := make([]PublicBetView, 0, len(views))
	for _, v := range views {
		if v.UserID == userID {
			continue
		}
		out = append(out, v)
	}
	return out
}
