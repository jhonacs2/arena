/**
 * Renderiza una escena de Excalidraw a SVG.
 *
 * El `.excalidraw` es la fuente: se abre en excalidraw.com, se edita a mano y se
 * guarda encima. Este módulo lo convierte al `.svg` que consumen las
 * diapositivas de Marp, así que el diagrama vive en un solo lugar y no hay dos
 * versiones que se separen.
 *
 * Dos límites, a propósito y documentados:
 *
 * 1. **`roughness` se ignora.** El trazo «a mano alzada» de Excalidraw lo genera
 *    rough.js con su propio ruido; reproducirlo acá sería portar una librería
 *    entera para un efecto que el neobrutalismo no quiere. Los diagramas se
 *    autoran con `roughness: 0` (estilo «arquitecto»), y entonces lo que se ve
 *    en Excalidraw es lo que sale en el SVG.
 * 2. **Se soportan los tipos que usamos:** rectangle, ellipse, diamond, line,
 *    arrow y text. Si aparece otro, `render` lo reporta en vez de dibujar mal en
 *    silencio.
 */

const FONTS = {
  1: "'Bricolage Grotesque', 'Segoe UI', system-ui, sans-serif",
  2: "'Public Sans', 'Segoe UI', system-ui, sans-serif",
  3: "'Martian Mono', ui-monospace, 'Cascadia Mono', monospace",
};

const esc = (s) =>
  String(s)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');

const num = (n) => (Math.round(n * 100) / 100).toString();

/** Redondeo de esquinas. Excalidraw usa `roundness: null` para el ángulo vivo. */
const radiusOf = (el) => {
  if (!el.roundness) return 0;
  // type 3 = radio proporcional al lado menor, con techo. Es lo que hace la app.
  const smaller = Math.min(Math.abs(el.width), Math.abs(el.height));
  return Math.min(smaller * 0.25, 32);
};

/**
 * Los colores literales del `.excalidraw` se reescriben a `var(--token, #hex)`
 * cuando coinciden con un token de la paleta. Así el diagrama sigue el modo
 * oscuro sin que Excalidraw tenga que saber qué es un token: el fallback es el
 * hex que ya estaba, y un color elegido a mano en la app queda tal cual.
 *
 * Adivinar el token por el hex no siempre alcanza. En claro `text`, `border`,
 * `shadow` y `accent-ink` son **todos** ink-900 — pero en oscuro se separan:
 * `text` se vuelve tiza y `shadow` se vuelve casi negro. Un elemento puede
 * entonces declarar el token a mano en `customData`, que Excalidraw preserva al
 * guardar:
 *
 *     "customData": { "strokeToken": "border", "backgroundToken": "shadow" }
 */
function makePaint(palette) {
  return (hex, explicit) => {
    if (explicit) return `var(--dg-${explicit}, ${hex ?? 'currentColor'})`;
    if (!hex || hex === 'transparent') return 'none';
    const token = palette.get(hex.toLowerCase());
    return token ? `var(--${token}, ${hex})` : hex;
  };
}

/** Puntos absolutos de una línea o flecha. */
const absPoints = (el) => (el.points ?? [[0, 0]]).map(([x, y]) => [el.x + x, el.y + y]);

/** Nombre estable de marcador por (color, forma, punta). */
const markerId = (color, shape, side) =>
  `mk-${shape}-${side}-${color.replace(/[^a-z0-9]/gi, '')}`;

function markerDef(id, shape, side, stroke) {
  // `orient="auto"` alinea con la tangente; para la punta de inicio se dibuja
  // espejada, que es más simple y más predecible que auto-start-reverse.
  const flip = side === 'start' ? ' transform="rotate(180 5 5)"' : '';
  const body =
    shape === 'triangle'
      ? `<path d="M 0 0 L 10 5 L 0 10 z" fill="${stroke}"${flip} />`
      : shape === 'dot'
        ? `<circle cx="5" cy="5" r="4" fill="${stroke}" />`
        : shape === 'bar'
          ? `<path d="M 5 0 L 5 10" stroke="${stroke}" stroke-width="2" fill="none"${flip} />`
          : `<path d="M 1 1 L 9 5 L 1 9" stroke="${stroke}" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"${flip} />`;

  return (
    `<marker id="${id}" viewBox="0 0 10 10" refX="${shape === 'dot' ? 5 : side === 'start' ? 1 : 9}" refY="5" ` +
    `markerWidth="5" markerHeight="5" orient="auto" markerUnits="strokeWidth">${body}</marker>`
  );
}

