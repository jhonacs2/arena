/**
 * Verifica que la paleta de docs/design/tokens.json cumpla WCAG AA.
 *
 *   node scripts/check-contrast.mjs           informe completo
 *   node scripts/check-contrast.mjs --quiet   solo falla o calla
 *
 * `docs/design/CLAUDE.md` declara el contraste AA como regla dura sin excepciones. El
 * neobrutalismo falla exactamente acá — acentos saturados sobre blanco se ven
 * bien y no se leen. Una regla que no se mide es una intención, así que esto
 * corre dentro de scripts/verify.mjs.
 *
 * También avisa si un color oklch cae fuera del gamut sRGB: el navegador lo
 * recortaría en silencio y el contraste real dejaría de ser el calculado.
 */

import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const ROOT = join(dirname(fileURLToPath(import.meta.url)), '..');
const tokens = JSON.parse(readFileSync(join(ROOT, 'docs/design/tokens.json'), 'utf8'));

const quiet = process.argv.includes('--quiet');
const THRESHOLD = { aa: 4.5, aaLarge: 3.0 };

/** oklch → sRGB lineal. Devuelve también si cayó fuera de gamut. */
function oklchToLinearRgb([L, C, H]) {
  const h = (H * Math.PI) / 180;
  const a = C * Math.cos(h);
  const b = C * Math.sin(h);

  const l_ = L + 0.3963377774 * a + 0.2158037573 * b;
  const m_ = L - 0.1055613458 * a - 0.0638541728 * b;
  const s_ = L - 0.0894841775 * a - 1.291485548 * b;

  const l = l_ ** 3;
  const m = m_ ** 3;
  const s = s_ ** 3;

  const rgb = [
    4.0767416621 * l - 3.3077115913 * m + 0.2309699292 * s,
    -1.2684380046 * l + 2.6097574011 * m - 0.3413193965 * s,
    -0.0041960863 * l - 0.7034186147 * m + 1.707614701 * s,
  ];

  const inGamut = rgb.every((c) => c >= -0.0005 && c <= 1.0005);
  return { rgb: rgb.map((c) => Math.min(Math.max(c, 0), 1)), inGamut };
}

const luminance = ([r, g, b]) => 0.2126 * r + 0.7152 * g + 0.0722 * b;

function contrast(a, b) {
  const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return (hi + 0.05) / (lo + 0.05);
}

function toHex([r, g, b]) {
  const enc = (c) => (c <= 0.0031308 ? 12.92 * c : 1.055 * c ** (1 / 2.4) - 0.055);
  return (
    '#' +
    [r, g, b]
      .map((c) => Math.round(enc(c) * 255).toString(16).padStart(2, '0'))
      .join('')
  );
}

// ── Gamut de todas las primitivas y sedas ─────────────────────────────────
const failures = [];
const resolved = new Map();

for (const [group, entries] of [
  ['primitives', tokens.primitives],
  ['silks', tokens.silks],
]) {
  for (const [name, def] of Object.entries(entries)) {
    const { rgb, inGamut } = oklchToLinearRgb(def.oklch);
    resolved.set(name, rgb);
    if (!inGamut) {
      failures.push(`${group}.${name} — oklch(${def.oklch.join(' ')}) cae fuera del gamut sRGB`);
    }
  }
}

// ── Separación entre superficies ──────────────────────────────────────────
// Con las tres superficies a la misma luminancia, una pantalla oscura se ve
// plana y sucia: el borde queda como único indicio y termina compitiendo con
// el texto. No es una regla de contraste WCAG, es una regla del sistema.
const lightnessOf = (name) =>
  tokens.primitives[name]?.oklch[0] ?? tokens.silks[name]?.oklch[0];

const surfaceRows = [];

for (const [theme, map] of Object.entries(tokens.semantic)) {
  for (const [a, b] of tokens.surfaceSeparation.pairs) {
    const deltaL = Math.abs(lightnessOf(map[a]) - lightnessOf(map[b]));
    const ok = deltaL >= tokens.surfaceSeparation.minDeltaL;

    surfaceRows.push({ theme, a, b, nameA: map[a], nameB: map[b], deltaL, ok });
    if (!ok) {
      failures.push(
        `${theme}: --${a} y --${b} son casi el mismo color ` +
          `(${map[a]} vs ${map[b]}, ΔL ${deltaL.toFixed(3)}, mínimo ${tokens.surfaceSeparation.minDeltaL})`,
      );
    }
  }
}

// ── Pares de contraste ────────────────────────────────────────────────────
const rows = [];

for (const pair of tokens.contrastPairs) {
  const map = tokens.semantic[pair.theme];
  const fgName = map[pair.fg];
  const bgName = map[pair.bg];
  const fg = resolved.get(fgName);
  const bg = resolved.get(bgName);

  if (!fg || !bg) {
    failures.push(`par ${pair.theme}/${pair.fg}-sobre-${pair.bg} apunta a una primitiva inexistente`);
    continue;
  }

  const ratio = contrast(fg, bg);
  const min = THRESHOLD[pair.level];
  const ok = ratio >= min;

  rows.push({ ...pair, fgName, bgName, ratio, min, ok });
  if (!ok) {
    failures.push(
      `${pair.theme}: ${pair.what} — ${fgName} sobre ${bgName} da ${ratio.toFixed(2)}:1, necesita ${min}:1`,
    );
  }
}

// ── Informe ───────────────────────────────────────────────────────────────
if (!quiet) {
  const pad = (s, n) => String(s).padEnd(n);
  let theme = '';
  console.log('\n  Contraste WCAG · docs/design/tokens.json\n');

  for (const r of rows) {
    if (r.theme !== theme) {
      theme = r.theme;
      console.log(`  ── ${theme === 'light' ? 'claro' : 'oscuro'} ${'─'.repeat(52)}`);
    }
    console.log(
      `  ${r.ok ? '✓' : '✗'} ${pad(r.what, 26)} ${pad(`${r.fgName} / ${r.bgName}`, 26)}` +
        `${r.ratio.toFixed(2).padStart(6)}:1  (min ${r.min})`,
    );
  }

  console.log('\n  ── separación entre superficies ' + '─'.repeat(30));
  for (const r of surfaceRows) {
    console.log(
      `  ${r.ok ? '✓' : '✗'} ${pad(r.theme === 'light' ? 'claro' : 'oscuro', 8)}` +
        `${pad(`--${r.a} / --${r.b}`, 34)}ΔL ${r.deltaL.toFixed(3)}` +
        `  (min ${tokens.surfaceSeparation.minDeltaL})`,
    );
  }

  console.log('\n  ── sedas ' + '─'.repeat(52));
  const swatches = Object.keys(tokens.silks)
    .map((n) => `${n} ${toHex(resolved.get(n))}`)
    .join('   ');
  console.log('  ' + swatches.replace(/(.{72}\S*)\s/g, '$1\n  '));
  console.log('');
}

if (failures.length) {
  console.error('\n  ✗ Contraste — la paleta no cumple:\n');
  for (const f of failures) console.error(`    · ${f}`);
  console.error('');
  process.exit(1);
}

if (!quiet) console.log('  ✓ Toda la paleta cumple AA.\n');
