#!/usr/bin/env node
/**
 * Escribe el deck de reveal.js de cada sesión a partir de su slides.md.
 *
 *   node scripts/gen-reveal.mjs                          todas las sesiones
 *   node scripts/gen-reveal.mjs sesiones/s00-typescript  una sola
 *   node scripts/gen-reveal.mjs --check                  falla si alguno quedó desfasado
 *
 * `slides.md` es el original y sigue siendo Marp: el guión de cada diapositiva
 * vive en sus comentarios y el instructor lo edita ahí. Este script traduce ese
 * mismo archivo a reveal.js, que es lo que se proyecta.
 *
 *   <!-- _class: codigo -->   →  <section class="codigo">
 *   <!-- ...el resto... -->   →  <aside class="notes">   (tecla S)
 *   ![w:900](diagramas/x.svg) →  el SVG pegado inline
 *
 * Tres decisiones que no son obvias:
 *
 * 1. **El markdown se convierte acá y no en el navegador.** Reveal trae un
 *    plugin de markdown que haría lo mismo en tiempo de carga, pero entonces el
 *    HTML del deck sería un `<textarea>` con todo adentro: no se puede leer, no
 *    se puede diffear y no se puede pegar un SVG en el medio sin pelearse con el
 *    parser. Acá el HTML sale escrito.
 *
 * 2. **Los diagramas se pegan inline y se les saca el modo oscuro.** El SVG que
 *    genera gen-diagram-svg.mjs trae su propio `prefers-color-scheme`, y dentro
 *    de un `<img>` ese media query lo resuelve el sistema operativo de la
 *    máquina que proyecta. Una diapositiva clara con el diagrama en oscuro es
 *    exactamente lo que sesiones/CLAUDE.md prohíbe: el deck se ve siempre igual.
 *
 * 3. **Los fondos a sangre salen de tokens.json.** `data-background-color` cubre
 *    también las bandas del proyector cuando la sala no es 16:9. El color no se
 *    escribe acá: se lee del JSON, como todo el resto del color del repo.
 */

import { readFileSync, writeFileSync, existsSync, readdirSync } from 'node:fs';
import { dirname, join, relative, basename } from 'node:path';
import { fileURLToPath } from 'node:url';

const ROOT = join(dirname(fileURLToPath(import.meta.url)), '..');
const args = process.argv.slice(2);
const CHECK = args.includes('--check');
const explicit = args.filter((a) => !a.startsWith('--'));

const OUTPUT_NAME = 'slides.reveal.html';

const tokens = JSON.parse(readFileSync(join(ROOT, 'docs/design/tokens.json'), 'utf8'));

/** Un color de tokens.json, listo para escribir en un atributo. */
const oklchOf = (name) => {
  const def = tokens.primitives[name] ?? tokens.silks[name];
  if (!def) throw new Error(`no existe la primitiva --${name} en tokens.json`);
  return `oklch(${def.oklch.join(' ')})`;
};

/**
 * Las tres clases a sangre y de dónde sale su fondo. Las demás diapositivas
 * usan --surface, que ya pinta .reveal-viewport en el tema.
 */
const FULL_BLEED = {
  // Las claves van entre comillas porque no son identificadores: son los
  // nombres de clase que escribe slides.md, y esos están en castellano igual
  // que en el tema de Marp.
  'portada': () => oklchOf('ink-900'),
  'bloque': () => oklchOf(tokens.semantic.light.accent),
  'ojo': () => oklchOf(tokens.semantic.light.live),
};

// ══ MARKDOWN ═══════════════════════════════════════════════════════════════
// El subconjunto que usan las diapositivas del módulo, y nada más: títulos,
// párrafos, listas, tablas, citas, código y la imagen del diagrama.

const escapeHtml = (text) =>
  text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

/**
 * Marcas de línea: negrita, énfasis, código y enlaces.
 *
 * El código va primero y sale de la pasada guardado aparte: adentro de un
 * `code` no hay negrita ni enlaces, y `**` o `_` son caracteres literales.
 */
// Marcador de un `code` mientras corren las demás marcas. Es un carácter del
// área de uso privado: no existe en ningún teclado ni en ninguna diapositiva,
// así que no hay forma de que choque con el texto del guión.
const CODE_MARK = String.fromCharCode(0xe000);

