# S1 · Primer componente — guión

> Documento maestro de la sesión. Todo lo demás (slides, misiones, quiz) sale de acá.
> **Antes de dar la clase:** cronometrar contra estos bloques. Si algo se pasa 3 minutos, se recorta acá y no en vivo.

| | |
|---|---|
| **Concepto único** | Un componente es una clase con un template, y hay cuatro caminos entre los dos. |
| **Al final saben** | Crear un componente standalone · usar los cuatro bindings · explicar por qué `[(ngModel)]` necesita `FormsModule`. |
| **Requisito previo** | Tema 0 (TypeScript, asíncrono) |
| **Archivos que se tocan** | `lab/starter/src/app/sesiones/s01/` · `project/frontend/starter/src/app/features/races/` |

### Sobre la historia y la filosofía

El temario pide «filosofía e historia de Angular · Web Components · CLI». Van **como encuadre de tres minutos dentro del bloque 0:12, no como clase aparte**.

El motivo es concreto: en la primera sesión hay que salir con algo funcionando en pantalla. Una persona que escribió su primer componente entiende para qué sirvió AngularJS mucho mejor que una que escuchó veinte minutos de historia sin haber escrito nada. La historia vuelve, con contexto, en S11 con los NgModules.

---

## 0:00 — Pregunta de apertura · 5 min

> **«Pensá en una web que uses todos los días. Cuando cambia un número en pantalla —el carrito, un contador, un saldo— ¿quién lo cambió? Escribí lo primero que se te ocurra.»**

Responden **en el chat**. Sin juicio, sin corregir, sin «casi». Se leen dos o tres en voz alta y se sigue.

Va a haber de todo: «el servidor», «JavaScript», «React», «no sé». **Todas sirven.** La pregunta no busca la respuesta correcta: busca que noten que hay alguien moviendo el DOM, y que hoy ese alguien va a ser Angular.

- [x] Pregunta abierta, sin respuesta correcta
- [x] Se responde en una línea
- [x] Conecta con el concepto de hoy sin adelantarlo

---

## 0:05 — Wayground de Tema 0 · 7 min

Correr **`sesiones/s00-typescript/wayground.csv`** — TypeScript, el material asíncrono.

> Es la única sesión donde el quiz no es «de la anterior»: no hay anterior. De S2 en adelante se corre el de la sesión previa.

| Si falla mucho | Decir |
|---|---|
| `type RaceStatus = string` | «`string` acepta `'galopando'`. La unión de literales es lo que hace que el compilador te frene.» |
| `Horse` en vez de `Horse \| undefined` | «Una carrera sin caballos no debería existir, pero el tipo no lo puede garantizar. `strictNullChecks` te obliga a decidir qué pasa.» |
| `readonly` | «`readonly` no congela el objeto: impide `push` y asignar por índice. Hoy alcanza con eso.» |

**No más de 30 segundos por pregunta.** Si necesita más, va a la tarea.

---

## 0:12 — Concepto en diagrama · 8 min

**El editor está cerrado.** No hay VS Code en pantalla. Solo el diagrama.

Diapositivas: `slides.md` · Diagrama: `diagramas/componente-y-template.svg`

### 0:12 — Tres minutos de encuadre

- **2010, AngularJS.** La idea nueva era que el HTML pudiera tener lógica declarativa. Funcionó tan bien que se copió en todos lados.
- **2016, Angular 2.** Reescritura completa. Se trajo de los **Web Components** la idea central: *el navegador ya sabe de componentes*. Un `@Component` de Angular es esa idea, con tipado y herramientas encima.
- **2023–2024, standalone.** Angular sacó del medio la capa de `NgModule`. Hoy un componente declara sus propias dependencias en `imports`, sin intermediarios. **Así vamos a trabajar todo el curso.** Los NgModules se ven en S11, como contexto para leer código viejo.

> Una frase para dejar: *«Angular no inventó los componentes. Los hizo tipados, con herramientas y con un ciclo de vida.»*

