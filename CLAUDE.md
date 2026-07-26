# CLAUDE.md — Hipódromo (proyecto ancla, Módulo Angular · Talento DH 8va Versión)

> Documento de contexto para Claude Code. Léelo completo antes de escribir código.
> **Regla cero: este proyecto es Angular 18. No Angular 19, 20 ni 21.**

---

## 1. Qué es esto y para quién

Aplicación educativa de **simulación de apuestas de carreras de caballos**. No maneja dinero real: el saldo es virtual y se otorga al registrarse. Es el proyecto ancla de un módulo de Angular de **11 sesiones de 2 horas** para desarrolladores en formación.

Se producen **cuatro artefactos** desde este mismo spec. No los mezcles:

| Artefacto | Carpeta | Contenido |
|---|---|---|
| **Solución de referencia** | `project/frontend/solution/` | App completa y funcional. Solo la ve el instructor. |
| **Starter del alumno** | `project/frontend/starter/` | Misma estructura, mismos archivos, misma configuración — pero los cuerpos que corresponden a un ejercicio están vacíos y marcados con `// TODO(S3): …`. Compila y levanta sin errores desde el minuto cero. |
| **Lab** | `lab/solution/` y `lab/starter/` | App aparte, una ruta por sesión (`/s01` … `/s11`). Acá se enseña el concepto **aislado**, sin el ruido del dominio de apuestas. |
| **Materiales de clase** | `sesiones/sNN-*/` | Guión, slides, misiones, quiz, exit ticket. |

Cuando generes el starter, **nunca borres un archivo**: vacía su implementación y deja el TODO con el número de sesión. Un archivo faltante rompe el árbol; un método vacío no.

### El starter funciona a medias, no está vacío

Un starter en blanco no enseña: enseña una pantalla en blanco. El starter tiene que **levantar mostrando algo y andar mal de una forma concreta**, y el ejercicio es arreglar eso.

```ts
// TODO(S1): esto siempre devuelve 0 porque la cuota está clavada en 0.
protected get pagoPotencial(): number {
  return potentialPayout(this.monto, 0);
}
```

Tres razones, y la tercera es la que no se negocia:

1. Es más parecido al trabajo real que a un ejercicio de escuela.
2. El alumno ve el efecto de su cambio de inmediato, contra algo que ya se movía.
3. **`noUnusedLocals` no deja compilar un starter vacío.** Si `favourite` y `potentialPayout` están importados y nadie los usa, `tsc` falla — y el starter tiene que compilar desde el minuto cero. Un andamiaje que ya los usa resuelve las dos cosas a la vez.

### El lab y el hipódromo son dos cosas distintas

El guión de clase reparte los bloques entre los dos proyectos. Respetalo:

- **0:20 live coding · 0:35 Misión 1 · 1:10 predice y ejecuta** → `lab/`
- **1:25 Misión 2** → `project/frontend/starter/`

El concepto se aprende aislado y se aplica en el proyecto. Es lo que evita que la primera clase de `input()` se pelee con el modelo de dominio de las apuestas.

---

## 2. Mapa del workspace

```
a/                              ← workspace del instructor. Tiene TODO.
├── CLAUDE.md
├── docs/
│   ├── contract/               ← ⭐ FUENTE DE VERDAD (§7)
│   ├── design/                 ← tokens.json · tokens.md · IMAGES.md · assets/
│   └── curriculum.md           ← sesión → concepto → lab → misión ancla
├── project/                    ← lo que se publica al alumno
│   ├── backend/                ← monolito Go
│   └── frontend/{solution,starter}
├── lab/{solution,starter}
├── sesiones/                   ← _plantilla/ + s00 … s11
├── scripts/                    ← verify.mjs y generadores
└── theme/marp-neobrutal.css
```

**Repo del alumno.** Este workspace **no** se publica. `scripts/publish-student-repo.sh` genera un repo con solo `project/backend` + `project/frontend/starter` + `lab/starter`, y se taggea `s01`, `s02`… al cerrar cada clase. El que se atrasa hace `git checkout s03`. La solución nunca sale de acá.

---

## 3. Stack y versiones — bloqueadas

