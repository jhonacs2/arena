#!/usr/bin/env node
/**
 * Genera los mocks tipados del frontend desde docs/contract/seed/.
 *
 *   node scripts/gen-mocks.mjs           genera
 *   node scripts/gen-mocks.mjs --check   falla si están desfasados
 *
 * S1–S4 trabajan con datos hardcodeados, como pide el temario. Pero
 * "hardcodeado" no quiere decir "inventado": estas constantes salen del mismo
 * seed que carga el backend Go, así que la card que el alumno escribe en S1
 * funciona sin cambios cuando en S7 se conecta al servidor real.
 */

import { readFileSync, writeFileSync, existsSync, mkdirSync } from 'node:fs';
import { dirname, join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

const ROOT = join(dirname(fileURLToPath(import.meta.url)), '..');
const CHECK = process.argv.includes('--check');

const TARGETS = [
  'project/frontend/solution/src/app/core/mocks',
  'project/frontend/starter/src/app/core/mocks',
];

const readSeed = (file) =>
  JSON.parse(readFileSync(join(ROOT, 'docs/contract/seed', file), 'utf8'));

const races = readSeed('races.json');
const users = readSeed('users.json');
const bets = readSeed('bets.json');
const results = readSeed('results.json');
const leaderboard = readSeed('leaderboard.json');

/** Serializa a TypeScript legible: comillas simples y sin comillas en claves. */
function ts(value, indent = 0) {
  const pad = '  '.repeat(indent);
  const padIn = '  '.repeat(indent + 1);

  if (Array.isArray(value)) {
    if (value.length === 0) return '[]';
    return `[\n${value.map((v) => padIn + ts(v, indent + 1)).join(',\n')},\n${pad}]`;
  }
  if (value !== null && typeof value === 'object') {
    const entries = Object.entries(value).filter(([k]) => !k.startsWith('_'));
    if (entries.length === 0) return '{}';
    const body = entries.map(([k, v]) => `${padIn}${k}: ${ts(v, indent + 1)}`).join(',\n');
    return `{\n${body},\n${pad}}`;
  }
  if (typeof value === 'string') return `'${value.replace(/'/g, "\\'")}'`;
  return String(value);
}

const header = (what) => `/**
 * ${what}
 *
 * GENERADO desde docs/contract/seed/ por scripts/gen-mocks.mjs — no editar a
 * mano. Es el MISMO dataset que carga el backend Go, así que un componente
 * escrito contra estos datos funciona sin cambios cuando en S7 se conecta al
 * servidor real.
 */
`;

// La misma regla de rebase que aplica el backend (docs/contract/README.md).
// Sin esto, en una clase de agosto la próxima carrera aparecería en marzo.
const rebaseHelper = `
/**
 * Las fechas del dataset están ancladas a un instante fijo y se desplazan al
 * cargar el módulo, igual que hace el backend: \`docs/contract/README.md\`,
 * regla de rebase. Así la carrera en vivo siempre está en vivo y la próxima
 * siempre está por venir, se dé la clase el mes que se dé.
 */
const ANCHOR = Date.parse('2026-03-14T12:00:00Z');
const OFFSET = Date.now() - ANCHOR;

function rebase(iso: string): string {
  return new Date(Date.parse(iso) + OFFSET).toISOString().replace(/\\.\\d{3}Z$/, 'Z');
}
`;

const files = {
  'races.mock.ts':
    header('Las 8 carreras del dataset, con sus 54 caballos.') +
    `\nimport type { Race } from '../models';\n` +
    rebaseHelper +
    `\nexport const RACES: readonly Race[] = ${ts(races)}.map((race) => ({\n` +
    `  ...race,\n  startsAt: rebase(race.startsAt),\n})) as readonly Race[];\n`,

  'bets.mock.ts':
    header('Las 34 apuestas del dataset — 30 liquidadas y 4 pendientes.') +
    `\nimport type { Bet } from '../models';\n` +
    rebaseHelper +
    `\nexport const BETS: readonly Bet[] = ${ts(bets)}.map((bet) => ({\n` +
    `  ...bet,\n  placedAt: rebase(bet.placedAt),\n})) as readonly Bet[];\n`,

  'users.mock.ts':
    header('Los 12 usuarios del dataset. Contraseña de todos: Carrera123!') +
    `\nimport type { User } from '../models';\n` +
    `\nexport const USERS: readonly User[] = ${ts(users)};\n` +
    `\n/** El usuario por defecto de las demos, mientras no haya sesión real. */\n` +
    `export const CURRENT_USER: User = USERS[0]!;\n`,

  'results.mock.ts':
    header('Podios de las carreras terminadas. `payouts` se completa por usuario.') +
    `\nimport type { RaceResult } from '../models';\n` +
    rebaseHelper +
    `\nexport const RESULTS: readonly RaceResult[] = ${ts(results)}.map((result) => ({\n` +
    `  ...result,\n  finishedAt: rebase(result.finishedAt),\n  payouts: [],\n})) as readonly RaceResult[];\n`,

  'leaderboard.mock.ts':
    header('Tabla de posiciones. Calculada desde las apuestas liquidadas.') +
    `\nimport type { LeaderboardEntry } from '../models';\n` +
    `\nexport const LEADERBOARD_ALL: readonly LeaderboardEntry[] = ${ts(leaderboard.all)};\n` +
    `\nexport const LEADERBOARD_DAILY: readonly LeaderboardEntry[] = ${ts(leaderboard.daily)};\n`,

  'index.ts':
    header('Punto de entrada de los mocks.') +
    `\nexport * from './bets.mock';\nexport * from './leaderboard.mock';\n` +
    `export * from './races.mock';\nexport * from './results.mock';\nexport * from './users.mock';\n`,
};

const stale = [];
let written = 0;

for (const target of TARGETS) {
  const dir = join(ROOT, target);
  // Solo se escribe donde el proyecto ya existe.
  if (!existsSync(join(ROOT, target.split('/src/')[0]))) continue;

  for (const [name, content] of Object.entries(files)) {
    const path = join(dir, name);
    const current = existsSync(path) ? readFileSync(path, 'utf8') : null;
    if (current === content) continue;

    if (CHECK) {
      stale.push(`${target}/${name}`);
    } else {
      mkdirSync(dir, { recursive: true });
      writeFileSync(path, content, 'utf8');
      written++;
      console.log(`  ${relative(ROOT, path)}`);
    }
  }
}

if (stale.length) {
  console.error('\n  ✗ Mocks desfasados respecto del seed:');
  for (const s of stale) console.error(`    · ${s}`);
  console.error('\n    Correr: node scripts/gen-mocks.mjs\n');
  process.exit(1);
}

if (!CHECK) console.log(`  ${written} archivo(s) generados.`);
