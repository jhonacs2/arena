# S4 · Directivas y pipes — guión

> **Esto es un teleprompter, no un resumen.** Lo que está entre comillas se dice. Lo que está en gris se hace. Léelo de corrido antes de dar la clase, con cronómetro.

| | |
|---|---|
| **Concepto único** | Un pipe transforma lo que se ve. Una directiva cambia cómo se comporta un elemento. Ninguno de los dos toca el componente. |
| **Al final saben** | Usar los pipes que ya vienen · escribir un pipe propio con parámetro · explicar puro contra impuro · escribir una directiva de atributo · leer una directiva estructural y saber qué hace el asterisco. |
| **Requisito previo** | S3. Signals y control flow. |
| **Archivos** | `lab/starter/src/app/sessions/s04/` · `project/frontend/starter/src/app/shared/` |

---

## Glosario de la sesión

| Palabra | En una frase |
|---|---|
| **Pipe** | Una función con nombre que se usa en el template para transformar un valor. |
| **Pipe puro** | El que Angular llama solo cuando cambia el valor de entrada. Es el valor por defecto. |
| **Pipe impuro** | El que Angular llama cada vez que revisa el componente. |
| **Encadenar** | Pasar el resultado de un pipe a otro: `{{ x \| a \| b }}`. |
| **Parámetro de pipe** | Lo que va después de los dos puntos: `{{ x \| money: 'USD' }}`. |
| **Directiva** | Una pieza sin template propio que le agrega algo a un elemento que ya existe. |
| **Directiva de atributo** | La que cambia cómo se ve o se comporta un elemento. |
| **Directiva estructural** | La que decide si un elemento existe y cuántas veces. Lleva asterisco. |
| **`ng-template`** | Un pedazo de HTML guardado que no se dibuja hasta que alguien decide dibujarlo. |
| **`host`** | El bloque donde una directiva declara qué le hace al elemento del que cuelga. |
| **`LOCALE_ID`** | El idioma con el que formatean los pipes incorporados. Por defecto, `en-US`. |

---

## 0:00 · Pregunta de apertura — 5 min

**En pantalla:** diapositiva 2.

> «Desde la primera clase venimos escribiendo esto en los templates:»
>
> ```html
> {{ horse.odds.toFixed(2) }}
> ```
>
> «Está en la parrilla, va a estar en el historial de apuestas y va a estar en el resultado de cada carrera. **¿Cuántas veces te parece que va a estar escrito el día que la aplicación esté terminada?** Tira un número en el chat.»

**Espera 90 segundos.** Van a decir 5, 20, 50. **Todos sirven.**

> «El número exacto no importa. Lo que importa es que **es más de uno**, y que el día que alguien pida mostrar las cuotas con coma en vez de punto hay que encontrarlos todos.»
>
> «Hoy vamos a dejarlo escrito una sola vez.»

---

## 0:05 · Wayground de S3 — 7 min

**Correr:** `sesiones/s03-signals/wayground.csv`.

| Si falla | Decir |
|---|---|
| El `push` sobre un signal | «Se actualiza la mitad de la pantalla. Es el peor de los tres.» |
| `track` obligatorio | «Es la única forma que tiene Angular de reconocer una fila.» |
| El `computed` que ordena | «`sort` ordena en el lugar. `[...]` antes, siempre.» |

**Máximo 30 segundos por pregunta.**

---

## 0:12 · El concepto — 8 min

> **El editor está cerrado.**

### 0:12 — El problema · 2 min

**En pantalla:** diapositiva 5.

> «Mira lo que tiene adentro un template cualquiera de los que escribimos: `toFixed(2)` para las cuotas, un método `formatMoney()` en el componente para los importes, y una condición repetida dos veces para poner la clase del favorito.»
>
> «Ninguna de esas tres cosas es lógica de esta pantalla. Son **formas de mostrar**, y las formas de mostrar se repiten en toda la aplicación.»

