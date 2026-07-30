# S4 · Ejercicio 2 — Un solo lugar para cada formato

**En parejas · 20 minutos · proyecto `project/frontend/starter`**

---

## Enunciado

`horse.odds.toFixed(2)` está escrito en la parrilla, y va a estar en el historial
de apuestas, en el resultado de cada carrera y en el panel de la carrera en vivo.
El día que alguien pida las cuotas con coma en vez de punto hay que encontrarlos
todos.

Además, `toFixed` tiene un problema que nadie notó todavía: **siempre usa el
punto como separador decimal**, escriba lo que escriba el resto de la pantalla.

Escribe dos pipes y una directiva, y deja cada formato en un solo lugar.

## Estado inicial

El punto de partida es **el listado que terminaste en S3**, con el filtro y la
búsqueda funcionando.

```bash
cd project/frontend/starter
npm start
```

## Datos

Las ocho carreras de `core/mocks`. El saldo del proyecto es virtual y entero: no
hay centavos y no hay moneda real, así que la unidad se llama `pts`.

---

## Requisitos

### 1. `MoneyPipe`, en `shared/pipes/`

```
{{ 1250 | money }}     → 1.250 pts
{{ 1250 | money: '' }} → 1.250
```

Sin decimales, con separador de miles, y la unidad como parámetro con valor por
defecto.

**`pts` tiene que estar escrito una sola vez en todo el proyecto**, y ese lugar
es este archivo.

### 2. `OddsPipe`, en `shared/pipes/`

```
{{ 2.4 | odds }} → 2,40
{{ 9 | odds }}   → 9,00
```

Siempre dos decimales, **y con coma**. Reemplaza a todos los `toFixed(2)`.

### 3. `FavouriteDirective`, en `shared/directives/`

```html
<li class="runner" [appFavourite]="horse.id === view.favourite?.id">
```

- Pone la clase `is-favourite`.
- Pone `data-favourite-label` con el texto, solo cuando corresponde.
- El texto es un input con valor por defecto `Favorito`.

El favorito de una carrera **es información, no decoración**: quien navega con
lector de pantalla también tiene que enterarse. Por eso va el atributo y no solo
un color.

### 4. El idioma de la aplicación

Declara `LOCALE_ID` en `app.config.ts`. Sin eso, cualquier pipe incorporado que
se use más adelante va a formatear en inglés.

### 5. Que no quede ninguno suelto

Al terminar, `toFixed(` no aparece ni una vez en `src/app/features/`.

---

## Resultado esperado

**La aplicación se ve casi igual.** Las dos diferencias visibles:

```
Parrilla · por cuota
  3  ▮ Payador · FAVORITO        2,75    ← coma, y el rótulo lo pone la directiva
  7  ▮ Trueno Manso              3,10
  1  ▮ Viento Norte              4,00    ← antes decía 4.00

Payador paga    275 pts
```

## Restricciones

- Los tres archivos nuevos van en `shared/` y **ninguno sabe qué es una
  carrera**. Si aparece la palabra `Race` o `Horse`, están en la carpeta
  equivocada.
- Los pipes reciben **el valor que transforman**, no el objeto que lo contiene.
- Los dos pipes son puros.
- Prohibido `any`. `standalone: true` en los tres.

## Autoevaluación

- [ ] `npm run build` pasa
- [ ] Las cuotas salen con coma y con dos decimales, todas
- [ ] `grep -r "toFixed" src/app/features/` no devuelve nada
- [ ] La palabra `pts` está escrita **una** vez en todo el proyecto
- [ ] La fila del favorito se ve distinta y anuncia que lo es
- [ ] `odds.pipe.ts` no menciona ni `Race` ni `Horse`
- [ ] No quedó ningún `TODO(S4)`

---

## Pistas

<details>
<summary>Pista 1 — por qué <code>toFixed</code> estaba mal</summary>

```js
(2.4).toFixed(2)   // '2.40'  — siempre punto, en cualquier idioma
```

`Intl.NumberFormat('es', { minimumFractionDigits: 2, maximumFractionDigits: 2 })`
devuelve `'2,40'`. Es la misma diferencia que hay entre escribir una fecha a mano
y usar el formateador del navegador.
</details>

<details>
<summary>Pista 2 — el separador de miles que no aparece</summary>

En español, el formato por defecto **no agrupa los números de cuatro cifras**:
1250 sale `1250`. Para un número suelto es lo correcto; para un importe, no.

Se resuelve con `useGrouping: true`, y ese es exactamente el tipo de detalle que
un pipe deja escrito una vez.
</details>

<details>
<summary>Pista 3 — dónde va cada archivo</summary>

```
shared/pipes/money.pipe.ts
shared/pipes/odds.pipe.ts
shared/pipes/index.ts
shared/directives/favourite.directive.ts
shared/directives/index.ts
```

`shared/` no importa de `features/` ni de `core/`. Es la regla de dependencias
del proyecto, y hoy vuelve a ser media consigna.
</details>

<details>
<summary>Solución completa</summary>

`correccion.md`, Parte B.
</details>

## Extensión

Agrega un pipe `relativeTime` que muestre cuánto falta para que largue una
carrera: `en 8 min`, `hace 2 h`, `largando`.

Y después contesta, por escrito, la pregunta difícil: **¿puede ser puro?**

Piénsalo así: el valor de entrada es la fecha de largada, y esa fecha no cambia
nunca. Pero lo que hay que mostrar sí cambia, porque pasa el tiempo.

La respuesta honesta es que un pipe no es la herramienta correcta para esto, y
entender por qué vale más que el pipe.

---

> **Referencia:** `conceptos.md` tiene todos los conceptos de la sesión con sus
> ejemplos.
