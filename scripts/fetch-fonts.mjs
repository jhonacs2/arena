#!/usr/bin/env node
/**
 * Descarga las tres tipografías del sistema y las deja auto-hospedadas.
 *
 *   node scripts/fetch-fonts.mjs
 *
 * `docs/design/tokens.md` §3: nada de CDN de fuentes. Un aula tiene wifi de
 * aula, y una clase de dos horas no se puede caer porque Google Fonts tarde.
 * Los .woff2 se commitean: son parte del repo, como cualquier otro asset.
 *
 * Las tres son variables y OFL. Se piden con los ejes completos para poder
 * usar `font-variation-settings` — el eje de ancho de Bricolage es lo que deja
 * condensar los títulos largos en móvil sin cambiar de familia.
 */

import { writeFileSync, mkdirSync, existsSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const ROOT = join(dirname(fileURLToPath(import.meta.url)), '..');

const FAMILIES = [
  {
    slug: 'bricolage-grotesque',
    query: 'Bricolage+Grotesque:opsz,wdth,wght@12..96,75..100,200..800',
    family: 'Bricolage Grotesque',
    role: 'display',
  },
  {
    slug: 'public-sans',
    query: 'Public+Sans:ital,wght@0,100..900;1,100..900',
    family: 'Public Sans',
    role: 'cuerpo',
  },
  {
    slug: 'martian-mono',
    query: 'Martian+Mono:wdth,wght@75..112.5,100..800',
    family: 'Martian Mono',
    role: 'números',
  },
];

// Solo los subconjuntos que necesita el español. `latin` cubre las vocales con
// tilde, la eñe y la diéresis; `latin-ext` entra por si algún nombre lo pide.
const WANTED_SUBSETS = new Set(['latin', 'latin-ext']);

// Sin este User-Agent, Google devuelve TTF en vez de WOFF2.
const MODERN_UA =
  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36';

const TARGETS = [
  'project/frontend/solution/public/fonts',
  'project/frontend/starter/public/fonts',
  'lab/solution/public/fonts',
  'lab/starter/public/fonts',
];

/** Extrae los descriptores de un bloque @font-face, para replicarlos. */
function blocksOf(css) {
  const out = [];
  const re = /\/\*\s*([\w[\]-]+)\s*\*\/\s*@font-face\s*\{([^}]+)\}/g;
  let match;
  while ((match = re.exec(css)) !== null) {
    const [, subset, body] = match;
    const get = (name) => body.match(new RegExp(`${name}:\\s*([^;]+);`))?.[1]?.trim();
    const url = body.match(/url\((https:[^)]+\.woff2)\)/)?.[1];
    if (!url) continue;
    out.push({
      subset,
      url,
      style: get('font-style') ?? 'normal',
      weight: get('font-weight') ?? '400',
      stretch: get('font-stretch'),
      unicodeRange: get('unicode-range'),
    });
  }
  return out;
}

const faceRules = [];
let downloaded = 0;

for (const family of FAMILIES) {
  const cssUrl = `https://fonts.googleapis.com/css2?family=${family.query}&display=swap`;
  const res = await fetch(cssUrl, { headers: { 'User-Agent': MODERN_UA } });
  if (!res.ok) {
    console.error(`  ✗ ${family.family}: Google devolvió ${res.status}`);
    process.exit(1);
  }
  const css = await res.text();
  const blocks = blocksOf(css).filter((b) => WANTED_SUBSETS.has(b.subset));

  if (blocks.length === 0) {
    console.error(`  ✗ ${family.family}: no se encontró ningún subconjunto latino`);
    process.exit(1);
  }

  for (const block of blocks) {
    const italic = block.style === 'italic';
    const name = `${family.slug}-${block.subset}${italic ? '-italic' : ''}.woff2`;

    const fontRes = await fetch(block.url, { headers: { 'User-Agent': MODERN_UA } });
    if (!fontRes.ok) {
      console.error(`  ✗ ${name}: ${fontRes.status}`);
      process.exit(1);
    }
    const bytes = Buffer.from(await fontRes.arrayBuffer());

    for (const target of TARGETS) {
      const dir = join(ROOT, target);
      // Solo se escribe donde el proyecto ya existe.
      if (!existsSync(join(ROOT, target.split('/public/')[0]))) continue;
      mkdirSync(dir, { recursive: true });
      writeFileSync(join(dir, name), bytes);
    }
    downloaded++;
    console.log(`  ${name.padEnd(42)} ${(bytes.length / 1024).toFixed(0).padStart(4)} KB`);

    faceRules.push(
      [
        '@font-face {',
        `  font-family: '${family.family}';`,
        `  src: url('/fonts/${name}') format('woff2');`,
        `  font-style: ${block.style};`,
        `  font-weight: ${block.weight};`,
        block.stretch ? `  font-stretch: ${block.stretch};` : null,
        // swap: el texto se ve con la fuente de reserva mientras carga. En una
        // clase, media pantalla en blanco esperando una fuente es peor que un
        // salto tipográfico.
        '  font-display: swap;',
        block.unicodeRange ? `  unicode-range: ${block.unicodeRange};` : null,
        '}',
      ]
        .filter(Boolean)
        .join('\n'),
    );
  }
}

const header = `/* Generado por scripts/fetch-fonts.mjs — no editar a mano.
 *
 * Las tres familias son OFL y están auto-hospedadas: docs/design/tokens.md §3.
 * Nada de CDN — un aula tiene wifi de aula.
 *
${FAMILIES.map((f) => ` *   ${f.family.padEnd(22)} ${f.role}`).join('\n')}
 */

`;

for (const target of TARGETS) {
  const dir = join(ROOT, target);
  if (!existsSync(dir)) continue;
  writeFileSync(join(dirname(dir), '..', 'src', 'fonts.css'), header + faceRules.join('\n\n') + '\n', 'utf8');
}

console.log(`\n  ${downloaded} archivos · ${faceRules.length} reglas @font-face · fonts.css escrito`);
console.log('  Licencias: las tres son SIL Open Font License 1.1');
