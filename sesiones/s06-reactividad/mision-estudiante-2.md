# S6 · Ejercicio 2 — El buscador que no molesta al servidor

**En parejas · 20 minutos · proyecto `project/frontend/starter`**

---

## Enunciado

El buscador de carreras filtra en cada tecla. Hoy eso no se nota, porque filtra
una constante que ya está en memoria.

En la sesión 7 esa búsqueda va a ser una petición al servidor. Escribir
`gran premio` van a ser **once peticiones**, y la undécima puede contestar antes
que la décima.

Ponle el freno ahora, mientras es barato comprobarlo.

## Estado inicial

El punto de partida es **el listado que terminaste en S5**, con `RaceStore` y
`BetStore` funcionando.

---

## Requisitos

### 1. El campo emite eventos, no valores

`[(ngModel)]` se va. Lo que hace falta es el **evento** de cada pulsación, para
empujarlo a un flujo:

```html
<input [value]="draft()" (input)="onSearch($any($event.target).value)" />
```

### 2. Dos textos distintos

Aquí está la parte que hay que pensar:

| Qué | Quién lo tiene | Cuándo cambia |
|---|---|---|
| Lo que se ve en el campo | el componente | en cada tecla |
| Lo que filtra la lista | `RaceStore` | cuando la persona deja de escribir |

**Dejaron de ser el mismo texto.** Si el campo leyera del store, se vaciaría
mientras se escribe.

### 3. El flujo

```ts
this.typed.pipe(debounceTime(250), distinctUntilChanged(), takeUntilDestroyed())
```

- `debounceTime(250)`
- `distinctUntilChanged()`
- `takeUntilDestroyed()`

La suscripción va en el **constructor**, que es un contexto de inyección.

### 4. «Ver todas» sigue funcionando

El botón que limpia la búsqueda tiene que limpiar **las dos** cosas: lo que se ve
en el campo y lo que filtra el store.

Si solo limpia el store, el campo queda con texto y la lista muestra todo. Si
solo limpia el campo, la lista sigue filtrada.

---

## Resultado esperado

**Se ve exactamente igual que al terminar S5.** Escribiendo `payador` letra por
letra, la lista ahora se actualiza **una vez**, al final, en vez de siete.

La forma de comprobarlo sin un servidor: escribe muy rápido y mira si la lista
parpadea en cada tecla. No debe.

## Restricciones

- `RaceStore` **no cambia**. El debounce es de la pantalla, no del store.
- Prohibido guardar la suscripción en una propiedad y cancelarla en
  `ngOnDestroy`: hay una línea para eso.
- Prohibido `any` fuera del `$any($event.target)` del template, que es la forma
  estándar de tipar el objetivo de un evento del DOM.
- `OnPush`.

## Autoevaluación

- [ ] `npm run build` pasa
- [ ] Escribir rápido no hace parpadear la lista en cada tecla
- [ ] El texto se ve en el campo mientras se escribe
- [ ] «Ver todas» limpia el campo **y** la lista
- [ ] `race.store.ts` no cambió
- [ ] La suscripción se corta sola
- [ ] No quedó ningún `TODO(S6)`

---

## Pistas

<details>
<summary>Pista 1 — por qué dos textos y no uno</summary>

Prueba con uno solo: haz que el `[value]` lea `races.query()`.

Vas a escribir una letra y ver cómo desaparece, porque el store todavía tiene el
texto viejo. El campo necesita el valor inmediato; la lista, el retrasado.
</details>

<details>
<summary>Pista 2 — dónde va la suscripción</summary>

```ts
constructor() {
  this.typed
    .pipe(debounceTime(250), distinctUntilChanged(), takeUntilDestroyed())
    .subscribe((value) => this.races.setQuery(value));
}
```

En el constructor o en un campo. Adentro de un método sale `NG0203`, igual que
`inject()` en S5, y por el mismo motivo.
</details>

<details>
<summary>Pista 3 — <code>$any</code> en el template</summary>

```html
(input)="onSearch($any($event.target).value)"
```

`$event.target` es `EventTarget`, que no tiene `value`. `$any` es la única forma
razonable de resolverlo en un template, y no cuenta como el `any` prohibido: es
sintaxis de Angular, no de TypeScript.
</details>

<details>
<summary>Solución completa</summary>

`correccion.md`, Parte B.
</details>

## Extensión

Escribe en un comentario, al final de `race-list.component.ts`, la respuesta a
esto:

> En la sesión 7 la búsqueda va a ser una petición al servidor. **¿Qué líneas de
> este archivo van a cambiar ese día?**

La respuesta esperada es incómoda de creer y es correcta: **ninguna**. El
componente empuja textos a un flujo y el store le contesta; de dónde saca el
store las carreras no le importa.

---

> **Referencia:** `conceptos.md` tiene todos los conceptos de la sesión con sus
> ejemplos.
