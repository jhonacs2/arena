# CLAUDE.md — Los materiales de clase

> Complementa el [`CLAUDE.md` de la raíz](../CLAUDE.md).

Una carpeta por sesión, todas con los mismos archivos. La plantilla está en `_plantilla/`; el mapa de qué se enseña en cada una, en [`docs/curriculum.md`](../docs/curriculum.md).

---

## La regla que manda sobre todas

> **Nadie usa algo que no sabe qué es.**

Enseñar «poné `{{ }}` para interpolar» no es enseñar: es dictar sintaxis. El alumno la copia, le funciona, y no puede explicar por qué — así que la primera vez que falla, no tiene con qué razonar.

**Todo concepto se presenta en este orden, sin saltear pasos:**

| | | Para qué |
|---|---|---|
| 1 | **El problema** | Que sientan la incomodidad antes de darles la solución |
| 2 | **Qué es** | Una definición en una frase, en castellano, sin jerga sin explicar |
| 3 | **Cómo se escribe** | Recién acá aparece la sintaxis |
| 4 | **Cómo se rompe** | El error concreto que van a cometer, y qué dice |

Si un bloque del guión llega al paso 3 sin haber pasado por el 1 y el 2, está mal escrito. Volvé y arreglalo.

**Ningún término se usa antes de definirse.** Cada `guion.md` abre con un **glosario de la sesión**: cada palabra nueva, en el orden en que aparece, con su definición de una línea. Si una palabra está en el glosario de una sesión anterior, se puede usar; si no está en ninguno, hay que definirla o sacarla.

---

## El guión es un teleprompter, no un resumen

El instructor lo lee **mientras da la clase**. Si tiene que interpretarlo, ya se perdió.

**Sí:**

- Frases literales para decir, entre comillas y en su renglón.
- Qué hay en pantalla en cada momento: *«diapositiva 7»*, *«editor cerrado»*, *«navegador al lado del código»*.
- El código exacto que se escribe en el live coding, en el orden en que se escribe.
- Las preguntas que van a hacer, con su respuesta preparada.
- Marcas de tiempo cada pocos minutos dentro de los bloques largos, no solo al inicio.

**No:**

- Viñetas telegráficas que hay que expandir en vivo (*«explicar binding»* no es un guión).
- Suponer que el instructor recuerda por qué se decidió algo. Si importa, está escrito ahí.
- Mandar a leer otro archivo en medio de un bloque. Lo que se necesita a las 0:27 está a las 0:27.

**La prueba:** leelo de corrido, cronómetro en mano, sin abrir nada más. Si en algún punto tenés que parar a pensar qué sigue, ese punto está mal escrito.

---

## Los archivos de cada sesión

`sesiones/README.md` los detalla. En resumen:

| Archivo | Bloque | Qué es |
|---|---|---|
| `guion.md` | todos | **El documento maestro.** Todo lo demás sale de acá |
| `slides.md` | 0:12+ | Marp. El guión de cada diapositiva va en sus speaker notes |
| `diagramas/*.svg` | 0:12 | El concepto dibujado, con el editor cerrado |
| `mision-1.md` | 0:35 | Individual, en `lab/starter` |
| `mision-2.md` | 1:25 | En parejas, en `project/frontend/starter` |
| `correccion.md` | 1:45 | **De vacío a funcionando, paso a paso** |
| `predice-y-ejecuta/` | 1:10 | Snippets rotos + `respuestas.md` |
| `wayground.csv` | 0:05 de la **siguiente** | Preguntas sobre **esta** sesión |
| `exit-ticket.md` · `tarea.md` | 1:55 | |

### `correccion.md`

Es el archivo que hace que el starter pueda estar vacío. Lleva **de la pantalla en blanco al resultado**, y sirve para tres cosas: guiar el bloque de las 1:45, rescatar a quien se trabó, y que el alumno se autocorrija después.

Cada paso lleva **qué se escribe, dónde, y por qué ahí**. Un paso que solo muestra código copiado no sirve: el «por qué» es lo único que no pueden sacar del `solution/`.

### El desfase del Wayground

> `sesiones/sNN/wayground.csv` tiene preguntas **sobre la sesión NN** y se corre al empezar la **sesión NN+1**.

El bloque 0:05 de S1 usa `s00-typescript/wayground.csv`. El quiz de S11 va a la evaluación asíncrona de cierre.

---

## Cómo se produce una sesión

1. Copiar `_plantilla/` a `sesiones/sNN-<slug>/`. Slug corto y en español: `s01-primer-componente`.
2. Escribir **primero el glosario y después `guion.md`**. Si el glosario no cierra, la sesión tiene dos temas y hay que partirla.
3. **Lab:** crear `lab/{solution,starter}/src/app/sesiones/sNN/`, poner `disponible: true` en `sesiones.ts` **y** sumar la ruta en `app.routes.ts` — ver [`lab/CLAUDE.md`](../lab/CLAUDE.md).
4. **Hipódromo:** el slice en `solution/`. En `starter/` **no existe**: se construye en clase, y `correccion.md` es el camino.
5. `slides.md`, misiones, predicciones, `wayground.csv`, exit ticket, tarea.
6. `node scripts/verify.mjs` y commit `feat(sNN): …`.

**Si un tema se necesita antes de su sesión, se da hecho y se nombra.** `@for` es de S3 pero S1 necesita una lista: viene escrito y el enunciado aclara en qué sesión se ve a fondo. La regla completa está en `docs/curriculum.md`.

> **Las respuestas de «predice y ejecuta» se verifican en el navegador antes de escribirlas.** En S1 la respuesta que parecía obvia era la contraria: `class="a {{ b }}"` y `[class.x]` **no compiten, se combinan**. Escribir de memoria es cómo se le enseña algo falso a treinta personas a la vez.

---

## El tema de las diapositivas tiene tres trampas

Marp no es CSS normal. Las tres ya costaron tiempo una vez:

1. **Nunca escribas `*/` dentro de un comentario del tema.** Un glob como `sesiones/s01-*/slides.md` cierra el comentario ahí y corrompe todo lo que sigue. El deck sale **sin un solo color** y nada avisa.
2. **Los tokens se declaran en `section`, no en `:root`** — ni en `:root, section`: Marp reescribe la lista de selectores y el navegador la descarta entera. `gen-tokens-css.mjs` ya lo resuelve; no lo «arregles».
3. **`::after` es de Marp** (la paginación) y `<header>` también. Lo que pongas ahí hereda su `opacity` y se ve apagado.

Las diapositivas se exportan **siempre en claro**: si dependieran de `prefers-color-scheme`, el mismo archivo saldría distinto según la máquina.

```bash
npx @marp-team/marp-cli --theme-set theme/ --html --allow-local-files sesiones/sNN-*/slides.md
```
