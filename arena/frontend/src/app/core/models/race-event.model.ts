import type { Bet, PublicBet, RaceDetail, ResultEntry } from './race.model';

/**
 * Los eventos del socket de `api.md` §WebSocket. Van **servidor → cliente**: el
 * cliente no manda nada salvo el handshake, porque apostar es un `POST`.
 *
 * **El sobre lleva `type` y `raceId`, y la carga va adentro.** Este archivo decía
 * antes que el sobre era «una asunción» porque `api.md` listaba las cargas sin el
 * envoltorio, y la asunción salió mal: el backend manda `room.state` con la sala
 * anidada en `race`, y acá se leían los campos como si estuvieran planos.
 *
 * Costó caro y conviene entender por qué. `event.participants` daba `undefined`,
 * se escribía tal cual en el signal, y `participants().length` reventaba el render
 * de toda la pantalla. El síntoma no era «faltan los participantes»: era que **no
 * se podía seleccionar un caballo**, porque con el template roto el clic
 * actualizaba el estado y la vista nunca se volvía a pintar. Nada de eso lo ve
 * TypeScript: un campo que no viene es `undefined` y sigue de largo.
 *
 * Ahora el sobre está escrito en `api.md`, que es donde se mira antes de asumir.
 */

/**
 * Lo primero que llega al conectarse: la sala completa.
 *
 * `race` es **el mismo `RaceDetail` que devuelve `GET /races/:id`**, y eso es a
 * propósito — el cliente tiene un solo modelo para la foto y para el vivo.
 */
export interface RoomStateEvent {
  readonly type: 'room.state';
  readonly raceId: string;
  readonly race: RaceDetail;
}

export interface RoomJoinedEvent {
  readonly type: 'room.joined';
  readonly raceId: string;
  readonly userId: string;
  readonly username: string;
  readonly participantCount: number;
}

/**
 * Alguien apostó.
 *
 * **No trae el caballo, y no es un olvido:** si lo trajera, los últimos en
 * apostar copiarían a los primeros y la apuesta dejaría de medir criterio. En el
 * backend el campo directamente no existe en la estructura, así que no hay forma
 * de filtrarlo por descuido.
 */
export interface BetPlacedEvent {
  readonly type: 'bet.placed';
  readonly raceId: string;
  readonly userId: string;
  readonly username: string;
  readonly amount: number;
  readonly betCount: number;
}

export interface RaceStartedEvent {
  readonly type: 'race.started';
  readonly raceId: string;
  readonly startedAt: string;
  /** Al largar, el servidor revela las apuestas que hasta ahora venían tapadas. */
  readonly bets: readonly PublicBet[];
}

export interface HorsePosition {
  readonly horseId: string;
  /** Avance normalizado, 0 en la largada y 1 en el disco. */
  readonly progress: number;
  /** Puesto en ese instante, 1 el que va adelante. */
  readonly place: number;
}

export interface RaceTickEvent {
  readonly type: 'race.tick';
  readonly raceId: string;
  /** Segundos de carrera. A 10 Hz llega uno cada 100 ms. */
  readonly t: number;
  readonly positions: readonly HorsePosition[];
}

/**
 * Terminó. **El sobre se arma por destinatario**, no se difunde igual a todos:
 * difundir el mismo objeto filtraría cuánto cobró cada uno.
 *
 * No trae `points`: la nota suma los puntos regalados, que el paquete de carreras
 * no conoce. Salen de `GET /api/me`.
 */
export interface RaceFinishedEvent {
  readonly type: 'race.finished';
  readonly raceId: string;
  readonly results: readonly ResultEntry[];
  /** Las del resto, ya reveladas, sin el pago de nadie. */
  readonly bets: readonly PublicBet[];
  /** La propia, liquidada. `null` si no apostó. El pago está en `myBet.payout`. */
  readonly myBet: Bet | null;
  /** `null` cuando el saldo no cambió, para no tocar el widget al pedo. */
  readonly balance: number | null;
}

export interface RaceCancelledEvent {
  readonly type: 'race.cancelled';
  readonly raceId: string;
  readonly reason: string;
  /** Lo devuelto a quien recibe el evento. Íntegro: exactamente lo apostado. */
  readonly myRefund: number | null;
  readonly balance: number | null;
}

export type RaceEvent =
  | RoomStateEvent
  | RoomJoinedEvent
  | BetPlacedEvent
  | RaceStartedEvent
  | RaceTickEvent
  | RaceFinishedEvent
  | RaceCancelledEvent;
