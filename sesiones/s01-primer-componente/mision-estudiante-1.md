# S1 · Ejercicio 1 — El mostrador de la cafetería

**Individual · 15 minutos · proyecto `lab/starter`**

---

## Enunciado

Una cafetería necesita la pantalla de su mostrador. La pantalla muestra **un**
café: su nombre, su origen y su precio. Indica si hay stock y permite marcarlo
como agotado. Además arma un pedido: quién lo lleva y cuántas unidades, con el
total calculado.

Construí esa pantalla desde cero, usando los cuatro tipos de binding.

## Estado inicial

```bash
cd lab/starter
npm start
```

En <http://localhost:4200> vas a ver la pantalla de inicio y **la barra de la
izquierda vacía**. Eso es correcto: todavía no hay ninguna sesión registrada.

**El componente de esta sesión no existe.** Los archivos que hay que tocar están
marcados con `TODO(S1)`.

## Datos

Usá exactamente este objeto:

```ts
{ name: 'Yirgacheffe', origin: 'Etiopía', price: 42, available: true }
```

El pedido arranca vacío: sin nombre de cliente y con cantidad 1.

---

## Requisitos

### 1. La pantalla existe y es alcanzable

Creá el componente con el CLI de Angular y hacelo accesible en la ruta `/s01`.

Hacen falta **dos** registros, en dos archivos distintos: la ruta y el índice de
sesiones. Buscá `TODO(S1)`.

**Verificación:** la dirección `/s01` abre el componente, y «Primer componente»
aparece en la barra de la izquierda.

### 2. Interpolación

Declará el café en la clase y mostrá en el template su nombre, su origen y su
precio.

**Verificación:** cambiá `price` en el `.ts` y guardá. Lo que se ve cambia sin que
toques el HTML.

### 3. Property binding

El contenedor del producto lleva la clase CSS `product--soldout` **solamente
cuando no hay stock**. Junto a él, un texto que dice `Disponible` o `Agotado`
según corresponda.

La clase `product`, que va siempre, tiene que seguir estando.

### 4. Event binding

Un botón que invierte la disponibilidad cada vez que se lo toca. Su texto cambia
según el estado: `Marcar agotado` cuando hay stock, `Marcar disponible` cuando no.

Al modificar el café, **construí un objeto nuevo** en lugar de reasignar sus
propiedades:

```ts
this.coffee = { ...this.coffee, available: !this.coffee.available };
```

### 5. Two-way binding

Dos campos de entrada: el nombre del cliente (texto) y la cantidad (número).

- El total es `precio × cantidad`, y se actualiza **mientras se escribe**.
- El botón `Agregar al pedido` está deshabilitado mientras falte el nombre del
  cliente o la cantidad sea menor que 1.

---

## Resultado esperado

Con stock, cliente «Ana» y cantidad 3:

```
┌──────────────────────────────────────┐
│  Yirgacheffe                         │
│  Etiopía · $42                       │
│  Disponible                          │
│                                      │
│  [ Marcar agotado ]                  │
│                                      │
│  Cliente   [ Ana          ]          │
│  Cantidad  [ 3            ]          │
│                                      │
│  Total: $126                         │
│  [ Agregar al pedido ]  ← habilitado │
└──────────────────────────────────────┘
```

Sin stock: el contenedor lleva `product--soldout`, el texto dice `Agotado` y el
botón dice `Marcar disponible`.

Con el cliente vacío: `Agregar al pedido` está deshabilitado.

## Restricciones

- Todos los datos y todos los cálculos van **en la clase**, no en el template.
- No uses `document.querySelector` ni `addEventListener`. Todo pasa por bindings.
- No hace falta CSS. Si querés, el de `lab/solution` se copia tal cual — **no es
  parte del ejercicio** y que se vea feo hoy no importa.

## Autoevaluación

- [ ] `/s01` abre, y «Primer componente» aparece en la barra de la izquierda
- [ ] Cambiar el precio en el `.ts` cambia lo que se ve
- [ ] El botón de disponibilidad cambia **el aspecto** y **su propio texto**
- [ ] Escribir en los campos actualiza el total mientras escribís
- [ ] `Agregar al pedido` está deshabilitado hasta que hay cliente y cantidad ≥ 1
- [ ] La consola del navegador no tiene errores en rojo

---

## Pistas

<details>
<summary>Pista 1 — el error de <code>ngModel</code></summary>

Si te aparece esto, es el error esperado y el más importante del ejercicio:

```
Can't bind to 'ngModel' since it isn't a known property of 'input'.
```

Leelo entero. Dice que `ngModel` **no es una propiedad conocida**. Un componente
standalone declara lo que usa. ¿Dónde lo declara?

Está explicado en `conceptos.md` §9.
</details>

<details>
<summary>Pista 2 — <code>imports</code>, y las dos clases juntas</summary>

En el `@Component`, la propiedad `imports`. Va `FormsModule`, de `@angular/forms`.

Y para el requisito 3: `class="x"` pone la clase **siempre**;
`[class.x]="condición"` la pone solo si la condición es verdadera. **Las dos
formas se pueden usar en el mismo elemento** y no se pisan.
</details>

<details>
<summary>Solución completa</summary>

`correccion.md`, Parte A: el paso a paso de vacío a funcionando, con el por qué de
cada decisión. Usalo para destrabarte y seguí desde ahí — no lo copies entero.
</details>

## Extensión

Agregá un **descuento por cantidad**: a partir de 5 unidades, un 10 % menos. Que
se refleje en el total y que aparezca un aviso solo cuando el descuento aplica.

No hace falta ningún binding nuevo: con los cuatro de hoy alcanza.

---

> **Referencia:** `conceptos.md` tiene todos los conceptos de la sesión con sus
> ejemplos. Está pensado para tenerlo abierto al lado del editor.
