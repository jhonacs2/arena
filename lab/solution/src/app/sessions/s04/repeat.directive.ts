import { Directive, effect, inject, input, TemplateRef, ViewContainerRef } from '@angular/core';

/**
 * S4 · Repite un pedazo de template N veces.
 *
 *   <span *appRepeat="coffee.rating">●</span>
 *
 * Esta es una **directiva estructural**: no cambia un elemento, decide **si
 * existe y cuántas veces**. El asterisco es la pista.
 *
 * El `*` es azúcar sintáctica. Estas dos líneas son exactamente lo mismo:
 *
 *   <span *appRepeat="3">●</span>
 *
 *   <ng-template [appRepeat]="3">
 *     <span>●</span>
 *   </ng-template>
 *
 * Un `<ng-template>` es un pedazo de HTML que **no se dibuja**: queda guardado
 * y alguien decide después si se pinta, cuántas veces y dónde. Eso es lo que
 * hacían `*ngIf` y `*ngFor` antes de que llegaran `@if` y `@for`.
 *
 * > En Angular 18 casi nunca hace falta escribir una directiva estructural
 * > propia: el control flow de S3 cubre los casos comunes y se lee mucho mejor.
 * > Está aquí para poder leer código que las usa, que es todavía la mayoría.
 */
@Directive({
  selector: '[appRepeat]',
  standalone: true,
})
export class RepeatDirective {
  /** Cuántas veces se dibuja el template. */
  readonly appRepeat = input.required<number>();

  /** El pedazo de HTML guardado, sin dibujar. */
  private readonly template = inject(TemplateRef<unknown>);

  /** El lugar del DOM donde se van a insertar las copias. */
  private readonly container = inject(ViewContainerRef);

  constructor() {
    // `effect` corre cuando cambia algún signal que lee. Es la pieza de signals
    // que S3 no llegó a ver: sirve cuando hace falta que algo **pase**, no que
    // se calcule.
    effect(() => {
      const times = Math.max(0, Math.trunc(this.appRepeat()));
      this.container.clear();
      for (let index = 0; index < times; index += 1) {
        this.container.createEmbeddedView(this.template);
      }
    });
  }
}
