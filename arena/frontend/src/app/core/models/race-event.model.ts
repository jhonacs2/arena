import type { Participant, PublicBet, RaceStatus, ResultEntry } from './race.model';

/**
 * Los eventos del socket de `api.md` §WebSocket. Van **servidor → cliente**: el
 * cliente no manda nada salvo el handshake, porque apostar es un `POST`.
 *
 * ⚠️ **Asunción del sobre.** `api.md` lista el nombre del evento y su carga pero
 * no el envoltorio JSON. Acá se asume un objeto plano discriminado por `type`:
 *
 * ```json
 * { "type": "race.tick", "t": 42, "positions": [{ "horseId": "…", "progress": 0.61 }] }
 * ```
 *
 * Si el backend elige otra forma (por ejemplo `{ event, data }`), el único
 * archivo que cambia es este más el `map` de `race-socket.ts`.
 */
export interface RoomStateEvent {
  readonly type: 'room.state';
  readonly status: RaceStatus;
  readonly participants: readonly Participant[];
  readonly bets: readonly PublicBet[];
}

export interface RoomJoinedEvent {
  readonly type: 'room.joined';
  readonly userId: string;
  readonly username: string;
  readonly participantCount: number;
}

export interface BetPlacedEvent {
  readonly type: 'bet.placed';
  readonly userId: string;
  readonly username: string;
  readonly amount: number;
  /** `null` mientras la carrera está `open`. Al largar se revelan todas juntas. */
  readonly horseId: string | null;
}

export interface RaceStartedEvent {
  readonly type: 'race.started';
  readonly startedAt: string;
  /** Al largar, el servidor revela las apuestas que hasta ahora venían tapadas. */
  readonly bets: readonly PublicBet[];
}

export interface HorsePosition {
  readonly horseId: string;
  /** Avance normalizado, 0 en la largada y 1 en el disco. */
  readonly progress: number;
}

export interface RaceTickEvent {
  readonly type: 'race.tick';
  /** Número de tick. A 10 Hz, `t / 10` son los segundos de carrera. */
  readonly t: number;
  readonly positions: readonly HorsePosition[];
}

export interface RaceFinishedEvent {
  readonly type: 'race.finished';
  readonly results: readonly ResultEntry[];
  /** El pago **propio**. El sobre se arma por destinatario, no se difunde igual. */
  readonly payout: number;
  readonly balance: number;
  readonly points: number;
}

export interface RaceCancelledEvent {
  readonly type: 'race.cancelled';
  readonly reason: string;
}

export type RaceEvent =
  | RoomStateEvent
  | RoomJoinedEvent
  | BetPlacedEvent
  | RaceStartedEvent
  | RaceTickEvent
  | RaceFinishedEvent
  | RaceCancelledEvent;
