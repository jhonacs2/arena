/**
 * Conversión de color oklch → sRGB, y de ahí a hex y a contraste WCAG.
 *
 * Vive acá porque la usan dos consumidores con motivos distintos:
 * `check-contrast.mjs` para medir la paleta, y `gen-diagram-svg.mjs` porque
 * Excalidraw guarda los colores en hex y no entiende oklch. Tener la matriz
 * OKLab escrita dos veces es exactamente el tipo de duplicación que después se
 * desincroniza sin que nada avise.
 */

/** oklch → sRGB lineal. Devuelve también si cayó fuera de gamut. */
export function oklchToLinearRgb([L, C, H]) {
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

export const luminance = ([r, g, b]) => 0.2126 * r + 0.7152 * g + 0.0722 * b;

export function contrast(a, b) {
  const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return (hi + 0.05) / (lo + 0.05);
}

/** sRGB lineal → `#rrggbb`. Aplica la curva de transferencia sRGB. */
export function toHex([r, g, b]) {
  const enc = (c) => (c <= 0.0031308 ? 12.92 * c : 1.055 * c ** (1 / 2.4) - 0.055);
  return (
    '#' +
    [r, g, b]
      .map((c) =>
        Math.round(enc(c) * 255)
          .toString(16)
          .padStart(2, '0'),
      )
      .join('')
  );
}

/** Atajo para el caso más común: un oklch del tokens.json directo a hex. */
export const oklchToHex = (oklch) => toHex(oklchToLinearRgb(oklch).rgb);
