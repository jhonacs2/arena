# S2 · Anatomía de un componente — guión

> **Esto es un teleprompter, no un resumen.** Lo que está entre comillas se dice. Lo que está en gris se hace. Leelo de corrido antes de dar la clase, con cronómetro: si tenés que parar a pensar qué sigue, avisame y lo arreglo.

| | |
|---|---|
| **Concepto único** | Los datos bajan por `input()`, los avisos suben por `output()`, y `ng-content` es un hueco que llena el padre. |
| **Al final saben** | Decidir qué se lleva un componente hijo y qué se queda en el padre · escribir `input()`, `input.required()`, `model()` y `output()` · usar `ng-content` con y sin `select` · nombrar los tres ganchos del ciclo de vida que se usan. |
| **Requisito previo** | S1. Los cuatro bindings, y que un componente es una clase más un template. |
| **Archivos** | `lab/starter/src/app/sessions/s02/` · `project/frontend/starter/src/app/features/races/` |

---

## Glosario de la sesión

Todo lo que se nombra hoy, en el orden en que aparece. **Ninguna de estas palabras se usa antes de definirla.** Si te escuchás diciendo una que no explicaste, parás y la explicás.

| Palabra | En una frase |
|---|---|
| **Padre** | El componente que usa a otro adentro de su template. |
| **Hijo** | El componente usado. No sabe quién lo usa. |
| **Componer** | Armar una pantalla juntando componentes, en vez de escribir todo el marcado junto. |
| **`input()`** | Una puerta de entrada: un dato que el padre le pasa al hijo. |
| **`input.required()`** | Lo mismo, pero sin ese dato el componente **no compila**. |
| **`output()`** | Una puerta de salida: un aviso que el hijo manda al padre. |
| **`emit()`** | Mandar el aviso por un `output()`. |
| **`$event`** | En un `(output)`, el dato que el hijo pasó a `emit()`. No es un evento del DOM. |
| **`model()`** | Entrada y salida a la vez: habilita `[(propiedad)]` sobre un componente propio. |
| **Proyección de contenido** | Que el padre meta HTML propio adentro de la etiqueta del hijo. |
| **`ng-content`** | El hueco donde cae ese HTML. |
| **`select`** | El filtro que separa un hueco de otro: `<ng-content select="[card-tag]" />`. |
| **Ciclo de vida** | Los momentos por los que pasa un componente: se crea, cambian sus datos, se destruye. |
| **`ngOnInit`** | Corre una vez, cuando los inputs ya llegaron. |
| **`ngOnChanges`** | Corre cada vez que el padre cambia un input. |
| **`ngOnDestroy`** | Corre cuando el componente se va de la pantalla. |

---

## 0:00 · Pregunta de apertura — 5 min

**En pantalla:** diapositiva 2.

> «La semana pasada hicimos el listado de carreras. Ocho carreras, todas iguales, todas escritas una sola vez adentro de un `@for`.»
>
> «Ahora imaginate que el jueves te piden mostrar **una sola carrera** en la página de inicio, con el mismo aspecto. ¿Qué harías, con lo que sabés hoy?»
>
> «Escribilo en el chat, sin pensarlo mucho.»

**Esperá 90 segundos en silencio.**

Van a decir «copio y pego el HTML», «hago otro `@for` de un elemento», «no sé». **Todas sirven.** No corrijas ninguna.

Leé dos o tres y cerrá así:

> «El que dijo copiar y pegar tiene razón: es lo único que se puede hacer hoy. Y también es exactamente el problema — **a partir del jueves hay dos lugares que hay que acordarse de cambiar juntos**.»
>
> «Hoy vamos a ver cómo se hace para que haya uno solo.»

> ⚠️ No expliques nada todavía. Este bloque es para que hablen.

---

## 0:05 · Wayground de S0 y S1 — 7 min

**Correr:** `sesiones/s01-primer-componente/wayground.csv`.

Máximo **30 segundos** de explicación por pregunta, y solo si la falló más de un tercio:

