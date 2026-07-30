# S0 · Ejercicio 1 — Apretar los tipos del menú

**Individual · 15 minutos · proyecto `lab/starter`**

---

## Enunciado

El menú de una cafetería está escrito en TypeScript y funciona. También compila
cuando está mal: acepta tamaños que no existen, obliga a inventar notas de cata
para los cafés que no tienen, promete un café más barato aunque el menú esté
vacío, y deja que cualquiera le agregue productos desde cualquier parte del
programa.

Apretá los tipos hasta que ninguna de esas cuatro cosas se pueda escribir.

## Estado inicial

```bash
cd lab/starter
npm start
```

En <http://localhost:4200/s00> vas a ver el pizarrón del café: una tabla de
precios por tamaño, el más barato, la primera página del menú y un pedido con su
total. **Todo eso ya funciona y tiene que seguir funcionando al terminar.**

El archivo del ejercicio es **`src/app/sessions/s00/menu.ts`**. Los cinco lugares
que hay que tocar están marcados con `TODO(S0)`.

**El archivo `s00.component.ts` no se toca.** Los componentes son el tema de la
clase que viene; si algo se rompe ahí, el arreglo está en `menu.ts`.

## Datos

El menú es el que ya está escrito. No lo cambies, salvo lo que pidan los
requisitos:

| Café | Origen | Precio | Nota de cata |
|---|---|---|---|
| Yirgacheffe | Etiopía | 42 | cítrico y floral |
| Huila | Colombia | 38 | *no tiene* |
| Cerrado | Brasil | 30 | chocolate y nuez |
| Antigua | Guatemala | 45 | *no tiene* |

Los tres tamaños son `chico`, `mediano` y `grande`, y multiplican el precio por
0,8 · 1 · 1,3.

---

## Requisitos

### 1. Los tamaños son tres, y el tipo lo dice

`CoffeeSize` acepta hoy cualquier cadena. Tiene que aceptar exactamente esas tres.

**Verificación:** agregá al final del archivo `sizeFactor('mediana');`. Tiene que
aparecer un error. Después borralo.

### 2. Las notas de cata son opcionales

Huila y Antigua no tienen nota. Hoy llevan cadena vacía porque el tipo obliga a
poner algo.

- Marcá el campo como opcional.
- Borrá los dos `notes: ''` del menú.
- Arreglá `describe()`, que sigue comparando contra la cadena vacía.

**Verificación:** en el pizarrón, la fila de Huila dice `Huila · Colombia` y no
`Huila · Colombia — undefined`.

### 3. El menú no se edita

Ni la lista ni los campos de cada café se pueden modificar en el lugar. Vale
también para `OrderLine` y `Order`.

**Verificación:** agregá al final `MENU.push({ id: 'c9', name: 'Trucho', origin: 'Ninguno', price: 1 });`.
Tiene que aparecer un error. Después borralo.

### 4. El `switch` que se queja solo

Reescribí `sizeFactor()` con un `switch` que cubra los tres tamaños **y no tenga
caso por defecto**.

**Verificación:** agregale `| 'jumbo'` al tipo de tamaños. Tiene que aparecer un
error en `sizeFactor`, no en el tipo. Después sacá `| 'jumbo'`.

### 5. Puede no haber un café más barato

`cheapest()` promete devolver un café siempre, y con un menú vacío explota. Decí
la verdad en el tipo de retorno y sacá el `!`.

**Verificación:** `cheapest([])` no tiene que romper, y el pizarrón tiene que
seguir mostrando el Cerrado.

---

## Resultado esperado

La pantalla se ve **igual que al empezar**, con un solo cambio visible: la fila
de Huila y la de Antigua ya no dicen `— undefined`.

```
┌─────────────────────────────────────────────────────────┐
│  Café                        chico   mediano   grande   │
│  Yirgacheffe · Etiopía —                                │
│    cítrico y floral            34        42       55    │
│  Huila · Colombia              30        38       49    │
│  Cerrado · Brasil —                                     │
│    chocolate y nuez            24        30       39    │
│  Antigua · Guatemala           36        45       59    │
└─────────────────────────────────────────────────────────┘
```

Lo que cambió de verdad no se ve: hay cinco cosas que antes se podían escribir y
ahora no.

## Restricciones

- Prohibido `any`. Si aparece uno, el ejercicio está resuelto al revés.
- Prohibido `!` y `as` para hacer callar un error. Los dos son promesas, y hoy el
  tema es no prometer lo que no se puede cumplir.
- No cambies los nombres de los campos ni el comportamiento de las funciones.
  Solo los tipos, y lo mínimo que haga falta adentro para que sigan andando.
- No toques `s00.component.ts`, `s00.component.html` ni `s00.component.css`.

## Autoevaluación

- [ ] `npx tsc --noEmit` no imprime nada
- [ ] El pizarrón se sigue viendo igual, sin ningún `undefined`
- [ ] `npm test` pasa
- [ ] Las cuatro verificaciones de los requisitos 1, 3 y 4 dan error, y las borré
- [ ] No quedó ningún `any`, ningún `!` ni ningún `as` en el archivo
- [ ] No quedó ningún `TODO(S0)` del 1 al 5

---

## Pistas

<details>
<summary>Pista 1 — el error de las notas</summary>

Si borraste los `notes: ''` antes de marcar el campo como opcional, el error es:

```
TS2741: Property 'notes' is missing in type '{ id: string; … }' but required in type 'Coffee'.
```

Y tiene razón: el tipo todavía dice que la nota es obligatoria. El orden que
menos duele es **primero el tipo, después los datos**.
</details>

<details>
<summary>Pista 2 — dónde va exactamente <code>readonly</code></summary>

Son **dos lugares distintos** y hacen cosas distintas:

```ts
readonly price: number;              // el campo no se puede reasignar
readonly horses: readonly Horse[];   // …y además la lista no se puede modificar
```

El primero solo, sin el segundo, deja pasar `push`. Lo vas a ver otra vez en el
bloque de predicciones.
</details>

<details>
<summary>Pista 3 — el tipo de <code>cheapest</code></summary>

Un valor que puede no estar se escribe con una unión: `Coffee | undefined`.

Con el `!` afuera, `menu[0]` pasa a ser `Coffee | undefined` y el compilador no
te va a dejar leerle el precio hasta que preguntes si existe. Esa pregunta es la
solución, no el obstáculo.
</details>

<details>
<summary>Solución completa</summary>

`correccion.md`, Parte A: el paso a paso con el porqué de cada decisión. Usalo
para destrabarte y seguí desde ahí — no lo copies entero.
</details>

## Extensión

Quedan dos `TODO(S0)` más, el 6 y el 7, y son los dos temas que hoy solo miramos:

- **6 · Genéricos.** `CoffeePage` sirve solo para cafés. Convertila en `Page<T>`
  y hacé que `firstPage` la acompañe, de manera que `firstPage(MENU, 3)` siga
  funcionando sin decirle a mano qué contiene.
- **7 · Utility types.** `CoffeeDraft` está copiada a mano de `Coffee`. Derivala
  del original con `Omit`, y comprobá que agregarle un campo a `Coffee` ahora
  aparece solo en el borrador.

---

> **Referencia:** `conceptos.md` tiene todos los conceptos de la sesión con sus
> ejemplos. Está pensado para tenerlo abierto al lado del editor.
