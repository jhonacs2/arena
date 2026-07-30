#!/usr/bin/env node
/**
 * Punta a punta contra un backend que está corriendo de verdad.
 *
 *   node arena/scripts/e2e.mjs
 *
 *   ARENA_E2E_URL        http://localhost:8099   (por defecto)
 *   ARENA_E2E_USER       instructor
 *   ARENA_E2E_PASSWORD   la del instructor. Sin esto no corre.
 *
 * Prueba **lo que ningún test de Go cubre: el cableado**. Los tests de paquete
 * verifican cada pieza con su doble o contra Postgres, pero nadie comprueba que
 * `main.go` haya enchufado la regla de liquidación, el hub y el runner. Un backend
 * con `Rule` nula compila, pasa los 12 paquetes en verde, y no liquida una sola
 * carrera.
 *
 * Y verifica la economía sobre datos reales, de punta a punta:
 *
 *   · el piso de 10 puntos, con un alumno fundido de verdad
 *   · la conservación del pool: el total de monedas del curso no se mueve
 *   · que los puntos regalados se sumen POR ENCIMA del piso
 *   · el reparto del resto de la división, que es donde se pierde una moneda
 *
 * Corre contra una base con datos y no los borra. Crea sus propios códigos,
 * alumnos y carrera, así que se puede correr contra la instancia de desarrollo sin
 * pisar nada — pero **no contra la de producción**: deja tres alumnos de prueba.
 */

const BASE = (process.env.ARENA_E2E_URL ?? 'http://localhost:8099').replace(/\/$/, '');
const API = `${BASE}/api`;
const USER = process.env.ARENA_E2E_USER ?? 'instructor';
const PASSWORD = process.env.ARENA_E2E_PASSWORD;

const RED = '\x1b[31m';
const GREEN = '\x1b[32m';
const GREY = '\x1b[90m';
const BOLD = '\x1b[1m';
const OFF = '\x1b[0m';

if (!PASSWORD) {
  console.log(`\n  ${GREY}Sin ARENA_E2E_PASSWORD: no hay contra qué correr. Se saltea.${OFF}\n`);
  process.exit(0);
}

const failures = [];

function check(label, got, want) {
  const ok = JSON.stringify(got) === JSON.stringify(want);
  console.log(`  ${ok ? GREEN + '✓' : RED + '✗'}${OFF} ${label}${ok ? '' : ` — dio ${got}, se esperaba ${want}`}`);
  if (!ok) failures.push(label);
}