| Si falla | Decir |
|---|---|
| `class` vs `[class.x]` | «No compiten, se suman. `class` pone lo que va siempre, `[class.x]` lo que va a veces.» |
| `(click)="contador + 1"` | «La expresión se evalúa y se tira. Interpolación y expresiones **leen**; para escribir hace falta una asignación o un método.» |
| `imports` del standalone | «Si el template usa algo de Angular, va en `imports`. Hoy vamos a agregar el primer caso nuevo: **un componente propio**.» |

Si algo necesita más de 30 segundos, va a la tarea. **No te enganches acá.**

---

## 0:12 · El concepto — 8 min

> **El editor está cerrado.** No hay VS Code en pantalla, no hay terminal. Solo diapositivas.

### 0:12 — El problema, con números · 2 min

**En pantalla:** diapositiva 5.

> «El listado de la clase pasada tiene, adentro del `@for`, unas veinte líneas de HTML: el botón, las clases de estado, el nombre, la hora, el pie con los competidores y el favorito.»
>
> «Copiarlas a la página de inicio no es difícil. El problema es el mes que viene, cuando alguien pida agregar la distancia de la carrera. Hay que acordarse de los dos lugares. Y en seis meses son cinco lugares, y siempre hay uno que quedó viejo.»
>
> «El HTML repetido no es feo: **es un compromiso de mantenimiento que nadie firmó**.»

### 0:14 — Las dos puertas · 3 min

**En pantalla:** diapositiva 6 — `diagramas/datos-bajan-avisos-suben.svg`.

**Los términos que se definen acá:** *padre*, *hijo*, *`input()`*, *`output()`*.

> «Un componente que se usa adentro de otro tiene exactamente dos puertas, y la clase de hoy es saber qué pasa por cada una.»

Señalá la flecha que baja:

> «**Los datos bajan.** El padre le pasa al hijo lo que necesita para dibujarse: la carrera, la hora ya formateada, si está abierta. Eso es `input()`, y es **de solo lectura**: el hijo mira, no toca.»

Señalá la flecha que sube:

> «**Los avisos suben.** El hijo no decide nada importante: cuando lo tocan, avisa. Eso es `output()`. Y el padre decide qué hacer, porque el padre es el único que sabe qué pasa con las otras siete tarjetas.»

Y ahora la frase que quiero que se lleven:

> «**El hijo nunca modifica lo que le prestaron. Pide, y el padre decide.**»

**La pregunta que van a hacer, y hay que contestarla acá:**

| Preguntan | Respondé |
|---|---|
| «¿Y por qué el hijo no cambia el dato y listo?» | «Porque entonces habría dos lugares donde se decide lo mismo, y tarde o temprano dicen cosas distintas. Con una sola dirección, cuando algo está mal sabés dónde mirar.» |

### 0:17 — El hueco · 3 min

**En pantalla:** diapositiva 7, el mismo diagrama.

**El término que se define acá:** *proyección de contenido*, *`ng-content`*.

> «Hay una tercera cosa, y no es un dato: es **marcado**.»
>
> «La tarjeta de una carrera tiene una pastilla de estado arriba. Podría ser un `input()` de texto… hasta que alguien quiere poner un icono al lado. Y después un contador. Y cada vez hay que agregarle un input nuevo al hijo, que ya son seis.»

Señalá el hueco punteado del diagrama:

> «`ng-content` da vuelta el problema. El hijo dice *acá hay un hueco* y el padre mete lo que quiera adentro de la etiqueta. El hijo **no sabe qué es** y no lo puede leer: solo le reserva el lugar.»
>
> «Y si hay más de un hueco, se los distingue con `select`. Uno para la pastilla, otro para lo que sea.»

> **Si vas tarde:** de este bloque se puede recortar la lista de preguntas. El diagrama y la frase de «los datos bajan, los avisos suben», no.

---

## 0:20 · Live coding — 15 min

**Decilo textual antes de empezar:**

> «Cierren el editor. Los próximos quince minutos yo escribo y ustedes miran. No copien: van a hacer esto mismo después.»

**En pantalla:** VS Code y el navegador lado a lado. Proyecto: **`lab/demo`**, ruta `/s02`.

