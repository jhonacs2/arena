#!/usr/bin/env node
/**
 * Escribe el bloque de custom properties de docs/design/tokens.json dentro de
 * los archivos CSS que lo declaren. Una sola fuente de verdad para la paleta.
 *
 *   node scripts/gen-tokens-css.mjs           inyecta en todos los destinos
 *   node scripts/gen-tokens-css.mjs --check   falla si algún destino está desfasado
 *
 * El archivo destino marca la zona a reemplazar así:
 *
 *   ／* @tokens:start *／
 *   ...lo que haya acá se reemplaza...
 *   ／* @tokens:end *／
 *
 * Todo lo que esté fuera de los marcadores no se toca. Así el CSS se escribe a
 * mano y solo la paleta se genera.
 */

import { readFileSync, writeFileSync, existsSync } from 'node:fs';
import { dirname, join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

const ROOT = join(dirname(fileURLToPath(import.meta.url)), '..');
const CHECK = process.argv.includes('--check');

const TARGETS = [
  'theme/marp-neobrutal.css',
  'project/frontend/solution/src/styles.css',
  'project/frontend/starter/src/styles.css',
  'lab/solution/src/styles.css',
  'lab/starter/src/styles.css',
];

const START = '/* @tokens:start */';
const END = '/* @tokens:end */';

const tokens = JSON.parse(readFileSync(join(ROOT, 'docs/design/tokens.json'), 'utf8'));

const oklch = ([l, c, h]) => `oklch(${l} ${c} ${h})`;
const pad = (s, n) => s.padEnd(n);

function block() {
  const lines = [];
  const emit = (s = '') => lines.push(s);

  emit('  /* Generado desde docs/design/tokens.json — no editar a mano.');
  emit('     Cambiar un color: se edita el JSON y se corre node scripts/gen-tokens-css.mjs */');
  emit('');
  emit('  :root {');

  emit('    /* primitivas */');
  for (const [name, def] of Object.entries(tokens.primitives)) {
    emit(`    --${pad(name + ':', 13)} ${oklch(def.oklch)};`);
  }

  emit('');
  emit('    /* sedas de jockey — los 10 colores registrados del deporte.');
  emit('       Solo dentro del SVG de la seda. Ningún texto se pinta encima. */');
  for (const [name, def] of Object.entries(tokens.silks)) {
    emit(`    --${pad('silk-' + name + ':', 13)} ${oklch(def.oklch)};`);
  }

  emit('');
  emit('    color-scheme: light dark;');
  emit('  }');

  for (const [theme, map] of Object.entries(tokens.semantic)) {
    emit('');
    const selector =
      theme === 'light'
        ? '  :root, [data-theme="light"] {'
        : '  @media (prefers-color-scheme: dark) { :root:not([data-theme="light"]) {';
    emit(selector);
    emit(`    /* semánticos — ${theme === 'light' ? 'claro' : 'oscuro'}. El oscuro invierte el borde a tiza: es un modo diseñado, no un filtro. */`);
    for (const [semantic, primitive] of Object.entries(map)) {
      emit(`    --${pad(semantic + ':', 15)} var(--${primitive});`);
    }
    emit(theme === 'light' ? '  }' : '  } }');
  }

  // El selector explícito tiene que poder ganarle a la media query en ambos sentidos.
  emit('');
  emit('  [data-theme="dark"] {');
  for (const [semantic, primitive] of Object.entries(tokens.semantic.dark)) {
    emit(`    --${pad(semantic + ':', 15)} var(--${primitive});`);
  }
  emit('  }');

  return lines.join('\n');
}

const generated = block();
const stale = [];
let written = 0;

for (const target of TARGETS) {
  const path = join(ROOT, target);
  if (!existsSync(path)) continue;

  const text = readFileSync(path, 'utf8');
  const from = text.indexOf(START);
  const to = text.indexOf(END);

  if (from === -1 || to === -1) {
    stale.push(`${target} — no tiene los marcadores ${START} … ${END}`);
    continue;
  }

  const next = text.slice(0, from + START.length) + '\n' + generated + '\n' + text.slice(to);
  if (next === text) continue;

  if (CHECK) stale.push(`${target} — desfasado respecto de tokens.json`);
  else {
    writeFileSync(path, next, 'utf8');
    written++;
    console.log(`  actualizado  ${relative(ROOT, path)}`);
  }
}

if (stale.length) {
  console.error('\n  ✗ Tokens CSS:');
  for (const s of stale) console.error(`    · ${s}`);
  console.error(CHECK ? '\n    Correr: node scripts/gen-tokens-css.mjs\n' : '');
  process.exit(1);
}

if (!CHECK) console.log(`  ${written} archivo(s) al día.`);