```
Angular      18.2.x   (exacto, sin ^)
TypeScript   5.5.x
RxJS         7.8.x
Node         ^20.11 || ^22     (ver .nvmrc — el equipo está en 22.22.3)
Go           1.26.x
```

Instala con `npm install --save-exact`. `package.json` no debe contener `^` ni `~` en `@angular/*`.

---

## 4. APIs prohibidas (no existen en Angular 18)

Esta es la sección más importante del documento. Estas APIs aparecen mucho en material reciente y **no compilan** en 18:

| API | Llega en | Qué usar en 18 |
|---|---|---|
| `resource()`, `rxResource()` | v19 | Servicio + `signal()` + `HttpClient` manual |
| `httpResource()` | v20 | `HttpClient` + `toSignal()` |
| `linkedSignal()` | v19 | `computed()` + un `signal()` de override |
| Signal Forms (`form()`, `Control`, schemas) | v20/21 | **Reactive Forms** (`FormBuilder`, `FormGroup`) |
| `afterRenderEffect()` | v19 | `afterNextRender()` / `effect()` |
| `provideZonelessChangeDetection()` | v19 | No usar. En 18 se llama `provideExperimentalZonelessChangeDetection()` y **no lo usamos**: el proyecto corre con zone.js. |
| `standalone: true` implícito | v19 | En 18 **hay que escribirlo explícitamente** en cada `@Component`, `@Directive` y `@Pipe`. |

Sí disponibles en 18, úsalas: `signal()`, `computed()`, `effect()`, `input()`, `input.required()`, `output()`, `model()`, `@if/@for/@switch`, `@defer`, `toSignal()`, `toObservable()`, `takeUntilDestroyed()`, guards e interceptores funcionales, `inject()`.

`@let` y las signal queries (`viewChild()`, `contentChild()`) existen en 18 pero en developer preview — **no las uses**, confunden al alumno.

### Verificación mecánica (obligatoria)

Una lista en prosa se degrada a lo largo de una sesión larga. **Ejecutá `node scripts/verify.mjs` después de cada feature.**

```bash
node scripts/verify.mjs            # todo
node scripts/verify.mjs --fast     # sin builds, mientras iterás
node scripts/verify.mjs contrato   # un solo grupo: contrato | diseño | código
```

Verifica el contrato (coherencia del seed, integridad referencial de las apuestas, leaderboard golden, fixture de la carrera, copias sincronizadas, catálogo de errores Go ↔ markdown), el diseño (contraste AA, colisiones de sedas, tokens al día), el backend (gofmt, vet, tests, build) y el frontend (APIs prohibidas en `.ts` **y `.html`**, `standalone` y `OnPush`, `any` y `console.log`, fuentes auto-hospedadas, `tsc --noEmit`, build de producción y tests de navegador).

Saltea con gracia lo que todavía no existe. `scripts/verify.sh` es un envoltorio: la lógica está en el `.mjs` porque los alumnos están en Windows, macOS y Linux.

**Los comentarios no cuentan.** El verificador los vacía antes de buscar: escribir "acá no usamos `NgModule`" tiene que ser posible, y explicar por qué algo *no* está es parte del material.

### Nada generado se edita a mano

| Genera | Script | Desde |
|---|---|---|
| `core/mocks/*.ts` | `gen-mocks.mjs` | `docs/contract/seed/` |
| bloque de tokens en `styles.css` | `gen-tokens-css.mjs` | `docs/design/tokens.json` |
| `race-ticks.jsonl` | `gen-race-ticks.mjs` | el simulador |
| hoja de sedas y `silks.golden.ts` | `gen-silks-specimen.mjs` | el seed |
| `public/contract/`, `internal/seed/data/` | `sync-contract.mjs` | `docs/contract/` |
| `public/fonts/` y `src/fonts.css` | `fetch-fonts.mjs` | Google Fonts, una sola vez |

`verify.mjs` corre todos en modo `--check` y falla si algo quedó desfasado.

Si falla, **arreglá antes de seguir**. No expliques por qué falló y continúes.

**Única excepción a `NgModule`:** la ruta `/s11` del lab, donde se enseña NgModules como contexto legacy. `verify.mjs` ya la exime.

