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
import { oklchToLinearRgb, contrast, toHex } from './lib/oklch.mjs';

const ROOT = join(dirname(fileURLToPath(import.meta.url)), '..');
const tokens = JSON.parse(readFileSync(join(ROOT, 'docs/design/tokens.json'), 'utf8'));

const quiet = process.argv.includes('--quiet');
const THRESHOLD = { aa: 4.5, aaLarge: 3.0 };

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