> **Antes de entrar al aula:** `node scripts/prep-demo.mjs` y después `cd lab/demo && npm start`.
>
> La pantalla de S2 **ya está escrita, con todo adentro de un componente**. Eso es a propósito: hoy no se construye una pantalla nueva, se parte una en dos. La secuencia de tecleo completa está en **`mision-profe.md`**.

### 0:20 — Nace el hijo, vacío · 2 min

```bash
ng generate component sessions/s02/coffee-card --flat
```

Cortá el `<article class="card">` entero del padre y pegalo en el template del hijo.

**Guardá. Se rompe todo**, y está bien:

> «Diecisiete errores. Todos dicen lo mismo: `item` no existe acá. Claro que no: `item` era del `@for` del padre. **Esta es la parte que importa de la clase** — cada error es una cosa que el hijo necesita de afuera, y cada una va a ser un input.»

### 0:22 — `input.required()` · 3 min

```ts
readonly coffee = input.required<Coffee>();
```

Reemplazá `item.coffee` por `coffee()` en el template del hijo.

> «`input()` no es una propiedad: es una **función que devuelve el valor**. Por eso en el template va con paréntesis. Es la misma forma que van a ver en la clase que viene con signals, y no es casualidad.»

Usalo en el padre sin pasarle nada, a propósito:

```html
<app-coffee-card />
```

> 🔴 **Rotura deliberada 1.**

```
NG8008: Required input 'coffee' from component CoffeeCardComponent must be specified.
```

> «`required` quiere decir que el compilador te lo exige. Es el mismo tipo de promesa que los tipos de S0: no es una convención escrita en un comentario, es un error de compilación.»

Ahora sí, pasáselo: `[coffee]="item.coffee"`.

### 0:25 — `input()` opcional y `output()` · 4 min

```ts
readonly featured = input(false);
readonly ordered = output<OrderRequest>();
```

> «El opcional lleva su valor por defecto adentro de los paréntesis. Sin `[featured]`, vale `false` y no hay que escribir nada.»

En el hijo, donde antes llamaba al método del padre:

```ts
protected order(): void {
  if (!this.coffee().available) return;
  this.ordered.emit({ coffee: this.coffee(), quantity: this.quantity() });
}
```

Y en el padre:

```html
(ordered)="take($event)"
```

> «Los paréntesis son los mismos de `(click)` de la clase pasada. La diferencia es que `click` lo inventó el navegador y `ordered` lo inventé yo, hace treinta segundos.»
>
> «Y `$event` acá **no es un evento del DOM**: es exactamente el objeto que puse en `emit()`, con su tipo. Pasá el mouse por encima: dice `OrderRequest`.»

**Ahora la rotura que más cuesta encontrar en un proyecto de verdad.** Borrá el `(ordered)` del padre.

> 🔴 **Rotura deliberada 2.** Tocá «Pedir» en el navegador. **No pasa nada. Y no hay ningún error.**

> «Nadie está escuchando. El hijo emite igual, al vacío. Esto compila, pasa el build, se despliega — y el botón no hace nada. Guárdense esta sensación para el bloque de predicciones.»

Volvé a poner el `(ordered)`.

### 0:29 — `model()` · 2 min

```ts
readonly quantity = model(1);
```

En el padre: `[(quantity)]="item.quantity"`.

> «`model()` es `input()` y `output()` en la misma línea. Habilita los corchetes con paréntesis adentro sobre **un componente propio**, que es lo que la clase pasada solo se podía hacer con `ngModel` sobre un input de HTML.»
>
> «Y fijate el detalle: el hijo lo cambia con `quantity.set(…)`, no asignando. Es un signal, y en la clase que viene vamos a ver qué significa eso.»

### 0:31 — `ng-content` · 2 min

En el hijo, donde iba el rótulo del café del día:

```html
<div class="card__tag">
  <ng-content select="[card-tag]" />
</div>
```

Y en el padre, adentro de la etiqueta:

```html
<app-coffee-card …>
  <span card-tag class="tag">Café del día</span>
</app-coffee-card>
```

> «Todo lo que el padre escribe entre la etiqueta que abre y la que cierra **no lo dibuja el padre**: viaja adentro del hijo y cae en el hueco.»