### 0:14 — Las tres piezas · 4 min

**En pantalla:** diapositiva 6 — `diagramas/pipes-y-directivas.svg`.

**Los términos que se definen aquí:** *pipe*, *directiva de atributo*, *directiva estructural*.

Señala cada tarjeta:

> «**Un pipe transforma lo que se ve.** Entra un valor, sale otro. Es una función con nombre que se puede usar desde el template, y no sabe quién la llamó — por eso sirve para cualquier pantalla.»
>
> «**Una directiva de atributo cambia cómo se comporta un elemento que ya existe.** No dibuja nada: se cuelga de una etiqueta y le agrega una clase, un atributo, un comportamiento.»
>
> «**Una directiva estructural decide si un elemento existe y cuántas veces.** El asterisco es la pista de que hay una detrás.»

Y ahora la pregunta que siempre aparece, contestada antes de que la hagan:

> «¿Y en qué se diferencia de un componente? En una sola cosa: **un componente trae su propio template. Una directiva no tiene template**, le agrega algo a un elemento ajeno.»

### 0:18 — Lo que se declara · 2 min

**En pantalla:** diapositiva 7.

> «Las tres piezas se declaran en `imports`, exactamente igual que un componente. Es la misma regla desde la primera clase: **si el template lo usa, se declara**.»
>
> «Y hay una que ya usaron sin saberlo: `@if` y `@for` **no** se declaran, porque no son directivas — son sintaxis del template. `*ngIf` y `*ngFor`, que es lo que había antes, sí lo eran.»

> **Si vas tarde:** se puede recortar el bloque de las 0:18. El diagrama, no.

---

## 0:20 · Live coding — 15 min

**En pantalla:** proyecto **`lab/demo`**, ruta `/s04`. La secuencia completa está en **`mision-profe.md`**.

### 0:20 — Los pipes que ya vienen · 3 min

En el template, sin tocar el componente:

```html
<p class="card__origin">{{ coffee.origin | uppercase }}</p>
<p class="card__stock">{{ coffee.stock | number }} en depósito</p>
```

**El origen sale en mayúsculas. El dato no cambió.**

> «Ese es todo el trato de un pipe: **transforma lo que se ve, no lo que hay**. Si mañana hace falta el nombre tal cual para buscar, sigue estando tal cual.»

Ahora agrega el porcentaje y **mira el resultado**:

```html
{{ share(coffee) | percent: '1.0-1' }}
```

> «Sale `54.5%`, con punto. Y toda la pantalla está en español.»
>
> «Los pipes incorporados formatean según el idioma de la aplicación, y ese idioma **por defecto es `en-US`**. Se arregla en un solo lugar, en `app.config.ts`.»

Muéstralo, sin detenerte:

```ts
registerLocaleData(localeEs);
providers: [{ provide: LOCALE_ID, useValue: 'es' }, …]
```

> «Dos líneas, una vez en la vida del proyecto. Ahora dice `54,5 %`.»

### 0:23 — Un pipe propio · 4 min

```bash
ng generate pipe sessions/s04/money
```

```ts
@Pipe({ name: 'money', standalone: true, pure: true })
export class MoneyPipe implements PipeTransform {
  transform(value: number, symbol = '$'): string {
    const formatted = new Intl.NumberFormat('es', {
      maximumFractionDigits: 0,
      useGrouping: true,
    }).format(value);

    return `${symbol} ${formatted}`;
  }
}
```

Úsalo **antes de declararlo**, a propósito:

```html
<p class="card__price num">{{ coffee.price | money }}</p>
```

> 🔴 **Rotura deliberada 1.**

```
NG8004: No pipe found with name 'money'.
```

> «*No encontré ningún pipe que se llame money.* Y tiene razón: lo escribí, pero no le dije a este componente que lo iba a usar.»

Agrégalo a `imports` y funciona. Después, el parámetro:

```html
<p class="card__alt num">{{ coffee.price | money: 'USD' }}</p>
```

