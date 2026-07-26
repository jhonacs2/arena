/** Apuestas y tabla de posiciones — `docs/contract/openapi.yaml`. */

export type BetStatus = 'pending' | 'won' | 'lost';

/**
 * `raceName` y `horseName` vienen denormalizados a propósito: el historial se
 * pinta sin ir a buscar la carrera ni el caballo. Es una decisión del contrato
 * (`docs/contract/README.md`), no un descuido del backend.
 */
export interface Bet {
  readonly id: string;
  readonly userId: string;
  readonly raceId: string;
  readonly raceName: string;
  readonly horseId: string;
  readonly horseName: string;
  readonly amount: number;
  /** Congelada al momento de apostar. */
  readonly odds: number;
  readonly status: BetStatus;
  /** `0` mientras esté pendiente o si perdió. */
  readonly payout: number;
  readonly placedAt: string;
}

/** Respuesta de `POST /bets`: la apuesta y el saldo ya descontado. */
export interface BetCreated {
  readonly bet: Bet;
  readonly balance: number;
}

export type LeaderboardPeriod = 'daily' | 'all';

export interface LeaderboardEntry {
  readonly rank: number;
  readonly userId: string;
  readonly displayName: string;
  /** Pagos menos montos apostados. Puede ser negativo. */
  readonly profit: number;
  /** Apuestas liquidadas en el período. */
  readonly bets: number;
  readonly wins: number;
}

export interface Leaderboard {
  readonly period: LeaderboardPeriod;
  readonly entries: readonly LeaderboardEntry[];
}

/** Límites del dominio, del contrato. Los usa el validador de S8. */
export const MIN_BET_AMOUNT = 10;
export const MAX_BET_AMOUNT = 5000;

/** Lo que se cobra si gana. El backend redondea igual. */
export function potentialPayout(amount: number, odds: number): number {
  return Math.round(amount * odds);
}
