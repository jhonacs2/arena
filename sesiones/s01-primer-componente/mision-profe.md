# S1 · Misión profe — El live coding del bloque 0:20

**Solo instructor · 15 minutos · `lab/demo`**

La secuencia exacta que escribís en vivo, en orden, con lo que se dice en cada
paso y las dos roturas deliberadas. Está pensado para el segundo monitor: el
`guion.md` lleva la clase, esto lleva el teclado.

---

## Antes de entrar al aula

```bash
node scripts/prep-demo.mjs
cd lab/demo
npm start
```

`lab/demo` es una copia descartable de `lab/starter`, así que arranca en el mismo
estado en que están los alumnos: **el componente de S1 no existe.** Es el lienzo
correcto para escribirlo en vivo.

> **No trabajes en `lab/solution`.** No hay que borrarle nada a la solución de
> referencia para dar la clase — para eso está `demo/`. Y si el bloque sale mal,
> `node scripts/prep-demo.mjs` te devuelve el lienzo limpio en un segundo, así
> que se puede ensayar tres veces.

**En pantalla:** VS Code y el navegador lado a lado, el navegador en
<http://localhost:4200>.

**Decilo textual antes de empezar:**

> «Cierren el editor. Los próximos quince minutos yo escribo y ustedes miran. No
> copien: van a hacer esto mismo después, y con las manos libres se entiende
> mejor. Si me equivoco, avisen.»

---

## 0:20 — El CLI crea el componente · 2 min

```bash
ng generate component sessions/s01 --flat
```

> «Esto es el **CLI**, la herramienta de línea de comandos de Angular. Le pedí un
> componente y me creó cuatro archivos: el TypeScript, el HTML, el CSS y uno de
> tests.»
>
> «Podría haberlos creado a mano. Uso el CLI porque los nombra igual siempre y
> porque no me olvido de ninguno. **No hace nada mágico**: mírenlos, están
> vacíos.»

Abrí el `.ts` y leelo en voz alta, línea por línea:

> «`@Component` es un **decorador**: le dice a Angular “esta clase no es una clase
> cualquiera, es un componente”. Y adentro va la configuración.»
> «`selector` es el nombre de la etiqueta con la que se usa.»
> «`standalone: true` es lo que hablamos: se declara solo.»
> «`templateUrl` y `styleUrl` apuntan a los otros dos archivos.»

## 0:22 — Que aparezca en pantalla · 2 min

Dos archivos, y hay que tocar los dos:

```ts
// app.routes.ts
{
  path: 's01',
  loadComponent: () => import('./sessions/s01/s01.component').then((m) => m.S01Component),
},
```

```ts
// sessions.ts
{ id: 's01', title: 'Primer componente', available: true },
```

Navegá a `/s01`.

> «Fíjense que hicieron falta **dos cosas**: la ruta, para que la dirección
> exista, y el índice, para que aparezca en el menú de la izquierda. Son dos
> archivos distintos y hay que tocar los dos. Es el error más común de la clase
> que viene.»

Señalá la página con el texto que dejó el CLI:

> «Eso es lo mínimo que existe. De acá para arriba, todo lo que aparezca lo vamos
> a poner nosotros.»

## 0:24 — Interpolación · 3 min

En la clase:

```ts
protected coffee = { name: 'Yirgacheffe', origin: 'Etiopía', price: 42, available: true };
```

En el template:

```html
<h2>{{ coffee.name }}</h2>
<p>{{ coffee.origin }}</p>
<p>{{ coffee.price }}</p>
```

> «Las llaves dobles dicen: “Angular, poné acá el valor de esta expresión, como
> texto”.»

**La demostración que vale por toda la explicación:** cambiá `price: 42` a
`price: 55` y guardá.

> «Cambié **un solo lugar** —la clase— y la pantalla se actualizó sola. Eso es el
> trato: el dato vive en un lugar, y el template lo mira. Nunca más vas a ir a
> buscar un elemento para escribirle el número.»

## 0:27 — Property binding · 3 min · **rotura deliberada 1**

> «Ahora quiero que la card se vea distinta cuando no hay stock. La clase CSS ya
> existe, se llama `product--soldout`.»

