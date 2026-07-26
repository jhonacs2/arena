/**
 * Sedas de jockey — el elemento firma del sistema (`docs/design/tokens.md` §1).
 *
 * Cada caballo tiene su casaca, derivada de su `id` con una función pura:
 * mismo id, misma seda, siempre. En la lista, en el detalle, en la carrera en
 * vivo y en el historial.
 *
 * Es lo que hace que 54 caballos tengan identidad visual sin un solo archivo
 * de imagen, y que el color de acento de cada card se elija solo.
 *
 * ⚠️ Esta es la MISMA especificación que implementa `scripts/gen-silks-specimen.mjs`.
 * Si se cambia acá, hay que cambiarla allá y regenerar la hoja de muestra.
 * `silk.util.spec.ts` compara las dos y falla si se separan.
 */

export type SilkColor =
  | 'black'
  | 'white'
  | 'red'
  | 'blue'
  | 'yellow'
  | 'green'
  | 'orange'
  | 'purple'
  | 'pink'
  | 'cyan';

export type SilkBody =
  | 'solid'
  | 'halves'
  | 'quarters'
  | 'stripes'
  | 'hoops'
  | 'chevron'
  | 'sash'
  | 'star'
  | 'diamond'
  | 'seams';

export type SilkSleeves = 'plain' | 'alt' | 'hooped' | 'striped';

export interface SilkSpec {
  readonly primary: SilkColor;
  readonly secondary: SilkColor;
  readonly body: SilkBody;
  readonly sleeves: SilkSleeves;
}

/** Los diez colores registrados del deporte. El orden es parte del contrato. */
export const SILK_COLORS: readonly SilkColor[] = [
  'black',
  'white',
  'red',
  'blue',
  'yellow',
  'green',
  'orange',
  'purple',
  'pink',
  'cyan',
];

const SILK_BODIES: readonly SilkBody[] = [
  'solid',
  'halves',
  'quarters',
  'stripes',
  'hoops',
  'chevron',
  'sash',
  'star',
  'diamond',
  'seams',
];

const SILK_SLEEVES: readonly SilkSleeves[] = ['plain', 'alt', 'hooped', 'striped'];

/** Luminancia de cada color, de `docs/design/tokens.json`. */
const LIGHTNESS: Readonly<Record<SilkColor, number>> = {
  black: 0.2,
  white: 0.98,
  red: 0.55,
  blue: 0.48,
  yellow: 0.85,
  green: 0.52,
  orange: 0.7,
  purple: 0.45,
  pink: 0.75,
  cyan: 0.78,
};

/**
 * Separación mínima de luminancia entre los dos colores de una seda.
 *
 * Sin esta regla salen sedas azul-sobre-violeta que a 24 px se ven como un
 * cuadrado sólido, y el sistema deja de identificar caballos.
 */
const MIN_LIGHTNESS_GAP = 0.22;

/** FNV-1a de 32 bits. */
function fnv1a(text: string): number {
  let hash = 0x811c9dc5;
  for (let i = 0; i < text.length; i++) {
    hash ^= text.charCodeAt(i);
    hash = Math.imul(hash, 0x01000193) >>> 0;
  }
  return hash;
}

/** Sin la mezcla final, ids parecidos dan sedas parecidas. */
function mix32(hash: number): number {
  let h = (hash ^ (hash >>> 15)) >>> 0;
  h = Math.imul(h, 0x2545f491) >>> 0;
  return (h ^ (h >>> 13)) >>> 0;
}

/**
 * `id` → especificación de seda. Función pura, sin estado y sin azar.
 *
 * 10 × 9 × 10 × 4 = 3600 combinaciones. Con 54 caballos no se repite ninguna
 * dentro de una misma carrera, que es la única condición que importa: entre
 * carreras distintas, compartir seda es lo que pasa en el hipódromo real —
 * las sedas son del dueño, no del caballo.
 */
export function silkFromId(id: string): SilkSpec {
  let hash = fnv1a(id);
  const next = (): number => (hash = mix32(hash));

  const primary = SILK_COLORS[hash % SILK_COLORS.length] ?? 'black';

  let secondary: SilkColor = primary;
  for (let attempt = 0; attempt < 32; attempt++) {
    next();
    const candidate = SILK_COLORS[hash % SILK_COLORS.length] ?? 'black';
    if (
      candidate !== primary &&
      Math.abs(LIGHTNESS[candidate] - LIGHTNESS[primary]) >= MIN_LIGHTNESS_GAP
    ) {
      secondary = candidate;
      break;
    }
  }

  next();
  const body = SILK_BODIES[hash % SILK_BODIES.length] ?? 'solid';
  next();
  const sleeves = SILK_SLEEVES[hash % SILK_SLEEVES.length] ?? 'plain';

  return { primary, secondary, body, sleeves };
}

// ── Geometría ─────────────────────────────────────────────────────────────

/** viewBox 0 0 40 38. Cuerpo en 12..28, mangas en 3..12 y 28..37. */
const BODY = { x: 12, y: 4, w: 16, h: 30 } as const;
const SLEEVE = { w: 9, h: 14, y: 6, left: 3, right: 28 } as const;

