import { httpResource } from '@angular/common/http';
import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';
import {
  FormField,
  FormRoot,
  applyEach,
  form,
  min,
  minLength,
  required,
  submit,
} from '@angular/forms/signals';
import { RouterLink } from '@angular/router';

import { api } from '../../core/api/api';
import { toApiError } from '../../core/api/api-error';
import { AdminService } from '../../core/data/admin.service';
import type { RaceSummary, RaceStatus } from '../../core/models';
import { OddsPipe } from '../../shared/format/odds.pipe';
import { WhenPipe } from '../../shared/format/when.pipe';
import { StatusBadge } from '../../shared/ui/badge/status-badge';
import { Button } from '../../shared/ui/button/button';
import { Callout } from '../../shared/ui/callout/callout';
import { FieldErrors } from '../../shared/ui/field/field-errors';

interface HorseRow {
  number: number;
  name: string;
  nominalOdds: number;
}

interface RaceDraft {
  name: string;
  scheduledAt: string;
  horses: HorseRow[];
}

/** `2026-07-29T22:30`, el formato que espera un `datetime-local`. */
function localInputValue(date: Date): string {
  const pad = (value: number): string => String(value).padStart(2, '0');
  return (
    `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}` +
    `T${pad(date.getHours())}:${pad(date.getMinutes())}`
  );
}

const startingHorses = (): HorseRow[] => [
  { number: 1, name: '', nominalOdds: 300 },
  { number: 2, name: '', nominalOdds: 450 },
  { number: 3, name: '', nominalOdds: 600 },
  { number: 4, name: '', nominalOdds: 900 },
];

/**
 * Armar carreras y operarlas.
 *
 * **El instructor es el operador: nada arranca solo.** Las cuatro acciones son las
 * transiciones de `decisiones.md` §3, y cada botón aparece solo en el estado desde
 * el que la transición existe. La validación de verdad la hace el servidor: si dos
 * pestañas mandan «largar» a la vez, una recibe `INVALID_TRANSITION`.
 */
@Component({
  selector: 'app-admin-races',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    RouterLink,
    FormField,
    FormRoot,
    StatusBadge,
    Button,
    Callout,
    FieldErrors,
    OddsPipe,
    WhenPipe,
  ],
  templateUrl: './admin-races.html',
  styleUrls: [
    '../../shared/ui/surface.css',
    '../../shared/ui/field/form-controls.css',
    './admin-races.css',
  ],
})
export class AdminRaces {
  private readonly admin = inject(AdminService);

  protected readonly races = httpResource<{ items: RaceSummary[] }>(() => api('/races'), {
    defaultValue: { items: [] },
  });

  private readonly model = signal<RaceDraft>({
    name: '',
    scheduledAt: localInputValue(new Date(Date.now() + 30 * 60 * 1000)),
    horses: startingHorses(),
  });

  protected readonly draft = form(this.model, (path) => {
    required(path.name, { message: 'La carrera necesita un nombre.' });
    required(path.scheduledAt, { message: 'Poné una hora.' });
    minLength(path.horses, 2, { message: 'Una carrera necesita al menos dos caballos.' });

    // `applyEach` aplica el mismo esquema a cada ítem del arreglo, incluso a los
    // que todavía no existen: agregar una fila no requiere volver a validar nada.
    applyEach(path.horses, (horse) => {
      required(horse.name, { message: 'Falta el nombre.' });
      // Cuotas ×100 en entero. 101 es 1,01: por debajo de eso la casa perdería
      // plata en cada apuesta, lo que no es una carrera sino un regalo.
      min(horse.nominalOdds, 101, { message: 'Mínimo 101, que es 1,01.' });
    });
  });

  private readonly _failure = signal<string | null>(null);
  protected readonly failure = this._failure.asReadonly();

  private readonly _busy = signal<string | null>(null);
  protected readonly busy = this._busy.asReadonly();

  protected readonly listError = computed(() =>
    this.races.status() === 'error' ? toApiError(this.races.error()).message : null,
  );

  protected readonly ordered = computed(() => {
    const weight: Readonly<Record<RaceStatus, number>> = {
      running: 0,
      open: 1,
      draft: 2,
      finished: 3,
      cancelled: 4,
    };
    return [...this.races.value().items].sort((a, b) => weight[a.status] - weight[b.status]);
  });

  protected addHorse(): void {
    this.model.update((draft) => ({
      ...draft,
      horses: [
        ...draft.horses,
        { number: draft.horses.length + 1, name: '', nominalOdds: 500 },
      ],
    }));
  }

  protected removeHorse(index: number): void {
    this.model.update((draft) => ({
      ...draft,
      // Se renumeran las filas: los dorsales de una carrera son 1..n sin huecos.
      horses: draft.horses
        .filter((_horse, position) => position !== index)
        .map((horse, position) => ({ ...horse, number: position + 1 })),
    }));
  }

  protected async create(): Promise<void> {
    this._failure.set(null);

    await submit(this.draft, {
      action: async () => {
        const { name, scheduledAt, horses } = this.model();
        try {
          await this.admin.createRace({
            name: name.trim(),
            scheduledAt: new Date(scheduledAt).toISOString(),
            horses: horses.map((horse) => ({
              number: horse.number,
              name: horse.name.trim(),
              nominalOdds: horse.nominalOdds,
            })),
          });
          this.model.set({
            name: '',
            scheduledAt: localInputValue(new Date(Date.now() + 30 * 60 * 1000)),
            horses: startingHorses(),
          });
          this.races.reload();
        } catch (cause) {
          this._failure.set(toApiError(cause).message);
        }
        return undefined;
      },
    });
  }

  protected async operate(raceId: string, action: 'open' | 'start' | 'cancel'): Promise<void> {
    this._busy.set(raceId);
    this._failure.set(null);
    try {
      if (action === 'open') await this.admin.openRace(raceId);
      else if (action === 'start') await this.admin.startRace(raceId);
      else await this.admin.cancelRace(raceId, 'Cancelada por el instructor');
      this.races.reload();
    } catch (cause) {
      this._failure.set(toApiError(cause).message);
    } finally {
      this._busy.set(null);
    }
  }
}
