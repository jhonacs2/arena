# S1 · Primer componente — guión

> **Esto es un teleprompter, no un resumen.** Lo que está entre comillas se dice. Lo que está en gris se hace. Leelo de corrido antes de dar la clase, con cronómetro: si tenés que parar a pensar qué sigue, avisame y lo arreglo.

| | |
|---|---|
| **Concepto único** | Un componente es una clase y un template, y hay cuatro caminos entre los dos. |
| **Al final saben** | Explicar qué es un componente · crear uno con el CLI · usar los cuatro bindings · decir por qué Angular repinta. |
| **Requisito previo** | S0 (TypeScript). Nada más. |
| **Archivos** | `lab/starter/src/app/` · `project/frontend/starter/src/app/` |

---

## Glosario de la sesión

Todo lo que se nombra hoy, en el orden en que aparece. **Ninguna de estas palabras se usa antes de definirla.** Si te escuchás diciendo una que no explicaste, parás y la explicás.

| Palabra | En una frase |
|---|---|
| **DOM** | El árbol de elementos que el navegador tiene en memoria y dibuja en pantalla. |
| **Framework** | Un conjunto de reglas y herramientas que se encarga de las partes repetitivas por vos. |
| **Componente** | Una pieza de interfaz con su propio HTML, su propio estilo y su propia lógica, en una carpeta. |
| **Clase** | El archivo `.ts` del componente: los datos y las decisiones. |
| **Template** | El archivo `.html` del componente: lo que se ve. No es HTML común — tiene instrucciones para Angular. |
| **Standalone** | Un componente que declara solo lo que usa, sin depender de un archivo lejano que lo registre. |
| **Binding** | Una conexión entre la clase y el template, para que no haya que sincronizarlos a mano. |
| **Interpolación** | El binding que pone un valor como **texto**: `{{ }}`. |
| **Property binding** | El binding que pone un valor en una **propiedad** de un elemento: `[ ]`. |
| **Event binding** | El binding que escucha un **evento** del navegador: `( )`. |
| **Two-way binding** | Los dos anteriores juntos, en un solo lugar: `[( )]`. |
| **Detección de cambios** | El momento en que Angular revisa si algo cambió y repinta lo que haga falta. |
| **CLI** | La herramienta de línea de comandos de Angular. Crea proyectos y archivos por vos. |

---

## 0:00 · Pregunta de apertura — 5 min

**En pantalla:** diapositiva 2.

> «Pensá en una web que uses todos los días. Cuando cambia un número en pantalla —el carrito, un contador, un saldo— **¿quién lo cambió?** Escribí en el chat lo primero que se te ocurra.»

**Esperá 90 segundos en silencio.** Si nadie escribe, respondé vos primero para romper el hielo.

Van a decir «el servidor», «JavaScript», «React», «no sé». **Todas sirven.** No corrijas ninguna.

Leé dos o tres en voz alta y cerrá así:

> «Todas tienen algo de razón. Lo que quiero que se lleven es que **alguien tuvo que agarrar ese pedazo de pantalla y cambiarlo**. No se actualizó solo. Hoy vamos a ver quién hace eso en Angular, y por qué eso es todo el asunto.»

> ⚠️ No expliques nada todavía. Este bloque es para que hablen, no para que aprendan.

---

## 0:05 · Wayground de S0 — 7 min

**Correr:** `sesiones/s00-typescript/wayground.csv`.

> «De acá en adelante arrancamos siempre así: siete minutos repasando lo de la clase pasada. No se corrige nota, se corrigen ideas.»

Máximo **30 segundos** de explicación por pregunta, y solo si la falló más de un tercio:

