# S4 · Misión profe — El live coding del bloque 0:20

**Solo instructor · 15 minutos · `lab/demo`**

---

## Antes de entrar al aula

```bash
node scripts/prep-demo.mjs
cd lab/demo
npm start
```

Suma la ruta `/s04` a mano, más `available: true` en `sessions.ts`.

**Y quita las dos líneas del idioma de `app.config.ts`** —`registerLocaleData` y
el proveedor de `LOCALE_ID`—, porque el bloque de las 0:20 las agrega en vivo. Es
el único preparativo que no es mecánico; si te lo olvidas, el porcentaje ya sale
bien y se pierde la demostración.

**En pantalla:** VS Code y el navegador en <http://localhost:4200/s04>.

---

## 0:20 — Los pipes que ya vienen · 3 min

En el template, sin tocar el componente:

```html
<p class="card__origin">{{ coffee.origin | uppercase }}</p>
```

> «Ese es todo el trato de un pipe: **transforma lo que se ve, no lo que hay**.
> El dato sigue siendo “Etiopía”; si mañana hace falta para buscar, está intacto.»

Agrega el número y el porcentaje:

```html
<p class="card__stock">
  {{ coffee.stock | number }} en depósito · {{ share(coffee) | percent: '1.0-1' }} del total
</p>
```

**Y ahora mira la pantalla y quédate callado un segundo.**

> «Sale `54.5%`. Con punto. Y toda la pantalla está en español.»
>
> «Los pipes incorporados formatean según el idioma de la aplicación, y ese
> idioma **por defecto es `en-US`**. No es un error de Angular: es que nadie le
> dijo en qué idioma está esto.»

En `app.config.ts`:

```ts
import { registerLocaleData } from '@angular/common';
import localeEs from '@angular/common/locales/es';

registerLocaleData(localeEs);

providers: [{ provide: LOCALE_ID, useValue: 'es' }, …]
```

> «Dos líneas, una sola vez en la vida del proyecto. Ahora dice `54,5 %`.»

---

## 0:23 — Un pipe propio · 4 min

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

> «`transform` es el único método que hace falta. El primer argumento es el valor
> que viene por la tubería; lo que devuelve es lo que se ve.»

**Úsalo sin declararlo**, a propósito:

```html
<p class="card__price num">{{ coffee.price | money }}</p>
```

> 🔴 **Rotura deliberada 1.**

```
NG8004: No pipe found with name 'money'.
```

> «*No encontré ningún pipe que se llame money.* Y tiene razón: lo escribí, pero
> no le dije a **este** componente que lo iba a usar. Es la misma regla de
> `FormsModule` en la primera clase y de los componentes en la segunda.»

Agrégalo a `imports`. Funciona.

Ahora el parámetro:

```html
<p class="card__alt num">{{ coffee.price | money: 'USD' }}</p>
```

> «Lo que va después de los dos puntos es el segundo argumento de `transform`.»

**Y si alguien pregunta por el `useGrouping`,** vale la pena contestarlo porque
es real:

> «En español, el formato por defecto **no agrupa los números de cuatro cifras**:
> 4200 se escribe “4200”. Para un número suelto es correcto; para un importe, no.
> Es exactamente el tipo de detalle que un pipe deja resuelto en un solo lugar en
> vez de en veinte templates.»

---

## 0:27 — Puro contra impuro · 3 min

**Baja hasta el panel de contadores.** Ya está en la pantalla, no hay que
escribir nada.

Toca «Provocar una detección de cambios» cinco o seis veces, despacio.

> «Los dos pipes hacen exactamente lo mismo: devuelven el texto tal cual. Se
> diferencian en **una palabra**.»
>
> «El puro se quedó en uno. El impuro sube con cada clic.»

Muestra las dos declaraciones, una al lado de la otra:

```ts
@Pipe({ name: 'countPure', standalone: true, pure: true })
@Pipe({ name: 'countImpure', standalone: true, pure: false })
```

> «Un pipe **puro** —que es el valor por defecto, no hace falta escribirlo—
> Angular lo llama solo cuando **cambia el valor de entrada**. Si es el mismo,
> reutiliza el resultado anterior sin ejecutar nada.»
>
> «Uno **impuro** corre cada vez que Angular revisa este componente.»

Y el matiz que conviene decir, porque vuelve a las 1:20:

