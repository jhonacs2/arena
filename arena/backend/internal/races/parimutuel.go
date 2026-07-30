package races

import (
	"context"
	"fmt"
	"sort"
)

// PariMutuel reparte lo apostado entre quienes acertaron. Es la economía decidida
// de Arena — ver `docs/contract/decisiones.md` §0.
//
//	pago = monto × pool / poolGanador
//
// No hay casa: el pool sale de las apuestas y vuelve a ellas. Eso lo hace **suma
// cero**, y es la mitad de la decisión: la otra mitad es el piso de 10 puntos, que
// vive en la vista `user_scores` del esquema. Las dos van juntas porque con suma
// cero y sin piso la nota que gana un alumno saldría de la de otro.
//
// Implementa SettlementRule y es la única pieza que sabe de la economía. El resto
// del paquete no cambia si algún día se elige otra.
type PariMutuel struct{}

var _ SettlementRule = PariMutuel{}

// Settle calcula los pagos y escribe la liquidación.
//
// Devuelve error si el reparto no conserva el pool. Es una comprobación
// redundante con el CHECK del esquema, a propósito: acá el mensaje dice *qué*
// cuenta no cerró, y la transacción se revierte antes de tocar un saldo.
func (PariMutuel) Settle(
	ctx context.Context, tx Tx, race Race, bets []Bet, winnerHorseID string,
) (Payouts, error) {
	// Solo cuentan las apuestas todavía en juego. Una que ya está `won`, `lost` o
	// `refunded` se liquidó antes, y sumarla al pool pagaría con dinero que ya
	// salió — `finish` la saltea, pero el pool se calcula acá y tiene que
	// coincidir con lo que `finish` va a pagar.
	var live []Bet
	var pool, winningPool int64
	for _, bet := range bets {
		if bet.Status != BetStatusPlaced {
			continue
		}
		live = append(live, bet)
		pool += bet.Amount
		if bet.HorseID == winnerHorseID {
			winningPool += bet.Amount
		}
	}

	settlement := Settlement{
		RaceID:        race.ID,
		WinnerHorseID: winnerHorseID,
		Pool:          pool,
		WinningPool:   winningPool,
	}

	// Nadie le apostó al ganador: no hay a quién pagarle, y quedarse el pool sería
	// inventar una casa que este modelo no tiene. Se devuelve cada apuesta íntegra.
	//
	// El pool vacío —una carrera sin apuestas— cae acá también, y es correcto: se
	// devuelve nada, y queda la fila de liquidación registrando que corrió.
	if winningPool == 0 {
		settlement.Refunded = true
		settlement.PaidOut = pool // devolver todo también agota el pool
		if err := tx.InsertSettlement(ctx, settlement); err != nil {
			return Payouts{}, err
		}
		return Payouts{RefundAll: true}, nil
	}

	// ── El reparto ────────────────────────────────────────────────────────────
	//
	// División entera, que casi nunca es exacta: con un pool de 800 sobre 300
	// ganadores, los pagos truncados suman 799 y sobra 1 moneda. Ese resto
	// —*breakage* en las carreras de verdad— **no se descarta**: si se descartara,
	// el pari-mutuel dejaría de ser suma cero y el curso perdería monedas que
	// nadie perdió apostando.
	winners := make([]Bet, 0, len(live))
	for _, bet := range live {
		if bet.HorseID == winnerHorseID {
			winners = append(winners, bet)
		}
	}

	// Orden determinístico, porque decide a quién le toca la moneda del resto: el
	// que más puso primero, y a igual monto el que apostó antes. Sin un orden
	// fijo, la misma carrera liquidada dos veces daría pagos distintos.
	sort.Slice(winners, func(i, j int) bool {
		if winners[i].Amount != winners[j].Amount {
			return winners[i].Amount > winners[j].Amount
		}
		if !winners[i].CreatedAt.Equal(winners[j].CreatedAt) {
			return winners[i].CreatedAt.Before(winners[j].CreatedAt)
		}
		return winners[i].ID < winners[j].ID
	})

	byBet := make(map[string]int64, len(winners))
	var truncated int64
	for _, bet := range winners {
		// `amount * pool` no desborda con los montos de un curso: con 30 alumnos y
		// el tope de apuesta del contrato queda seis órdenes de magnitud por debajo
		// de int64. Si alguna vez sube el tope, esto es lo que hay que revisar.
		payout := bet.Amount * pool / winningPool
		byBet[bet.ID] = payout
		truncated += payout
	}

	// La moneda que sobró, de a una, hasta agotarla.
	//
	// Nunca hay más resto que ganadores: con división exacta la suma daría el pool
	// justo, así que el resto es la suma de las partes fraccionarias, cada una menor
	// que 1. Con n ganadores, resto ≤ n−1. El `min` es cinturón por si esa
	// invariante se rompiera alguna vez: antes prefiero perder una moneda que
	// entrar en pánico liquidando la nota de treinta personas.
	remainder := pool - truncated
	if remainder > int64(len(winners)) {
		remainder = int64(len(winners))
	}
	for i := int64(0); i < remainder; i++ {
		byBet[winners[i].ID]++
	}

	// Se vuelve a sumar en vez de asumir el pool: si el reparto tuviera un error,
	// esta cuenta es la que lo encuentra. Darle el valor esperado y después
	// compararlo con sí mismo no verifica nada.
	var paid int64
	for _, payout := range byBet {
		paid += payout
	}
	if paid != pool {
		return Payouts{}, fmt.Errorf(
			"la liquidación no conserva: se pagan %d de un pool de %d en la carrera %s",
			paid, pool, race.ID)
	}
	settlement.PaidOut = paid
	if err := tx.InsertSettlement(ctx, settlement); err != nil {
		return Payouts{}, err
	}

	return Payouts{ByBet: byBet}, nil
}
