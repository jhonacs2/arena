import { ChangeDetectionStrategy, Component, computed, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';

export interface Order {
  readonly id: number;
  readonly customer: string;
  readonly coffee: string;
}

/**
 * S5 · Un solo mostrador, con todo adentro.
 *
 * Esto funciona, y tiene un techo: **la comanda vive dentro de este
 * componente**. El día que haya un segundo mostrador —o un panel de cocina, o
 * una pantalla de retiro— no hay forma de que vean la misma comanda.
 *
 * Con lo de S2 tampoco alcanza: `input()` y `output()` conectan padre e hijo, y
 * dos mostradores uno al lado del otro son hermanos. Obligar al padre a hacer
 * de intermediario de algo que no le importa es la solución que se toma cuando
 * no se conoce la de hoy.
 *
 * Los lugares que hay que tocar están marcados con `TODO(S5)`.
 */
@Component({
  selector: 'app-s05',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [FormsModule],
  templateUrl: './s05.component.html',
  styleUrl: './s05.component.css',
})
export class S05Component {
  /**
   * TODO(S5) · 2 — Toda esta parte se va a un servicio con
   * `providedIn: 'root'`, para que la comanda sea una sola en toda la
   * aplicación y cualquiera pueda pedirla.
   */
  private readonly _orders = signal<readonly Order[]>([]);
  private nextId = 1;

  protected readonly orders = this._orders.asReadonly();
  protected readonly count = computed(() => this.orders().length);
  protected readonly lastCustomer = computed(() => this.orders().at(-1)?.customer ?? '');

  /** TODO(S5) · 4 — El nombre del café no es un servicio: es configuración. */
  protected readonly shopName = 'Café Compilado';

  protected readonly coffees = ['Yirgacheffe', 'Huila', 'Cerrado'] as const;

  protected readonly customer = signal('');
  protected readonly coffee = signal<string>('Yirgacheffe');

  /**
   * TODO(S5) · 3 — Y el cuaderno se va a un servicio SIN `providedIn`, provisto
   * en el mostrador, para que cada uno tenga el suyo.
   */
  protected readonly notes = signal<readonly string[]>([]);
  protected readonly note = signal('');

  protected take(): void {
    const name = this.customer().trim();
    if (name === '') return;

    this._orders.update((orders) => [
      ...orders,
      { id: this.nextId++, customer: name, coffee: this.coffee() },
    ]);
    this.customer.set('');
  }

  protected remove(id: number): void {
    this._orders.update((orders) => orders.filter((order) => order.id !== id));
  }

  protected clear(): void {
    this._orders.set([]);
  }

  protected jot(): void {
    const text = this.note().trim();
    if (text === '') return;

    this.notes.update((notes) => [...notes, text]);
    this.note.set('');
  }

  protected erase(): void {
    this.notes.set([]);
  }
}
