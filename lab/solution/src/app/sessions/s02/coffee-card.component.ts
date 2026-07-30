import {
  ChangeDetectionStrategy,
  Component,
  input,
  model,
  output,
  type OnChanges,
  type OnDestroy,
  type OnInit,
} from '@angular/core';

import type { Coffee, OrderRequest } from './menu';

/**
 * S2 · El hijo.
 *
 * Todo lo que este componente sabe del mundo entra por sus `input()`, y todo
 * lo que tiene para decir sale por su `output()`. No importa `MENU`, no sabe
 * qué es una comanda y no sabe cuántas tarjetas hay: por eso se puede usar en
 * cualquier lado.
 *
 * Las cuatro puertas del componente:
 *
 *   input()          entra un dato, de solo lectura
 *   input.required() entra un dato, y sin él no compila
 *   model()          entra y sale: two-way binding
 *   output()         sale un aviso
 *
 * Y una quinta que no es un dato sino marcado: `<ng-content>`, donde el padre
 * mete HTML propio.
 */
@Component({
  selector: 'app-coffee-card',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  templateUrl: './coffee-card.component.html',
  styleUrl: './coffee-card.component.css',
})
export class CoffeeCardComponent implements OnInit, OnChanges, OnDestroy {
  /**
   * Sin este dato la tarjeta no tiene sentido, y `required` lo hace obligatorio
   * **en tiempo de compilación**: usar `<app-coffee-card />` sin `[coffee]` no
   * compila. Es el mismo tipo de promesa que los tipos de S0.
   */
  readonly coffee = input.required<Coffee>();

  /** Opcional, con valor por defecto. Sin `[featured]`, vale `false`. */
  readonly featured = input(false);

  /**
   * `model()` es `input()` y `output()` a la vez: el padre escribe con
   * `[(quantity)]` y el hijo puede cambiarlo con `quantity.set(…)`.
   */
  readonly quantity = model(1);

  /** El aviso hacia afuera. El padre decide qué hacer con él. */
  readonly ordered = output<OrderRequest>();

  /** Cuántas veces cambió alguno de los inputs. Se ve en la tarjeta. */
  protected changes = 0;

  /** Se llena en `ngOnInit`, no en la declaración. Ver el porqué abajo. */
  protected mountedAt = '';

  /** Avisa al padre cuando la tarjeta se va de la pantalla. */
  readonly destroyed = output<string>();

  /**
   * `ngOnInit` corre **una vez**, después de que Angular llenó los inputs.
   *
   * Por eso el nombre del café no se puede leer en el constructor: ahí todavía
   * no llegó nada.
   */
  ngOnInit(): void {
    this.mountedAt = new Intl.DateTimeFormat('es', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    }).format(new Date());
  }

  /** Corre cada vez que el padre cambia el valor de un input. */
  ngOnChanges(): void {
    this.changes += 1;
  }

  /** Corre cuando el componente se destruye. Es el lugar de soltar cosas. */
  ngOnDestroy(): void {
    this.destroyed.emit(this.coffee().name);
  }

  protected get total(): number {
    return this.coffee().price * this.quantity();
  }

  protected add(step: number): void {
    const next = this.quantity() + step;
    if (next >= 1 && next <= 20) this.quantity.set(next);
  }

  protected order(): void {
    if (!this.coffee().available) return;
    this.ordered.emit({ coffee: this.coffee(), quantity: this.quantity() });
  }
}
