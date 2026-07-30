# S0 · Misión profe — El live coding del bloque 0:20

**Solo instructor · 15 minutos · `lab/demo`**

La secuencia exacta que escribís en vivo, en orden, con lo que se dice en cada
paso y las tres roturas deliberadas. Está pensado para el segundo monitor: el
`guion.md` lleva la clase, esto lleva el teclado.

---

## Antes de entrar al aula

```bash
node scripts/prep-demo.mjs
cd lab/demo
npm start
```

`lab/demo` es una copia descartable de `lab/starter`, así que arranca en el mismo
estado en que están los alumnos: **los tipos de `menu.ts` están flojos**. Es el
lienzo correcto para apretarlos en vivo.

> **No trabajes en `lab/solution`.** No hay que romperle nada a la solución de
> referencia para dar la clase — para eso está `demo/`. Y si el bloque sale mal,
> `node scripts/prep-demo.mjs` te devuelve el lienzo limpio en un segundo, así
> que se puede ensayar tres veces.

**En pantalla:** VS Code y el navegador lado a lado, el navegador en
<http://localhost:4200/s00>. El archivo abierto es
`src/app/sessions/s00/menu.ts`.

**Y una tercera ventana que hoy importa más que otros días:** la terminal donde
corre `npm start`, visible. Los errores de tipo salen ahí con archivo y línea, y
que la vean aparecer en vivo es medio bloque.

**Decilo textual antes de empezar:**

> «Cierren el editor. Los próximos quince minutos yo escribo y ustedes miran. No
> copien: van a hacer esto mismo después, y con las manos libres se entiende
> mejor. Si me equivoco, avisen.»

---

## 0:20 — Inferencia y anotación · 2 min

**No escribas nada todavía.** Apoyá el mouse sobre `MENU` y dejá que se vea el
tooltip.

> «Miren lo que dice el editor cuando apoyo el mouse: ya sabe que esto es una
> lista de cafés. **Nadie se lo dijo.** Lo dedujo del valor. A eso se le dice
> **inferencia**, y es la razón por la que TypeScript no es escribir el doble.»

Ahora sí, al final del archivo:

```ts
let price = 42;
price = 'cuarenta y dos';
```

> 🔴 **Rotura deliberada 1.** Leé el error completo, en voz alta, del editor o de
> la terminal:

```
TS2322: Type 'string' is not assignable to type 'number'.
```

> «*El tipo `string` no se puede asignar al tipo `number`*. Yo nunca escribí
> “number” en ningún lado: lo dedujo de que puse 42. Y ahora me lo sostiene.»

**Borrá las dos líneas.**

> «Regla práctica: cuando el valor está ahí, no se anota. Se anota cuando
> todavía no está — los parámetros de una función, lo que devuelve, lo que viene
> de afuera.»

---

## 0:22 — La unión de literales · 3 min

Al final del archivo:

```ts
sizeFactor('mediana');
```

**Guardá. No pasa nada.**

> «Esto compila. Es el mismo error que vimos a las 0:05, ahora con TypeScript
> prendido. ¿Por qué pasa?»

Subí hasta `CoffeeSize` y mostralo:

```ts
export type CoffeeSize = string;
```

> «Porque el parámetro es `CoffeeSize`, y `CoffeeSize` hoy es `string`. Y
> `'mediana'` es un `string` perfectamente válido.»

Cambiá esa línea:

```ts
export type CoffeeSize = 'chico' | 'mediano' | 'grande';
```

**Guardá.** El error aparece solo, cincuenta líneas más abajo:

```
TS2345: Argument of type '"mediana"' is not assignable to parameter of type 'CoffeeSize'.
```

> «Cambié **una línea** y el compilador encontró un problema que estaba a
> cincuenta líneas de distancia. Eso es lo que compra apretar un tipo.»
>
> «`type` le pone nombre a un tipo. La barra vertical es una **unión**: “esto o
> esto o esto”. Y cada una de esas tres cadenas entre comillas es un **literal**:
> un tipo que tiene un solo valor adentro.»

**Borrá la línea de prueba.**

---

## 0:25 — El `switch` que se queja solo · 3 min

Reemplazá el cuerpo de `sizeFactor`:

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

> «Fíjense que no hay `default`, y sin embargo compila. TypeScript sabe que
> `CoffeeSize` tiene tres valores, ve que los cubrí a los tres, y por eso sabe
> que esta función siempre devuelve algo.»

**Ahora la demostración que vale por todo el bloque.** Volvé al tipo y agregale
un cuarto tamaño:

```ts
export type CoffeeSize = 'chico' | 'mediano' | 'grande' | 'jumbo';
```

> 🔴 **Rotura deliberada 2.** El error **no** aparece donde escribiste:

```
TS2366: Function lacks ending return statement and return type does not include 'undefined'.
```

> «Agregué un tamaño arriba y el compilador me llevó al único lugar del programa
> donde falta decidir qué hacer con él. **Eso** es lo que un comentario nunca va
> a hacer por ustedes.»

**Sacá `| 'jumbo'`.**

---

## 0:28 — Opcional y narrowing · 3 min

Mostrá el `MENU` y señalá con el cursor los dos `notes: ''`.

