# S1 · Conceptos — Primer componente y data binding

> **Para qué es este archivo.** La clase es en vivo y no queda grabada. Cuando te
> sientes a hacer la tarea, esto es lo que tenés en vez de la memoria: cada
> concepto que vimos, con su definición y **los ejemplos exactos que corrimos en
> clase**. No es un resumen decorativo — está pensado para tenerlo abierto al
> lado del editor.

**Índice**

1. [El problema: por qué existe un framework](#1-el-problema-por-qué-existe-un-framework)
2. [El DOM](#2-el-dom)
3. [Componente, clase y template](#3-componente-clase-y-template)
4. [El decorador `@Component`](#4-el-decorador-component)
5. [Standalone e `imports`](#5-standalone-e-imports)
6. [Las cuatro ataduras (data binding)](#6-las-cuatro-ataduras-data-binding)
7. [Detección de cambios](#7-detección-de-cambios)
8. [Dos archivos para que una pantalla exista](#8-dos-archivos-para-que-una-pantalla-exista)
9. [Los errores de hoy y qué significan](#9-los-errores-de-hoy-y-qué-significan)
10. [Glosario](#10-glosario)

---

## 1. El problema: por qué existe un framework

Antes de Angular, para que un número en pantalla cambiara había que ir a buscar
el elemento y escribirle encima:

```js
let count = 0;
const output = document.querySelector('#output');

document.querySelector('#add').addEventListener('click', () => {
  count = count + 1;
  output.textContent = count;   // ← sin esta línea, el número no se mueve
});
```

**El dato y lo que se ve son dos cosas separadas, y mantenerlas iguales es tu
trabajo.** `count` ya vale 1 sin la última línea; la pantalla sigue mostrando 0.
Con un contador se sostiene. Con cuarenta campos, dos listas y un formulario, te
olvidás de una línea y la pantalla miente.

**Angular existe para que esa última línea no la escribas nunca más.**

---

## 2. El DOM

> **El DOM** (*Document Object Model*) es la representación en memoria que el
> navegador hace de tu HTML: un árbol de objetos, uno por etiqueta, que se puede
> leer y modificar desde JavaScript.

`document.querySelector('#output')` no busca en el archivo `.html` — busca en ese
árbol. Y `output.textContent = count` modifica el árbol, y el navegador repinta.

Esto importa para hoy porque **es lo que Angular hace en tu lugar**. No es magia:
es el mismo `textContent` de siempre, escrito por otro.

---

## 3. Componente, clase y template

> **Un componente** es una clase de TypeScript emparejada con un pedazo de HTML.

| | Archivo | Qué tiene |
|---|---|---|
| **La clase** | `.component.ts` | Los datos y las decisiones |
| **El template** | `.component.html` | Lo que se ve |

La clase no sabe nada de HTML. El template no toma decisiones. Entre los dos hay
**ataduras** (*bindings*), y son el tema de hoy.

![Componente y template](diagramas/componente-y-template.svg)

---

## 4. El decorador `@Component`

Esto es lo que creó el CLI, línea por línea:

```ts
import { Component, ChangeDetectionStrategy } from '@angular/core';

@Component({
  selector: 'app-s01',                    // la etiqueta con la que se usa: <app-s01>
  standalone: true,                       // se declara solo (ver §5)
  imports: [],                            // lo que usa su template
  templateUrl: './s01.component.html',    // dónde está el HTML
  styleUrl: './s01.component.css',        // dónde está el CSS
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class S01Component {}
```

> **Un decorador** es una anotación que se le pega a una clase para agregarle
> información. `@Component` le dice a Angular: «esta clase no es una clase
> cualquiera, es un componente», y adentro va la configuración.

El comando que los crea:

```bash
ng generate component sessions/s01 --flat
```

Crea cuatro archivos: el `.ts`, el `.html`, el `.css` y uno de tests. **No hace
nada mágico** — los mismos archivos se pueden crear a mano. El CLI se usa porque
los nombra igual siempre y porque no se olvida de ninguno.

> `changeDetection: OnPush` es una regla del curso desde el día uno. Por qué
> importa se ve en la sesión 10; por ahora va siempre y alcanza con eso.

---

## 5. Standalone e `imports`

> **`standalone: true`** significa que el componente **declara por sí mismo todo
> lo que su template necesita**. Nadie lo declara por él.

Esa lista de cosas que necesita es `imports`:

```ts
import { FormsModule } from '@angular/forms';

@Component({
  standalone: true,
  imports: [FormsModule],   // ← sin esto, [(ngModel)] no compila
  // …
})
```

**La regla completa, que vale para todo el curso:**

> Si el template usa algo de Angular —una directiva, un componente de otro
> archivo, el router— **tiene que estar en `imports`**.

En Angular 18 `standalone: true` **se escribe siempre**, explícitamente. (En
versiones posteriores pasó a ser el comportamiento por defecto; acá no.)

---

## 6. Las cuatro ataduras (data binding)

> **Binding** quiere decir **atadura**. Es una conexión declarada entre un dato
> de la clase y algo del template, que Angular mantiene al día por vos.

| | Sintaxis | Dirección | Ata |
|---|---|---|---|
| **Interpolación** | `{{ expresión }}` | clase → template | texto |
| **Property binding** | `[propiedad]="expresión"` | clase → template | una propiedad |
| **Event binding** | `(evento)="método()"` | template → clase | un evento del DOM |
| **Two-way** | `[(ngModel)]="propiedad"` | los dos sentidos | valor de un input |

### 6.1 Interpolación — `{{ }}`

Pone el valor de una expresión **como texto**.

```ts
protected coffee = { name: 'Yirgacheffe', origin: 'Etiopía', price: 42, available: true };
```

```html
<h2>{{ coffee.name }}</h2>
<p>{{ coffee.origin }}</p>
<p>{{ coffee.price }}</p>
```

**La demostración que hicimos en clase:** cambiar `price: 42` a `price: 55` en el
`.ts` y guardar. La pantalla se actualiza sola.

> Cambiás **un solo lugar** —la clase— y lo que se ve sigue. El dato vive en un
> lugar y el template lo mira.

### 6.2 Property binding — `[ ]`

Ata una propiedad de un elemento a una expresión.

```html
<!-- así la clase queda puesta SIEMPRE. No es lo que queremos. -->
<div class="product product--soldout">

<!-- así depende de un dato -->
<div class="product" [class.product--soldout]="!coffee.available">
```

**Los corchetes cambian todo:**

| | Qué es lo que va entre comillas |
|---|---|
| `atributo="algo"` | **texto literal**. La cadena `algo`. |
| `[propiedad]="algo"` | **una expresión de TypeScript** que Angular evalúa. |

Y las dos formas **conviven en el mismo elemento**:

```html
<div class="product" [class.product--soldout]="!coffee.available">
```

> `class` pone lo que va siempre, `[class.x]` lo que va a veces. **No se pisan,
> se suman.** El resultado en el DOM es `class="product product--soldout"`.
>
> Esto fue el primer «predice y ejecuta» de la clase, y la respuesta intuitiva
> —«gana uno de los dos»— es la equivocada.

Otras formas de property binding que vas a usar:

```html
<button [disabled]="!customer">Agregar</button>
<div [attr.aria-pressed]="isOpen">…</div>
<img [src]="horse.silkUrl" [alt]="horse.name" />
```

### 6.3 Event binding — `( )`

Escucha un evento del DOM y llama a un método.

```ts
protected toggleAvailability(): void {
  this.coffee = { ...this.coffee, available: !this.coffee.available };
}
```

```html
<button type="button" (click)="toggleAvailability()">
  {{ coffee.available ? 'Marcar agotado' : 'Marcar disponible' }}
</button>
```

> `click` no lo inventó Angular: es el evento de siempre. Lo que agrega Angular
> es que en vez de escribir `addEventListener`, **decís qué método llamar**.

Sobre el `{ ...this.coffee, ... }`: creamos un objeto **nuevo** en vez de
modificar el que estaba. Es una regla del curso; por qué importa se ve en la
sesión 3.

### 6.4 Two-way binding — `[( )]`

Para cuando la información va en los dos sentidos. El caso típico es un input.

```ts
protected customer = '';
```

```html
<input type="text" [(ngModel)]="customer" />
<p>Hola, {{ customer }}</p>
```

**Necesita `FormsModule` en `imports`.** Sin eso no compila — ver §9.

> `[(ngModel)]` es literalmente `[ngModel]` **más** `(ngModelChange)`: property
> binding y event binding juntos, con azúcar sintáctica. Los corchetes adentro de
> los paréntesis se recuerdan como *«banana in a box»*: `[()]`.

---

## 7. Detección de cambios

La pregunta que hay que poder contestar: **¿cómo se entera Angular de que un dato
cambió?**

> **Angular no vigila tus datos. Revisa después de que pasó algo.**

«Algo» es un evento del navegador: un click, una tecla, un temporizador, una
respuesta de red. Cuando eso ocurre, Angular vuelve a evaluar los bindings y
actualiza lo que haya cambiado.

Por eso, al tocar el botón del ejemplo, se actualizaron **las dos cosas** —el
texto del botón y la clase del `div`— sin que nadie las actualizara a mano: hubo
un click, y Angular revisó todo.

Esta idea es la base de `OnPush`, que se ve a fondo en la sesión 10.

---

## 8. Dos archivos para que una pantalla exista

Que el componente exista no alcanza para poder navegar a él. En el lab hacen
falta **dos** cosas, en dos archivos distintos:

```ts
// app.routes.ts — para que la dirección exista
{ path: 's01', loadComponent: () => import('./sessions/s01/s01.component').then((m) => m.S01Component) }
```

```ts
// sessions.ts — para que aparezca en el menú de la izquierda
{ id: 's01', title: 'Primer componente', available: true }
```

> Tocar uno solo es **el error más común**. Si la dirección `/s01` funciona pero
> no aparece en el menú, te falta el índice. Si aparece en el menú y al hacer
> click no pasa nada, te falta la ruta.

---

## 9. Los errores de hoy y qué significan

### NG8002 — `ngModel` no es una propiedad conocida

```
NG8002: Can't bind to 'ngModel' since it isn't a known property of 'input'.
```

**Qué dice:** no puedo enlazar a `ngModel` porque no es una propiedad conocida de
`input`.

**Y tiene razón:** `ngModel` no existe en HTML, lo trae Angular. El componente es
standalone, así que declara solo lo que usa — y nunca le dijimos que iba a usar
`ngModel`.

**Cómo se arregla:**

```ts
import { FormsModule } from '@angular/forms';
// …
imports: [FormsModule],
```

**Cuándo vuelve a aparecer:** cada vez que uses algo de Angular en el template y
te olvides de `imports`. Con `@if` y `@for` no pasa (son sintaxis del template,
no directivas importables), pero sí con cada componente que uses adentro de otro,
desde la sesión 2.

### Una expresión que no se asigna a nada

```html
<p>{{ count + 1 }}</p>
```

Esto **muestra** `count + 1`, pero no cambia `count`. La interpolación **lee**,
nunca escribe. Fue el tercer «predice y ejecuta».

### La pantalla no cambia cuando cambio el `.ts`

Casi siempre es una de tres: el servidor no está corriendo, hay un error de
compilación arriba en la terminal que tapó todo, o estás mirando el archivo de
otro proyecto (`solution/` en vez de `starter/`).

---

## 10. Glosario

| Palabra | Qué es |
|---|---|
| **DOM** | El árbol de objetos que el navegador arma a partir del HTML |
| **Componente** | Una clase de TypeScript emparejada con un pedazo de HTML |
| **Clase** | El `.ts`: los datos y las decisiones |
| **Template** | El `.html`: lo que se ve |
| **Decorador** | Una anotación pegada a una clase, como `@Component` |
| **Selector** | El nombre de etiqueta con el que se usa un componente |
| **Standalone** | Que el componente declara por sí mismo lo que necesita |
| **`imports`** | La lista de lo que el template usa |
| **Binding** | Una atadura entre un dato de la clase y algo del template |
| **Interpolación** | El binding de texto: `{{ }}` |
| **Property binding** | El binding de una propiedad: `[ ]` |
| **Event binding** | El binding de un evento: `( )` |
| **Two-way binding** | Los dos sentidos a la vez: `[( )]` |
| **Detección de cambios** | La revisión que Angular hace después de que pasó algo |
| **CLI** | La herramienta de línea de comandos de Angular (`ng`) |

---

## Para la tarea

Con esto alcanza para hacer `tarea.md` sin nada más. Si algo de acá no te cierra,
es mejor preguntarlo al empezar la sesión que viene que arrastrarlo: **todo lo de
la sesión 2 se apoya en esto.**

Lo que **no** vimos hoy y no hace falta todavía: recorrer listas (`@for`),
condicionales en el template (`@if`), pasar datos de un componente a otro
(`input()`), servicios, y HTTP. Cada uno tiene su sesión.
