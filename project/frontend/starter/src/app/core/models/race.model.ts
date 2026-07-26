/**
 * Carreras y caballos — contraparte de `docs/contract/openapi.yaml`.
 *
 * Los nombres de campo son el contrato: hay structs Go con exactamente estos
 * nombres, y un test golden de cada lado que compara contra los mismos
 * archivos de `docs/contract/samples/`. Si acá cambia un nombre, se rompe el
 * otro lado en silencio.
 */

export type RaceStatus = 'upcoming' | 'live' | 'finished';

export interface Horse {
  readonly id: string;
  readonly name: string;
  /** Posición de partida. Única dentro de la carrera. */
  readonly number: number;
  /** Cuota decimal. El pago es `amount × odds`. */
  readonly odds: number;
}

export interface Race {
  readonly id: string;
  readonly name: string;
  /** ISO 8601 en UTC. */
  readonly startsAt: string;
  readonly status: RaceStatus;
  /**
   * El listado también los trae, no solo el detalle. Está decidido en
   * `docs/contract/README.md`: si fueran opcionales, el alumno pelearía con
   * `undefined` desde S1 sin haber visto todavía qué es un tipo opcional.
   */
  readonly horses: readonly Horse[];
}

export interface PodiumEntry {
  readonly place: 1 | 2 | 3;
  readonly horseId: string;
  readonly horseName: string;
  readonly number: number;
  readonly odds: number;
}

export interface Payout {
  readonly betId: string;
  readonly horseId: string;
  /** Lo que apostó. */
  readonly stake: number;
  /** Lo que cobró. `0` si perdió. */
  readonly amount: number;
}

export interface RaceResult {
  readonly raceId: string;
  readonly finishedAt: string;
  readonly podium: readonly PodiumEntry[];
  /** Solo las apuestas del usuario autenticado. Vacío sin sesión. */
  readonly payouts: readonly Payout[];
}

/**
 * El favorito es el de menor cuota; el empate se rompe por número de partida.
 *
 * Devuelve `undefined` con una carrera sin caballos: el contrato exige al
 * menos dos, pero el tipo no lo puede garantizar y `noUncheckedIndexedAccess`
 * obliga a hacerse cargo.
 */
export function favourite(race: Race): Horse | undefined {
  return race.horses.reduce<Horse | undefined>((best, horse) => {
    if (!best) return horse;
    if (horse.odds < best.odds) return horse;
    if (horse.odds === best.odds && horse.number < best.number) return horse;
    return best;
  }, undefined);
}
