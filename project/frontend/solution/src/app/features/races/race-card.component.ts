import { ChangeDetectionStrategy, Component, input, output } from '@angular/core';

import type { Horse, Race } from '../../core/models';

/**
 * S2 · Una carrera del programa.
 *
 * Salió de adentro del `@for` de `race-list`, y el criterio para decidir qué
 * se lleva y qué se queda fue uno solo:
 *
 *   **Lo que la tarjeta necesita para dibujarse, entra por `input()`.
 *   Lo que decide la pantalla, sale por `output()`.**
 *
 * Por eso `selected` es un input y no un estado propio: cuál está abierta es
 * una decisión del listado —solo puede haber una— y la tarjeta no tiene forma
 * de saber qué pasa en las otras siete. Ella avisa que la tocaron y se calla.
 *
 * Y por eso no importa `RACES` ni `favourite()`: si los importara, esta
 * tarjeta solo serviría para el listado del programa. Así sirve para
 * cualquier lugar donde haya una carrera que mostrar.
 */
@Component({
  selector: 'app-race-card',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  templateUrl: './race-card.component.html',
  styleUrl: './race-card.component.css',
})
export class RaceCardComponent {
  /** Sin carrera no hay tarjeta: `required` lo vuelve un error de compilación. */
  readonly race = input.required<Race>();

  /** Ya formateada por el padre. La tarjeta no decide cómo se escribe una fecha. */
  readonly time = input.required<string>();

  /** Puede no haber: una carrera sin caballos no tiene favorito. */
  readonly favourite = input<Horse | undefined>(undefined);

  /** Lo decide el listado, no la tarjeta. */
  readonly selected = input(false);

  /** «Me tocaron.» Qué hacer con eso es problema del padre. */
  readonly toggled = output<Race>();
}
