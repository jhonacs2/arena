# S3 · Conceptos — Signals y control flow

> **Para qué es este archivo.** La clase es en vivo y no queda grabada. Cuando te
> sientes a hacer la tarea, esto es lo que tenés en vez de la memoria.

**Índice**

1. [Estado y derivado](#1-estado-y-derivado)
2. [Qué es un signal](#2-qué-es-un-signal)
3. [`set` y `update`](#3-set-y-update)
4. [`computed`](#4-computed)
5. [Inmutabilidad: la regla de la sesión](#5-inmutabilidad-la-regla-de-la-sesión)
6. [Control flow: `@if`, `@for`, `@switch`](#6-control-flow-if-for-switch)
7. [Los tres errores de hoy](#7-los-tres-errores-de-hoy)
8. [Glosario](#8-glosario)

---

## 1. Estado y derivado

Antes de escribir un signal hay una pregunta más vieja: **de todo lo que hay en
la pantalla, ¿qué hay que guardar?**

El tablero de la comanda muestra seis cosas. **Solo tres son estado:**

| | Qué |
|---|---|
| **Estado** | las comandas · el filtro elegido · el texto del buscador |
| **Derivado** | cuántas hay pendientes · cuánto falta cobrar · cuáles son las más caras |

> **Estado** es lo que nadie puede deducir de otra cosa.
> **Derivado** es lo que sale de calcular sobre el estado.

Guardar un derivado es firmar un compromiso: cada vez que cambie el estado hay
que acordarse de actualizarlo. En el tablero, la comanda cambia en **cuatro**
métodos. Cuatro lugares donde olvidarse.

---

## 2. Qué es un signal

> **Un signal es un valor que avisa cuando cambia.**

```ts
protected readonly orders = signal<readonly Order[]>(INITIAL_ORDERS);
```

Se lee **llamándolo**:

```html
@for (order of orders(); track order.id) { … }
```

Y ahí está la respuesta a la pregunta que quedó de S2: **por qué `coffee()` se
leía con paréntesis.** Un `input()` es un signal. Al leerlo, el signal **anota
quién lo leyó** — y por eso después puede avisarle.

Un dato común no puede hacer eso: es un número, se queda ahí y no sabe quién lo
está mirando.

![El signal avisa](diagramas/el-signal-avisa.svg)

---

## 3. `set` y `update`

```ts
this.orders.set(INITIAL_ORDERS);                          // reemplaza el valor entero
this.orders.update((orders) => [...orders, nuevaOrden]);  // parte de lo que hay
```

| | Cuándo |
|---|---|
| **`set`** | el valor nuevo no depende del anterior |
| **`update`** | necesitás mirar lo que había |

Los ejemplos exactos de la clase:

```ts
protected advance(id: string): void {
  this.orders.update((orders) =>
    orders.map((order) => (order.id === id ? { ...order, status: nextStatus(order.status) } : order)),
  );
}

protected remove(id: string): void {
  this.orders.update((orders) => orders.filter((order) => order.id !== id));
}
```

`map` devuelve un array nuevo y el `{ ...order }` un objeto nuevo. **Nada de lo
que estaba se modifica: se reemplaza.**

---

## 4. `computed`

> Un **`computed`** es un valor derivado de otros signals, que se recalcula solo
> cuando alguno de ellos cambia.

```ts
protected readonly pendingTotal = computed(() =>
  this.orders()
    .filter((order) => order.status !== 'served')
    .reduce((sum, order) => sum + lineTotal(order), 0),
);
```

**Lo que no hay que escribir es lo importante:** no se guarda en ninguna
propiedad, y no se actualiza en `advance`, ni en `remove`, ni en `add`, ni en
`reset`. Cuatro lugares donde no hay que acordarse de nada.

### Está memoizado

Un `computed` **guarda su último resultado** y lo devuelve hasta que su fuente
cambie. No se recalcula en cada repintado.

Por eso se puede llamar desde el template sin culpa — a diferencia de un método,
que sí corre en cada detección de cambios. Es la respuesta a la pregunta que
quedó abierta en S1.

### Un `computed` no tiene `set`

Si te dan ganas de escribirle uno, ese valor no era derivado: era estado, y va en
un `signal`.

---

## 5. Inmutabilidad: la regla de la sesión

> **Nunca modifiques lo que había. Poné algo nuevo.**

```ts
this.orders.update((orders) => [...orders, nueva]);   // ✅
this.orders().push(nueva);                            // ❌
```

### Qué pasa exactamente con el `push`

Esto es lo que corrimos en clase, y **no es lo que uno diría**:

| Qué se ve | Qué pasa |
|---|---|
| La lista del `@for` | **se actualiza**: aparece la fila nueva |
| El contador `{{ orders().length }}` | **se actualiza** |
| El total, que sale de un `computed` | **no se mueve. Nunca más.** |

**Media pantalla dice una cosa y la otra media dice otra.** Sin error, sin
advertencia, sin nada en la consola. Y no se arregla tocando otros botones: ese
`computed` no se va a recalcular jamás.

**El porqué:** `push` cambió lo que hay *adentro* del array. Pero el signal no
guarda el contenido: guarda **el array**, y el array es el mismo de siempre. Para
él no pasó nada, así que no avisó.

¿Y por qué la lista sí se actualizó? Porque el `@for` lee `orders()` directo y
Angular relee el template cada vez que revisa. El `computed` no: está memoizado
contra un aviso que nunca llegó.

### Por eso el `readonly` está desde S0

```ts
protected readonly orders = signal<readonly Order[]>(INITIAL_ORDERS);
```

```
Property 'push' does not exist on type 'readonly Order[]'.
```

**No es una formalidad: es lo que hace que este bug no se pueda escribir.**

### Y `sort` es igual de peligroso

```ts
[...this.visible()].sort(…)   // ✅ ordena una copia
this.visible().sort(…)        // ❌ ordena el original
```

`sort()` ordena **en el lugar** y devuelve el mismo array. Un `computed` que
ordena su fuente le cambia el orden a todos los que estaban leyendo ese signal
— fue el tercer «predice y ejecuta».

---

## 6. Control flow: `@if`, `@for`, `@switch`

Son sintaxis del template: **no van en `imports`**, a diferencia de los
componentes y de `FormsModule`.

### `@if`

```html
@if (visible().length === 0) {
  <p class="empty">…</p>
} @else {
  <ul>…</ul>
}
```

Y con alias, para no llamar dos veces a lo mismo:

```html
@if (selected(); as view) {
  <h2>{{ view.race.name }}</h2>
}
```

### `@for` — y `track` es obligatorio

```html
@for (order of visible(); track order.id) { … } @empty { … }
```

Sin `track` **no compila**:

```
NG5002: @for loop must have a "track" expression
```

> **Qué es `track`.** Es cómo Angular reconoce que la fila de Ana **sigue siendo
> la de Ana** cuando la lista se reordena o se filtra. Con `track order.id` la
> mueve de lugar. Sin nada con qué reconocerla, la destruiría y la volvería a
> crear — y con ella se va el foco, el scroll y lo que el usuario estuviera
> haciendo ahí.

`$index` sirve **solo** si no hay id y la lista no se reordena nunca. En cuanto
se reordena, miente.

`@empty` es el bloque para cuando la lista está vacía.

### `@switch`

```html
@switch (filter()) {
  @case ('pending') { No queda ninguna pendiente. La barra está al día. }
  @case ('ready') { Ninguna lista para entregar. }
  @default { No hay comandas que coincidan con la búsqueda. }
}
```

Sirve para lo mismo que el `switch` de TypeScript de S0: el valor es una unión
cerrada y se cubren los casos.

---

## 7. Los tres errores de hoy

### El `push` silencioso

Ya está arriba, en §5. **Es el error de la sesión.** Sin mensaje, sin log, con
media pantalla vieja.

### NG5002 — falta `track`

```
@for loop must have a "track" expression
```

Es de compilación, así que se encuentra solo. Es de las pocas cosas que Angular
te obliga a escribir.

### El `computed` que ordena su fuente

```ts
readonly sorted = computed(() => this.numbers().sort((a, b) => a - b));
```

Este ordenó el array que estaba adentro del signal. **Otro `computed` que solo
leía ese mismo signal empezó a devolver los números ordenados sin que nadie
tocara nada.**

Un valor que nadie cambió, cambió — y el culpable está en otro archivo.

---

## 8. Glosario

| Palabra | Qué es |
|---|---|
| **Estado** | Lo que se guarda porque no se puede deducir |
| **Derivado** | Lo que se calcula a partir del estado |
| **Signal** | Un valor que avisa cuando cambia |
| **`set`** | Reemplaza el valor entero |
| **`update`** | Recibe lo que hay, devuelve lo que va a haber |
| **`computed`** | Un derivado que se recalcula solo |
| **Memoizado** | Que guarda su resultado hasta que la fuente cambie |
| **Mutar** | Modificar lo que ya existe: `push`, `sort`, asignar |
| **Inmutable** | Poner un valor nuevo en vez de modificar el que había |
| **Control flow** | `@if`, `@for`, `@switch` en el template |
| **`track`** | Cómo Angular reconoce que una fila es la misma |

---

## Para la tarea

Lo que **no** vimos hoy: `effect()` —para cuando hace falta que algo *pase*, no
que se calcule—, cómo se comparte un signal entre componentes que no son padre e
hijo (sesión 5), y de dónde salen los datos cuando no están en una constante
(sesiones 6 y 7).

Y una que se va a entender recién en la sesión 10: **por qué todo esto funciona
con `OnPush`**. La respuesta corta es que un signal avisa, y con eso Angular sabe
exactamente qué componente revisar.
