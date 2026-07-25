/**
 * Genera docs/contract/fixtures/race-ticks.jsonl — la grabación completa de
 * race_005, corrida 164, desde la sesión de usr_001 (Ana Robles).
 *
 *   node scripts/gen-race-ticks.mjs
 *
 * Todo sale del simulador de scripts/lib/race-sim.mjs, que es la misma
 * especificación que implementa el backend Go (docs/contract/race-simulation.md).
 * Nada acá está escrito a mano: ni el podio, ni los pagos, ni el leaderboard.
 *
 * Determinístico. Correrlo dos veces produce el mismo archivo.
 */

import { readFileSync, writeFileSync, mkdirSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { simulate } from './lib/race-sim.mjs';

const ROOT = join(dirname(fileURLToPath(import.meta.url)), '..');
const OUT = join(ROOT, 'docs/contract/fixtures/race-ticks.jsonl');

const RACE_ID = 'race_005';
const RUN_INDEX = 164;
// El servidor cuenta 60 s antes de largar (docs/contract/ws-events.md). La
// grabación arranca a 10: nadie mira un minuto de cuenta regresiva en una demo,
// y los eventos de countdown son idénticos entre sí salvo el número.
const FIXTURE_COUNTDOWN_SECONDS = 10;
const VIEWER = 'usr_001'; // desde qué sesión se grabó: payouts y balance son de este usuario

const readSeed = (f) => JSON.parse(readFileSync(join(ROOT, 'docs/contract/seed', f), 'utf8'));
const races = readSeed('races.json');
const users = readSeed('users.json');
const bets = readSeed('bets.json');
const results = readSeed('results.json');

const race = races.find((r) => r.id === RACE_ID);
const sim = simulate(RACE_ID, RUN_INDEX, race.horses);
const winner = sim.podium[0].horseId;

// Las fechas del fixture están ancladas al mismo instante que el seed.
const startsAt = race.startsAt;

const lines = [];
const emit = (offsetMs, event) => lines.push(JSON.stringify({ _offsetMs: offsetMs, ...event }));

// ── Cuenta regresiva, 1 Hz ────────────────────────────────────────────────
for (let s = FIXTURE_COUNTDOWN_SECONDS; s >= 1; s--) {
  emit((FIXTURE_COUNTDOWN_SECONDS - s) * 1000, { type: 'race.countdown', raceId: RACE_ID, secondsLeft: s });
}

const T0 = FIXTURE_COUNTDOWN_SECONDS * 1000;

// ── Largada y ticks ───────────────────────────────────────────────────────
emit(T0, { type: 'race.started', raceId: RACE_ID, startedAt: startsAt });

for (const tick of sim.ticks) {
  emit(T0 + Math.round(tick.t * 1000), {
    type: 'race.tick',
    raceId: RACE_ID,
    t: tick.t,
    positions: tick.positions,
  });
}

const T_END = T0 + Math.round(sim.duration * 1000);

// ── Liquidación ───────────────────────────────────────────────────────────
// Igual que el backend: gana quien apostó al del puesto 1, con la cuota
// congelada al apostar. Todo se calcula, nada se escribe a mano.
const settled = bets
  .filter((b) => b.raceId === RACE_ID && b.status === 'pending')
  .map((b) => ({
    ...b,
    status: b.horseId === winner ? 'won' : 'lost',
    payout: b.horseId === winner ? Math.round(b.amount * b.odds) : 0,
  }));

// `payouts` de race.finished trae solo las apuestas del usuario conectado.
emit(T_END + 200, {
  type: 'race.finished',
  raceId: RACE_ID,
  podium: sim.podium.map((p) => p.horseId),
  payouts: settled
    .filter((b) => b.userId === VIEWER)
    .map((b) => ({ betId: b.id, amount: b.payout })),
});

const viewer = users.find((u) => u.id === VIEWER);
const credited = settled
  .filter((b) => b.userId === VIEWER)
  .reduce((sum, b) => sum + b.payout, 0);
emit(T_END + 400, { type: 'balance.updated', balance: viewer.balance + credited });

// ── Leaderboard recalculado con la carrera ya liquidada ───────────────────
const allBets = bets.map((b) => settled.find((s) => s.id === b.id) ?? b);
const finishedRaces = new Set([...results.map((r) => r.raceId), RACE_ID]);
const userById = new Map(users.map((u) => [u.id, u]));

const acc = new Map();
for (const bet of allBets) {
  if (bet.status === 'pending' || !finishedRaces.has(bet.raceId)) continue;
  const e = acc.get(bet.userId) ?? { profit: 0, bets: 0, wins: 0 };
  e.profit += bet.payout - bet.amount;
  e.bets += 1;
  if (bet.status === 'won') e.wins += 1;
  acc.set(bet.userId, e);
}

const entries = [...acc.entries()]
  .map(([userId, e]) => ({ userId, displayName: userById.get(userId).displayName, ...e }))
  .sort((a, b) => b.profit - a.profit || b.wins - a.wins || a.displayName.localeCompare(b.displayName, 'es'))
  .slice(0, 20)
  .map((e, i) => ({ rank: i + 1, ...e }));

emit(T_END + 600, { type: 'leaderboard.updated', entries });

mkdirSync(dirname(OUT), { recursive: true });
writeFileSync(OUT, lines.join('\n') + '\n', 'utf8');

// ── Informe ───────────────────────────────────────────────────────────────
const mid = sim.ticks[Math.floor(sim.ticks.length / 2)];
const midPlace = mid.positions.find((p) => p.horseId === winner).place;
const last = [...sim.ticks.at(-1).positions].sort((a, b) => a.place - b.place);
const style = sim.runners.find((r) => r.horseId === winner).style;

console.log(`race-ticks.jsonl · ${lines.length} eventos · ${sim.ticks.length} ticks · ${sim.duration}s`);
console.log(`  ${RACE_ID} corrida ${RUN_INDEX}, grabada desde ${VIEWER}`);
console.log(`  gana ${sim.podium[0].horseName} @${sim.podium[0].odds} (${style}), iba ${midPlace}º a mitad`);
console.log(`  llega por ${(last[0].progress - last[1].progress).toFixed(3)} sobre ${sim.podium[1].horseName}`);
console.log(`  paga a ${VIEWER}: ${credited}`);
