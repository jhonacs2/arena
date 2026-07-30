# CLAUDE.md — Hipódromo · Módulo Angular · Talento DH 8va

> **Regla cero: el material didáctico es Angular 18. No Angular 19, 20 ni 21.**
>
> Vale para `project/` y `lab/` — todo lo que el alumno lee o escribe. **`arena/` es la excepción y es Angular 22**: es una app que los alumnos *usan*, no material que leen, y tiene su propio [`arena/CLAUDE.md`](arena/CLAUDE.md).

Este archivo tiene solo lo que vale **en todo el repo**. El resto está repartido por carpeta y se carga cuando trabajás ahí:

| Si tocás… | Leé también |
|---|---|
| `project/frontend/` | [`project/frontend/CLAUDE.md`](project/frontend/CLAUDE.md) — convenciones Angular, estructura, mocks por etapa |
| `lab/` | [`lab/CLAUDE.md`](lab/CLAUDE.md) — el mundo del lab y cómo sumarle una sesión |
| `sesiones/` | [`sesiones/CLAUDE.md`](sesiones/CLAUDE.md) — cómo se escribe una clase que se pueda dar |
| `docs/design/` o cualquier `.css` | [`docs/design/CLAUDE.md`](docs/design/CLAUDE.md) — la paleta y las reglas visuales |
| `docs/contract/` | [`docs/contract/CLAUDE.md`](docs/contract/CLAUDE.md) — regla contrato-primero |
| `project/backend/` | [`project/backend/CLAUDE.md`](project/backend/CLAUDE.md) — está congelado |
| `arena/` | [`arena/CLAUDE.md`](arena/CLAUDE.md) — **otro producto: Angular 22, Go, Supabase** |

**Estado actual y qué sigue: [`docs/plan.md`](docs/plan.md).**

---

## 1. Qué es esto

Aplicación educativa de **simulación de apuestas de carreras de caballos**. Saldo virtual, cero dinero real. Es el proyecto ancla de un módulo de Angular de **11 sesiones de 2 horas** para desarrolladores en formación.

Se producen **cuatro artefactos** desde este mismo spec. No los mezcles:

| Artefacto | Carpeta | Qué es |
|---|---|---|
| **Solución de referencia** | `project/frontend/solution/` | App completa. Solo la ve el instructor. |
| **Starter del alumno** | `project/frontend/starter/` | Lo que recibe el alumno. Ver la regla de abajo. |
| **Lab** | `lab/solution/` y `lab/starter/` | App aparte, una ruta por sesión. El concepto **aislado**. |
| **Materiales de clase** | `sesiones/sNN-*/` | Guión, slides, ejercicios, apunte, corrección. |

Y una cuarta copia que **no es un artefacto**, es una herramienta: `demo/`.

> El live coding **no se da en `solution/`**. Se da en `demo/`, una copia descartable de `starter/` que genera `scripts/prep-demo.mjs` y que está en `.gitignore`.

Escribir en el guión «borrá tal carpeta antes de empezar» es pedirle al instructor que mutile su solución de referencia cinco minutos antes de dar la clase. Y es innecesario: el lienzo correcto para el live coding —el proyecto en el estado justo anterior a esta sesión— **ya es `starter/`**. Detalle en [`sesiones/CLAUDE.md`](sesiones/CLAUDE.md).

### La regla del starter

> **El alumno tiene que ver aparecer lo que construye.** Si la pantalla ya estaba hecha, no aprendió: le apareció mágicamente y no sabe qué puso él.

El starter arranca **con lo mínimo para que el concepto de la sesión sea lo que se construye**. En S1 eso es casi nada: la paleta, la configuración y una pantalla vacía. De ahí en adelante, cada sesión hereda lo que construyó la anterior y agrega lo suyo desde cero.

Dos consecuencias:

- **Nunca borres un archivo de configuración** (`angular.json`, `tsconfig`, `styles.css`, fuentes). Eso no es el ejercicio y romperlo cuesta media clase.
- **Sí se borra el código que es el ejercicio.** Si el componente de la sesión es lo que se construye, en el starter no existe: lo crean ellos, con el CLI.

Cada sesión tiene su **`correccion.md`**: el paso a paso de vacío a funcionando, para guiar en vivo y para que el alumno se autocorrija.

### Y su `conceptos.md`

> **Las clases son en vivo y no quedan grabadas. La memoria es frágil.**

