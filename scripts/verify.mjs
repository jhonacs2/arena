#!/usr/bin/env node
/**
 * Verificación mecánica del proyecto. Se corre después de cada feature.
 *
 *   node scripts/verify.mjs            todo
 *   node scripts/verify.mjs --fast     saltea los builds (útil mientras se itera)
 *   node scripts/verify.mjs contrato   solo un grupo: contrato | diseño | código
 *
 * `CLAUDE.md` §3 lo dice sin vueltas: una lista de reglas en prosa se degrada a
 * lo largo de una sesión larga. Esto es la versión que no se degrada.
 *
 * Saltea con gracia lo que todavía no existe: en Fase 0 no hay apps Angular y
 * el script igual verifica el contrato y el diseño.
 */

import { readFileSync, existsSync, readdirSync, statSync } from 'node:fs';
import { execFileSync } from 'node:child_process';
import { dirname, join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

const ROOT = join(dirname(fileURLToPath(import.meta.url)), '..');
const args = process.argv.slice(2);
const FAST = args.includes('--fast');
const ONLY = args.find((a) => !a.startsWith('--'));

const results = [];
const check = (group, name, fn) => {
  if (ONLY && !group.toLowerCase().startsWith(ONLY.toLowerCase())) return;
  let problems = [];
  try {
    problems = fn() ?? [];
  } catch (err) {
    problems = [`la verificación explotó: ${err.message}`];
  }
  results.push({ group, name, problems });
};

const readJson = (p) => JSON.parse(readFileSync(join(ROOT, p), 'utf8'));
const round = (n) => Math.round(n * 100) / 100;

// ══ CONTRATO ═══════════════════════════════════════════════════════════════

const races = readJson('docs/contract/seed/races.json');
const users = readJson('docs/contract/seed/users.json');
const bets = readJson('docs/contract/seed/bets.json');
const raceResults = readJson('docs/contract/seed/results.json');
const leaderboard = readJson('docs/contract/seed/leaderboard.json');

const raceById = new Map(races.map((r) => [r.id, r]));
const horseById = new Map(races.flatMap((r) => r.horses.map((h) => [h.id, { ...h, raceId: r.id }])));
const userById = new Map(users.map((u) => [u.id, u]));
const resultByRace = new Map(raceResults.map((r) => [r.raceId, r]));

check('contrato', 'los JSON del seed son coherentes entre sí', () => {
  const bad = [];

  const dupRaces = races.length - new Set(races.map((r) => r.id)).size;
  if (dupRaces) bad.push(`${dupRaces} id de carrera repetido`);

  const allHorseIds = races.flatMap((r) => r.horses.map((h) => h.id));
  const dupHorses = allHorseIds.length - new Set(allHorseIds).size;
  if (dupHorses) bad.push(`${dupHorses} id de caballo repetido`);

  for (const race of races) {
    const numbers = race.horses.map((h) => h.number).sort((a, b) => a - b);
    const expected = numbers.map((_, i) => i + 1);
    if (numbers.join() !== expected.join()) {
      bad.push(`${race.id}: los números de partida no son 1..${race.horses.length} (son ${numbers.join(',')})`);
    }
    if (race.horses.some((h) => h.odds < 1.01)) bad.push(`${race.id}: hay una cuota menor a 1.01`);
  }

  for (const u of users) {
    if (u.balance < 0) bad.push(`${u.id}: saldo negativo`);
  }

  return bad;
});

check('contrato', 'cada apuesta apunta a datos que existen y los copia bien', () => {
  const bad = [];
  for (const bet of bets) {
    const race = raceById.get(bet.raceId);
    const horse = horseById.get(bet.horseId);

    if (!userById.has(bet.userId)) { bad.push(`${bet.id}: userId inexistente`); continue; }
    if (!race) { bad.push(`${bet.id}: raceId inexistente`); continue; }
    if (!horse) { bad.push(`${bet.id}: horseId inexistente`); continue; }
    if (horse.raceId !== bet.raceId) { bad.push(`${bet.id}: el caballo no corre en esa carrera`); continue; }

    // Los campos denormalizados tienen que coincidir con el origen.
    if (bet.raceName !== race.name) bad.push(`${bet.id}: raceName dice "${bet.raceName}", la carrera se llama "${race.name}"`);
    if (bet.horseName !== horse.name) bad.push(`${bet.id}: horseName dice "${bet.horseName}", el caballo se llama "${horse.name}"`);
    if (bet.odds !== horse.odds) bad.push(`${bet.id}: odds ${bet.odds} ≠ cuota del caballo ${horse.odds}`);

    if (!Number.isInteger(bet.amount)) bad.push(`${bet.id}: amount no es entero`);
    if (bet.amount < 10 || bet.amount > 5000) bad.push(`${bet.id}: amount ${bet.amount} fuera de [10, 5000]`);

    const result = resultByRace.get(bet.raceId);
    const winner = result?.podium.find((p) => p.place === 1)?.horseId;

    if (race.status === 'finished') {
      const shouldWin = bet.horseId === winner;
      if (bet.status !== (shouldWin ? 'won' : 'lost')) {
        bad.push(`${bet.id}: carrera terminada, el estado debería ser ${shouldWin ? 'won' : 'lost'} y es "${bet.status}"`);
      }
      const expected = shouldWin ? round(bet.amount * bet.odds) : 0;
      if (bet.payout !== expected) bad.push(`${bet.id}: payout ${bet.payout}, esperado ${expected}`);
      if (!Number.isInteger(bet.payout)) bad.push(`${bet.id}: payout no es entero`);
    } else {
      if (bet.status !== 'pending') bad.push(`${bet.id}: la carrera no terminó pero el estado es "${bet.status}"`);
      if (bet.payout !== 0) bad.push(`${bet.id}: pendiente con payout ${bet.payout}`);
    }

    if (new Date(bet.placedAt) >= new Date(race.startsAt)) {
      bad.push(`${bet.id}: se apostó después de la largada`);
    }
  }
  return bad;
});

check('contrato', 'los resultados calzan con las carreras terminadas', () => {
  const bad = [];
  const finished = races.filter((r) => r.status === 'finished');

  for (const race of finished) {
    if (!resultByRace.has(race.id)) bad.push(`${race.id} está finished pero no tiene resultado`);
  }
  for (const result of raceResults) {
    const race = raceById.get(result.raceId);
    if (!race) { bad.push(`resultado de ${result.raceId}: la carrera no existe`); continue; }
    if (race.status !== 'finished') bad.push(`${result.raceId}: hay resultado pero el estado es "${race.status}"`);
    if (result.podium.length !== 3) bad.push(`${result.raceId}: el podio tiene ${result.podium.length} puestos, deberían ser 3`);
    if (new Date(result.finishedAt) <= new Date(race.startsAt)) bad.push(`${result.raceId}: terminó antes de largar`);

    const places = result.podium.map((p) => p.place).sort();
    if (places.join() !== '1,2,3') bad.push(`${result.raceId}: los puestos no son 1,2,3`);

    for (const p of result.podium) {
      const horse = horseById.get(p.horseId);
      if (!horse || horse.raceId !== result.raceId) { bad.push(`${result.raceId}: ${p.horseId} no corre en esa carrera`); continue; }
      if (p.horseName !== horse.name) bad.push(`${result.raceId}/${p.horseId}: horseName no coincide`);
      if (p.number !== horse.number) bad.push(`${result.raceId}/${p.horseId}: number no coincide`);
      if (p.odds !== horse.odds) bad.push(`${result.raceId}/${p.horseId}: odds no coinciden`);
    }
  }
  return bad;
});

check('contrato', 'el leaderboard golden coincide con lo calculado desde las apuestas', () => {
  const bad = [];
  const dayOf = (iso) => iso.slice(0, 10);
  const today = dayOf([...raceResults].sort((a, b) => b.finishedAt.localeCompare(a.finishedAt))[0].finishedAt);

  const compute = (period) => {
    const acc = new Map();
    for (const bet of bets) {
      if (bet.status === 'pending') continue;
      const result = resultByRace.get(bet.raceId);
      if (!result) continue;
      if (period === 'daily' && dayOf(result.finishedAt) !== today) continue;

      const e = acc.get(bet.userId) ?? { profit: 0, bets: 0, wins: 0 };
      e.profit += bet.payout - bet.amount;
      e.bets += 1;
      if (bet.status === 'won') e.wins += 1;
      acc.set(bet.userId, e);
    }
    return [...acc.entries()]
      .map(([userId, e]) => ({ userId, displayName: userById.get(userId).displayName, ...e }))
      .sort(
        (a, b) =>
          b.profit - a.profit || b.wins - a.wins || a.displayName.localeCompare(b.displayName, 'es'),
      )
      .slice(0, 20)
      .map((e, i) => ({ rank: i + 1, ...e }));
  };

  for (const period of ['all', 'daily']) {
    const expected = compute(period);
    const actual = leaderboard[period];
    if (expected.length !== actual.length) {
      bad.push(`${period}: golden tiene ${actual.length} entradas, el cálculo da ${expected.length}`);
      continue;
    }
    for (let i = 0; i < expected.length; i++) {
      const e = expected[i];
      const a = actual[i];
      for (const k of ['rank', 'userId', 'displayName', 'profit', 'bets', 'wins']) {
        if (e[k] !== a[k]) bad.push(`${period}[${i}].${k}: golden ${a[k]}, calculado ${e[k]}`);
      }
    }
  }
  return bad;
});

check('contrato', 'los samples reflejan el seed', () => {
  const bad = [];
  const dir = join(ROOT, 'docs/contract/samples');
  const files = readdirSync(dir).filter((f) => f.endsWith('.json'));
  if (!files.length) bad.push('no hay ningún sample');

  for (const f of files) {
    try {
      JSON.parse(readFileSync(join(dir, f), 'utf8'));
    } catch (err) {
      bad.push(`${f} no es JSON válido: ${err.message}`);
    }
  }

  const detail = readJson('docs/contract/samples/races.detail.200.json');
  const seedRace = raceById.get(detail.id);
  if (!seedRace) bad.push('races.detail apunta a una carrera que no está en el seed');
  else if (JSON.stringify(seedRace) !== JSON.stringify({ id: detail.id, name: detail.name, startsAt: detail.startsAt, status: detail.status, horses: detail.horses })) {
    bad.push('races.detail.200.json se separó de races.json');
  }

  for (const period of ['all', 'daily']) {
    const sample = readJson(`docs/contract/samples/leaderboard.${period}.200.json`);
    if (JSON.stringify(sample.entries) !== JSON.stringify(leaderboard[period])) {
      bad.push(`leaderboard.${period}.200.json se separó de seed/leaderboard.json`);
    }
  }

  return bad;
});

check('contrato', 'el fixture de la carrera es consistente', () => {
  const path = join(ROOT, 'docs/contract/fixtures/race-ticks.jsonl');
  if (!existsSync(path)) return ['falta race-ticks.jsonl — correr node scripts/gen-race-ticks.mjs'];

  const bad = [];
  const events = readFileSync(path, 'utf8').trim().split('\n').map((l) => JSON.parse(l));
  const ticks = events.filter((e) => e.type === 'race.tick');
  const finished = events.find((e) => e.type === 'race.finished');

  const race = raceById.get(ticks[0].raceId);
  if (!race) return ['el fixture apunta a una carrera que no está en el seed'];

  const fieldIds = new Set(race.horses.map((h) => h.id));
  const lastByHorse = new Map();

  for (const tick of ticks) {
    if (tick.positions.length !== race.horses.length) {
      bad.push(`t=${tick.t}: ${tick.positions.length} posiciones, la carrera tiene ${race.horses.length} caballos`);
      break;
    }
    for (const p of tick.positions) {
      if (!fieldIds.has(p.horseId)) { bad.push(`t=${tick.t}: ${p.horseId} no corre en ${race.id}`); break; }
      if (p.progress < 0 || p.progress > 1) { bad.push(`t=${tick.t}: progress ${p.progress} fuera de [0,1]`); break; }
      const prev = lastByHorse.get(p.horseId) ?? 0;
      if (p.progress < prev - 1e-9) { bad.push(`t=${tick.t}: ${p.horseId} retrocedió (${prev} → ${p.progress})`); break; }
      lastByHorse.set(p.horseId, p.progress);
    }
    const places = tick.positions.map((p) => p.place).sort((a, b) => a - b);
    if (places.join() !== places.map((_, i) => i + 1).join()) {
      bad.push(`t=${tick.t}: los puestos no son 1..${race.horses.length}`);
      break;
    }
  }

  const lastTick = ticks.at(-1);
  const podiumFromTick = [...lastTick.positions].sort((a, b) => a.place - b.place).slice(0, 3).map((p) => p.horseId);
  if (podiumFromTick.join() !== finished.podium.join()) {
    bad.push(`el podio de race.finished (${finished.podium.join()}) no coincide con el último tick (${podiumFromTick.join()})`);
  }

  const payoutBet = bets.find((b) => b.id === finished.payouts[0]?.betId);
  if (!payoutBet) bad.push('race.finished paga una apuesta que no está en el seed');
  else if (finished.payouts[0].amount !== round(payoutBet.amount * payoutBet.odds)) {
    bad.push(`el pago del fixture no es amount × odds de ${payoutBet.id}`);
  }

  return bad;
});

// ══ DISEÑO ═════════════════════════════════════════════════════════════════

const runScript = (file, extraArgs = []) => {
  try {
    execFileSync(process.execPath, [join(ROOT, 'scripts', file), ...extraArgs], { stdio: 'pipe' });
    return [];
  } catch (err) {
    const out = `${err.stdout ?? ''}${err.stderr ?? ''}`.trim();
    return out.split('\n').filter((l) => l.trim().startsWith('·')).map((l) => l.replace(/^\s*·\s*/, ''))
      .concat(out.includes('·') ? [] : [out.split('\n').slice(-3).join(' ')]);
  }
};

check('contrato', 'las copias sincronizadas están al día', () => runScript('sync-contract.mjs', ['--check']));

check('contrato', 'el catálogo de errores de Go coincide con error-codes.md', () => {
  const goFile = join(ROOT, 'project/backend/internal/contract/errors.go');
  if (!existsSync(goFile)) return [];

  // Los códigos documentados salen de la tabla del markdown: `CODIGO` en la
  // primera columna. Los de Go, de las constantes tipadas Code.
  const doc = new Set(
    [...readFileSync(join(ROOT, 'docs/contract/error-codes.md'), 'utf8').matchAll(/^\|\s*`([A-Z_]+)`\s*\|/gm)]
      .map((m) => m[1]),
  );
  const code = new Set(
    [...readFileSync(goFile, 'utf8').matchAll(/^\s*Code\w+\s+Code\s*=\s*"([A-Z_]+)"/gm)].map((m) => m[1]),
  );

  const bad = [];
  for (const c of doc) if (!code.has(c)) bad.push(`${c} está documentado pero no existe en errors.go`);
  for (const c of code) if (!doc.has(c)) bad.push(`${c} existe en errors.go pero no está en error-codes.md`);
  return bad;
});

/**
 * Todo lo que produce un generador. Cada uno es determinístico, así que volver
 * a correrlos no debería cambiar nada: si cambia algo, o el archivo se editó a
 * mano o alguien tocó una constante sin regenerar.
 */
const GENERATED = [
  'docs/contract/fixtures/race-ticks.jsonl',
  'docs/design/assets/silks-specimen.svg',
  'project/frontend/solution/src/app/shared/ui/silk/silks.golden.ts',
  'project/frontend/starter/src/app/shared/ui/silk/silks.golden.ts',
];

check('contrato', 'los archivos generados están al día', () => {
  const before = new Map(
    GENERATED.filter((f) => existsSync(join(ROOT, f))).map((f) => [f, readFileSync(join(ROOT, f), 'utf8')]),
  );

  for (const script of ['gen-race-ticks.mjs', 'gen-silks-specimen.mjs']) {
    const problems = runScript(script);
    if (problems.length) return problems;
  }

  const stale = [];
  for (const [file, content] of before) {
    if (readFileSync(join(ROOT, file), 'utf8') !== content) {
      stale.push(`${file} no coincide con lo que produce su generador`);
    }
  }
  return stale;
});

check('contrato', 'los mocks del frontend salen del seed', () => runScript('gen-mocks.mjs', ['--check']));

check('código', 'los identificadores están en inglés', () => runScript('check-language.mjs', ['--quiet']));

check('diseño', 'la paleta cumple contraste AA', () => runScript('check-contrast.mjs', ['--quiet']));

check('diseño', 'la gramática de sedas no colisiona dentro de una carrera', () => runScript('gen-silks-specimen.mjs'));

check('diseño', 'los tokens CSS están al día', () => runScript('gen-tokens-css.mjs', ['--check']));

check('diseño', 'los diagramas y sus SVG coinciden', () => runScript('gen-diagram-svg.mjs', ['--check']));

// Solo mira las sesiones que YA tienen su deck escrito: el de reveal es
// opcional sesión por sesión, y una sesión sin deck no es un error.
check('diseño', 'los decks de reveal salen de su slides.md', () => runScript('gen-reveal.mjs', ['--check']));

// ══ BACKEND GO ═════════════════════════════════════════════════════════════

const BACKEND = join(ROOT, 'project/backend');
const hasBackend = existsSync(join(BACKEND, 'go.mod'));

/** Corre un comando en el backend y devuelve su salida si falla. */
function runGo(args, { expectEmptyOutput = false } = {}) {
  try {
    const out = execFileSync('go', args, { cwd: BACKEND, stdio: 'pipe', encoding: 'utf8' });
    if (expectEmptyOutput && out.trim()) {
      return out.trim().split('\n').map((l) => `sin formatear: ${l}`);
    }
    return [];
  } catch (err) {
    const out = `${err.stdout ?? ''}${err.stderr ?? ''}`.trim();
    return out.split('\n').filter((l) => l.trim()).slice(0, 8);
  }
}

check('backend', 'el código está formateado', () => {
  if (!hasBackend) return [];
  // gofmt -l lista los archivos que NO están formateados. Salida vacía = todo bien.
  const out = execFileSync('gofmt', ['-l', '.'], { cwd: BACKEND, stdio: 'pipe', encoding: 'utf8' });
  return out.trim() ? out.trim().split('\n').map((f) => `${f} necesita gofmt`) : [];
});

check('backend', 'go vet no encuentra problemas', () => (hasBackend ? runGo(['vet', './...']) : []));

check('backend', 'los tests pasan, incluidos los golden del contrato', () =>
  hasBackend && !FAST ? runGo(['test', './...']) : [],
);

check('backend', 'compila para producción', () =>
  hasBackend && !FAST ? runGo(['build', '-o', process.platform === 'win32' ? 'NUL' : '/dev/null', '.']) : [],
);

// ══ CÓDIGO ANGULAR ═════════════════════════════════════════════════════════

/** Proyectos Angular que existan. En Fase 0 todavía no hay ninguno. */
const angularProjects = [
  'project/frontend/solution',
  'project/frontend/starter',
  'lab/solution',
  'lab/starter',
].filter((p) => existsSync(join(ROOT, p, 'package.json')));

function walk(dir, exts) {
  const out = [];
  for (const entry of readdirSync(dir)) {
    if (entry === 'node_modules' || entry === '.angular' || entry === 'dist') continue;
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) out.push(...walk(full, exts));
    else if (exts.some((e) => entry.endsWith(e))) out.push(full);
  }
  return out;
}

