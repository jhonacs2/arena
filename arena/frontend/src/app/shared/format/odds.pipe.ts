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

/**
 * Acá **no** hay un `payoutOf`, y su ausencia es a propósito.
 *
 * Existió mientras la economía era de cuota fija: `amount * oddsAtBet / 100`.
 * Con pari-mutuel (`decisiones.md` §1) el pago sale de repartir el pozo, y no se
 * puede calcular en el cliente: depende de cuánto se apostó en total y de cuántos
 * acertaron, y el servidor **tapa a qué caballo apostó cada uno** mientras la
 * carrera está abierta. Cualquier función que devolviera un número acá estaría
 * inventando una promesa que el backend no va a cumplir.
 *
 * El pago real llega liquidado, en `bet.payout`.
 */

/** Puntos = `max(10, floor(saldo / 100))`. El piso está en `decisiones.md` §1. */
export const pointsOf = (balance: number): number => Math.max(10, Math.floor(balance / 100));