> «Lo que va después de los dos puntos es el segundo argumento de `transform`. Se pueden pasar varios, separados por más dos puntos.»

### 0:27 — Puro contra impuro · 3 min

**En pantalla:** el panel de contadores de la pantalla.

> «Los dos pipes de abajo hacen exactamente lo mismo: devuelven el texto tal cual. Se diferencian en **una palabra**.»

Toca «Provocar una detección de cambios» cinco o seis veces.

> «El puro se quedó en uno. El impuro sube con cada clic.»
>
> «Un pipe **puro** —que es el valor por defecto— Angular lo llama solo cuando **cambia el valor de entrada**. Si es el mismo, reutiliza el resultado.»
>
> «Uno **impuro** corre cada vez que Angular revisa este componente. Si el componente fuera de detección por defecto en vez de `OnPush`, sería cada clic de cualquier parte de la aplicación.»

**Y la regla:**

> «Se usa impuro cuando el resultado depende de algo que el valor de entrada no ve — el reloj, por ejemplo. Casi siempre hay una forma mejor, y casi siempre esa forma es un `computed` de la clase pasada.»

### 0:30 — Una directiva de atributo · 3 min

```bash
ng generate directive sessions/s04/highlight
```

```ts
@Directive({
  selector: '[appHighlight]',
  standalone: true,
  host: {
    '[class.is-highlighted]': 'appHighlight()',
    '[attr.data-highlight-label]': 'appHighlight() ? highlightLabel() : null',
  },
})
export class HighlightDirective {
  readonly appHighlight = input(false);
  readonly highlightLabel = input('Destacado');
}
```

En el template, reemplaza los dos atributos repetidos:

```html
<li class="card" [appHighlight]="coffee.id === featuredId" highlightLabel="Café del día">
```

> «El selector entre corchetes quiere decir *cualquier elemento que tenga este atributo*. Y el input se llama igual que el selector, así se escribe una sola vez.»
>
> «Fíjate qué desapareció del template: **el nombre de la clase CSS**. La pantalla ya no decide cómo se ve un elemento destacado; solo dice cuál lo está.»

**Y quita el `standalone: true`,** a propósito:

> 🔴 **Rotura deliberada 2.**

```
TS-992011: The directive 'HighlightDirective' appears in 'imports', but is not standalone
and cannot be imported directly. It must be imported via an NgModule.
```

> «Es el mismo error que tendrían con un componente. En Angular 18 se escribe siempre, en los tres: componentes, directivas y pipes.»

### 0:33 — La estructural, para leerla · 2 min

**No la escribas: muéstrala.**

```html
<span class="bean" *appRepeat="coffee.rating">●</span>
```

> «Las dos líneas de abajo son **exactamente lo mismo**. El asterisco es azúcar sintáctica.»

```html
<span *appRepeat="3">●</span>

<ng-template [appRepeat]="3">
  <span>●</span>
</ng-template>
```

> «Un `ng-template` es un pedazo de HTML que **no se dibuja**: queda guardado, y alguien decide después si se pinta, cuántas veces y dónde. Eso es lo que hacían `*ngIf` y `*ngFor` antes de que existieran `@if` y `@for`.»
>
> «En Angular 18 casi nunca hace falta escribir una propia. Está aquí para que puedan **leer** el código que las usa, que todavía es la mayoría.»

---

## 0:35 · Misión 1 — 15 min

**Enunciado en `mision-estudiante-1.md`.**

> «Ahora ustedes, en `lab/starter`. La pantalla funciona y todo el formateo está adentro del componente. El trabajo es sacarlo: un pipe y dos directivas.»

**Dilo antes de largar:**

> «Empieza por los pipes que ya vienen, que son los de una línea. El pipe propio después. Las directivas al final, que son las que más cuestan.»

**Reloj de pistas:**

