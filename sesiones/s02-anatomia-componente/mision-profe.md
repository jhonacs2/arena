# S2 · Misión profe — El live coding del bloque 0:20

**Solo instructor · 15 minutos · `lab/demo`**

La secuencia exacta que escribís en vivo, con las dos roturas deliberadas.
Segundo monitor: el `guion.md` lleva la clase, esto lleva el teclado.

---

## Antes de entrar al aula

```bash
node scripts/prep-demo.mjs
cd lab/demo
npm start
```

Y **sumá la ruta `/s02` a mano antes de empezar**, porque declararla es tarea de
los alumnos y hoy no es el tema:

```ts
// lab/demo/src/app/app.routes.ts
{
  path: 's02',
  title: 'S2 · Anatomía de un componente · Lab',
  loadComponent: () => import('./sessions/s02/s02.component').then((m) => m.S02Component),
},
```

Más `available: true` en `sessions.ts`. Treinta segundos, y evita empezar la
clase peleando con una ruta.

**En pantalla:** VS Code y el navegador lado a lado, en
<http://localhost:4200/s02>. Se ve la carta de cuatro cafés, funcionando.

**Decilo textual antes de empezar:**

> «Cierren el editor. Los próximos quince minutos yo escribo y ustedes miran. No
> copien: van a hacer esto mismo después.»

---

## 0:20 — Nace el hijo, vacío · 2 min

```bash
ng generate component sessions/s02/coffee-card --flat
```

> «Mismo comando de la clase pasada. Cuatro archivos, vacíos.»

Ahora **cortá** —cortá, no copies— el `<article class="card">` entero del
template del padre, y pegalo en `coffee-card.component.html`.

**Guardá.**

> 🔴 **Rotura deliberada 1.** La terminal se llena de errores. Contalos en voz
> alta.

> «Diecisiete errores, y todos dicen lo mismo: `item` no existe acá. Claro que
> no: `item` era la variable del `@for` del padre, y este archivo es otro
> componente.»
>
> «**Esta es la parte que importa de la clase.** Cada error es una cosa que el
> hijo necesita de afuera. No hay que adivinar los inputs: la lista ya está
> escrita, en la terminal.»

---

## 0:22 — `input.required()` · 3 min

En `coffee-card.component.ts`:

```ts
import { Component, ChangeDetectionStrategy, input } from '@angular/core';
import type { Coffee } from './menu';

readonly coffee = input.required<Coffee>();
```

Y en el template del hijo, reemplazá `item.coffee` por `coffee()`.

> «Fíjense en los paréntesis. `input()` no devuelve un dato: devuelve una
> **función** que da el dato. Por eso en el template va `coffee().name` y no
> `coffee.name`.»
>
> «Es la misma forma que van a ver la clase que viene con signals, y no es
> casualidad: es exactamente lo mismo.»

Ahora usalo en el padre, **mal a propósito**:

```html
<app-coffee-card />
```

(Y acordate de sumar `CoffeeCardComponent` a los `imports` del padre, o el error
que sale es el otro.)

```
NG8008: Required input 'coffee' from component CoffeeCardComponent must be specified.
```

> «`required` no es documentación: es el compilador exigiéndotelo. El mismo tipo
> de promesa que los tipos de la sesión 0.»

Arreglalo:

```html
<app-coffee-card [coffee]="item.coffee" />
```

---

## 0:25 — `input()` opcional y `output()` · 4 min

```ts
readonly featured = input(false);
readonly ordered = output<OrderRequest>();
```

> «El opcional lleva el valor por defecto adentro de los paréntesis. Si el padre
> no escribe `[featured]`, vale `false`.»

El método del hijo:

```ts
protected order(): void {
  if (!this.coffee().available) return;
  this.ordered.emit({ coffee: this.coffee(), quantity: this.quantity() });
}
```

En el padre:

```html
<app-coffee-card [coffee]="item.coffee" (ordered)="take($event)" />
```

**Pasá el mouse por encima de `$event`** y mostrá el tooltip:

> «`$event` acá no es un evento del DOM: es exactamente el objeto que puse en
> `emit()`. Mírenlo — dice `OrderRequest`. El tipo viaja con el aviso.»

**Ahora la rotura que más cuesta encontrar.** Borrá el `(ordered)="take($event)"`
del padre y andá al navegador. Tocá «Pedir».

> 🔴 **Rotura deliberada 2.** No pasa nada. Dejá el silencio tres segundos.

> «Nada. No hay error en la consola, no hay error en la terminal, el botón se
> aprieta y se hunde. El hijo está emitiendo al vacío, porque nadie se suscribió.»
>
> «Guárdense esta sensación. Vuelve en el bloque de predicciones.»