function inline(text) {
  const codes = [];
  const masked = text.replace(/`([^`]+)`/g, (_, body) => {
    codes.push(`<code>${escapeHtml(body)}</code>`);
    return CODE_MARK;
  });

  // La negrita se aplica sobre el texto enmascarado y no sobre los pedazos
  // sueltos: en la tabla de las 0:12 hay un `**Opcionales y \`undefined\`**`, y
  // con el código ya reemplazado esos dos asteriscos no se encuentran nunca.
  const marked = escapeHtml(masked)
    .replace(/\[([^\]]+)\]\(([^)\s]+)\)/g, '<a href="$2">$1</a>')
    .replace(/\*\*([\s\S]+?)\*\*/g, '<strong>$1</strong>')
    .replace(/(^|[^*\w])\*([^*\n]+)\*(?![*\w])/g, '$1<em>$2</em>');

  return marked
    .split(CODE_MARK)
    .map((part, index) => (index === 0 ? part : codes[index - 1] + part))
    .join('');
}

/** Una fila de tabla en celdas, sin los pipes de los extremos. */
const cellsOf = (line) =>
  line
    .trim()
    .replace(/^\|/, '')
    .replace(/\|$/, '')
    .split('|')
    .map((cell) => cell.trim());

const isTableDivider = (line) => /^\s*\|?\s*:?-{2,}:?\s*(\|\s*:?-{2,}:?\s*)*\|?\s*$/.test(line);

/**
 * Markdown → HTML. Recibe el cuerpo de una diapositiva o de una nota.
 * `onImage` decide qué hacer con `![...](...)`: en la diapositiva pega el SVG,
 * en las notas no debería aparecer nunca.
 */
function renderMarkdown(source, onImage = null) {
  const lines = source.split('\n');
  const html = [];
  let i = 0;

  const paragraph = [];
  const flushParagraph = () => {
    if (!paragraph.length) return;
    html.push(`<p>${inline(paragraph.join(' '))}</p>`);
    paragraph.length = 0;
  };

  while (i < lines.length) {
    const line = lines[i];

    // ── Código con vallas ────────────────────────────────────────────────
    const fence = line.match(/^```(\S*)\s*$/);
    if (fence) {
      flushParagraph();
      const language = fence[1] || 'plaintext';
      const body = [];
      i++;
      while (i < lines.length && !/^```\s*$/.test(lines[i])) body.push(lines[i++]);
      i++;
      // Sin data-trim: la sangría del snippet es parte de lo que se enseña.
      html.push(
        `<pre><code class="language-${language}">${escapeHtml(body.join('\n'))}\n</code></pre>`,
      );
      continue;
    }

    if (!line.trim()) {
      flushParagraph();
      i++;
      continue;
    }

    // ── Imagen sola en su línea ──────────────────────────────────────────
    const image = line.match(/^!\[([^\]]*)\]\(([^)\s]+)\)\s*$/);
    if (image && onImage) {
      flushParagraph();
      html.push(onImage(image[2], image[1]));
      i++;
      continue;
    }

    // ── Título ───────────────────────────────────────────────────────────
    const heading = line.match(/^(#{1,6})\s+(.*)$/);
    if (heading) {
      flushParagraph();
      const level = heading[1].length;
      html.push(`<h${level}>${inline(heading[2].trim())}</h${level}>`);
      i++;
      continue;
    }

    // ── Tabla ────────────────────────────────────────────────────────────
    if (line.trim().startsWith('|') && isTableDivider(lines[i + 1] ?? '')) {
      flushParagraph();
      const head = cellsOf(line);
      i += 2;
      const body = [];
      while (i < lines.length && lines[i].trim().startsWith('|')) body.push(cellsOf(lines[i++]));

      const headRow = head.map((cell) => `<th>${inline(cell)}</th>`).join('');
      const bodyRows = body
        .map((row) => `<tr>${row.map((cell) => `<td>${inline(cell)}</td>`).join('')}</tr>`)
        .join('\n');
      html.push(`<table>\n<thead><tr>${headRow}</tr></thead>\n<tbody>\n${bodyRows}\n</tbody>\n</table>`);
      continue;
    }

    // ── Listas ───────────────────────────────────────────────────────────
    const bullet = line.match(/^\s*[-*+]\s+(.*)$/);
    const numbered = line.match(/^\s*\d+[.)]\s+(.*)$/);
    if (bullet || numbered) {
      flushParagraph();
      const tag = bullet ? 'ul' : 'ol';
      const re = bullet ? /^\s*[-*+]\s+(.*)$/ : /^\s*\d+[.)]\s+(.*)$/;
      const items = [];
      while (i < lines.length) {
        const match = lines[i].match(re);
        if (match) {
          items.push(match[1]);
          i++;
        } else if (lines[i].trim() && /^\s{2,}\S/.test(lines[i]) && items.length) {
          // continuación indentada del ítem anterior
          items[items.length - 1] += ' ' + lines[i].trim();
          i++;
        } else break;
      }
      html.push(`<${tag}>\n${items.map((item) => `<li>${inline(item)}</li>`).join('\n')}\n</${tag}>`);
      continue;
    }

    // ── Cita ─────────────────────────────────────────────────────────────
    if (/^>\s?/.test(line)) {
      flushParagraph();
      const quoted = [];
      while (i < lines.length && /^>\s?/.test(lines[i])) quoted.push(lines[i++].replace(/^>\s?/, ''));
      html.push(`<blockquote>${renderMarkdown(quoted.join('\n'))}</blockquote>`);
      continue;
    }

    paragraph.push(line.trim());
    i++;
  }

  flushParagraph();
  return html.join('\n');
}

