# S1 · Misión 2 — El programa de carreras

**En parejas · 20 minutos · `project/frontend/starter`**

El concepto de hoy, ahora en el hipódromo. Es el único bloque de la sesión que toca el proyecto ancla.

**Conducción por turnos:** 10 minutos escribe uno mientras el otro dicta, y a los 10 se invierte. El que dicta no toca el teclado; el que escribe no decide.

---

## Arrancar

```bash
cd project/frontend/starter
npm start
```

Abrí <http://localhost:4200>. La pantalla **ya funciona a medias**: las ocho carreras están, pero les falta todo el texto y el panel de la derecha se comporta raro.

No arrancás de una hoja en blanco. Arrancás de algo que anda mal, que es más parecido al trabajo real.

Archivos:

- `src/app/features/races/race-list.component.ts`
- `src/app/features/races/race-list.component.html`

Buscá `TODO(S1)`. **El CSS ya está: no lo toques.**

> El `@for` viene hecho. El control flow (`@if`, `@for`, `@switch`) se ve a fondo en S3; hoy el trabajo son los bindings de adentro.

## Qué construir

**En el HTML**

1. **Interpolación** — la etiqueta de estado, el nombre de la carrera, la hora y cuántos competidores tiene. Los datos ya están preparados en `carreras`.
2. **Property binding** — las clases de estado del botón:
   - `carrera--viva` cuando el estado es `live`
   - `carrera--terminada` cuando es `finished`
   - `carrera--abierta` cuando **es** la carrera seleccionada
   - `[attr.aria-pressed]` con la misma condición que `carrera--abierta`
3. **Event binding** — que `(click)` llame a `seleccionar(vista)`.
4. **Two-way binding** — el input del monto, enlazado con `monto`. Y mostrar `pagoPotencial`.

**En el TypeScript**

5. `pagoPotencial` siempre devuelve 0 porque la cuota está clavada en `0`. Poné la del favorito de la carrera seleccionada. Ojo: puede no haber ninguna seleccionada.
6. `seleccionar()` selecciona siempre. Hacé que tocar **la misma** carrera dos veces la deseleccione.

## Listo cuando

- [ ] Las ocho carreras muestran nombre, hora, estado y cantidad de competidores
- [ ] La que está en vivo se distingue de las terminadas sin leer el texto
- [ ] Tocar una carrera abre su parrilla; tocar la misma la cierra
- [ ] Cambiar el monto actualiza el pago **mientras escribís**
- [ ] `npm run build` pasa sin errores
- [ ] Recorrer con `Tab` llega a todas las carreras y se ve dónde está el foco

## Una cosa que vas a notar

El orden de la lista no ayuda: las terminadas aparecen mezcladas con las que faltan. **Está bien que te moleste.** Ordenar y filtrar es S3 — hoy solo se pintan los datos como vienen.

## Si terminan antes

Mostrá en cada carrera **cuánto falta para que largue** («en 8 min», «hace 2 h»). Todo el cálculo va en la clase, no en el template.

Pista: `new Date(carrera.startsAt).getTime() - Date.now()`.

## Se revisa en el bloque de las 1:45

Con la rúbrica de siempre. Si quieren ofrecer su solución, avisen mientras circulo.
