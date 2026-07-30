import { httpResource } from '@angular/common/http';
import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';
import { FormField, FormRoot, form, max, min, required, submit } from '@angular/forms/signals';

import { api } from '../../core/api/api';
import { toApiError } from '../../core/api/api-error';
import { AdminService } from '../../core/data/admin.service';
import type { InviteCode } from '../../core/models';
import { CoinsPipe } from '../../shared/format/coins.pipe';
import { WhenPipe } from '../../shared/format/when.pipe';
import { Button } from '../../shared/ui/button/button';
import { Callout } from '../../shared/ui/callout/callout';
import { FieldErrors } from '../../shared/ui/field/field-errors';

interface CodeBatch {
  count: number;
  coinsGranted: number;
  note: string;
}

/**
 * Generar códigos y verlos.
 *
 * Los códigos recién creados se muestran en un bloque grande y monoespaciado
 * porque el instructor los va a **dictar en voz alta o copiar a un chat**. Es el
 * mismo motivo por el que el alfabeto no tiene `I`, `L`, `O`, `U`, `0` ni `1`.
 */
@Component({
  selector: 'app-admin-codes',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [FormField, FormRoot, Button, Callout, FieldErrors, CoinsPipe, WhenPipe],
  templateUrl: './admin-codes.html',
  styleUrls: [
    '../../shared/ui/surface.css',
    '../../shared/ui/field/form-controls.css',
    './admin-codes.css',
  ],
})
export class AdminCodes {
  private readonly admin = inject(AdminService);

  protected readonly codes = httpResource<{ items: InviteCode[] }>(() => api('/admin/codes'), {
    defaultValue: { items: [] },
  });

  private readonly model = signal<CodeBatch>({ count: 10, coinsGranted: 1000, note: '' });

  protected readonly batch = form(this.model, (path) => {
    required(path.count);
    min(path.count, 1, { message: 'Al menos uno.' });
    max(path.count, 200, { message: 'De a 200 como máximo.' });
    min(path.coinsGranted, 1, { message: 'Tiene que otorgar monedas.' });
  });

  private readonly _created = signal<readonly string[]>([]);
  protected readonly created = this._created.asReadonly();

  private readonly _failure = signal<string | null>(null);
  protected readonly failure = this._failure.asReadonly();

  protected readonly pending = computed(
    () => this.codes.value().items.filter((code) => code.redeemedAt === null).length,
  );

  protected readonly listError = computed(() =>
    this.codes.status() === 'error' ? toApiError(this.codes.error()).message : null,
  );

  protected async generate(): Promise<void> {
    this._failure.set(null);

    await submit(this.batch, {
      action: async () => {
        const { count, coinsGranted, note } = this.model();
        try {
          const response = await this.admin.createCodes({ count, coinsGranted, note: note.trim() });
          this._created.set(response.codes);
          this.codes.reload();
        } catch (cause) {
          this._failure.set(toApiError(cause).message);
        }
        return undefined;
      },
    });
  }
}
