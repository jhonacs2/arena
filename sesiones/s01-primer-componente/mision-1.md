# S1 · Misión 1 — El mostrador

**Individual · 15 minutos · `lab/starter`**

Lo mismo que acabás de ver, pero lo escribís vos. **Desde cero: el componente no existe.**

Trabarse es parte del ejercicio. En los primeros minutos no hay ayuda, a propósito.

---

## Arrancar

```bash
cd lab/starter
npm start
```

Abrí <http://localhost:4200>. Vas a ver la pantalla de inicio y **la barra de la izquierda vacía**. Está bien: no hay ninguna sesión todavía. La vas a hacer aparecer vos.

## Qué construir

**1 · Que la pantalla exista.**

Creá el componente con el CLI y hacelo alcanzable en `/s01`. Acordate: son **dos archivos** —la ruta y el índice—, y hay que tocar los dos. Buscá `TODO(S1)`.

> Antes de seguir, comprobalo: la dirección `/s01` tiene que abrir algo, y «Primer componente» tiene que aparecer en la barra.

**2 · Interpolación.** Poné en la clase un objeto `cafe` con nombre, origen, precio y disponible. Mostralos.

> Comprobalo: cambiá el precio en el `.ts`. Si la pantalla no cambia sola, todavía no está.

**3 · Property binding.** Que el contenedor lleve la clase `producto--agotado` **solo** cuando no hay stock, y que el texto diga «Disponible» o «Agotado».

**4 · Event binding.** Un botón que alterne la disponibilidad, con el texto cambiando según el estado.

**5 · Two-way binding.** Dos inputs, para el nombre del cliente y la cantidad. Mostrá el total y deshabilitá el botón de agregar cuando falte algo.

> El CSS ya está en `lab/solution` si querés copiarlo, pero **no es parte del ejercicio**. Que se vea feo hoy no importa.

## Listo cuando

Esto lo verificás vos, mirando la pantalla:

- [ ] `/s01` abre, y «Primer componente» aparece en la barra de la izquierda
- [ ] Cambiar el precio en el `.ts` cambia lo que se ve
- [ ] «Marcar agotado» cambia el aspecto **y** el texto del botón
- [ ] Escribir en los inputs actualiza el total **mientras escribís**
- [ ] El botón está deshabilitado hasta que hay nombre y cantidad
- [ ] La consola del navegador no tiene errores en rojo

## Si te trabás

<details>
<summary>Pista 1 — a los 5 minutos</summary>

Si te aparece este error, es el esperado y es el más importante de la clase:

```
Can't bind to 'ngModel' since it isn't a known property of 'input'.
```

Leelo entero. Dice que `ngModel` **no es una propiedad conocida**. Un componente standalone declara lo que usa. ¿Dónde lo declara?
</details>

<details>
<summary>Pista 2 — a los 10 minutos</summary>

En el `@Component`, la propiedad `imports`. Va `FormsModule`, de `@angular/forms`.

Y para el punto 3: `class="x"` pone la clase **siempre**. `[class.x]="condición"` la pone solo si la condición da verdadero. Podés usar las dos juntas en el mismo elemento.
</details>

<details>
<summary>Si te quedaste muy atrás</summary>

`correccion.md`, Parte A. Tiene el paso a paso completo. Usalo para destrabarte y seguí desde ahí — no lo copies entero.
</details>

## Si terminaste antes

Agregá un **descuento por cantidad**: a partir de 5 unidades, un 10 % menos. Que se vea en el total y que aparezca un aviso solo cuando el descuento aplica.

No hace falta ningún binding nuevo. Con lo de hoy alcanza.
