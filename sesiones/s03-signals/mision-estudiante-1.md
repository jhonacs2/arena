# S3 · Ejercicio 1 — El tablero que se calcula solo

**Individual · 15 minutos · proyecto `lab/starter`**

---

## Enunciado

El tablero de la comanda muestra una lista de pedidos y deja avanzarlos,
quitarlos y agregar uno nuevo. Funciona.

Lo que no tiene es nada derivado: no se puede filtrar por estado, no se puede
buscar, no hay contadores y no se sabe cuánto falta cobrar. Agregar cada una de
esas cosas guardándolas en una propiedad significa acordarse de actualizarlas en
los cuatro lugares donde la comanda cambia.

Pasá el estado a signals y hacé que todo lo demás se calcule solo.

## Estado inicial

```bash
cd lab/starter
npm start
```

La ruta `/s03` **no existe todavía**. El componente sí está, en
`src/app/sessions/s03/`, con seis `TODO(S3)`.

## Datos

Las cinco comandas iniciales ya están en `orders.ts`, junto con
`STATUS_LABELS`, `lineTotal()` y `nextStatus()`. No se tocan.

Los cuatro filtros son `Todas`, `Pendientes`, `Listas` y `Entregadas`.

---

## Requisitos

### 1. La pantalla es alcanzable

Declará la ruta `/s03` y hacela aparecer en la barra lateral. Dos archivos.

### 2. La comanda es un signal

`orders` pasa a ser un `signal<readonly Order[]>`, y `advance`, `remove`, `add`
y `reset` lo cambian con `update` o `set`.

**Ni un `push`, ni un `sort`, ni una asignación a una propiedad de un pedido.**

### 3. El filtro y la búsqueda

Dos signals más: el filtro elegido y el texto del buscador.

- Una fila de pestañas con los cuatro filtros. La activa se ve distinta.
- Cada pestaña muestra **cuántas comandas hay de ese estado** — del total, no de
  lo que se está viendo.
- Un campo de búsqueda que filtra por cliente **o** por café, sin distinguir
  mayúsculas.

El buscador necesita `FormsModule` en `imports`, como en S1.

### 4. Todo lo demás es `computed`

| Qué | De dónde sale |
|---|---|
| Lo que se ve | la comanda, filtrada por estado y por texto |
| Los contadores | la comanda entera |
| El total por cobrar | lo que no está entregado |
| Las más caras | lo que se ve, ordenado por importe |

**Ninguno de los cuatro se guarda en una propiedad ni se actualiza a mano.**

### 5. El control flow

- El `@for` recorre lo que se ve y lleva **`track` por id**.
- El texto del botón de avanzar sale de un `@switch` sobre el estado:
  `Marcar lista` → `Entregar` → `Entregada`.
- Cuando no hay nada que mostrar, un `@switch` sobre el filtro elige el mensaje:

  | Filtro | Mensaje |
  |---|---|
  | Pendientes | `No queda ninguna pendiente. La barra está al día.` |
  | Listas | `Ninguna lista para entregar.` |
  | Entregadas | `Todavía no se entregó ninguna.` |
  | cualquier otro | `No hay comandas que coincidan con la búsqueda.` |

### 6. Ordenar sin romper

El panel «Las más caras» muestra lo que se ve ordenado de mayor a menor importe.

**El orden de la comanda no cambia.** Si al mirar ese panel la lista de arriba se
reordena, está mal.

---

## Resultado esperado

```
[ Todas 5 ] [ Pendientes 2 ] [ Listas 2 ] [ Entregadas 1 ]

Buscar por cliente o café  [                    ]

En pantalla 5 de 5
  Ana     2 × Yirgacheffe   Pendiente   84   [Marcar lista] [Quitar]
  Beto    1 × Huila         Lista       38   [Entregar]     [Quitar]
  Carla   3 × Cerrado       Entregada   90   [Entregada]    [Quitar]
  …

┌─ Por cobrar ─┐  ┌─ Las más caras ──┐
│     209      │  │ Carla · 90       │
└──────────────┘  │ Ana · 84         │
                  └──────────────────┘
```

Con el filtro en «Pendientes» y todo marcado como listo:
`No queda ninguna pendiente. La barra está al día.`

## Restricciones

- Prohibido `push`, `sort`, `splice` y asignar propiedades sobre el estado.
- Prohibido guardar en una propiedad algo que se pueda calcular.
- Prohibido `any`. `standalone: true` y `OnPush`.
- No cambies `orders.ts`.

## Autoevaluación

- [ ] `/s03` abre y aparece en la barra lateral
- [ ] `npm test` pasa — los tests que ya estaban **siguen** pasando
- [ ] Hay exactamente **tres** signals, y ninguno guarda algo derivado
- [ ] Agregar una comanda mueve la lista **y** el total, a la vez
- [ ] Filtrar por «Pendientes» y vaciarlo muestra el mensaje de la barra al día
- [ ] Mirar el panel de las más caras no reordena la lista de arriba
- [ ] No quedó ningún `TODO(S3)`

---

## Pistas

<details>
<summary>Pista 1 — el orden que menos duele</summary>

Primero convertí `orders` en signal y poné paréntesis hasta que compile. **Sin
agregar nada nuevo.** La pantalla va a quedar igual que antes, y eso está bien.

Recién después agregá lo derivado. Si arrancás por el filtro, vas a estar
peleando dos cosas a la vez.
</details>

<details>
<summary>Pista 2 — <code>set</code> o <code>update</code></summary>

```ts
this.orders.set(INITIAL_ORDERS);                       // reemplazo el valor entero
this.orders.update((orders) => [...orders, nueva]);    // parto de lo que hay
```

Si necesitás mirar lo que había, es `update`.
</details>

<details>
<summary>Pista 3 — un computed no se actualiza</summary>

Si te dan ganas de escribir `this.visible.set(...)`, pará: un `computed` no tiene
`set`. Se recalcula solo.

Y si un valor de verdad necesita `set`, entonces no era derivado: era estado, y
va en un `signal`.
</details>

<details>
<summary>Pista 4 — el orden que no rompe nada</summary>

```ts
[...this.visible()].sort(…)   // ordena una copia
this.visible().sort(…)        // ordena el original, y rompe el de arriba
```

`sort()` ordena **en el lugar** y devuelve el mismo array. El `[...]` de adelante
es todo lo que hace falta.
</details>

<details>
<summary>Solución completa</summary>

`correccion.md`, Parte A.
</details>

## Extensión

Agregá un botón «Entregar todas las listas» que avance de una sola vez todas las
que estén en `ready`.

Tiene que ser **un solo `update`**, no un `update` por comanda. Si escribís un
bucle que llama a `advance` cinco veces, funciona igual — pero anotá por qué
preferimos el otro.

---

> **Referencia:** `conceptos.md` tiene todos los conceptos de la sesión con sus
> ejemplos.
