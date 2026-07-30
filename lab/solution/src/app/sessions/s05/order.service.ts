import { Injectable, computed, signal } from '@angular/core';

export interface Order {
  readonly id: number;
  readonly customer: string;
  readonly coffee: string;
}

/**
 * S5 · La comanda compartida.
 *
 * `providedIn: 'root'` quiere decir **una sola instancia para toda la
 * aplicación**. Cualquiera que la pida recibe exactamente el mismo objeto, así
 * que dos componentes que no se conocen entre sí ven la misma comanda.
 *
 * Es lo que en S2 no se podía hacer: ahí los datos bajaban por `input()` y los
 * avisos subían por `output()`, y eso solo funciona entre padre e hijo. Dos
 * hermanos no tienen forma de hablarse — salvo obligando al padre a hacer de
 * intermediario de algo que no le importa.
 *
 * El servicio es ese lugar común.
 *
 * La forma de exponer el estado es la misma de S3, con una vuelta más:
 *
 *   private readonly _orders = signal(…);        se escribe adentro
 *   readonly orders = this._orders.asReadonly(); se lee afuera
 *
 * Así nadie de afuera puede llamar a `set` ni a `update`: para cambiar la
 * comanda hay que pasar por un método, y los métodos son el contrato.
 */
@Injectable({ providedIn: 'root' })
export class OrderService {
  private readonly _orders = signal<readonly Order[]>([]);
  private nextId = 1;

  /** Solo lectura hacia afuera. Es la mitad del diseño. */
  readonly orders = this._orders.asReadonly();

  readonly count = computed(() => this.orders().length);

  readonly lastCustomer = computed(() => this.orders().at(-1)?.customer ?? '');

  add(customer: string, coffee: string): void {
    const name = customer.trim();
    if (name === '') return;

    this._orders.update((orders) => [...orders, { id: this.nextId++, customer: name, coffee }]);
  }

  remove(id: number): void {
    this._orders.update((orders) => orders.filter((order) => order.id !== id));
  }

  clear(): void {
    this._orders.set([]);
  }
}
