# S1 · Ejercicio 2 — El programa de carreras

**En parejas · 20 minutos · proyecto `project/frontend/starter`**

---

## Enunciado

El hipódromo necesita su pantalla principal: el **programa de carreras**. Lista
las carreras de la jornada con su hora, su estado y cuántos competidores tiene
cada una. Al tocar una carrera se abre su parrilla —los caballos que corren— y al
tocarla de nuevo se cierra. Dentro de la parrilla, un simulador calcula cuánto
pagaría una apuesta al favorito.

Construí esa pantalla desde cero, usando los mismos cuatro bindings del ejercicio
anterior, ahora sobre datos reales del proyecto.

## Modalidad

**Conducción por turnos.** Diez minutos escribe uno mientras el otro dicta, y a
los diez se invierte. El que dicta no toca el teclado; el que escribe no decide.

## Estado inicial

```bash
cd project/frontend/starter
npm start
```

Abre en `/sistema`, la muestra del sistema de diseño: ahí están los colores, las
tipografías y las sedas de los caballos. **Es la referencia visual, no el
producto.**

El producto no tiene ninguna pantalla todavía. Esa es la de hoy.

## Datos

Ya están en el proyecto. **No inventes datos.**

| Qué | Dónde |
|---|---|
| Las 8 carreras con sus 54 caballos | `core/mocks` |
| Los tipos `Race` y `Horse` | `core/models` |
| `favourite(race)` — devuelve el caballo favorito | `core/models` |
| `potentialPayout(amount, odds)` — calcula el pago | `core/models` |
| `<app-silk>` — dibuja la casaca de un caballo | `shared/ui` |
| El CSS de esta pantalla | `solution/` — copialo tal cual |

---

## Requisitos

### 1. La pantalla existe y es alcanzable

Creá el componente con el CLI, publicalo en la ruta `/carreras` y agregá el enlace
en el encabezado. Buscá `TODO(S1)`.

`/carreras` tiene que ser la pantalla principal de la aplicación.

### 2. Los datos preparados en la clase

Por cada carrera, calculá **en la clase**: la hora formateada, el caballo favorito
y la etiqueta de estado en castellano.

**Ni un cálculo en el template.**

### 3. Interpolación

Mostrá, por carrera: el estado, el nombre, la hora y la cantidad de competidores.

### 4. Property binding

El botón de cada carrera lleva estas clases y este atributo:

| Se aplica | Cuando |
|---|---|
| `race--live` | el estado es `live` |
| `race--finished` | el estado es `finished` |
| `race--open` | es la carrera seleccionada |
| `[attr.aria-pressed]` | con la misma condición que `race--open` |

### 5. Event binding

Tocar una carrera abre su parrilla. Tocar **la misma** la cierra.

Solo puede haber una parrilla abierta a la vez.

### 6. Two-way binding

En la parrilla, un campo con el monto de la apuesta. Debajo, cuánto pagaría esa
apuesta al favorito, actualizado **mientras se escribe**.

Usá `potentialPayout(amount, odds)`; no rehagas la cuenta.

---

## Resultado esperado

```
PROGRAMA DE CARRERAS

┌────────────────────────────────────────────────┐
│ EN VIVO   Clásico del Recuerdo   14:30   8 ▸  │  ← race--live
├────────────────────────────────────────────────┤
│ CERRADA   Handicap de Otoño      13:00   6    │  ← race--finished
├────────────────────────────────────────────────┤
│ ABIERTA   Premio Talento         15:15   7 ▾  │  ← race--open, abierta
│  ┌──────────────────────────────────────────┐  │
│  │ [seda] 1  Viento Norte      3.40         │  │
│  │ [seda] 2  Luz de Enero      5.10         │  │
│  │ …                                        │  │
│  │                                          │  │
│  │ Favorito: Viento Norte (3.40)            │  │
│  │ Monto  [ 500        ]                    │  │
│  │ Pagaría: $1.700                          │  │
│  └──────────────────────────────────────────┘  │
└────────────────────────────────────────────────┘
```

Una carrera en vivo tiene que distinguirse de una terminada **sin leer el texto**.

## Restricciones

- Todos los cálculos en la clase.
- Un solo estado abierto a la vez.
- No modifiques los datos de `core/mocks`.
- El `@for` que recorre la lista **está escrito en `correccion.md`**: copialo.
  Recorrer listas es la sesión 3; hoy el trabajo son los bindings de adentro.

## Autoevaluación

- [ ] `/carreras` abre y es la pantalla principal
- [ ] Las ocho carreras muestran nombre, hora, estado y cantidad de competidores
- [ ] La carrera en vivo se distingue de las terminadas sin leer el texto
- [ ] Tocar una carrera abre su parrilla; tocar la misma la cierra
- [ ] Cambiar el monto actualiza el pago mientras se escribe
- [ ] `npm run build` pasa sin errores
- [ ] Con `Tab` se llega a todas las carreras y se ve dónde está el foco

---

## Una observación intencional

El orden de la lista no ayuda: las carreras terminadas aparecen mezcladas con las
que faltan. **Está bien que moleste.** Ordenar y filtrar es la sesión 3; hoy los
datos se pintan como vienen.

## Pistas

<details>
<summary>Si la parrilla no se cierra al volver a tocar</summary>

Guardá el **id** de la carrera abierta, no un booleano. Al tocar una carrera,
compará: si el id que llega es el que ya está guardado, guardá `null`.
</details>

<details>
<summary>Solución completa</summary>

`correccion.md`, Parte B. Para destrabarse, no para copiar.
</details>

## Extensión

Mostrá en cada carrera **cuánto falta para que largue**: «en 8 min», «hace 2 h»,
«largando».

Todo el cálculo va en la clase. Punto de partida:

```ts
new Date(race.startsAt).getTime() - Date.now()
```

---

> **Referencia:** `conceptos.md` tiene todos los conceptos de la sesión con sus
> ejemplos.