Volvé a poner el `(ordered)`.

---

## 0:29 — `model()` · 2 min

```ts
readonly quantity = model(1);
```

En el hijo, los botones de cantidad:

```ts
protected add(step: number): void {
  const next = this.quantity() + step;
  if (next >= 1 && next <= 20) this.quantity.set(next);
}
```

En el padre:

```html
[(quantity)]="item.quantity"
```

> «Los corchetes con paréntesis adentro, sobre un componente **mío**. La clase
> pasada eso solo se podía con `ngModel`, sobre un `<input>` de HTML.»
>
> «Y miren cómo escribe el hijo: `quantity.set(…)`, no `quantity = …`. Es un
> signal. La clase que viene es entera sobre eso.»

Mostralo en el navegador: subí la cantidad y bajala.

---

## 0:31 — `ng-content` · 2 min

En el hijo, arriba de todo:

```html
<div class="card__tag">
  <ng-content select="[card-tag]" />
</div>
```

En el padre, **adentro** de la etiqueta:

```html
<app-coffee-card [coffee]="item.coffee" …>
  @if (item.coffee.id === featuredId) {
    <span card-tag class="tag">Café del día</span>
  }
</app-coffee-card>
```

> «Todo lo que el padre escriba entre la etiqueta que abre y la que cierra no lo
> dibuja el padre: **viaja adentro del hijo** y cae en el hueco que le
> corresponde.»

Y el segundo hueco, el sin `select`:

```html
<div class="card__slot">
  <ng-content />
</div>
```

> «Este es el cajón de sastre: recibe todo lo que no matcheó con ningún otro. Va
> **uno solo** por componente. Qué pasa si ponés dos lo vemos a las 1:20.»

---

## 0:33 — El ciclo de vida · 2 min

```ts
export class CoffeeCardComponent implements OnInit, OnChanges, OnDestroy {
  ngOnInit(): void {
    this.mountedAt = /* la hora */;
  }

  ngOnChanges(): void {
    this.changes += 1;
  }

  ngOnDestroy(): void {
    this.destroyed.emit(this.coffee().name);
  }
}
```

> «Tres momentos. Con estos tres alcanza para todo el curso.»

Mostralos funcionando, en este orden:

1. **Recargá.** Cada tarjeta dice la hora en que se montó, y `ngOnChanges ×1`.
2. **Subí una cantidad.** El contador pasa a **×2**. Preguntá por qué, y esperá:

   > «El cambio lo hizo el hijo. ¿Por qué corre `ngOnChanges`, que es el gancho
   > de *cambió un input*?»
   >
   > «Porque `model()` es entrada **y** salida. El hijo emite, el padre guarda el
   > valor nuevo en `item.quantity`, y ese valor **vuelve a bajar** por el mismo
   > binding. Da toda la vuelta. Para el hijo es un input que cambió, y tiene
   > razón: cambió.»

3. **Tocá «Sacar el Antigua de la carta».** Aparece el aviso en el panel de
   ciclo de vida.

> «Nadie llamó a `ngOnDestroy`. Angular lo llamó por vos, justo antes de sacarlo
> de la pantalla. Hoy solo avisa un nombre; en la clase 6 va a ser lo que evita
> que la aplicación se coma la memoria.»

---

## Orden de sacrificio

Si a las 0:30 vas por el paso de las 0:25, recortá en este orden:

| | Qué se saca | Por qué se puede |
|---|---|---|
| 1.º | El ciclo de vida de **0:33** | Está en `conceptos.md` §7 y en el enunciado como requisito aparte |
| 2.º | `model()` de **0:29** | Se puede contar en treinta segundos sin escribirlo |
| 3.º | El segundo `ng-content` de **0:31** | Con el `select` ya se entendió la idea |

**Lo que no se sacrifica nunca:** el corte de las **0:20** con sus diecisiete
errores, y la rotura del `output()` sin escuchar de las **0:25**. Son las dos
que sostienen la Misión 1 y el bloque de predicciones.

---

## Si algo sale mal

| Pasa | Qué hacer |
|---|---|
| El hijo no aparece en pantalla | Falta `CoffeeCardComponent` en los `imports` del padre. Es el error número uno y conviene que pase. |
| `coffee.name` no compila | Falta el paréntesis: `coffee().name`. |
| Quedó todo hecho un desastre | `node scripts/prep-demo.mjs` en otra terminal y `demo/` vuelve a cero. Perdés el paso, no la clase. |
