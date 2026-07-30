import { ChangeDetectionStrategy, Component, computed, input } from '@angular/core';

import type { RaceStatus } from '../../../core/models';

/** El texto que ve el usuario, en castellano. La clase CSS, en inglés. */
const LABEL: Readonly<Record<RaceStatus, string>> = {
  draft: 'Borrador',
  open: 'Abierta',
  running: 'Corriendo',
  finished: 'Terminada',
  cancelled: 'Cancelada',
};

/**
 * El estado de una carrera.
 *
 * `running` es el único que late: es el estado que cambia solo y el que el
 * alumno tiene que poder encontrar de un barrido por la pantalla. Los otros
 * cuatro son quietos porque no pasa nada en ellos.
 */
@Component({
  selector: 'app-status-badge',
  changeDetection: ChangeDetectionStrategy.OnPush,
  styleUrl: './status-badge.css',
  template: `<span [class]="'badge badge--' + status()">{{ label() }}</span>`,
})
export class StatusBadge {
  readonly status = input.required<RaceStatus>();
  protected readonly label = computed(() => LABEL[this.status()]);
}
