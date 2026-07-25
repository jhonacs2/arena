/**
 * Genera docs/design/assets/silks-specimen.svg — la hoja de muestra de las
 * sedas de jockey: los 10 patrones de la gramática y las 54 sedas reales del
 * seed, cada una derivada del `id` de su caballo.
 *
 *   node scripts/gen-silks-specimen.mjs
 *
 * Esta es la implementación de referencia de `silkFromId()`. En Fase 2 se
 * porta tal cual a `shared/ui/silk/silk.util.ts` — misma función hash, mismas
 * reglas de rechazo, mismo orden de ejes. Si la seda de un caballo cambia
 * entre esta hoja y la app, algo se portó mal.
 */

import { readFileSync, writeFileSync, mkdirSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const ROOT = join(dirname(fileURLToPath(import.meta.url)), '..');
const OUT = join(ROOT, 'docs/design/assets/silks-specimen.svg');

const tokens = JSON.parse(readFileSync(join(ROOT, 'docs/design/tokens.json'), 'utf8'));
const races = JSON.parse(readFileSync(join(ROOT, 'docs/contract/seed/races.json'), 'utf8'));

// ── Gramática ─────────────────────────────────────────────────────────────
const BODY_PATTERNS = [
  'solid', 'halves', 'quarters', 'stripes', 'hoops',
  'chevron', 'sash', 'star', 'diamond', 'seams',
];
const SLEEVE_PATTERNS = ['plain', 'alt', 'hooped', 'striped'];

const COLOR_NAMES = Object.keys(tokens.silks);
const LIGHTNESS = Object.fromEntries(
  Object.entries(tokens.silks).map(([n, d]) => [n, d.oklch[0]]),
);

const MIN_LIGHTNESS_GAP = 0.22;

/** FNV-1a 32 bits. Determinístico y trivial de reimplementar en TypeScript. */
function hash(id) {
  let h = 0x811c9dc5;
  for (let i = 0; i < id.length; i++) {
    h ^= id.charCodeAt(i);
    h = Math.imul(h, 0x01000193) >>> 0;
  }
  return h;
}

/** id → especificación de seda. Función pura. */
export function silkFromId(id) {
  let h = hash(id);
  const next = () => (h = Math.imul(h ^ (h >>> 15), 0x2545f491) >>> 0);

  const primary = COLOR_NAMES[h % COLOR_NAMES.length];

  // Rechazo: el secundario tiene que ser otro color Y separarse en luminancia.
  // Sin esto salen sedas azul-sobre-violeta que a 24 px son un cuadrado sólido.
  let secondary = primary;
  for (let attempt = 0; attempt < 32; attempt++) {
    next();
    const candidate = COLOR_NAMES[h % COLOR_NAMES.length];
    if (
      candidate !== primary &&
      Math.abs(LIGHTNESS[candidate] - LIGHTNESS[primary]) >= MIN_LIGHTNESS_GAP
    ) {
      secondary = candidate;
      break;
    }
  }

  next();
  const body = BODY_PATTERNS[h % BODY_PATTERNS.length];
  next();
  const sleeves = SLEEVE_PATTERNS[h % SLEEVE_PATTERNS.length];

  return { primary, secondary, body, sleeves };
}

// ── Render ────────────────────────────────────────────────────────────────
// viewBox 0 0 40 38. Cuerpo 12..28 (16 ancho), mangas 3..12 y 28..37.
const BODY = { x: 12, y: 4, w: 16, h: 30 };

const cssVar = (name) => `var(--silk-${name})`;

function bodyShapes(p, s) {
  const { x, y, w, h } = BODY;
  const half = w / 2;
  const bg = `<rect x="${x}" y="${y}" width="${w}" height="${h}" fill="${cssVar(p)}"/>`;

  switch (arguments[2]) {
    case 'solid':
      return bg;

    case 'halves':
      return bg + `<rect x="${x + half}" y="${y}" width="${half}" height="${h}" fill="${cssVar(s)}"/>`;

    case 'quarters':
      return (
        bg +
        `<rect x="${x + half}" y="${y}" width="${half}" height="${h / 2}" fill="${cssVar(s)}"/>` +
        `<rect x="${x}" y="${y + h / 2}" width="${half}" height="${h / 2}" fill="${cssVar(s)}"/>`
      );

    case 'stripes':
      return (
        bg +
        [1, 3].map((i) => `<rect x="${x + i * (w / 4)}" y="${y}" width="${w / 4}" height="${h}" fill="${cssVar(s)}"/>`).join('')
      );

    case 'hoops':
      return (
        bg +
        [1, 3].map((i) => `<rect x="${x}" y="${y + i * (h / 5)}" width="${w}" height="${h / 5}" fill="${cssVar(s)}"/>`).join('')
      );

    case 'chevron':
      return (
        bg +
        [0, 1, 2].map((i) => {
          const top = y + 3 + i * 9;
          return `<path d="M${x},${top} L${x + half},${top + 5} L${x + w},${top} L${x + w},${top + 4} L${x + half},${top + 9} L${x},${top + 4} Z" fill="${cssVar(s)}"/>`;
        }).join('')
      );

    case 'sash':
      return bg + `<path d="M${x},${y + h - 6} L${x + w - 6},${y} L${x + w},${y} L${x + w},${y + 4} L${x + 4},${y + h} L${x},${y + h} Z" fill="${cssVar(s)}"/>`;

    case 'star': {
      const cx = x + half;
      const cy = y + h / 2;
      const pts = Array.from({ length: 10 }, (_, i) => {
        const r = i % 2 === 0 ? 6.5 : 2.7;
        const a = (Math.PI / 5) * i - Math.PI / 2;
        return `${(cx + r * Math.cos(a)).toFixed(2)},${(cy + r * Math.sin(a)).toFixed(2)}`;
      }).join(' ');
      return bg + `<polygon points="${pts}" fill="${cssVar(s)}"/>`;
    }

    case 'diamond': {
      const cx = x + half;
      const cy = y + h / 2;
      return bg + `<polygon points="${cx},${cy - 8} ${cx + 6},${cy} ${cx},${cy + 8} ${cx - 6},${cy}" fill="${cssVar(s)}"/>`;
    }

    case 'seams':
      return bg + `<rect x="${x + half - 2}" y="${y}" width="4" height="${h}" fill="${cssVar(s)}"/>`;

    default:
      return bg;
  }
}

function sleeveShapes(p, s, pattern) {
  const sides = [
    { x: 3, y: 6 },
    { x: 28, y: 6 },
  ];
  const w = 9;
  const h = 14;

  return sides
    .map(({ x, y }) => {
      switch (pattern) {
        case 'plain':
          return `<rect x="${x}" y="${y}" width="${w}" height="${h}" fill="${cssVar(p)}"/>`;
        case 'alt':
          return `<rect x="${x}" y="${y}" width="${w}" height="${h}" fill="${cssVar(s)}"/>`;
        case 'hooped':
          return (
            `<rect x="${x}" y="${y}" width="${w}" height="${h}" fill="${cssVar(p)}"/>` +
            [1, 3].map((i) => `<rect x="${x}" y="${y + i * (h / 4)}" width="${w}" height="${h / 4}" fill="${cssVar(s)}"/>`).join('')
          );
        case 'striped':
          return (
            `<rect x="${x}" y="${y}" width="${w}" height="${h}" fill="${cssVar(p)}"/>` +
            `<rect x="${x + w / 3}" y="${y}" width="${w / 3}" height="${h}" fill="${cssVar(s)}"/>`
          );
        default:
          return '';
      }
    })
    .join('');
}

/** Devuelve el <g> de una seda, sin viewBox: se compone dentro de la hoja. */
function silkGroup(spec) {
  const { primary: p, secondary: s, body, sleeves } = spec;
  return (
    `<g stroke="var(--ink)" stroke-width="1.5" stroke-linejoin="round">` +
    sleeveShapes(p, s, sleeves) +
    `<g>${bodyShapes(p, s, body)}</g>` +
    `<rect x="${BODY.x}" y="${BODY.y}" width="${BODY.w}" height="${BODY.h}" fill="none"/>` +
    `<rect x="3" y="6" width="9" height="14" fill="none"/>` +
    `<rect x="28" y="6" width="9" height="14" fill="none"/>` +
    `<path d="M17,4 L23,4 L21.5,7.5 L18.5,7.5 Z" fill="var(--ink)" stroke="none"/>` +
    `</g>`
  );
}

// ── Hoja de muestra ───────────────────────────────────────────────────────
const horses = races.flatMap((r) => r.horses.map((h) => ({ ...h, race: r.name })));

const CELL_W = 92;
const CELL_H = 108;
const COLS = 9;
const gridRows = Math.ceil(horses.length / COLS);

const grammarY = 96;
const gridY = grammarY + 150;
const HEIGHT = gridY + gridRows * CELL_H + 48;
const WIDTH = COLS * CELL_W + 48;

const swatches = COLOR_NAMES.map(
  (n, i) =>
    `<rect x="${24 + i * 46}" y="52" width="38" height="18" fill="${cssVar(n)}" stroke="var(--ink)" stroke-width="2"/>` +
    `<text x="${43 + i * 46}" y="82" class="tiny" text-anchor="middle">${n}</text>`,
).join('');

const grammar = BODY_PATTERNS.map((pattern, i) => {
  const spec = { primary: 'red', secondary: 'white', body: pattern, sleeves: 'hooped' };
  const x = 24 + i * 46;
  return (
    `<g transform="translate(${x},${grammarY + 22}) scale(1.05)">${silkGroup(spec)}</g>` +
    `<text x="${x + 21}" y="${grammarY + 78}" class="tiny" text-anchor="middle">${pattern}</text>`
  );
}).join('');

const grid = horses
  .map((h, i) => {
    const col = i % COLS;
    const row = Math.floor(i / COLS);
    const x = 24 + col * CELL_W;
    const y = gridY + row * CELL_H;
    const spec = silkFromId(h.id);
    const label = h.name.length > 11 ? h.name.slice(0, 10) + '…' : h.name;
    return (
      `<g transform="translate(${x + 14},${y}) scale(1.5)">${silkGroup(spec)}</g>` +
      // El número NUNCA va sobre la seda: cuadrado aparte, tinta sobre tiza.
      `<rect x="${x + 2}" y="${y + 58}" width="18" height="18" fill="var(--chalk)" stroke="var(--ink)" stroke-width="2"/>` +
      `<text x="${x + 11}" y="${y + 71}" class="num" text-anchor="middle">${h.number}</text>` +
      `<text x="${x + 26}" y="${y + 71}" class="name">${label}</text>` +
      `<text x="${x + 2}" y="${y + 87}" class="tiny">${spec.primary}/${spec.secondary}</text>` +
      `<text x="${x + 2}" y="${y + 98}" class="tiny">${spec.body} · ${spec.sleeves}</text>`
    );
  })
  .join('');

const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${WIDTH} ${HEIGHT}" width="${WIDTH}" height="${HEIGHT}" font-family="ui-monospace, 'Cascadia Mono', Menlo, monospace">
  <style>
    :root, svg {
      --ink:   oklch(0.17 0.02 155);
      --chalk: oklch(0.97 0.008 110);
${COLOR_NAMES.map((n) => `      --silk-${n}: oklch(${tokens.silks[n].oklch.join(' ')});`).join('\n')}
    }
    @media (prefers-color-scheme: dark) {
      :root, svg { --ink: oklch(0.97 0.008 110); --chalk: oklch(0.17 0.02 155); }
    }
    .h1   { font-size: 19px; font-weight: 700; fill: var(--ink); }
    .h2   { font-size: 11px; font-weight: 700; fill: var(--ink); letter-spacing: .09em; }
    .name { font-size: 10px; fill: var(--ink); }
    .num  { font-size: 11px; font-weight: 700; fill: var(--ink); }
    .tiny { font-size: 8px; fill: var(--ink); opacity: .62; }
  </style>

  <rect width="100%" height="100%" fill="var(--chalk)"/>

  <text x="24" y="32" class="h1">Sedas de jockey — hoja de muestra</text>
  <text x="24" y="46" class="h2">LOS 10 COLORES REGISTRADOS</text>
  ${swatches}

  <text x="24" y="${grammarY + 8}" class="h2">LOS 10 PATRONES DE CUERPO</text>
  ${grammar}

  <text x="24" y="${gridY - 18}" class="h2">LAS 54 SEDAS DEL SEED — DERIVADAS DEL id, NO ELEGIDAS</text>
  ${grid}
</svg>
`;

mkdirSync(dirname(OUT), { recursive: true });
writeFileSync(OUT, svg, 'utf8');

// ── Chequeos de la gramática ──────────────────────────────────────────────
const key = (s) => `${s.primary}|${s.secondary}|${s.body}|${s.sleeves}`;
const specs = horses.map((h) => ({ id: h.id, ...silkFromId(h.id) }));

// Requisito duro: dos caballos de LA MISMA carrera no pueden compartir seda.
// Es el único caso donde se ven lado a lado y hay que distinguirlos de un vistazo.
const clashes = [];
for (const race of races) {
  const seen = new Map();
  for (const horse of race.horses) {
    const k = key(silkFromId(horse.id));
    if (seen.has(k)) clashes.push(`${race.name}: ${seen.get(k)} y ${horse.name} comparten seda`);
    else seen.set(k, horse.name);
  }
}

// Entre carreras distintas, compartir seda es aceptable — en el hipódromo real
// las sedas son del dueño, no del caballo, y se repiten entre jornadas. Se
// informa, no se falla: en `my-bets` cada fila lleva además el nombre en texto.
const global = new Map();
const crossRace = [];
for (const race of races) {
  for (const horse of race.horses) {
    const k = key(silkFromId(horse.id));
    if (global.has(k)) crossRace.push(`${global.get(k)} ≡ ${race.name}/${horse.name}`);
    else global.set(k, `${race.name}/${horse.name}`);
  }
}

const badGap = specs.filter(
  (s) => Math.abs(LIGHTNESS[s.secondary] - LIGHTNESS[s.primary]) < MIN_LIGHTNESS_GAP,
);

const spread = {};
for (const s of specs) spread[s.body] = (spread[s.body] ?? 0) + 1;

console.log(`silks-specimen.svg · ${horses.length} sedas · ${WIDTH}×${HEIGHT}`);
console.log(`  combinaciones posibles: ${COLOR_NAMES.length * (COLOR_NAMES.length - 1) * BODY_PATTERNS.length * SLEEVE_PATTERNS.length}`);
console.log(`  colisiones dentro de una carrera: ${clashes.length}   ← tiene que ser 0`);
console.log(`  coincidencias entre carreras:     ${crossRace.length}   ← aceptable`);
for (const c of crossRace) console.log(`      ${c}`);
console.log(`  con ΔL insuficiente: ${badGap.length}`);
console.log(`  reparto de patrones: ${BODY_PATTERNS.map((p) => `${p} ${spread[p] ?? 0}`).join(', ')}`);

if (clashes.length || badGap.length) {
  console.error('\n  ✗ Gramática de sedas:');
  for (const c of [...clashes, ...badGap.map((s) => `${s.id}: ΔL insuficiente`)]) {
    console.error(`    · ${c}`);
  }
  process.exit(1);
}
