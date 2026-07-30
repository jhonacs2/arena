import { Directive, input } from '@angular/core';

/**
 * S4 · Resalta un elemento cuando se cumple una condición.
 *
 *   <li [appHighlight]="item.featured">…</li>
 *   <li [appHighlight]="true" highlightLabel="Recomendado">…</li>
 *
 * Una **directiva de atributo** cambia cómo se ve o cómo se comporta un
 * elemento que ya existe. No dibuja nada nuevo: se cuelga de una etiqueta y le
 * agrega algo.
 *
 * La diferencia con un componente, en una línea:
 *
 *   Un componente **trae su propio template**. Una directiva **no tiene
 *   template**: le agrega comportamiento a un elemento ajeno.
 *
 * `host` reemplaza a los decoradores `@HostBinding` y `@HostListener`: es la
 * forma recomendada en Angular 18 y deja todo el contrato del elemento en un
 * solo lugar, a la vista.
 */
@Directive({
  selector: '[appHighlight]',
  standalone: true,
  host: {
    '[class.is-highlighted]': 'appHighlight()',
    // Un dato para el lector de pantalla: sin esto el resaltado es solo color,
    // y quien no ve el color no se entera de nada.
    '[attr.data-highlight-label]': 'appHighlight() ? highlightLabel() : null',
  },
})
export class HighlightDirective {
  /**
   * El input se llama igual que el selector, así se puede escribir
   * `[appHighlight]="condición"` en una sola vez en lugar de repetir el nombre.
   */
  readonly appHighlight = input(false);

  /** Lo que se anuncia cuando está resaltado. */
  readonly highlightLabel = input('Destacado');
}
