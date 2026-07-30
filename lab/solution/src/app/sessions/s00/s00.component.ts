import { ChangeDetectionStrategy, Component } from '@angular/core';

import { MENU, cheapest, describe, firstPage, orderTotal, priceFor } from './menu';
import type { CoffeeSize, Order, OrderLine } from './menu';

/**
 * S0 · El pizarrón de Café Compilado.
 *
 * ⚠️ **Este componente viene dado y no es el ejercicio de hoy.** Los
 * componentes se ven en S1 y el `@for` en S3; acá está para que se vea en
 * pantalla lo que hacen las funciones de `menu.ts`, que sí es el material de
 * esta sesión.
 *
 * Todo lo que muestra sale de `menu.ts`. Si cambiás un tipo allá y algo deja
 * de compilar, el error va a aparecer acá: eso también es parte de la clase.
 */
@Component({
  selector: 'app-s00',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  templateUrl: './s00.component.html',
  styleUrl: './s00.component.css',
})
export class S00Component {
  /** Los tres tamaños, para armar la tabla de precios. */
  protected readonly sizes: readonly CoffeeSize[] = ['chico', 'mediano', 'grande'];

  /** La primera página del menú: tres cafés de los cuatro que hay. */
  protected readonly page = firstPage(MENU, 3);

  /** El menú entero, para la tabla de precios. */
  protected readonly menu = MENU;

  /** Puede no haber: con un menú vacío no hay «el más barato». */
  protected readonly best = cheapest(MENU);

  protected readonly order: Order = {
    customer: 'Ana',
    lines: MENU.slice(0, 2).map(
      (coffee, index): OrderLine => ({
        coffee,
        size: index === 0 ? 'grande' : 'chico',
        quantity: index === 0 ? 2 : 1,
      }),
    ),
  };

  protected readonly total = orderTotal(this.order);

  /** Las funciones de `menu.ts`, expuestas para poder llamarlas del template. */
  protected readonly describe = describe;
  protected readonly priceFor = priceFor;
}