| Si falla | Decir |
|---|---|
| `type RaceStatus = string` | «`string` acepta `'galopando'`. La unión de literales es lo que hace que el compilador te frene antes de que llegue al navegador.» |
| `Horse` en vez de `Horse \| undefined` | «Una carrera sin caballos no debería existir, pero el tipo no lo puede garantizar. TypeScript te obliga a decidir qué pasa si no hay.» |
| `readonly lines: OrderLine[]` deja hacer `push` | «Son dos `readonly` y no es repetido: uno protege el campo, el otro la lista. Es el que más se olvida.» |
| Los tipos «quedan» al ejecutar | «Se borran. Lo que corre en el navegador es JavaScript. Hoy vamos a ver quién hace esa traducción: el CLI.» |

Si algo necesita más de 30 segundos, va a la tarea. **No te enganches acá.**

---

## 0:12 · El concepto — 8 min

> **El editor está cerrado.** No hay VS Code en pantalla, no hay terminal. Solo diapositivas.
> Si alguien pregunta «¿y el código?», la respuesta es: «en diez minutos. Primero quiero que puedan dibujar esto en una servilleta.»

### 0:12 — El problema, antes de nombrar nada · 2 min

**En pantalla:** diapositiva 5.

> «Imaginate que tenés un contador en una página, sin ningún framework. Escribís el número en el HTML. El usuario toca un botón, vos sumás uno a una variable… y la pantalla sigue mostrando el número viejo.»
>
> «¿Por qué? Porque **el HTML ya se dibujó**. Ese número dejó de ser tu variable en el momento en que se pintó. Para que cambie, tenés que ir a buscar el elemento y escribirle el valor nuevo, a mano.»
>
> «Con un contador es molesto. Con veinte carreras, cada una con ocho caballos y sus cuotas, es **imposible de sostener**: cada vez que cambia un dato tenés que acordarte de todos los lugares de la pantalla que lo muestran.»

**El término que se define acá:** *DOM* — el árbol de elementos que el navegador tiene en memoria y dibuja.

> «Ese árbol es el DOM. Sin framework, mantenerlo sincronizado con tus datos es trabajo tuyo, y es el trabajo que más bugs genera en cualquier interfaz.»

### 0:14 — De dónde viene Angular, en tres minutos · 3 min

**En pantalla:** diapositiva 6.

> «En 2010 apareció **AngularJS** con una idea nueva: que el HTML pudiera decir de dónde salen sus datos. En vez de ir a buscar el elemento, escribías en el HTML *acá va el nombre*, y el framework se encargaba.»
>
> «En 2016 lo reescribieron entero: **Angular 2**, que es el que llega hasta hoy. Se trajeron una idea del propio navegador —los **Web Components**—: que una parte de la pantalla pueda ser una pieza cerrada, con su HTML, su estilo y su comportamiento adentro.»
>
> «Eso es un **componente**. Angular no lo inventó: lo hizo tipado, con herramientas y con un ciclo de vida.»
>
> «Y en 2024 sacaron del medio una capa que molestaba: antes, cada componente había que anotarlo en un archivo aparte llamado NgModule. Hoy un componente **se declara solo**. A eso se le dice **standalone**, y así vamos a trabajar todo el curso.»

**Preguntas que van a hacer:**

| Preguntan | Respondé |
|---|---|
| «¿Y los NgModules ya no existen?» | «Existen, y hay miles de proyectos con ellos. Los vemos en S10, al cierre, para que puedan leer código de antes. Hoy no los necesitamos.» |
| «¿Angular y AngularJS son lo mismo?» | «No. Comparten el nombre y nada más. Si buscan algo y ven `$scope`, es AngularJS y no les sirve.» |
| «¿Por qué 18 y no la última?» | «Porque es la que vamos a ver en proyectos reales ahora. Y porque hay APIs de 19 y 20 que no compilan acá: la lista está en el repo.» |

### 0:17 — Un componente son dos cosas · 3 min

**En pantalla:** diapositiva 7 — `diagramas/componente-y-template.svg`.

