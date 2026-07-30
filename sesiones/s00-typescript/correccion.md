# S0 · Corrección — de los tipos flojos a los tipos que dicen la verdad

**Bloque 1:45 · instructor y alumno**

Las dos misiones, paso a paso, con **qué se escribe, dónde y por qué ahí**. El
«por qué» es lo único que no se puede sacar leyendo `solution/`.

> **Cómo se usa.** En clase, en pantalla mientras se revisa una solución. En
> casa, para destrabarse: buscá el paso donde estás, leelo, y seguí desde ahí.
> Copiarlo entero no sirve — el ejercicio es ver aparecer y desaparecer errores.

---

# Parte A · El menú del lab

**Archivo:** `lab/starter/src/app/sessions/s00/menu.ts`
**Referencia terminada:** `lab/solution/src/app/sessions/s00/menu.ts`

## Antes de empezar

Dejá dos cosas a la vista: el editor y **la terminal donde corre `npm start`**.
Casi todo el ejercicio consiste en hacer aparecer un error y hacerlo desaparecer,
y la terminal es donde aparece con archivo y línea.

## Paso 1 · Los tamaños son tres

```ts
// antes
export type CoffeeSize = string;

// después
export type CoffeeSize = 'chico' | 'mediano' | 'grande';
```

**Por qué ahí y no en cada función.** El tipo se declara **una vez** y lo usan
`sizeFactor`, `priceFor` y `OrderLine`. Si en vez de esto se escribieran las tres
cadenas en cada firma, el día que aparezca un cuarto tamaño habría que acordarse
de todos los lugares. Un alias es un solo lugar.

**Cómo se comprueba:** al final del archivo, `sizeFactor('mediana');`

```
TS2345: Argument of type '"mediana"' is not assignable to parameter of type 'CoffeeSize'.
```

Borrar la línea de prueba.

## Paso 2 · Las notas de cata son opcionales

**El orden importa: primero el tipo, después los datos.** Al revés, el error que
aparece es sobre datos incompletos y confunde.

```ts
// en la interfaz Coffee
notes?: string;
```

Ahora sí, borrar los dos `notes: ''` del `MENU` — los de Huila y Antigua.

Y arreglar `describe()`, que quedó preguntando por algo que ya no existe:

```ts
// antes
if (coffee.notes === '') {
  return base;
}

// después
if (coffee.notes === undefined) {
  return base;
}
```

**Por qué esto es más que cosmética.** Con `notes: string`, «no tiene nota» y
«tiene una nota vacía» eran indistinguibles: cualquiera que leyera ese campo
tenía que saber, por fuera del tipo, que la cadena vacía significaba ausencia.
Con `notes?: string` la ausencia está en el tipo, y el compilador obliga a
tratarla.

**Cómo se comprueba:** el pizarrón dice `Huila · Colombia`, sin `— undefined`.

> **El error más común de este paso** es borrar los `notes: ''` antes de tocar la
> interfaz:
>
> ```
> TS2741: Property 'notes' is missing in type … but required in type 'Coffee'.
> ```
>
> No está mal resuelto: está resuelto en el orden que duele.

## Paso 3 · El menú no se edita

```ts
export interface Coffee {
  readonly id: string;
  readonly name: string;
  readonly origin: string;
  readonly price: number;
  readonly notes?: string;
}

export interface OrderLine {
  readonly coffee: Coffee;
  readonly size: CoffeeSize;
  readonly quantity: number;
}

export interface Order {
  readonly customer: string;
  readonly lines: readonly OrderLine[];
}

export const MENU: readonly Coffee[] = [ … ];
```

**Los dos `readonly` de `lines` no son un error de tipeo.** Hacen cosas
distintas:

| Dónde | Qué impide |
|---|---|
| `readonly lines: …` | reasignar el campo: `order.lines = []` |
| `…: readonly OrderLine[]` | modificar la lista: `order.lines.push(…)` |

Uno solo de los dos deja pasar la mitad. Es el segundo «predice y ejecuta» de la
clase.