Escribí primero **el error**, a propósito:

```html
<div class="product product--soldout">
```

> «Así queda puesta siempre, y no es lo que quiero. Necesito que dependa de un
> dato.»

Y ahora la forma correcta:

```html
<div class="product" [class.product--soldout]="!coffee.available">
```

> «Los corchetes cambian todo. Sin corchetes, lo que va entre comillas es **texto
> literal**. Con corchetes, es **una expresión de TypeScript** que Angular
> evalúa.»
>
> «Y fíjense que puedo tener las dos cosas en el mismo elemento: `class` con lo
> que va siempre, y `[class.algo]` con lo que va a veces. **No se pisan, se
> suman.**»

Cambiá `available` a `false` y mostrá el resultado.

## 0:30 — Event binding · 3 min

> «Hasta acá la información va en un solo sentido: de la clase al template. Ahora
> el otro.»

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

> «Los paréntesis escuchan un evento del navegador. `click` no lo inventó
> Angular: es el mismo de siempre. Lo que agrega Angular es que en vez de escribir
> un `addEventListener`, decís qué método llamar.»

Tocá el botón. Andá y volvé un par de veces.

> «Y acá está lo de la detección de cambios: **tocaron el botón, entonces pasó
> algo, entonces Angular revisó**. Por eso el texto del botón y la clase del div
> se actualizaron los dos, sin que yo tocara nada más.»

> Si alguien pregunta por el `{ ...this.coffee, ... }`: «creo un objeto nuevo en
> vez de modificar el que estaba. Es una regla del curso y en la clase 3 van a ver
> por qué importa.» **No te extiendas más que eso.**

## 0:33 — Two-way binding · 4 min · **rotura deliberada 2**

> «Falta el cuarto: cuando la información tiene que ir en los dos sentidos. El
> caso típico es un input.»

En la clase:

```ts
protected customer = '';
```

En el template:

```html
<input type="text" [(ngModel)]="customer" />
<p>Hola, {{ customer }}</p>
```

**Guardá. Va a fallar.** Leé el error completo en voz alta, del navegador o de la
terminal:

```
NG8002: Can't bind to 'ngModel' since it isn't a known property of 'input'.
```

> «Buenísimo que falle. Léanlo conmigo: *no puedo enlazar a `ngModel` porque no es
> una propiedad conocida de `input`*.»
>
> «Y tiene razón: `ngModel` no existe en HTML. Es algo que trae Angular. Y acá
> está la parte importante — **este componente es standalone, o sea que declara
> solo lo que usa**. Yo nunca le dije que iba a usar `ngModel`.»

Agregá el import:

```ts
import { FormsModule } from '@angular/forms';
// …
imports: [FormsModule],
```

> «Ahora sí. Y quiero que se lleven la regla completa: **si el template usa algo
> de Angular, tiene que estar en `imports`.** Vale para esto, para el router, y
> desde la clase que viene, para cada componente que usen adentro de otro.»

Escribí en el input y mostrá cómo el `<p>` de abajo cambia mientras escribís.

> «Los dos sentidos: lo que escribo va a la clase, y lo que está en la clase se ve
> abajo. `[(ngModel)]` es literalmente `[ngModel]` más `(ngModelChange)` —
> property binding y event binding juntos, con azúcar sintáctica.»

---

## Al terminar el bloque

El resultado tiene que quedar **en pantalla** durante el ejercicio 1: es la
referencia visual de lo que van a construir. No cierres el navegador.

Si querés el lienzo limpio otra vez —para ensayar, o porque el bloque se
descarriló— `node scripts/prep-demo.mjs` y volvés al estado inicial.

## Qué hacer si te quedás sin tiempo

El orden de sacrificio, de lo primero que se recorta a lo último:

1. **La rotura deliberada 1** (el `class` puesto siempre). Se puede contar en vez
   de escribirla.
2. **El `{ ...this.coffee }`**: dejá el objeto mutado y avisá que en S3 se corrige.
3. **Nunca recortes la rotura 2** (NG8002). Ese error es el que más les va a
   pasar en la Misión 1, y verlo acá primero es la mitad del valor del bloque.
