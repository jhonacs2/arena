/**
 * Las carreras. El ciclo de vida está en `decisiones.md` §3:
 *
 *   draft ──▶ open ──▶ running ──▶ finished
 *     │         │
 *     └─────────┴──▶ cancelled
 */
export type RaceStatus = 'draft' | 'open' | 'running' | 'finished' | 'cancelled';

export interface Horse {
  readonly id: string;
  readonly number: number;
  readonly name: string;
  /** Cuota ×100 en entero: `340` es 3,40. Nunca se hace aritmética con float. */
  readonly odds: number;
}

export interface Participant {
  readonly userId: string;
  readonly username: string;
}

/** La apuesta propia, tal como la resuelve el servidor. */
export interface Bet {
  readonly id: string;
  readonly horseId: string;
  readonly horseName: string;
  readonly amount: number;
  /** La cuota **congelada** en el momento de apostar. No se recalcula nunca. */
  readonly oddsAtBet: number;
  readonly potentialPayout: number;
}

/**
 * Una apuesta de otro alumno, vista desde la sala.
 *
 * `horseId` llega **en `null` mientras la carrera está `open`**: revelarlo haría
 * que los últimos en apostar copien a los primeros (`api.md`, WebSocket).
 */
export interface PublicBet {
  readonly userId: string;
  readonly username: string;
  readonly amount: number;
  readonly horseId: string | null;
}

export interface RaceSummary {
  readonly id: string;
  readonly name: string;
  readonly status: RaceStatus;
  readonly scheduledAt: string;
  readonly horseCount: number;
  readonly participantCount: number;
  /** Viene resuelto: sin esto el frontend haría una llamada por carrera. */
  readonly myBet: Bet | null;
}

export interface ResultEntry {
  readonly position: number;
  readonly horseId: string;
  readonly horseName: string;
}

export interface RaceDetail {
  readonly id: string;
  readonly name: string;
  readonly status: RaceStatus;
  readonly scheduledAt: string;
  readonly horses: readonly Horse[];
  readonly participants: readonly Participant[];
  readonly bets: readonly PublicBet[];
  readonly myBet: Bet | null;
  readonly results: readonly ResultEntry[] | null;
  /** El pago propio, si la carrera terminó. Se arma por destinatario. */
  readonly myPayout: number | null;
}

export interface PlaceBetRequest {
  readonly horseId: string;
  readonly amount: number;
}

export interface PlaceBetResponse {
  readonly bet: Bet;
  readonly balance: number;
  readonly points: number;
}