**Cómo se comprueba:** al final del archivo,
`MENU.push({ id: 'c9', name: 'Trucho', origin: 'Ninguno', price: 1 });`

```
TS2339: Property 'push' does not exist on type 'readonly Coffee[]'.
```

Borrar la línea de prueba.

## Paso 4 · El `switch` que se queja solo

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

**Por qué sin `default`.** Con un `default` esto compilaría siempre, incluso
cuando falte un caso — que es exactamente el comportamiento que tenía la versión
con `if` y `return 1`. Sin `default`, TypeScript comprueba que los tres casos de
la unión estén cubiertos: si mañana aparece un cuarto tamaño, esta función deja
de compilar y el error señala el lugar donde hay que decidir.

**Cómo se comprueba:** agregar `| 'jumbo'` al tipo.

```
TS2366: Function lacks ending return statement and return type does not include 'undefined'.
```

Sacar `| 'jumbo'`.

## Paso 5 · Puede no haber café más barato

```ts
export function cheapest(menu: readonly Coffee[]): Coffee | undefined {
  let best: Coffee | undefined = undefined;
  for (const coffee of menu) {
    if (best === undefined || coffee.price < best.price) best = coffee;
  }
  return best;
}
```

**Por qué desaparece el `!`.** El `!` no arreglaba nada: le decía al compilador
que se callara. La lista puede estar vacía y el tipo ahora lo dice.

**Por qué el acumulador arranca en `undefined` y no en `menu[0]`.** Porque
`menu[0]` es justamente lo que puede no existir. Arrancando en `undefined`, la
primera vuelta del `for` lo llena y no hay ningún caso especial que recordar.

**Y lo que hay que mostrar en clase:** la pantalla **sigue funcionando**. El
template ya tenía escrito qué hacer si no hay café más barato:

```html
@if (best) { … } @else { <p class="empty">El menú está vacío.</p> }
```

> El tipo no agregó trabajo. Hizo visible el trabajo que ya había que hacer.

## Pasos 6 y 7 · La extensión

```ts
export interface Page<T> {
  readonly items: readonly T[];
  readonly total: number;
}

export function firstPage<T>(items: readonly T[], size: number): Page<T> {
  return { items: items.slice(0, size), total: items.length };
}

export type CoffeeDraft = Omit<Coffee, 'id'>;
```

**`T` es un hueco con nombre.** `Page<Coffee>` es una página de cafés, `Page<Order>`
una de pedidos, y es la misma interfaz. En `firstPage`, el `<T>` de la firma dice
«esta función funciona con lo que sea»; el que llama no tiene que escribir nada,
porque TypeScript deduce `T` del argumento: `firstPage(MENU, 3)` es
`Page<Coffee>` solo.

**`Omit` deriva en vez de copiar.** La versión escrita a mano funcionaba hasta el
día en que `Coffee` gana un campo. Con `Omit<Coffee, 'id'>`, ese día no existe.

---

# Parte B · Los modelos del hipódromo

**Archivo:** `project/frontend/starter/src/app/core/models/race.model.ts`
**Referencia terminada:** `project/frontend/solution/src/app/core/models/race.model.ts`

Es la misma idea de la Parte A sobre datos reales, y con una diferencia
importante: acá **la verdad no la decide el que escribe, la decide el contrato**.
`docs/contract/openapi.yaml` está en el repo y se abre.

## Paso 1 · `RaceStatus`

```ts
export type RaceStatus = 'upcoming' | 'live' | 'finished';
```

Sale del esquema `RaceStatus` del contrato, que tiene esos tres en un `enum`. Se
copian tal cual, incluidas las minúsculas: son los que manda el backend Go.

## Paso 2 y 3 · `readonly` en todo lo que llega del servidor

```ts
export interface Horse {
  readonly id: string;
  readonly name: string;
  readonly number: number;
  readonly odds: number;
}

export interface Race {
  readonly id: string;
  readonly name: string;
  readonly startsAt: string;
  readonly status: RaceStatus;
  readonly horses: readonly Horse[];
}
```

