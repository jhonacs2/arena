# Plan de construcción

Orden de trabajo del módulo. Una fase por vez; cada una termina con `node scripts/verify.mjs` en verde y un commit.

| | Fase | Estado |
|---|---|---|
| **0** | Contrato y esqueleto | ✅ **terminada** |
| **1** | Backend Go, completo y congelado | ✅ **terminada** |
| **2** | Baseline visual del hipódromo | ✅ **terminada** |
| **3** | S1 · Primer componente | ✅ **terminada** — falta darla |
| **3.5** | Revisión del formato de los materiales | ✅ **terminada** |
| **3.75** | S0 · TypeScript, y S11 repartida | ✅ **terminada** — falta darla |
| **4** | S2 · Anatomía de un componente | ✅ **terminada** — falta darla |
| **5–12** | Una sesión por fase, S3 … S10 | ⬜ siguiente |
| **13** | Publicación y deploy | ⬜ |

---

## Por qué este orden

El riesgo que mata este tipo de proyecto no es el orden de construcción, es la **deriva de contrato**: el frontend asume `startsAt`, el backend devuelve `start_time`, y S7 se convierte en una clase de depuración en lugar de una de `HttpClient`.

La Fase 0 congela el contrato como artefacto ejecutable **antes** de escribir cualquiera de los dos lados, y ambos asertan contra los mismos archivos JSON. Con eso, el orden de lo que viene después deja de ser crítico — pero conviene igual hacer el backend primero, porque una vez terminado no se toca más y libera toda la atención para el frontend, que es donde vive la pedagogía.

---

## Fase 0 — Contrato y esqueleto ✅

- `git init`, árbol, `.gitignore`, `.editorconfig`, `.nvmrc`
- `docs/contract/` completo: `openapi.yaml` (13 endpoints), `ws-events.md`, `error-codes.md` (18 códigos), seed (12 usuarios · 8 carreras · 54 caballos · 34 apuestas), 12 samples, fixture de 462 eventos
- `docs/design/`: `tokens.json` canónico, `tokens.md`, `IMAGES.md`, hoja de muestra de las 54 sedas
- `docs/curriculum.md` — el mapa de las 11 sesiones
- `sesiones/_plantilla/` — los 9 archivos de cada sesión
- `theme/marp-neobrutal.css`
- `scripts/`: `verify.mjs`, `check-contrast.mjs`, `gen-tokens-css.mjs`, `gen-race-ticks.mjs`, `gen-silks-specimen.mjs`, `sync-contract.mjs`
- `CLAUDE.md` v2

`node scripts/verify.mjs` corre hoy **20 verificaciones** sobre contrato, diseño, backend y —cuando exista— el código Angular.

## Fase 1 — Backend Go, completo y congelado ✅

Un binario, **una sola dependencia externa** (`github.com/coder/websocket`), sin base de datos, con el dataset embebido.

- Los 13 endpoints de `openapi.yaml`, con el sobre de error uniforme y los 18 códigos del catálogo
- Hub WebSocket multiplexado por sala. `race.finished` se arma **por destinatario**: difundir el mismo objeto filtraría lo que cobró cada uno
- Simulador a 10 Hz que implementa `race-simulation.md`. El test golden reproduce el fixture tick por tick, así que la app se ve igual contra el mock y contra el servidor
- Calendario que larga carreras solo: la primera a los 30 s, después el ciclo del dataset cada 2 h
- JWT HS256 escrito a mano, PBKDF2 de stdlib, refresh de un solo uso, verificación por correo con un emisor de desarrollo que imprime el enlace en consola
- Estado en memoria con copia atómica en disco; `RESET=1` vuelve al dataset limpio
- Dockerfile distroless con binario estático

Probado de punta a punta contra el servidor real: **26 verificaciones**, incluidos 22 eventos de cuenta regresiva, 419 ticks y la liquidación con el saldo consistente entre socket y REST.

**Desde acá el backend está congelado:** `CLAUDE.md` §12.

## Fase 2 — Baseline visual del hipódromo ✅

El piso sobre el que se para S1. Angular 18.2.14 con versiones exactas, standalone en todo, `strict` más `noUncheckedIndexedAccess`.

- `styles.css` con `@layer reset, tokens, base, utilities` y el bloque de tokens inyectado desde `tokens.json`
- Las tres tipografías variables auto-hospedadas en woff2 — 356 KB en el repo, cero CDN
- `silkFromId()` portado a TypeScript, `<app-silk>`, y un test que cruza el port contra la implementación de referencia en JS para los 54 caballos
- `shared/ui/`: button, skeleton, empty-state, logo. **Sin badge ni race-card**: son el ejercicio de S2
- `core/models/` (contraparte del contrato), `core/mocks/` (generado desde el seed), `core/theme/`
- `layout/shell` con salto al contenido, navegación y el interruptor de tema
- `features/sistema`: la muestra del sistema de diseño, que es la prueba visible de que todo esto funciona
- `starter/` creado con el mismo baseline — a partir de S1 los dos divergen