---

## 5. Convenciones de código

- **Standalone en todo.** Cero `NgModule` (salvo `lab/**/s11`). Bootstrap con `bootstrapApplication` + `app.config.ts`.
- **`inject()` siempre**, nunca inyección por constructor. Es lo que se enseña en clase.
- **`ChangeDetectionStrategy.OnPush` en todos los componentes**, sin excepción.
- **Estado con signals**, no con `BehaviorSubject`. RxJS se reserva para HTTP, WebSocket y eventos del DOM.
- **Inmutabilidad estricta.** Nunca `array.push()` sobre estado; siempre `signal.update(v => [...v, x])`. Esto se evalúa en clase.
- **Nada de NgRx.** El estado vive en servicios con signals privados de escritura y expuestos de solo lectura:
  ```ts
  private readonly _races = signal<Race[]>([]);
  readonly races = this._races.asReadonly();
  ```
- Interceptores y guards **funcionales** (`HttpInterceptorFn`, `CanActivateFn`).
- `provideRouter(routes, withComponentInputBinding())` para que los params de ruta lleguen como `input()`.
- Nada de `any`. `strict: true` en `tsconfig.json`, más `noUncheckedIndexedAccess`, `noUnusedLocals` y `noUnusedParameters`.
- Nombres de archivo en kebab-case, componentes con sufijo real: `race-card.component.ts`.
- **Todo el texto de la UI en español. Los identificadores del código, en inglés.** Sin excepciones en ninguno de los dos sentidos.

---

## 6. Estructura de la app Angular

```
src/app/
├── core/                       # singletons, se provee una sola vez
│   ├── auth/
│   │   ├── auth.service.ts         # estado de sesión (signals), login/logout/refresh
│   │   ├── auth.interceptor.ts     # inyecta Bearer, refresca en 401
│   │   ├── auth.guard.ts           # canActivate funcional
│   │   └── verified.guard.ts       # bloquea si el correo no está verificado
│   ├── http/
│   │   ├── api.config.ts           # InjectionToken con la baseUrl
│   │   ├── error.interceptor.ts    # normaliza errores del backend
│   │   └── mock-backend.interceptor.ts  # responde desde el seed (§8)
│   ├── ws/
│   │   ├── socket.service.ts       # conexión única, reconexión, multiplexado por sala
│   │   ├── mock-socket.service.ts  # reproduce race-ticks.jsonl (§8)
│   │   └── ws-events.model.ts      # tipos de los eventos del servidor
│   ├── theme/theme.service.ts      # claro / oscuro / sistema
│   ├── mocks/                      # GENERADO por gen-mocks.mjs desde el seed
│   └── models/                     # Race, Horse, Bet, User, Page, ApiError
├── shared/
│   ├── ui/                         # silk, button, skeleton, empty-state, logo
│   ├── pipes/                      # money.pipe, odds.pipe
│   └── directives/
├── layout/
│   ├── shell.component.ts          # header + router-outlet
│   └── balance-widget.component.ts # saldo en vivo, escucha balance.updated
├── features/
│   ├── sistema/        muestra del sistema de diseño (no es del producto)
│   ├── auth/           login | register | verify-email | resend-verification
│   ├── races/          race-list | race-detail | race-live
│   ├── bets/           my-bets
│   └── leaderboard/
├── app.config.ts
├── app.routes.ts
└── app.component.ts
```

**Regla de dependencias:** `features/` puede importar de `core/` y `shared/`. `shared/` no importa de `features/` ni de `core/`. `core/` no importa de `features/`. Si necesitás romperla, no la rompas: mové el código.

**Lo que NO va en `shared/ui/`:** `<app-badge>` y `<app-race-card>` son el ejercicio de S2. Si están hechos de antemano, esa sesión se queda sin práctica. Vale para cualquier pieza que aparezca en la columna "qué construye el alumno" de §9: se deja para su sesión.

---

## 7. El contrato — regla contrato-primero

**`docs/contract/` es la fuente de verdad.** El backend Go y el frontend Angular se escriben contra ella, nunca uno contra el otro.