/**
 * Una figura del SVG. El template las pinta con un `@for`, así la geometría
 * vive en TypeScript —donde se puede testear— y no repartida en el HTML.
 */
export type SilkShape =
  | { readonly kind: 'rect'; readonly x: number; readonly y: number; readonly w: number; readonly h: number; readonly fill: string }
  | { readonly kind: 'poly'; readonly points: string; readonly fill: string }
  | { readonly kind: 'path'; readonly d: string; readonly fill: string };

const paint = (color: SilkColor): string => `var(--silk-${color})`;

const rect = (x: number, y: number, w: number, h: number, fill: string): SilkShape => ({
  kind: 'rect',
  x,
  y,
  w,
  h,
  fill,
});

function bodyShapes(spec: SilkSpec): SilkShape[] {
  const { x, y, w, h } = BODY;
  const half = w / 2;
  const first = paint(spec.primary);
  const second = paint(spec.secondary);
  const base = rect(x, y, w, h, first);

  switch (spec.body) {
    case 'solid':
      return [base];

    case 'halves':
      return [base, rect(x + half, y, half, h, second)];

    case 'quarters':
      return [
        base,
        rect(x + half, y, half, h / 2, second),
        rect(x, y + h / 2, half, h / 2, second),
      ];

    case 'stripes':
      return [base, ...[1, 3].map((i) => rect(x + i * (w / 4), y, w / 4, h, second))];

    case 'hoops':
      return [base, ...[1, 3].map((i) => rect(x, y + i * (h / 5), w, h / 5, second))];

    case 'chevron':
      return [
        base,
        ...[0, 1, 2].map((i) => {
          const top = y + 3 + i * 9;
          return {
            kind: 'path' as const,
            fill: second,
            d: `M${x},${top} L${x + half},${top + 5} L${x + w},${top} L${x + w},${top + 4} L${x + half},${top + 9} L${x},${top + 4} Z`,
          };
        }),
      ];

    case 'sash':
      return [
        base,
        {
          kind: 'path',
          fill: second,
          d: `M${x},${y + h - 6} L${x + w - 6},${y} L${x + w},${y} L${x + w},${y + 4} L${x + 4},${y + h} L${x},${y + h} Z`,
        },
      ];

    case 'star': {
      const cx = x + half;
      const cy = y + h / 2;
      const points = Array.from({ length: 10 }, (_, i) => {
        const radius = i % 2 === 0 ? 6.5 : 2.7;
        const angle = (Math.PI / 5) * i - Math.PI / 2;
        return `${(cx + radius * Math.cos(angle)).toFixed(2)},${(cy + radius * Math.sin(angle)).toFixed(2)}`;
      }).join(' ');
      return [base, { kind: 'poly', points, fill: second }];
    }

    case 'diamond': {
      const cx = x + half;
      const cy = y + h / 2;
      return [
        base,
        { kind: 'poly', fill: second, points: `${cx},${cy - 8} ${cx + 6},${cy} ${cx},${cy + 8} ${cx - 6},${cy}` },
      ];
    }

    case 'seams':
      return [base, rect(x + half - 2, y, 4, h, second)];
  }
}

function sleeveShapes(spec: SilkSpec): SilkShape[] {
  const { w, h, y } = SLEEVE;
  const first = paint(spec.primary);
  const second = paint(spec.secondary);

  return [SLEEVE.left, SLEEVE.right].flatMap((x): SilkShape[] => {
    switch (spec.sleeves) {
      case 'plain':
        return [rect(x, y, w, h, first)];
      case 'alt':
        return [rect(x, y, w, h, second)];
      case 'hooped':
        return [
          rect(x, y, w, h, first),
          ...[1, 3].map((i) => rect(x, y + i * (h / 4), w, h / 4, second)),
        ];
      case 'striped':
        return [rect(x, y, w, h, first), rect(x + w / 3, y, w / 3, h, second)];
    }
  });
}

/** Contornos: el borde de tinta que define la silueta en los dos temas. */
const OUTLINES: readonly SilkShape[] = [
  rect(BODY.x, BODY.y, BODY.w, BODY.h, 'none'),
  rect(SLEEVE.left, SLEEVE.y, SLEEVE.w, SLEEVE.h, 'none'),
  rect(SLEEVE.right, SLEEVE.y, SLEEVE.w, SLEEVE.h, 'none'),
];

export interface SilkDrawing {
  readonly spec: SilkSpec;
  readonly fills: readonly SilkShape[];
  readonly outlines: readonly SilkShape[];
  /** El cuello, en tinta. */
  readonly collar: string;
}

/** Todo lo que el template necesita para pintar una seda. */
export function silkDrawing(id: string): SilkDrawing {
  const spec = silkFromId(id);
  return {
    spec,
    fills: [...sleeveShapes(spec), ...bodyShapes(spec)],
    outlines: OUTLINES,
    collar: 'M17,4 L23,4 L21.5,7.5 L18.5,7.5 Z',
  };
}