/**
 * @param scene     el JSON del `.excalidraw`
 * @param palette   Map de `#hex` (minúscula) → nombre de custom property
 * @param darkCss   bloque CSS que redefine esas custom properties en oscuro
 * @param ambiguous Set de `#hex` que en claro son un color y en oscuro varios
 * @returns {{svg: string, problems: string[]}}
 */
export function render(
  scene,
  { palette = new Map(), darkCss = '', padding = 24, ambiguous = new Set() } = {},
) {
  const problems = [];
  const paint = makePaint(palette);

  // Para un hex ambiguo, el rol del elemento alcanza para elegir bien: el trazo
  // de una forma es su borde, el «trazo» de un texto es el texto. Sin este
  // default, cualquier rectángulo dibujado con el color por defecto exigiría un
  // `customData` a mano — y un checker que grita por nada se empieza a ignorar.
  const strokeRole = (el) =>
    ambiguous.has((el.strokeColor ?? '').toLowerCase())
      ? el.type === 'text'
        ? 'text'
        : 'border'
      : undefined;

  const elements = (scene.elements ?? []).filter((el) => !el.isDeleted);
  if (!elements.length) return { svg: '', problems: ['la escena no tiene elementos'] };

  // ── Caja envolvente ─────────────────────────────────────────────────────
  let minX = Infinity;
  let minY = Infinity;
  let maxX = -Infinity;
  let maxY = -Infinity;
  for (const el of elements) {
    const xs = el.points ? absPoints(el).map((p) => p[0]) : [el.x, el.x + el.width];
    const ys = el.points ? absPoints(el).map((p) => p[1]) : [el.y, el.y + el.height];
    minX = Math.min(minX, ...xs);
    maxX = Math.max(maxX, ...xs);
    minY = Math.min(minY, ...ys);
    maxY = Math.max(maxY, ...ys);
  }
  // La sombra dura del neobrutalismo se dibuja como un rectángulo más, así que
  // ya está contada. El padding es solo aire.
  minX -= padding;
  minY -= padding;
  maxX += padding;
  maxY += padding;

  const width = Math.round(maxX - minX);
  const height = Math.round(maxY - minY);

  // ── Cuerpo ──────────────────────────────────────────────────────────────
  const markers = new Map();
  const byId = new Map(elements.map((el) => [el.id, el]));
  const body = [];

  for (const el of elements) {
    // El texto atado a un contenedor se dibuja en su propio turno, centrado
    // sobre el contenedor: si se dibujara acá quedaría en el origen.
    const stroke = paint(el.strokeColor, el.customData?.strokeToken ?? strokeRole(el));
    const fill = paint(el.backgroundColor, el.customData?.backgroundToken);
    const sw = el.strokeWidth ?? 2;
    const dash =
      el.strokeStyle === 'dashed'
        ? ` stroke-dasharray="${sw * 4} ${sw * 3}"`
        : el.strokeStyle === 'dotted'
          ? ` stroke-dasharray="${sw} ${sw * 2}"`
          : '';
    const op = (el.opacity ?? 100) / 100;
    const opAttr = op === 1 ? '' : ` opacity="${num(op)}"`;

    // El rayado de Excalidraw también lo dibuja rough.js. En vez de renderizarlo
    // como relleno sólido —que se vería distinto en el SVG que en la app— se
    // reporta: los diagramas se autoran con `fillStyle: "solid"`.
    if (fill !== 'none' && el.fillStyle && el.fillStyle !== 'solid') {
      problems.push(`${el.type} ${el.id}: fillStyle "${el.fillStyle}" no se renderiza — usá "solid"`);
    }

    switch (el.type) {
      case 'rectangle': {
        const r = radiusOf(el);
        body.push(
          `<rect x="${num(el.x)}" y="${num(el.y)}" width="${num(el.width)}" height="${num(el.height)}"` +
            (r ? ` rx="${num(r)}"` : '') +
            ` fill="${fill}" stroke="${stroke}" stroke-width="${sw}"${dash}${opAttr} />`,
        );
        break;
      }
      case 'ellipse':
        body.push(
          `<ellipse cx="${num(el.x + el.width / 2)}" cy="${num(el.y + el.height / 2)}" ` +
            `rx="${num(Math.abs(el.width) / 2)}" ry="${num(Math.abs(el.height) / 2)}" ` +
            `fill="${fill}" stroke="${stroke}" stroke-width="${sw}"${dash}${opAttr} />`,
        );
        break;
      case 'diamond': {
        const cx = el.x + el.width / 2;
        const cy = el.y + el.height / 2;
        const pts = [
          [cx, el.y],
          [el.x + el.width, cy],
          [cx, el.y + el.height],
          [el.x, cy],
        ]
          .map(([x, y]) => `${num(x)},${num(y)}`)
          .join(' ');
        body.push(
          `<polygon points="${pts}" fill="${fill}" stroke="${stroke}" stroke-width="${sw}"${dash}${opAttr} />`,
        );
        break;
      }
      case 'line':
      case 'arrow': {
        const pts = absPoints(el);
        const d = pts.map(([x, y], i) => `${i ? 'L' : 'M'} ${num(x)} ${num(y)}`).join(' ');

        let ends = '';
        for (const [side, shape] of [
          ['start', el.startArrowhead],
          ['end', el.endArrowhead ?? (el.type === 'arrow' ? 'arrow' : null)],
        ]) {
          if (!shape) continue;
          const id = markerId(stroke, shape, side);
          if (!markers.has(id)) markers.set(id, markerDef(id, shape, side, stroke));
          ends += ` marker-${side}="url(#${id})"`;
        }

        body.push(
          `<path d="${d}" fill="none" stroke="${stroke}" stroke-width="${sw}" ` +
            `stroke-linecap="round" stroke-linejoin="round"${dash}${ends}${opAttr} />`,
        );
        break;
      }
      case 'text': {
        const size = el.fontSize ?? 16;
        const lh = size * (el.lineHeight ?? 1.25);
        const family = FONTS[el.fontFamily] ?? FONTS[2];
        const lines = String(el.text ?? '').split('\n');

        // Un texto atado a un contenedor se centra sobre él, no sobre su propia
        // caja: Excalidraw guarda la caja del texto desactualizada a menudo.
        const box = el.containerId ? byId.get(el.containerId) : null;
        const align = el.textAlign ?? (box ? 'center' : 'left');
        const left = box ? box.x : el.x;
        const boxW = box ? box.width : el.width;
        const anchor = align === 'center' ? 'middle' : align === 'right' ? 'end' : 'start';
        const tx = align === 'center' ? left + boxW / 2 : align === 'right' ? left + boxW : left;

        const blockH = lines.length * lh;
        const top = box
          ? box.y + (box.height - blockH) / 2
          : (el.verticalAlign === 'middle' && el.height
              ? el.y + (el.height - blockH) / 2
              : el.y);

        // Excalidraw no tiene negrita: el peso se deduce del tamaño, que es como
        // se autoran los diagramas —título grande, cuerpo chico— y así lo que se
        // ve en la app y lo que sale en el SVG coinciden.
        const weight = size >= 20 ? '800' : size >= 17 ? '700' : '500';
        const tspans = lines
          .map(
            (line, i) =>
              `<tspan x="${num(tx)}" y="${num(top + i * lh + lh / 2)}">${esc(line)}</tspan>`,
          )
          .join('');

        body.push(
          `<text font-family="${family}" font-size="${num(size)}" font-weight="${weight}" ` +
            `fill="${stroke}" text-anchor="${anchor}" dominant-baseline="central"${opAttr}>${tspans}</text>`,
        );
        break;
      }
      default:
        problems.push(`tipo de elemento sin soporte: ${el.type}`);
    }
  }

  const defs = markers.size ? `\n  <defs>\n    ${[...markers.values()].join('\n    ')}\n  </defs>` : '';
  const bg = paint(scene.appState?.viewBackgroundColor ?? 'transparent');

  const svg =
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="${num(minX)} ${num(minY)} ${width} ${height}" ` +
    `width="${width}" height="${height}" role="img">\n` +
    (darkCss ? `  <style>\n${darkCss}\n  </style>\n` : '') +
    defs +
    (bg === 'none' ? '' : `\n  <rect x="${num(minX)}" y="${num(minY)}" width="${width}" height="${height}" fill="${bg}" />`) +
    '\n  ' +
    body.join('\n  ') +
    '\n</svg>\n';

  return { svg, problems };
}