Build de producción en **75 kB** de transferencia inicial. Probado en navegador, en claro y oscuro, y medido de 320 a 768 px sin desbordamiento horizontal.

Un defecto real que encontró el cruce JS↔TS: había **dos mezcladores de hash distintos** —el generador de sedas usaba una mezcla de dos pasos y el simulador de carreras una de tres—. Unificados; de paso, las coincidencias de seda entre carreras bajaron de 1 a 0.

## Fase 3 — S1 · Primer componente ✅

- **Lab creado** (`lab/solution` y `lab/starter`): una app con una ruta por sesión, dominio propio —una cafetería— para que el concepto se practique sin el ruido del hipódromo
- **Lab `/s01`**: los cuatro bindings en un solo componente standalone, con tests que los verifican desde afuera
- **Hipódromo**: `race-list` con las 8 carreras, panel de parrilla con sedas y simulador de apuesta al favorito
- **`sesiones/s00-typescript/`**: material del tema 0 y su quiz, que alimenta el bloque 0:05 de S1
- **`sesiones/s01-primer-componente/`**: los 9 archivos — guión de 12 bloques, 24 diapositivas Marp, diagrama, dos misiones, tres ejercicios de predicción con respuestas, quiz, exit ticket y tarea

Dos decisiones que quedaron escritas como regla y no como caso particular:

- **El starter funciona a medias, no está vacío** (`CLAUDE.md` §1). Además de ser mejor pedagogía, es lo único que compila: `noUnusedLocals` no deja un starter con imports sin usar.
- **Cuando un tema se necesita antes de su sesión, se da hecho y se nombra** (`docs/curriculum.md`). `@for` es de S3 pero S1 necesita una lista.

Tres cosas que se arreglaron porque se verificaron en vez de suponerse:

- El deck de Marp salía **sin un solo color**: un `sesiones/s01-*/` dentro de un comentario del tema CSS cerraba el comentario en el `*/` del glob y corrompía todo lo que seguía
- La franja de la portada usaba `::after`, que **Marp ya usa para la paginación** — y era un `linear-gradient`, prohibido por §10. Ahora es un borde sólido
- La respuesta del primer «predice y ejecuta» era **la contraria** de lo que había escrito: `class` y `[class.x]` no compiten, se combinan

## Fase 3.5 — Revisión del formato de los materiales ✅

Cuatro cambios que salieron de leer S1 completa. Se aplicaron a S1 y a `_plantilla/`, así que S2 los hereda.

- **Los diagramas se autoran en Excalidraw.** El `.excalidraw` es la fuente y el `.svg` lo genera `gen-diagram-svg.mjs`. El motivo es que el bloque de las 0:12 se da **dibujando**: agregar una flecha en vivo mientras alguien pregunta vale más que un SVG prolijo intocable.
- **El live coding no se da en `solution/`.** `prep-demo.mjs` copia `starter/` → `demo/` en un segundo. El guión decía «borrá `sessions/s01` antes de empezar» — mutilar la solución de referencia cinco minutos antes de dar la clase, y además innecesario: el lienzo correcto ya era `starter/`.
- **`mision-profe` y `mision-estudiante-N`.** El nombre dice quién teclea. Los ejercicios de alumno se reescribieron como **ejercicio de libro** —enunciado, datos, requisitos numerados, resultado esperado dibujado, restricciones, autoevaluación, pistas escalonadas—, sin voz de instructor: se leen en casa, sin el contexto del aula.
- **`conceptos.md`, el apunte.** Las clases son en vivo y no quedan grabadas. Cada sesión deja cada concepto con su definición, **los ejemplos exactos que se corrieron en clase** y los errores con su mensaje literal. Se anuncia en voz alta a las 1:55: un apunte que nadie sabe que existe no existe.

Tres defectos reales que aparecieron al mirar en vez de suponer:

- El generador de diagramas deducía el token del color, y en claro `text`, `border`, `shadow` y `accent-ink` son **los cuatro** ink-900. En oscuro se separan: la sombra dura salía **blanca** y el texto de la pastilla quedaba claro sobre claro. Se resolvió con `customData` explícito, que Excalidraw preserva al guardar, más un aviso del generador para el único caso genuinamente ambiguo (un relleno).
- El diagrama mostraba `cafe`, `cliente`, `cantidad`, `agregar()` — **código en español**, contra §3.
- `'Marcar available'` en el guión **y en `lab/solution`**: texto de usuario en inglés, resto del barrido de renombrado.

Y una decisión de nombre: la carpeta se llama `demo/` y no `clase/` porque `check-language.mjs` marcó `prep-clase.mjs` con razón —los nombres de archivo son código—. `demo/` además dice *qué es* en vez de *cuándo se usa*.

