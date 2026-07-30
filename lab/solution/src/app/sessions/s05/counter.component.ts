import { ChangeDetectionStrategy, Component, inject, input, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { SHOP_NAME } from './shop.token';
import { NotepadService } from './notepad.service';
import { OrderService } from './order.service';

/**
 * S5 · Un mostrador.
 *
 * En la pantalla hay **dos copias** de este componente, una al lado de la otra.
 * Sirven para ver la diferencia sin explicarla:
 *
 *   - La comanda (`OrderService`) es la MISMA en los dos. Un pedido tomado en
 *     el mostrador de la izquierda aparece a la derecha.
 *   - El cuaderno (`NotepadService`) es DISTINTO en cada uno, porque está
 *     declarado en `providers` de este componente.
 *
 * Ninguno de los dos mostradores sabe que el otro existe. No hay `input()` ni
 * `output()` entre ellos: no son padre e hijo, son hermanos.
 */
@Component({
  selector: 'app-counter',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [FormsModule],
  // Esta línea es la clase entera: una instancia de NotepadService POR CADA
  // <app-counter> que exista en pantalla.
  providers: [NotepadService],
  templateUrl: './counter.component.html',
  styleUrl: './counter.component.css',
})
export class CounterComponent {
  readonly label = input.required<string>();

  /**
   * `inject()` pide una dependencia. Se llama en el campo de la clase, que es
   * un **contexto de inyección**: el momento en que Angular está construyendo
   * este componente y sabe a qué inyector preguntarle.
   */
  protected readonly orders = inject(OrderService);
  protected readonly notepad = inject(NotepadService);

  /** Un valor de configuración, no un servicio. Ver `shop.token.ts`. */
  protected readonly shopName = inject(SHOP_NAME);

  protected readonly customer = signal('');
  protected readonly note = signal('');

  protected readonly coffees = ['Yirgacheffe', 'Huila', 'Cerrado'] as const;
  protected readonly coffee = signal<string>('Yirgacheffe');

  protected take(): void {
    this.orders.add(this.customer(), this.coffee());
    this.customer.set('');
  }

  protected jot(): void {
    this.notepad.write(this.note());
    this.note.set('');
  }
}
