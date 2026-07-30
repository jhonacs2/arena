# S3 · Ejercicio 2 — Filtrar el programa

**En parejas · 20 minutos · proyecto `project/frontend/starter`**

---

## Enunciado

El listado muestra las ocho carreras del programa, todas juntas. En un hipódromo
de verdad son cuarenta, y lo primero que hace cualquiera es buscar la suya o ver
solo las que están por largar.

Agregá filtro por estado y búsqueda, con el estado en signals y todo lo demás
derivado.

## Estado inicial

El punto de partida es **el listado que terminaste en S2**, con
`<app-race-card>` y `<app-badge>` ya funcionando. Si te quedó a medias,
`sesiones/s02-anatomia-componente/correccion.md` lo deja andando.

```bash
cd project/frontend/starter
npm start
```

## Datos

Las ocho carreras de `core/mocks`. Los cuatro filtros son `Todas`, `En vivo`,
`Por largar` y `Terminadas`, y sus contadores salen del programa entero.

---

## Requisitos

### 1. Solo tres cosas son estado

Poné en signals **el filtro**, **la búsqueda** y **cuál carrera está abierta**.
Nada más.

**Las ocho carreras no son estado.** Vienen de una constante y no cambian nunca;
un signal para algo que no cambia es ruido. En S7, cuando las traiga el servidor,
sí lo va a ser.

### 2. Guardá el id, no la carrera

El signal de la carrera abierta guarda **su id**. La carrera se deriva de ahí con
un `computed`.

### 3. El filtro y la búsqueda

- Cuatro pestañas con su contador, sacado del programa entero.
- Un buscador que encuentre por **nombre de carrera o nombre de caballo**, sin
  distinguir mayúsculas.

### 4. Lo derivado

| Qué | De dónde sale |
|---|---|
| Las carreras que se ven | el programa, filtrado por estado y por texto |
| Los contadores | el programa entero |
| La carrera abierta | el id, buscado entre las que se ven |
| El pago posible | el monto y la cuota del favorito |
| La parrilla ordenada | los caballos de la carrera abierta, por cuota |

### 5. El control flow

- `@for` con **`track view.race.id`**.
- Cuando no hay resultados, un `@switch` sobre el filtro elige el mensaje, y
  además hay un botón «Ver todas» que limpia filtro y búsqueda.

### 6. La parrilla se ordena sin tocar el dataset

El panel de detalle muestra los caballos **de menor a mayor cuota**, y el empate
se rompe por número de partida.

`core/mocks` **no cambia de orden**. Si al abrir una carrera el dataset queda
reordenado, está mal.

---

## Resultado esperado

```
[ Todas 8 ] [ En vivo 1 ] [ Por largar 3 ] [ Terminadas 4 ]

Buscar por carrera o caballo  [ payador            ]

┌────────────────────────────────────────┐
│ [EN VIVO]                30 jul, 14:23 │
│ Gran Premio Nacional                   │
│ 8 competidores · favorito Payador      │
│ Se está corriendo ahora                │
└────────────────────────────────────────┘
```

**Y la prueba que hay que hacer sí o sí:** abrí una carrera terminada, y sin
cerrarla tocá el filtro «En vivo».

**El panel de detalle tiene que cerrarse solo.** Nadie escribe eso: sale de haber
guardado el id y derivado el resto.

## Restricciones

- Prohibido `push`, `sort`, `splice` sobre el estado o sobre los mocks.
- Prohibido guardar en un signal algo que se pueda derivar.
- Prohibido `any`. `OnPush` en todo.
- No se toca `<app-race-card>` ni `<app-badge>`: lo de S2 sigue igual.

## Autoevaluación

- [ ] `npm run build` pasa
- [ ] Hay **tres** signals de estado, y las carreras no son uno de ellos
- [ ] Los contadores no cambian al filtrar
- [ ] Buscar `payador` deja una sola carrera
- [ ] Abrir una carrera y filtrarla afuera **cierra el panel solo**
- [ ] Abrir una carrera no reordena la lista
- [ ] No quedó ningún `TODO(S3)`

---

## Pistas

<details>
<summary>Pista 1 — por qué el id y no la carrera</summary>

Probá las dos. Guardá la carrera entera, abrí una terminada y filtrá por «En
vivo»: el panel sigue mostrando una carrera que ya no está en la lista.

Para arreglarlo con el objeto guardado habría que acordarse de limpiarlo en cada
cambio de filtro **y** en cada cambio de búsqueda. Con el id, el `computed` que
la busca entre las visibles no la encuentra y devuelve `undefined`. Sale gratis.
</details>

<details>
<summary>Pista 2 — buscar también por caballo</summary>

```ts
private matches(view: RaceView, text: string): boolean {
  if (view.race.name.toLowerCase().includes(text)) return true;
  return view.race.horses.some((horse) => horse.name.toLowerCase().includes(text));
}
```
</details>

<details>
<summary>Pista 3 — los contadores no son de lo que se ve</summary>

Si el contador de «En vivo» dice 0 cuando estás filtrando por «Terminadas», lo
estás calculando sobre la lista filtrada. Tiene que salir del programa entero.
</details>

<details>
<summary>Solución completa</summary>

`correccion.md`, Parte B.
</details>

## Extensión

Agregá un orden configurable al listado: por hora de largada o por cuota del
favorito, con dos botones.

Va como **un signal más** —el criterio— y **un `computed` más** —la lista
ordenada—. Si te sale agregando un tercer signal con la lista ya ordenada,
volvé a leer el requisito 4.

---

> **Referencia:** `conceptos.md` tiene todos los conceptos de la sesión con sus
> ejemplos.
