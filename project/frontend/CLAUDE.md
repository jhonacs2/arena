# CLAUDE.md — Frontend del hipódromo

> Complementa el [`CLAUDE.md` de la raíz](../../CLAUDE.md). Las APIs prohibidas y `verify.mjs` están allá.

`solution/` es la referencia del instructor. `starter/` es lo que recibe el alumno. Los dos tienen la misma configuración; lo que cambia es cuánto código de producto hay.

---

## Convenciones de código

- **Standalone en todo.** Cero `NgModule`. Bootstrap con `bootstrapApplication` + `app.config.ts`.
- **`inject()` siempre**, nunca inyección por constructor. Es lo que se enseña en clase.
- **`ChangeDetectionStrategy.OnPush` en todos los componentes**, sin excepción.
- **Estado con signals**, no con `BehaviorSubject`. RxJS se reserva para HTTP, WebSocket y eventos del DOM.
- **Inmutabilidad estricta.** Nunca `array.push()` sobre estado; siempre `signal.update(v => [...v, x])`. Se evalúa en clase.
- **Nada de NgRx.** El estado vive en servicios con signals privados de escritura y expuestos de solo lectura:
  ```ts
  private readonly _races = signal<Race[]>([]);
  readonly races = this._races.asReadonly();
  ```
- Interceptores y guards **funcionales** (`HttpInterceptorFn`, `CanActivateFn`).
- `provideRouter(routes, withComponentInputBinding())` para que los params lleguen como `input()`.
- Nada de `any`. `strict` más `noUncheckedIndexedAccess`, `noUnusedLocals` y `noUnusedParameters`.
- Archivos en kebab-case, con sufijo real: `race-card.component.ts`.
- **Todo el texto de la UI en español. Los identificadores del código, en inglés.** Sin excepciones en ninguno de los dos sentidos.

---

## Estructura

```
src/app/
├── core/                       # singletons, se proveen una sola vez
│   ├── auth/                       auth.service · auth.interceptor · auth.guard · verified.guard
│   ├── http/                       api.config · error.interceptor · mock-backend.interceptor
│   ├── ws/                         socket.service · mock-socket.service · ws-events.model
│   ├── theme/theme.service.ts      claro / oscuro / sistema
│   ├── mocks/                      GENERADO por gen-mocks.mjs desde el seed
│   └── models/                     Race, Horse, Bet, User, Page, ApiError
├── shared/
│   ├── ui/                         silk, button, skeleton, empty-state, logo
│   ├── pipes/                      money.pipe, odds.pipe
│   └── directives/
├── layout/                     shell.component · balance-widget.component
└── features/                   sistema · auth · races · bets · leaderboard
```

**Regla de dependencias:** `features/` puede importar de `core/` y `shared/`. `shared/` no importa de `features/` ni de `core/`. `core/` no importa de `features/`. Si necesitás romperla, no la rompas: mové el código.

### Lo que NO se construye antes de su sesión

`<app-badge>` y `<app-race-card>` son el ejercicio de S2. Si están hechos de antemano, esa sesión se queda sin práctica.

**Vale para cualquier pieza de la columna «qué construye el alumno» de `docs/curriculum.md`:** se deja para su sesión, tanto en `solution/` como en `starter/`. La solución se escribe cuando se escribe la clase, no antes.

---

## Datos falsos, por etapa

El frontend arranca antes que el backend y **las primeras seis sesiones no lo necesitan**. Pero los datos falsos no se inventan: salen del mismo `docs/contract/seed/` que carga Go.

| Sesiones | Qué usa | Por qué |
|---|---|---|
| S1–S4 | `core/mocks/` — consts tipadas | Datos hardcodeados, como pide el temario |
| S5–S6 | `RaceStore` con signals; `of(fixture).pipe(delay(400))` | Async real sin HTTP; habilita debounce y `toSignal()` |
| S7+ | `HttpClient` real + token `API_URL` | Apunta al Go real **o** a `provideMockBackend()`. Cambiar es una línea en `environment.ts` |
| S10 | `MockSocketService` reproduce `race-ticks.jsonl` a 10 Hz | **Seguro de clase**: si el deploy se cae, la carrera igual corre |

El interceptor de mock responde desde el seed con los mismos códigos de error del catálogo: un componente escrito contra el mock funciona contra el backend real sin tocar una línea.

---

## Tests

`solution/` testea lo que el alumno tiene que lograr — es la referencia de la misión.

`starter/` testea **solo lo que ya viene hecho**, así pasa desde el minuto cero. Lo que es el ejercicio se comprueba con los criterios de «Listo cuando» del enunciado, mirando la pantalla. Aprender a mirar y decidir si está bien es parte de la clase; un test verde no reemplaza eso.

```bash
npm start        # http://localhost:4200
npm run build
npm test
```
