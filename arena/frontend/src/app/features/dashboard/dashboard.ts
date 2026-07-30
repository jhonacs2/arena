import { httpResource } from '@angular/common/http';
import { ChangeDetectionStrategy, Component, computed, effect, inject } from '@angular/core';
import { RouterLink } from '@angular/router';

import { api } from '../../core/api/api';
import { toApiError } from '../../core/api/api-error';
import { SessionStore } from '../../core/auth/session.store';
import { LEDGER_REASON_LABEL, type LedgerEntry, type Me, type RaceSummary } from '../../core/models';
import { CoinsPipe, SignedCoinsPipe } from '../../shared/format/coins.pipe';
import { OddsPipe } from '../../shared/format/odds.pipe';
import { WhenPipe } from '../../shared/format/when.pipe';
import { StatusBadge } from '../../shared/ui/badge/status-badge';
import { Callout } from '../../shared/ui/callout/callout';
import { EmptyState } from '../../shared/ui/empty-state/empty-state';

/**
 * El tablero del alumno: su saldo, su historial y las carreras.
 *
 * Las tres lecturas son `httpResource`. Cada una trae su propio `isLoading()` y
 * su `error()`, así que las tres tarjetas manejan los tres estados —cargando,
 * vacío, error— sin tres signals cableados a mano por pantalla.
 */
@Component({
  selector: 'app-dashboard',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [
    RouterLink,
    StatusBadge,
    Callout,
    EmptyState,
    CoinsPipe,
    SignedCoinsPipe,
    OddsPipe,
    WhenPipe,
  ],
  templateUrl: './dashboard.html',
  styleUrls: ['../../shared/ui/surface.css', './dashboard.css'],
})
export class Dashboard {
  protected readonly session = inject(SessionStore);

  protected readonly me = httpResource<Me>(() => api('/me'));

  protected readonly races = httpResource<{ items: RaceSummary[] }>(() => api('/races'), {
    defaultValue: { items: [] },
  });

  protected readonly ledger = httpResource<{ items: LedgerEntry[] }>(
    () => api('/me/transactions'),
    { defaultValue: { items: [] } },
  );

  protected readonly reasonLabel = LEDGER_REASON_LABEL;

  /** Las abiertas primero: son las únicas donde todavía se puede hacer algo. */
  protected readonly ordered = computed(() => {
    const weight = { open: 0, running: 1, finished: 2, cancelled: 3, draft: 4 };
    return [...this.races.value().items].sort((a, b) => weight[a.status] - weight[b.status]);
  });

  protected readonly racesError = computed(() =>
    this.races.status() === 'error' ? toApiError(this.races.error()).message : null,
  );

  protected readonly ledgerError = computed(() =>
    this.ledger.status() === 'error' ? toApiError(this.ledger.error()).message : null,
  );

  constructor() {
    // El saldo autoritativo es el del servidor. La cabecera muestra el que tiene
    // guardado; cuando `/me` responde, se sincroniza.
    effect(() => {
      const me = this.me.value();
      if (me !== undefined) this.session.sync(me);
    });
  }
}