| Min | Pista, sin resolver |
|---|---|
| 0:43 | «Si el error dice `No pipe found with name`, el pipe existe pero este componente no lo declaró.» |
| 0:47 | «El selector de una directiva de atributo va entre corchetes, y el input se llama igual que el selector.» |

---

## 0:50 · Comparten pantalla — 10 min

**Preguntas, no corriges.**

1. «¿Por qué esto es un pipe y aquello una directiva?»
2. «Si mañana hay que mostrar precios en otra pantalla, ¿qué tocas?»
3. «¿Tu pipe sabe qué es un café?» — *(la respuesta correcta es «no»)*

**Lo más probable:** alguien hizo el pipe impuro «por las dudas», o le puso al pipe un parámetro que es el café entero.

> «Si el pipe recibe el café, ya no sirve para las carreras. Un pipe recibe **el valor que transforma**, no el objeto que lo contiene.»

---

## 1:00 · Descanso — 10 min

---

## 1:10 · Predice y ejecuta — 15 min

**Respuestas verificadas contra el compilador:** `predice-y-ejecuta/respuestas.md`.

| Min | Snippet | Casi todos predicen | Pasa |
|---|---|---|---|
| 1:10 | Un pipe usado sin declararlo | «se ve el valor sin transformar» | **No compila.** `NG8004` |
| 1:15 | Una directiva sin `standalone: true` | «anda igual» | **No compila.** `TS-992011` |
| 1:20 | Un pipe impuro con `OnPush` | «corre en cada clic de la app» | **Solo cuando se revisa ese componente** |

Cierra con:

> «Los dos primeros te los frena el compilador. El tercero no falla nunca: **funciona, y es lento**. Y lo lento no aparece en el proyecto de la clase, con cuatro cafés. Aparece con cuatro mil filas y un usuario que dice que la aplicación se traba.»

---

## 1:25 · Misión 2, en parejas — 20 min

**Enunciado en `mision-estudiante-2.md`.**

**Tres cosas antes de largar:**

> «Uno: los tres archivos nuevos van en `shared/`, y **ninguno sabe qué es una carrera**. Si en el pipe aparece la palabra `Race`, está en la carpeta equivocada.»
>
> «Dos: el pipe de cuotas reemplaza a `toFixed(2)`, y de paso arregla algo que `toFixed` hacía mal. Van a ver qué en cuanto lo prueben.»
>
> «Tres: acuérdense de `LOCALE_ID`. Sin eso, media pantalla formatea en inglés.»

---

## 1:45 · Code review en vivo — 10 min

Rúbrica del curso, más la pregunta de hoy:

> «Tapa el nombre del archivo. Leyendo solo el pipe, **¿podrías decir de qué aplicación es?** Si la respuesta es sí, sabe demasiado.»

Y el cierre:

> «Conté los `toFixed` que quedaron en el proyecto. Eran cinco y ahora es uno. Y el día que alguien pida las cuotas con otro formato, se cambia ahí — no en cinco lugares, ni en los tres que se van a escribir el mes que viene.»

---

## 1:55 · Exit ticket y tarea — 5 min

**Exit ticket:** `exit-ticket.md`. **Tarea:** `tarea.md`, leída en voz alta.

**Y el aviso de la próxima:**

> «La clase que viene: inyección de dependencias. Vamos a sacar los datos de adentro de los componentes, que es lo último que les queda encima.»

---

## Después de la clase

- [ ] Leer los exit tickets. Lo confuso arranca S5.
- [ ] Revisar `wayground.csv` de **esta** sesión — se corre al empezar S5.
- [ ] Aplicar la corrección de S4 al `starter/` publicado y taggear `s05`.

### Notas de la corrida real

| | |
|---|---|
| ¿Cuántos hicieron el pipe impuro «por las dudas»? | |
| ¿Se entendió el asterisco sin escribir la directiva estructural? | |
| ¿Qué pregunta no supe contestar? | |
| ¿Qué sacaría o agregaría? | |