> «Fíjate que dije *revisa este componente*, no *en cada clic de la aplicación*.
> Este componente es `OnPush`, así que solo se revisa cuando pasa algo adentro.
> Si fuera de detección por defecto, el número subiría con cualquier clic de
> cualquier parte.»

**La regla:**

> «Impuro se usa cuando el resultado depende de algo que el valor de entrada no
> ve — el reloj, por ejemplo. Casi siempre hay una forma mejor, y casi siempre
> esa forma es un `computed` de la clase pasada.»

---

## 0:30 — Una directiva de atributo · 3 min

```bash
ng generate directive sessions/s04/highlight
```

Antes de escribir nada, **muestra lo que hay en el template**:

```html
<li
  class="card"
  [class.is-highlighted]="coffee.id === featuredId"
  [attr.data-highlight-label]="coffee.id === featuredId ? 'Café del día' : null"
>
```

> «La misma condición, escrita dos veces. Y el nombre de la clase CSS decidido
> por la pantalla.»

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

> «El selector entre corchetes quiere decir *cualquier elemento que tenga este
> atributo*. Y el input se llama igual que el selector, así se escribe una sola
> vez en vez de dos.»
>
> «`host` es donde la directiva declara qué le hace al elemento del que cuelga.
> Reemplaza a `@HostBinding` y `@HostListener`, que es lo que vas a ver en código
> más viejo.»

El template queda así:

```html
<li class="card" [appHighlight]="coffee.id === featuredId" highlightLabel="Café del día">
```

> «Y fíjate qué desapareció: **el nombre de la clase CSS**. La pantalla ya no
> decide cómo se ve un elemento destacado; solo dice cuál lo está.»

**Ahora quita el `standalone: true`:**

> 🔴 **Rotura deliberada 2.**

```
TS-992011: The directive 'HighlightDirective' appears in 'imports', but is not standalone
and cannot be imported directly. It must be imported via an NgModule.
```

> «El mismo error que tendrían con un componente. En Angular 18 se escribe
> siempre, en los tres: componentes, directivas y pipes.»

Vuelve a ponerlo.

---

## 0:33 — La estructural, para leerla · 2 min

**No la escribas. Ábrela y muéstrala.**

```html
<span class="bean" *appRepeat="coffee.rating">●</span>
```

Y al lado, las dos formas equivalentes:

```html
<span *appRepeat="3">●</span>

<ng-template [appRepeat]="3">
  <span>●</span>
</ng-template>
```

> «Son **exactamente lo mismo**. El asterisco es azúcar sintáctica: Angular
> reescribe la primera en la segunda antes de compilar.»
>
> «Un `ng-template` es un pedazo de HTML que **no se dibuja**: queda guardado, y
> alguien decide después si se pinta, cuántas veces y dónde. Eso es lo que hacían
> `*ngIf` y `*ngFor` antes de que existieran `@if` y `@for`.»

Cambia un `rating` en los datos y muestra los puntos apareciendo.

> «En Angular 18 casi nunca hace falta escribir una propia: el control flow de la
> clase pasada cubre los casos comunes y se lee mucho mejor. Está aquí para que
> puedan **leer** el código que las usa, que todavía es la mayoría.»

---

## Orden de sacrificio

| | Qué se saca | Por qué se puede |
|---|---|---|
| 1.º | La estructural de **0:33** | Es de lectura; está entera en `conceptos.md` §6 |
| 2.º | El `LOCALE_ID` de **0:20** | Se puede contar en treinta segundos sin escribirlo |
| 3.º | El parámetro del pipe de **0:23** | Está en el enunciado como requisito aparte |

**Lo que no se sacrifica nunca:** el `NG8004` de las **0:23** y el panel de puro
contra impuro de las **0:27**. Son los dos que sostienen el bloque de
predicciones.

---

## Si algo sale mal

| Pasa | Qué hacer |
|---|---|
| `NG8004` cuando ya lo declaraste | El `name` del `@Pipe` no coincide con lo que escribiste en el template. Se comparan cadenas, no clases. |
| La directiva no hace nada | Falta declararla en `imports`, o el selector no coincide: entre corchetes y con el mismo prefijo. |
| El porcentaje sigue en inglés | `registerLocaleData` va **fuera** del objeto de configuración, en el cuerpo del módulo. |
| Quedó todo hecho un desastre | `node scripts/prep-demo.mjs` y `demo/` vuelve a cero. |
