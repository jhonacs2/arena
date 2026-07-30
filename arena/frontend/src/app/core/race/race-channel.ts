import { Injectable, inject } from '@angular/core';
import { Observable, concat, defer, of } from 'rxjs';
import { filter, map } from 'rxjs/operators';
import { webSocket } from 'rxjs/webSocket';

import { environment } from '../../../environments/environment';
import { MockWorld } from '../api/mock/mock-world';
import { SessionStore } from '../auth/session.store';
import type { RaceEvent } from '../models';

/**
 * El canal de la carrera.
 *
 * Es una clase abstracta y no una interfaz porque hace de token de inyección: la
 * pantalla de la carrera pide `RaceChannel` y no sabe —ni le importa— si atrás
 * hay un WebSocket de verdad o el mock.
 *
 * Los eventos van **servidor → cliente**. El cliente no manda nada salvo el
 * handshake: apostar es un `POST`, no un mensaje de socket.
 */
@Injectable()
export abstract class RaceChannel {
  abstract connect(raceId: string): Observable<RaceEvent>;
}

@Injectable()
export class RaceSocket extends RaceChannel {
  private readonly session = inject(SessionStore);

  override connect(raceId: string): Observable<RaceEvent> {
    return defer(() => {
      const token = this.session.accessToken() ?? '';
      const scheme = location.protocol === 'https:' ? 'wss' : 'ws';
      const params = new URLSearchParams({ raceId, token });
      const url = `${scheme}://${location.host}${environment.apiBaseUrl}/ws?${params}`;

      return webSocket<RaceEvent>({ url });
    });
  }
}

@Injectable()
export class MockRaceChannel extends RaceChannel {
  private readonly world = inject(MockWorld);
  private readonly session = inject(SessionStore);

  override connect(raceId: string): Observable<RaceEvent> {
    return defer(() => {
      const userId = this.world.userIdFromToken(this.session.accessToken());
      if (userId === null) return of<RaceEvent>();

      // `room.state` primero, como en el socket real: al conectarse se recibe la
      // sala completa y después las novedades.
      const state = of<RaceEvent>(this.world.roomState(userId, raceId));
      const live = this.world.messages.pipe(
        // `race.finished` va dirigido: cada uno recibe su propio pago.
        filter((message) => message.raceId === raceId),
        filter((message) => message.to === null || message.to === userId),
        map((message) => message.event),
      );

      return concat(state, live);
    });
  }
}

export const provideRaceChannel = () => ({
  provide: RaceChannel,
  useClass: environment.useMockBackend ? MockRaceChannel : RaceSocket,
});
