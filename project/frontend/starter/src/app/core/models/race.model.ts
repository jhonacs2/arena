/**
 * Carreras y caballos — contraparte de `docs/contract/openapi.yaml`.
 *
 * Los nombres de campo son el contrato: hay structs Go con exactamente estos
 * nombres, y un test golden de cada lado que compara contra los mismos
 * archivos de `docs/contract/samples/`. Si acá cambia un nombre, se rompe el
 * otro lado en silencio.
 *
 * ⚠️ **Este archivo es la Misión 2 de S0.** Los nombres están bien; los tipos
 * son más flojos que el contrato. Compila igual —ese es el problema— y hay
 * cosas que el contrato garantiza y que estos tipos no dicen. Cada una está
 * marcada con `TODO(S0)`.
 *
 * La fuente de verdad es `docs/contract/openapi.yaml`. No se adivina: se abre.
 */

/**
 * En qué momento está una carrera.
 *
 * TODO(S0) · 1 — El contrato dice que hay exactamente tres estados:
 * `upcoming`, `live` y `finished`. Con `string`, esto compila:
 *
 *   if (race.status === 'galopando') { … }
 *
 * y no se cumple nunca. Escribí la unión de literales y volvé a probar esa
 * comparación: el error tiene que aparecer antes de abrir el navegador.
 */
export type RaceStatus = string;

/**
 * TODO(S0) · 2 — Un caballo del dataset no se edita en el lugar. Hoy esto
 * compila y no debería:
 *
 *   horse.odds = 1.01;
 */
export interface Horse {
  id: string;
  name: string;
  /** Posición de partida. Única dentro de la carrera. */
  number: number;
  /** Cuota decimal. El pago es `amount × odds`. */
  odds: number;
}

/**
 * TODO(S0) · 3 — Mismo problema que `Horse`, y uno más: la lista de caballos
 * admite `push`. Además de `readonly` en cada campo, el array va
 * `readonly Horse[]`.
 */
export interface Race {
  id: string;
  name: string;
  /** ISO 8601 en UTC. */
  startsAt: string;
  status: RaceStatus;
  /**
   * El listado también los trae, no solo el detalle. Está decidido en
   * `docs/contract/README.md`: si fueran opcionales, el alumno pelearía con
   * `undefined` desde S1 sin haber visto todavía qué es un tipo opcional.
   */
  horses: Horse[];
}

/**
 * TODO(S0) · 4 — El podio tiene tres lugares y el contrato lo dice: `place`
 * vale 1, 2 o 3. `number` acepta `0`, `-7` y `99`.
 */
export interface PodiumEntry {
  place: number;
  horseId: string;
  horseName: string;
  number: number;
  odds: number;
}

export interface Payout {
  betId: string;
  horseId: string;
  /** Lo que apostó. */
  stake: number;
  /** Lo que cobró. `0` si perdió. */
  amount: number;
}

export interface RaceResult {
  raceId: string;
  finishedAt: string;
  podium: PodiumEntry[];
  /** Solo las apuestas del usuario autenticado. Vacío sin sesión. */
  payouts: Payout[];
}

/**
 * El favorito es el de menor cuota; el empate se rompe por número de partida.
 *
 * TODO(S0) · 5 — Ese `!` es una promesa que este código no puede cumplir. El
 * contrato exige al menos dos caballos por carrera, pero el **tipo** no lo
 * garantiza: `favourite({ …, horses: [] })` explota en tiempo de ejecución con
 * `Cannot read properties of undefined`.
 *
 * Decí la verdad en el tipo de retorno y sacá el `!`. El compilador te va a
 * llevar de la mano: con `noUncheckedIndexedAccess` prendido, `horses[0]` ya
 * es `Horse | undefined` y no hay forma de ignorarlo sin mentir de nuevo.
 */
export function favourite(race: Race): Horse {
  let best = race.horses[0]!;
  for (const horse of race.horses) {
    if (horse.odds < best.odds) best = horse;
    else if (horse.odds === best.odds && horse.number < best.number) best = horse;
  }
  return best;
}
