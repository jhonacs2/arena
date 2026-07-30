import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { api } from '../api/api';
import { SessionStore } from '../auth/session.store';
import type { PlaceBetRequest, PlaceBetResponse, RaceDetail } from '../models';

/**
 * Las escrituras del lado del alumno.
 *
 * Las **lecturas** no están acá: cada pantalla arma su propio `httpResource`, que
 * es reactivo al parámetro de la ruta y expone `isLoading()` y `error()` sin que
 * nadie tenga que cablear tres signals a mano.
 *
 * `resource()` es para leer. Apostar muta el saldo de alguien, y un `resource`
 * cancela su carga en vuelo cuando cambia la request: cancelar una apuesta a
 * medio camino es exactamente lo que no queremos.
 */
@Injectable({ providedIn: 'root' })
export class RaceService {
  private readonly http = inject(HttpClient);
  private readonly session = inject(SessionStore);

  /** Entrar a la sala. Idempotente: se puede llamar cada vez que se abre la pantalla. */
  join(raceId: string): Promise<RaceDetail> {
    return firstValueFrom(this.http.post<RaceDetail>(api(`/races/${raceId}/join`), {}));
  }

  async placeBet(raceId: string, body: PlaceBetRequest): Promise<PlaceBetResponse> {
    const response = await firstValueFrom(
      this.http.post<PlaceBetResponse>(api(`/races/${raceId}/bet`), body),
    );
    // El saldo autoritativo es el que devolvió el servidor, no `saldo - monto`.
    this.session.setBalance(response.balance);
    return response;
  }
}
