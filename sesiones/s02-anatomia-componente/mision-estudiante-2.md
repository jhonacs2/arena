# S2 · Ejercicio 2 — La tarjeta de una carrera

**En parejas · 20 minutos · proyecto `project/frontend/starter`**

---

## Enunciado

El listado de carreras dibuja las ocho tarjetas adentro de su propio `@for`. La
semana que viene hay que mostrar una carrera destacada en la portada, y en la
sesión 10 el mismo aspecto va a hacer falta en la pantalla de la carrera en vivo.
Con el marcado donde está hoy, eso son tres copias del mismo HTML.

Sacá la tarjeta a un componente reutilizable y la pastilla de estado a una
primitiva del sistema de diseño.

## Estado inicial

El punto de partida es **el listado que escribiste en S1**, en
`src/app/features/races/`. Si te quedó a medias, `sesiones/s01-primer-componente/correccion.md`
lo deja funcionando en diez minutos; arrancá desde ahí.

```bash
cd project/frontend/starter
npm start
```

## Datos

Los mismos de siempre: las ocho carreras de `core/mocks`. No se inventan datos y
no se cambia ningún nombre de campo.

---

## Requisitos

### 1. `<app-badge>` — la pastilla

Una primitiva del sistema de diseño, en **`src/app/shared/ui/badge/`**.

- El texto entra por **`<ng-content>`**, no por un `input()`.
- El tono entra por un `input()` con estos valores y ninguno más: `neutral`,
  `live`, `success`, `accent`. Por defecto, `neutral`.
- Se exporta desde `shared/ui/index.ts`.

**No sabe qué es una carrera.** Si en su archivo aparece la palabra `Race`, está
en la carpeta equivocada.

### 2. `<app-race-card>` — la tarjeta

En **`src/app/features/races/`**, porque sí sabe qué es una carrera.

| Entrada | Tipo | |
|---|---|---|
| `race` | `Race` | **Obligatoria** |
| `time` | `string` | **Obligatoria.** Ya formateada por el padre |
| `favourite` | `Horse \| undefined` | Opcional |
| `selected` | `boolean` | Opcional, por defecto `false` |

| Salida | Tipo | Cuándo |
|---|---|---|
| `toggled` | `Race` | Cuando alguien toca la tarjeta |

**La tarjeta no decide cuál está abierta.** Solo puede haber una abierta a la
vez, y eso únicamente lo sabe el listado.

### 3. Los dos huecos de la tarjeta

- Uno **con `select`**, arriba, donde el listado pone el `<app-badge>` con el
  texto y el tono del estado.
- Uno **sin `select`**, abajo, donde el listado pone —**solo en la carrera en
  vivo**— el texto `Se está corriendo ahora`.

El hueco de abajo no tiene que ocupar lugar en las siete tarjetas que no
proyectan nada.

### 4. El listado queda con lo suyo

Después del corte, `race-list` se queda con exactamente dos cosas: **preparar los
datos** y **decidir cuál está abierta**. El marcado de una carrera ya no está
ahí, y su CSS tampoco.

El panel de detalle, la parrilla y el simulador **no se tocan**.

---

## Resultado esperado

**La pantalla se ve exactamente igual que al terminar S1.** Ni un píxel de
diferencia — salvo el texto nuevo bajo la carrera en vivo.

```
┌────────────────────────────────────────┐
│ [TERMINADA]              29 jul, 22:38 │
│ Clásico Apertura                       │
│ 6 competidores · favorito Viento Norte │
└────────────────────────────────────────┘
┌────────────────────────────────────────┐
│ [EN VIVO]                30 jul, 14:23 │   ← la pastilla en rojo
│ Gran Premio Nacional                   │
│ 8 competidores · favorito Payador      │
│ Se está corriendo ahora                │   ← proyectado, solo acá
└────────────────────────────────────────┘
```

Lo que cambió no se ve: hay dos piezas que se pueden usar en otra pantalla.

## Restricciones

- `shared/` **no importa de `features/`**. Es la regla de dependencias del
  proyecto y hoy es la mitad del ejercicio.
- La tarjeta no importa `RACES` ni `favourite()`.
- El CSS de la tarjeta se muda con la tarjeta.
- `standalone: true` y `OnPush` en los dos componentes nuevos. Prohibido `any`.
- No se cambia el comportamiento: tocar la carrera abierta la sigue cerrando.

## Autoevaluación

- [ ] `npm run build` pasa
- [ ] La pantalla se ve igual que antes, y el detalle sigue funcionando
- [ ] Tocar una carrera la abre; tocarla de nuevo la cierra
- [ ] La pastilla de la carrera en vivo se ve en rojo y las otras no
- [ ] `Se está corriendo ahora` aparece **solo** en la carrera en vivo
- [ ] `badge.component.ts` no menciona la palabra `Race` ni una vez
- [ ] `race-card.component.ts` no importa nada de `core/mocks`
- [ ] En `race-list.component.css` no quedó ninguna regla `.race*`

---

## Pistas

<details>
<summary>Pista 1 — por qué el texto de la pastilla va por ng-content</summary>

Si el texto fuera un `input()`, la pastilla serviría para textos y nada más. El
día que alguien quiera meterle un icono al lado, o un número, hay que agregarle
un input nuevo — y después otro.

Con `<ng-content>` el padre mete lo que quiera y la pastilla no se entera. Se
usa así:

```html
<app-badge tone="live">En vivo</app-badge>
```
</details>

<details>
<summary>Pista 2 — dos huecos en el mismo componente</summary>

```html
<span class="race__status">
  <ng-content select="[card-status]" />
</span>
…
<span class="race__extra">
  <ng-content />
</span>
```

Y el padre marca cuál va a dónde con un atributo:

```html
<app-race-card …>
  <app-badge card-status [tone]="view.tone">{{ view.statusLabel }}</app-badge>
  …
</app-race-card>
```

El que **no** lleva `select` recibe todo lo demás, y va uno solo.
</details>

<details>
<summary>Pista 3 — el hueco vacío que ocupa lugar</summary>

Si las siete tarjetas que no proyectan nada abajo quedan con un renglón en
blanco, es porque el contenedor del hueco existe igual. Se arregla en el CSS del
hijo:

```css
.race__extra:empty {
  display: none;
}
```
</details>

<details>
<summary>Pista 4 — el tono de la pastilla</summary>

`tone` es una unión de literales, como las de S0. Y la traducción de estado a
tono es una decisión **del listado**, no de la tarjeta ni de la pastilla: va
donde ya está `STATUS_LABELS`, con la misma forma.
</details>

<details>
<summary>Solución completa</summary>

`correccion.md`, Parte B: los tres archivos terminados, con el porqué de cada
decisión.
</details>

## Extensión

Poné `<app-race-card>` una segunda vez, arriba del listado, mostrando solo la
carrera en vivo, con una pastilla `accent` que diga `DESTACADA` y sin nada
proyectado abajo.

Si te lleva más de dos minutos y hay que tocar el archivo de la tarjeta, el corte
quedó atado a la pantalla. Anotá qué te faltó: eso es material para el code
review.

---

> **Referencia:** `conceptos.md` tiene todos los conceptos de la sesión con sus
> ejemplos. Está pensado para tenerlo abierto al lado del editor.
