import { ChangeDetectionStrategy, Component } from '@angular/core';

import { CoffeeCardComponent } from './coffee-card.component';
import { MENU, type Coffee, type OrderRequest } from './menu';

/**
 * S2 · El padre.
 *
 * Acá vive el estado —la carta y la comanda— y las decisiones. Las tarjetas no
 * saben nada de esto: reciben un café y avisan cuando alguien pide.
 *
 * La regla que explica la sesión entera:
 *
 *   **Los datos bajan por `input()`. Los avisos suben por `output()`.**
 *
 * El hijo nunca modifica lo que le prestaron. Pide, y el padre decide.
 */

/** Un café de la carta con la cantidad que el usuario eligió en su tarjeta. */
interface MenuItem {
  readonly coffee: Coffee;
  /** No es `readonly`: `[(quantity)]` escribe acá desde el hijo. */
  quantity: number;
}

/** El café que se puede sacar de la carta, para ver el desmontaje. */
const LAST_ID = 'c4';

@Component({
  selector: 'app-s02',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [CoffeeCardComponent],
  templateUrl: './s02.component.html',
  styleUrl: './s02.component.css',
})
export class S02Component {
  protected readonly items: readonly MenuItem[] = MENU.map((coffee) => ({ coffee, quantity: 1 }));

  /** El café del día se muestra distinto. Es un `input()` opcional del hijo. */
  protected readonly featuredId = 'c1';

  /** La comanda, en texto. Se arma con lo que emiten las tarjetas. */
  protected orders: readonly string[] = [];

  /** Lo que avisaron las tarjetas al destruirse. */
  protected farewells: readonly string[] = [];

  /** Controla si la última tarjeta existe, para poder verla desmontarse. */
  protected showLast = true;

  /** Lo que se dibuja. Sacar un café de acá es lo que dispara `ngOnDestroy`. */
  protected get visibleItems(): readonly MenuItem[] {
    return this.showLast ? this.items : this.items.filter((item) => item.coffee.id !== LAST_ID);
  }

  protected get total(): number {
    return this.orders.length;
  }

  /**
   * Escucha el `output()` del hijo.
   *
   * `$event` es exactamente lo que el hijo pasó a `emit()`, con su tipo. No es
   * un evento del DOM: es un dato del componente.
   */
  protected take(request: OrderRequest): void {
    this.orders = [
      ...this.orders,
      `${request.quantity} × ${request.coffee.name} · ${request.quantity * request.coffee.price}`,
    ];
  }

  protected sayGoodbye(name: string): void {
    this.farewells = [...this.farewells, name];
  }

  protected clear(): void {
    this.orders = [];
  }

  protected toggleLast(): void {
    this.showLast = !this.showLast;
  }
}