async function call(method, path, body, token) {
  const res = await fetch(API + path, {
    method,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  const text = await res.text();
  if (!res.ok) {
    throw new Error(`${method} ${path} → ${res.status}: ${text.slice(0, 300)}`);
  }
  return text ? JSON.parse(text) : {};
}

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

// Único nombre de usuario por corrida: el script se puede correr dos veces seguidas
// sin chocar con `USERNAME_TAKEN`.
const run = Number(process.hrtime.bigint() % 100000n);

try {
  console.log(`\n  ${BOLD}ARENA · punta a punta${OFF}  ${GREY}${BASE}${OFF}\n`);

  // ── El instructor reparte códigos ────────────────────────────────────────
  const admin = await call('POST', '/auth/login', { username: USER, password: PASSWORD });
  const tok = admin.accessToken;

  const { codes } = await call('POST', '/admin/codes',
    { count: 3, coinsGranted: 1000, note: `e2e ${run}` }, tok);
  console.log(`  ${GREY}códigos: ${codes.join(' · ')}${OFF}`);

  // ── Los alumnos canjean ──────────────────────────────────────────────────
  const people = [
    ['Ana', 'Gómez', `ana${run}`],
    ['Bruno', 'Díaz', `bruno${run}`],
    ['Carla', 'Ruiz', `carla${run}`],
  ];
  const students = {};
  for (const [i, [firstName, lastName, username]] of people.entries()) {
    const out = await call('POST', '/auth/redeem', {
      code: codes[i], firstName, lastName, username, password: 'Alumno1234!',
    });
    students[username] = { token: out.accessToken, id: out.user.id };
    check(`${username} arranca con 1000 monedas`, out.balance, 1000);
    check(`${username} arranca con 10 puntos`, out.points, 10);
  }
  const [ana, bruno, carla] = people.map(([, , u]) => u);

  // ── La carrera ───────────────────────────────────────────────────────────
  const created = await call('POST', '/admin/races', {
    name: 'Clásico de prueba',
    horses: [
      { number: 1, name: 'Tormenta', odds: 250 },
      { number: 2, name: 'Relámpago', odds: 410 },
      { number: 3, name: 'Viento Norte', odds: 180 },
    ],
  }, tok);
  // Ojo: POST /admin/races envuelve en `race`; GET /races/{id} devuelve la carrera
  // pelada. Es una inconsistencia del contrato, no un error de acá.
  const raceID = created.race.id;
  const horses = Object.fromEntries(created.race.horses.map((h) => [h.number, h.id]));

  await call('POST', `/admin/races/${raceID}/open`, undefined, tok);

  // Los mismos montos que `parimutuel_test.go` y `schema.test.sql`: si gana el 3,
  // el pool es 800 sobre 300 y el reparto deja resto.
  const plan = [[ana, 3, 100], [bruno, 3, 200], [carla, 1, 500]];
  for (const [username, horse, amount] of plan) {
    await call('POST', `/races/${raceID}/join`, undefined, students[username].token);
    const out = await call('POST', `/races/${raceID}/bet`,
      { horseId: horses[horse], amount }, students[username].token);
    check(`${username} pagó su apuesta de ${amount}`, out.balance, 1000 - amount);
  }

  // ── Largar y esperar la llegada ──────────────────────────────────────────
  await call('POST', `/admin/races/${raceID}/start`, undefined, tok);
  console.log(`  ${GREY}largó — la simulación tarda unos 45 s${OFF}`);

  let race;
  for (let i = 0; i < 120; i++) {
    await sleep(1000);
    race = await call('GET', `/races/${raceID}`, undefined, tok);
    if (race.status === 'finished') break;
  }
  if (race.status !== 'finished') {
    throw new Error(`la carrera quedó en "${race.status}" después de 120 s`);
  }
  const winner = race.results.find((r) => r.position === 1);
  console.log(`  ${GREY}ganó ${winner.horseName} (número ${winner.number})${OFF}`);

  // ── La economía ──────────────────────────────────────────────────────────
  const scoresOf = async () => {
    const { items } = await call('GET', '/admin/scores', undefined, tok);
    return Object.fromEntries(items.map((s) => [s.username, s]));
  };
  let scores = await scoresOf();
  const mine = [ana, bruno, carla].map((u) => scores[u]);

  // Pari-mutuel es suma cero: se apostaron 800 y el pool volvió entero a los
  // alumnos, así que los tres juntos siguen teniendo las 3000 con las que empezaron
  // pase lo que pase en la carrera.
  check('el pool vuelve entero: los tres suman 3000',
    mine.reduce((sum, s) => sum + s.balance, 0), 3000);

  // Uno de los tres perdió su apuesta. Sin el piso tendría menos de 10 puntos.
  const below = mine.filter((s) => s.points < 10).map((s) => s.username);
  check('nadie queda por debajo del piso de 10 puntos', below, []);

  // Si el backend mangleara UTF-8, los apellidos saldrían mal en el panel de notas.
  check('el apellido con acento vuelve intacto', scores[ana].lastName, 'Gómez');
  check('el nombre de la carrera vuelve intacto', race.name, 'Clásico de prueba');

  // ── El piso, forzado ─────────────────────────────────────────────────────
  // Se funde al que menos tiene, para ver el piso actuando y no deducirlo.
  const broke = mine.reduce((a, b) => (a.balance <= b.balance ? a : b)).username;
  if (scores[broke].balance > 0) {
    await call('POST', `/admin/users/${students[broke].id}/gift`,
      { coins: -scores[broke].balance, note: 'e2e: fundir para probar el piso' }, tok);
  }
  scores = await scoresOf();
  check(`${broke} con 0 monedas conserva 10 puntos`, scores[broke].points, 10);

  // Y el regalo de puntos se suma POR ENCIMA del piso, no compite con él.
  await call('POST', `/admin/users/${students[broke].id}/grant-points`,
    { points: 250, reason: 'e2e: explicó @for en el code review' }, tok);
  scores = await scoresOf();
  check(`${broke} suma el regalo sobre el piso`, scores[broke].points, 12.5);
} catch (err) {
  console.log(`\n  ${RED}${BOLD}✗ ${err.message}${OFF}\n`);
  process.exit(1);
}

if (failures.length) {
  console.log(`\n  ${RED}${BOLD}✗ ${failures.length} comprobaciones fallaron.${OFF}\n`);
  process.exit(1);
}
console.log(`\n  ${GREEN}${BOLD}✓ punta a punta en verde.${OFF}\n`);
