/**
 * Genera docs/contract/fixtures/race-ticks.jsonl — una grabación completa de
 * race_005 (Gran Premio Nacional, 8 caballos).
 *
 * Determinístico: sin Math.random ni Date.now. Correrlo dos veces produce
 * exactamente el mismo archivo, así el fixture puede vivir en git sin ruido.
 *
 *   node scripts/gen-race-ticks.mjs
 *
 * Lo consumen:
 *   - MockSocketService del frontend, que lo reproduce en tiempo real (S10).
 *   - El test golden del backend, que compara la forma de sus eventos.
 */

import { writeFileSync, mkdirSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const ROOT = join(dirname(fileURLToPath(import.meta.url)), '..');
const OUT = join(ROOT, 'docs/contract/fixtures/race-ticks.jsonl');

const RACE_ID = 'race_005';
const TICK_HZ = 10;
const COUNTDOWN_SECONDS = 10; // la grabación arranca a 10 s de la largada, no a 60
const STARTED_AT = '2026-03-14T12:00:00Z';
const FINISHED_AT = '2026-03-14T12:00:44Z';

/**
 * finishTime define el orden de llegada. style define *cómo* corre:
 *   front  — arranca fuerte y se apaga
 *   even   — ritmo parejo
 *   closer — arranca atrás y remata
 * El favorito gana rematando, así el marcador cambia de manos y la clase
 * tiene algo que mirar.
 */
const FIELD = [
  { id: 'hrs_029', name: 'Payador',  finishTime: 44.8, style: 'closer', phase: 0.31 },
  { id: 'hrs_031', name: 'Espuela',  finishTime: 45.1, style: 'front',  phase: 1.74 },
  { id: 'hrs_035', name: 'Bagual',   finishTime: 45.4, style: 'even',   phase: 2.96 },
  { id: 'hrs_028', name: 'Yatay',    finishTime: 45.8, style: 'front',  phase: 4.12 },
  { id: 'hrs_030', name: 'Tacuara',  finishTime: 46.2, style: 'even',   phase: 5.38 },
  { id: 'hrs_033', name: 'Chañar',   finishTime: 46.7, style: 'closer', phase: 0.87 },
  { id: 'hrs_032', name: 'Lucero',   finishTime: 47.3, style: 'front',  phase: 2.05 },
  { id: 'hrs_034', name: 'Tordillo', finishTime: 48.1, style: 'even',   phase: 3.63 },
];

const SHAPE = { front: 0.82, even: 1.0, closer: 1.22 };

const RACE_DURATION = FIELD[0].finishTime; // gana el primero de la lista

const round = (n, d) => Number(n.toFixed(d));

/** Ondulación suave y determinística. Simula el vaivén sin romper la monotonía. */
function wobble(phase, t) {
  return 0.0045 * Math.sin(phase + t * 0.9) + 0.0022 * Math.sin(phase * 2.7 + t * 2.3);
}

const lines = [];
const emit = (offsetMs, event) => lines.push(JSON.stringify({ _offsetMs: offsetMs, ...event }));

// ── Cuenta regresiva, 1 Hz ────────────────────────────────────────────────
for (let s = COUNTDOWN_SECONDS; s >= 1; s--) {
  emit((COUNTDOWN_SECONDS - s) * 1000, {
    type: 'race.countdown',
    raceId: RACE_ID,
    secondsLeft: s,
  });
}

const T0 = COUNTDOWN_SECONDS * 1000;

// ── Largada ───────────────────────────────────────────────────────────────
emit(T0, { type: 'race.started', raceId: RACE_ID, startedAt: STARTED_AT });

// ── Ticks a 10 Hz ─────────────────────────────────────────────────────────
const previous = new Map(FIELD.map((h) => [h.id, 0]));
const totalTicks = Math.round(RACE_DURATION * TICK_HZ);

for (let i = 1; i <= totalTicks; i++) {
  const t = round(i / TICK_HZ, 1);

  const raw = FIELD.map((horse) => {
    const u = Math.min(t / horse.finishTime, 1);
    const base = Math.pow(u, SHAPE[horse.style]);
    // La ondulación se apaga cerca de la llegada: nadie zigzaguea en el disco.
    const noisy = base + wobble(horse.phase, t) * (1 - u);
    const progress = Math.min(Math.max(noisy, previous.get(horse.id)), 1);
    previous.set(horse.id, progress);
    return { horseId: horse.id, progress };
  });

  const order = [...raw].sort(
    (a, b) => b.progress - a.progress || a.horseId.localeCompare(b.horseId),
  );
  const placeOf = new Map(order.map((p, idx) => [p.horseId, idx + 1]));

  emit(T0 + i * (1000 / TICK_HZ), {
    type: 'race.tick',
    raceId: RACE_ID,
    t,
    positions: raw.map((p) => ({
      horseId: p.horseId,
      progress: round(p.progress, 3),
      place: placeOf.get(p.horseId),
    })),
  });
}

const T_END = T0 + totalTicks * (1000 / TICK_HZ);

// ── Llegada ───────────────────────────────────────────────────────────────
// payouts trae solo las apuestas del usuario conectado. La grabación es la de
// usr_001 (Ana Robles), que tenía bet_005 sobre Payador: 400 × 2.75 = 1100.
emit(T_END + 200, {
  type: 'race.finished',
  raceId: RACE_ID,
  finishedAt: FINISHED_AT,
  podium: ['hrs_029', 'hrs_031', 'hrs_035'],
  payouts: [{ betId: 'bet_005', amount: 1100 }],
});

emit(T_END + 400, { type: 'balance.updated', balance: 6100 });

emit(T_END + 600, {
  type: 'leaderboard.updated',
  entries: [
    { rank: 1,  userId: 'usr_005', displayName: 'Elena Quiroga',   profit: 2100, bets: 3, wins: 2 },
    { rank: 2,  userId: 'usr_010', displayName: 'Joaquín Ferrer',  profit: 1790, bets: 4, wins: 3 },
    { rank: 3,  userId: 'usr_009', displayName: 'Irene Castro',    profit: 1700, bets: 2, wins: 1 },
    { rank: 4,  userId: 'usr_004', displayName: 'Diego Paredes',   profit: 1110, bets: 3, wins: 2 },
    { rank: 5,  userId: 'usr_001', displayName: 'Ana Robles',      profit: 955,  bets: 5, wins: 3 },
    { rank: 6,  userId: 'usr_002', displayName: 'Bruno Salas',     profit: 500,  bets: 3, wins: 1 },
    { rank: 7,  userId: 'usr_007', displayName: 'Gabriela Nieto',  profit: 300,  bets: 2, wins: 1 },
    { rank: 8,  userId: 'usr_012', displayName: 'Lautaro Mendive', profit: 125,  bets: 3, wins: 1 },
    { rank: 9,  userId: 'usr_006', displayName: 'Facundo Ibarra',  profit: -300, bets: 3, wins: 1 },
    { rank: 10, userId: 'usr_011', displayName: 'Karina Villalba', profit: -350, bets: 2, wins: 0 },
    { rank: 11, userId: 'usr_008', displayName: 'Hugo Lemos',      profit: -750, bets: 3, wins: 0 },
  ],
});

mkdirSync(dirname(OUT), { recursive: true });
writeFileSync(OUT, lines.join('\n') + '\n', 'utf8');

console.log(`race-ticks.jsonl · ${lines.length} eventos · ${totalTicks} ticks · ${RACE_DURATION}s`);
