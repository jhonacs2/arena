# S0 · Conceptos — TypeScript

> **Para qué es este archivo.** La clase es en vivo y no queda grabada. Cuando te
> sientes a hacer la tarea, esto es lo que tenés en vez de la memoria: cada
> concepto que vimos, con su definición y **los ejemplos exactos que corrimos en
> clase**. No es un resumen decorativo — está pensado para tenerlo abierto al
> lado del editor.

**Índice**

1. [El problema: tres programas que no fallaron](#1-el-problema-tres-programas-que-no-fallaron)
2. [Qué es un tipo](#2-qué-es-un-tipo)
3. [El compilador, y por qué los tipos se borran](#3-el-compilador-y-por-qué-los-tipos-se-borran)
4. [Inferencia y anotación](#4-inferencia-y-anotación)
5. [`interface` y `type`](#5-interface-y-type)
6. [Uniones y literales](#6-uniones-y-literales)
7. [Opcionales, `undefined` y narrowing](#7-opcionales-undefined-y-narrowing)
8. [`readonly`](#8-readonly)
9. [Genéricos](#9-genéricos)
10. [Utility types](#10-utility-types)
11. [Módulos: `import` y `export`](#11-módulos-import-y-export)
12. [Las dos salidas de emergencia: `!` y `as`](#12-las-dos-salidas-de-emergencia--y-as)
13. [Los errores de hoy y qué significan](#13-los-errores-de-hoy-y-qué-significan)
14. [Glosario](#14-glosario)

---

## 1. El problema: tres programas que no fallaron

Los tres que corrimos al empezar la clase, en JavaScript puro:

```js
const price = '42';
console.log(price * 2);   // 84
console.log(price + 2);   // '422'
```

```js
const race = { name: 'Clásico Apertura' };
console.log(race.nombre); // undefined
```

```js
function sizeFactor(size) {
  if (size === 'chico') return 0.8;
  if (size === 'grande') return 1.3;
  return 1;
}
sizeFactor('mediana');    // 1
```

**Ninguno tiró un error. Los tres están mal.** Un texto que se comporta como
número o como texto según el operador, una propiedad mal escrita que devuelve
`undefined` en silencio, y un valor inválido que se cobra como si fuera válido.

> El error ya estaba escrito en el archivo. El problema no era el error: era que
> **nadie lo leyó hasta que fue tarde**.

---

## 2. Qué es un tipo

> **Un tipo es el conjunto de valores que una cosa puede tener.**

Nada más que eso. `number` es el conjunto de todos los números. `string` es el
conjunto de todas las cadenas — infinitas, incluidas `'mediana'`, `'asdf'` y la
cadena vacía.

![Un tipo es un conjunto](diagramas/un-tipo-es-un-conjunto.svg)

Cuando decimos que el tamaño de un café es `string`, estamos diciendo *cualquiera
de esas infinitas sirve*. Y no es lo que queremos decir: queremos decir tres.

| | Qué dice | Qué pasa con `'mediana'` |
|---|---|---|
| `string` | cualquier cadena | devuelve 1 y nadie avisa |
| `'chico' \| 'mediano' \| 'grande'` | estas tres | no compila, con archivo y línea |

**Apretar un tipo es achicar ese conjunto hasta que sea el que corresponde.** Es
todo lo que hicimos hoy.

---

## 3. El compilador, y por qué los tipos se borran

> **El compilador** es el programa que lee tu TypeScript, revisa que los tipos
> cierren, y escribe el JavaScript que se va a ejecutar.

Y acá está lo que más cuesta el primer día:

> **TypeScript no existe en el navegador.** No hay ningún motor de TypeScript en
> ningún lado. Los tipos se borran antes de correr.

Este archivo:

```ts
const price: number = 42;
```

se convierte, literalmente, en esto:

```js
const price = 42;
```

Los tipos son un **andamio**: sostienen mientras se construye y no van a la obra
terminada. Sirven para el rato en que escribís, que es donde se cometen los
errores.

En este curso el compilador corre en **modo `strict`**, más algunas opciones
extra, en los cuatro proyectos. No se apaga ningún día. La que más se nota:

> **`noUncheckedIndexedAccess`** — `horses[0]` no es `Horse`, es
> `Horse | undefined`. Porque la lista podría estar vacía, y el compilador no
> tiene forma de saber que no lo está.

---

## 4. Inferencia y anotación

> **Anotación** es escribir el tipo a mano, después de dos puntos.
> **Inferencia** es cuando TypeScript lo deduce solo y no hace falta escribirlo.

```ts
let price = 42;              // inferido: number. Nadie escribió "number".
price = 'cuarenta y dos';    // ✗ TS2322
```

```
TS2322: Type 'string' is not assignable to type 'number'.
```

**La regla práctica:**

| | Se anota |
|---|---|
| Un valor que ya está ahí | **No.** Ya se sabe qué es |
| Un parámetro de función | **Sí.** El valor todavía no existe |
| Lo que devuelve una función | **Sí.** Es la promesa que la función hace |
| Algo que viene de afuera | **Sí.** Nadie puede deducirlo |

Por eso TypeScript no es escribir el doble: adentro de las funciones casi no se
anota nada.

---

## 5. `interface` y `type`

> **`interface`** le pone nombre a la forma de un objeto.
> **`type`** le pone nombre a **cualquier** tipo, no solo a objetos.

```ts
export interface Coffee {
  readonly id: string;
  readonly name: string;
  readonly origin: string;
  readonly price: number;
  readonly notes?: string;
}
```

```ts
export type CoffeeSize = 'chico' | 'mediano' | 'grande';
```

**Cuál usar, en este curso:**

| | |
|---|---|
| La forma de un objeto | `interface` |
| Todo lo demás — uniones, alias, tipos derivados | `type` |

Es una convención, no una ley del lenguaje: `type` también puede describir
objetos. La seguimos porque hace que leer un archivo de modelos sea rápido — lo
que empieza con `interface` es una cosa, lo que empieza con `type` es una regla
sobre valores.

---

## 6. Uniones y literales

> Una **unión** es un tipo que es «esto o esto otro»: se escribe con `|`.
> Un **literal** es un tipo con un solo valor adentro: `'grande'` es un tipo.

```ts
export type CoffeeSize = 'chico' | 'mediano' | 'grande';
```

Y los números también:

```ts
readonly place: 1 | 2 | 3;
```

### El `switch` que se queja solo

El ejemplo que corrimos en clase:

```ts
export function sizeFactor(size: CoffeeSize): number {
  switch (size) {
    case 'chico':
      return 0.8;
    case 'mediano':
      return 1;
    case 'grande':
      return 1.3;
  }
}
```

**No tiene `default`, y compila igual.** TypeScript sabe que la unión tiene tres
valores, ve que están los tres cubiertos, y por eso sabe que la función siempre
devuelve un número.

Y el momento que hay que recordar: **le agregamos un cuarto tamaño al tipo**…

```ts
export type CoffeeSize = 'chico' | 'mediano' | 'grande' | 'jumbo';
```

…y el error apareció en `sizeFactor`, no en el tipo:

```
TS2366: Function lacks ending return statement and return type does not include 'undefined'.
```

> Cambiás el tipo en un lugar y el compilador te lleva a **todos** los lugares
> donde falta decidir. Un `default` habría tapado eso, y por eso no lo ponemos.

---

## 7. Opcionales, `undefined` y narrowing

> **`undefined`** es el valor que significa «acá no hay nada».
> Un campo **opcional** se marca con `?` y quiere decir que puede no estar.

Antes, en el menú, dos cafés sin nota de cata llevaban cadena vacía:

```ts
{ id: 'c2', name: 'Huila', origin: 'Colombia', price: 38, notes: '' },
```

**La cadena vacía era una mentira chiquita:** en todo el programa, «no tiene
nota» y «tiene una nota vacía» eran indistinguibles. La verdad se escribe así:

```ts
readonly notes?: string;      // el tipo del campo es string | undefined
```

### Narrowing

```ts
export function describe(coffee: Coffee): string {
  const base = `${coffee.name} · ${coffee.origin}`;
  if (coffee.notes === undefined) {
    return base;
  }
  return `${base} — ${coffee.notes}`;
}
```

**Lo que hicimos en clase fue apoyar el mouse sobre `coffee.notes` arriba y abajo
del `if`:**

| Dónde | Qué dice el editor |
|---|---|
| arriba del `if` | `string \| undefined` |
| después del `return` | `string` |

> **Narrowing** es eso: adentro de un `if`, el tipo se achica, porque ahí el
> compilador sabe más que afuera.

### Lo mismo, en un valor de retorno

```ts
export function cheapest(menu: readonly Coffee[]): Coffee | undefined {
  let best: Coffee | undefined = undefined;
  for (const coffee of menu) {
    if (best === undefined || coffee.price < best.price) best = coffee;
  }
  return best;
}
```

Con un menú vacío no hay café más barato, y el tipo lo dice. Quien lo use tiene
que decidir qué mostrar en ese caso — que es lo que el pizarrón ya hacía:

```html
@if (best) { … } @else { <p class="empty">El menú está vacío.</p> }
```

> El tipo no agregó trabajo. **Hizo visible el trabajo que ya había que hacer.**

---

## 8. `readonly`

> **`readonly`** marca que algo no se puede modificar. Es una marca **para el
> compilador**: no congela nada en tiempo de ejecución.

```ts
MENU.push({ id: 'c9', name: 'Trucho', origin: 'Ninguno', price: 1 });
```

```
TS2339: Property 'push' does not exist on type 'readonly Coffee[]'.
```

**Y la trampa, que fue el segundo «predice y ejecuta» de la clase:** son dos
lugares distintos y hacen cosas distintas.

```ts
readonly lines: OrderLine[];            // el campo no se reasigna…
                                        // …pero lines.push(…) COMPILA

readonly lines: readonly OrderLine[];   // ahora sí, las dos cosas
```

| Dónde va | Qué impide |
|---|---|
| `readonly campo:` | reasignar el campo: `order.lines = []` |
| `: readonly Tipo[]` | modificar la lista: `order.lines.push(…)` |

**Por qué está en todo el curso:** en la sesión 3, cuando el estado viva en
signals, modificar un array en el lugar va a hacer que la pantalla **no se
actualice**. `readonly` es lo que hace que ese error no se pueda escribir.

---

## 9. Genéricos

> Un **genérico** es un tipo con un hueco, que se rellena al usarlo.

```ts
export interface Page<T> {
  readonly items: readonly T[];
  readonly total: number;
}
```

`Page<Coffee>` es una página de cafés. `Page<Order>`, una de pedidos. Es la misma
interfaz: sin el hueco habría que escribir `CoffeePage` y `OrderPage`, iguales
salvo una palabra.

En el hipódromo existe exactamente este tipo, y lo van a usar `Page<Race>` y
`Page<Bet>`.

Las funciones también:

```ts
export function firstPage<T>(items: readonly T[], size: number): Page<T> {
  return { items: items.slice(0, size), total: items.length };
}
```

**No hay que escribir el hueco al llamarla:** `firstPage(MENU, 3)` ya es
`Page<Coffee>`, porque TypeScript deduce `T` del argumento.

---

## 10. Utility types

> Un **utility type** arma un tipo a partir de otro, en vez de copiarlo.

Los tres que se usan en el curso:

| | Qué hace | Ejemplo |
|---|---|---|
| `Omit<T, 'x'>` | `T` sin el campo `x` | `type CoffeeDraft = Omit<Coffee, 'id'>` |
| `Pick<T, 'x' \| 'y'>` | solo `x` e `y` de `T` | `type RaceSummary = Pick<Race, 'id' \| 'name'>` |
| `Partial<T>` | todos los campos de `T`, opcionales | para un formulario a medio llenar |

**Por qué derivar y no copiar.** Un `CoffeeDraft` escrito a mano con los cuatro
campos funciona perfecto, hasta el día en que `Coffee` gana un quinto. Ese día la
copia se queda vieja y **nadie se entera**. Con `Omit`, ese día no existe.

---

## 11. Módulos: `import` y `export`

> Cada archivo `.ts` es un **módulo**. Lo que lleva `export` se puede usar desde
> otro archivo; lo que no, es privado del archivo.

```ts
// menu.ts
export interface Coffee { … }
export const MENU: readonly Coffee[] = [ … ];
```

```ts
// s00.component.ts
import { MENU, cheapest, priceFor } from './menu';
import type { Coffee, Order } from './menu';
```

**`import type`** trae solo el tipo, y desaparece por completo al compilar. No es
obligatorio, pero deja claro qué se está trayendo: `MENU` existe cuando el
programa corre, `Coffee` no.

Todo el proyecto está armado así, y es lo mismo que vas a ver en Angular desde la
clase que viene.

---

## 12. Las dos salidas de emergencia: `!` y `as`

Las dos hacen que el compilador se calle. **Ninguna verifica nada.**

```ts
const best = menu[0]!;                 // "confiá en mí, existe"
const race = JSON.parse(text) as Race; // "confiá en mí, tiene esta forma"
```

El primero fue el que rompió en clase:

```
TypeError: Cannot read properties of undefined (reading 'price')
```

> **Un tipo es una promesa. `!` y `as` son vos jurando, sin que nadie te
> verifique.** Cada vez que te tiente escribir uno, la pregunta es una sola:
> *¿puedo cumplir esta promesa?* Si la respuesta no es un sí rotundo, lo que
> falta es decir la verdad en el tipo.

En este curso las dos están **prohibidas en los ejercicios**, junto con `any`.

---

## 13. Los errores de hoy y qué significan

### TS2322 — asignar un tipo a otro que no lo acepta

```
Type 'string' is not assignable to type 'number'.
```

**Qué dice:** el valor que estás poniendo no pertenece al conjunto que declaraste.
Aparece al reasignar, al armar un objeto y al devolver algo de una función.

### TS2345 — un argumento que no entra en el parámetro

```
Argument of type '"mediana"' is not assignable to parameter of type 'CoffeeSize'.
```

**Qué dice:** llamaste a una función con algo que no está en la unión. Es el error
que aparece al apretar `CoffeeSize`, y el que hace visible el bug del principio de
la clase.

### TS2366 — falta un caso

```
Function lacks ending return statement and return type does not include 'undefined'.
```

**Qué dice:** hay un camino por el que la función no devuelve nada. En un `switch`
sin `default` sobre una unión, significa: **te falta un caso**.

### TS2339 — esa propiedad no existe en ese tipo

```
Property 'push' does not exist on type 'readonly Coffee[]'.
```

**Qué dice:** un array `readonly` no tiene `push`, ni `pop`, ni `sort`. No es que
esté bloqueado: es que ese tipo directamente no tiene esos métodos.

### TS2532 / TS18048 — puede ser `undefined`

```
Object is possibly 'undefined'.
```

**Qué dice:** estás usando algo que el tipo dice que puede no estar. **La solución
no es un `!`**: es preguntar antes, con un `if` o con `?.`.

### TS2367 — esta comparación no puede dar verdadero nunca

```
This comparison appears to be unintentional because the types 'RaceStatus' and '"galopando"' have no overlap.
```

**Qué dice:** estás comparando contra un valor que no está en el conjunto. Es el
error más lindo de todos, porque encuentra un `if` que nunca se cumple — algo que
en JavaScript se descubre meses después.

### TS2540 — es de solo lectura

```
Cannot assign to 'odds' because it is a read-only property.
```

### TS2741 — falta un campo

```
Property 'notes' is missing in type … but required in type 'Coffee'.
```

**Cuándo apareció hoy:** al borrar los `notes: ''` **antes** de marcar el campo
como opcional. El orden que menos duele es primero el tipo, después los datos.

---

## 14. Glosario

| Palabra | Qué es |
|---|---|
| **Compilador** | El programa que lee tu TypeScript, lo revisa y escribe el JavaScript |
| **Tipo** | El conjunto de valores que una cosa puede tener |
| **Anotación** | Escribir el tipo a mano: `price: number` |
| **Inferencia** | Cuando TypeScript deduce el tipo solo |
| **`interface`** | Un nombre para la forma de un objeto |
| **Alias de tipo** | Un nombre para cualquier tipo. Se escribe con `type` |
| **Literal** | Un tipo con un solo valor adentro: `'grande'`, `1` |
| **Unión** | «Esto o esto otro»: `A \| B` |
| **Narrowing** | Cuando el tipo se achica adentro de un `if` |
| **Opcional** | Un campo que puede no estar: `notes?: string` |
| **`undefined`** | El valor que significa «acá no hay nada» |
| **`strict`** | El modo del compilador que no deja pasar lo dudoso |
| **`noUncheckedIndexedAccess`** | La opción que hace que `lista[0]` sea `T \| undefined` |
| **Genérico** | Un tipo con un hueco: `Page<T>` |
| **Utility type** | Un tipo armado a partir de otro: `Omit`, `Pick`, `Partial` |
| **`readonly`** | Marca que impide reasignar un campo o modificar un array |
| **Módulo** | Cada archivo `.ts`, con sus `export` e `import` |
| **`!`** | «Confiá en mí, esto existe». No verifica nada |
| **`as`** | «Confiá en mí, esto tiene esta forma». No verifica nada |

---

## Para la tarea

Con esto alcanza para hacer `tarea.md` sin nada más. Si algo de acá no te cierra,
es mejor preguntarlo al empezar la sesión que viene que arrastrarlo: **todo el
curso está escrito en TypeScript, y estos tipos son los que vamos a usar once
clases seguidas.**

Lo que **no** vimos hoy y no hace falta todavía: clases y herencia, `enum`,
decoradores, tipos condicionales, y todo lo que sea Angular. Los decoradores
aparecen la clase que viene, con `@Component`, y no hace falta entender cómo
funcionan para usarlos.
