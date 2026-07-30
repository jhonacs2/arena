package races

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// settleWith corre la regla contra el doble en memoria y devuelve los pagos y la
// liquidación escrita. Pasa por InTx porque la regla escribe dentro de la
// transacción, y el doble reproduce los CHECK del esquema.
func settleWith(t *testing.T, bets []Bet, winnerHorseID string) (Payouts, Settlement) {
	t.Helper()

	store := newMemStore()
	var payouts Payouts
	var settlement Settlement

	err := store.InTx(context.Background(), func(tx Tx) error {
		var err error
		payouts, err = PariMutuel{}.Settle(
			context.Background(), tx, Race{ID: "race-1"}, bets, winnerHorseID)
		if err != nil {
			return err
		}
		var ok bool
		settlement, ok, err = tx.Settlement(context.Background(), "race-1")
		if err != nil {
			return err
		}
		if !ok {
			t.Fatal("la regla no escribió la liquidación")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Settle: %v", err)
	}
	return payouts, settlement
}

// bet arma una apuesta. El orden de creación importa: decide a quién le toca la
// moneda del resto cuando dos pusieron lo mismo.
func bet(id, userID, horseID string, amount int64, createdOffset time.Duration) Bet {
	return Bet{
		ID:        id,
		RaceID:    "race-1",
		UserID:    userID,
		HorseID:   horseID,
		Amount:    amount,
		Status:    BetStatusPlaced,
		CreatedAt: time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC).Add(createdOffset),
	}
}

// El caso que no cierra en enteros, que es el que importa. Es el mismo que asierta
// `docs/contract/schema.test.sql`, a propósito: si el Go y el SQL discrepan, uno de
// los dos está mal y hay que saber cuál.
//
//	pool 800, pool ganador 300
//	Ana   100 × 800 / 300 = 266,66 → 266
//	Bruno 200 × 800 / 300 = 533,33 → 533
//	suma 799, sobra 1 → al de apuesta mayor → Bruno 534
func TestPariMutuelReparteElRestoAlQueMasApostó(t *testing.T) {
	bets := []Bet{
		bet("bet-ana", "ana", "horse-3", 100, 0),
		bet("bet-bruno", "bruno", "horse-3", 200, time.Second),
		bet("bet-carla", "carla", "horse-1", 500, 2*time.Second),
	}

	payouts, settlement := settleWith(t, bets, "horse-3")

	if payouts.RefundAll {
		t.Fatal("hubo aciertos: no corresponde devolver")
	}
	if got := payouts.ByBet["bet-ana"]; got != 266 {
		t.Errorf("Ana cobra %d, esperado 266", got)
	}
	if got := payouts.ByBet["bet-bruno"]; got != 534 {
		t.Errorf("Bruno cobra %d, esperado 534 (533 truncado + la moneda del resto)", got)
	}
	if _, ok := payouts.ByBet["bet-carla"]; ok {
		t.Error("Carla no acertó: no debería tener pago")
	}

	if settlement.Pool != 800 || settlement.WinningPool != 300 {
		t.Errorf("pool %d / ganador %d, esperado 800 / 300", settlement.Pool, settlement.WinningPool)
	}
	if settlement.PaidOut != 800 {
		t.Errorf("se pagaron %d de un pool de 800: se perdió el resto", settlement.PaidOut)
	}
}

// A igual monto gana el que apostó antes, y a igual instante el id menor. Sin un
// orden total, la misma carrera liquidada dos veces daría pagos distintos.
func TestPariMutuelDesempataPorAntigüedad(t *testing.T) {
	bets := []Bet{
		bet("bet-z", "zoe", "horse-1", 100, time.Second),
		bet("bet-a", "ana", "horse-1", 100, 0),
		bet("bet-m", "max", "horse-2", 100, 2*time.Second),
	}

	// pool 300, ganador 200 → cada uno 100 × 300 / 200 = 150. Suma 300, sin resto.
	payouts, _ := settleWith(t, bets, "horse-1")
	if payouts.ByBet["bet-a"] != 150 || payouts.ByBet["bet-z"] != 150 {
		t.Fatalf("sin resto los dos cobran 150: ana %d, zoe %d",
			payouts.ByBet["bet-a"], payouts.ByBet["bet-z"])
	}

	// Con un tercero al ganador el reparto deja resto: 400/300 no es exacto.
	bets = append(bets, bet("bet-b", "bruno", "horse-1", 100, 3*time.Second))
	payouts, s := settleWith(t, bets, "horse-1")
	// pool 400, ganador 300 → 100 × 400 / 300 = 133 cada uno. Suma 399, sobra 1.
	// Los tres pusieron 100, así que gana el más antiguo: bet-a.
	if payouts.ByBet["bet-a"] != 134 {
		t.Errorf("el más antiguo se lleva la moneda del resto: ana cobró %d, esperado 134",
			payouts.ByBet["bet-a"])
	}
	if s.PaidOut != s.Pool {
		t.Errorf("no conserva: %d de %d", s.PaidOut, s.Pool)
	}
}

// Nadie le apostó al ganador: se devuelve todo. Quedarse el pool sería inventar una
// casa que este modelo no tiene.
func TestPariMutuelDevuelveSiNadieAcertó(t *testing.T) {
	bets := []Bet{
		bet("bet-1", "ana", "horse-1", 100, 0),
		bet("bet-2", "bruno", "horse-2", 250, time.Second),
	}

	payouts, s := settleWith(t, bets, "horse-9")

	if !payouts.RefundAll {
		t.Fatal("nadie acertó: corresponde devolver")
	}
	if len(payouts.ByBet) != 0 {
		t.Errorf("no debería haber pagos individuales: %v", payouts.ByBet)
	}
	if !s.Refunded || s.WinningPool != 0 || s.PaidOut != s.Pool || s.Pool != 350 {
		t.Errorf("liquidación mal: %+v", s)
	}
}

// Una carrera sin apuestas se liquida igual, en cero. Deja la fila que registra que
// corrió, y no explota en una división por cero.
func TestPariMutuelCarreraSinApuestas(t *testing.T) {
	payouts, s := settleWith(t, nil, "horse-1")

	if !payouts.RefundAll {
		t.Error("sin apuestas el camino es el de devolución")
	}
	if s.Pool != 0 || s.PaidOut != 0 || !s.Refunded {
		t.Errorf("liquidación mal: %+v", s)
	}
}

// Un solo acertante se lleva el pool entero, incluido lo de los que perdieron.
func TestPariMutuelUnSoloGanadorSeLlevaTodo(t *testing.T) {
	bets := []Bet{
		bet("bet-1", "ana", "horse-1", 100, 0),
		bet("bet-2", "bruno", "horse-2", 300, time.Second),
		bet("bet-3", "carla", "horse-2", 600, 2*time.Second),
	}

	payouts, s := settleWith(t, bets, "horse-1")

	if got := payouts.ByBet["bet-1"]; got != 1000 {
		t.Errorf("Ana cobra %d, esperado 1000 (el pool entero)", got)
	}
	if s.PaidOut != 1000 {
		t.Errorf("se pagaron %d de 1000", s.PaidOut)
	}
}

// Las apuestas ya liquidadas NO entran al pool. Si entraran, se pagaría con dinero
// que ya salió y la carrera no conservaría.
func TestPariMutuelIgnoraApuestasYaLiquidadas(t *testing.T) {
	stale := bet("bet-vieja", "zoe", "horse-1", 900, 0)
	stale.Status = BetStatusRefunded

	bets := []Bet{
		stale,
		bet("bet-1", "ana", "horse-1", 100, time.Second),
		bet("bet-2", "bruno", "horse-2", 100, 2*time.Second),
	}

	payouts, s := settleWith(t, bets, "horse-1")

	if s.Pool != 200 {
		t.Errorf("pool %d, esperado 200: la apuesta ya devuelta no cuenta", s.Pool)
	}
	if got := payouts.ByBet["bet-1"]; got != 200 {
		t.Errorf("Ana cobra %d, esperado 200", got)
	}
	if _, ok := payouts.ByBet["bet-vieja"]; ok {
		t.Error("no se le paga a una apuesta ya liquidada")
	}
}

// La conservación en mil carreras al azar. Es la propiedad que sostiene la economía
// entera: el total de monedas del curso solo cambia cuando el instructor regala.
//
// Al azar y no a mano porque el error que se busca —una moneda perdida en el
// truncamiento— aparece con combinaciones raras de montos, no con las redondas que
// uno elige al escribir un test.
func TestPariMutuelConservaSiempre(t *testing.T) {
	rng := rand.New(rand.NewSource(20260730))

	for caso := 0; caso < 1000; caso++ {
		horses := []string{"horse-1", "horse-2", "horse-3", "horse-4"}
		n := 1 + rng.Intn(12)

		var bets []Bet
		for i := 0; i < n; i++ {
			bets = append(bets, bet(
				fmt.Sprintf("bet-%d", i),
				fmt.Sprintf("user-%d", i),
				horses[rng.Intn(len(horses))],
				1+rng.Int63n(999),
				time.Duration(i)*time.Second,
			))
		}
		winner := horses[rng.Intn(len(horses))]

		payouts, s := settleWith(t, bets, winner)

		var paid int64
		for _, amount := range payouts.ByBet {
			paid += amount
		}
		if payouts.RefundAll {
			// La devolución la reparte `finish`, no la regla: acá se comprueba que
			// la liquidación diga que se devolvió el pool entero.
			if s.PaidOut != s.Pool {
				t.Fatalf("caso %d: devolución de %d con pool %d", caso, s.PaidOut, s.Pool)
			}
			continue
		}
		if paid != s.Pool {
			t.Fatalf("caso %d: los pagos suman %d y el pool es %d (ganador %s, %d apuestas)",
				caso, paid, s.Pool, winner, len(bets))
		}
		// Y nadie cobra de menos por el truncamiento: cada ganador se lleva al menos
		// lo que puso, porque el pool nunca es menor que el pool ganador.
		for _, b := range bets {
			if b.HorseID != winner {
				continue
			}
			if payouts.ByBet[b.ID] < b.Amount {
				t.Fatalf("caso %d: %s puso %d y cobra %d",
					caso, b.ID, b.Amount, payouts.ByBet[b.ID])
			}
		}
	}
}
