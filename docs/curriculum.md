# Currículo — 11 sesiones de 2 horas

Mapa de qué se enseña, dónde se practica y qué se entrega en cada sesión. Es el índice que usan las fases 3–13 del plan: cada fase produce una fila completa de esta tabla.

---

## El guión de 12 bloques

Toda sesión tiene la misma forma. La estructura es fija; lo que cambia es el contenido.

| Min | Bloque | Dura | Dónde | Qué hace el instructor |
|---|---|---|---|---|
| 0:00 | Pregunta de apertura | 5 | chat | Lee, no juzga, no corrige |
| 0:05 | **Wayground de la sesión ANTERIOR** | 7 | Wayground | Corre el quiz, comenta los errores |
| 0:12 | Concepto en diagrama | 8 | slides | **Editor cerrado.** Solo diagrama |
| 0:20 | Live coding narrado | 15 | `lab/` | Escribe y narra. **Ellos no copian, miran** |
| 0:35 | Misión 1 | 15 | `lab/starter` | **Silencio.** Disponible, no interviene |
| 0:50 | Dos alumnos comparten pantalla | 10 | pantalla del alumno | **Pregunta, no corrige** |
| 1:00 | Descanso | 10 | — | — |
| 1:10 | Predice y ejecuta | 15 | `lab/` | Muestra código roto. Predicen, después ejecutan |
| 1:25 | Misión 2, en parejas | 20 | `project/frontend/starter` | Circula entre parejas |
| 1:45 | Code review en vivo | 10 | solución de un alumno | Revisa con la rúbrica, en voz alta |
| 1:55 | Exit ticket + tarea | 5 | formulario | Reparte y cierra |

**Dos proyectos, dos momentos.** El concepto se aprende **aislado en el lab** (0:20, 0:35, 1:10) y se aplica **en el hipódromo** (1:25). Es lo que evita que la primera clase de `input()` se pelee con el modelo de dominio de las apuestas.

### El desfase del Wayground

> `sesiones/sNN/wayground.csv` contiene preguntas **sobre la sesión NN**, y se corre al empezar la **sesión NN+1**.

Es el detalle que más fácil se rompe al armar los materiales. Dos consecuencias:

- El bloque 0:05 de **S0** no es un quiz: no hay sesión anterior. Es un diagnóstico en vivo con tres programas de JavaScript que no fallan y están mal.
- El quiz de **S10** no se corre en clase: entra en la evaluación asíncrona de cierre.

---

## Las 11 sesiones

| # | Tema | Lab (`/sNN`) | Misión 2 — hipódromo | Predice y ejecuta |
|---|---|---|---|---|
| **S0** | Tipos y anotaciones · `interface` y `type` · uniones de literales · opcionales y `undefined` · narrowing · `readonly` · genéricos · utility types · módulos ES | El menú de Café Compilado: apretar los tipos de `menu.ts` | `core/models/race.model.ts` — los tipos del contrato, y cinco líneas que tienen que dejar de compilar | Widening de una propiedad de objeto · `readonly` que protege el campo pero no la lista · `as` que compila y explota |
| **S1** | Filosofía e historia de Angular · Web Components · CLI · primer standalone · binding uni y bidireccional | Primer componente standalone: interpolación, `[prop]`, `(event)`, `[(ngModel)]` | Shell con header + listado de carreras con datos hardcodeados | `{{ }}` donde va `[ ]` · `[(ngModel)]` sin importar `FormsModule` |
| **S2** | Anatomía de un componente · segmentación de templates · `input()` `output()` `model()` · ciclo de vida · `ng-content` | Extraer `<app-coffee-card>` de una pantalla que ya funciona: las cuatro puertas y los dos huecos | `<app-race-card>` en `features/` · `<app-badge>` en `shared/ui/`, con proyección | `input.required()` sin binding · `output()` que nadie escucha · `ng-content` duplicado |
| **S3** | Signals: `signal` `computed` `set/update` · control flow `@if/@for/@switch` · inmutabilidad | El tablero de la comanda: tres signals y todo lo demás derivado | Filtro por estado + búsqueda · `@for` con `track` · el panel que se cierra solo | `push` sobre un signal y **media** pantalla se queda vieja · `@for` sin `track` · `computed` que ordena su fuente |
| **S4** | Directivas de atributo y estructurales · directivas propias · pipes incorporados y propios · `LOCALE_ID` | Sacar el formateo del componente: un pipe con parámetro y dos directivas | `money.pipe` · `odds.pipe` · directiva de caballo favorito | Pipe usado sin declarar · directiva sin `standalone: true` · pipe impuro con `OnPush` |
| **S5** | DI a fondo · servicios · `inject()` · jerarquía de inyectores · `InjectionToken` | Dos mostradores y una comanda: `providedIn: 'root'` contra `providers` del componente | `RaceStore` y `BetStore` con signals · `API_URL` como token | Declarado en los dos lados: dos instancias y ningún error · `inject()` en un método · un servicio que pide otro |
| **S6** | Reactividad · promesas contra observables · frío y caliente · `debounceTime` `distinctUntilChanged` `switchMap` · `takeUntilDestroyed` · los tres estados | El buscador del catálogo: de siete búsquedas a una | Buscador de carreras con debounce · interop con `toSignal()` | `mergeMap` y gana la respuesta vieja · `catchError` afuera del `switchMap` · un observable frío que no hace nada |
| **S7** | `HttpClient` · GET/POST/PUT/DELETE · interceptores funcionales · `catchError` | Un GET contra el mock y los tres estados | Conectar al backend real · `auth.interceptor` · `error.interceptor` · cargando / vacío / error | Interceptor que no clona la request · `catchError` que se traga el error y devuelve `of(null)` |
| **S8** | Reactive Forms · `FormBuilder` · validadores custom · errores de formulario | Formulario con validador custom | Login, registro y formulario de apuesta con validación de saldo | Validador que devuelve `boolean` en vez de `ValidationErrors` · `formControlName` sin `formGroup` alrededor |
| **S9** | Routing operativo · `provideRouter` · params · guards funcionales · lazy `loadComponent` | Dos rutas, un param, un guard | App multipantalla · `authGuard` y `verifiedGuard` · lazy por feature | La ruta `**` declarada antes que las demás · guard que no devuelve nada · param que llega `undefined` sin `withComponentInputBinding()` |
| **S10** | WebSockets · zona y detección de cambios · `OnPush` · `@defer` · **NgModules como contexto legacy (30 min)** | Socket mock que empuja a un signal · un `NgModule` mínimo, solo para leerlo | `race-live`: carrera en vivo, saldo reactivo, leaderboard con `@defer` | **El clásico**: llega el evento, el signal cambia, la vista no repinta |

