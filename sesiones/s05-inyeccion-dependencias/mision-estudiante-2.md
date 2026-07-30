# S5 · Ejercicio 2 — El estado sale de la pantalla

**En parejas · 20 minutos · proyecto `project/frontend/starter`**

---

## Enunciado

El filtro, la búsqueda y la carrera abierta viven dentro de `race-list`. Funciona
mientras haya una sola pantalla que los use — y en la sesión 10 va a haber una
segunda, la de la carrera en vivo, que necesita saber exactamente lo mismo.

Muda el estado a un servicio y deja en la pantalla solo lo que de verdad es suyo.

## Estado inicial

El punto de partida es **el listado que terminaste en S4**, con los pipes y la
directiva ya funcionando. Si te quedó a medias,
`sesiones/s04-directivas-pipes/correccion.md` lo deja andando.

---

## Requisitos

### 1. `RaceStore`, en `core/races/`

Se muda **tal cual**: los `computed` que escribiste en S3 se copian y pegan. Esto
no es una reescritura, es un cambio de dueño.

| Qué expone | Cómo |
|---|---|
| `filter` · `query` | signals de solo lectura |
| `counts` · `visible` · `selected` · `lineup` | `computed` |
| `setFilter` · `setQuery` · `toggle` · `clearFilters` | métodos |

- El programa entero **sigue sin ser un signal**: viene de una constante.
- La carrera abierta se sigue derivando **del id**.
- Los signals de escritura son `private`.

### 2. `BetStore`, en `core/bets/`

El simulador de apuesta, y **inyecta a `RaceStore`**.

| Qué expone | De dónde sale |
|---|---|
| `amount` | signal de solo lectura |
| `target` | el favorito de la carrera abierta, del otro store |
| `payout` | el monto por la cuota del favorito |
| `isValid` | los límites del contrato: `MIN_BET_AMOUNT` y `MAX_BET_AMOUNT` |

Un servicio puede pedir servicios. Se hace con `inject()`, igual que en un
componente.

### 3. `API_URL`, en `core/config/`

Un `InjectionToken<string>` con `providedIn: 'root'` y una `factory` que devuelva
`'/api'`.

Todavía no lo usa nadie: en S7 va a apuntar al backend Go o al mock, y cambiarlo
tiene que ser **una línea**.

Escribe en su comentario **por qué no es una constante exportada**.

### 4. Lo que se queda en la pantalla

`race-list` conserva únicamente lo de presentación:

- `STATUS_LABELS` y `STATUS_TONES`
- el formato de la hora
- la traducción de lo que trae el store a lo que dibuja el template

**Si algo de eso también se fue al store, está de más ahí:** otra vista de las
mismas carreras podría querer otras etiquetas y otro formato de hora.

### 5. El monto sigue funcionando

`[(ngModel)]` necesita algo con `set`, y el store no lo expone. Resuélvelo con un
getter y un setter en el componente que deleguen en el store.

---

## Resultado esperado

**La aplicación se ve exactamente igual que al terminar S4.** Todo sigue
funcionando: el filtro, la búsqueda, el panel que se cierra solo, el pago.

Lo único visible que se agrega es el aviso cuando el monto queda fuera de rango:

```
Monto  [ 9000 ]
Payador paga  24.750 pts
El monto tiene que estar entre 10 y 5000.
```

## Restricciones

- `core/` **no importa de `features/`**. Es la regla de dependencias, y hoy es
  media consigna.
- Los signals de escritura de los stores son `private`.
- Prohibido `new RaceStore()`.
- Prohibido `any`. `OnPush` en todo.

## Autoevaluación

- [ ] `npm run build` pasa
- [ ] La aplicación se ve igual y todo sigue funcionando
- [ ] Desde `race-list` no hay forma de llamar a `set` sobre el filtro
- [ ] `BetStore` inyecta `RaceStore` y no duplica nada suyo
- [ ] En `race-list` no quedó ni un `signal(` de estado
- [ ] `STATUS_LABELS` **sí** quedó en `race-list`
- [ ] No quedó ningún `TODO(S5)`

---

## Pistas

<details>
<summary>Pista 1 — qué se muda y qué se queda</summary>

La pregunta para cada cosa: **¿otra pantalla necesitaría exactamente esto?**

- El filtro y la búsqueda: sí → al store.
- Cuál carrera está abierta: sí → al store.
- Que `finished` se muestre como «Terminada»: no necesariamente. Otra vista
  podría querer «Finalizada», o un icono → se queda.
</details>

<details>
<summary>Pista 2 — un servicio que inyecta a otro</summary>

```ts
@Injectable({ providedIn: 'root' })
export class BetStore {
  private readonly races = inject(RaceStore);
  …
}
```

Igual que en un componente, y en el campo de la clase. El inyector no distingue
quién pregunta.
</details>

<details>
<summary>Pista 3 — <code>[(ngModel)]</code> contra un store de solo lectura</summary>

```ts
protected get amount(): number {
  return this.bets.amount();
}

protected set amount(value: number) {
  this.bets.setAmount(value);
}
```

El componente hace de traductor. El store sigue siendo el único que escribe su
estado.
</details>

<details>
<summary>Solución completa</summary>

`correccion.md`, Parte B.
</details>

## Extensión

Escribe un test de `RaceStore` **sin ningún componente**:

```ts
TestBed.configureTestingModule({});
const store = TestBed.inject(RaceStore);
store.setFilter('live');
expect(store.visible().length).toBe(1);
```

Y después contesta, en un comentario: **¿por qué esto no se podía hacer cuando el
estado vivía en `race-list`?**

La respuesta es media razón por la que se mueve el estado a un servicio, y la
otra media es la de la sesión 10.

---

> **Referencia:** `conceptos.md` tiene todos los conceptos de la sesión con sus
> ejemplos.