### 0:15 — Cinco minutos de diagrama

Dos cajas: **la clase** (TypeScript, los datos) y **el template** (HTML, lo que se ve). Cuatro flechas entre ellas:

```
             ┌──────────────┐                    ┌──────────────┐
             │              │  {{ interpolación }}│              │
             │    CLASE     │ ──────────────────▶ │   TEMPLATE   │
             │              │   [property]="…"    │              │
             │  nombre      │ ──────────────────▶ │   lo que se  │
             │  precio      │                     │      ve      │
             │  agregar()   │ ◀────────────────── │              │
             │              │    (evento)="…"     │              │
             │              │ ◀──────────────────▶│              │
             └──────────────┘   [(ngModel)]="…"   └──────────────┘
```

**La analogía:** un mostrador de cafetería. La clase es la trastienda —lo que hay, cuánto cuesta, quién atiende—; el template es la vidriera. Interpolación es escribir el precio en el cartel. Property binding es apagar la luz del cartel cuando se acabó. Event binding es el timbre que suena cuando alguien pide. Y two-way es la libreta del pedido: la escribe el cliente y la lee el mozo.

**Que puedan dibujar las dos cajas y las cuatro flechas antes de escribir una línea.** Si arranca el código acá, copian sintaxis sin modelo mental.

---

## 0:20 — Live coding narrado · 15 min

**Ellos no copian. Miran.** Decirlo explícito al empezar: *«cierren el editor, esto lo hacemos juntos en 15 minutos y después lo hacen ustedes»*.

Proyecto: `lab/solution` → ruta `/s01`

| Min | Qué escribo | Qué narro mientras escribo |
|---|---|---|
| 0:20 | `ng serve` y recorro la estructura | «Esto lo generó el CLI. `main.ts` arranca la app, `app.config.ts` la configura. Cero `NgModule`.» |
| 0:22 | El `@Component` vacío: selector, `standalone: true`, template | «`standalone: true` en Angular 18 hay que escribirlo. En 19 pasa a ser el valor por defecto — si copian código de un blog reciente, ojo con eso.» |
| 0:25 | La propiedad `cafe` y su interpolación | «Escribo el precio una sola vez, en la clase. El template lo lee.» **Cambiar el precio en vivo y que vean la recarga.** |
| 0:29 | `[class.producto--agotado]` | «Sin corchetes es la cadena literal; con corchetes es una expresión de TypeScript. Es *la* diferencia.» |
| 0:33 | `(click)="alternarDisponibilidad()"` | «Los paréntesis van del template a la clase. El DOM avisa, la clase decide.» |
| 0:37 | `[(ngModel)]` **sin** `FormsModule` → error | **Romperlo a propósito.** Leer el error en voz alta antes de arreglarlo. |
| 0:39 | Agregar `FormsModule` a `imports` | «Los imports son de ESTE componente. Standalone quiere decir que se declara solo.» |

> El error de `ngModel` es el más frecuente de la sesión. Que lo vean acá, con la pantalla del instructor, hace que a las 0:35 lo reconozcan solos.

---

## 0:35 — Misión 1 · 15 min

Enunciado: `mision-1.md` · Trabajan en `lab/starter`, ruta `/s01` · **Individual**

**Estás en silencio.** Disponible si preguntan, pero no circulás ofreciendo ayuda. Los quince minutos de pelearse solo con el error son la clase.

Si a los 8 minutos más de la mitad está trabada en lo mismo → una pista para todos, en voz alta, sin resolver. La más probable: *«¿alguien ya vio el error de `ngModel`? ¿Dónde dice Angular que hay que declarar lo que usa un componente?»*

---

## 0:50 — Dos alumnos comparten pantalla · 10 min

**Preguntás, no corregís.** Aunque esté mal. Aunque duela.

1. ¿Qué esperabas que pasara acá?
2. ¿Qué pasó?
3. ¿Cómo lo averiguaste?
4. Si tuvieras que explicarle esta línea a alguien que no estuvo hoy, ¿qué le decís?

