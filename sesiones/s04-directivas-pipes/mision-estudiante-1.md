# S4 · Ejercicio 1 — Sacar el formateo del componente

**Individual · 15 minutos · proyecto `lab/starter`**

---

## Enunciado

La carta funciona, y todo el formateo vive adentro del componente: un método que
arma los importes, otro que fabrica un array de longitud N solo para poder
dibujar N puntos, y una condición repetida dos veces para resaltar el café del
día.

Ninguna de esas tres cosas es lógica de esta pantalla: son formas de mostrar, y
las formas de mostrar se repiten en toda la aplicación.

Sácalas a un pipe y dos directivas.

## Estado inicial

```bash
cd lab/starter
npm start
```

La ruta `/s04` **no existe todavía**. El componente sí está, en
`src/app/sessions/s04/`, con cuatro `TODO(S4)`.

## Datos

Los cuatro cafés ya están en el componente y no se tocan. Los precios están en
pesos enteros; el puntaje va de 0 a 5.

---

## Requisitos

### 1. Los pipes que ya vienen

Sin escribir ninguno propio, y sin tocar el componente:

| Qué | Cómo se ve |
|---|---|
| El origen | `ETIOPÍA` |
| El stock | `12` con separador de miles cuando corresponda |
| La proporción | `54,5 %` |

**El porcentaje tiene que salir con coma y en español.** Si sale `54.5%`, falta
declarar el idioma de la aplicación: son dos líneas en `app.config.ts`, y ya
están escritas ahí.

### 2. Un pipe propio: `money`

```
{{ 4200 | money }}        → $ 4.200
{{ 4200 | money: 'USD' }} → USD 4.200
```

- Sin decimales, con separador de miles.
- El símbolo es un **parámetro con valor por defecto**.
- Es **puro**. Escríbelo aunque sea el valor por defecto.
- El método `formatMoney()` desaparece del componente.

### 3. Una directiva de atributo: `appHighlight`

```html
<li class="card" [appHighlight]="coffee.id === featuredId" highlightLabel="Café del día">
```

- Pone la clase `is-highlighted` cuando la condición es verdadera.
- Pone el atributo `data-highlight-label` con el texto, **solo** cuando está
  resaltado. Cuando no, el atributo no existe.
- El texto es un input con valor por defecto `Destacado`.

**Del template desaparecen las dos condiciones repetidas y el nombre de la clase
CSS.**

### 4. Una directiva estructural: `*appRepeat`

```html
<span class="bean" *appRepeat="coffee.rating">●</span>
```

Dibuja el contenido tantas veces como diga el número. El método `beansFor()`
desaparece del componente.

Necesita `TemplateRef` y `ViewContainerRef`, y las dos se piden con `inject()`.

---

## Resultado esperado

**La pantalla se ve igual que al empezar**, con dos diferencias visibles: el
origen en mayúsculas y el porcentaje en español.

```
┌──────────────────────┐
│ CAFÉ DEL DÍA         │  ← el rótulo lo pone la directiva
│ ETIOPÍA              │
│ Yirgacheffe          │
│ $ 4.200              │
│ USD 4.200            │
│ ● ● ● ● ●            │  ← cinco, dibujados por *appRepeat
│ 12 en depósito ·     │
│ 54,5 % del total     │
└──────────────────────┘
```

Lo que cambió no se ve: el componente ya no formatea nada.

## Restricciones

- El pipe **no sabe qué es un café**. Recibe un número, no el objeto.
- Prohibido `any`. Los tres archivos nuevos llevan `standalone: true`.
- No cambies los datos ni el CSS.
- El componente queda **sin** `formatMoney()` ni `beansFor()`.

## Autoevaluación

- [ ] `/s04` abre y aparece en la barra lateral
- [ ] `npm test` pasa — los tests que ya estaban **siguen** pasando
- [ ] El porcentaje sale `54,5 %`, con coma
- [ ] Los cafés que no son el del día **no** tienen el atributo del rótulo
- [ ] El componente no tiene ni un método de formateo
- [ ] Abriendo solo `money.pipe.ts`, no hay forma de saber que es de una cafetería
- [ ] No quedó ningún `TODO(S4)`

---

## Pistas

<details>
<summary>Pista 1 — <code>No pipe found with name 'money'</code></summary>

El pipe existe, pero **este componente** no lo declaró. Va en `imports`, igual
que un componente o que `FormsModule`.

Y si ya está en `imports` y el error sigue: el `name` del `@Pipe` tiene que
coincidir **exacto** con lo que escribes en el template. Se comparan cadenas, no
clases.
</details>

<details>
<summary>Pista 2 — el selector de una directiva</summary>

```ts
@Directive({
  selector: '[appHighlight]',
  standalone: true,
  host: {
    '[class.is-highlighted]': 'appHighlight()',
  },
})
export class HighlightDirective {
  readonly appHighlight = input(false);
}
```

Los corchetes en el selector quieren decir *cualquier elemento que tenga este
atributo*. Y el input se llama igual que el selector para poder escribir
`[appHighlight]="condición"` una sola vez.
</details>

<details>
<summary>Pista 3 — el atributo que a veces no está</summary>

```ts
'[attr.data-highlight-label]': 'appHighlight() ? highlightLabel() : null',
```

`null` **quita** el atributo. Una cadena vacía lo deja puesto y vacío, que no es
lo mismo.
</details>

<details>
<summary>Pista 4 — la directiva estructural</summary>

```ts
private readonly template = inject(TemplateRef<unknown>);
private readonly container = inject(ViewContainerRef);
```

`TemplateRef` es el pedazo de HTML guardado; `ViewContainerRef` es el lugar donde
insertarlo. El ciclo es: limpiar el contenedor y crear N vistas.

Para que se rehaga cuando cambia el número, envuélvelo en un `effect()`.
</details>

<details>
<summary>Solución completa</summary>

`correccion.md`, Parte A.
</details>

## Extensión

Agrega un pipe `truncate` que corte un texto largo y le ponga puntos suspensivos:

```
{{ 'Café de finca con notas cítricas' | truncate: 12 }}  → Café de fin…
```

Y contesta, en un comentario: **¿tiene que ser puro o impuro?** ¿Por qué?

---

> **Referencia:** `conceptos.md` tiene todos los conceptos de la sesión con sus
> ejemplos.