const sourceFiles = () =>
  angularProjects.flatMap((p) => {
    const src = join(ROOT, p, 'src');
    return existsSync(src) ? walk(src, ['.ts', '.html']) : [];
  });

/**
 * S10 cierra con NgModules como contexto legacy. La ruta /s10 del lab TIENE que
 * contener un NgModule — es el material de la clase. Se exime solo ahí.
 *
 * (Era la ruta /s11 hasta que S11 se repartió: docs/curriculum.md.)
 */
const isNgModuleExempt = (file) => /[\\/]lab[\\/].*[\\/]s10[\\/]/.test(file);

/**
 * Vacía los comentarios conservando los saltos de línea, para no correr los
 * números de línea del informe.
 *
 * Sin esto, un comentario que dice "acá no usamos NgModule" hace fallar la
 * verificación de NgModule. Y ese comentario es justamente el que hay que
 * escribir: explicar por qué algo NO está es parte del material de clase.
 */
function stripComments(text) {
  const keepNewlines = (match) => match.replace(/[^\n]/g, ' ');
  return text
    .replace(/\/\*[\s\S]*?\*\//g, keepNewlines) // bloque /* … */
    .replace(/<!--[\s\S]*?-->/g, keepNewlines) // comentario HTML
    .replace(/(^|[^:])\/\/[^\n]*/g, (m, before) => before + ' '.repeat(m.length - before.length)); // línea //
}

check('código', 'no hay APIs de Angular 19+ ni NgModule', () => {
  if (!angularProjects.length) return [];
  const banned = [
    [/\bresource\s*\(/, 'resource() — llega en v19'],
    [/\brxResource\b/, 'rxResource() — llega en v19'],
    [/\bhttpResource\b/, 'httpResource() — llega en v20'],
    [/\blinkedSignal\b/, 'linkedSignal() — llega en v19'],
    [/\bafterRenderEffect\b/, 'afterRenderEffect() — llega en v19'],
    [/\bprovideZonelessChangeDetection\b/, 'provideZonelessChangeDetection() — llega en v19'],
    [/\bNgModule\b/, 'NgModule — prohibido por convención'],
    [/\bviewChild\s*\(/, 'viewChild() — developer preview en 18, no va en el starter'],
    [/\bcontentChild\s*\(/, 'contentChild() — developer preview en 18, no va en el starter'],
  ];
  const bad = [];
  for (const file of sourceFiles()) {
    const lines = stripComments(readFileSync(file, 'utf8')).split('\n');
    for (const [re, why] of banned) {
      if (re.source.includes('NgModule') && isNgModuleExempt(file)) continue;
      const line = lines.findIndex((l) => re.test(l));
      if (line >= 0) bad.push(`${relative(ROOT, file)}:${line + 1} — ${why}`);
    }
  }
  return bad;
});

check('código', 'todo componente es standalone y OnPush', () => {
  if (!angularProjects.length) return [];
  const bad = [];
  for (const file of sourceFiles().filter((f) => f.endsWith('.ts'))) {
    const text = readFileSync(file, 'utf8');
    if (!/@Component\s*\(/.test(text)) continue;
    const rel = relative(ROOT, file);
    if (!/standalone\s*:\s*true/.test(text)) bad.push(`${rel} — falta standalone: true (en 18 es explícito)`);
    if (!/ChangeDetectionStrategy\.OnPush/.test(text)) bad.push(`${rel} — falta ChangeDetectionStrategy.OnPush`);
  }
  return bad;
});

check('código', 'no hay any ni console.log', () => {
  if (!angularProjects.length) return [];
  const bad = [];
  for (const file of sourceFiles().filter((f) => f.endsWith('.ts'))) {
    const rel = relative(ROOT, file);
    // Sin comentarios: nombrar `any` para explicar por qué no se usa es válido.
    stripComments(readFileSync(file, 'utf8'))
      .split('\n')
      .forEach((line, i) => {
        if (/:\s*any\b|<any>|as\s+any\b/.test(line)) bad.push(`${rel}:${i + 1} — any`);
        // console.error sí se permite: es el catch de arranque de main.ts, y un
        // error que nadie ve es peor que un log.
        if (/console\.(log|debug|info)\s*\(/.test(line)) bad.push(`${rel}:${i + 1} — console.log`);
      });
  }
  return bad;
});

check('código', 'las fuentes están auto-hospedadas', () => {
  // Un enlace a fonts.googleapis.com haría que la app dependa de la red en
  // plena clase. Es una regla dura de docs/design/tokens.md §3, y por eso se
  // escanea también el CSS y el index.html, no solo el TypeScript.
  const bad = [];
  for (const p of angularProjects) {
    const src = join(ROOT, p, 'src');
    if (!existsSync(src)) continue;

    for (const file of walk(src, ['.ts', '.html', '.css'])) {
      if (/fonts\.(googleapis|gstatic)\.com/.test(readFileSync(file, 'utf8'))) {
        bad.push(`${relative(ROOT, file)} — enlaza a Google Fonts en vez de usar la copia local`);
      }
    }
    if (!existsSync(join(src, 'fonts.css'))) {
      bad.push(`${p} — falta src/fonts.css; correr node scripts/fetch-fonts.mjs`);
    }
    if (!existsSync(join(ROOT, p, 'public/fonts'))) {
      bad.push(`${p} — falta public/fonts; correr node scripts/fetch-fonts.mjs`);
    }
  }
  return bad;
});

/** Chrome para correr Karma sin ventana. */
function chromeBinary() {
  if (process.env['CHROME_BIN']) return process.env['CHROME_BIN'];
  const candidates = [
    'C:/Program Files/Google/Chrome/Application/chrome.exe',
    'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe',
    '/usr/bin/google-chrome',
    '/usr/bin/chromium',
    '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
  ];
  return candidates.find((c) => existsSync(c));
}

check('código', 'compila en producción y tipa sin errores', () => {
  if (!angularProjects.length || FAST) return [];
  const bad = [];
  for (const p of angularProjects) {
    const cwd = join(ROOT, p);
    for (const [label, cmd] of [
      ['tsc --noEmit', ['npx', 'tsc', '--noEmit']],
      ['ng build --configuration production', ['npx', 'ng', 'build', '--configuration', 'production']],
    ]) {
      try {
        execFileSync(cmd[0], cmd.slice(1), { cwd, stdio: 'pipe', shell: process.platform === 'win32' });
      } catch (err) {
        const out = `${err.stdout ?? ''}${err.stderr ?? ''}`.trim().split('\n').slice(0, 6).join('\n      ');
        bad.push(`${p} — ${label} falló:\n      ${out}`);
      }
    }
  }
  return bad;
});

check('código', 'los tests del navegador pasan', () => {
  if (!angularProjects.length || FAST) return [];

  const chrome = chromeBinary();
  if (!chrome) return ['no se encontró Chrome ni Edge; definí CHROME_BIN para correr los tests'];

  const bad = [];
  for (const p of angularProjects) {
    // Un proyecto sin specs hace fallar a `ng test` con TS18003. Que todavía
    // no tenga tests no es un error: los va sumando cada sesión.
    const src = join(ROOT, p, 'src');
    if (!existsSync(src) || walk(src, ['.spec.ts']).length === 0) continue;

    try {
      execFileSync('npx', ['ng', 'test', '--watch=false', '--browsers=ChromeHeadless'], {
        cwd: join(ROOT, p),
        stdio: 'pipe',
        shell: process.platform === 'win32',
        env: { ...process.env, CHROME_BIN: chrome },
      });
    } catch (err) {
      const out = `${err.stdout ?? ''}${err.stderr ?? ''}`;
      const failures = out
        .split('\n')
        .filter((l) => /FAILED|Expected/.test(l))
        .slice(0, 5)
        .map((l) => l.trim());
      bad.push(`${p} — ${failures.length ? failures.join(' · ') : 'los tests fallaron'}`);
    }
  }
  return bad;
});

// ══ INFORME ════════════════════════════════════════════════════════════════

const GREY = '\x1b[90m';
const RED = '\x1b[31m';
const GREEN = '\x1b[32m';
const BOLD = '\x1b[1m';
const OFF = '\x1b[0m';

console.log('');
let group = '';
for (const r of results) {
  if (r.group !== group) {
    group = r.group;
    console.log(`  ${BOLD}${group.toUpperCase()}${OFF}`);
  }
  const ok = r.problems.length === 0;
  console.log(`  ${ok ? GREEN + '✓' : RED + '✗'}${OFF} ${r.name}`);
  for (const p of r.problems) console.log(`      ${RED}${p}${OFF}`);
}

if (!angularProjects.length) {
  console.log(`\n  ${GREY}Sin apps Angular todavía — las verificaciones de código quedan en espera.${OFF}`);
} else if (FAST) {
  console.log(`\n  ${GREY}--fast: no se corrieron los builds.${OFF}`);
}

const failed = results.filter((r) => r.problems.length);
if (failed.length) {
  console.log(`\n  ${RED}${BOLD}✗ ${failed.length} de ${results.length} verificaciones fallaron.${OFF}\n`);
  process.exit(1);
}
console.log(`\n  ${GREEN}${BOLD}✓ ${results.length} verificaciones, todo en verde.${OFF}\n`);
