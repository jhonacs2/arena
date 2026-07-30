import { Pipe, type PipeTransform } from '@angular/core';

/**
 * S4 · Un importe del hipódromo.
 *
 *   {{ 1250 | money }}        → 1.250 pts
 *   {{ 1250 | money:'' }}     → 1.250
 *
 * El saldo es virtual y entero: no hay centavos y no hay moneda real, así que
 * no se usa `| currency`. La unidad se llama «pts» en toda la aplicación, y
 * este pipe es el único lugar donde ese nombre está escrito.
 *
 * Es puro: Angular lo llama solo cuando cambia el importe.
 */
@Pipe({
  name: 'money',
  standalone: true,
  pure: true,
})
export class MoneyPipe implements PipeTransform {
  transform(value: number, unit = 'pts'): string {
    // `useGrouping: true` a propósito: en español, el valor por defecto no
    // agrupa los números de cuatro cifras. Para un importe sí queremos «1.250».
    const formatted = new Intl.NumberFormat('es', {
      maximumFractionDigits: 0,
      useGrouping: true,
    }).format(value);

    return unit === '' ? formatted : `${formatted} ${unit}`;
  }
}