### Cuando un tema se necesita antes de su sesión

`@for` es de S3, pero no se puede mostrar un listado sin bucle en S1. Cuando eso pase, la regla es:

> **La pieza adelantada se da hecha en el starter, se nombra en voz alta durante el live coding, y se dice en qué sesión se ve a fondo.**

En S1 el `@for` viene escrito y el enunciado lo aclara: *«el control flow es S3; hoy el trabajo son los bindings de adentro»*. Así S1 puede construir una lista sin quedarse con el tema de S3.

Lo mismo aplica a `standalone: true` (se usa desde S1, se explica cuando toca) y a `OnPush` (está en todos los componentes desde el principio, se explica en S10).

### S0 rompe dos reglas, y las dos a propósito

TypeScript era **tema 0 asíncrono** y pasó a ser la primera clase en vivo. El motivo: el vocabulario de tipos sostiene las once sesiones que siguen, y leerlo solo no es lo mismo que ver aparecer y desaparecer un error en pantalla.

Eso obliga a dos excepciones, y las dos están escritas donde corresponde:

- **En `/s00` del lab, la ruta y el componente vienen dados también en `starter/`.** De S1 en adelante crear la ruta *es* el ejercicio; en S0 no puede serlo, porque las rutas son S9 y los componentes son S1. Lo que sí es el ejercicio está en `sessions/s00/menu.ts`, TypeScript puro. Detalle en [`lab/CLAUDE.md`](../lab/CLAUDE.md).
- **El bloque 0:05 no es un Wayground**, es un diagnóstico en vivo: no hay sesión anterior de la cual preguntar.

Y la rúbrica del code review tampoco aplica entera: los puntos 1 y 2 hablan de componentes, y en S0 no hay ninguno. `sesiones/s00-typescript/guion.md` trae la rúbrica de cinco puntos que se usa ese día, y el guión dice en voz alta que es la excepción.

### S11 se repartió

El módulo son 11 clases y TypeScript ocupa una. Lo que era S11 —NgModules legacy, build de producción, deploy y code review cruzado— se repartió así:

| Qué era de S11 | Dónde quedó |
|---|---|
| NgModules como contexto legacy (30 min) | **S10**, después del bloque de WebSockets |
| Build de producción y deploy | Cierre asíncrono, guiado por `docs/plan.md` fase 14 |
| Code review cruzado entre equipos | Cierre asíncrono, con la misma rúbrica de siempre |

Numeración: **S0 … S10**. Las sesiones S1 a S10 conservan exactamente el número que ya tenían, así que las rutas `/sNN` del lab, los `TODO(SN)` del starter y los tags publicados siguen valiendo.

---

## Qué produce cada sesión

Una carpeta `sesiones/sNN-<slug>/` con estos archivos, siempre los mismos:

```
guion.md              los 12 bloques con sus minutos y qué decir en cada uno
slides.md             Marp. El guión vive en las speaker notes del MISMO archivo
diagramas/*.svg       para el bloque 0:12, sin editor abierto
mision-estudiante-1.md           enunciado + criterios de listo (lab, individual)
mision-estudiante-2.md           enunciado + criterios de listo (hipódromo, en parejas)
predice-y-ejecuta/    snippets rotos + respuestas.md con la explicación
wayground.csv         preguntas SOBRE esta sesión — se corren en la SIGUIENTE
exit-ticket.md        3 preguntas
tarea.md              consigna asíncrona + criterio de entrega
```

Más, fuera de `sesiones/`: el slice de la solución de referencia, el slice del starter con sus `// TODO(Sn)`, y la ruta `/sNN` del lab en ambas versiones.

---

## Rúbrica del code review en vivo (bloque 1:45)

La misma en las 11 sesiones, para que el alumno sepa contra qué lo miran desde el primer día. Se revisa en este orden y en voz alta:

1. ¿Es `standalone: true` y `OnPush`?
2. ¿El estado se actualiza sin mutar? (`update(v => [...v, x])`, nunca `push`)
3. ¿Hay `any`? ¿Hay `console.log`? ¿Imports sin usar?
4. ¿La vista maneja cargando, vacío y error?
5. ¿El nombre dice lo que la cosa hace?
6. ¿Respeta la regla de dependencias — `features/` → `core/` + `shared/`, nunca al revés?

Los puntos 1 a 3 los verifica `scripts/verify.mjs`. Del 4 al 6 los verifica una persona, que es justamente por qué el bloque existe.

**S0 es la excepción y hay que decirlo en voz alta:** ese día no hay ni un componente en juego, así que los puntos 1 y 2 no aplican. Su guión trae una rúbrica de cinco puntos propia, sobre tipos, y aclara que desde S1 se vuelve a esta.
