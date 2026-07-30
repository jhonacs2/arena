import { httpResource } from '@angular/common/http';
import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';

import { api } from '../../core/api/api';
import { toApiError } from '../../core/api/api-error';
import { AdminService } from '../../core/data/admin.service';
import type { ScoreRow } from '../../core/models';
import { CoinsPipe } from '../../shared/format/coins.pipe';
import { Button } from '../../shared/ui/button/button';
import { Callout } from '../../shared/ui/callout/callout';
import { EmptyState } from '../../shared/ui/empty-state/empty-state';

/**
 * El panel de nota, y el regalo de monedas.
 *
 * Muestra monedas **y** puntos porque son la misma cosa vista de dos maneras y el
 * instructor necesita las dos: las monedas para entender qué pasó en las
 * carreras, los puntos para la planilla. El mapeo final a la calificación lo
 * decide el instructor — no lo decide la app (`decisiones.md` §1).
 *
 * El regalo acepta números negativos: es un ajuste, y queda en el ledger con
 * quién lo hizo. **Nunca se edita la historia de la nota de alguien**, se compensa.
 */
@Component({
  selector: 'app-admin-scores',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Button, Callout, EmptyState, CoinsPipe],
  templateUrl: './admin-scores.html',
  styleUrls: [
    '../../shared/ui/surface.css',
    '../../shared/ui/field/form-controls.css',
    './admin-scores.css',
  ],
})
export class AdminScores {
  private readonly admin = inject(AdminService);

  protected readonly scores = httpResource<{ items: ScoreRow[] }>(() => api('/admin/scores'), {
    defaultValue: { items: [] },
  });

  /** A quién se le está regalando. `null` cierra el formulario. */
  private readonly _target = signal<ScoreRow | null>(null);
  protected readonly target = this._target.asReadonly();

  protected readonly coins = signal(100);
  protected readonly note = signal('');
  protected readonly working = signal(false);

  private readonly _failure = signal<string | null>(null);
  protected readonly failure = this._failure.asReadonly();

  protected readonly listError = computed(() =>
    this.scores.status() === 'error' ? toApiError(this.scores.error()).message : null,
  );

  protected readonly totalCoins = computed(() =>
    this.scores.value().items.reduce((sum, row) => sum + row.balance, 0),
  );

  protected open(row: ScoreRow): void {
    this._target.set(row);
    this.coins.set(100);
    this.note.set('');
    this._failure.set(null);
  }

  protected close(): void {
    this._target.set(null);
  }

  protected onCoins(event: Event): void {
    const input = event.target;
    if (!(input instanceof HTMLInputElement)) return;
    // El monto es entero. `parseInt` sobre el valor crudo, y `0` si quedó vacío:
    // un `NaN` viajando hacia el servidor es un 500 esperando a pasar.
    const parsed = Number.parseInt(input.value, 10);
    this.coins.set(Number.isNaN(parsed) ? 0 : parsed);
  }

  protected onNote(event: Event): void {
    const input = event.target;
    if (input instanceof HTMLInputElement) this.note.set(input.value);
  }

  protected async give(): Promise<void> {
    const row = this._target();
    if (row === null) return;

    this.working.set(true);
    this._failure.set(null);
    try {
      await this.admin.gift(row.userId, { coins: this.coins(), note: this.note().trim() });
      this._target.set(null);
      this.scores.reload();
    } catch (cause) {
      this._failure.set(toApiError(cause).message);
    } finally {
      this.working.set(false);
    }
  }
}