Sin apunte, el alumno se sienta a hacer la tarea con lo que recuerda de dos horas. Cada sesión deja un **`conceptos.md`** con cada concepto, su definición, **los ejemplos exactos que se corrieron en clase** y los errores con su mensaje literal. Se anuncia en voz alta al cerrar: un apunte que nadie sabe que existe no existe.

### El lab y el hipódromo son dos cosas distintas

- **0:20 live coding · 0:35 Misión 1 · 1:10 predice y ejecuta** → `lab/`
- **1:25 Misión 2** → `project/frontend/starter/`

El concepto se aprende aislado y se aplica en el proyecto.

---

## 2. Mapa del workspace

```
a/                              ← workspace del instructor. Tiene TODO.
├── CLAUDE.md                   ← este archivo
├── docs/
│   ├── contract/               ← ⭐ fuente de verdad del backend
│   ├── design/                 ← tokens.json · tokens.md · IMAGES.md · assets/
│   ├── curriculum.md           ← sesión → concepto → lab → misión ancla
│   └── plan.md                 ← estado y qué sigue
├── project/{backend, frontend/{solution,starter,demo*}}
├── lab/{solution,starter,demo*}
├── sesiones/                   ← _plantilla/ + s00 … s10
├── scripts/                    ← verify.mjs y generadores
└── theme/marp-neobrutal.css

* demo/ no está en git: la genera scripts/prep-demo.mjs y es descartable.
```

**Este workspace no se publica.** `scripts/publish-student-repo.sh` genera el repo del alumno con solo `project/backend` + `project/frontend/starter` + `lab/starter`, y se taggea `s01`, `s02`… al cerrar cada clase.

---

## 3. El idioma

> **El texto que ve el usuario, en español. El código, en inglés.** Siempre, en todo el repo.

«Código» es todo lo que se nombra: variables, funciones, tipos, propiedades, clases CSS, custom properties, nombres de archivo y de test. **No** es el contenido de los strings, ni los comentarios, ni el texto del HTML, ni las URLs — eso es lo que lee el alumno.

```ts
// bien
protected readonly races = RACES.map((race) => …);
<span class="race__name">{{ view.race.name }}</span>

// mal
protected readonly carreras = RACES.map((carrera) => …);
<span class="carrera__nombre">{{ vista.carrera.name }}</span>
```

Por qué importa en un curso: el alumno va a leer código en inglés toda su vida profesional. Un proyecto con `carrera.seleccionada` le enseña un dialecto que no existe fuera de esta clase.

Lo verifica `scripts/check-language.mjs`, que corre dentro de `verify.mjs`. Si aparece una palabra en español que la lista no conoce, se agrega ahí — sale más barato que revisar a ojo.

**Las URLs no son código.** `/carreras` y `/sistema` son navegación que el usuario lee, y van en español. Lo que está adentro del `.ts` que las declara, no.

---

## 4. Stack y versiones — bloqueadas

```
Angular      18.2.x   (exacto, sin ^)
TypeScript   5.5.x
RxJS         7.8.x
Node         ^20.11 || ^22     (ver .nvmrc — el equipo está en 22.22.3)
Go           1.26.x
```

Instalá con `npm install --save-exact`. `package.json` no debe contener `^` ni `~` en `@angular/*`.

---

## 5. APIs prohibidas (no existen en Angular 18)

**La sección más importante del repo.** Aparecen mucho en material reciente y **no compilan** en 18.

> Esta tabla rige en `project/` y `lab/`. **En `arena/` no** — ahí corre Angular 22 y estas APIs existen. `verify.mjs` solo escanea los cuatro proyectos del material; Arena tiene su propio verificador.

| API | Llega en | Qué usar en 18 |
|---|---|---|
| `resource()`, `rxResource()` | v19 | Servicio + `signal()` + `HttpClient` manual |
| `httpResource()` | v20 | `HttpClient` + `toSignal()` |
| `linkedSignal()` | v19 | `computed()` + un `signal()` de override |
| Signal Forms (`form()`, `Control`, schemas) | v20/21 | **Reactive Forms** (`FormBuilder`, `FormGroup`) |
| `afterRenderEffect()` | v19 | `afterNextRender()` / `effect()` |
| `provideZonelessChangeDetection()` | v19 | En 18 se llama `provideExperimentalZonelessChangeDetection()` y **no lo usamos**: corremos con zone.js. |
| `standalone: true` implícito | v19 | En 18 **se escribe explícitamente** en cada `@Component`, `@Directive` y `@Pipe`. |