// ══ DIAGRAMAS ══════════════════════════════════════════════════════════════

/** Recorta el bloque que empieza en `start` contando llaves. */
function cutBlock(text, start) {
  let depth = 0;
  for (let i = text.indexOf('{', start); i < text.length; i++) {
    if (text[i] === '{') depth++;
    else if (text[i] === '}' && --depth === 0) return text.slice(0, start) + text.slice(i + 1);
  }
  return text;
}

/**
 * Deja el SVG listo para pegarlo en la diapositiva: sin el modo oscuro, con
 * los ids prefijados —dos diagramas en un mismo deck comparten nombres de
 * marker— y con el ancho que pedía la sintaxis `w:900` de Marp.
 */
function inlineDiagram(svgPath, width, prefix) {
  let svg = readFileSync(svgPath, 'utf8').replace(/<\?xml[^>]*\?>\s*/, '');

  const dark = svg.indexOf('@media (prefers-color-scheme: dark)');
  if (dark !== -1) svg = cutBlock(svg, dark);

  svg = svg
    .replace(/\sid="([^"]+)"/g, ` id="${prefix}-$1"`)
    .replace(/url\(#([^)]+)\)/g, `url(#${prefix}-$1)`)
    .replace(/(\s(?:xlink:)?href=")#([^"]+)"/g, `$1#${prefix}-$2"`);

  const style = width ? ` style="width: ${width}px"` : '';
  return svg.replace(/^<svg\s/, `<svg class="diagram"${style} `);
}

// ══ DIAPOSITIVAS ═══════════════════════════════════════════════════════════

/** Separa el frontmatter YAML del cuerpo. Solo se leen las claves que se usan. */
function splitFrontMatter(text) {
  const match = text.match(/^---\r?\n([\s\S]*?)\r?\n---\r?\n/);
  if (!match) return { meta: {}, body: text };

  const meta = {};
  for (const line of match[1].split(/\r?\n/)) {
    const pair = line.match(/^([\w-]+):\s*(.*)$/);
    if (pair) meta[pair[1]] = pair[2].trim().replace(/^['"]|['"]$/g, '');
  }
  return { meta, body: text.slice(match[0].length) };
}

/**
 * Corta el cuerpo en diapositivas. El `---` de separación tiene que estar solo
 * en su línea y fuera de una valla de código: en las diapositivas de S0 hay
 * tablas con `|---|---|` y no son separadores.
 */
function splitSlides(body) {
  const slides = [];
  let current = [];
  let inFence = false;

  for (const line of body.split(/\r?\n/)) {
    if (/^```/.test(line)) inFence = !inFence;
    if (!inFence && /^---\s*$/.test(line)) {
      slides.push(current.join('\n'));
      current = [];
      continue;
    }
    current.push(line);
  }
  slides.push(current.join('\n'));

  return slides.map((s) => s.trim()).filter(Boolean);
}

function buildSlide(raw, context) {
  let className = '';
  let body = raw.replace(/<!--\s*_class:\s*([\w-]+)\s*-->/g, (_, name) => {
    className = name;
    return '';
  });

  const notes = [];
  body = body.replace(/<!--([\s\S]*?)-->/g, (_, note) => {
    notes.push(note.trim());
    return '';
  });

  const attributes = [];
  if (className) attributes.push(`class="${className}"`);
  const background = FULL_BLEED[className];
  if (background) attributes.push(`data-background-color="${background()}"`);

  const parts = [];
  if (context.header && className !== 'portada') {
    parts.push(`<div class="deck-header">${escapeHtml(context.header)}</div>`);
  }

  // El contenido va envuelto porque reveal le escribe `display: block` inline a
  // la sección que está en pantalla: cualquier `display: flex` en el CSS pierde
  // contra eso, y centrar los separadores de bloque a fuerza de !important
  // sería pelearse con la librería en vez de usarla.
  const rendered = [
    renderMarkdown(body.trim(), (source, alt) => {
      const size = alt.match(/^w:(\d+)$/);
      const file = join(context.dir, source);
      if (!existsSync(file)) throw new Error(`la diapositiva apunta a ${source} y no existe`);
      return inlineDiagram(file, size ? size[1] : null, basename(source, '.svg'));
    }),
  ].join('\n');

  parts.push(`<div class="slide-body">\n${rendered}\n</div>`);

  if (notes.length) {
    parts.push(`<aside class="notes">\n${renderMarkdown(notes.join('\n\n'))}\n</aside>`);
  }

  const open = attributes.length ? `<section ${attributes.join(' ')}>` : '<section>';
  return `${open}\n${parts.join('\n')}\n</section>`;
}

