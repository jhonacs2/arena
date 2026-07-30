import { ChangeDetectionStrategy, Component, computed, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { BetStore } from '../../core/bets/bet.store';
import { type Race, type RaceStatus } from '../../core/models';
import { RaceStore, type RaceFilter, type RaceView } from '../../core/races/race.store';
import { FavouriteDirective } from '../../shared/directives/favourite.directive';
import { MoneyPipe } from '../../shared/pipes/money.pipe';
import { OddsPipe } from '../../shared/pipes/odds.pipe';
import { BadgeComponent, type BadgeTone } from '../../shared/ui/badge/badge.component';
import { SilkComponent } from '../../shared/ui/silk/silk.component';
import { RaceCardComponent } from './race-card.component';

/**
 * S1 + S2 + S3 + S4 + S5 · Listado de carreras.
 *
 * La historia del archivo es la historia del curso:
 *
 *   S1 · todo el marcado aquí adentro
 *   S2 · la tarjeta se fue a <app-race-card>
 *   S3 · el estado se fue a signals
 *   S4 · el formateo se fue a pipes y directivas
 *   S5 · el estado se fue a un servicio
 *
 * Lo que queda es lo único que de verdad es de esta pantalla: **cómo se ve**.
 * Las etiquetas de los estados, los tonos de las pastillas y el formato de la
 * hora son decisiones de presentación; el filtro, la búsqueda y cuál carrera
 * está abierta ya no le pertenecen, porque otras pantallas los van a necesitar.
 */

const STATUS_LABELS: Record<RaceStatus, string> = {
  upcoming: 'Por largar',
  live: 'En vivo',
  finished: 'Terminada',
};

const STATUS_TONES: Record<RaceStatus, BadgeTone> = {
  upcoming: 'neutral',
  live: 'live',
  finished: 'neutral',
};

const FILTER_LABELS: Record<RaceFilter, string> = {
  all: 'Todas',
  live: 'En vivo',
  upcoming: 'Por largar',
  finished: 'Terminadas',
};

const TIME_FORMAT = new Intl.DateTimeFormat('es', {
  day: '2-digit',
  month: 'short',
  hour: '2-digit',
  minute: '2-digit',
});

@Component({
  selector: 'app-race-list',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    FormsModule,
    SilkComponent,
    RaceCardComponent,
    BadgeComponent,
    FavouriteDirective,
    MoneyPipe,
    OddsPipe,
  ],
  templateUrl: './race-list.component.html',
  styleUrl: './race-list.component.css',
})
export class RaceListComponent {
  /** Los dos stores. La pantalla ya no guarda nada. */
  protected readonly races = inject(RaceStore);
  protected readonly bets = inject(BetStore);

  protected readonly filters: readonly RaceFilter[] = ['all', 'live', 'upcoming', 'finished'];
  protected readonly filterLabels = FILTER_LABELS;

  /**
   * Lo de presentación se deriva de lo que trae el store.
   *
   * Esto **sí** es de la pantalla: otra vista de las mismas carreras podría
   * querer otras etiquetas, otros tonos o la hora en otro formato.
   */
  protected readonly visible = computed(() =>
    this.races.visible().map((view) => ({
      view,
      time: TIME_FORMAT.format(new Date(view.race.startsAt)),
      statusLabel: STATUS_LABELS[view.race.status],
      tone: STATUS_TONES[view.race.status],
    })),
  );

  protected readonly selectedTime = computed(() => {
    const race = this.races.selected()?.race;
    return race === undefined ? '' : TIME_FORMAT.format(new Date(race.startsAt));
  });

  /** `[(ngModel)]` necesita algo con `set`; el store no lo expone. */
  protected get query(): string {
    return this.races.query();
  }

  protected set query(value: string) {
    this.races.setQuery(value);
  }

  protected get amount(): number {
    return this.bets.amount();
  }

  protected set amount(value: number) {
    this.bets.setAmount(value);
  }

  protected isSelected(view: RaceView): boolean {
    return this.races.selected()?.race.id === view.race.id;
  }

  protected selectFilter(filter: RaceFilter): void {
    this.races.setFilter(filter);
  }

  protected select(race: Race): void {
    this.races.toggle(race.id);
  }
}
