#!/usr/bin/env node
/**
 * `sesiones/**\/diagramas/*.excalidraw` → el `.svg` de al lado.
 *
 *   node scripts/gen-diagram-svg.mjs           regenera
 *   node scripts/gen-diagram-svg.mjs --check    falla si alguno quedó desfasado
 *
 * ¿Por qué Excalidraw y no SVG a mano? Porque el diagrama se usa **en vivo**. El
 * bloque de las 0:12 se da con el editor cerrado, dibujando: poder abrir el
 * `.excalidraw`, agregarle una flecha mientras alguien pregunta y guardarlo vale
 * más que un SVG prolijo que nadie puede tocar en clase.
 *
 * El `.excalidraw` es la fuente y se edita en excalidraw.com («Abrir» → el
 * archivo → «Guardar en...» encima). El `.svg` es derivado: lo consumen las
 * diapositivas de Marp, que no saben leer Excalidraw.
 *
 * Los colores se guardan en hex porque es lo único que Excalidraw entiende, y
 * acá se reescriben a `var(--dg-token, #hex)` con los dos temas embebidos. Así
 * el diagrama sigue el modo oscuro sin dejar de abrirse en la app.
 */

import { readFileSync, writeFileSync, existsSync, readdirSync, statSync } from 'node:fs';
import { dirname, join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';
import { render } from './lib/excalidraw.mjs';
import { oklchToHex } from './lib/oklch.mjs';

const ROOT = join(dirname(fileURLToPath(import.meta.url)), '..');
const CHECK = process.argv.includes('--check');
const tokens = JSON.parse(readFileSync(join(ROOT, 'docs/design/tokens.json'), 'utf8'));

// ── Paleta ─────────────────────────────────────────────────────────────────
// Un `#hex` puede corresponder a más de un token semántico (en claro `border` y
// `shadow` son los dos ink-900). Gana el primero, que es el que da el nombre más
// útil; lo que importa es que el valor sea el correcto en los dos temas.
const hexOf = (primitive) => {
  const def = tokens.primitives[primitive] ?? tokens.silks[primitive];
  if (!def) throw new Error(`el token "${primitive}" no está en tokens.json`);
  return oklchToHex(def.oklch);
};

const palette = new Map();
const themeVars = { light: [], dark: [] };
const claimants = new Map(); // hex en claro → [nombres semánticos que lo usan]

for (const [theme, map] of Object.entries(tokens.semantic)) {
  for (const [name, primitive] of Object.entries(map)) {
    const hex = hexOf(primitive);
    themeVars[theme].push(`--dg-${name}: ${hex};`);
    if (theme !== 'light') continue;
    if (!palette.has(hex.toLowerCase())) palette.set(hex.toLowerCase(), `dg-${name}`);
    claimants.set(hex.toLowerCase(), [...(claimants.get(hex.toLowerCase()) ?? []), name]);
  }
}

/**
 * Hexes que en claro son el mismo color pero en oscuro se separan. Ahí adivinar
 * el token por el hex elige mal en silencio: `text` y `shadow` son los dos
 * ink-900 en claro, y en oscuro uno es tiza y el otro casi negro. Un elemento
 * que use uno de estos colores tiene que declarar el token en `customData`.
 */
const ambiguous = new Set();
for (const [hex, names] of claimants) {
  const darks = new Set(names.map((n) => hexOf(tokens.semantic.dark[n])));
  if (darks.size > 1) ambiguous.add(hex);
}

const darkCss = [
  `    svg { ${themeVars.light.join(' ')} }`,
  `    @media (prefers-color-scheme: dark) {`,
  `      svg { ${themeVars.dark.join(' ')} }`,
  `    }`,
].join('\n');

// ── Recorrido ──────────────────────────────────────────────────────────────
function* walk(dir) {
  if (!existsSync(dir)) return;
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) yield* walk(full);
    else if (entry.endsWith('.excalidraw')) yield full;
  }
}

const problems = [];
let written = 0;
let checked = 0;

for (const source of walk(join(ROOT, 'sesiones'))) {
  const rel = relative(ROOT, source).replace(/\\/g, '/');
  const target = source.replace(/\.excalidraw$/, '.svg');

  let scene;
  try {
    scene = JSON.parse(readFileSync(source, 'utf8'));
  } catch (err) {
    problems.push(`${rel} — no es JSON válido: ${err.message}`);
    continue;
  }

  // El trazo ambiguo se resuelve por el rol del elemento (ver `render`). El
  // **relleno** no: una forma pintada de ink-900 puede ser la sombra dura del
  // neobrutalismo o un bloque de tinta sólido, y en oscuro esos dos van a
  // extremos opuestos de la escala. Ese es el único caso que hay que declarar.
  for (const el of scene.elements ?? []) {
    const hex = el.isDeleted ? null : el.backgroundColor?.toLowerCase();
    if (hex && ambiguous.has(hex) && !el.customData?.backgroundToken) {
      problems.push(
        `${rel} — ${el.type} ${el.id}: relleno ${hex} es ambiguo en oscuro ` +
          `(${claimants.get(hex).join(', ')}); declará customData.backgroundToken`,
      );
    }
  }

  const { svg, problems: bad } = render(scene, { palette, darkCss, ambiguous });
  for (const p of bad) problems.push(`${rel} — ${p}`);
  if (!svg) continue;

  checked++;
  const before = existsSync(target) ? readFileSync(target, 'utf8') : null;

  if (CHECK) {
    if (before !== svg) {
      problems.push(
        `${rel.replace(/\.excalidraw$/, '.svg')} está desfasado — corré node scripts/gen-diagram-svg.mjs`,
      );
    }
  } else if (before !== svg) {
    writeFileSync(target, svg, 'utf8');
    written++;
    console.log(`  ${rel.replace(/\.excalidraw$/, '.svg')}`);
  }
}

if (problems.length) {
  for (const p of problems) console.error(`  ✗ ${p}`);
  process.exit(1);
}

if (!CHECK) {
  console.log(`${checked} diagrama(s) · ${written} svg actualizado(s)`);
}
