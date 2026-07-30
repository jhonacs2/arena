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

| Archivo | Bloque | Quién lo lee | Qué es |
|---|---|---|---|
| `guion.md` | todos | instructor | **El documento maestro.** Todo lo demás sale de acá |
| `slides.md` | 0:12+ | ambos | Marp. El guión de cada diapositiva va en sus speaker notes |
| `diagramas/*.excalidraw` | 0:12 | ambos | El concepto dibujado. El `.svg` de al lado es generado |
| `mision-profe.md` | 0:20 | instructor | **La secuencia de tecleo del live coding**, en `demo/` |
| `mision-estudiante-1.md` | 0:35 | alumno | Ejercicio individual, en `lab/starter` |
| `mision-estudiante-2.md` | 1:25 | alumno | Ejercicio en parejas, en `project/frontend/starter` |
| `correccion.md` | 1:45 | ambos | **De vacío a funcionando, paso a paso** |
| `conceptos.md` | después | alumno | **El apunte de referencia de la sesión** |
| `predice-y-ejecuta/` | 1:10 | ambos | Snippets rotos + `respuestas.md` |
| `wayground.csv` | 0:05 de la **siguiente** | instructor | Preguntas sobre **esta** sesión |
| `exit-ticket.md` · `tarea.md` | 1:55 | ambos | |

### Misión profe y misión estudiante

El nombre dice **quién teclea**. Antes se llamaban `mision-1` y `mision-2`, y eso escondía que había tres cosas distintas mezcladas: lo que demuestra el instructor y los dos ejercicios de los alumnos.

- **`mision-profe.md`** — lo que escribís vos en vivo a las 0:20, en orden, con las frases y **las roturas deliberadas marcadas**. Va en el segundo monitor: el guión lleva la clase, esto lleva el teclado. Cierra con un **orden de sacrificio** por si te quedás sin tiempo.
- **`mision-estudiante-N.md`** — lo que codean ellos.

### Los ejercicios se escriben como ejercicio de libro, no como nota de profe

Un enunciado que dice «lo mismo que acabás de ver, pero lo escribís vos» o «trabarse es parte del ejercicio» **no es un enunciado**: es manejo de clase, y va en `guion.md`. El alumno abre el archivo en su casa, sin el contexto del aula, y necesita un problema planteado.

La estructura, siempre la misma:

| | Sección |
|---|---|
| 1 | **Enunciado** — el escenario en prosa, dos o tres frases |
| 2 | **Estado inicial** — cómo arrancar y qué va a ver |
| 3 | **Datos** — literales exactos, o de dónde salen |
| 4 | **Requisitos** — numerados, precisos, cada uno verificable |
| 5 | **Resultado esperado** — dibujado o transcripto, concreto |
| 6 | **Restricciones** — qué no se puede usar |
| 7 | **Autoevaluación** — checklist |
| 8 | **Pistas** — en `<details>`, escalonadas |
| 9 | **Extensión** — para quien termina antes |

Sin segunda persona de instructor, sin «acordate de», sin explicar por qué el ejercicio es así.

### `conceptos.md` — el apunte

> **Las clases son en vivo y no quedan grabadas. La memoria es frágil.**

Sin este archivo, el alumno se sienta a hacer la tarea con lo que recuerda de dos horas de clase. Con él, tiene cada concepto con su definición y **los ejemplos exactos que se corrieron en clase** — no ejemplos nuevos: los mismos, para que reconozca lo que vio.

Lleva, en este orden: el problema que motivó el tema, las definiciones, los ejemplos de clase, **los errores de la sesión con su mensaje literal y su arreglo**, un glosario, y qué **no** se vio todavía.

**Y hay que anunciarlo en voz alta a las 1:55.** Un apunte que nadie sabe que existe no existe.

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
3. **Lab:** crear `lab/{solution,starter}/src/app/sessions/sNN/`, poner `available: true` en `sessions.ts` **y** sumar la ruta en `app.routes.ts` — ver [`lab/CLAUDE.md`](../lab/CLAUDE.md).
4. **Hipódromo:** el slice en `solution/`. En `starter/` **no existe**: se construye en clase, y `correccion.md` es el camino.
5. `slides.md`, `mision-profe.md`, los dos ejercicios, `conceptos.md`, predicciones, `wayground.csv`, exit ticket, tarea.
6. `node scripts/verify.mjs` y commit `feat(sNN): …`.

