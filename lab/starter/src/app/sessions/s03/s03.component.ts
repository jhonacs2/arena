import { ChangeDetectionStrategy, Component } from '@angular/core';

import { INITIAL_ORDERS, STATUS_LABELS, lineTotal, nextStatus, type Order } from './orders';

/**
 * S3 · El tablero de la comanda, sin signals.
 *
 * Esto funciona: la lista se ve, los botones andan, y los cambios se hacen sin
 * mutar —`map`, `filter`, spread—, como venimos haciendo desde S1.
 *
 * Lo que no tiene es **nada derivado**. No hay filtro, no hay búsqueda, no hay
 * contadores y no hay total. Y cada vez que quieras agregar uno vas a tener
 * dos opciones malas: guardarlo en otra propiedad y acordarte de actualizarla
 * en los cuatro lugares donde cambia la comanda, o calcularlo con un método
 * que corre en **cada** detección de cambios.
 *
 * `computed()` es la tercera opción, y es la clase de hoy.
 *
 * Los lugares que hay que tocar están marcados con `TODO(S3)`.
 */
@Component({
  selector: 'app-s03',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  // TODO(S3) · 3 — el buscador necesita FormsModule, como en S1.
  templateUrl: './s03.component.html',
  styleUrl: './s03.component.css',
})
export class S03Component {
  protected readonly statusLabels = STATUS_LABELS;
  protected readonly lineTotal = lineTotal;

  /**
   * TODO(S3) · 2 — La comanda es una propiedad común: cambia, y la pantalla se
   * entera porque el clic que la cambió pasó por el template de este mismo
   * componente. Nadie avisó nada; se revisó todo por las dudas.
   *
   * Convertila en un `signal`, y `advance`, `remove` y `add` en `update`.
   */
  protected orders: readonly Order[] = INITIAL_ORDERS;

  // TODO(S3) · 3 — acá van los signals del filtro y de la búsqueda.

  // TODO(S3) · 4 — y acá los computed: lo que se ve, los contadores, el total
  // por cobrar y la lista ordenada por precio.

  protected advance(id: string): void {
    this.orders = this.orders.map((order) =>
      order.id === id ? { ...order, status: nextStatus(order.status) } : order,
    );
  }

  protected remove(id: string): void {
    this.orders = this.orders.filter((order) => order.id !== id);
  }

  protected add(): void {
    const number = this.orders.length + 1;
    this.orders = [
      ...this.orders,
      {
        id: `o${number}-${this.orders.length}`,
        customer: `Cliente ${number}`,
        coffee: 'Yirgacheffe',
        quantity: 1,
        price: 42,
        status: 'pending' as const,
      },
    ];
  }

  protected reset(): void {
    this.orders = INITIAL_ORDERS;
  }
}
