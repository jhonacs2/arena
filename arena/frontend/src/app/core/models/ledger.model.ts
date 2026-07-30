/**
 * El ledger es **append-only** (`decisiones.md` §4): no se edita ni se borra un
 * movimiento, se compensa con otro. Es lo que le permite al alumno entender por
 * qué tiene la nota que tiene.
 */
export type LedgerReason =
  | 'code_redeemed'
  | 'bet_placed'
  | 'bet_won'
  | 'bet_refunded'
  | 'gift'
  | 'adjustment';

export interface LedgerEntry {
  readonly id: number;
  /** Monedas, entero con signo. Positivo entra, negativo sale. */
  readonly delta: number;
  readonly reason: LedgerReason;
  readonly balanceAfter: number;
  readonly createdAt: string;
  readonly raceName?: string;
  readonly note?: string;
}

/** El texto que ve el alumno para cada motivo. En castellano, como todo lo visible. */
export const LEDGER_REASON_LABEL: Readonly<Record<LedgerReason, string>> = {
  code_redeemed: 'Código canjeado',
  bet_placed: 'Apuesta',
  bet_won: 'Apuesta ganada',
  bet_refunded: 'Apuesta devuelta',
  gift: 'Regalo del instructor',
  adjustment: 'Ajuste',
};