Agregá el segundo hueco, el que no tiene `select`, y proyectá el aviso de sin stock.

> «El que no lleva `select` es el cajón de sastre: recibe todo lo que no matcheó con ningún otro. Va **uno solo** por componente, y en un rato vamos a ver qué pasa si ponés dos.»

### 0:33 — El ciclo de vida · 2 min

```ts
ngOnInit(): void { … }
ngOnChanges(): void { this.changes += 1; }
ngOnDestroy(): void { this.destroyed.emit(this.coffee().name); }
```

> «Tres momentos, y con estos tres alcanza para todo el curso.»

| Gancho | Cuándo | Para qué |
|---|---|---|
| `ngOnInit` | una vez, con los inputs ya llenos | preparar lo que depende de los inputs |
| `ngOnChanges` | cada vez que el padre cambia un input | reaccionar a un dato nuevo |
| `ngOnDestroy` | cuando se va de la pantalla | soltar lo que quedó abierto |

> «Y el que más importa a largo plazo es el último. Hoy solo avisa un nombre; en la clase 6 va a ser lo que evita que la aplicación se coma la memoria.»

**Tocá el botón que saca el Antigua de la carta** y mostrá el aviso apareciendo en el panel.

> «Miren: el componente se destruyó y avisó al irse. Nadie lo llamó — Angular llama a ese método por vos.»

---

## 0:35 · Misión 1 — 15 min

**En pantalla:** diapositiva 17. Enunciado en `mision-estudiante-1.md`.

> «Ahora ustedes, en `lab/starter`. La pantalla está hecha y funciona. El trabajo es partirla en dos: sacar la tarjeta a su propio componente y darle las cuatro puertas.»

**Decí esto antes de largar, porque es la trampa del ejercicio:**

> «El orden que menos duele es: primero cortás y pegás el HTML, **dejás que se rompa todo**, y después cada error te dice qué input falta. No trates de adivinar los inputs antes de cortar.»
>
> «Y acordate de que un componente propio también va en `imports`. Es lo mismo que `FormsModule` de la clase pasada.»

> **Estás en silencio.**

**Reloj de pistas** — solo si más de la mitad está trabada en lo mismo:

| Min | Pista, en voz alta, sin resolver |
|---|---|
| 0:43 | «Si el hijo no aparece, mirá `imports` del padre. Es el mismo error de la clase pasada, con otro nombre.» |
| 0:47 | «En el template del hijo, un `input()` se lee **con paréntesis**. `coffee` es la función; `coffee()` es el café.» |

---

## 0:50 · Comparten pantalla — 10 min

Dos personas. **Una que le funciona y una que no** — a la segunda pedile permiso antes.

> **Preguntás, no corregís.**

1. «¿Cómo decidiste qué se llevaba el hijo y qué se quedaba?»
2. «¿Qué error te apareció primero, y qué te dijo?»
3. «Si mañana hay que mostrar un café en otra pantalla, ¿qué tocás?»
4. «¿Por qué esto es `input` y aquello es `ng-content`?»

**Lo más probable que aparezca:** alguien se llevó la comanda al hijo, y ahora cada tarjeta tiene su propia lista de pedidos.

> «No está mal escrito: está mal repartido, y es el error más instructivo del día. Fijate qué pasa si pedís dos cafés distintos — hay dos comandas y ninguna tiene todo. La comanda es de la pantalla, no de la tarjeta.»

---

## 1:00 · Descanso — 10 min

> «Diez minutos. Vuelvan puntuales.»

---

## 1:10 · Predice y ejecuta — 15 min

**Los archivos:** `predice-y-ejecuta/`. **Las respuestas:** `predice-y-ejecuta/respuestas.md`, verificadas contra el compilador — no las improvises.

**El orden no se saltea:** mostrar → predecir 60 segundos → ejecutar → explicar.

| Min | Snippet | Casi todos predicen | Pasa |
|---|---|---|---|
| 1:10 | `<app-coffee-card />` sin `[coffee]` | «anda con undefined» | **No compila.** `NG8008` |
| 1:15 | `output()` que nadie escucha | «da error» o «avisa algo» | **Compila y no pasa nada.** Silencio total |
| 1:20 | Dos `<ng-content />` iguales | «se duplica el contenido» | **Compila, y aparece una sola vez** |

