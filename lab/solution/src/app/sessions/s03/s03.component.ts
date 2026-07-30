import { ChangeDetectionStrategy, Component, computed, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';

import {
  FILTER_LABELS,
  INITIAL_ORDERS,
  STATUS_LABELS,
  lineTotal,
  nextStatus,
  type Order,
  type OrderFilter,
} from './orders';

/**
 * S3 · El tablero de la comanda.
 *
 * Tres ideas, y las tres se ven en este archivo:
 *
 *   signal()    un valor que AVISA cuando cambia
 *   computed()  un valor DERIVADO de otros, que se recalcula solo
 *   inmutable   nunca se modifica lo que había; se pone algo nuevo
 *
 * Y la regla que las une, que es la que hay que poder explicar:
 *
 *   **La vista se repinta porque el signal avisó, no porque alguien la mandó.**
 *   Si modificás el array que había adentro, el signal no se entera: para él
 *   sigue siendo el mismo array. Y la pantalla se queda vieja.
 */
@Component({
  selector: 'app-s03',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [FormsModule],
  templateUrl: './s03.component.html',
  styleUrl: './s03.component.css',
})
export class S03Component {
  protected readonly statusLabels = STATUS_LABELS;
  protected readonly filters: readonly OrderFilter[] = ['all', 'pending', 'ready', 'served'];
  protected readonly filterLabels = FILTER_LABELS;
  protected readonly lineTotal = lineTotal;

  // ── El estado. Tres signals y nada más ──────────────────────────────────

  /** La comanda entera. `readonly Order[]` para que no se pueda hacer push. */
  private readonly orders = signal<readonly Order[]>(INITIAL_ORDERS);

  protected readonly filter = signal<OrderFilter>('all');
  protected readonly query = signal('');

  // ── Lo derivado. Nada de esto se guarda: se calcula ─────────────────────

  /**
   * Lo que se ve, después del filtro y la búsqueda.
   *
   * `computed()` se recalcula **solo cuando cambia alguno de los signals que
   * lee**, y guarda el resultado hasta entonces. Por eso esto puede vivir en
   * el template sin costo: no es un método que corre en cada repintado.
   */
  protected readonly visible = computed<readonly Order[]>(() => {
    const status = this.filter();
    const text = this.query().trim().toLowerCase();

    return this.orders()
      .filter((order) => status === 'all' || order.status === status)
      .filter((order) => text === '' || `${order.customer} ${order.coffee}`.toLowerCase().includes(text));
  });

  /** Cuántas hay de cada estado. Se recalcula solo cuando cambia la comanda. */
  protected readonly counts = computed(() => {
    const orders = this.orders();
    return {
      all: orders.length,
      pending: orders.filter((order) => order.status === 'pending').length,
      ready: orders.filter((order) => order.status === 'ready').length,
      served: orders.filter((order) => order.status === 'served').length,
    };
  });

  /** Lo que hay que cobrar por lo que todavía no se entregó. */
  protected readonly pendingTotal = computed(() =>
    this.orders()
      .filter((order) => order.status !== 'served')
      .reduce((sum, order) => sum + lineTotal(order), 0),
  );

  /** Las más caras primero, sin tocar el orden de la comanda. */
  protected readonly byPrice = computed<readonly Order[]>(() =>
    // `[...]` primero: sort() ordena EN EL LUGAR y devolvería el mismo array
    // que está adentro del signal. Ordenarlo sería mutarlo.
    [...this.visible()].sort((a, b) => lineTotal(b) - lineTotal(a)),
  );

  // ── Los cambios. Siempre un valor NUEVO ─────────────────────────────────

  /**
   * `update` recibe lo que hay y devuelve lo que va a haber.
   *
   * `map` devuelve un array nuevo, y el `{ ...order }` un objeto nuevo. Nada
   * de lo que estaba se modifica: se reemplaza.
   */
  protected advance(id: string): void {
    this.orders.update((orders) =>
      orders.map((order) => (order.id === id ? { ...order, status: nextStatus(order.status) } : order)),
    );
  }

  protected remove(id: string): void {
    this.orders.update((orders) => orders.filter((order) => order.id !== id));
  }

  /** Una comanda nueva al final. Nunca `push`. */
  protected add(): void {
    const number = this.orders().length + 1;
    this.orders.update((orders) => [
      ...orders,
      {
        id: `o${number}-${orders.length}`,
        customer: `Cliente ${number}`,
        coffee: 'Yirgacheffe',
        quantity: 1,
        price: 42,
        status: 'pending' as const,
      },
    ]);
  }

  /** `set` reemplaza el valor entero, sin mirar el anterior. */
  protected reset(): void {
    this.orders.set(INITIAL_ORDERS);
    this.filter.set('all');
    this.query.set('');
  }

  protected selectFilter(filter: OrderFilter): void {
    this.filter.set(filter);
  }
}
