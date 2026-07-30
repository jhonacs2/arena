import { Pipe, type PipeTransform } from '@angular/core';

/**
 * S4 · Un importe, formateado.
 *
 *   {{ 1234 | money }}        → $ 1.234
 *   {{ 1234 | money:'USD' }}  → USD 1.234
 *
 * Un pipe es **una función con nombre que se puede usar en el template**. Nada
 * más que eso: recibe un valor, devuelve otro, y no sabe quién lo llamó.
 *
 * Por eso no necesita saber qué es un café ni qué es una carrera: sirve para
 * cualquier número que haya que mostrar como dinero.
 *
 * `pure: true` es el valor por defecto y está escrito a propósito: significa
 * que Angular vuelve a llamarlo **solo cuando cambia el valor de entrada**. Un
 * pipe impuro corre en cada detección de cambios, y eso se ve en el bloque de
 * predicciones.
 */
@Pipe({
  name: 'money',
  standalone: true,
  pure: true,
})
export class MoneyPipe implements PipeTransform {
  transform(value: number, symbol = '$'): string {
    // `Intl` es del navegador, no de Angular. El pipe solo le pone nombre y lo
    // hace reutilizable desde cualquier template.
    //
    // `useGrouping: true` está escrito a mano por una razón concreta: en
    // español, el valor por defecto **no agrupa los números de cuatro cifras**
    // —4200 se escribe «4200» y no «4.200»—. Para un número suelto es lo
    // correcto; para un importe, no. Es el tipo de detalle que un pipe deja
    // resuelto en un solo lugar en vez de en veinte templates.
    const formatted = new Intl.NumberFormat('es', {
      maximumFractionDigits: 0,
      useGrouping: true,
    }).format(value);

    return `${symbol} ${formatted}`;
  }
}
