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

- El bloque 0:05 de **S1** corre `sesiones/s00-typescript/wayground.csv` — el tema 0, que va asíncrono antes de empezar.
- El quiz de **S11** no se corre en clase: entra en la evaluación asíncrona de cierre.

---

## Las 11 sesiones

| # | Tema | Lab (`/sNN`) | Misión 2 — hipódromo | Predice y ejecuta |
|---|---|---|---|---|
| **S1** | Filosofía e historia de Angular · Web Components · CLI · primer standalone · binding uni y bidireccional | Primer componente standalone: interpolación, `[prop]`, `(event)`, `[(ngModel)]` | Shell con header + listado de carreras con datos hardcodeados | `{{ }}` donde va `[ ]` · `[(ngModel)]` sin importar `FormsModule` |
| **S2** | Anatomía de un componente · segmentación de templates · `input()` `output()` `model()` · ciclo de vida · `ng-content` | Componente hijo con entrada y salida; `ng-content` con y sin `select` | `<app-race-card>` reutilizable · `<app-badge>` con proyección | `input.required()` sin binding · `output()` que nadie escucha · `ng-content` duplicado |
| **S3** | Signals: `signal` `computed` `set/update` · control flow `@if/@for/@switch` · inmutabilidad | Contador y lista filtrada con signals | Filtro por estado + búsqueda · `@for` con `track` · ordenar sin mutar | `array.push()` sobre un signal y la vista no repinta · `@for` sin `track` · `computed` que muta su fuente |
| **S4** | Directivas de atributo y estructurales · directivas custom · pipes built-in y custom | Directiva de resaltado + pipe de formato | `money.pipe` · `odds.pipe` · directiva de caballo favorito | Pipe impuro que corre en cada CD · directiva sin `standalone: true` · pipe usado sin importar |
| **S5** | DI a fondo · servicios · `inject()` · jerarquía de inyectores · `InjectionToken` | Un servicio `providedIn: 'root'` vs provisto en el componente | `RaceStore` y `BetStore` con signals · `API_URL` como token | Dos instancias del mismo servicio sin querer · `inject()` fuera de contexto de inyección |
| **S6** | Reactividad · promesas vs observables · hot vs cold · `map` `filter` `switchMap` · `takeUntilDestroyed` | Buscador con `debounceTime` | Buscador de carreras con debounce · interop con `toSignal()` | Suscripción que nunca se libera · `mergeMap` donde va `switchMap` · un cold observable que "no hace nada" |
| **S7** | `HttpClient` · GET/POST/PUT/DELETE · interceptores funcionales · `catchError` | Un GET contra el mock y los tres estados | Conectar al backend real · `auth.interceptor` · `error.interceptor` · cargando / vacío / error | Interceptor que no clona la request · `catchError` que se traga el error y devuelve `of(null)` |
| **S8** | Reactive Forms · `FormBuilder` · validadores custom · errores de formulario | Formulario con validador custom | Login, registro y formulario de apuesta con validación de saldo | Validador que devuelve `boolean` en vez de `ValidationErrors` · `formControlName` sin `formGroup` alrededor |
| **S9** | Routing operativo · `provideRouter` · params · guards funcionales · lazy `loadComponent` | Dos rutas, un param, un guard | App multipantalla · `authGuard` y `verifiedGuard` · lazy por feature | La ruta `**` declarada antes que las demás · guard que no devuelve nada · param que llega `undefined` sin `withComponentInputBinding()` |
| **S10** | WebSockets · zona y detección de cambios · `OnPush` · `@defer` | Socket mock que empuja a un signal | `race-live`: carrera en vivo, saldo reactivo, leaderboard con `@defer` | **El clásico**: llega el evento, el signal cambia, la vista no repinta |
| **S11** | NgModules como contexto legacy (30 min) · build de producción · deploy · code review cruzado | Un `NgModule` mínimo, solo para leerlo | App desplegada · revisión cruzada entre equipos | Build que pasa en dev y explota en `--configuration production` |

### Cuando un tema se necesita antes de su sesión

`@for` es de S3, pero no se puede mostrar un listado sin bucle en S1. Cuando eso pase, la regla es:

> **La pieza adelantada se da hecha en el starter, se nombra en voz alta durante el live coding, y se dice en qué sesión se ve a fondo.**

En S1 el `@for` viene escrito y el enunciado lo aclara: *«el control flow es S3; hoy el trabajo son los bindings de adentro»*. Así S1 puede construir una lista sin quedarse con el tema de S3.

Lo mismo aplica a `standalone: true` (se usa desde S1, se explica cuando toca) y a `OnPush` (está en todos los componentes desde el principio, se explica en S10).

### S11 rompe el guión, a propósito

Los 30 minutos de NgModules legacy y el deploy en vivo no entran en el molde de 12 bloques. S11 usa una variante: pregunta de apertura, Wayground de S10, los 30 min de NgModules, descanso, deploy guiado, y **code review cruzado entre equipos** en lugar de misiones. Va documentado en `sesiones/s11-*/guion.md`.

---

## Tema 0 — TypeScript, asíncrono

Va **antes de S1**, 100% asíncrono. Produce el `wayground.csv` que se corre en el bloque 0:05 de S1.

Tipos y anotaciones · interfaces y `type` · uniones y literales · genéricos básicos · `strict` y `strictNullChecks` · narrowing · utility types (`Pick`, `Omit`, `Partial`) · módulos ES.

El material sale del propio contrato: `Race`, `Horse`, `Bet` y `RaceStatus` son los ejemplos. El alumno tipa el dominio real antes de la primera clase, y llega a S1 con el vocabulario puesto.

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
