# S3 · Misión profe — El live coding del bloque 0:20

**Solo instructor · 15 minutos · `lab/demo`**

---

## Antes de entrar al aula

```bash
node scripts/prep-demo.mjs
cd lab/demo
npm start
```

Y **sumá la ruta `/s03` a mano**, más `available: true` en `sessions.ts`.
Declarar rutas es tarea de los alumnos y hoy no es el tema.

**En pantalla:** VS Code y el navegador en <http://localhost:4200/s03>. Se ve el
tablero de la comanda con cinco filas, funcionando.

**Y una cuarta ventana que hoy importa:** el navegador **a la vista todo el
tiempo**. La rotura de las 0:26 no se ve en la terminal.

---

## 0:20 — De propiedad a signal · 3 min

```ts
import { signal } from '@angular/core';

protected orders = signal<readonly Order[]>(INITIAL_ORDERS);
```

Guardá. **Se rompe el template**: `orders` ya no es un array.

Poné los paréntesis: `orders` → `orders()`, en los cuatro métodos y en el `@for`.

> «Un cambio mecánico. Y miren la pantalla: **exactamente igual que antes**.»
>
> «Vale decirlo en voz alta: hasta acá no ganamos nada. Cambiamos cómo se
> escribe y no qué hace. Lo que se gana aparece en el próximo paso, y si no lo
> ven venir esto parece burocracia.»

---

## 0:23 — El primer `computed` · 3 min

```ts
import { computed } from '@angular/core';

protected readonly pendingTotal = computed(() =>
  this.orders()
    .filter((order) => order.status !== 'served')
    .reduce((sum, order) => sum + lineTotal(order), 0),
);
```

Y en el template, el panel:

```html
<p class="panel__lead num">{{ pendingTotal() }}</p>
```

**Tocá «Marcar lista», «Entregar», «Quitar» y «Agregar comanda».** El número
sigue solo.

> «Miren lo que **no** escribí. No lo guardé en ninguna propiedad. Y no lo
> actualicé en `advance`, ni en `remove`, ni en `add`, ni en `reset`.»
>
> «**Cuatro lugares donde no hay que acordarse de nada.** Ese es el trato.»

Si alguien pregunta si se recalcula todo el tiempo:

> «No. Guarda el resultado y lo devuelve hasta que `orders` cambie. Se dice que
> está **memoizado**. Por eso se puede llamar desde el template sin culpa.»

---

## 0:26 — La rotura · 4 min

**Es el bloque de la clase. No lo apures.**

Cambiá `add()` por la versión que muta:

```ts
protected add(): void {
  this.orders().push({ id: 'x', customer: 'Nuevo', coffee: 'Huila', quantity: 1, price: 38, status: 'pending' });
}
```

**No compila:**

```
Property 'push' does not exist on type 'readonly Order[]'.
```

> «Miren esto, que no estaba en el plan: **el tipo de la sesión 0 no me deja
> escribir el bug**. Voy a sacarle el `readonly` un segundo para poder
> mostrárselos — y quiero que se queden con este gesto, porque es exactamente el
> momento en el que hay que parar y preguntarse qué se está por hacer.»

Cambiá el tipo a `Order[]`, guardá, y **tocá «Agregar comanda»**.

> 🔴 **Rotura deliberada.** No digas nada durante tres segundos. Dejá que miren.

Después, señalando con el cursor:

> «La fila apareció: son seis. El encabezado dice seis. Y el total de acá
> abajo… **no se movió**.»
>
> «Y no se va a mover más. Miren.» — **tocá otros botones, filtrá, escribí en el
> buscador.** El total sigue viejo.

> «No hay error en la consola. No hay advertencia en la terminal. **Media
> pantalla dice una cosa y la otra media dice otra**, y van a seguir así hasta
> que alguien recargue.»

**El porqué, con el diagrama de las 0:12 si hace falta volver a él:**

