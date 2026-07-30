import { Pipe, type PipeTransform } from '@angular/core';

/**
 * S4 · Una cuota, siempre con dos decimales.
 *
 *   {{ 2.4 | odds }}   → 2,40
 *   {{ 9 | odds }}     → 9,00
 *
 * Sin esto, en la misma columna conviven «9» y «2,4» y la parrilla se lee
 * torcida. Venía resuelto con `horse.odds.toFixed(2)` repetido en cada
 * template desde S1; el pipe lo deja en un solo lugar **y** arregla algo que
 * `toFixed` hacía mal: `toFixed` siempre usa el punto como separador decimal,
 * escriba lo que escriba el resto de la pantalla.
 */
@Pipe({
  name: 'odds',
  standalone: true,
  pure: true,
})
export class OddsPipe implements PipeTransform {
  transform(value: number): string {
    return new Intl.NumberFormat('es', {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(value);
  }
}
