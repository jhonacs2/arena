#!/usr/bin/env node
/**
 * Verificación mecánica de Arena. Se corre después de cada feature.
 *
 *   node scripts/verify-arena.mjs             todo
 *   node scripts/verify-arena.mjs --fast      saltea builds y tests
 *   node scripts/verify-arena.mjs esquema     un grupo: contrato | esquema |
 *                                             backend | frontend | diseño | deploy
 *
 * Mismo espíritu que `scripts/verify.mjs`: si falla, se arregla antes de seguir.
 * Dos diferencias, porque Arena no es material de clase:
 *
 * - **No verifica APIs de Angular 19+.** Arena es Angular 22 y ahí `resource()`,
 *   `httpResource()` y Signal Forms están permitidas — ver `arena/CLAUDE.md`.
 * - **Sí verifica la plata.** Las monedas son nota: el ledger tiene que cerrar,
 *   ningún endpoint de instructor puede quedar sin chequeo de rol y no puede
 *   haber un `float` en el camino de un monto.
 *
 * Saltea con gracia lo que todavía no existe. Un check salteado se imprime en
 * gris con el motivo y **no** rompe el exit code: mientras el backend y el
 * frontend se escriben, esto igual verifica el esquema, la paleta y el
 * despliegue. Lo que no se saltea nunca es un check que *puede* correr.
 */

import { existsSync, readFileSync, readdirSync, statSync } from 'node:fs';
import { execFileSync } from 'node:child_process';
import { dirname, join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

const ROOT = join(dirname(fileURLToPath(import.meta.url)), '..');
const ARENA = join(ROOT, 'arena');
const BACKEND = join(ARENA, 'backend');
const FRONTEND = join(ARENA, 'frontend');
const DEPLOY = join(ARENA, 'deploy');

const args = process.argv.slice(2);
const FAST = args.includes('--fast');
const ONLY = args.find((a) => !a.startsWith('--'));

/** Sin tildes y en minúscula, para que `diseño` y `diseno` filtren igual. */
const fold = (s) => s.toLowerCase().normalize('NFD').replace(/[̀-ͯ]/g, '');

/** Un check devuelve esto cuando no hay nada que verificar todavía. */
const SKIP = (reason) => ({ skip: reason });

const results = [];
const check = (group, name, fn) => {
  if (ONLY && !fold(group).startsWith(fold(ONLY))) return;
  let problems = [];
  let skip = null;
  try {
    const out = fn() ?? [];
    if (out && !Array.isArray(out) && out.skip) skip = out.skip;
    else problems = out;
  } catch (err) {
    problems = [`la verificación explotó: ${err.message}`];
  }
  results.push({ group, name, problems, skip });
};

const read = (path) => readFileSync(path, 'utf8');
const rel = (path) => relative(ROOT, path).replace(/\\/g, '/');

function walk(dir, exts, skipDirs = ['node_modules', '.angular', 'dist', 'vendor', '.git']) {
  if (!existsSync(dir)) return [];
  const out = [];
  for (const entry of readdirSync(dir)) {
    if (skipDirs.includes(entry)) continue;
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) out.push(...walk(full, exts, skipDirs));
    else if (exts.some((e) => entry.endsWith(e))) out.push(full);
  }
  return out;
}

/**
 * Vacía los comentarios conservando los saltos de línea, para no correr los
 * números de línea del informe. Igual que en `verify.mjs`: un comentario que
 * dice «acá el rol se chequea en el middleware» no puede hacer pasar —ni
 * fallar— una verificación.
 */