> «Un componente son **dos archivos que se hablan**.»
>
> «A la izquierda, **la demo**: un archivo TypeScript. Ahí viven los datos y las decisiones. Cuánto sale el café, si hay stock, qué pasa cuando alguien pide.»
>
> «A la derecha, **el template**: un archivo HTML. Es lo que se ve. Y ojo con esto, porque es la primera trampa: **el template no es HTML común**. Es HTML con instrucciones adentro, y esas instrucciones las lee Angular antes de dibujar.»
>
> «Entre los dos hay cuatro caminos. A cada camino se le dice **binding**, que quiere decir *atadura*: atás un pedazo del template a un pedazo de la clase, y Angular se encarga de que no se despeguen.»

Señalá cada flecha del diagrama mientras decís:

> «Uno. De la clase al template, como **texto**: eso es la **interpolación**.»
> «Dos. De la clase al template, pero a una **propiedad** del elemento —si está deshabilitado, qué clase tiene—: eso es **property binding**.»
> «Tres. Del template a la clase: el usuario hizo algo y hay que enterarse. **Event binding**.»
> «Cuatro. Los dos sentidos a la vez, en un solo lugar. **Two-way binding**.»

**Y ahora la pregunta que casi nadie hace y todos necesitan:**

> «¿Cómo se entera Angular de que algo cambió?»
>
> «No hay magia y no hay un vigilante mirando tus variables. Angular espera a que pase **algo**: un clic, una tecla, una respuesta del servidor. Cuando eso pasa, revisa los bindings de la pantalla y repinta lo que haga falta. A ese momento se le dice **detección de cambios**.»
>
> «Guardensé esa frase: **Angular repinta después de que pasó algo.** En la clase 10 vamos a ver un caso donde “pasa algo” pero Angular no se entera, y va a tener todo el sentido del mundo.»

> **Si vas tarde:** lo único que se puede recortar de este bloque son los tres minutos de historia. El diagrama y la detección de cambios, no.

---

## 0:20 · Live coding — 15 min

**Decilo textual antes de empezar:**

> «Cierren el editor. Los próximos quince minutos yo escribo y ustedes miran. No copien: van a hacer esto mismo después, y con las manos libres se entiende mejor. Si me equivoco, avisen.»

**En pantalla:** VS Code y el navegador lado a lado. Proyecto: **`lab/demo`**.

> **Antes de entrar al aula:** `node scripts/prep-demo.mjs` y después `cd lab/demo && npm start`.
>
> `lab/demo` es una copia descartable de `lab/starter`: arranca en el mismo estado en que están los alumnos, con el componente de S1 sin existir. **No le borres nada a `lab/solution`** — tu solución de referencia queda intacta, y si el bloque se descarrila, `prep-demo.mjs` te devuelve el lienzo limpio en un segundo.
>
> La secuencia de tecleo completa, con las dos roturas deliberadas, está en **`mision-profe.md`**. Ponela en el segundo monitor: este guión lleva la clase, ese archivo lleva el teclado.

### 0:20 — El CLI crea el componente · 2 min

```bash
cd lab/demo
ng generate component sessions/s01 --flat
```

> «Esto es el **CLI**, la herramienta de línea de comandos de Angular. Le pedí un componente y me creó cuatro archivos: el TypeScript, el HTML, el CSS y uno de tests.»
>
> «Podría haberlos creado a mano. Uso el CLI porque los nombra igual siempre y porque no me olvido de ninguno. **No hace nada mágico**: mírenlos, están vacíos.»

Abrí el `.ts` y leelo en voz alta, línea por línea:

> «`@Component` es un **decorador**: le dice a Angular “esta clase no es una clase cualquiera, es un componente”. Y adentro va la configuración.»
> «`selector` es el nombre de la etiqueta con la que se usa.»
> «`standalone: true` es lo que hablamos: se declara solo.»
> «`templateUrl` y `styleUrl` apuntan a los otros dos archivos.»

### 0:22 — Que aparezca en pantalla · 2 min

