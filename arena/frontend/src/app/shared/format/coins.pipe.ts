import { Pipe, type PipeTransform } from '@angular/core';

/**
 * Monedas. Siempre enteras (`arena/CLAUDE.md` §5).
 *
 * El separador de miles se agrupa a mano y no con `Intl`: agrupar es lo único
 * que hace falta, y así no queda ninguna puerta abierta a que a alguien se le
 * ocurra pasarle un decimal y que «funcione».
 */
function group(value: number): string {
  const digits = String(Math.abs(Math.trunc(value)));
  let out = '';
  for (let i = 0; i < digits.length; i++) {
    if (i > 0 && (digits.length - i) % 3 === 0) out += '.';
    out += digits[i];
  }
  return out;
}

@Pipe({ name: 'coins' })
export class CoinsPipe implements PipeTransform {
  transform(value: number): string {
    return group(value);
  }
}

/**
 * El mismo número con su signo, para el ledger.
 *
 * Usa el signo menos de verdad (U+2212) y no el guion del teclado: en cifras
 * tabulares el guion queda alto y angosto, y una columna de movimientos con
 * guiones se ve como una lista de viñetas.
 */
@Pipe({ name: 'signedCoins' })
export class SignedCoinsPipe implements PipeTransform {
  transform(value: number): string {
    const sign = value < 0 ? '−' : '+';
    return `${sign}${group(value)}`;
  }
}
