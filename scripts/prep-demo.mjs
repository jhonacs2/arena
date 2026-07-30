#!/usr/bin/env node
/**
 * Prepara las copias donde se da el live coding.
 *
 *   node scripts/prep-demo.mjs           antes de cada clase
 *   node scripts/prep-demo.mjs --clean   las borra
 *
 * El problema que resuelve: el guión decía «trabajás sobre `lab/solution` y
 * borrá `sessions/s01` antes de empezar». Eso es pedirle al instructor que
 * mutile su propia solución cinco minutos antes de dar la clase, con el riesgo
 * de olvidarse de restaurarla. Y era innecesario: el lienzo correcto para el
 * live coding —el proyecto en el estado justo anterior a esta sesión— es
 * exactamente lo que ya es `starter/`.
 *
 * Entonces hay tres copias con tres dueños, y ninguna se pisa:
 *
 *   solution/   la referencia. Solo la ve el instructor. Nunca se toca en clase.
 *   starter/    lo que reciben los alumnos. Es lo que se publica.
 *   demo/       copia descartable de starter/ donde el instructor escribe en vivo.
 *
 * `demo/` está en `.gitignore`: es de usar y tirar. Se regenera en un segundo
 * cuantas veces haga falta, así que se puede ensayar el bloque tres veces y
 * arrancar limpio cada vez.
 *
 * `node_modules` no se copia: se enlaza al del starter con una junction, que en
 * Windows no pide permisos de administrador. Sin eso, preparar la clase serían
 * dos `npm install` y varios minutos.
 */

import { cpSync, rmSync, existsSync, symlinkSync, mkdirSync, readdirSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const ROOT = join(dirname(fileURLToPath(import.meta.url)), '..');
const CLEAN = process.argv.includes('--clean');

/** starter → demo, por proyecto. */
const COPIES = [
  ['lab/starter', 'lab/demo'],
  ['project/frontend/starter', 'project/frontend/demo'],
];

const SKIP = new Set(['node_modules', 'dist', '.angular', 'out-tsc', 'coverage', '.git']);

if (CLEAN) {
  for (const [, target] of COPIES) {
    const full = join(ROOT, target);
    if (existsSync(full)) {
      rmSync(full, { recursive: true, force: true });
      console.log(`  borrado ${target}`);
    }
  }
  process.exit(0);
}

const problems = [];

for (const [source, target] of COPIES) {
  const from = join(ROOT, source);
  const to = join(ROOT, target);

  if (!existsSync(from)) {
    problems.push(`falta ${source}`);
    continue;
  }

  rmSync(to, { recursive: true, force: true });
  mkdirSync(to, { recursive: true });

  for (const entry of readdirSync(from)) {
    if (SKIP.has(entry)) continue;
    cpSync(join(from, entry), join(to, entry), { recursive: true });
  }

  // El enlace a node_modules es lo que hace que esto tarde un segundo y no
  // cinco minutos. Si falla, la copia sirve igual: solo hay que instalar.
  const deps = join(from, 'node_modules');
  if (existsSync(deps)) {
    try {
      symlinkSync(deps, join(to, 'node_modules'), 'junction');
    } catch (err) {
      problems.push(`${target}: no se pudo enlazar node_modules (${err.code}) — corré npm install ahí`);
    }
  } else {
    problems.push(`${target}: ${source}/node_modules no existe — corré npm install en ${source} primero`);
  }

  console.log(`  ${source} → ${target}`);
}

if (problems.length) {
  console.log('');
  for (const p of problems) console.log(`  · ${p}`);
}

console.log(`
  Listo. El live coding se da en demo/, no en solution/.
  Para volver a empezar de cero: node scripts/prep-demo.mjs
  Para borrarlas al terminar:    node scripts/prep-demo.mjs --clean
`);
