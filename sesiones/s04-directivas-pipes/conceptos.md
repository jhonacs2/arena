# S4 · Conceptos — Directivas y pipes

> **Para qué es este archivo.** La clase es en vivo y no queda grabada. Cuando te
> sientes a hacer la tarea, esto es lo que tienes en vez de la memoria.

**Índice**

1. [El problema: el formato repetido](#1-el-problema-el-formato-repetido)
2. [Pipes](#2-pipes)
3. [Los pipes que ya vienen, y el idioma](#3-los-pipes-que-ya-vienen-y-el-idioma)
4. [Un pipe propio](#4-un-pipe-propio)
5. [Puro contra impuro](#5-puro-contra-impuro)
6. [Directivas](#6-directivas)
7. [Los errores de hoy](#7-los-errores-de-hoy)
8. [Glosario](#8-glosario)

---

## 1. El problema: el formato repetido

Desde S1 veníamos escribiendo esto en cada template:

```html
{{ horse.odds.toFixed(2) }}
```

Y en el componente, un método para los importes. **Ninguna de las dos es lógica
de la pantalla:** son formas de mostrar, y las formas de mostrar se repiten en
toda la aplicación.

El día que alguien pide las cuotas con coma en vez de punto, hay que encontrarlos
todos.

---

## 2. Pipes

> **Un pipe es una función con nombre que se usa desde el template para
> transformar un valor.**

```html
{{ coffee.origin | uppercase }}
```

Entra un valor, sale otro. **El dato no cambia**: si mañana hace falta el nombre
tal cual para buscar, sigue estando tal cual.

Y no sabe quién lo llamó, que es lo que lo hace reutilizable.

![Pipes y directivas](diagramas/pipes-y-directivas.svg)

---

## 3. Los pipes que ya vienen, y el idioma

Vienen con Angular y se declaran en `imports`, igual que todo lo demás:

| Pipe | Ejemplo |
|---|---|
| `uppercase` · `lowercase` · `titlecase` | `{{ 'etiopía' \| uppercase }}` → `ETIOPÍA` |
| `number` | `{{ 12345 \| number }}` → `12.345` |
| `percent` | `{{ 0.545 \| percent: '1.0-1' }}` → `54,5 %` |
| `currency` | para dinero real, con código de moneda |
| `date` | `{{ race.startsAt \| date: 'short' }}` |
| `slice` | corta arrays y textos |
| `json` | para depurar |

### El idioma no es automático

**Los pipes incorporados formatean según el idioma de la aplicación, y ese idioma
por defecto es `en-US`.**

Sin esto, `{{ 0.545 | percent }}` sale `54.5%` en una pantalla escrita entera en
español:

```ts
// app.config.ts
import { registerLocaleData } from '@angular/common';
import localeEs from '@angular/common/locales/es';

registerLocaleData(localeEs);

export const appConfig: ApplicationConfig = {
  providers: [{ provide: LOCALE_ID, useValue: 'es' }, …],
};
```

**Dos pasos, y hacen falta los dos:** registrar los datos del idioma y decir cuál
usar. Una sola vez en la vida del proyecto.

---

## 4. Un pipe propio

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

| | |
|---|---|
| `name` | con lo que se escribe en el template. **Se comparan cadenas**, no clases |
| `transform` | el único método necesario. El primer argumento es el valor |
| Los demás argumentos | son los parámetros: `{{ x \| money: 'USD' }}` |

Se encadenan:

```html
{{ nombre | lowercase | titlecase }}
```

### El detalle del `useGrouping`

En español, el formato por defecto **no agrupa los números de cuatro cifras**:
4200 se escribe «4200». Para un número suelto es lo correcto; para un importe,
no.

Es exactamente el tipo de detalle que un pipe deja resuelto en un solo lugar en
vez de en veinte templates.

---

## 5. Puro contra impuro

| | Cuándo lo llama Angular |
|---|---|
| **`pure: true`** — el valor por defecto | solo cuando **cambia el valor de entrada** |
| **`pure: false`** | cada vez que **revisa el componente** |

Lo que se vio en la pantalla de clase: dos pipes idénticos salvo por esa palabra,
contando cuántas veces los llamaron. Tras seis clics, el puro seguía en **1** y el
impuro iba por **7**.

### El matiz que importa

«Cada vez que revisa el componente» **no** es «en cada clic de la aplicación**.
El componente de la clase es `OnPush`, así que solo se revisa cuando pasa algo
adentro. Si fuera de detección por defecto, subiría con cualquier clic de
cualquier parte.

Es decir: un pipe impuro es caro, **y con detección por defecto es mucho más
caro**.

### Cuándo se usa impuro

Cuando el resultado depende de algo que el valor de entrada no ve — el reloj, por
ejemplo. Casi siempre hay una forma mejor, y casi siempre esa forma es un
`computed` de S3.

---

## 6. Directivas

> **Una directiva es una pieza sin template propio que le agrega algo a un
> elemento que ya existe.**

Y la diferencia que se pregunta siempre:

| | |
|---|---|
| **Componente** | trae su propio template. Dibuja algo nuevo |
| **Directiva** | no tiene template. Se cuelga de una etiqueta ajena |

### De atributo

Cambia cómo se ve o cómo se comporta un elemento.

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

```html
<li class="card" [appHighlight]="coffee.id === featuredId" highlightLabel="Café del día">
```

- **El selector entre corchetes** quiere decir *cualquier elemento que tenga este
  atributo*.
- **El input se llama igual que el selector**, para poder escribirlo una vez.
- **`host`** es donde la directiva declara qué le hace al elemento. Reemplaza a
  `@HostBinding` y `@HostListener`, que es lo que se ve en código más viejo.
- **`null` quita el atributo.** Una cadena vacía lo deja puesto y vacío, que no es
  lo mismo.

Lo que desaparece del template es **el nombre de la clase CSS**: la pantalla ya
no decide cómo se ve un elemento destacado, solo dice cuál lo está.

### Estructural

Decide **si un elemento existe y cuántas veces**. Lleva asterisco.

```html
<span class="bean" *appRepeat="coffee.rating">●</span>
```

**El asterisco es azúcar sintáctica.** Estas dos son exactamente lo mismo:

```html
<span *appRepeat="3">●</span>

<ng-template [appRepeat]="3">
  <span>●</span>
</ng-template>
```

> Un **`ng-template`** es un pedazo de HTML que **no se dibuja**: queda guardado,
> y alguien decide después si se pinta, cuántas veces y dónde.

Eso es lo que hacían `*ngIf` y `*ngFor` antes de que existieran `@if` y `@for`.

Por dentro se piden dos cosas con `inject()`:

```ts
private readonly template = inject(TemplateRef<unknown>);   // el HTML guardado
private readonly container = inject(ViewContainerRef);      // dónde insertarlo
```

> En Angular 18 casi nunca hace falta escribir una propia: el control flow de S3
> cubre los casos comunes y se lee mucho mejor. Está en el material para poder
> **leer** el código que las usa, que todavía es la mayoría.

---

## 7. Los errores de hoy

### NG8004 — el pipe no está declarado

```
No pipe found with name 'money'.
```

El pipe existe, pero **este componente** no lo declaró. Va en `imports`.

Y si ya está: el `name` del `@Pipe` tiene que coincidir **exacto** con lo que se
escribe en el template.

### TS-992011 — falta `standalone`

```
The directive 'HighlightDirective' appears in 'imports', but is not standalone
and cannot be imported directly. It must be imported via an NgModule.
```

En Angular 18 se escribe siempre, en los tres: componentes, directivas y pipes.

### El porcentaje en inglés

No es un error: nadie declaró el idioma. Se arregla en `app.config.ts`, y se
arregla una sola vez.

---

## 8. Glosario

| Palabra | Qué es |
|---|---|
| **Pipe** | Una función con nombre que transforma un valor en el template |
| **Pipe puro** | Corre solo cuando cambia la entrada. Es el valor por defecto |
| **Pipe impuro** | Corre cada vez que se revisa el componente |
| **Parámetro** | Lo que va después de los dos puntos: `\| money: 'USD'` |
| **Encadenar** | Pasar el resultado de un pipe a otro |
| **Directiva** | Una pieza sin template que le agrega algo a un elemento |
| **De atributo** | Cambia cómo se ve o se comporta |
| **Estructural** | Decide si existe y cuántas veces. Lleva asterisco |
| **`ng-template`** | HTML guardado que no se dibuja hasta que alguien lo pinta |
| **`host`** | Donde la directiva declara qué le hace al elemento |
| **`LOCALE_ID`** | El idioma con el que formatean los pipes incorporados |

---

## Para la tarea

Lo que **no** vimos hoy: `AsyncPipe`, que aparece en S6 cuando haya observables;
`KeyValuePipe`; y las directivas que necesitan escuchar eventos del elemento
—también con `host`—, que van a hacer falta recién en S8.
