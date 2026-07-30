# S3 · Corrección — de propiedades a signals

**Bloque 1:45 · instructor y alumno**

---

# Parte A · El tablero del lab

**Archivo:** `lab/starter/src/app/sessions/s03/`
**Referencia terminada:** `lab/solution/src/app/sessions/s03/`

## Paso 0 · Que la pantalla exista

La ruta `/s03` en `app.routes.ts` y `available: true` en `sessions.ts`. Dos
archivos, como siempre.

## Paso 1 · El estado, y nada más que el estado

```ts
private readonly orders = signal<readonly Order[]>(INITIAL_ORDERS);
protected readonly filter = signal<OrderFilter>('all');
protected readonly query = signal('');
```

**Tres. Ni uno más.** Todo lo que se pueda calcular a partir de estos tres no va
en un signal.

> **Por qué `orders` es `private` y los otros dos `protected`.** El template no
> tiene por qué leer la comanda entera: lee `visible()`. Cerrarlo obliga a pasar
> por lo derivado, que es justamente la disciplina que la sesión enseña.

Y los cambios, con `update`:

```ts
protected advance(id: string): void {
  this.orders.update((orders) =>
    orders.map((order) => (order.id === id ? { ...order, status: nextStatus(order.status) } : order)),
  );
}
```

**Por qué `map` y no un bucle que asigne.** `map` devuelve un array nuevo y el
spread un objeto nuevo. Si en cambio se hiciera `order.status = …`, el objeto
sería el mismo, el array sería el mismo, y el signal no avisaría nada.

## Paso 2 · Lo derivado

```ts
protected readonly visible = computed<readonly Order[]>(() => {
  const status = this.filter();
  const text = this.query().trim().toLowerCase();

  return this.orders()
    .filter((order) => status === 'all' || order.status === status)
    .filter((order) => text === '' || `${order.customer} ${order.coffee}`.toLowerCase().includes(text));
});
```

**Por qué los dos filtros van encadenados y no en una condición gigante.** Se lee
en dos renglones lo que hace cada uno, y agregar un tercer criterio es una línea
más. Es una decisión de lectura, no de rendimiento: sobre cinco elementos da
igual.

```ts
protected readonly counts = computed(() => {
  const orders = this.orders();
  return {
    all: orders.length,
    pending: orders.filter((order) => order.status === 'pending').length,
    ready: orders.filter((order) => order.status === 'ready').length,
    served: orders.filter((order) => order.status === 'served').length,
  };
});
```

**Por qué los contadores salen de `orders()` y no de `visible()`.** Si salieran de
lo que se ve, al filtrar por «Pendientes» los otros tres contadores dirían 0 — y
entonces no sirven para nada: su única función es decirte qué vas a encontrar
**antes** de filtrar.

Es el error más común del ejercicio.

```ts
protected readonly byPrice = computed<readonly Order[]>(() =>
  [...this.visible()].sort((a, b) => lineTotal(b) - lineTotal(a)),
);
```

**El `[...]` no es decorativo.** `sort()` ordena en el lugar y devolvería el mismo
array que hay adentro del signal. Sin la copia, mirar este panel reordena la
lista de arriba.

## Paso 3 · El control flow

```html
@for (order of visible(); track order.id) { … }
```

**`track` por id, no por `$index`.** Con el filtro puesto, las filas cambian de
posición. El índice identifica la posición, no la fila.

```html
@if (visible().length === 0) {
  <p class="empty">
    @switch (filter()) {
      @case ('pending') { No queda ninguna pendiente. La barra está al día. }
      …
      @default { No hay comandas que coincidan con la búsqueda. }
    }
  </p>
}
```

**Por qué el mensaje depende del filtro.** Un «no hay nada» genérico no distingue
entre *está el filtro mal puesto* y *de verdad no queda trabajo*. Son dos
situaciones opuestas para el que está atendiendo la barra.

---

# Parte B · El programa del hipódromo

**Referencia terminada:** `project/frontend/solution/src/app/features/races/race-list.component.ts`

## Paso 1 · Qué es estado acá

```ts
private readonly all: readonly RaceView[] = RACES.map((race) => ({ … }));

protected readonly filter = signal<RaceFilter>('all');
protected readonly query = signal('');
private readonly selectedId = signal<string | null>(null);
protected readonly amount = signal(100);
```

**`all` no es un signal, y es lo más importante del ejercicio.** Viene de una
constante y no cambia nunca. Un signal para algo que no cambia es ruido: agrega
paréntesis, sugiere que puede cambiar, y no compra nada.

> En S7, cuando las carreras las traiga `HttpClient`, **sí** va a ser un signal.
> Y ese cambio va a tocar una línea, porque todo lo demás ya se deriva.

## Paso 2 · El id, no el objeto

```ts
protected readonly selected = computed<RaceView | undefined>(() => {
  const id = this.selectedId();
  return id === null ? undefined : this.visible().find((view) => view.race.id === id);
});
```

**Esta es la decisión que se cobra sola.** Guardando el objeto entero, si el
usuario abre una carrera terminada y después filtra por «En vivo», el panel sigue
mostrando una carrera que ya no está en la lista. Para arreglarlo habría que
acordarse de limpiarlo en cada cambio de filtro **y** en cada cambio de búsqueda.

Derivándolo del id, el `find` no la encuentra y devuelve `undefined`. **El panel
se cierra solo, y nadie escribió esa línea.**

## Paso 3 · La parrilla ordenada

```ts
protected readonly lineup = computed<readonly Horse[]>(() => {
  const horses = this.selected()?.race.horses ?? [];
  return [...horses].sort((a, b) => a.odds - b.odds || a.number - b.number);
});
```

Mismo `[...]` de siempre. Sin él, abrir una carrera **reordena el dataset** —y el
dataset es el mismo objeto que van a leer todas las demás pantallas.

El `|| a.number - b.number` rompe el empate por número de partida, igual que
`favourite()` de S0.

## Paso 4 · El template

```html
@if (selected(); as view) { … } @else { … }
```

El `as` evita llamar tres veces a lo mismo y deja el template más corto. No es
obligatorio.

---

## Los cinco errores que aparecen todos los años

| Lo que hacen | Por qué no alcanza | Qué decirles |
|---|---|---|
| Guardan `visible` en un signal y lo actualizan a mano | Cuatro lugares donde olvidarse | «Contá cuántas líneas tocás para agregar un filtro nuevo.» |
| Los contadores salen de lo filtrado | Dicen 0 justo cuando servirían | «Filtrá por Terminadas y mirá el contador de En vivo.» |
| `sort()` sin la copia | Reordena el original | «Abrí el panel y mirá la lista de arriba.» |
| Guardan la carrera abierta, no el id | El panel no se cierra al filtrar | «Abrí una terminada y filtrá por En vivo.» |
| Ponen las ocho carreras en un signal | Ruido: no cambian nunca | «¿Qué línea de tu código llama a `set` sobre eso?» |

---

## Cómo se verifica que quedó bien

```bash
cd lab/starter              && npx tsc --noEmit && npm test
cd project/frontend/starter && npx tsc --noEmit && npm run build
node scripts/verify.mjs --fast
```

Y a ojo, que es lo que el verificador no puede hacer:

- Agregar una comanda mueve **la lista y el total a la vez**.
- Filtrar con una carrera abierta **cierra el panel**.
- Abrir una carrera **no reordena** nada.
- `grep -c "signal(" race-list.component.ts` devuelve **4**, no diez.
