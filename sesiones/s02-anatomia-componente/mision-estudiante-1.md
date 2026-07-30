# S2 · Ejercicio 1 — Partir la carta en dos

**Individual · 15 minutos · proyecto `lab/starter`**

---

## Enunciado

La carta de la cafetería funciona: cuatro cafés, cada uno con su cantidad, su
total y su botón de pedir. Está escrita entera adentro de un solo componente, y
por eso no se puede mostrar un café en ninguna otra pantalla sin copiar veinte
líneas de HTML.

Sacá la tarjeta de un café a su propio componente, con sus entradas y sus
salidas, de manera que la pantalla siga viéndose y funcionando igual.

## Estado inicial

```bash
cd lab/starter
npm start
```

La ruta `/s02` **no existe todavía**: declararla es el primer requisito. El
componente de la pantalla sí está, en `src/app/sessions/s02/`, y los lugares que
hay que tocar están marcados con `TODO(S2)`.

## Datos

Los cuatro cafés ya están en `menu.ts` y no se tocan. El Cerrado es el único sin
stock, y el Yirgacheffe es el café del día.

---

## Requisitos

### 1. La pantalla es alcanzable

Declará la ruta `/s02` y hacela aparecer en la barra lateral. Son **dos**
archivos, como en S1.

**Verificación:** `/s02` abre la carta y «Anatomía de un componente» aparece en
el menú de la izquierda.

### 2. La tarjeta es un componente

Creá `<app-coffee-card>` con el CLI y mudale el `<article class="card">` entero,
con su CSS.

**El CSS se muda con el marcado.** Un componente cuyo estilo vive en el archivo
de otro no se puede llevar a ninguna parte, que es justo lo que veníamos a
arreglar.

### 3. Las entradas

| Entrada | Tipo | Comportamiento |
|---|---|---|
| `coffee` | `Coffee` | **Obligatoria.** Sin ella, no compila |
| `featured` | `boolean` | Opcional. Sin ella vale `false` |
| `quantity` | `number` | Va y vuelve: el padre la escribe con `[(quantity)]` |

**Verificación:** escribí `<app-coffee-card />` sin `[coffee]`. Tiene que
aparecer un error de compilación. Después borralo.

### 4. La salida

Cuando alguien toca «Pedir», la tarjeta **avisa**; no escribe en la comanda. El
padre escucha ese aviso y agrega la línea.

El aviso lleva el café y la cantidad juntos, en un solo objeto.

### 5. El contenido proyectado

Dos huecos en la tarjeta:

- Uno **con `select`**, arriba, donde el padre pone el rótulo `Café del día`.
- Uno **sin `select`**, en el medio, donde el padre pone el aviso
  `Vuelve el jueves.` de los cafés sin stock.

La tarjeta no decide ninguno de los dos textos.

### 6. El ciclo de vida

La tarjeta implementa los tres ganchos:

- `ngOnInit` guarda la hora en que se montó, y se muestra al pie.
- `ngOnChanges` cuenta cuántas veces cambió un input, y se muestra al lado.
- `ngOnDestroy` emite un segundo `output()` con el nombre del café.

En la pantalla, un panel «Ciclo de vida» lista lo que avisaron las tarjetas al
destruirse, y un botón saca el Antigua de la carta y lo devuelve.

---

## Resultado esperado

```
┌──────────────────┐ ┌──────────────────┐
│ CAFÉ DEL DÍA     │ │                  │
│ Yirgacheffe      │ │ Huila            │
│ Etiopía          │ │ Colombia         │
│ 42               │ │ 38               │
│  − [ 1 ] +       │ │  − [ 1 ] +       │
│ Total: 42        │ │ Total: 38        │
│ [ Pedir ]        │ │ [ Pedir ]        │
│ montado 14:03    │ │ montado 14:03    │
│ ngOnChanges ×1   │ │ ngOnChanges ×1   │
└──────────────────┘ └──────────────────┘

[ Sacar el Antigua de la carta ]

Comanda                    Ciclo de vida
1 × Yirgacheffe · 42       ngOnDestroy · Antigua
```

La pantalla se ve **igual que al empezar**, más el panel de ciclo de vida.

## Restricciones

- La tarjeta **no importa `MENU`** ni sabe qué es una comanda. Si necesita
  importar algo de la pantalla, el corte está mal hecho.
- La tarjeta **no escribe** en la comanda. Avisa.
- Prohibido `any`. Los dos componentes son `standalone: true` y `OnPush`.
- No cambies el comportamiento: mismos textos, mismos límites de cantidad (1 a
  20), mismo botón deshabilitado sin stock.

## Autoevaluación

- [ ] `/s02` abre y aparece en la barra lateral
- [ ] `npm test` pasa — los tests que ya estaban **siguen** pasando
- [ ] `npx tsc --noEmit` no imprime nada
- [ ] Tapando el archivo del padre, se puede decir qué hace la tarjeta y qué necesita
- [ ] La tarjeta no importa nada de la pantalla
- [ ] Pedir dos cafés distintos deja **dos** líneas en **una sola** comanda
- [ ] Sacar el Antigua de la carta deja su nombre en el panel de ciclo de vida
- [ ] No quedó ningún `TODO(S2)`

---

## Pistas

<details>
<summary>Pista 1 — el orden que menos duele</summary>

**Cortá primero, pensá después.** Mudá el `<article>` entero al hijo y guardá sin
arreglar nada: van a aparecer unos quince errores, todos diciendo que `item` no
existe.

Esa lista de errores **es** la lista de inputs que hace falta. No hay que
adivinarla.
</details>

<details>
<summary>Pista 2 — el hijo no aparece</summary>

Si en la pantalla no se ve nada donde debería estar la tarjeta, o el navegador se
queja de una etiqueta desconocida, falta esto en el `@Component` del padre:

```ts
imports: [CoffeeCardComponent],
```

Es la misma regla de `FormsModule` en S1: **si el template usa algo, va en
`imports`**. Ahora vale también para los componentes propios.
</details>

<details>
<summary>Pista 3 — <code>coffee.name</code> no compila</summary>

Un `input()` es una **función**. En el template se lee con paréntesis:

```html
{{ coffee().name }}
```

Sin los paréntesis estás pidiéndole el `name` a la función, no al café.
</details>

<details>
<summary>Pista 4 — qué va en el aviso</summary>

`menu.ts` ya tiene el tipo escrito:

```ts
export interface OrderRequest {
  readonly coffee: Coffee;
  readonly quantity: number;
}
```

Se usa así: `output<OrderRequest>()` en el hijo, `(ordered)="take($event)"` en el
padre. Ese `$event` **es** el `OrderRequest`.
</details>

<details>
<summary>Solución completa</summary>

`correccion.md`, Parte A: el paso a paso con el porqué de cada decisión.
</details>

## Extensión

La tarjeta tiene un `ngOnChanges` que cuenta cambios de input. Probá esto y
explicá lo que ves, en un comentario:

1. Subí la cantidad con el botón `+`. El contador se mueve, aunque el cambio lo
   haya hecho la tarjeta y no la pantalla. **¿Por qué?**
2. Sacá el Antigua de la carta y devolvelo. ¿Qué dice su contador al volver, y
   qué dice el de las otras tarjetas?

La respuesta a la 1 es lo que hace que `model()` sea distinto de una propiedad
común, y la 2 dice algo sobre qué significa «el mismo componente».

---

> **Referencia:** `conceptos.md` tiene todos los conceptos de la sesión con sus
> ejemplos. Está pensado para tenerlo abierto al lado del editor.
