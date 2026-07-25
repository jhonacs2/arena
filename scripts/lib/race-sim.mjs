/**
 * Simulador de carrera — implementación de referencia en JavaScript.
 *
 * El backend Go implementa EXACTAMENTE este algoritmo (internal/sim/sim.go).
 * La especificación normativa está en docs/contract/race-simulation.md.
 *
 * Que los dos lados produzcan la misma carrera no es un capricho: es lo que
 * hace que `MockSocketService` y el backend real muestren lo mismo, y por eso
 * el punto 5 de la definición de terminado ("se ve igual en los dos") se puede
 * verificar en lugar de prometer.
 *
 * Determinístico: mismo (raceId, runIndex, caballos) → misma carrera, siempre.
 * Sin Math.random, sin Date.now.
 */

export const TICK_HZ = 10;
export const COUNTDOWN_SECONDS = 60;

export const BASE_DURATION = 42.0; // segundos que tarda el favorito nominal
export const ODDS_SPREAD = 4.5;   // cuánto más tarda el de cuota más alta
export const JITTER = 5.0;         // ventana aleatoria; es lo que permite el batacazo

/** front, even, closer. El exponente define la curva de esfuerzo. */
export const SHAPES = [0.82, 1.0, 1.22];
export const STYLE_NAMES = ['front', 'even', 'closer'];

const WOBBLE_A1 = 0.0045;
const WOBBLE_F1 = 0.9;
const WOBBLE_A2 = 0.0022;
const WOBBLE_F2 = 2.3;
const WOBBLE_PHASE_MULT = 2.7;

// ── Aleatoriedad determinística ───────────────────────────────────────────

/** FNV-1a de 32 bits. Trivial de reimplementar en cualquier lenguaje. */
export function fnv1a(text) {
  let h = 0x811c9dc5;
  for (let i = 0; i < text.length; i++) {
    h ^= text.charCodeAt(i);
    h = Math.imul(h, 0x01000193) >>> 0;
  }
  return h;
}

/** Mezcla final, para que ids parecidos no den valores parecidos. */
export function mix32(h) {
  h = (h ^ (h >>> 15)) >>> 0;
  h = Math.imul(h, 0x2545f491) >>> 0;
  h = (h ^ (h >>> 13)) >>> 0;
  return h;
}

/** Devuelve un número en [0, 1) derivado de la clave completa. */
export function rnd(raceId, runIndex, horseId, salt) {
  return mix32(fnv1a(`${raceId}/${runIndex}/${horseId}/${salt}`)) / 4294967296;
}

// ── Preparación de la carrera ─────────────────────────────────────────────

/**
 * @param {string} raceId
 * @param {number} runIndex  corrida de esa carrera; 0 es la primera
 * @param {{id:string,name:string,number:number,odds:number}[]} horses
 */
export function prepare(raceId, runIndex, horses) {
  // La cuota más baja es el favorito. El empate se rompe por número de partida,
  // para que el orden no dependa del orden del array.
  const byOdds = [...horses].sort((a, b) => a.odds - b.odds || a.number - b.number);
  const skillOf = new Map(byOdds.map((h, i) => [h.id, horses.length > 1 ? i / (horses.length - 1) : 0]));

  const runners = horses.map((horse) => {
    const skill = skillOf.get(horse.id);
    const finishTime =
      BASE_DURATION + ODDS_SPREAD * skill + JITTER * (rnd(raceId, runIndex, horse.id, 't') - 0.5);
    const styleIndex = Math.floor(rnd(raceId, runIndex, horse.id, 's') * SHAPES.length);
    const phase = rnd(raceId, runIndex, horse.id, 'p') * 2 * Math.PI;

    return {
      horseId: horse.id,
      name: horse.name,
      number: horse.number,
      odds: horse.odds,
      finishTime,
      shape: SHAPES[styleIndex],
      style: STYLE_NAMES[styleIndex],
      phase,
    };
  });

  // La carrera termina cuando cruza el primero, no cuando llegan todos.
  //
  // Redondeo hacia ARRIBA a la décima, no al más cercano: con el más cercano,
  // un tiempo de 42.14 daba duración 42.1 y el último tick caía antes de la
  // llegada — el ganador quedaba en 0.998 y la carrera terminaba sin que nadie
  // cruzara. Pasaba en 1 de cada 4 corridas.
  const duration = Math.ceil(Math.min(...runners.map((r) => r.finishTime)) * 10) / 10;

  return { raceId, runIndex, runners, duration };
}

// ── Progreso ──────────────────────────────────────────────────────────────

const round1 = (n) => Math.round(n * 10) / 10;
const round3 = (n) => Math.round(n * 1000) / 1000;

function wobble(phase, t) {
  return (
    WOBBLE_A1 * Math.sin(phase + t * WOBBLE_F1) +
    WOBBLE_A2 * Math.sin(phase * WOBBLE_PHASE_MULT + t * WOBBLE_F2)
  );
}

/**
 * Genera todos los ticks de la carrera. `progress` nunca retrocede y el orden
 * del array es el de `horses`: ordenar por `place` es tarea del cliente.
 */
export function ticks(prepared) {
  const previous = new Map(prepared.runners.map((r) => [r.horseId, 0]));
  const total = Math.round(prepared.duration * TICK_HZ);
  const out = [];

  for (let i = 1; i <= total; i++) {
    const t = round1(i / TICK_HZ);

    const raw = prepared.runners.map((r) => {
      const u = Math.min(t / r.finishTime, 1);
      // La ondulación se apaga cerca del disco: nadie zigzaguea en la llegada.
      const noisy = Math.pow(u, r.shape) + wobble(r.phase, t) * (1 - u);
      const progress = Math.min(Math.max(noisy, previous.get(r.horseId)), 1);
      previous.set(r.horseId, progress);
      return { horseId: r.horseId, progress };
    });

    const ordered = [...raw].sort(
      (a, b) => b.progress - a.progress || a.horseId.localeCompare(b.horseId),
    );
    const placeOf = new Map(ordered.map((p, idx) => [p.horseId, idx + 1]));

    out.push({
      t,
      positions: raw.map((p) => ({
        horseId: p.horseId,
        progress: round3(p.progress),
        place: placeOf.get(p.horseId),
      })),
    });
  }

  return out;
}

/** Podio: los tres primeros del último tick. */
export function podium(prepared, allTicks) {
  const last = allTicks[allTicks.length - 1];
  return [...last.positions]
    .sort((a, b) => a.place - b.place)
    .slice(0, 3)
    .map((p, i) => {
      const runner = prepared.runners.find((r) => r.horseId === p.horseId);
      return {
        place: i + 1,
        horseId: runner.horseId,
        horseName: runner.name,
        number: runner.number,
        odds: runner.odds,
      };
    });
}

/** Corre todo de una: preparar, generar ticks y resolver el podio. */
export function simulate(raceId, runIndex, horses) {
  const prepared = prepare(raceId, runIndex, horses);
  const allTicks = ticks(prepared);
  return { ...prepared, ticks: allTicks, podium: podium(prepared, allTicks) };
}