Elegir **una que funciona y una que no**. La que no funciona enseña más — pedir permiso antes.

Lo más probable que aparezca: alguien escribió `class="producto--agotado"` en vez de `[class.producto--agotado]="…"`. Es la mejor pantalla posible para compartir.

---

## 1:00 — Descanso · 10 min

Volver puntual: los quince minutos de después son los más densos.

---

## 1:10 — Predice y ejecuta · 15 min

Carpeta: `predice-y-ejecuta/` · Respuestas: `predice-y-ejecuta/respuestas.md`

Para cada snippet, en este orden y sin saltearse pasos:

1. Se muestra el código. **No se ejecuta.**
2. *«¿Qué va a pasar? Escribilo en el chat.»* — 60 segundos.
3. Se ejecuta.
4. Se explica la diferencia entre lo que dijeron y lo que pasó.

**El paso 2 es todo el ejercicio.** Ejecutar primero lo convierte en una demo.

| # | Qué está roto | Qué predicen casi todos | Qué pasa de verdad |
|---|---|---|---|
| 1 | `class="a {{ b }}"` mezclado con `[class.x]` | «Se pisan; gana una» | **Se combinan.** Quedan las tres clases — verificado en el navegador |
| 2 | `[(ngModel)]` sin `FormsModule` | «Anda igual, ngModel es de Angular» | **No compila.** `Can't bind to 'ngModel'…` |
| 3 | `(click)="contador + 1"` en vez de `contador = contador + 1` | «Suma uno» | Evalúa la expresión y tira el resultado. El número no se mueve |

---

## 1:25 — Misión 2, en parejas · 20 min

Enunciado: `mision-2.md` · Trabajan en `project/frontend/starter` · **En parejas**

Acá el concepto aterriza en el hipódromo. Es el único bloque que toca el proyecto ancla.

**Conducción por turnos:** 10 minutos escribe uno y dicta el otro; a los 10 se invierte.

Dos cosas para decir antes de largar:

- **El `@for` se los damos hecho.** El control flow es S3. Hoy el trabajo son los bindings de adentro.
- **El starter ya funciona a medias.** No arrancan de una pantalla en blanco: arrancan de algo que anda mal, y eso es más parecido al trabajo real que a un ejercicio.

Circulás entre las parejas. Escuchás más de lo que hablás. La pareja que termina recibe la extensión del final del enunciado.

---

## 1:45 — Code review en vivo · 10 min

Una solución de la Misión 2, con permiso de la pareja. Con la rúbrica de `docs/curriculum.md`, en voz alta y en este orden:

1. ¿`standalone: true` y `OnPush`?
2. ¿Actualiza el estado sin mutar?
3. ¿`any`, `console.log`, imports sin usar?
4. ¿Maneja cargando, vacío y error?
5. ¿El nombre dice lo que la cosa hace?
6. ¿Respeta la regla de dependencias?

En S1, el punto 4 casi no aplica —los datos son hardcodeados— y conviene decirlo: *«hoy no hay estado de carga porque no hay nada que cargar. En S7 vuelve, y va a ser obligatorio.»*

Empezar por algo que está bien hecho. Siempre hay algo.

---

## 1:55 — Exit ticket y tarea · 5 min

Formulario: `exit-ticket.md` — tres preguntas, tres minutos.

Tarea: `tarea.md`. **Leerla en voz alta antes de cortar.** Si se manda solo por chat, no se hace.

---

## Después de la clase

- [ ] Leer los exit tickets. Lo confuso arranca S2.
- [ ] Escribir `sesiones/s01-primer-componente/wayground.csv` con lo que más falló hoy — se corre al empezar S2.
- [ ] Anotar acá abajo qué bloque se pasó de tiempo y por qué.

### Notas de la corrida real

*Se completa después de dar la clase. Es lo que hace que S2 salga mejor.*