> «Estos dos cafés no tienen nota de cata. Les pusimos cadena vacía porque el
> tipo nos obligó a poner algo. Es una mentira chiquita, y las mentiras chiquitas
> se propagan: ahora, en todo el programa, “no tiene nota” y “tiene una nota
> vacía” son lo mismo.»

En la interfaz `Coffee`:

```ts
  notes?: string;
```

> «El signo de pregunta dice **puede no estar**. Y ahora el tipo de
> `coffee.notes` ya no es `string`: es `string | undefined`.»

**Borrá los dos `notes: ''` del `MENU` y guardá.** Andá al navegador.

> 🔴 **Rotura deliberada 3.** El pizarrón ahora dice:

```
Huila · Colombia — undefined
```

**No lo arregles todavía.** Dejalo en pantalla dos segundos.

> «Nadie se quejó. Compila. Y está mal, porque `describe` sigue preguntando si la
> nota es la cadena vacía, y ya no hay ninguna cadena vacía en ningún lado.»

Ahora sí, en `describe`:

```ts
  if (coffee.notes === undefined) {
    return base;
  }
  return `${base} — ${coffee.notes}`;
```

**Y hacé esto, que es el momento del bloque:** apoyá el mouse sobre
`coffee.notes` **arriba** del `if`, y después **abajo**.

> «Arriba dice `string | undefined`. Abajo dice `string`, a secas.»
>
> «El compilador entendió que si el código llegó hasta acá, no puede ser
> `undefined`. A eso se le dice **narrowing**: adentro del `if` el tipo se
> achica, porque ahí ya sabe más que afuera.»

---

## 0:31 — Lo que puede no estar · 2 min

En la consola del navegador, o como una línea temporal en el archivo:

```ts
cheapest([]);
```

```
TypeError: Cannot read properties of undefined (reading 'price')
```

Mostrá la línea culpable de `cheapest`:

```ts
let best = menu[0]!;
```

> «Ese signo de exclamación es una promesa: le estoy diciendo al compilador
> *confiá en mí, siempre hay al menos uno*. Y no la puedo cumplir.»
>
> «La regla es corta: **si te tienta poner un `!`, casi siempre lo que falta es
> decir la verdad en el tipo.**»

Cambiá la firma y sacá el `!`:

```ts
export function cheapest(menu: readonly Coffee[]): Coffee | undefined {
  let best: Coffee | undefined = undefined;
  for (const coffee of menu) {
    if (best === undefined || coffee.price < best.price) best = coffee;
  }
  return best;
}
```

Volvé al navegador: **el pizarrón sigue funcionando**.

> «Y esto es lo que quiero que vean: la pantalla ya sabía qué hacer si no hay
> café más barato, porque el template tiene los dos casos escritos. El tipo no
> agregó trabajo: **hizo visible el trabajo que ya había que hacer.**»

---

## 0:33 — `readonly` · 2 min

Al final del archivo:

```ts
MENU.push({ id: 'c9', name: 'Trucho', origin: 'Ninguno', price: 1 });
```

**Guardá. Compila.**

> «Estoy modificando el menú del negocio desde cualquier parte del programa, y
> nadie me dice nada.»

Subí al `MENU` y a la interfaz `Coffee`, y poné `readonly` donde va:

```ts
export interface Coffee {
  readonly id: string;
  readonly name: string;
  readonly origin: string;
  readonly price: number;
  readonly notes?: string;
}

export const MENU: readonly Coffee[] = [ … ];
```

```
TS2339: Property 'push' does not exist on type 'readonly Coffee[]'.
```

> «`readonly` no congela nada en tiempo de ejecución: es una marca para el
> compilador. Y alcanza, porque el que se equivoca es el que escribe.»
>
> «Esto va a estar en **todo** el curso, y en la clase 3 va a ser la diferencia
> entre que la pantalla se actualice y que no. Por ahora quédense con esto: los
> datos que vienen de afuera no se editan en el lugar.»

**Borrá el `push`.**

---

## Orden de sacrificio

Si a las 0:30 vas por el paso de las 0:25, recortá en este orden:

| | Qué se saca | Por qué se puede |
|---|---|---|
| 1.º | El `push` de **0:33** | Queda dicho en `conceptos.md` §7 y es un `TODO` del ejercicio igual |
| 2.º | La demo del `!` de **0:31** | El error de runtime se puede contar en diez segundos sin ejecutarlo |
| 3.º | El `'jumbo'` de **0:25** | Duele, pero el `switch` ya se ve compilando sin `default` |

**Lo que no se sacrifica nunca:** la unión de literales de **0:22** y el
narrowing de **0:28**. Son los dos que se usan en las dos misiones.

---

## Si algo sale mal

| Pasa | Qué hacer |
|---|---|
| El error no aparece en la terminal | El `ng serve` se quedó colgado en un error anterior. `Ctrl+C` y `npm start` de nuevo — treinta segundos. |
| El tooltip del hover no sale | El servidor de TypeScript de VS Code se cayó. `Ctrl+Shift+P` → *TypeScript: Restart TS Server*. |
| Quedó el archivo hecho un desastre | `node scripts/prep-demo.mjs` en otra terminal y `demo/` vuelve a cero. Perdés el paso, no la clase. |