> «`push` cambió lo que hay adentro del array. Pero el signal no guarda el
> contenido: guarda **el array**. Y el array es el mismo de siempre. Para él no
> pasó nada, así que no avisó, así que el `computed` no se recalculó.»
>
> «¿Y por qué la lista sí se actualizó? Porque el `@for` lee `orders()` directo y
> Angular relee el template cuando revisa. El `computed` no: está memoizado
> contra un aviso que nunca llegó.»

Arreglalo:

```ts
protected add(): void {
  this.orders.update((orders) => [...orders, { … }]);
}
```

Y **volvé a poner el `readonly`** en el tipo.

> «`update` recibe lo que hay y devuelve lo que va a haber. El spread arma un
> array nuevo, el signal ve que es otro, avisa, y todo se acomoda.»
>
> «Y ahora sí entienden por qué el `readonly` está desde la primera clase. **No
> es una formalidad: es lo que hace que este bug no se pueda escribir.**»

---

## 0:30 — `@for` con `track` · 2 min

Sacá el `track` del `@for` del tablero:

```
NG5002: @for loop must have a "track" expression
```

> «Angular te obliga, y es de las pocas cosas que te obliga.»
>
> «`track` es cómo reconoce que la fila de Ana **sigue siendo la de Ana** cuando
> la lista se reordena o se filtra. Con `track order.id` la mueve de lugar. Sin
> nada con qué reconocerla, la destruiría y la volvería a crear — y con ella se
> va el foco, el scroll y cualquier cosa que el usuario estuviera haciendo.»

Volvé a poner `track order.id`.

Si alguien pregunta por `$index`:

> «Sirve cuando no hay id y la lista no se reordena nunca. En cuanto se
> reordena, `$index` miente: la fila 0 pasa a ser otra cosa y Angular cree que
> es la misma.»

---

## 0:32 — `@switch` y el mensaje de vacío · 3 min

```html
@if (visible().length === 0) {
  <p class="empty">
    @switch (filter()) {
      @case ('pending') { No queda ninguna pendiente. La barra está al día. }
      @case ('ready') { Ninguna lista para entregar. }
      @case ('served') { Todavía no se entregó ninguna. }
      @default { No hay comandas que coincidan con la búsqueda. }
    }
  </p>
}
```

Filtrá por «Pendientes» y marcá las dos como listas, hasta que la lista quede
vacía. Que se vea el mensaje cambiando.

> «`@switch` es un `@if` con varias ramas, y sirve para lo mismo que el `switch`
> de TypeScript de la sesión 0: el valor es una unión cerrada, cubrís los casos.»
>
> «Y miren qué gana el usuario. Un “no hay nada” genérico no le dice si el
> filtro está mal puesto o si de verdad no queda nada por hacer. Esto sí.»

---

## Orden de sacrificio

| | Qué se saca | Por qué se puede |
|---|---|---|
| 1.º | El `@switch` de **0:32** | Está en el enunciado y en `conceptos.md` §6 |
| 2.º | El `track` de **0:30** | El error es de compilación: lo van a encontrar solos |
| 3.º | Nada más | |

**Lo que no se sacrifica nunca:** la rotura de las **0:26**. Es la sesión
entera. Si vas muy tarde, dala directo después del primer `computed` y dejá el
resto para el enunciado.

---

## Si algo sale mal

| Pasa | Qué hacer |
|---|---|
| `orders is not a function` | Le pusiste paréntesis a algo que no es un signal, o al revés. |
| `Cannot read properties of undefined` en el template | Un `computed` que lee otro que todavía no se declaró. El orden de declaración importa. |
| La rotura no se ve | Fijate que el total esté leyendo el `computed` y no calculándose en el template con un método. Con un método se recalcula igual y el bug no aparece. |
| Quedó todo hecho un desastre | `node scripts/prep-demo.mjs` y `demo/` vuelve a cero. |