// ══ EL DECK ════════════════════════════════════════════════════════════════

/**
 * Sangra el HTML de una diapositiva para poder leerlo, salteando lo que esté
 * adentro de un `<pre>`.
 *
 * Con `white-space: pre`, cuatro espacios de prolijidad son cuatro espacios de
 * sangría falsa en el snippet proyectado — y la sangría del código es una de
 * las cosas que la clase mira.
 */
function indentSlide(html, pad) {
  let inCode = false;
  return html
    .split('\n')
    .map((line) => {
      const opens = line.includes('<pre>');
      const closes = line.includes('</pre>');
      const indented = inCode ? line : pad + line;
      if (opens && !closes) inCode = true;
      else if (closes) inCode = false;
      return indented;
    })
    .join('\n');
}

function buildDeck(dir) {
  const source = readFileSync(join(dir, 'slides.md'), 'utf8');
  const { meta, body } = splitFrontMatter(source);
  const toTheme = relative(dir, join(ROOT, 'theme')).replace(/\\/g, '/');

  const context = { dir, header: meta.header ?? '' };
  const slides = splitSlides(body).map((raw) => buildSlide(raw, context));

  const title = meta.header || basename(dir);
  const paginate = String(meta.paginate) === 'true';

  return `<!doctype html>
<html lang="es">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>${escapeHtml(title)}</title>

<!-- Generado por scripts/gen-reveal.mjs desde slides.md — no editar a mano.
     Se regenera con: node scripts/gen-reveal.mjs ${relative(ROOT, dir).replace(/\\/g, '/')} -->

<link rel="stylesheet" href="${toTheme}/reveal/reset.css">
<link rel="stylesheet" href="${toTheme}/reveal/reveal.css">
<link rel="stylesheet" href="${toTheme}/reveal-neobrutal.css">
</head>
<body>

<div class="reveal">
  <div class="slides">
${slides.map((slide) => indentSlide(slide, '    ')).join('\n\n')}
  </div>
</div>

<script src="${toTheme}/reveal/reveal.js"></script>
<script src="${toTheme}/reveal/plugin/highlight/highlight.js"></script>
<script src="${toTheme}/reveal/plugin/notes/notes.js"></script>
<script src="${toTheme}/reveal/plugin/zoom/zoom.js"></script>
<script>
  Reveal.initialize({
    width: 1280,
    height: 720,
    margin: 0,
    // El contenido arranca arriba: las diapositivas a sangre se centran solas
    // con flex en el tema, y el resto quiere columna de lectura.
    center: false,
    hash: true,
    slideNumber: ${paginate ? "'c/t'" : 'false'},
    transition: 'none',
    // Sin ondas ni desvanecidos: el movimiento comunica estado o no está.
    backgroundTransition: 'none',
    plugins: [RevealHighlight, RevealNotes, RevealZoom],
  });
</script>

</body>
</html>
`;
}

// ══ CLI ════════════════════════════════════════════════════════════════════

const sessionDirs = () => {
  if (explicit.length) return explicit.map((p) => join(ROOT, p));
  const base = join(ROOT, 'sesiones');
  return readdirSync(base)
    .map((entry) => join(base, entry))
    .filter((dir) => existsSync(join(dir, 'slides.md')));
};

const stale = [];
let written = 0;

for (const dir of sessionDirs()) {
  const out = join(dir, OUTPUT_NAME);
  if (CHECK && !existsSync(out)) continue;

  let deck;
  try {
    deck = buildDeck(dir);
  } catch (err) {
    stale.push(`${relative(ROOT, dir)} — ${err.message}`);
    continue;
  }

  const before = existsSync(out) ? readFileSync(out, 'utf8') : null;
  if (before === deck) continue;

  if (CHECK) stale.push(`${relative(ROOT, out)} — desfasado respecto de slides.md`);
  else {
    writeFileSync(out, deck, 'utf8');
    written++;
    console.log(`  ${before === null ? 'escrito     ' : 'actualizado '} ${relative(ROOT, out)}`);
  }
}

if (stale.length) {
  console.error('\n  ✗ Decks de reveal:');
  for (const s of stale) console.error(`    · ${s}`);
  console.error(CHECK ? '\n    Correr: node scripts/gen-reveal.mjs\n' : '');
  process.exit(1);
}

if (!CHECK) console.log(`  ${written} deck(s) al día.`);
