#!/usr/bin/env node
/**
 * Verifica que los IDENTIFICADORES del código estén en inglés.
 *
 *   node scripts/check-language.mjs           informe completo
 *   node scripts/check-language.mjs --quiet    solo falla o calla
 *
 * La regla del proyecto es una sola y no tiene excepciones:
 *
 *   **El texto que ve el usuario, en español. El código, en inglés.**
 *
 * «Código» incluye todo lo que se nombra: variables, funciones, tipos,
 * propiedades, clases CSS, custom properties, archivos y nombres de test.
 * No incluye el contenido de los strings, los comentarios ni el texto del
 * HTML — eso es lo que lee el alumno y va en español.
 *
 * Por qué importa en un curso: el alumno va a leer código en inglés toda su
 * vida profesional. Un proyecto con `carrera.seleccionada` le enseña un
 * dialecto que no existe fuera de esta clase.
 */

import { readFileSync, readdirSync, statSync, existsSync } from 'node:fs';
import { dirname, join, relative, basename } from 'node:path';
import { fileURLToPath } from 'node:url';

const ROOT = join(dirname(fileURLToPath(import.meta.url)), '..');
const quiet = process.argv.includes('--quiet');

/**
 * Palabras en español que no pueden aparecer en un identificador.
 *
 * Es una lista curada, no un diccionario: atrapa lo que este proyecto usa y
 * lo que un hispanohablante escribe sin pensar. Si aparece una nueva, se
 * agrega acá — es más barato que revisar a ojo.
 */
// Ojo: no van acá las palabras que se escriben igual en inglés — error, color,
// total, menu, panel. Marcarlas haría que el checker griete por nada y que
// se lo empiece a ignorar.
const SPANISH = [
  // dominio del hipódromo
  'carrera', 'caballo', 'apuesta', 'apostar', 'jinete', 'seda', 'cuota', 'saldo',
  'ganador', 'podio', 'parrilla', 'corredor', 'dorsal', 'largada', 'pista', 'monto',
  'pago', 'favorito', 'competidor', 'jornada', 'hipodromo',
  // dominio del lab
  'cafe', 'pedido', 'comanda', 'cliente', 'mostrador', 'producto', 'precio',
  'cantidad', 'disponible', 'agotado', 'origen',
  // curso
  'sesion', 'sesiones', 'leccion', 'clase', 'alumno', 'ejercicio', 'mision',
  // genéricas de programación
  'boton', 'tarjeta', 'cabecera', 'encabezado', 'pie', 'barra', 'pantalla',
  'ventana', 'campo', 'formulario', 'etiqueta', 'titulo', 'subtitulo', 'mensaje',
  'aviso', 'exito', 'cargando', 'vacio', 'lista', 'tabla', 'fila',
  'columna', 'contenedor', 'contenido', 'seccion', 'pagina', 'inicio', 'detalle',
  'resumen', 'estado', 'tema', 'claro', 'oscuro', 'colores', 'fuente',
  'tamano', 'ancho', 'alto', 'borde', 'sombra', 'fondo', 'texto', 'numero',
  'nombre', 'valor', 'dato', 'datos', 'usuario', 'entrada', 'salida', 'nota',
  'muestra', 'escala', 'familia', 'tablero', 'programa', 'simulador', 'rotulo',
  'hueso', 'contorno', 'cuello', 'marca', 'saltar', 'apilado', 'chico', 'grande',
  'vista', 'bloque', 'texto',
  // verbos
  'agregar', 'quitar', 'borrar', 'eliminar', 'crear', 'guardar', 'cargar',
  'buscar', 'filtrar', 'ordenar', 'mostrar', 'ocultar', 'abrir', 'cerrar',
  'seleccionar', 'elegir', 'cambiar', 'alternar', 'limpiar', 'reiniciar',
  'aplicar', 'ciclar', 'formatear', 'calcular', 'obtener', 'poner', 'accionar',
  'cobrar', 'liquidar', 'validar', 'enviar', 'recibir', 'actualizar',
  // adjetivos y modificadores frecuentes
  'primario', 'secundario', 'terciario', 'fantasma', 'peligro', 'activo',
  'activa', 'abierta', 'cerrada', 'viva', 'terminada', 'pendiente', 'nuevo',
  'primero', 'ultimo', 'siguiente', 'anterior', 'todos', 'todas', 'ninguno',
];

const spanishRe = new RegExp(`(?:^|[^a-z])(${SPANISH.join('|')})(?:s|es)?(?:$|[^a-z])`, 'i');

/** Devuelve la palabra española encontrada, o null. */
function spanishWordIn(identifier) {
  // camelCase / PascalCase / kebab / snake → palabras sueltas
  const words = identifier
    .replace(/([a-z0-9])([A-Z])/g, '$1 $2')
    .replace(/[-_.]/g, ' ')
    .toLowerCase()
    .split(/\s+/)
    .filter(Boolean);

  for (const word of words) {
    const bare = word.replace(/(es|s)$/, '');
    if (SPANISH.includes(word) || SPANISH.includes(bare)) return word;
  }
  return null;
}