Sí disponibles en 18, usalas: `signal()`, `computed()`, `effect()`, `input()`, `input.required()`, `output()`, `model()`, `@if/@for/@switch`, `@defer`, `toSignal()`, `toObservable()`, `takeUntilDestroyed()`, guards e interceptores funcionales, `inject()`.

`@let` y las signal queries (`viewChild()`, `contentChild()`) existen en 18 pero en developer preview — **no las uses**.

---

## 6. Verificación mecánica (obligatoria)

Una lista en prosa se degrada a lo largo de una sesión larga. **Corré `node scripts/verify.mjs` después de cada feature.**

```bash
node scripts/verify.mjs            # todo
node scripts/verify.mjs --fast     # sin builds, mientras iterás
node scripts/verify.mjs contrato   # un grupo: contrato | diseño | backend | código
```

Verifica el contrato, el diseño (contraste AA, sedas, tokens), el backend (gofmt, vet, tests, build) y el frontend (APIs prohibidas en `.ts` **y `.html`**, `standalone` y `OnPush`, `any` y `console.log`, fuentes auto-hospedadas, `tsc --noEmit`, build y tests de navegador).

Saltea con gracia lo que todavía no existe. Si falla, **arreglá antes de seguir**. No expliques por qué falló y continúes.

**Los comentarios no cuentan.** El verificador los vacía antes de buscar: escribir «acá no usamos `NgModule`» tiene que ser posible.

### Nada generado se edita a mano

| Genera | Script |
|---|---|
| `core/mocks/*.ts` | `gen-mocks.mjs` |
| bloque de tokens en cada `styles.css` | `gen-tokens-css.mjs` |
| `race-ticks.jsonl` | `gen-race-ticks.mjs` |
| hoja de sedas y `silks.golden.ts` | `gen-silks-specimen.mjs` |
| `public/contract/`, `internal/seed/data/` | `sync-contract.mjs` |
| `public/fonts/` y `src/fonts.css` | `fetch-fonts.mjs` |
| `sesiones/**/diagramas/*.svg` | `gen-diagram-svg.mjs` — la fuente es el `.excalidraw` |

`verify.mjs` los corre en modo `--check` y falla si algo quedó desfasado.

---

## 7. Definición de terminado

| | Desde |
|---|---|
| 1. `node scripts/verify.mjs` pasa | siempre |
| 2. No hay `any`, ni `console.log`, ni imports sin usar | siempre |
| 3. Todos los componentes son `standalone: true` y `OnPush` | siempre |
| 4. Se probó **en el navegador**, no solo compilando | siempre |
| 5. Se probó **en claro y en oscuro** | siempre |
| 6. Recorrido por teclado con foco visible | siempre |
| 7. El starter arranca sin el código de la sesión, y su `correccion.md` lleva de vacío a funcionando | siempre |
| 8. La ruta `/sNN` del lab existe en `solution/`. En `starter/` **no**: crearla es el ejercicio | siempre |
| 9. Los materiales de `sesiones/sNN-*/` están completos | siempre |
| 10. `conceptos.md` alcanza para hacer la tarea **sin la memoria de la clase** | siempre |
| 11. Los ejercicios de alumno se leen como ejercicio de libro, sin voz de instructor | siempre |
| 12. La vista maneja los tres estados: cargando, vacío, error | **S7** |
| 13. Se probó contra el mock **y** contra el backend real, y se ve igual | **S7** |

Los puntos 12 y 13 **no aplican antes de S7**: hasta ahí los datos son constantes tipadas y no hay nada que cargar ni que pueda fallar. Decirlo en voz alta en el code review de las sesiones tempranas es mejor que exigir algo imposible y acostumbrar a saltear la lista.

---

## 8. Cómo trabajar

- Avanzá **una sesión a la vez**. Terminá S(N) completa —solución, starter, lab, materiales, verificación— antes de tocar S(N+1).
- Antes de escribir un archivo nuevo, decí en qué carpeta va y por qué.
- Si el spec es ambiguo o falta un endpoint, **preguntá**. No inventes rutas.
- Si necesitás una API que §5 prohíbe, **pará y decilo**. No busques un rodeo silencioso.
- Commits pequeños con prefijo de sesión: `feat(s04): pipe de cuotas`.
- **Verificá antes de afirmar.** Si escribís que algo pasa, corrélo y miralo. En S1 la respuesta «obvia» de un ejercicio de predicción resultó ser la contraria.
