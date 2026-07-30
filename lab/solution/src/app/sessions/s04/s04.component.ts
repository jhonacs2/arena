import { ChangeDetectionStrategy, Component, computed, signal } from '@angular/core';
import { DecimalPipe, PercentPipe, UpperCasePipe } from '@angular/common';

import { CALL_COUNT, CountImpurePipe, CountPurePipe, resetCallCount } from './call-count.pipe';
import { HighlightDirective } from './highlight.directive';
import { MoneyPipe } from './money.pipe';
import { RepeatDirective } from './repeat.directive';

/**
 * S4 · La carta, con pipes y directivas.
 *
 * Lo que hay que mirar en este archivo es **lo que no está**: no hay ni una
 * función que formatee dinero, ni una que arme la clase del café del día, ni
 * un `toFixed`. Todo eso se fue a piezas que se usan desde el template y que
 * sirven en cualquier otra pantalla.
 *
 *   Un **pipe** transforma lo que se ve.
 *   Una **directiva** cambia cómo se comporta un elemento.
 *
 * Ninguno de los dos toca el componente.
 */

interface Coffee {
  readonly id: string;
  readonly name: string;
  readonly origin: string;
  readonly price: number;
  /** De 0 a 5. La dibuja la directiva estructural. */
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
  imports: [
    // Los pipes y las directivas van en `imports`, igual que los componentes.
    // Es la misma regla desde S1: si el template lo usa, se declara.
    UpperCasePipe,
    DecimalPipe,
    PercentPipe,
    MoneyPipe,
    HighlightDirective,
    RepeatDirective,
    CountPurePipe,
    CountImpurePipe,
  ],
  templateUrl: './s04.component.html',
  styleUrl: './s04.component.css',
})
export class S04Component {
  protected readonly menu = MENU;
  protected readonly featuredId = 'c1';
  protected readonly totalStock = TOTAL_STOCK;

  /** Solo para que el usuario pueda provocar detecciones de cambios. */
  protected readonly clicks = signal(0);

  protected readonly counts = computed(() => {
    // Se lee `clicks()` para que este computed se recalcule con cada clic y la
    // pantalla muestre los contadores al día.
    this.clicks();
    return { pure: CALL_COUNT.pure, impure: CALL_COUNT.impure };
  });

  protected tick(): void {
    this.clicks.update((value) => value + 1);
  }

  protected resetCounters(): void {
    resetCallCount();
    this.clicks.update((value) => value + 1);
  }

  protected share(coffee: Coffee): number {
    return this.totalStock === 0 ? 0 : coffee.stock / this.totalStock;
  }
}