Sumá la ruta en `app.routes.ts` y `available: true` en `sessions.ts`. Navegá a `/s01`.

> «Fíjense que hicieron falta **dos cosas**: la ruta, para que la dirección exista, y el índice, para que aparezca en el menú de la izquierda. Son dos archivos distintos y hay que tocar los dos. Es el error más común de la clase que viene.»

**Ahora se ve una página en blanco con el texto que el CLI dejó.** Señalalo:

> «Eso es lo mínimo que existe. De acá para arriba, todo lo que aparezca lo vamos a poner nosotros.»

### 0:24 — Interpolación · 3 min

Escribí en la clase:

```ts
protected coffee = { name: 'Yirgacheffe', origin: 'Etiopía', price: 42, available: true };
```

Y en el template:

```html
<h2>{{ coffee.name }}</h2>
<p>{{ coffee.origin }}</p>
<p>{{ coffee.price }}</p>
```

> «Las llaves dobles dicen: “Angular, poné acá el valor de esta expresión, como texto”.»

**Ahora hacé la demostración que vale por toda la explicación:** cambiá `price: 42` a `price: 55` y guardá.

> «Cambié **un solo lugar** —la clase— y la pantalla se actualizó sola. Eso es el trato: el dato vive en un lugar, y el template lo mira. Nunca más vas a ir a buscar un elemento para escribirle el número.»

### 0:27 — Property binding · 3 min

> «Ahora quiero que la card se vea distinta cuando no hay stock. La clase CSS ya existe, se llama `product--soldout`.»

Escribí primero **el error**, a propósito:

```html
<div class="product product--soldout">
```

> «Así queda puesta siempre, y no es lo que quiero. Necesito que dependa de un dato.»

Y ahora la forma correcta:

```html
<div class="product" [class.product--soldout]="!coffee.available">
```

> «Los corchetes cambian todo. Sin corchetes, lo que va entre comillas es **texto literal**. Con corchetes, es **una expresión de TypeScript** que Angular evalúa.»
>
> «Y fíjense que puedo tener las dos cosas en el mismo elemento: `class` con lo que va siempre, y `[class.algo]` con lo que va a veces. **No se pisan, se suman.**»

Cambiá `available` a `false` y mostrá el resultado.

### 0:30 — Event binding · 3 min

> «Hasta acá la información va en un solo sentido: de la clase al template. Ahora el otro.»

En la clase:

```ts
protected toggleAvailability(): void {
  this.coffee = { ...this.coffee, available: !this.coffee.available };
}
```

En el template:

```html
<button type="button" (click)="toggleAvailability()">
  {{ coffee.available ? 'Marcar agotado' : 'Marcar disponible' }}
</button>
```

> «Los paréntesis escuchan un evento del navegador. `click` no lo inventó Angular: es el mismo de siempre. Lo que agrega Angular es que en vez de escribir un `addEventListener`, decís qué método llamar.»

Tocá el botón. Andá y volvé un par de veces.

> «Y acá está lo de la detección de cambios: **tocaron el botón, entonces pasó algo, entonces Angular revisó**. Por eso el texto del botón y la clase del div se actualizaron los dos, sin que yo tocara nada más.»

> Fijate en el `{ ...this.coffee, ... }`. Si alguien pregunta: «creo un objeto nuevo en vez de modificar el que estaba. Es una regla del curso y en la clase 3 van a ver por qué importa.» **No te extiendas más que eso.**

### 0:33 — Two-way binding, y el error · 4 min

> «Falta el cuarto: cuando la información tiene que ir en los dos sentidos. El caso típico es un input.»

En la clase: `protected customer = '';`

En el template:

```html
<input type="text" [(ngModel)]="customer" />
<p>Hola, {{ customer }}</p>
```

**Guardá. Va a fallar.** Leé el error completo en voz alta, del navegador o de la terminal:

```
NG8002: Can't bind to 'ngModel' since it isn't a known property of 'input'.
```

> «Buenísimo que falle. Léanlo conmigo: *no puedo enlazar a `ngModel` porque no es una propiedad conocida de `input`*.»
>
> «Y tiene razón: `ngModel` no existe en HTML. Es algo que trae Angular. Y acá está la parte importante — **este componente es standalone, o sea que declara solo lo que usa**. Yo nunca le dije que iba a usar `ngModel`.»

Agregá el import:

```ts
import { FormsModule } from '@angular/forms';
// …
imports: [FormsModule],
```

> «Ahora sí. Y quiero que se lleven la regla completa: **si el template usa algo de Angular, tiene que estar en `imports`.** Vale para esto, para el router, y desde la clase que viene, para cada componente que usen adentro de otro.»

Escribí en el input y mostrá cómo el `<p>` de abajo cambia mientras escribís.

> «Los dos sentidos: lo que escribo va a la clase, y lo que está en la clase se ve abajo. `[(ngModel)]` es literalmente `[ngModel]` más `(ngModelChange)` — property binding y event binding juntos, con azúcar sintáctica.»

---

## 0:35 · Misión 1 — 15 min

**En pantalla:** diapositiva 15 con el enunciado. Enunciado completo en `mision-estudiante-1.md`.

> «Ahora ustedes. Mismo mostrador, en `lab/starter`. El componente **no existe**: lo crean con el CLI, igual que recién. Quince minutos.»

> **Estás en silencio.** Disponible si preguntan, pero no circulás ofreciendo ayuda. Los quince minutos de pelearse con el error **son la demo**.

**Reloj de pistas** — solo si más de la mitad está trabada en lo mismo:

| Min | Pista, en voz alta, sin resolver |
|---|---|
| 0:43 | «¿A alguien le apareció un error con `ngModel`? Vuelvan a leerlo entero: dice dónde está el problema.» |
| 0:47 | «Acordate: para que la ruta aparezca en el menú hay que tocar **dos** archivos.» |

---

## 0:50 · Comparten pantalla — 10 min

Dos personas. **Una que le funciona y una que no** — a la segunda pedile permiso antes.

> **Preguntás, no corregís.** Aunque veas el error en el primer segundo.

1. «¿Qué esperabas que pasara acá?»
2. «¿Qué pasó?»
3. «¿Cómo te diste cuenta?»
4. «Si tuvieras que explicarle esta línea a alguien que no vino hoy, ¿qué le decís?»

**Lo más probable que aparezca:** alguien escribió `class="product--soldout"` en vez de `[class.product--soldout]="…"`. Es la mejor pantalla posible para compartir — no la arregles vos, dejá que la encuentren entre todos.

---

## 1:00 · Descanso — 10 min

> «Diez minutos. Vuelvan puntuales, que lo que viene es lo mejor de la clase.»

---

## 1:10 · Predice y ejecuta — 15 min

**Los archivos:** `predice-y-ejecuta/`. **Las respuestas:** `predice-y-ejecuta/respuestas.md`, y están verificadas en el navegador — no las improvises.

**El orden no se saltea:**

1. Mostrás el código. **No lo ejecutás.**
2. «¿Qué va a pasar? Escribilo en el chat.» — **60 segundos de reloj.**
3. Recién ahí, ejecutás.
4. Explicás la diferencia entre lo que dijeron y lo que pasó.

> El paso 2 es todo el ejercicio. Si ejecutás primero, esto es una demo y no aprende nadie.

| Min | Snippet | Casi todos predicen | Pasa |
|---|---|---|---|
| 1:10 | `class="a {{ b }}"` junto a `[class.x]` | «se pisan, gana una» | **Quedan las tres clases.** Se combinan |
| 1:15 | `[(ngModel)]` sin `FormsModule` | «anda igual» | **No compila.** `NG8002` |
| 1:20 | `(click)="contador + 1"` | «suma uno» | Se queda en **0**, y **no hay ningún error** |