/** Vacía comentarios y strings conservando los saltos de línea. */
function stripNoise(text) {
  const blank = (m) => m.replace(/[^\n]/g, ' ');
  return text
    .replace(/\/\*[\s\S]*?\*\//g, blank)
    .replace(/<!--[\s\S]*?-->/g, blank)
    .replace(/(^|[^:])\/\/[^\n]*/g, (m, before) => before + ' '.repeat(m.length - before.length))
    .replace(/`(?:\\.|[^`\\])*`/g, blank)
    .replace(/'(?:\\.|[^'\\\n])*'/g, blank)
    .replace(/"(?:\\.|[^"\\\n])*"/g, blank);
}

function walk(dir, exts) {
  const out = [];
  if (!existsSync(dir)) return out;
  for (const entry of readdirSync(dir)) {
    if (['node_modules', 'dist', '.angular', '.git'].includes(entry)) continue;
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) out.push(...walk(full, exts));
    else if (exts.some((e) => entry.endsWith(e))) out.push(full);
  }
  return out;
}

const SCOPES = [
  'project/frontend/solution/src',
  'project/frontend/starter/src',
  'lab/solution/src',
  'lab/starter/src',
  'project/backend/internal',
  'scripts',
];

const findings = [];
const add = (file, line, kind, name, word) =>
  findings.push({ file: relative(ROOT, file), line, kind, name, word });

for (const scope of SCOPES) {
  const base = join(ROOT, scope);

  // ── Nombres de archivo ──────────────────────────────────────────────────
  for (const file of walk(base, ['.ts', '.css', '.html', '.go', '.mjs'])) {
    const word = spanishWordIn(basename(file).replace(/\.[^.]+$/, ''));
    if (word) add(file, 0, 'archivo', basename(file), word);
  }

  // ── Identificadores en TypeScript, Go y JavaScript ──────────────────────
  for (const file of walk(base, ['.ts', '.go', '.mjs'])) {
    const lines = stripNoise(readFileSync(file, 'utf8')).split('\n');

    lines.forEach((line, i) => {
      // declaraciones: const X, let X, function X, class X, interface X,
      // type X, propiedad X:, método X(
      const patterns = [
        /\b(?:const|let|var|function|class|interface|type|enum|func)\s+([A-Za-z_$][\w$]*)/g,
        /(?:^|[{;,])\s*(?:readonly\s+|protected\s+|private\s+|public\s+|static\s+)*([A-Za-z_$][\w$]*)\s*[:(=]/g,
      ];

      for (const re of patterns) {
        for (const match of line.matchAll(re)) {
          const name = match[1];
          if (!name || name.length < 3) continue;
          const word = spanishWordIn(name);
          if (word) add(file, i + 1, 'identificador', name, word);
        }
      }
    });
  }

  // ── Clases CSS y custom properties ──────────────────────────────────────
  for (const file of walk(base, ['.css'])) {
    const lines = stripNoise(readFileSync(file, 'utf8')).split('\n');
    lines.forEach((line, i) => {
      for (const match of line.matchAll(/\.([a-z][\w-]*)/gi)) {
        const word = spanishWordIn(match[1]);
        if (word) add(file, i + 1, 'clase CSS', '.' + match[1], word);
      }
      for (const match of line.matchAll(/--([a-z][\w-]*)\s*:/gi)) {
        const word = spanishWordIn(match[1]);
        if (word) add(file, i + 1, 'custom property', '--' + match[1], word);
      }
    });
  }

  // ── Clases usadas en el HTML ────────────────────────────────────────────
  for (const file of walk(base, ['.html'])) {
    const lines = readFileSync(file, 'utf8').split('\n');
    lines.forEach((line, i) => {
      for (const match of line.matchAll(/\bclass(?:\.([\w-]+))?\s*=\s*"([^"]*)"/g)) {
        const candidates = match[1] ? [match[1]] : match[2].split(/\s+/);
        for (const candidate of candidates) {
          if (!candidate || candidate.includes('{{')) continue;
          const word = spanishWordIn(candidate);
          if (word) add(file, i + 1, 'clase en HTML', candidate, word);
        }
      }
    });
  }
}

// ── Informe ───────────────────────────────────────────────────────────────
const unique = new Map();
for (const f of findings) unique.set(`${f.file}:${f.line}:${f.name}`, f);
const list = [...unique.values()];

if (!quiet) {
  if (list.length === 0) {
    console.log('\n  ✓ Todos los identificadores están en inglés.\n');
  } else {
    console.log(`\n  Identificadores en español — ${list.length} hallazgos\n`);
    const byFile = new Map();
    for (const f of list) {
      if (!byFile.has(f.file)) byFile.set(f.file, []);
      byFile.get(f.file).push(f);
    }
    for (const [file, items] of [...byFile].sort((a, b) => b[1].length - a[1].length)) {
      console.log(`  ${file}  (${items.length})`);
      const names = [...new Set(items.map((i) => `${i.name}`))];
      console.log(`      ${names.slice(0, 14).join(', ')}${names.length > 14 ? ` …y ${names.length - 14} más` : ''}`);
    }
    console.log('');
  }
}

if (list.length > 0) {
  if (quiet) {
    console.error(`  ${list.length} identificadores en español. Correr node scripts/check-language.mjs para verlos.`);
  }
  process.exit(1);
}