> **Si un campo o un endpoint no está en `docs/contract/openapi.yaml`, no existe.**
> Cuando haga falta algo nuevo se agrega **primero** ahí, y recién después se implementa en los dos lados. Si el spec es ambiguo o falta un endpoint, **preguntá**. No inventes rutas.

| Archivo | Qué es |
|---|---|
| `docs/contract/openapi.yaml` | Los 13 endpoints con todos los esquemas. Normativo. |
| `docs/contract/ws-events.md` | El contrato del WebSocket. OpenAPI no cubre sockets. |
| `docs/contract/error-codes.md` | Catálogo **cerrado** de `error.code`. El frontend hace `switch` sobre el código, nunca sobre el mensaje. |
| `docs/contract/seed/` | El dataset. Lo cargan **los dos** lados. |
| `docs/contract/samples/` | Una respuesta canónica por endpoint. Golden test de ambos lados. |
| `docs/contract/fixtures/race-ticks.jsonl` | Grabación de una carrera en vivo a 10 Hz. |
| `docs/contract/README.md` | Regla de rebase de fechas y las decisiones de forma que tomé. **Leelo.** |

### Resumen operativo

Base `/api/v1`. Autenticación por `Authorization: Bearer <accessToken>`. Error uniforme:

```json
{ "error": { "code": "INVALID_CREDENTIALS", "message": "…", "details": {} } }
```

| Método | Ruta | Nota |
|---|---|---|
| POST | `/auth/register` | `201`, dispara correo de verificación (Resend) |
| POST | `/auth/verify` · `/auth/resend-verification` | el correo lleva a `{FRONT_URL}/verificar?token=…` |
| POST | `/auth/login` · `/auth/refresh` · `/auth/logout` | refresh token de **un solo uso** |
| GET | `/me` | |
| GET | `/races?status=&page=&size=` | paginado `{items, page, size, total}`; el listado **sí** trae `horses[]` |
| GET | `/races/:id` · `/races/:id/results` | `payouts` trae solo las apuestas del usuario autenticado |
| POST | `/bets` | rechaza si la carrera arrancó, si no hay saldo o si el monto sale de `[10, 5000]` |
| GET | `/bets/me?page=` · `/leaderboard?period=daily\|all` | |

WebSocket: `wss://{HOST}/ws?token={accessToken}` — el token va por query string porque el navegador no permite headers en el handshake. Una sola conexión, multiplexada por sala. `race.tick` llega a ~10 Hz.

**La simulación de la carrera es autoridad del servidor.** El front solo pinta `positions`. No interpoles físicas ni calcules ganadores en el cliente.

⚠️ **Detección de cambios y WebSocket.** Los eventos del socket llegan fuera de la zona de Angular y con `OnPush` la UI puede no repintar aunque el signal cambie. El `SocketService` escribe en signals y el componente los lee. **Si la carrera no se anima en el navegador, el problema es la zona, no el binding.** Es el contenido central de S10.

---

## 8. Datos falsos, por etapa

El frontend arranca antes que el backend y **las primeras seis sesiones no lo necesitan**. Pero los datos falsos no se inventan: salen del mismo `docs/contract/seed/` que carga Go.

| Sesiones | Qué usa | Por qué |
|---|---|---|
| S1–S4 | `core/mocks/` — consts tipadas | Datos hardcodeados, como pide el temario |
| S5–S6 | `RaceStore` con signals; `of(fixture).pipe(delay(400))` | Async real sin HTTP; habilita debounce y `toSignal()` |
| S7+ | `HttpClient` real + token `API_URL` | Apunta al Go real **o** a `provideMockBackend()`. Cambiar de uno a otro es una línea en `environment.ts` |
| S10 | `MockSocketService` reproduce `race-ticks.jsonl` a 10 Hz | **Seguro de clase**: si el deploy se cae, la carrera igual corre |

El interceptor de mock responde desde el seed con los mismos códigos de error del catálogo. Un componente escrito contra el mock funciona contra el backend real sin tocar una línea.

---

## 9. Sesiones y materiales

Son **11 sesiones de 2 horas**. El tema 0 (TypeScript) va 100% asíncrono, antes de S1. El mapa completo está en `docs/curriculum.md`; la plantilla de materiales en `sesiones/_plantilla/`.

