# S5 · Ejercicio 1 — Una comanda, dos mostradores

**Individual · 15 minutos · proyecto `lab/starter`**

---

## Enunciado

El mostrador funciona y la comanda vive dentro del componente. Eso tiene un
techo: el día que haya un segundo mostrador —o un panel de cocina, o una pantalla
de retiro— no hay forma de que vean la misma comanda.

Con `input()` y `output()` tampoco alcanza: conectan padre e hijo, y dos
mostradores uno al lado del otro son hermanos.

Saca el estado a servicios y decide, para cada uno, cuántos tiene que haber.

## Estado inicial

```bash
cd lab/starter
npm start
```

La ruta `/s05` **no existe todavía**. El componente sí está, en
`src/app/sessions/s05/`, con cuatro `TODO(S5)`.

## Datos

Los tres cafés y la forma de la comanda ya están escritos. No se tocan.

---

## Requisitos

### 1. Dos mostradores

Extrae el mostrador a `<app-counter>` —con un `input.required<string>()` para su
nombre— y ponlo **dos veces** en la pantalla: `Mostrador A` y `Mostrador B`.

Además, declara la ruta `/s05` y hazla aparecer en la barra lateral.

### 2. La comanda es un servicio compartido

`OrderService`, con `providedIn: 'root'`.

- El estado va en un `signal` **privado**, y se expone con `asReadonly()`.
- Los cambios pasan por métodos: `add`, `remove`, `clear`.
- Expone `count` y `lastCustomer` como `computed`.

**Nadie de afuera puede llamar a `set` ni a `update`.**

Lo piden los dos mostradores **y también la pantalla**, que ya no recibe nada por
`input()`.

### 3. El cuaderno es un servicio por mostrador

`NotepadService`, **sin `providedIn`**, declarado en `providers` del mostrador.

Cada mostrador anota lo suyo y no ve lo del otro.

### 4. El nombre del café es un token

`SHOP_NAME`, un `InjectionToken<string>` con `providedIn: 'root'` y una `factory`
que devuelva `'Café Compilado'`.

Lo piden la pantalla y los dos mostradores. **Ninguno lo importa como constante.**

---

## Resultado esperado

```
┌── Mostrador A ────────┐  ┌── Mostrador B ────────┐
│ Café Compilado        │  │ Café Compilado        │
│                       │  │                       │
│ Cliente  [ Ana      ] │  │ Cliente  [         ]  │
│ [ Tomar pedido ]      │  │ [ Tomar pedido ]      │
│ La comanda tiene 1    │  │ La comanda tiene 1    │  ← los dos
│                       │  │                       │
│ Anotación [ Leche  ]  │  │ Anotación [        ]  │
│ · Falta leche         │  │ Cuaderno vacío        │  ← solo uno
└───────────────────────┘  └───────────────────────┘

La comanda  1
  Ana · Yirgacheffe   [ Quitar ]
```

**Las dos comprobaciones del ejercicio:**

1. Un pedido tomado en el A hace subir el contador de **los dos**.
2. Una anotación en el A **no** aparece en el B.

## Restricciones

- Prohibido `new OrderService()`. Los servicios se piden, no se construyen.
- Prohibido exponer el signal de escritura desde un servicio.
- Prohibido pasar la comanda por `input()`.
- Prohibido `any`. `standalone: true` y `OnPush`.

## Autoevaluación

- [ ] `/s05` abre y aparece en la barra lateral
- [ ] `npm test` pasa — los tests que ya estaban **siguen** pasando
- [ ] Tomar un pedido en un mostrador lo muestra en el otro
- [ ] Anotar en un mostrador **no** se ve en el otro
- [ ] Desde un componente no hay forma de llamar a `set` sobre la comanda
- [ ] El nombre del café se pide con `inject()`, no se importa
- [ ] No quedó ningún `TODO(S5)`

---

## Pistas

<details>
<summary>Pista 1 — <code>No provider for NotepadService</code></summary>

```
R3InjectorError[NotepadService -> NotepadService]: NullInjectorError:
No provider for NotepadService!
```

El servicio existe y compila. Lo que falta es que **alguien haya dicho dónde
vive**: o `providedIn` en el `@Injectable`, o el `providers` de un componente.
</details>

<details>
<summary>Pista 2 — <code>NG0203</code></summary>

```
NG0203: inject() must be called from an injection context…
```

`inject()` está adentro de un método o de un callback. Va en el **campo de la
clase**:

```ts
protected readonly orders = inject(OrderService);   // ✅
```

Cuando corre un método, Angular ya terminó de construir el componente y no sabe a
qué inyector preguntarle.
</details>

<details>
<summary>Pista 3 — el estado que no se puede escribir desde afuera</summary>

```ts
private readonly _orders = signal<readonly Order[]>([]);
readonly orders = this._orders.asReadonly();
```

El de abajo no tiene `set` ni `update`. Es la forma de que el servicio sea el
dueño de su estado y no un lugar donde cualquiera escribe.
</details>

<details>
<summary>Pista 4 — probar de verdad</summary>

Si probaste con **un solo mostrador**, no probaste nada: los dos servicios se
comportan igual con una sola copia. El segundo mostrador no es decoración, es la
comprobación.
</details>

<details>
<summary>Solución completa</summary>

`correccion.md`, Parte A.
</details>

## Extensión

Agrega un tercer componente, `<app-kitchen>`, que muestre **solo los pedidos
pendientes** y permita marcarlos como entregados.

No recibe nada por `input()` y no es hijo de los mostradores. Si te lleva más de
cinco minutos, algo del reparto quedó atado.

---

> **Referencia:** `conceptos.md` tiene todos los conceptos de la sesión con sus
> ejemplos.
