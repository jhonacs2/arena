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

/**
 * Cada destino dice dónde declarar los tokens y si quiere los dos temas.
 *
 * `selector` existe por Marp: sus diapositivas son elementos `<section>` y hay
 * que declarar las variables ahí. Ojo — tiene que ser `section` a secas: con
 * `:root, section`, el reescribiwlector de Marp genera una lista de selectores
 * que el navegador descarta entera, y el deck sale sin un solo color.
 *
 * `modes: 'light'` también es por Marp: si el tema dependiera de
 * `prefers-color-scheme`, el mismo archivo exportaría distinto según la
 * máquina desde la que se genere. Unas diapositivas tienen que verse igual
 * siempre — y proyectadas, claras.
 */
const TARGETS = [
  { file: 'theme/marp-neobrutal.css', selector: 'section', modes: 'light' },
  { file: 'project/frontend/solution/src/styles.css' },
  { file: 'project/frontend/starter/src/styles.css' },
  { file: 'lab/solution/src/styles.css' },
  { file: 'lab/starter/src/styles.css' },
];

const START = '/* @tokens:start */';
const END = '/* @tokens:end */';

const tokens = JSON.parse(readFileSync(join(ROOT, 'docs/design/tokens.json'), 'utf8'));

const oklch = ([l, c, h]) => `oklch(${l} ${c} ${h})`;
const pad = (s, n) => s.padEnd(n);

function block({ selector = ':root', modes = 'both' } = {}) {
  const lines = [];
  const emit = (s = '') => lines.push(s);

  emit('  /* Generado desde docs/design/tokens.json — no editar a mano.');
  emit('     Cambiar un color: se edita el JSON y se corre node scripts/gen-tokens-css.mjs */');
  emit('');
  emit(`  ${selector} {`);

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

  const semanticos = (map, label) => {
    emit(`    /* semánticos — ${label} */`);
    for (const [semantic, primitive] of Object.entries(map)) {
      emit(`    --${pad(semantic + ':', 15)} var(--${primitive});`);
    }
  };

  if (modes === 'light') {
    // Un solo tema, sin media query: las diapositivas tienen que exportar
    // igual desde cualquier máquina.
    emit('');
    emit(`  ${selector} {`);
    semanticos(tokens.semantic.light, 'claro. Las diapositivas son siempre claras: se proyectan');
    emit('  }');
    return lines.join('\n');
  }

  emit('');
  emit(`  ${selector}, [data-theme="light"] {`);
  semanticos(tokens.semantic.light, 'claro');
  emit('  }');

  emit('');
  emit(`  @media (prefers-color-scheme: dark) { ${selector}:not([data-theme="light"]) {`);
  semanticos(tokens.semantic.dark, 'oscuro. El borde se invierte a tiza: es un modo diseñado, no un filtro');
  emit('  } }');

  // El selector explícito tiene que poder ganarle a la media query en ambos sentidos.
  emit('');
  emit('  [data-theme="dark"] {');
  semanticos(tokens.semantic.dark, 'oscuro forzado');
  emit('  }');

  return lines.join('\n');
}

const stale = [];
let written = 0;

for (const target of TARGETS) {
  const path = join(ROOT, target.file);
  if (!existsSync(path)) continue;

  const text = readFileSync(path, 'utf8');
  const from = text.indexOf(START);
  const to = text.indexOf(END);

  if (from === -1 || to === -1) {
    stale.push(`${target.file} — no tiene los marcadores ${START} … ${END}`);
    continue;
  }

  const generated = block(target);
  const next = text.slice(0, from + START.length) + '\n' + generated + '\n' + text.slice(to);
  if (next === text) continue;

  if (CHECK) stale.push(`${target.file} — desfasado respecto de tokens.json`);
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