## Fase 3.75 — S0 · TypeScript, y S11 repartida ✅

TypeScript era el **tema 0 asíncrono**. Pasó a ser la **primera clase en vivo**: el vocabulario de tipos sostiene las once sesiones que siguen, y leerlo solo no es lo mismo que ver aparecer y desaparecer un error en pantalla.

El módulo son 11 clases, así que una tuvo que salir. Se repartió **S11**: los 30 minutos de NgModules legacy se suman a S10, y el build de producción, el deploy y el code review cruzado pasan al cierre asíncrono. **La numeración de S1 a S10 no cambió** — renumerar habría roto las rutas `/sNN` del lab, los `TODO(SN)` del starter y el tag `s01`, sin ganar nada pedagógico.

- **Lab `/s00`**: el menú de Café Compilado en `sessions/s00/menu.ts` — TypeScript puro, sin una línea de Angular. Uniones de literales, opcionales, narrowing, `readonly`, genéricos y `Omit`
- **Hipódromo**: `core/models/race.model.ts` del starter pasó a su versión **floja**, con seis `TODO(S0)`. Apretarlo hasta que diga lo que dice `openapi.yaml` es la Misión 2
- **`sesiones/s00-typescript/`**: los 12 archivos — guión de 12 bloques, 33 diapositivas Marp, diagrama, dos misiones, tres ejercicios de predicción, quiz, exit ticket, tarea y un README con el setup previo

Dos excepciones que quedaron escritas como regla, no como caso particular:

- **En `/s00` la ruta y el componente vienen dados también en `starter/`** (`lab/CLAUDE.md`). Crear la ruta es el ejercicio de todas las sesiones menos esta: las rutas son S9 y los componentes son S1, y en la clase 0 no se vio ninguno de los dos
- **El bloque 0:05 de S0 no es un Wayground sino un diagnóstico en vivo**, porque no hay sesión anterior de la cual preguntar

Y una decisión de método que se pagó sola: **los ocho mensajes de error del material están copiados de la salida de `tsc`, no escritos de memoria.** Dos de las tres respuestas del bloque «predice y ejecuta» son contraintuitivas —`readonly lines: OrderLine[]` **sí** deja hacer `push`, y `JSON.parse(t) as Race` **compila sin una advertencia**—, y son justamente las que no se pueden improvisar en clase.

## Fase 4 — S2 · Anatomía de un componente ✅

De una pantalla que funciona a dos piezas que se pueden mover. La sesión no agrega **ni una función**: la pantalla se ve igual al empezar y al terminar, y eso es lo que hay que decir en voz alta en el code review.

- **Lab `/s02`**: `<app-coffee-card>` con las cuatro puertas —`input()`, `input.required()`, `model()`, `output()`—, dos `<ng-content>` y los tres ganchos del ciclo de vida. El starter trae la pantalla **entera adentro de un componente**: hoy no se construye, se parte
- **Hipódromo**: `<app-badge>` en `shared/ui/` —el texto entra por `ng-content`, el tono por `input()`— y `<app-race-card>` en `features/races/`. `race-list` se queda con lo suyo: preparar datos y decidir cuál está abierta
- **`sesiones/s02-anatomia-componente/`**: los 12 archivos, diagrama incluido

Dos defectos reales que aparecieron por verificar en vez de suponer:

- **`<app-button>` emitía las clases `boton--*` y el CSS define `.button--*`.** Ninguna variante ni tamaño se estaba aplicando desde Fase 2, en los dos proyectos. Además de un bug, era un §3: los nombres de clase son código
- **Escribí en el guión que `ngOnChanges` no corre cuando el hijo cambia su propio `model()`.** Es al revés y lo dijo el navegador: el valor sube al padre y **vuelve a bajar** por el mismo binding, así que corre. Quedó convertido en la pregunta que se hace en vivo, que es mejor material que la afirmación equivocada

Y una respuesta de «predice y ejecuta» que la intuición falla: **dos `<ng-content>` iguales compilan sin una queja y el contenido aparece una sola vez** — el contenido proyectado se *mueve*, no se copia, y un nodo del DOM no puede estar en dos lugares.

## Fases 5–12 — Una sesión por fase

Cada una entrega: slice de la solución · slice del starter con sus `// TODO(Sn)` · ruta `/sNN` del lab en ambas versiones · los archivos de `sesiones/sNN-*/` · commit `feat(sNN): …` + tag en el repo del alumno.

**Recomendación:** dar S0 y S1, ajustar la plantilla del guión con lo que salga, y recién ahí seguir con S2. El guión de 12 bloques es una hipótesis hasta que se corre con alumnos reales.

## Fase 13 — Publicación

`publish-student-repo.sh` · tags por sesión · deploy del frontend · README del alumno con setup en 5 minutos.