Y lo mismo en `PodiumEntry`, `Payout` y `RaceResult`, incluidos los dos arrays de
`RaceResult`:

```ts
readonly podium: readonly PodiumEntry[];
readonly payouts: readonly Payout[];
```

**Por qué en un modelo de API esto no es opcional.** Estos objetos son una copia
de lo que dijo el servidor en un momento dado. Editarlos en el cliente no cambia
nada en el servidor: solo hace que la pantalla muestre algo que no es verdad, y
que el próximo `GET` lo pise sin avisar. `readonly` convierte esa disciplina en
algo que el compilador sostiene.

## Paso 4 · Los tres lugares del podio

```ts
export interface PodiumEntry {
  readonly place: 1 | 2 | 3;
  …
}
```

**Los literales también existen para los números.** `1 | 2 | 3` es una unión de
tres tipos, cada uno con un solo valor adentro, igual que `'chico' | 'mediano'`.

## Paso 5 · `favourite` puede no encontrar nada

```ts
export function favourite(race: Race): Horse | undefined {
  return race.horses.reduce<Horse | undefined>((best, horse) => {
    if (!best) return horse;
    if (horse.odds < best.odds) return horse;
    if (horse.odds === best.odds && horse.number < best.number) return horse;
    return best;
  }, undefined);
}
```

**El comportamiento no cambia:** menor cuota gana, y el empate se rompe por
número de partida. Lo único que cambió es que el caso «no hay caballos» dejó de
ser una promesa rota y pasó a ser un valor que el que llama tiene que mirar.

> Si en la Parte A saliste con un `for`, acá también podés. La versión con
> `reduce` es la que está en `solution/`; las dos son correctas y el ejercicio no
> es esa elección.

## La comprobación final

Las cinco líneas del enunciado tienen que dar estos cinco errores:

| | Línea | Error |
|---|---|---|
| a | `race.status === 'galopando'` | `TS2367: … the types 'RaceStatus' and '"galopando"' have no overlap.` |
| b | `horse.odds = 1.01` | `TS2540: Cannot assign to 'odds' because it is a read-only property.` |
| c | `race.horses.push(horse)` | `TS2339: Property 'push' does not exist on type 'readonly Horse[]'.` |
| d | `{ place: 0, … }` | `TS2322: Type '0' is not assignable to type '1 \| 2 \| 3'.` |
| e | `favourite(race).name` | `TS2532: Object is possibly 'undefined'.` |

**Y después se borran.** No son código del proyecto.

---

## Los cinco errores que aparecen todos los años

| Lo que escriben | Por qué no alcanza | Qué decirles |
|---|---|---|
| `status: string` con un comentario `// upcoming, live o finished` | El comentario no lo lee nadie ni lo verifica nada | «El comentario es la intención. La unión es la regla.» |
| `place: number` porque «total son números» | Acepta `0`, `-7` y `99` | «El tipo tiene que decir *cuáles* números, no *que* son números.» |
| `favourite(race)!` para hacer callar el error | El `!` vuelve exactamente al problema anterior, un renglón más abajo | «Cada `!` es una promesa. Preguntate si la podés cumplir.» |
| `readonly horses: Horse[]` | Impide reasignar el campo, no modificar la lista. `push` sigue compilando | «Van los dos `readonly`, y no es repetido.» |
| `notes: string \| undefined` en vez de `notes?: string` | Obliga a escribir `notes: undefined` en cada café que no tenga | «Se parecen, y no son lo mismo. Probá los dos y contame cuál te resulta más cómodo.» |

> El último **casi no es un error**, y conviene tratarlo así: es una diferencia
> real de significado y quien la encontró solo, entendió el tema.

---

## Cómo se verifica que quedó bien

```bash
cd lab/starter            && npx tsc --noEmit && npm test
cd project/frontend/starter && npx tsc --noEmit && npm run build
```

Y desde la raíz del repo, lo que corre el instructor:

```bash
node scripts/verify.mjs --fast
```
