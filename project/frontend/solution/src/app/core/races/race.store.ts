import { Injectable, computed, signal } from '@angular/core';

import { RACES } from '../mocks';
import { favourite, type Horse, type Race, type RaceStatus } from '../models';

/** El filtro incluye un valor que no es un estado. */
export type RaceFilter = RaceStatus | 'all';

/** Lo que una vista necesita de una carrera, ya preparado. */
export interface RaceView {
  readonly race: Race;
  readonly favourite: Horse | undefined;
}

/**
 * S5 · El estado de las carreras, fuera de la pantalla.
 *
 * En S3 esto vivía adentro de `race-list`, y funcionaba. El techo apareció
 * cuando hizo falta la misma información en otro lado: la portada, la pantalla
 * de la carrera en vivo, el widget del saldo. Todos tendrían que volver a
 * filtrar, volver a buscar y volver a decidir cuál está abierta.
 *
 * **Un store es un servicio que guarda estado.** No es un patrón nuevo ni una
 * librería: es lo de S3 con `providedIn: 'root'` encima.
 *
 * La forma de exponerlo es la de siempre en este proyecto:
 *
 *   private readonly _filter = signal(…);          se escribe adentro
 *   readonly filter = this._filter.asReadonly();   se lee afuera
 *
 * Nadie de afuera puede llamar a `set`. Para cambiar algo hay que pasar por un
 * método, y los métodos son el contrato del store.
 */
@Injectable({ providedIn: 'root' })
export class RaceStore {
  /**
   * El programa entero. **No es un signal**: viene de una constante y no cambia.
   * En S7, cuando lo traiga `HttpClient`, va a serlo — y va a cambiar esta
   * línea y ninguna otra.
   */
  private readonly all: readonly RaceView[] = RACES.map((race) => ({
    race,
    favourite: favourite(race),
  }));

  private readonly _filter = signal<RaceFilter>('all');
  private readonly _query = signal('');
  private readonly _selectedId = signal<string | null>(null);

  readonly filter = this._filter.asReadonly();
  readonly query = this._query.asReadonly();

  readonly counts = computed(() => ({
    all: this.all.length,
    live: this.all.filter((view) => view.race.status === 'live').length,
    upcoming: this.all.filter((view) => view.race.status === 'upcoming').length,
    finished: this.all.filter((view) => view.race.status === 'finished').length,
  }));

  readonly visible = computed<readonly RaceView[]>(() => {
    const status = this._filter();
    const text = this._query().trim().toLowerCase();

    return this.all
      .filter((view) => status === 'all' || view.race.status === status)
      .filter((view) => text === '' || this.matches(view, text));
  });

  /**
   * La carrera abierta, derivada del id.
   *
   * Si el filtro la deja afuera, el panel se cierra solo. Es la decisión de S3,
   * que se mudó tal cual: mover el estado a un servicio no cambia cómo se
   * piensa, cambia quién es el dueño.
   */
  readonly selected = computed<RaceView | undefined>(() => {
    const id = this._selectedId();
    return id === null ? undefined : this.visible().find((view) => view.race.id === id);
  });

  /** Los caballos de la carrera abierta, de menor a mayor cuota. */
  readonly lineup = computed<readonly Horse[]>(() => {
    const horses = this.selected()?.race.horses ?? [];
    return [...horses].sort((a, b) => a.odds - b.odds || a.number - b.number);
  });

  setFilter(filter: RaceFilter): void {
    this._filter.set(filter);
  }

  setQuery(query: string): void {
    this._query.set(query);
  }

  toggle(raceId: string): void {
    this._selectedId.update((current) => (current === raceId ? null : raceId));
  }

  clearFilters(): void {
    this._filter.set('all');
    this._query.set('');
  }

  private matches(view: RaceView, text: string): boolean {
    if (view.race.name.toLowerCase().includes(text)) return true;
    return view.race.horses.some((horse) => horse.name.toLowerCase().includes(text));
  }
}