Cerrá el bloque con esta pregunta:

> «De los tres, ¿cuál les habría costado más encontrar en un proyecto de verdad?»

Casi siempre eligen el tercero, y tienen razón:

> «Exacto. Los otros dos te frenan el build. Este te deja seguir. **Un binding que no falla no quiere decir que ande** — el silencio es el peor síntoma que puede tener un bug.»

---

## 1:25 · Misión 2, en parejas — 20 min

**En pantalla:** diapositiva 21. Enunciado en `mision-estudiante-2.md`.

> «Ahora al proyecto de verdad. En parejas, veinte minutos: diez escribe uno y dicta el otro, y a los diez se invierten. El que dicta no toca el teclado; el que escribe no decide.»

**Tres cosas para decir antes de largar:**

> «Uno: la pantalla **no existe**. La crean ustedes, con el CLI, igual que en el lab.»
>
> «Dos: los datos ya están. Son las ocho carreras de verdad, con sus 54 caballos, en `core/mocks`. No inventen datos.»
>
> «Tres: el `@for` para recorrer la lista se los damos escrito en la corrección. Recorrer listas es la clase 3; **hoy el trabajo son los bindings**.»

Circulás entre las parejas. **Escuchás más de lo que hablás.** La pareja que termina recibe la extensión del final del enunciado.

---

## 1:45 · Code review en vivo — 10 min

Una solución de la Misión 2, con permiso. **En pantalla, al lado, `correccion.md`.**

Rúbrica, en voz alta y en este orden:

1. ¿`standalone: true` y `OnPush`?
2. ¿Actualiza el estado sin mutar?
3. ¿`any`, `console.log`, imports sin usar?
4. ¿El nombre dice lo que la cosa hace?
5. ¿Está en la carpeta que le toca?

**Empezá por algo que está bien hecho.** Siempre hay algo.

Y decí esto, textual, porque si no lo van a esperar:

> «Se van a dar cuenta de que en la lista no hay estado de carga ni de error. Es correcto: **hoy no hay nada que cargar**, los datos ya están en el código. En la clase 7, cuando los pidamos a un servidor, los tres estados van a ser obligatorios.»

---

## 1:55 · Exit ticket y tarea — 5 min

**Exit ticket:** `exit-ticket.md`. Tres preguntas, tres minutos.

> «La tercera es la que más me sirve: qué quedó confuso. Vale “nada”, vale “todo”, vale una palabra. Con eso arranco la clase que viene.»

**Tarea:** `tarea.md`. **Leela en voz alta antes de cortar.** Una tarea que solo se manda por chat no se hace.

**Y decí esto, que es lo que más se olvida:**

> «Última cosa, y es importante. En el repo les dejé un archivo que se llama **`conceptos.md`**. Está todo lo de hoy: cada concepto con su definición y **los ejemplos exactos que corrimos acá**. La clase es en vivo y no queda grabada, así que cuando se sienten a hacer la tarea, eso es lo que tienen en vez de acordarse. Ténganlo abierto al lado del editor.»

> ⚠️ No lo saltees por falta de tiempo. Si nadie sabe que el apunte existe, es como si no existiera — y es lo único que tienen entre esta clase y la que viene.

---

## Después de la clase

- [ ] Leer los exit tickets. Lo confuso arranca S2.
- [ ] Escribir `wayground.csv` de **esta** sesión con lo que más falló — se corre al empezar S2.
- [ ] Completar las notas de abajo.

### Notas de la corrida real

*Completá después de dar la clase. Es lo que hace que S2 salga mejor.*

| | |
|---|---|
| ¿Qué bloque se pasó de tiempo? | |
| ¿Qué pregunta no supe contestar? | |
| ¿Qué error apareció que no estaba previsto? | |
| ¿Qué sacaría o agregaría? | |
