# S1 · Misión 1 — El mostrador

**Individual · 15 minutos · `lab/starter`, ruta `/s01`**

Es lo mismo que acabás de ver, pero lo escribís vos. Trabarse es parte del ejercicio: no hay ayuda en los primeros minutos, a propósito.

---

## Arrancar

```bash
cd lab/starter
npm start
```

Abrí <http://localhost:4200/s01>. Vas a ver el mostrador del **Café Compilado**, con el HTML escrito a mano y nada conectado.

Buscá `TODO(S1)` en `src/app/sesiones/s01/` — están numerados del 1 al 4, uno por cada tipo de binding.

## Qué construir

**1 · Interpolación.** El nombre, el origen y el precio están escritos a mano en el HTML. Cambialos por las propiedades de `cafe`.

> Comprobalo: cambiá el precio en `s01.component.ts` y recargá. Si el HTML sigue diciendo 42, todavía no está.

**2 · Property binding.** El `<div class="producto">` tiene que llevar además `producto--agotado` **solo** cuando el café no está disponible. Y el texto de estado tiene que decir «Disponible» o «Agotado».

**3 · Event binding.** El botón no hace nada. Que llame a `alternarDisponibilidad()`, y que su texto cambie según el estado.

**4 · Two-way binding.** Enlazá los dos inputs con `cliente` y `cantidad`. Después mostrá el `total` y deshabilitá el botón cuando `puedeAgregar` sea `false`.

## Listo cuando

Estos criterios los verificás vos, sin preguntar:

- [ ] Cambiar el precio en el `.ts` cambia lo que se ve
- [ ] «Marcar agotado» tacha el precio y pone el borde rojo
- [ ] Escribir un nombre y una cantidad actualiza el total **mientras escribís**
- [ ] El botón está deshabilitado hasta que hay nombre y cantidad
- [ ] «Agregar al pedido» suma una línea a la comanda
- [ ] La consola del navegador no tiene errores en rojo

## Si te trabás

<details>
<summary>Pista 1 — después de 5 minutos</summary>

`[(ngModel)]` no funciona sin más. El error dice:

```
Can't bind to 'ngModel' since it isn't a known property of 'input'.
```

Un componente standalone **declara lo que usa**. ¿Dónde se declara?
</details>

<details>
<summary>Pista 2 — después de 10 minutos</summary>

En el `@Component`, la propiedad `imports`. Va `FormsModule`, que viene de `@angular/forms`.

Y para el punto 2: `class="x"` pone la clase siempre. `[class.x]="condición"` la pone solo si la condición es verdadera. Podés usar las dos en el mismo elemento.
</details>

## Si terminaste antes

Agregá un **descuento por cantidad**: a partir de 5 unidades, un 10 % menos. Que se vea en el total y que aparezca un texto solo cuando el descuento aplica.

Pista: no hace falta ningún binding nuevo. Con lo que ya sabés alcanza.