function stripComments(text) {
  const blank = (m) => m.replace(/[^\n]/g, ' ');
  return text
    .replace(/\/\*[\s\S]*?\*\//g, blank)
    .replace(/<!--[\s\S]*?-->/g, blank)
    .replace(/(^|[^:])\/\/[^\n]*/g, (m, before) => before + ' '.repeat(m.length - before.length));
}

/** Corre otro script del repo y devuelve sus problemas si falla. */
function runScript(file, extraArgs = []) {
  try {
    execFileSync(process.execPath, [join(ROOT, 'scripts', file), ...extraArgs], { stdio: 'pipe' });
    return [];
  } catch (err) {
    const out = `${err.stdout ?? ''}${err.stderr ?? ''}`.trim();
    const bullets = out
      .split('\n')
      .filter((l) => l.trim().startsWith('·'))
      .map((l) => l.replace(/^\s*·\s*/, ''));
    return bullets.length ? bullets : [out.split('\n').slice(-3).join(' ')];
  }
}

function hasBinary(bin, probe = ['--version']) {
  try {
    execFileSync(bin, probe, { stdio: 'pipe' });
    return true;
  } catch {
    return false;
  }
}

// ══ CONTRATO ═══════════════════════════════════════════════════════════════

const CONTRACT = join(ARENA, 'docs/contract');
const SCHEMA_PATH = join(CONTRACT, 'schema.sql');
const API_PATH = join(CONTRACT, 'api.md');

check('contrato', 'los tres documentos del contrato están en su lugar', () => {
  const bad = [];
  for (const f of ['decisiones.md', 'schema.sql', 'api.md']) {
    if (!existsSync(join(CONTRACT, f))) bad.push(`falta arena/docs/contract/${f}`);
  }
  return bad;
});

check('contrato', 'los códigos de error de api.md son únicos y en inglés', () => {
  if (!existsSync(API_PATH)) return SKIP('todavía no hay api.md');
  const codes = [...read(API_PATH).matchAll(/^\|\s*`([A-Z][A-Z_]+)`\s*\|/gm)].map((m) => m[1]);
  if (!codes.length) return ['api.md no tiene la tabla de códigos de error'];

  const bad = [];
  const seen = new Set();
  for (const c of codes) {
    if (seen.has(c)) bad.push(`${c} está dos veces en la tabla`);
    seen.add(c);
    if (!/^[A-Z][A-Z_]*[A-Z]$/.test(c)) bad.push(`${c} no tiene la forma CODIGO_EN_MAYUSCULAS`);
  }
  return bad;
});

check('contrato', 'el catálogo de errores de Go coincide con api.md', () => {
  const goFiles = walk(BACKEND, ['.go']);
  if (!goFiles.length) return SKIP('todavía no hay backend');
  if (!existsSync(API_PATH)) return SKIP('todavía no hay api.md');

  const doc = new Set([...read(API_PATH).matchAll(/^\|\s*`([A-Z][A-Z_]+)`\s*\|/gm)].map((m) => m[1]));
  const code = new Set();
  for (const f of goFiles) {
    // Constantes de error, en cualquiera de las dos formas habituales:
    //   CodeForbidden Code = "FORBIDDEN"   ·   ErrForbidden = "FORBIDDEN"
    for (const m of stripComments(read(f)).matchAll(/\b(?:Code|Err)\w*\s*(?:\w+\s*)?=\s*"([A-Z][A-Z_]+)"/g)) {
      code.add(m[1]);
    }
  }
  if (!code.size) return SKIP('el backend todavía no declara constantes de error');

  const bad = [];
  for (const c of doc) if (!code.has(c)) bad.push(`${c} está en api.md pero no existe en el backend`);
  for (const c of code) if (!doc.has(c)) bad.push(`${c} existe en el backend pero no está en api.md`);
  return bad;
});

// ══ ESQUEMA ════════════════════════════════════════════════════════════════

/** El SQL sin comentarios de línea, para que un `-- create table x` no cuente. */
const schemaSql = () => (existsSync(SCHEMA_PATH) ? read(SCHEMA_PATH).replace(/--[^\n]*/g, '') : null);

check('esquema', 'el esquema es idempotente por construcción', () => {
  const sql = schemaSql();
  if (!sql) return SKIP('todavía no hay schema.sql');
  const low = sql.toLowerCase();
  const bad = [];

  for (const m of low.matchAll(/create\s+table\s+(?!if\s+not\s+exists)([\w."]+)/g)) {
    bad.push(`create table ${m[1]} sin "if not exists": la segunda pasada explota`);
  }
  for (const m of low.matchAll(/create\s+(?:unique\s+)?index\s+(?!if\s+not\s+exists|concurrently)([\w."]+)/g)) {
    bad.push(`create index ${m[1]} sin "if not exists"`);
  }
  for (const m of low.matchAll(/create\s+view\s+([\w."]+)/g)) {
    bad.push(`create view ${m[1]} sin "or replace"`);
  }
  for (const m of low.matchAll(/create\s+function\s+([\w."(]+)/g)) {
    bad.push(`create function ${m[1]} sin "or replace"`);
  }

  // `create type` no acepta "if not exists": tiene que ir dentro de un bloque
  // do $$ … exception when duplicate_object. Se verifica por posición, no
  // contando ocurrencias, para que un bloque de más no tape uno de menos.
  for (const m of low.matchAll(/create\s+type\s+(\w+)/g)) {
    const before = low.lastIndexOf('do $$', m.index);
    const after = low.indexOf('end $$', m.index);
    const block = before >= 0 && after > m.index ? low.slice(before, after) : '';
    if (!block.includes('duplicate_object')) {
      bad.push(`create type ${m[1]} no está dentro de un do $$ … exception when duplicate_object`);
    }
  }

  // Un trigger se recrea, así que necesita su drop previo.
  for (const m of low.matchAll(/create\s+trigger\s+(\w+)/g)) {
    if (!new RegExp(`drop\\s+trigger\\s+if\\s+exists\\s+${m[1]}\\b`).test(low)) {
      bad.push(`create trigger ${m[1]} sin "drop trigger if exists ${m[1]}" antes`);
    }
  }

  return bad;
});

/** Nombres de columna que son plata. En inglés, como todo el código del repo. */
const MONEY_COLUMN =
  /^(?:\w*_)?(?:coins?|balance|amount|delta|odds|payout|stake|price|refund|prize|fee|cash|coins_granted)(?:_\w*)?$/i;

check('esquema', 'no hay punto flotante en el esquema', () => {
  const sql = schemaSql();
  if (!sql) return SKIP('todavía no hay schema.sql');
  const bad = [];
  sql.split('\n').forEach((line, i) => {
    // `numeric` y `decimal` son exactos y no entran acá: el que rompe un centavo
    // en silencio es el binario. Las columnas de plata las mira el check de abajo.
    const m = line.match(/\b(real|double\s+precision|float4|float8|float\b)/i);
    if (m) bad.push(`schema.sql:${i + 1} — ${m[1].trim()}: con binario, 2.10 × 700 da 1469.9999999999998`);
  });
  return bad;
});

check('esquema', 'las columnas de plata son enteras', () => {
  const sql = schemaSql();
  if (!sql) return SKIP('todavía no hay schema.sql');
  const INTEGER = /^(?:bigint|integer|int|smallint|bigserial|serial)$/i;
  const bad = [];

  sql.split('\n').forEach((line, i) => {
    // `nombre tipo …` dentro de un create table. El tipo es el segundo token.
    const m = line.match(/^\s{2,}(\w+)\s+([a-z]+(?:\s+precision)?)\s*(\([^)]*\))?/i);
    if (!m || !MONEY_COLUMN.test(m[1])) return;
    if (!INTEGER.test(m[2])) {
      bad.push(
        `schema.sql:${i + 1} — ${m[1]} es ${m[2]}${m[3] ?? ''}: las monedas van en unidades (bigint) ` +
          `y las cuotas ×100 (int) — decisiones.md §5`,
      );
    }
  });
  return bad;
});

check('esquema', 'las reglas de decisiones.md están en el esquema y no en una convención', () => {
  const sql = schemaSql();
  if (!sql) return SKIP('todavía no hay schema.sql');
  const flat = sql.toLowerCase().replace(/\s+/g, ' ');
  const required = [
    [/check \(balance >= 0\)/, 'el piso del saldo en 0 (decisiones.md §1)'],
    [/unique \(race_id, user_id\)/, 'una apuesta por carrera y por alumno (decisiones.md §1)'],
    [/odds\w* >= 101/, 'la cuota mínima 1.01 guardada ×100'],
    [/\^\[a-z\]\{4\}-\[0-9\]\{4\}\$/, 'el formato AAAA-9999 del código de invitación'],
    [/before update or delete on coin_transactions/, 'el trigger que hace el ledger append-only'],
    [/create or replace view user_scores/, 'los puntos como vista, no como columna'],
    [/balance \/ 100/, 'puntos = floor(saldo / 100)'],
    [/redeemed_by is null and redeemed_at is null/, 'que no exista un código a medio canjear'],
  ];
  return required.filter(([re]) => !re.test(flat)).map(([, why]) => `el esquema no expresa ${why}`);
});

// ── Postgres de verdad, si hay ─────────────────────────────────────────────

const DB_URL = process.env['DATABASE_URL'] ?? '';
const HAS_PSQL = hasBinary('psql');
/** Con el daemon corriendo, no solo el CLI: `docker info` falla si no está. */
const HAS_DOCKER = hasBinary('docker', ['info', '--format', '{{.ServerVersion}}']);
const PSQL_IMAGE = process.env['ARENA_PSQL_IMAGE'] ?? 'postgres:17-alpine';
/**
 * Contenedor donde correr el psql, con `docker exec`. Es la única manera de
 * llegar a la base en el VPS: ahí Postgres corre en la red privada del compose y
 * no publica ningún puerto, así que desde el host no hay a dónde conectarse — y
 * la reconciliación del ledger es justo el check que hay que poder correr en
 * producción. Con `ARENA_PSQL_CONTAINER=arena-db`, `DATABASE_URL` es la que ve
 * ese contenedor: `postgres://arena:…@localhost:5432/arena`.
 */
const PSQL_CONTAINER = process.env['ARENA_PSQL_CONTAINER'] ?? '';

/**
 * Corre SQL contra $DATABASE_URL. Devuelve { ok, out }, o null si no hay con qué.
 *
 * Sin `psql` en el PATH usa el que viene en la imagen de Postgres. Vale la pena el
 * rodeo: la reconciliación del ledger es el check más importante de este script y
 * no puede depender de que alguien haya instalado el cliente de Postgres.
 */
function runSql(sql, { tuplesOnly = false } = {}) {
  if (!DB_URL) return null;
  const flags = ['-v', 'ON_ERROR_STOP=1', ...(tuplesOnly ? ['-t', '-A', '-q'] : []), '-f', '-'];
  const opts = { input: sql, stdio: 'pipe', encoding: 'utf8' };
  try {
    if (PSQL_CONTAINER && HAS_DOCKER) {
      const cmd = ['exec', '-i', PSQL_CONTAINER, 'psql', DB_URL, ...flags];
      return { ok: true, out: execFileSync('docker', cmd, opts) };
    }
    if (HAS_PSQL) return { ok: true, out: execFileSync('psql', [DB_URL, ...flags], opts) };
    if (HAS_DOCKER) {
      // Desde adentro del contenedor, `localhost` es el contenedor. Un Postgres
      // que corre en la máquina se alcanza por el gateway.
      const url = DB_URL.replace(/@(?:localhost|127\.0\.0\.1)\b/, '@host.docker.internal');
      const cmd = ['run', '--rm', '-i', '--add-host=host.docker.internal:host-gateway', PSQL_IMAGE, 'psql', url];
      return { ok: true, out: execFileSync('docker', [...cmd, ...flags], opts) };
    }
    return null;
  } catch (err) {
    return { ok: false, out: `${err.stdout ?? ''}${err.stderr ?? ''}`.trim() };
  }
}

/**
 * Las líneas que importan de la salida de psql. Sin esto el informe muestra
 * `DO / DO / CREATE TABLE …` —el progreso, que sale por stdout— y esconde el
 * `ERROR:` que sale por stderr, que es lo único que se necesita leer.
 */
function sqlErrorLines(out) {
  const lines = out.split('\n').filter((l) => l.trim());
  const errors = lines.filter((l) => /^(?:psql:)?.*\b(ERROR|FATAL|DETAIL|HINT|LINE)\b/.test(l));
  return (errors.length ? errors : lines.slice(-6)).slice(0, 6);
}

/** El motivo por el que un check contra la base no puede correr. */
function dbUnavailable() {
  if (!DB_URL) return 'sin DATABASE_URL: no se probó contra Postgres de verdad';
  if (!HAS_PSQL && !HAS_DOCKER) {
    return 'hay DATABASE_URL pero no hay psql ni Docker: no hay con qué hablarle a la base';
  }
  return null;
}

check('esquema', 'el esquema aplica limpio dos veces seguidas', () => {
  const sql = schemaSql();
  if (!sql) return SKIP('todavía no hay schema.sql');
  const why = dbUnavailable();
  if (why) return SKIP(why);

  const raw = read(SCHEMA_PATH);
  for (const pass of [1, 2]) {
    const result = runSql(raw);
    if (!result) return SKIP('no se pudo hablar con Postgres');
    if (!result.ok) {
      return [
        `la pasada ${pass} falló${pass === 2 ? ' — el esquema no es idempotente' : ''}:`,
        ...sqlErrorLines(result.out),
      ];
    }
  }
  return [];
});

check('esquema', 'la reconciliación del ledger da cero diferencias', () => {
  const reconcile = join(ARENA, 'scripts/reconcile.sql');
  if (!existsSync(reconcile)) return ['falta arena/scripts/reconcile.sql'];

  // Lo que se puede verificar sin base: que el SQL compare las dos cosas que
  // tiene que comparar. Un reconcile.sql que no toca el ledger pasaría siempre.
  const text = read(reconcile).toLowerCase();
  const bad = [];
  if (!/sum\(\s*delta\s*\)/.test(text)) bad.push('reconcile.sql no suma coin_transactions.delta');
  if (!/\.balance\b/.test(text)) bad.push('reconcile.sql no compara contra users.balance');
  if (bad.length) return bad;

  const why = dbUnavailable();
  if (why) return SKIP(why);

  const result = runSql(read(reconcile), { tuplesOnly: true });
  if (!result) return SKIP('no se pudo hablar con Postgres');
  if (!result.ok) return ['reconcile.sql no corrió:', ...sqlErrorLines(result.out)];

  const rows = result.out.split('\n').map((l) => l.trim()).filter(Boolean);
  return rows.slice(0, 10).map((r) => `el ledger no cierra: ${r}`);
});

// ══ BACKEND GO ═════════════════════════════════════════════════════════════

const goSources = () => walk(BACKEND, ['.go']).filter((f) => !f.endsWith('_test.go'));
const hasGoModule = () => existsSync(join(BACKEND, 'go.mod'));

/** Corre un comando de Go en arena/backend y devuelve su salida si falla. */
function runGo(cmdArgs, { expectEmptyOutput = false } = {}) {
  try {
    const out = execFileSync('go', cmdArgs, { cwd: BACKEND, stdio: 'pipe', encoding: 'utf8' });
    if (expectEmptyOutput && out.trim()) return out.trim().split('\n');
    return [];
  } catch (err) {
    return `${err.stdout ?? ''}${err.stderr ?? ''}`.trim().split('\n').filter((l) => l.trim()).slice(0, 8);
  }
}

/**
 * Un middleware de rol, escrito como se escribe: `requireAdmin`, `RequireRole`,
 * `adminOnly`, `mustBeAdmin`, `AdminMiddleware`, `roleGuard`…
 */
const ROLE_GUARD =
  /\b(?:\w*(?:require|must|only|ensure|guard|check|assert|verify|with)\w*(?:admin|role|instructor)\w*|\w*(?:admin|role|instructor)(?:only|require[ds]?|guard|middleware|check|gate)\w*)\b/i;

/** Un chequeo de rol hecho a mano dentro del handler. */
const ROLE_IN_BODY = /\bRole(?:Admin|Instructor)\b|\.Role\s*(?:==|!=)|\brole\s*(?:==|!=)\s*"|\b[iI]sAdmin\b/;

/** Registra rutas: cubre net/http 1.22, chi, gin y echo. */
const ROUTER_CALL =
  /([A-Za-z_]\w*)\s*(?:\.\w+)*\s*\.\s*(?:Handle|HandleFunc|Method|MethodFunc|Mount|Group|Route|With|Use|Get|Post|Put|Patch|Delete|Head|Options|Any|GET|POST|PUT|PATCH|DELETE)\s*\(/;

/**
 * La sentencia que empieza en la línea `i`, acumulando hasta que los paréntesis
 * y las llaves cierren. Sin esto, un `r.Route("/admin", func(r chi.Router) { … })`
 * con el `r.Use(RequireAdmin)` adentro se leería como una ruta sin protección.
 */
function statementAt(lines, i, maxLines = 40) {
  let depth = 0;
  let text = '';
  for (let j = i; j < Math.min(lines.length, i + maxLines); j++) {
    text += (j === i ? '' : '\n') + lines[j];
    for (const ch of lines[j]) {
      if ('([{'.includes(ch)) depth++;
      else if (')]}'.includes(ch)) depth--;
    }
    if (depth <= 0 && j > i) break;
    if (depth <= 0 && /[),;]\s*$/.test(lines[j])) break;
  }
  return text;
}

/** Cuerpo de cada función del backend, por nombre. */
function goFuncBodies(texts) {
  const bodies = new Map();
  for (const text of texts.values()) {
    for (const m of text.matchAll(/^func\s+(?:\([^)]*\)\s*)?([A-Za-z_]\w*)\s*\(/gm)) {
      let depth = 0;
      let start = -1;
      for (let k = m.index; k < text.length; k++) {
        if (text[k] === '{') {
          if (start < 0) start = k;
          depth++;
        } else if (text[k] === '}') {
          depth--;
          if (depth === 0 && start >= 0) {
            bodies.set(m[1], text.slice(start, k + 1));
            break;
          }
        }
      }
    }
  }
  return bodies;
}

check('backend', 'ningún handler bajo /admin/ queda sin chequeo de rol', () => {
  const files = walk(BACKEND, ['.go']);
  if (!files.length) return SKIP('todavía no hay backend');

  const texts = new Map(files.map((f) => [f, stripComments(read(f))]));
  const all = [...texts.values()].join('\n');
  const bodies = goFuncBodies(texts);

  /**
   * ¿El router `v` está detrás de un middleware de rol? Valen dos formas:
   *
   * - se **crea o configura** con el guard: `v := r.With(RequireAdmin)`,
   *   `v.Use(RequireAdmin)`;
   * - se **monta envuelto** por el guard: `mux.Handle("/api/admin/", requireAdmin(v))`.
   *
   * Lo que no vale es una línea que apenas lo menciona. Si valiera, un solo
   * `mux.Handle("/api/admin/", requireAdmin(sub))` blanquearía todas las demás
   * rutas colgadas de `mux`. Por eso la segunda forma exige que `v` aparezca
   * *adentro* del paréntesis del guard, y compara el nombre respetando
   * mayúsculas: `mux` no es `adminMux`.
   */
  const guardedRouter = (v) => {
    const setup = new RegExp(`\\b${v}\\b\\s*(?::=|=[^=]|\\.\\s*Use\\s*\\()`);
    const named = new RegExp(`\\b${v}\\b`);
    return all.split('\n').some((line) => {
      const guard = line.match(ROLE_GUARD);
      if (!guard) return false;
      if (setup.test(line)) return true;
      const at = line.search(named);
      return at > guard.index && line.slice(guard.index, at).includes('(');
    });
  };

  let adminRoutes = 0;
  const bad = [];

  for (const [file, text] of texts) {
    const lines = text.split('\n');
    lines.forEach((line, i) => {
      const path = line.match(/"([^"]*\/admin[^"]*)"/);
      if (!path || !ROUTER_CALL.test(line)) return;
      adminRoutes++;

      const statement = statementAt(lines, i);
      if (ROLE_GUARD.test(statement)) return;

      // El router que recibe la ruta, y —si la sentencia asigna— el que produce.
      const receiver = statement.match(ROUTER_CALL)?.[1];
      const assigned = statement.match(/^\s*(?:var\s+)?([A-Za-z_]\w*)\s*(?::=|=[^=])/)?.[1];
      if ([receiver, assigned].filter(Boolean).some(guardedRouter)) return;

      // Último recurso: el handler chequea el rol adentro.
      const after = statement.slice(statement.indexOf(path[0]) + path[0].length);
      const names = [...after.matchAll(/[A-Za-z_]\w*/g)].map((m) => m[0]);
      if (names.some((n) => bodies.has(n) && (ROLE_GUARD.test(bodies.get(n)) || ROLE_IN_BODY.test(bodies.get(n))))) {
        return;
      }

      // Una sola línea y sin espacios de más: el informe tiene que quedar en un
      // renglón para poder pegarlo en un mensaje.
      let handler = after.split('\n').find((l) => /\w/.test(l)) ?? '';
      handler = handler.replace(/^[\s,]+/, '').replace(/[\s),;{]+$/, '').replace(/\s+/g, ' ').slice(0, 60);
      if (/^func\b/.test(handler) || !handler) handler = 'el grupo';
      bad.push(`${rel(file)}:${i + 1} — ${path[1]} → ${handler} no pasa por ningún chequeo de rol`);
    });
  }

  if (!adminRoutes && !bad.length) return SKIP('el backend todavía no registra rutas /admin/');
  return bad;
});

check('backend', 'no hay float en el camino de un monto', () => {
  const files = goSources();
  if (!files.length) return SKIP('todavía no hay backend');

  // Identificadores de plata. Están en inglés porque el código está en inglés.
  const MONEY =
    /\b\w*(?:coin|balance|amount|delta|odds|payout|stake|bet|wager|gift|point|credit|debit|ledger|transaction|money|price|fee|cash|prize|refund)\w*\b/i;
  const FLOAT = /\bfloat(?:32|64)\b/;
  // Un struct que *es* plata: cualquier float adentro está mal, se llame como se llame.
  const MONEY_STRUCT = /\b\w*(?:bet|coin|ledger|transaction|wallet|balance|payout|odds|gift|score|money|prize)\w*\b/i;

  const bad = [];
  for (const file of files) {
    const lines = stripComments(read(file)).split('\n');
    let structName = null;
    let depth = 0;

    lines.forEach((line, i) => {
      const open = line.match(/type\s+(\w+)\s+struct\s*\{/);
      if (open) {
        structName = open[1];
        depth = 1;
      } else if (structName) {
        depth += (line.match(/\{/g) ?? []).length - (line.match(/\}/g) ?? []).length;
        if (depth <= 0) structName = null;
      }

      if (!FLOAT.test(line)) return;
      if (MONEY.test(line)) {
        bad.push(`${rel(file)}:${i + 1} — ${line.trim()}: los montos son int64 y las cuotas int ×100`);
      } else if (structName && MONEY_STRUCT.test(structName)) {
        bad.push(`${rel(file)}:${i + 1} — float dentro de ${structName}: esa struct es plata, va en enteros`);
      }
    });
  }
  return bad;
});

check('backend', 'el código está formateado', () => {
  if (!hasGoModule()) return SKIP('todavía no hay backend');
  const out = execFileSync('gofmt', ['-l', '.'], { cwd: BACKEND, stdio: 'pipe', encoding: 'utf8' });
  return out.trim() ? out.trim().split('\n').map((f) => `${f} necesita gofmt`) : [];
});

check('backend', 'go vet no encuentra problemas', () =>
  hasGoModule() ? runGo(['vet', './...']) : SKIP('todavía no hay backend'),
);

check('backend', 'compila para producción', () => {
  if (!hasGoModule()) return SKIP('todavía no hay backend');
  if (FAST) return SKIP('--fast');
  return runGo(['build', '-o', process.platform === 'win32' ? 'NUL' : '/dev/null', './...']);
});

check('backend', 'los tests pasan', () => {
  if (!hasGoModule()) return SKIP('todavía no hay backend');
  if (FAST) return SKIP('--fast');
  if (!walk(BACKEND, ['_test.go']).length) return SKIP('el backend todavía no tiene tests');
  return runGo(['test', './...']);
});

check('backend', 'las variables de entorno que lee el backend están en deploy/.env.example', () => {
  const files = goSources();
  const example = join(DEPLOY, '.env.example');
  if (!files.length) return SKIP('todavía no hay backend');
  if (!existsSync(example)) return ['falta arena/deploy/.env.example'];

  const declared = new Set([...read(example).matchAll(/^\s*(?:#\s*)?([A-Z][A-Z0-9_]*)=/gm)].map((m) => m[1]));
  const used = new Set();
  for (const f of files) {
    // `env("X", …)` pero también sus variantes tipadas `envInt`, `envBool`,
    // `envDuration`: con `\benv\(` a secas, `envInt("DB_MAX_CONNS", 10)` pasaba
    // sin documentar y el check daba una falsa tranquilidad.
    for (const m of stripComments(read(f)).matchAll(/(?:os\.(?:Getenv|LookupEnv)|\benv\w*)\s*\(\s*"([A-Z][A-Z0-9_]*)"/g)) {
      used.add(m[1]);
    }
  }
  return [...used]
    .filter((v) => !declared.has(v))
    .map((v) => `el backend lee ${v} y no está documentada en arena/deploy/.env.example`);
});

// ══ FRONTEND ═══════════════════════════════════════════════════════════════

const hasFrontend = () => existsSync(join(FRONTEND, 'package.json'));
const frontendSources = (exts) => walk(join(FRONTEND, 'src'), exts);

check('frontend', 'no hay any ni console.log', () => {
  if (!hasFrontend()) return SKIP('todavía no hay frontend');
  const files = frontendSources(['.ts']);
  if (!files.length) return SKIP('el frontend todavía no tiene código en src/');

  const bad = [];
  for (const file of files) {
    stripComments(read(file))
      .split('\n')
      .forEach((line, i) => {
        if (/:\s*any\b|<any>|as\s+any\b/.test(line)) bad.push(`${rel(file)}:${i + 1} — any`);
        // console.error se permite: es el catch de arranque de main.ts.
        if (/console\.(log|debug|info)\s*\(/.test(line)) bad.push(`${rel(file)}:${i + 1} — console.log`);
      });
  }
  return bad;
});

/** Corre un comando de npm/npx en arena/frontend. */
function runFrontend(label, cmd) {
  try {
    execFileSync(cmd[0], cmd.slice(1), {
      cwd: FRONTEND,
      stdio: 'pipe',
      shell: process.platform === 'win32',
    });
    return [];
  } catch (err) {
    const out = `${err.stdout ?? ''}${err.stderr ?? ''}`.trim().split('\n').slice(0, 6).join('\n      ');
    return [`${label} falló:\n      ${out}`];
  }
}

check('frontend', 'tipa sin errores', () => {
  if (!hasFrontend()) return SKIP('todavía no hay frontend');
  if (FAST) return SKIP('--fast');
  if (!existsSync(join(FRONTEND, 'node_modules'))) return SKIP('faltan las dependencias: npm ci en arena/frontend');

  /**
   * `npx tsc --noEmit` a secas **no compila nada** en un proyecto de Angular: el
   * `tsconfig.json` de la raíz es del estilo solución —`"files": []` más
   * `references`— así que tsc no encuentra ningún archivo de entrada y sale con 0.
   * El check pasaba con un error de tipos plantado a propósito, y así lo
   * descubrimos. Hay que apuntar a cada proyecto real.
   */
  const projects = ['tsconfig.app.json', 'tsconfig.spec.json'].filter((p) => existsSync(join(FRONTEND, p)));
  if (!projects.length) return runFrontend('tsc --noEmit', ['npx', 'tsc', '--noEmit']);

  return projects.flatMap((p) => runFrontend(`tsc --noEmit -p ${p}`, ['npx', 'tsc', '--noEmit', '-p', p]));
});

check('frontend', 'compila en producción', () => {
  if (!hasFrontend()) return SKIP('todavía no hay frontend');
  if (FAST) return SKIP('--fast');
  if (!existsSync(join(FRONTEND, 'node_modules'))) return SKIP('faltan las dependencias: npm ci en arena/frontend');
  return runFrontend('ng build --configuration production', ['npx', 'ng', 'build', '--configuration', 'production']);
});

// ══ DISEÑO ═════════════════════════════════════════════════════════════════

check('diseño', 'la paleta cumple contraste AA', () => runScript('check-contrast.mjs', ['--quiet']));

// ══ DEPLOY ═════════════════════════════════════════════════════════════════

const DEPLOY_FILES = [
  'Dockerfile',
  'docker-compose.yml',
  'docker-compose.prod.yml',
  '.env.example',
  'README.md',
  'arena-api.service',
  'cloudflared/config.yml',
  'cloudflared/cloudflared.service',
  'cloudflare/api-proxy.js',
  'cloudflare/wrangler.toml',
];

/**
 * Vacía los comentarios `#` conservando los saltos de línea. Los comentarios no
 * cuentan, igual que en `verify.mjs`: sin esto, el comentario del Dockerfile que
 * explica *por qué* va `CGO_ENABLED=0` hacía pasar el check aunque la línea que
 * lo usaba ya no estuviera. Lo descubrió la prueba en negativo, no la lectura.
 */
const stripHash = (text) =>
  text.replace(/(^|\s)#[^\n]*/g, (m, before) => before + ' '.repeat(m.length - before.length));

/** El contenido de un archivo de deploy, o null. Con `code`, sin comentarios. */
const deployFile = (name, { code = false } = {}) => {
  const p = join(DEPLOY, name);
  if (!existsSync(p)) return null;
  const text = read(p);
  return code ? stripHash(text) : text;
};

check('deploy', 'están todos los archivos de despliegue', () =>
  DEPLOY_FILES.filter((f) => !existsSync(join(DEPLOY, f))).map((f) => `falta arena/deploy/${f}`),
);

check('deploy', 'el Dockerfile produce un binario estático en una imagen sin shell', () => {
  const text = deployFile('Dockerfile', { code: true });
  if (!text) return SKIP('todavía no hay Dockerfile');
  const required = [
    [/AS\s+build/i, 'dos etapas (AS build)'],
    [/CGO_ENABLED=0/, 'CGO_ENABLED=0, para que el binario corra sin libc'],
    [/FROM\s+\S*distroless/, 'una imagen final distroless'],
    [/^USER\s+nonroot/m, 'USER nonroot: el proceso no corre como root'],
  ];
  return required.filter(([re]) => !re.test(text)).map(([, why]) => `al Dockerfile le falta ${why}`);
});

check('deploy', 'nada publica el puerto del backend fuera de la máquina', () => {
  const bad = [];

  // El de producción también, y sobre todo: es el que corre en el VPS que
  // comparte máquina con otras cosas.
  for (const name of ['docker-compose.yml', 'docker-compose.prod.yml']) {
    const compose = deployFile(name, { code: true });
    if (!compose) continue;
    // Una publicación sin interfaz (`"8080:8080"`) escucha en 0.0.0.0. El backend
    // no tiene que ser alcanzable desde afuera: el túnel sale hacia Cloudflare.
    for (const m of compose.matchAll(/^\s*-\s*"?([\d.:]+:)?(\d+):(\d+)"?\s*$/gm)) {
      const iface = m[1] ?? '';
      if (!iface.startsWith('127.0.0.1:')) {
        bad.push(`${name} publica ${m[0].trim()} sin atarlo a 127.0.0.1`);
      }
    }
    // Y la misma publicación con el puerto en una variable —`"127.0.0.1:${API_PORT}:8080"`—,
    // que el patrón de arriba no ve porque `${…}` no son dígitos. Es justo la
    // forma que tienen los dos composes, así que sin esto el check no miraba nada.
    for (const m of compose.matchAll(/^\s*-\s*"([^"\n]*\$\{[^"\n]*:\d+)"\s*$/gm)) {
      if (!m[1].startsWith('127.0.0.1:')) {
        bad.push(`${name} publica ${m[0].trim()} sin atarlo a 127.0.0.1`);
      }
    }
  }

  // Acá sí se lee el README con sus comentarios: es prosa, y un procedimiento que
  // le dice al usuario que abra un puerto es igual de malo escrito en prosa.
  for (const f of [
    'README.md',
    'docker-compose.yml',
    'docker-compose.prod.yml',
    'arena-api.service',
    'cloudflared/config.yml',
  ]) {
    const text = deployFile(f);
    if (!text) continue;
    if (/ufw\s+allow\s+\d+|firewall-cmd\s+--add-port|0\.0\.0\.0:\d+/.test(text)) {
      bad.push(`${f} abre un puerto de entrada: el punto del túnel es que no haya ninguno`);
    }
  }

  return bad;
});

check('deploy', 'el túnel apunta al backend por loopback y cierra con un 404', () => {
  const text = deployFile('cloudflared/config.yml', { code: true });
  if (!text) return SKIP('todavía no hay cloudflared/config.yml');
  const bad = [];

  for (const [key, why] of [
    [/^tunnel:\s*\S+/m, 'tunnel: con el id o el nombre del túnel'],
    [/^credentials-file:\s*\S+/m, 'credentials-file: con el JSON de credenciales'],
    [/^ingress:/m, 'ingress:'],
  ]) {
    if (!key.test(text)) bad.push(`a config.yml le falta ${why}`);
  }

  // `- service: …` (una regla sin hostname) y `service: …` (dentro de una regla).
  const services = [...text.matchAll(/^\s*(?:-\s*)?service:\s*(\S+)/gm)].map((m) => m[1]);
  if (!services.length) bad.push('config.yml no declara ningún service');
  for (const s of services) {
    if (s.startsWith('http_status:')) continue;
    if (!/^https?:\/\/(127\.0\.0\.1|localhost)(:\d+)?/.test(s)) {
      bad.push(`el ingress apunta a ${s}: tiene que ser loopback, el backend solo escucha ahí`);
    }
  }
  // La última regla es el catch-all. Sin ella, cloudflared no arranca.
  if (services.length && !services.at(-1).startsWith('http_status:')) {
    bad.push('la última regla del ingress tiene que ser un catch-all (service: http_status:404)');
  }

  return bad;
});

check('deploy', 'el túnel y el compose de producción hablan del mismo puerto', () => {
  const tunnel = deployFile('cloudflared/config.yml', { code: true });
  const compose = deployFile('docker-compose.prod.yml', { code: true });
  if (!tunnel || !compose) return SKIP('falta config.yml o docker-compose.prod.yml');

  // El puerto al que apunta el ingress y el que publica el compose son los dos
  // extremos del mismo cable. Tenerlos distintos da un 502 que no dice por qué, y
  // se descubre con la clase empezada.
  const ingress = tunnel.match(/service:\s*https?:\/\/(?:127\.0\.0\.1|localhost):(\d+)/);
  const published = compose.match(/"127\.0\.0\.1:\$\{API_PORT:-(\d+)\}:\d+"/);
  if (!ingress) return ['cloudflared/config.yml no apunta a un puerto de loopback'];
  if (!published) return ['docker-compose.prod.yml no publica el backend con ${API_PORT:-…}'];

  return ingress[1] === published[1]
    ? []
    : [
        `el túnel va al ${ingress[1]} y el compose publica el ${published[1]}: ` +
          'uno de los dos quedó sin cambiar',
      ];
});

check('deploy', 'las units de systemd arrancan solas y se reinician', () => {
  const bad = [];
  for (const f of ['arena-api.service', 'cloudflared/cloudflared.service']) {
    const text = deployFile(f, { code: true });
    if (!text) continue;
    for (const [re, why] of [
      [/^\[Unit\]/m, '[Unit]'],
      [/^\[Service\]/m, '[Service]'],
      [/^\[Install\]/m, '[Install]'],
      [/^Restart=/m, 'Restart='],
      [/^WantedBy=multi-user\.target/m, 'WantedBy=multi-user.target'],
    ]) {
      if (!re.test(text)) bad.push(`${f} no tiene ${why}`);
    }
  }
  return bad;
});

check('deploy', 'el compose y las units solo usan variables documentadas en .env.example', () => {
  const example = deployFile('.env.example');
  if (!example) return SKIP('todavía no hay .env.example');
  const declared = new Set([...example.matchAll(/^\s*(?:#\s*)?([A-Z][A-Z0-9_]*)=/gm)].map((m) => m[1]));

  const bad = [];
  for (const f of ['docker-compose.yml', 'docker-compose.prod.yml', 'arena-api.service']) {
    const text = deployFile(f, { code: true });
    if (!text) continue;
    for (const m of text.matchAll(/\$\{([A-Z][A-Z0-9_]*)(?::?[-?][^}]*)?\}/g)) {
      if (!declared.has(m[1])) bad.push(`${f} usa \${${m[1]}} y no está en .env.example`);
    }
  }
  return bad;
});

check('deploy', 'cada variable de .env.example tiene un comentario que la explica', () => {
  const example = deployFile('.env.example');
  if (!example) return SKIP('todavía no hay .env.example');
  const lines = example.split('\n');
  const bad = [];

  lines.forEach((line, i) => {
    const m = line.match(/^([A-Z][A-Z0-9_]*)=/);
    if (!m) return;
    // Se acepta el comentario en las líneas de arriba (salteando otras variables
    // del mismo bloque) o al final de la propia línea.
    let commented = /#\s*\S/.test(line.replace(/^[^#]*/, ''));
    for (let j = i - 1; j >= 0 && !commented; j--) {
      if (lines[j].trim().startsWith('#')) commented = true;
      else if (lines[j].trim() !== '') break;
    }
    if (!commented) bad.push(`${m[1]} no tiene ningún comentario que explique qué es`);
  });

  return bad;
});

check('deploy', 'el README avisa de la pausa de Supabase a los 7 días', () => {
  const text = deployFile('README.md');
  if (!text) return SKIP('todavía no hay README.md');
  const low = fold(text);
  const bad = [];
  if (!/(7|siete)\s*d[ií]as/.test(low)) bad.push('el README no menciona los 7 días de inactividad');
  if (!/pausa|pausad/.test(low)) bad.push('el README no dice que el proyecto se pausa');
  if (!/ping|cron|keep-?alive|latido/.test(low)) bad.push('el README no ofrece la salida del ping programado');
  if (!/plan\s+(pago|de\s+pago)|pro\b/.test(low)) bad.push('el README no ofrece la salida del plan pago');
  return bad;
});

check('deploy', 'el README dice que el túnel no autentica', () => {
  const text = deployFile('README.md');
  if (!text) return SKIP('todavía no hay README.md');
  const low = fold(text);
  // arena/CLAUDE.md §6 lo deja escrito y el README de despliegue es donde se lee
  // el día que se despliega: ocultar el origen no es un control de acceso.
  const bad = [];
  if (!/no autentica|no es un control de acceso|no reemplaza/.test(low)) {
    bad.push('el README no aclara que el túnel oculta el origen pero no autentica');
  }
  if (!/jwt/.test(low) || !/rol/.test(low)) {
    bad.push('el README no señala que la seguridad real son los JWT y la validación por rol');
  }
  return bad;
});

check('deploy', 'el compose es válido para Docker', () => {
  const compose = deployFile('docker-compose.yml');
  if (!compose) return SKIP('todavía no hay docker-compose.yml');
  if (!hasBinary('docker')) return SKIP('no hay docker en el PATH');
  try {
    execFileSync('docker', ['compose', '-f', join(DEPLOY, 'docker-compose.yml'), 'config', '-q'], {
      cwd: DEPLOY,
      stdio: 'pipe',
      encoding: 'utf8',
      env: { ...process.env, JWT_SECRET: 'x'.repeat(32), POSTGRES_PASSWORD: 'x', DATABASE_URL: 'postgres://x' },
    });
    return [];
  } catch (err) {
    const out = `${err.stdout ?? ''}${err.stderr ?? ''}`.trim();
    if (/daemon|dockerDesktop|pipe\/docker/i.test(out)) return SKIP('el daemon de Docker no está corriendo');
    return out.split('\n').filter((l) => l.trim()).slice(0, 6);
  }
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
  if (r.skip) {
    console.log(`  ${GREY}⊘ ${r.name} — ${r.skip}${OFF}`);
    continue;
  }
  const ok = r.problems.length === 0;
  console.log(`  ${ok ? GREEN + '✓' : RED + '✗'}${OFF} ${r.name}`);
  for (const p of r.problems) console.log(`      ${RED}${p}${OFF}`);
}

const skipped = results.filter((r) => r.skip);
const failed = results.filter((r) => !r.skip && r.problems.length);
const ran = results.length - skipped.length;

if (skipped.length) {
  console.log(`\n  ${GREY}${skipped.length} verificaciones salteadas: lo que todavía no existe no puede fallar.${OFF}`);
}
if (FAST) console.log(`  ${GREY}--fast: no se corrieron builds ni tests.${OFF}`);

if (failed.length) {
  console.log(`\n  ${RED}${BOLD}✗ ${failed.length} de ${ran} verificaciones fallaron.${OFF}\n`);
  process.exit(1);
}
console.log(`\n  ${GREEN}${BOLD}✓ ${ran} verificaciones, todo en verde.${OFF}\n`);
