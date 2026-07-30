import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { api } from '../api/api';
import type {
  CreateCodesRequest,
  CreateRaceRequest,
  GiftRequest,
  GiftResponse,
  RaceDetail,
} from '../models';

/**
 * Lo que solo puede el instructor.
 *
 * Que estos métodos existan en el bundle no le da permiso a nadie: **cada uno de
 * estos endpoints valida el rol en el servidor** (`api.md` §Instructor). El guard
 * de la ruta y el menú escondido son comodidad, no seguridad.
 */
@Injectable({ providedIn: 'root' })
export class AdminService {
  private readonly http = inject(HttpClient);

  createCodes(body: CreateCodesRequest): Promise<{ codes: string[] }> {
    return firstValueFrom(this.http.post<{ codes: string[] }>(api('/admin/codes'), body));
  }

  gift(userId: string, body: GiftRequest): Promise<GiftResponse> {
    return firstValueFrom(
      this.http.post<GiftResponse>(api(`/admin/users/${userId}/gift`), body),
    );
  }

  createRace(body: CreateRaceRequest): Promise<{ race: RaceDetail }> {
    return firstValueFrom(this.http.post<{ race: RaceDetail }>(api('/admin/races'), body));
  }

  openRace(raceId: string): Promise<{ race: RaceDetail }> {
    return firstValueFrom(this.http.post<{ race: RaceDetail }>(api(`/admin/races/${raceId}/open`), {}));
  }

  /** `open → running`. Cierra las apuestas en el servidor y fija la semilla. */
  startRace(raceId: string): Promise<{ race: RaceDetail }> {
    return firstValueFrom(
      this.http.post<{ race: RaceDetail }>(api(`/admin/races/${raceId}/start`), {}),
    );
  }

  /** Devuelve cada apuesta íntegra al saldo, en una transacción. */
  cancelRace(raceId: string, reason: string): Promise<{ race: RaceDetail }> {
    return firstValueFrom(
      this.http.post<{ race: RaceDetail }>(api(`/admin/races/${raceId}/cancel`), { reason }),
    );
  }
}
