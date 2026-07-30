import { ChangeDetectionStrategy, Component, inject } from '@angular/core';

import { SHOP_NAME } from './shop.token';
import { CounterComponent } from './counter.component';
import { OrderService } from './order.service';

/**
 * S5 · Dos mostradores y una comanda.
 *
 * Esta pantalla no tiene estado propio: lo pide. Todo lo que muestra sale del
 * mismo `OrderService` que usan los dos mostradores, y ninguno de los tres se
 * pasa nada por `input()`.
 *
 * Es la diferencia con S2 en una línea: **`input()` y `output()` conectan padre
 * e hijo. Un servicio conecta a cualquiera con cualquiera.**
 */
@Component({
  selector: 'app-s05',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [CounterComponent],
  templateUrl: './s05.component.html',
  styleUrl: './s05.component.css',
})
export class S05Component {
  /** Exactamente el mismo objeto que reciben los dos mostradores. */
  protected readonly orders = inject(OrderService);

  protected readonly shopName = inject(SHOP_NAME);
}
