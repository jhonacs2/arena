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
  /**
   * Cuota **nominal** ×100 en entero: `340` es 3,40.
   *
   * Es informativa y **no determina el pago**: la liquidación es pari-mutuel
   * (`decisiones.md` §1), así que lo que cobra el que acierta sale de repartir el
   * pozo, no de multiplicar por esta cuota. Sirve para dos cosas y ninguna más:
   * marcar cuál es el favorito, y alimentar la velocidad del simulador.
   */
  readonly nominalOdds: number;
}

export interface Participant {
  readonly userId: string;
  readonly username: string;
}

/**
 * La apuesta propia, tal como la resuelve el servidor.
 *
 * **No hay pago potencial.** Con pari-mutuel, lo que se cobra depende del pozo
 * completo y de cuántos acertaron, y las dos cosas siguen cambiando hasta que se
 * cierran las apuestas. Cualquier número que mostráramos al apostar sería una
 * promesa que el servidor no puede cumplir.
 *
 * `payout` llega en `null` hasta que la carrera se liquida. Ahí es el pago real.
 */
export interface Bet {
  readonly id: string;
  readonly horseId: string;
  readonly amount: number;
  readonly status: 'placed' | 'won' | 'lost' | 'refunded';
  readonly payout: number | null;
  readonly createdAt: string;
  readonly settledAt: string | null;
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
  readonly scheduledAt: string | null;
  readonly horses: readonly Horse[];
  readonly participants: readonly Participant[];
  readonly bets: readonly PublicBet[];
  /** El pago propio, ya liquidado, sale de `myBet.payout`. No es un campo aparte. */
  readonly myBet: Bet | null;
  readonly results: readonly ResultEntry[] | null;
}

export interface PlaceBetRequest {
  readonly horseId: string;
  readonly amount: number;
}

export interface PlaceBetResponse {
  readonly bet: Bet;
  readonly balance: number;
}
