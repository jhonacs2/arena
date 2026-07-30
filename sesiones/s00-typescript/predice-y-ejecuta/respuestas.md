# S0 · Predice y ejecuta — respuestas

> **Verificado con `tsc` 5.5.4 y el `tsconfig.json` del curso**, que tiene
> `strict` y `noUncheckedIndexedAccess`. Los mensajes están copiados de la
> salida, no escritos de memoria. Si cambiás el snippet, volvé a correrlo antes
> de dar la clase.

**El orden no se saltea:** mostrar → predecir 60 segundos → ejecutar → explicar.

---

## 1 · La misma cadena, dos tipos distintos

**Compila A. No compila B.**

```
TS2345: Argument of type 'string' is not assignable to parameter of type 'CoffeeSize'.
```

### Por qué

| | Qué escribimos | Qué tipo le dio TypeScript |
|---|---|---|
| **A** | `const size = 'grande'` | `'grande'` — el literal |
| **B** | `const config = { size: 'grande' }` | `string` |

Un `const` con un valor primitivo adentro **nunca va a cambiar**, así que
TypeScript se permite darle el tipo más chico posible: el literal `'grande'`.

Una **propiedad de un objeto sí puede cambiar** —`config.size = 'otra cosa'` es
legal—, así que TypeScript le da el tipo ancho: `string`. A eso se le dice
*widening*, ensanchar.

Es la misma cadena y el mismo `const`. Lo que cambia es **quién puede modificarla
después**.

### El arreglo, para mostrar al final

Dos caminos, y los dos sirven:

```ts
const config = { size: 'grande' } as const;   // congela todo el objeto
const config: { size: CoffeeSize } = { size: 'grande' };   // anota el tipo
```

### La frase para cerrar

> «El tipo no lo decide la letra que escribiste: lo decide **quién puede
> cambiarla más adelante**.»

---

## 2 · Un campo `readonly` con una lista adentro

**A y B no compilan. C sí compila.**

```
TS2540: Cannot assign to 'customer' because it is a read-only property.
TS2540: Cannot assign to 'lines' because it is a read-only property.
```

Y la tercera, `order.lines.push({ quantity: 1 })`, **pasa sin una sola queja**.

### Por qué

`readonly lines: OrderLine[]` protege **el campo**, no **la lista**. Nadie puede
reemplazar la lista por otra… y cualquiera puede vaciarla, ordenarla o agregarle
cosas.

```ts
readonly lines: readonly OrderLine[];
//       ↑              ↑
//   el campo       la lista
```

**Los dos `readonly` no son un error de tipeo.** Cada uno cierra una puerta
distinta, y la que casi todos dejan abierta es la segunda.

### La frase para cerrar

> «`readonly` en el campo dice *no me cambies la caja*. Lo que estaba adentro de
> la caja sigue siendo de todos.»

### Y por qué importa en este curso

En la sesión 3 el estado va a vivir en signals. Un `push` sobre un array de
estado hace que **la pantalla no se actualice**, y no da ningún error: es
exactamente el tipo de bug que se busca durante una hora. El `readonly` de adentro
es lo que hace que no se pueda escribir.

---

## 3 · Un dato que jura ser una carrera

**Compila perfecto. No hay un solo subrayado rojo.**

Y al ejecutar:

```
Clásico Apertura
Uncaught TypeError: Cannot read properties of undefined (reading 'length')
```

### Por qué

`JSON.parse` devuelve `any` — no tiene forma de saber qué había en el texto. El
`as Race` no verifica **nada**: le dice al compilador «tratá esto como una
carrera» y el compilador obedece.

El JSON tenía solo `name`. `race.horses` es `undefined`, y el tipo decía que era
una lista.

> **`as` no es una conversión. Es una promesa.** Y esta no se podía cumplir.

### Lo mismo vale para `!`

```ts
const best = menu[0]!;   // "confiá en mí, hay al menos uno"
```

Es la misma promesa con otra sintaxis, y fue la que rompió en el live coding:

```
TypeError: Cannot read properties of undefined (reading 'price')
```

### Qué se hace en su lugar

| En vez de | Va |
|---|---|
| `menu[0]!` | preguntar: `if (best === undefined)` o `?.` |
| `JSON.parse(t) as Race` | validar la forma antes de creerle |

Lo segundo se ve de verdad en la sesión 7, cuando los datos vengan de un servidor.
Hoy alcanza con saber que **el `as` no lo hace nadie por vos**.

---

## La pregunta de cierre del bloque

> «De los tres, ¿cuál les parece más peligroso en un proyecto de verdad?»

Casi siempre eligen el tercero, y tienen razón:

> «Exacto. Los dos primeros los descubrís escribiendo, y son molestos. El tercero
> compila, pasa el build, se despliega, y rompe con un usuario adelante. Cada vez
> que escriban `as` o `!`, la pregunta es una sola: **¿puedo cumplir esta
> promesa?**»
