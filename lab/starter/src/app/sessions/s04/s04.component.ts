import { ChangeDetectionStrategy, Component, computed, signal } from '@angular/core';

import { CALL_COUNT, CountImpurePipe, CountPurePipe, resetCallCount } from './call-count.pipe';

/**
 * S4 · La carta, con todo el formateo adentro del componente.
 *
 * Esto funciona. El problema es dónde vive cada cosa:
 *
 *   - `formatMoney()` es un método de ESTE componente. La pantalla de al lado
 *     que también muestre precios va a tener que copiarlo.
 *   - `beansFor()` arma un array de longitud N para poder dibujar N puntos con
 *     un `@for`. Es un rodeo: no hay ninguna lista, hay una cantidad.
 *   - La clase del café del día se decide en el template, con la condición
 *     repetida en dos atributos.
 *
 * Ninguna de las tres es un problema del componente: son formas de mostrar, y
 * tienen que poder usarse desde cualquier template.
 *
 * Los lugares que hay que tocar están marcados con `TODO(S4)`.
 */

interface Coffee {
  readonly id: string;
  readonly name: string;
  readonly origin: string;
  readonly price: number;
  readonly rating: number;
  readonly stock: number;
}

const MENU: readonly Coffee[] = [
  { id: 'c1', name: 'Yirgacheffe', origin: 'Etiopía', price: 4200, rating: 5, stock: 12 },
  { id: 'c2', name: 'Huila', origin: 'Colombia', price: 3800, rating: 4, stock: 3 },
  { id: 'c3', name: 'Cerrado', origin: 'Brasil', price: 3000, rating: 3, stock: 0 },
  { id: 'c4', name: 'Antigua', origin: 'Guatemala', price: 4500, rating: 4, stock: 7 },
];

const TOTAL_STOCK = MENU.reduce((sum, coffee) => sum + coffee.stock, 0);

@Component({
  selector: 'app-s04',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  // TODO(S4) · aquí van el pipe y las dos directivas cuando existan.
  imports: [CountPurePipe, CountImpurePipe],
  templateUrl: './s04.component.html',
  styleUrl: './s04.component.css',
})
export class S04Component {
  protected readonly menu = MENU;
  protected readonly featuredId = 'c1';
  protected readonly totalStock = TOTAL_STOCK;

  protected readonly clicks = signal(0);

  protected readonly counts = computed(() => {
    this.clicks();
    return { pure: CALL_COUNT.pure, impure: CALL_COUNT.impure };
  });

  /**
   * TODO(S4) · 2 — Esto no es lógica del componente: es una forma de mostrar
   * un número. Tiene que poder usarse desde cualquier template sin copiarlo.
   */
  protected formatMoney(value: number, symbol = '$'): string {
    const formatted = new Intl.NumberFormat('es', {
      maximumFractionDigits: 0,
      useGrouping: true,
    }).format(value);

    return `${symbol} ${formatted}`;
  }

  /**
   * TODO(S4) · 4 — Un array de longitud N, solo para poder recorrerlo con un
   * `@for` y dibujar N puntos. No hay ninguna lista: hay una cantidad.
   */
  protected beansFor(rating: number): readonly number[] {
    return Array.from({ length: Math.max(0, Math.trunc(rating)) }, (_, index) => index);
  }

  protected share(coffee: Coffee): number {
    return this.totalStock === 0 ? 0 : coffee.stock / this.totalStock;
  }

  protected tick(): void {
    this.clicks.update((value) => value + 1);
  }

  protected resetCounters(): void {
    resetCallCount();
    this.clicks.update((value) => value + 1);
  }
}