Cerrá el bloque con esta pregunta:

> «De los tres, ¿cuál les habría costado más encontrar?»

Casi siempre eligen el segundo, y tienen razón:

> «Exacto. El primero te frena el build. El tercero se ve raro en la pantalla y lo encontrás mirando. **El segundo no se ve**: el botón está, se puede tocar, no hay error en la consola, y no hace nada. Es el mismo silencio del `(click)="contador + 1"` de la clase pasada, y va a ser el mismo silencio del socket en la clase 10.»

---

## 1:25 · Misión 2, en parejas — 20 min

**En pantalla:** diapositiva 23. Enunciado en `mision-estudiante-2.md`.

> «Al proyecto de verdad. En parejas, veinte minutos: diez escribe uno y dicta el otro.»

**Tres cosas para decir antes de largar:**

> «Uno: el punto de partida es **el listado que ustedes escribieron la clase pasada**. Si les quedó a medias, la corrección de S1 lo deja andando en diez minutos y arrancan desde ahí.»
>
> «Dos: son **dos** componentes y viven en lugares distintos. `<app-badge>` es una primitiva y va en `shared/ui/` — no sabe qué es una carrera. `<app-race-card>` sí sabe, así que va en `features/races/`. Esa regla la van a usar todo el curso.»
>
> «Tres, y es la que se olvidan: cuando muevan el marcado, **muevan el CSS con él**. Un componente cuyo estilo vive en otro archivo no se puede llevar a ningún lado.»

Circulás entre las parejas. **Escuchás más de lo que hablás.**

---

## 1:45 · Code review en vivo — 10 min

Una solución de la Misión 2, con permiso. **En pantalla, al lado, `correccion.md`.**

Rúbrica, en voz alta y en este orden — ya es la del curso completa:

1. ¿`standalone: true` y `OnPush`?
2. ¿Actualiza el estado sin mutar?
3. ¿`any`, `console.log`, imports sin usar?
4. ¿El nombre dice lo que la cosa hace?
5. ¿Está en la carpeta que le toca — `shared/` no sabe qué es una carrera?

**Empezá por algo que está bien hecho.**

Y la pregunta que hace la sesión:

> «Tapá el archivo del padre un segundo. Leyendo **solo** el hijo, ¿podés decir para qué sirve y qué necesita? Si la respuesta es sí, el corte está bien hecho. Si tenés que ir a mirar quién lo usa, quedó atado.»

Y el cierre:

> «Se van a dar cuenta de que la pantalla se ve **exactamente igual** que la clase pasada. Es correcto: hoy no agregamos ni una función. Lo que cambió es que ahora hay una pieza que se puede usar en otro lado, y en la clase 10 el leaderboard la va a usar sin tocarle una línea.»

---

## 1:55 · Exit ticket y tarea — 5 min

**Exit ticket:** `exit-ticket.md`. Tres preguntas, tres minutos.

**Tarea:** `tarea.md`. **Leela en voz alta antes de cortar.**

**Y el recordatorio del apunte:**

> «`conceptos.md` tiene todo lo de hoy con los ejemplos exactos que corrimos, y los tres errores con su mensaje literal. Ténganlo abierto al lado del editor cuando hagan la tarea.»

**Y el aviso de la próxima:**

> «La clase que viene: **signals**. Van a entender por qué `coffee()` se lee con paréntesis, y vamos a hacer que el listado se filtre solo.»

---

## Después de la clase

- [ ] Leer los exit tickets. Lo confuso arranca S3.
- [ ] Revisar `wayground.csv` de **esta** sesión con lo que más falló — se corre al empezar S3.
- [ ] Aplicar la corrección de S2 al `starter/` publicado y taggear `s03`.
- [ ] Completar las notas de abajo.

### Notas de la corrida real

| | |
|---|---|
| ¿Qué bloque se pasó de tiempo? | |
| ¿Cuántos cortaron el HTML antes de pensar los inputs? | |
| ¿Qué pregunta no supe contestar? | |
| ¿Qué sacaría o agregaría? | |
