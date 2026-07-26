# S1 · Misión 2 — El programa de carreras

**En parejas · 20 minutos · `project/frontend/starter`**

El concepto de hoy, ahora en el hipódromo. **La pantalla no existe: la crean ustedes.**

**Conducción por turnos:** 10 minutos escribe uno mientras el otro dicta, y a los 10 se invierte. El que dicta no toca el teclado; el que escribe no decide.

---

## Arrancar

```bash
cd project/frontend/starter
npm start
```

Abrí <http://localhost:4200>. Va a abrir en `/sistema`, la muestra del sistema de diseño. **Esa pantalla es su referencia**, no el producto: ahí están los colores, las tipografías y las sedas de los caballos.

El producto todavía no tiene ninguna pantalla. Esa es la de hoy.

## Lo que ya está

- **Los datos.** Las 8 carreras con sus 54 caballos, en `core/mocks`. **No inventen datos.**
- **Los tipos.** `Race`, `Horse`, y dos ayudantes: `favourite(race)` y `potentialPayout(amount, odds)`, en `core/models`.
- **El diseño.** Colores, tipografías y `<app-silk>`, que dibuja la casaca de cada caballo.
- **El CSS de esta pantalla.** Está en `solution/`: cópienlo tal cual. Hoy no se pelean con estilos.

## Qué construir

**1 · Que la pantalla exista.** Componente con el CLI, ruta `/carreras`, y el enlace en el encabezado. Buscá `TODO(S1)`.

**2 · Los datos, preparados en la clase.** Por cada carrera: la hora formateada, el favorito y la etiqueta de estado. **En la clase, no en el HTML.**

**3 · Interpolación.** Estado, nombre, hora y cuántos competidores tiene cada carrera.

**4 · Property binding.** Las clases de estado del botón de cada carrera:

- `race--live` cuando el estado es `live`
- `race--finished` cuando es `finished`
- `race--open` cuando **es** la seleccionada
- `[attr.aria-pressed]` con la misma condición que la anterior

**5 · Event binding.** Tocar una carrera abre su parrilla. Tocar la misma la cierra.

**6 · Two-way binding.** El monto del simulador, y el pago que le correspondería al favorito.

> El `@for` que recorre la lista está en `correccion.md`, escrito. Recorrer listas es la clase 3 — **hoy el trabajo son los bindings de adentro**.

## Listo cuando

- [ ] `/carreras` abre y es la pantalla principal
- [ ] Las ocho carreras muestran nombre, hora, estado y cantidad de competidores
- [ ] La que está en vivo se distingue de las terminadas **sin leer el texto**
- [ ] Tocar una carrera abre su parrilla; tocar la misma la cierra
- [ ] Cambiar el monto actualiza el pago **mientras escribís**
- [ ] `npm run build` pasa sin errores
- [ ] Con `Tab` se llega a todas las carreras y se ve dónde está el foco

## Una cosa que van a notar

El orden de la lista no ayuda: las terminadas aparecen mezcladas con las que faltan. **Está bien que les moleste.** Ordenar y filtrar es la clase 3. Hoy se pintan los datos como vienen.

## Si se trabaron

`correccion.md`, Parte B. Para destrabarse, no para copiar.

## Si terminan antes

Mostrá en cada carrera **cuánto falta para que largue**: «en 8 min», «hace 2 h», «largando». Todo el cálculo va en la clase.

Pista: `new Date(race.startsAt).getTime() - Date.now()`.