| Sesión | Tema | Qué construye el alumno en el hipódromo |
|---|---|---|
| **S1** | Filosofía e historia de Angular · Web Components · CLI · primer standalone · binding uni y bidireccional | Layout base + listado de carreras con datos hardcodeados |
| **S2** | Anatomía de un componente · segmentación de templates · `input()`/`output()`/`model()` · ciclo de vida · `ng-content` | `<app-race-card>` reutilizable; `<app-badge>` con proyección |
| **S3** | Signals (`signal`, `computed`, `set/update`) · control flow `@if/@for/@switch` · inmutabilidad | Filtro por estado y búsqueda; `@for` con `track`; ordenar sin mutar |
| **S4** | Directivas de atributo y estructurales · directivas custom · pipes built-in y custom | `money.pipe`, `odds.pipe`, directiva de resaltado de favorito |
| **S5** | DI a fondo · servicios · `inject()` · jerarquía de inyectores · `InjectionToken` | `RaceStore` y `BetStore` con signals; `API_URL` como token |
| **S6** | Reactividad · promesas vs observables · hot/cold · `map`, `filter`, `switchMap` · `takeUntilDestroyed` | Buscador con debounce; interop `toSignal()` |
| **S7** | `HttpClient` · GET/POST/PUT/DELETE · interceptores funcionales · `catchError` | Conectar al backend real; `auth.interceptor`; `error.interceptor`; carga/vacío/error |
| **S8** | Reactive Forms · `FormBuilder` · validadores custom · errores de formulario | Login, registro y formulario de apuesta con validación de saldo |
| **S9** | Routing operativo · `provideRouter` · params · guards funcionales · lazy `loadComponent` | App multipantalla; `authGuard` y `verifiedGuard` |
| **S10** | WebSockets · zona y detección de cambios · `OnPush` · `@defer` | `race-live`: carrera en vivo, saldo reactivo, leaderboard diferido |
| **S11** | NgModules como contexto legacy (30 min) · build de producción · deploy · code review en vivo | App desplegada + revisión cruzada entre equipos |
| **Async** | TypeScript (tema 0) · verificación de correo · testing introductorio | Pantalla `/verificar`, reenvío de correo |

### El guión de 12 bloques

Toda sesión tiene la misma forma. Está en `sesiones/_plantilla/guion.md` y no se improvisa:

```
0:00  Pregunta de apertura (chat, sin juicio)                 5 min
0:05  Wayground de la sesión ANTERIOR                         7 min
0:12  Concepto en diagrama, SIN editor abierto                8 min
0:20  Live coding narrado — ellos NO copian, miran           15 min   → lab
0:35  Misión 1: lo mismo, pero ellos. Vos en silencio        15 min   → lab
0:50  2 alumnos comparten pantalla. Preguntás, no corregís   10 min
1:00  Descanso                                               10 min
1:10  "Predice y ejecuta": código roto a propósito           15 min   → lab
1:25  Misión 2 sobre el proyecto ancla, en parejas           20 min   → hipódromo
1:45  Code review en vivo de una solución de alumno          10 min
1:55  Exit ticket 3 preguntas + tarea asíncrona               5 min
```

> **Desfase del Wayground:** `sesiones/sNN/wayground.csv` contiene preguntas **sobre la sesión NN** y se corre al empezar la **sesión NN+1**. El bloque 0:05 de S1 usa `s00-typescript/wayground.csv`. El quiz de S11 va a la evaluación asíncrona de cierre.

S11 rompe el molde a propósito (30 min de NgModules + deploy + review cruzado). Va documentado en su propio `guion.md`.

---

## 10. UI

**Sin Tailwind. Sin Sass. Sin librería de componentes.** CSS nativo moderno con tokens. Razón pedagógica: el temario incluye *View Encapsulation*, y con utilidades atómicas ese tema queda sin nada que enseñar.

El sistema completo — paleta, tipografía, sedas, movimiento — está en **`docs/design/tokens.md`**. Los valores canónicos en `docs/design/tokens.json`.

### Reglas duras