### El live coding no se da en `solution/`

> **Nunca escribas en el guión «borrá tal carpeta antes de empezar».**

Es pedirle al instructor que mutile su propia solución de referencia cinco minutos antes de dar la clase, con el riesgo de olvidarse de restaurarla. Y es innecesario: el lienzo correcto —el proyecto en el estado justo anterior a esta sesión— **es exactamente `starter/`**.

```bash
node scripts/prep-demo.mjs      # starter/ → demo/, en un segundo
```

Tres copias, tres dueños, ninguna se pisa:

| | Quién | Se toca en clase |
|---|---|---|
| `solution/` | la referencia del instructor | **nunca** |
| `starter/` | lo que reciben los alumnos | es lo que se publica |
| `demo/` | copia descartable para el live coding | sí, y se tira |

`demo/` está en `.gitignore` y se regenera cuantas veces haga falta, así que el bloque se puede ensayar tres veces arrancando limpio.

**Si un tema se necesita antes de su sesión, se da hecho y se nombra.** `@for` es de S3 pero S1 necesita una lista: viene escrito y el enunciado aclara en qué sesión se ve a fondo. La regla completa está en `docs/curriculum.md`.

> **Las respuestas de «predice y ejecuta» se verifican en el navegador antes de escribirlas.** En S1 la respuesta que parecía obvia era la contraria: `class="a {{ b }}"` y `[class.x]` **no compiten, se combinan**. Escribir de memoria es cómo se le enseña algo falso a treinta personas a la vez.

---

## Los diagramas se autoran en Excalidraw

La fuente es el `.excalidraw`; el `.svg` de al lado lo genera `scripts/gen-diagram-svg.mjs` y lo consumen las diapositivas. **El `.svg` no se edita a mano** — `verify.mjs` falla si quedó desfasado.

Se edita en <https://excalidraw.com>: «Abrir» → el archivo → dibujar → «Guardar en...» encima. El motivo de que sea Excalidraw y no un SVG escrito a mano es que el bloque de las 0:12 se da **dibujando**: poder agregar una flecha en vivo mientras alguien pregunta vale más que un SVG prolijo que nadie puede tocar en clase.

Dos reglas al autorar:

- **`roughness: 0` y `fillStyle: "solid"`.** El trazo a mano alzada y el rayado los dibuja rough.js y el generador no los reproduce; con estos dos valores, lo que se ve en Excalidraw es lo que sale en el SVG. El generador avisa si aparece un rayado.
- **Declarar el token cuando el color es ambiguo.** En claro, `text`, `border`, `shadow` y `accent-ink` son **todos** ink-900; en oscuro se separan. El generador deduce el rol —trazo de forma es `border`, trazo de texto es `text`— pero un **relleno** con ese color puede ser una sombra o un bloque de tinta, y ahí hay que decirlo:

```json
"customData": { "backgroundToken": "shadow", "strokeToken": "shadow" }
```

Excalidraw preserva `customData` al guardar, así que sobrevive a que lo edites en la app. Sin esto la sombra dura sale **blanca** en modo oscuro y el texto sobre el acento queda claro sobre claro.

## El tema de las diapositivas tiene tres trampas

Marp no es CSS normal. Las tres ya costaron tiempo una vez:

1. **Nunca escribas `*/` dentro de un comentario del tema.** Un glob como `sesiones/s01-*/slides.md` cierra el comentario ahí y corrompe todo lo que sigue. El deck sale **sin un solo color** y nada avisa.
2. **Los tokens se declaran en `section`, no en `:root`** — ni en `:root, section`: Marp reescribe la lista de selectores y el navegador la descarta entera. `gen-tokens-css.mjs` ya lo resuelve; no lo «arregles».
3. **`::after` es de Marp** (la paginación) y `<header>` también. Lo que pongas ahí hereda su `opacity` y se ve apagado.

Las diapositivas se exportan **siempre en claro**: si dependieran de `prefers-color-scheme`, el mismo archivo saldría distinto según la máquina.

```bash
npx @marp-team/marp-cli --theme-set theme/ --html --allow-local-files sesiones/sNN-*/slides.md
```
