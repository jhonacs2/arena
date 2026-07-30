# S0 · Ejercicio 2 — El contrato del hipódromo

**En parejas · 20 minutos · proyecto `project/frontend/starter`**

---

## Enunciado

El hipódromo tiene un backend escrito en Go y un contrato que describe, campo por
campo, qué manda y qué garantiza. Los modelos del frontend copian bien los
**nombres** de ese contrato, pero sus **tipos** son más flojos que lo que el
contrato promete: dicen «una cadena» donde el contrato dice «una de estas tres»,
y dicen «siempre hay» donde el contrato no lo puede garantizar.

Apretá los tipos hasta que digan exactamente lo que el contrato dice.

## Estado inicial

```bash
cd project/frontend/starter
npm start
```

El archivo del ejercicio es **`src/app/core/models/race.model.ts`**. Los cinco
lugares que hay que tocar están marcados con `TODO(S0)`.

La fuente de verdad es **`docs/contract/openapi.yaml`**. Está en el repo y se
abre: nada de lo que se pide acá hay que adivinarlo.

## Datos

Del contrato, y son innegociables porque hay código Go del otro lado que espera
exactamente esto:

| Qué | Qué garantiza el contrato |
|---|---|
| `status` de una carrera | Exactamente tres valores: `upcoming`, `live`, `finished` |
| `place` del podio | Exactamente tres lugares: 1, 2, 3 |
| Carreras y caballos | Son datos que llegan del servidor. El frontend los muestra, no los edita |
| Caballos de una carrera | El contrato exige al menos dos, pero un tipo no puede exigir eso |

**Los nombres de los campos no se tocan.** Ni uno.

---

## Requisitos

### 1. Los estados de una carrera son tres

`RaceStatus` acepta hoy cualquier cadena.

### 2. Un caballo no se edita

Ningún campo de `Horse` se puede reasignar.

### 3. Una carrera no se edita, y su lista de caballos tampoco

Vale lo mismo para `Race`, `PodiumEntry`, `Payout` y `RaceResult`: campos y
listas.

### 4. El podio tiene tres lugares

`place` vale 1, 2 o 3. Ninguna otra cosa.

### 5. Puede no haber favorito

`favourite()` promete devolver un caballo siempre, y lo consigue con un `!` que
es mentira: con una carrera sin caballos explota en tiempo de ejecución. Decí la
verdad en el tipo de retorno y sacá el `!`.

---

## Resultado esperado

**La aplicación se ve exactamente igual que antes.** No cambia ni un píxel: los
tipos no existen cuando el programa corre.

Lo que cambia es lo que ya no se puede escribir. Pegá estas cinco líneas al final
del archivo, una por una, y comprobá que cada una da el error de la derecha:

```ts
declare const race: Race;
declare const horse: Horse;

// a
if (race.status === 'galopando') {
}

// b
horse.odds = 1.01;

// c
race.horses.push(horse);

// d
const entry: PodiumEntry = { place: 0, horseId: 'h1', horseName: 'X', number: 1, odds: 2 };

// e
favourite(race).name;
```

| | El error que tiene que aparecer |
|---|---|
| **a** | `TS2367: This comparison appears to be unintentional because the types 'RaceStatus' and '"galopando"' have no overlap.` |
| **b** | `TS2540: Cannot assign to 'odds' because it is a read-only property.` |
| **c** | `TS2339: Property 'push' does not exist on type 'readonly Horse[]'.` |
| **d** | `TS2322: Type '0' is not assignable to type '1 \| 2 \| 3'.` |
| **e** | `TS2532: Object is possibly 'undefined'.` |

**Cuando los cinco aparezcan, borrá el bloque entero.** No es código del
proyecto: es la prueba de que el ejercicio está hecho.

## Restricciones

- Prohibido `any`, `!` y `as`.
- No se cambia ningún nombre de campo. Son el contrato.
- No se cambia el comportamiento de `favourite()`: sigue devolviendo el de menor
  cuota, y el empate se sigue rompiendo por número de partida.
- No se toca ningún otro archivo. Si algo más deja de compilar, avisá: eso es
  material de clase.

## Autoevaluación

- [ ] `npm run build` pasa
- [ ] La aplicación se ve igual que antes
- [ ] Las cinco líneas de arriba dieron sus cinco errores, y las borré
- [ ] No quedó ningún `any`, ningún `!` ni ningún `as`
- [ ] No quedó ningún `TODO(S0)`
- [ ] Los nombres de los campos son los mismos que al empezar

---

## Pistas

<details>
<summary>Pista 1 — dónde dice el contrato los tres estados</summary>

`docs/contract/openapi.yaml`, en `components.schemas`. Hay un esquema que se
llama **`RaceStatus`** —el mismo nombre que el tipo— y tiene un `enum` con los
tres valores, escritos tal cual hay que copiarlos. `Race.status` no los repite:
lo referencia.

Lo mismo para el podio: esquema **`PodiumEntry`**, campo `place`, con
`enum: [1, 2, 3]`.
</details>

<details>
<summary>Pista 2 — el <code>readonly</code> de la lista</summary>

```ts
readonly horses: readonly Horse[];
```

Aparece dos veces y no es repetido: el primero impide reasignar el campo, el
segundo impide modificar la lista. Con uno solo, `push` sigue compilando.
</details>

<details>
<summary>Pista 3 — el favorito que puede no estar</summary>

Al sacar el `!`, `race.horses[0]` pasa a ser `Horse | undefined` — eso lo hace
`noUncheckedIndexedAccess`, que está prendido en el `tsconfig` del proyecto.

La salida no es volver a poner el `!`: es arrancar el acumulador en `undefined` y
preguntar antes de comparar cuotas.
</details>

<details>
<summary>Solución completa</summary>

`correccion.md`, Parte B: el archivo terminado, con el porqué de cada línea.
</details>

## Extensión

Abrí `src/app/core/models/api.model.ts` y leelo entero. Después contestá, por
escrito, en un comentario al final de ese archivo:

1. `Page<T>` es genérico. ¿Qué dos tipos distintos del proyecto lo van a usar, y
   qué se ganó con no escribir `RacePage` y `BetPage` por separado?
2. `ApiErrorCode` es una unión de dieciocho literales en vez de `string`. ¿Qué
   pasa el día que el backend agrega un código nuevo y alguien lo suma a esa
   unión?

La respuesta a la 2 es el motivo por el que la unión está ahí, y se va a cobrar
en la sesión 7.

---

> **Referencia:** `conceptos.md` tiene todos los conceptos de la sesión con sus
> ejemplos. Está pensado para tenerlo abierto al lado del editor.
