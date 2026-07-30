import { ChangeDetectionStrategy, Component, computed, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { RACES } from '../../core/mocks';
import { favourite, potentialPayout, type Horse, type Race, type RaceStatus } from '../../core/models';
import { BadgeComponent, type BadgeTone } from '../../shared/ui/badge/badge.component';
import { SilkComponent } from '../../shared/ui/silk/silk.component';
import { RaceCardComponent } from './race-card.component';

/**
 * S1 + S2 + S3 · Listado de carreras.
 *
 * En S1 el marcado de cada carrera estaba acá adentro. En S2 se fue a
 * `<app-race-card>`. En S3 el estado se fue a **signals**, y con eso el filtro
 * y la búsqueda dejaron de ser un problema: nada de lo que se ve está
 * guardado, se deriva.
 *
 * Las tres reglas de la sesión, escritas en este archivo:
 *
 *   1. Lo que **es** estado va en un `signal`. Acá son tres: el filtro, la
 *      búsqueda y cuál carrera está abierta. Nada más.
 *   2. Lo que se **deriva** va en un `computed`. No se guarda y no se
 *      sincroniza a mano; se recalcula solo cuando cambia su fuente.
 *   3. Nunca se modifica lo que había. Ni `push`, ni `sort` sobre el original.
 */

/** Lo que la view necesita, ya preparado. */
interface RaceView {
  readonly race: Race;
  readonly time: string;
  readonly favourite: Horse | undefined;
  readonly statusLabel: string;
  readonly tone: BadgeTone;
}

/** El filtro incluye un valor que no es un estado. */
type RaceFilter = RaceStatus | 'all';

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

@Component({
  selector: 'app-race-list',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [FormsModule, SilkComponent, RaceCardComponent, BadgeComponent],
  templateUrl: './race-list.component.html',
  styleUrl: './race-list.component.css',
})
export class RaceListComponent {
  protected readonly filters: readonly RaceFilter[] = ['all', 'live', 'upcoming', 'finished'];
  protected readonly filterLabels = FILTER_LABELS;

  /**
   * El programa entero, preparado una sola vez.
   *
   * **No es un signal, y es a propósito:** los datos vienen de una constante y
   * no cambian nunca. Un signal para algo que no cambia es ruido. En S7, cuando
   * los traiga `HttpClient`, sí va a serlo.
   */
  private readonly all: readonly RaceView[] = RACES.map((race) => ({
    race,
    time: this.formatTime(race.startsAt),
    favourite: favourite(race),
    statusLabel: STATUS_LABELS[race.status],
    tone: STATUS_TONES[race.status],
  }));

  // ── El estado: tres signals ─────────────────────────────────────────────

  protected readonly filter = signal<RaceFilter>('all');
  protected readonly query = signal('');

  /** Guarda el id, no la carrera: la carrera se deriva. */
  private readonly selectedId = signal<string | null>(null);

  protected readonly amount = signal(100);

  // ── Lo derivado ─────────────────────────────────────────────────────────

  /** Cuántas carreras hay de cada estado. Del programa entero, no de lo que se ve. */
  protected readonly counts = computed(() => ({
    all: this.all.length,
    live: this.all.filter((view) => view.race.status === 'live').length,
    upcoming: this.all.filter((view) => view.race.status === 'upcoming').length,
    finished: this.all.filter((view) => view.race.status === 'finished').length,
  }));

  protected readonly visible = computed<readonly RaceView[]>(() => {
    const status = this.filter();
    const text = this.query().trim().toLowerCase();

    return this.all
      .filter((view) => status === 'all' || view.race.status === status)
      .filter((view) => text === '' || this.matches(view, text));
  });

  /**
   * La carrera abierta.
   *
   * Se deriva del id, así que si el filtro la deja afuera, el panel se cierra
   * solo. Guardar el objeto entero obligaría a acordarse de limpiarlo a mano
   * en cada cambio de filtro — y a olvidarse una vez.
   */
  protected readonly selected = computed<RaceView | undefined>(() => {
    const id = this.selectedId();
    return id === null ? undefined : this.visible().find((view) => view.race.id === id);
  });

  protected readonly payout = computed(() =>
    potentialPayout(this.amount(), this.selected()?.favourite?.odds ?? 0),
  );

  /**
   * Los caballos de la carrera abierta, de menor a mayor cuota.
   *
   * El `[...]` no es decorativo: `sort()` ordena **en el lugar** y sin la copia
   * estaría reordenando el array que vino del dataset.
   */
  protected readonly lineup = computed<readonly Horse[]>(() => {
    const horses = this.selected()?.race.horses ?? [];
    return [...horses].sort((a, b) => a.odds - b.odds || a.number - b.number);
  });

  // ── Los cambios ─────────────────────────────────────────────────────────

  protected selectFilter(filter: RaceFilter): void {
    this.filter.set(filter);
  }

  protected select(race: Race): void {
    // Tocar la misma carrera dos veces la deselecciona: es lo que espera
    // cualquiera y ahorra un botón de cerrar.
    this.selectedId.update((current) => (current === race.id ? null : race.id));
  }

  protected clearFilters(): void {
    this.filter.set('all');
    this.query.set('');
  }

  private matches(view: RaceView, text: string): boolean {
    if (view.race.name.toLowerCase().includes(text)) return true;
    return view.race.horses.some((horse) => horse.name.toLowerCase().includes(text));
  }

  /**
   * Formatea la hora sin pipes: `| date` es de S4.
   *
   * `Intl.DateTimeFormat` es del navegador, no de Angular. Vale la pena que se
   * vea una vez a mano antes de que un pipe lo esconda.
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
