import { ChangeDetectionStrategy, Component, inject } from '@angular/core';

import { SessionStore } from '../../core/auth/session.store';
import { AdminCodes } from './admin-codes';
import { AdminRaces } from './admin-races';
import { AdminScores } from './admin-scores';

/**
 * El panel del instructor.
 *
 * Tres secciones en una sola página, sin pestañas: en clase se usan las tres en
 * el mismo minuto —abrir la carrera, ver quién apostó, regalar monedas al que
 * participó— y esconder dos tercios detrás de una pestaña obliga a ir y venir
 * mientras veinte personas esperan.
 */
@Component({
  selector: 'app-admin',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [AdminRaces, AdminScores, AdminCodes],
  templateUrl: './admin.html',
  styleUrl: './admin.css',
})
export class Admin {
  protected readonly session = inject(SessionStore);
}
