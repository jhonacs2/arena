import { Pipe, type PipeTransform } from '@angular/core';

/**
 * Cuotas. Llegan **×100 en entero**: `340` es 3,40 (`arena/CLAUDE.md` §5).
 *
 * El formateo es aritmética de enteros y concatenación, a propósito. Dividir por
 * 100 para mostrar parece inofensivo hasta que alguien reusa ese número para
 * calcular un pago: acá el `number` con coma no existe en ningún momento.
 */
@Pipe({ name: 'odds' })
export class OddsPipe implements PipeTransform {
  transform(value: number): string {
    const whole = Math.trunc(value / 100);
    const cents = String(Math.abs(value) % 100).padStart(2, '0');
    return `${whole},${cents}`;
  }
}

/** `payout = amount * oddsAtBet / 100`, división entera, redondeo hacia abajo. */
export const payoutOf = (amount: number, oddsAtBet: number): number =>
  Math.floor((amount * oddsAtBet) / 100);

/** Puntos = `floor(saldo / 100)`. Una función del saldo, nunca una columna. */
export const pointsOf = (balance: number): number => Math.floor(balance / 100);
