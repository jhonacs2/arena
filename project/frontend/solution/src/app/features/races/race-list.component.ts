import { ChangeDetectionStrategy, Component } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { RACES } from '../../core/mocks';
import { favourite, potentialPayout, type Horse, type Race } from '../../core/models';
import { BadgeComponent, type BadgeTone } from '../../shared/ui/badge/badge.component';
import { SilkComponent } from '../../shared/ui/silk/silk.component';
import { RaceCardComponent } from './race-card.component';

/**
 * S1 + S2 · Listado de carreras.
 *
 * Primera pantalla del producto. Datos hardcodeados: vienen de `core/mocks`,
 * que sale del mismo dataset que carga el backend — así, cuando en S7 se
 * conecte al servidor real, esta pantalla no cambia.
 *
 * En S1 el marcado de cada carrera estaba acá adentro. En S2 se fue a
 * `<app-race-card>`, y lo que quedó es lo que de verdad es del listado:
 * **preparar los datos y decidir cuál está abierta.** Ese es el corte, y es
 * el que hay que saber hacer: si la tarjeta supiera cuál está abierta, no
 * podría usarse en ningún otro lado.
 */

/** Lo que la view necesita, ya preparado. */
interface RaceView {
  readonly race: Race;
  readonly time: string;
  readonly favourite: Horse | undefined;
  readonly statusLabel: string;
  readonly tone: BadgeTone;
}

const STATUS_LABELS: Record<Race['status'], string> = {
  upcoming: 'Por largar',
  live: 'En vivo',
  finished: 'Terminada',
};

const STATUS_TONES: Record<Race['status'], BadgeTone> = {
  upcoming: 'neutral',
  live: 'live',
  finished: 'neutral',
};

@Component({
  selector: 'app-race-list',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [FormsModule, SilkComponent, RaceCardComponent, BadgeComponent],
  templateUrl: './race-list.component.html',
  styleUrl: './race-list.component.css',
})
export class RaceListComponent {
  /**
   * Los datos se preparan UNA vez, en la clase, no en el template.
   *
   * Podría hacerse con métodos llamados desde el HTML, pero entonces se
   * recalcularían en cada detección de cambios. Preparar acá es más rápido y
   * deja el template diciendo qué se ve, no cómo se calcula.
   */
  protected readonly races: readonly RaceView[] = RACES.map((race) => ({
    race,
    time: this.formatTime(race.startsAt),
    favourite: favourite(race),
    statusLabel: STATUS_LABELS[race.status],
    tone: STATUS_TONES[race.status],
  }));

  /** La carrera que el usuario tocó. `null` = ninguna. */
  protected selected: RaceView | null = null;

  /** Lo que se escribe en el simulador. Va y viene con `[(ngModel)]`. */
  protected amount = 100;

  protected get payout(): number {
    const odds = this.selected?.favourite?.odds ?? 0;
    return potentialPayout(this.amount, odds);
  }

  protected select(view: RaceView): void {
    // Tocar la misma carrera dos veces la deselecciona: es lo que espera
    // cualquiera y ahorra un botón de cerrar.
    this.selected = this.selected?.race.id === view.race.id ? null : view;
  }

  /**
   * Formatea la hora sin pipes: `| date` es de S4.
   *
   * `Intl.DateTimeFormat` es del navegador, no de Angular. Vale la pena que se
   * vea una vez a mano antes de que un pipe lo esconda.
   *
   * Va el día además de la hora: el programa cruza más de una jornada, y con
   * solo «sáb, 20:41» no se entiende por qué una carrera de las 04:56 aparece
   * antes que una de las 19:21.
   */
  private formatTime(iso: string): string {
    return new Intl.DateTimeFormat('es', {
      day: '2-digit',
      month: 'short',
      hour: '2-digit',
      minute: '2-digit',
    }).format(new Date(iso));
  }
}
