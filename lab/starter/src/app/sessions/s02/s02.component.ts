import { ChangeDetectionStrategy, Component } from '@angular/core';

import { MENU, type Coffee } from './menu';

/**
 * S2 · La carta, con todo adentro de un solo componente.
 *
 * Esto funciona. El problema no es que ande mal: es que **no se puede
 * reutilizar ni un pedazo**. La tarjeta de un café vive dentro del `@for` de
 * esta pantalla, así que para mostrar un café en otro lado hay que copiar y
 * pegar el marcado, y a partir de ahí son dos.
 *
 * El ejercicio de hoy es sacar esa tarjeta a su propio componente y darle las
 * cuatro puertas: `input()`, `input.required()`, `model()` y `output()`.
 *
 * Los lugares que hay que tocar están marcados con `TODO(S2)`.
 */

interface MenuItem {
  readonly coffee: Coffee;
  quantity: number;
}

@Component({
  selector: 'app-s02',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  // TODO(S2): cuando exista <app-coffee-card>, va acá.
  templateUrl: './s02.component.html',
  styleUrl: './s02.component.css',
})
export class S02Component {
  protected readonly items: readonly MenuItem[] = MENU.map((coffee) => ({ coffee, quantity: 1 }));

  /** El café del día se muestra distinto. */
  protected readonly featuredId = 'c1';

  protected orders: readonly string[] = [];

  protected get total(): number {
    return this.orders.length;
  }

  // TODO(S2) · 4 — Estos tres métodos se van con la tarjeta: son decisiones
  // de UNA tarjeta, no de la pantalla. El único que se queda acá es el que
  // escribe en la comanda, que es estado del padre.

  protected add(item: MenuItem, step: number): void {
    const next = item.quantity + step;
    if (next >= 1 && next <= 20) item.quantity = next;
  }

  protected totalFor(item: MenuItem): number {
    return item.coffee.price * item.quantity;
  }

  protected order(item: MenuItem): void {
    if (!item.coffee.available) return;
    this.orders = [
      ...this.orders,
      `${item.quantity} × ${item.coffee.name} · ${item.quantity * item.coffee.price}`,
    ];
  }

  protected clear(): void {
    this.orders = [];
  }
}
