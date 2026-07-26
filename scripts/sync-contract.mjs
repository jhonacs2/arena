#!/usr/bin/env node
/**
 * Copia docs/contract/ hacia los proyectos que lo consumen.
 *
 *   node scripts/sync-contract.mjs           copia
 *   node scripts/sync-contract.mjs --check   falla si algún destino está desfasado
 *
 * Hace falta porque `project/` se publica al alumno **sin** `docs/`: el backend
 * y el frontend tienen que llevar su propia copia del seed. La copia es
 * mecánica y `verify.mjs` corre el `--check`, así que no puede quedar vieja.
 *
 * Mismo mecanismo que gen-tokens-css.mjs: una fuente, copias verificadas.
 */

import { readFileSync, writeFileSync, existsSync, mkdirSync, readdirSync, rmSync, statSync } from 'node:fs';
import { dirname, join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

const ROOT = join(dirname(fileURLToPath(import.meta.url)), '..');
const CHECK = process.argv.includes('--check');
const SOURCE = join(ROOT, 'docs/contract');

/**
 * Cada destino declara qué subcarpetas necesita.
 *   seed      — el dataset, lo cargan backend y mock
 *   fixtures  — la grabación de la carrera
 *   samples   — golden tests de ambos lados
 * Solo se sincroniza lo que ya existe: en Fase 1 el frontend todavía no está.
 */
const TARGETS = [
  { dir: 'project/backend/internal/seed/data', parts: ['seed', 'fixtures', 'samples'] },
  { dir: 'project/frontend/solution/public/contract', parts: ['seed', 'fixtures'] },
  { dir: 'project/frontend/starter/public/contract', parts: ['seed', 'fixtures'] },
];

const HEADER = 'GENERADO — no editar acá. La fuente es docs/contract/. Correr: node scripts/sync-contract.mjs';

function filesIn(dir, prefix = '') {
  const out = [];
  for (const entry of readdirSync(dir).sort()) {
    const full = join(dir, entry);
    const rel = prefix ? `${prefix}/${entry}` : entry;
    if (statSync(full).isDirectory()) out.push(...filesIn(full, rel));
    else if (/\.(json|jsonl)$/.test(entry)) out.push({ rel, full });
  }
  return out;
}

const stale = [];
let copied = 0;
let skipped = 0;

for (const target of TARGETS) {
  const destRoot = join(ROOT, target.dir);
  // El destino se crea solo si su proyecto ya existe: en Fase 1 el frontend
  // todavía no estaba. La raíz del proyecto es lo que hay antes de `internal/`
  // o de `public/`.
  const projectRoot = join(ROOT, target.dir.replace(/\/(internal|public)\/.*$/, ''));
  if (!existsSync(projectRoot)) { skipped++; continue; }

  const wanted = new Map();
  for (const part of target.parts) {
    const from = join(SOURCE, part);
    if (!existsSync(from)) continue;
    for (const f of filesIn(from, part)) wanted.set(f.rel, f.full);
  }

  for (const [rel, from] of wanted) {
    const to = join(destRoot, rel);
    const source = readFileSync(from, 'utf8');
    const current = existsSync(to) ? readFileSync(to, 'utf8') : null;

    if (current === source) continue;
    if (CHECK) { stale.push(`${target.dir}/${rel}`); continue; }

    mkdirSync(dirname(to), { recursive: true });
    writeFileSync(to, source, 'utf8');
    copied++;
  }

  // Archivos que sobran en el destino: quedaron de una versión anterior del
  // contrato y confunden más que ayudar.
  if (existsSync(destRoot)) {
    for (const f of filesIn(destRoot)) {
      if (wanted.has(f.rel)) continue;
      if (CHECK) stale.push(`${target.dir}/${f.rel} — sobra, ya no está en docs/contract/`);
      else { rmSync(f.full); console.log(`  eliminado   ${relative(ROOT, f.full)}`); }
    }
  }

  if (!CHECK && !existsSync(join(destRoot, 'LEEME.txt'))) {
    mkdirSync(destRoot, { recursive: true });
    writeFileSync(join(destRoot, 'LEEME.txt'), HEADER + '\n', 'utf8');
  }
}

if (stale.length) {
  console.error('\n  ✗ Copias del contrato desfasadas:');
  for (const s of stale) console.error(`    · ${s}`);
  console.error('\n    Correr: node scripts/sync-contract.mjs\n');
  process.exit(1);
}

if (!CHECK) {
  console.log(`  ${copied} archivo(s) sincronizados${skipped ? `, ${skipped} destino(s) todavía no existen` : ''}.`);
}