- CSS nativo. `nesting`, `@layer`, custom properties y `color-mix()` cubren lo que se usaba Sass.
- `styles.css` global **solo** para tokens, reset y capas: `@layer reset, tokens, base, components, utilities;`. Todo lo demás vive en el `.css` del componente, encapsulado. Los estilos de componente quedan fuera de capas a propósito, para ganar sin `!important`.
- **Los tokens no se escriben a mano en el CSS.** Se editan en `tokens.json` y se corre `node scripts/gen-tokens-css.mjs`, que inyecta el bloque entre los marcadores `/* @tokens:start */` y `/* @tokens:end */`.
- Color en `oklch()`, nunca hex. Modo oscuro **diseñado**: reasigna tokens semánticos, jamás primitivos. En oscuro el borde se invierte a tiza.
- **Contraste AA (4,5:1) en todo texto, verificado por `scripts/check-contrast.mjs`.** El neobrutalismo falla exactamente acá.
- Nada de neumorfismo ni glassmorphism.
- Bordes sólidos de 3px, sombras duras `4px 4px 0` sin blur, cero gradientes, **radio 0 sin excepciones**.
- Movimiento con propósito: la animación comunica estado (la carrera avanzando, el saldo cambiando), nunca decora. Respetar `prefers-reduced-motion`.

### Las sedas de jockey son el elemento firma

Cada caballo tiene su seda, **generada como SVG de forma determinística desde su `id`** — patrón de cuerpo × 2 colores × mangas, con rechazo si los colores no separan ΔL ≥ 0,22. Implementación de referencia en `scripts/gen-silks-specimen.mjs`; hoja de muestra en `docs/design/assets/silks-specimen.svg`.

**Regla dura: ningún texto se pinta sobre una seda.** El número del caballo va en su cuadrado aparte, tinta sobre tiza.

### Layout

Mobile-first real (la carrera en vivo funciona a 360px) · **container queries** en `<app-race-card>` para que responda a su contenedor y no al viewport · grid bento en el dashboard · `:has()` para estado condicional sin clases extra.

### Imágenes

Las sedas y los avatares son SVG generado: **no hay archivos de imagen para el 80% de la UI.** Lo que sí hace falta está especificado en **`docs/design/IMAGES.md`** — seis piezas, con dimensiones, formato y prompt. Las genera el instructor. Mientras no existan, la app usa los placeholders de `docs/design/assets/`: **nunca hay una imagen rota**.

Si necesitás una imagen que no está en `IMAGES.md`, **agregala ahí primero** con su especificación. No metas un `<img>` apuntando a un archivo que no existe.

---

## 11. Definición de terminado

Una feature no está lista hasta que:

1. `node scripts/verify.mjs` pasa.
2. No hay `any`, ni `console.log`, ni imports sin usar.
3. Todos los componentes son `standalone: true` y `OnPush`.
4. La vista maneja los tres estados: cargando (skeleton), vacío (con acción), error (con reintento).
5. Se probó en el navegador **contra el mock y contra el backend real**, y se ve igual en los dos.
6. Recorrido por teclado con foco visible.
7. El equivalente en `starter/` existe, compila, y tiene su `// TODO(Sn)`.
8. La ruta `/sNN` del lab existe, en solución y starter.
9. Los materiales de `sesiones/sNN-*/` están completos.

---

## 12. Cómo trabajar

- **Fases.** El plan vive en `docs/plan.md`. Orden: contrato → backend → baseline visual → una sesión por fase.
- Avanzá **una sesión a la vez**. Terminá S1 completa (solución + starter + lab + materiales + verificación) antes de tocar S2.
- Antes de escribir un archivo nuevo, decí en qué carpeta va y por qué, según la regla de dependencias de §6.
- Si el spec es ambiguo o falta un endpoint, **preguntá**. No inventes rutas del backend.
- Si necesitás una API que §4 prohíbe, **pará y decilo**. No busques un rodeo silencioso.
- Commits pequeños con prefijo de sesión: `feat(s4): race store con signals`.
- **El backend se congela al terminar la Fase 1.** Desde ahí: no lo escribas, no lo modifiques, no propongas cambiarlo. Antes de ese punto es tuyo, y se escribe contra `docs/contract/`.
