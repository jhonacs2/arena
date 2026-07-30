# S3 · Signals y control flow — guión

> **Esto es un teleprompter, no un resumen.** Lo que está entre comillas se dice. Lo que está en gris se hace. Leelo de corrido antes de dar la clase, con cronómetro.

| | |
|---|---|
| **Concepto único** | Un signal es un valor que **avisa** cuando cambia. Si modificás lo que había adentro, no avisa — y media pantalla se queda vieja. |
| **Al final saben** | Decidir qué es estado y qué es derivado · escribir `signal`, `computed`, `set` y `update` · actualizar sin mutar · usar `@if`, `@for` con `track` y `@switch` · explicar por qué `coffee()` de S2 se leía con paréntesis. |
| **Requisito previo** | S2. `input()` y `output()`, y la sensación de que los paréntesis de `coffee()` eran raros. |
| **Archivos** | `lab/starter/src/app/sessions/s03/` · `project/frontend/starter/src/app/features/races/` |

---

## Glosario de la sesión

| Palabra | En una frase |
|---|---|
| **Estado** | Lo que la aplicación guarda porque nadie lo puede deducir de otra cosa. |
| **Derivado** | Lo que sale de calcular sobre el estado. No se guarda: se calcula. |
| **Signal** | Un valor que avisa cuando cambia. Se lee llamándolo: `orders()`. |
| **`set`** | Reemplaza el valor entero, sin mirar el anterior. |
| **`update`** | Recibe lo que hay y devuelve lo que va a haber. |
| **`computed`** | Un valor derivado de otros signals, que se recalcula solo. |
| **Memoizado** | Que guarda su último resultado y no lo vuelve a calcular hasta que su fuente cambie. |
| **Mutar** | Modificar lo que ya existe: `push`, `sort`, asignar una propiedad. |
| **Inmutable** | No modificar nunca lo que había: se pone un valor nuevo. |
| **Control flow** | Las instrucciones del template: `@if`, `@for`, `@switch`. |
| **`track`** | Cómo Angular reconoce que una fila de la lista **es la misma** que antes. |

---

## 0:00 · Pregunta de apertura — 5 min

**En pantalla:** diapositiva 2.

> «La clase pasada escribimos `coffee()`, con paréntesis, y varios me miraron raro. Alguno lo escribió sin paréntesis y no compiló.»
>
> «Hoy quiero empezar por ahí. **¿Por qué les parece que un dato se lee llamándolo, como si fuera una función?** Tiren cualquier cosa en el chat.»

**Esperá 90 segundos.** Van a decir «porque es una función», «para que se actualice», «no sé». **Todas sirven.**

> «El que dijo *para que se actualice* está muy cerca. Un dato común no puede avisar cuando cambia: es un número, se queda ahí. Para avisar hay que ser algo más que un número — y eso es lo que vamos a ver hoy.»

---

## 0:05 · Wayground de S2 — 7 min

**Correr:** `sesiones/s02-anatomia-componente/wayground.csv`.

| Si falla | Decir |
|---|---|
| `output()` que nadie escucha | «Compila y no pasa nada. Hoy vamos a ver un primo suyo, todavía más silencioso.» |
| `ng-content` duplicado | «Se mueve, no se copia. Un nodo del DOM no puede estar en dos lugares.» |
| `ngOnChanges` con `model()` | «Da toda la vuelta: sube al padre y vuelve a bajar.» |

**No te enganches.** Máximo 30 segundos por pregunta.

---

## 0:12 · El concepto — 8 min

> **El editor está cerrado.**

### 0:12 — Estado y derivado · 2 min

**En pantalla:** diapositiva 5.

> «Antes de hablar de signals hay una pregunta más vieja: **de todo lo que hay en una pantalla, ¿qué hay que guardar?**»
>
> «El tablero de la comanda tiene: las comandas, el filtro elegido, el texto del buscador, cuántas hay pendientes, cuánto falta cobrar y cuáles son las más caras. **Seis cosas. Solo tres son estado.**»

| | |
|---|---|
| **Estado** | las comandas · el filtro · la búsqueda |
| **Derivado** | cuántas pendientes · cuánto falta cobrar · las más caras |

> «Las tres de abajo salen de las tres de arriba. Guardarlas es firmar un compromiso: cada vez que cambie una comanda hay que acordarse de actualizar las tres. Y un día te olvidás de una, y la pantalla miente.»

### 0:14 — Qué es un signal · 3 min

**En pantalla:** diapositiva 6 — `diagramas/el-signal-avisa.svg`, **la fila de arriba solamente**.

**Los términos que se definen acá:** *signal*, *`computed`*.

> «Un **signal** es un valor que **avisa** cuando cambia. Eso es todo, y es todo lo que hace falta.»
>
> «Se lee llamándolo —`orders()`— y ahí está el porqué de los paréntesis de la clase pasada: al leerlo, el signal **anota quién lo leyó**. Por eso después puede avisarle.»

Señalá la cadena del diagrama:

> «`orders` avisa. `visible`, que es un `computed`, lo escuchaba, así que se recalcula. Y la vista, que estaba leyendo `visible`, repinta.»
>
> «Un `computed` **no guarda nada**: se calcula. Y no se calcula todo el tiempo: guarda su último resultado hasta que su fuente cambie. A eso se le dice **memoizado**, y es lo que hace que se pueda poner en el template sin culpa.»

### 0:17 — Y lo que pasa cuando mutás · 3 min

**En pantalla:** la fila de abajo del mismo diagrama.

**El término que se define acá:** *mutar*.

> «Ahora la parte que va a arruinarle la tarde a alguno, y por eso la vemos hoy y no cuando pase.»
>
> «`orders().push(unaComandaNueva)`. El array cambia: adentro hay un elemento más. Pero **para el signal es el mismo array de siempre** — la misma dirección de memoria. No pasó nada. No avisa.»

Y ahora lo que casi nadie espera:

> «¿Y qué se ve? Acá está lo peor: **la mitad de la pantalla se actualiza igual**. El `@for` que lee `orders()` directo se relee cuando Angular revisa, y muestra el elemento nuevo. Pero el `computed` que sacaba el total **nunca se recalcula**, porque nadie le avisó.»
>
> «Te queda una pantalla con seis filas y un total que dice cinco. Sin un error. Sin un log. Y va a seguir así para siempre, aunque toques otros botones.»

**La regla, y dejala en pantalla:**

> **Nunca modifiques lo que había. Poné algo nuevo.**
> `update(v => [...v, x])`, nunca `v.push(x)`.

> **Si vas tarde:** de este bloque no se recorta nada. La fila de abajo del diagrama es la sesión.

---

## 0:20 · Live coding — 15 min

**En pantalla:** VS Code y el navegador lado a lado. Proyecto: **`lab/demo`**, ruta `/s03`.

> **Antes de entrar al aula:** `node scripts/prep-demo.mjs`, sumá la ruta `/s03` a mano, y `npm start`. La secuencia completa está en **`mision-profe.md`**.

### 0:20 — De propiedad a signal · 3 min

```ts
protected orders = signal<readonly Order[]>(INITIAL_ORDERS);
```

En el template, `orders` pasa a `orders()`.

> «Un cambio mecánico: paréntesis en todos lados. Y la pantalla sigue exactamente igual, lo cual es raro y vale decirlo: **hasta acá no ganamos nada**. Lo que ganamos aparece en el próximo paso.»

### 0:23 — El primer `computed` · 3 min

```ts
protected readonly pendingTotal = computed(() =>
  this.orders()
    .filter((order) => order.status !== 'served')
    .reduce((sum, order) => sum + lineTotal(order), 0),
);
```

> «Miren lo que **no** escribí: no lo guardé en ninguna propiedad, y no lo actualicé en `advance`, ni en `remove`, ni en `add`, ni en `reset`. **Cuatro lugares donde no hay que acordarse de nada.**»

Tocá los botones y mostrá el total siguiendo solo.

### 0:26 — La rotura · 4 min

Cambiá `add()` por la versión que muta:

```ts
protected add(): void {
  this.orders().push({ …unaComandaNueva });
}
```

Hmm — no compila: `readonly Order[]` no tiene `push`.

> «Miren esto: **el tipo de la sesión 0 no me deja**. `readonly Order[]` no tiene `push`. Voy a sacarle el `readonly` un segundo para poder mostrarles el bug, y ese gesto —sacar el `readonly` para que algo compile— es exactamente el momento en el que hay que parar.»

Sacá el `readonly`, guardá, y **tocá «Agregar comanda»**.

> 🔴 **Rotura deliberada.** Dejá que lo miren tres segundos antes de decir nada.

> «La fila apareció. El contador de arriba dice seis. Y el total de abajo… **no se movió**. Y no se va a mover más: toquen los otros botones, filtren, busquen. Ese número se quedó viejo para siempre.»
>
> «No hay error. No hay advertencia. Media pantalla dice una cosa y la otra media dice otra.»

Arreglalo:

```ts
this.orders.update((orders) => [...orders, unaComandaNueva]);
```

Y volvé a poner el `readonly`.

> «Y ahora entienden por qué está el `readonly` desde la primera clase: **no es una formalidad, es lo que hace que este bug no se pueda escribir.**»

### 0:30 — `@for` con `track` · 2 min

Sacá el `track` del `@for` a propósito:

```
NG5002: @for loop must have a "track" expression
```

> «Angular te obliga. Y no es capricho: `track` es cómo reconoce que la fila de Ana **sigue siendo la de Ana** cuando cambia el orden. Con `track order.id` la mueve; sin nada con qué reconocerla, la destruiría y la volvería a crear — perdiendo el foco, el scroll y cualquier cosa que el usuario estuviera haciendo ahí.»

### 0:32 — `@switch` y el mensaje de vacío · 3 min

```html
@if (visible().length === 0) {
  @switch (filter()) {
    @case ('pending') { No queda ninguna pendiente. La barra está al día. }
    @default { No hay comandas que coincidan con la búsqueda. }
  }
}
```

> «`@switch` es un `@if` con varias ramas, y sirve para lo mismo que el `switch` de TypeScript de la sesión 0: cuando el valor es una unión cerrada, cubrís los casos y listo.»
>
> «Y fijate qué gana el usuario: un “no hay nada” genérico no le dice si el filtro está mal o si de verdad no queda nada. Esto sí.»

---

## 0:35 · Misión 1 — 15 min

**Enunciado en `mision-estudiante-1.md`.**

> «Ahora ustedes, en `lab/starter`. La lista ya anda; lo que no hay es nada derivado: ni filtro, ni búsqueda, ni contadores, ni total. Quince minutos.»

**Decí esto antes de largar:**

> «El orden que menos duele es: primero pasás `orders` a signal y ponés los paréntesis hasta que compile. Recién ahí agregás lo derivado. Si arrancás por el filtro, vas a estar peleando dos cosas a la vez.»

**Reloj de pistas:**

| Min | Pista, sin resolver |
|---|---|
| 0:43 | «Si algo no compila y dice `is not a function`, te falta un paréntesis. Si dice `not assignable`, te sobra.» |
| 0:47 | «Un `computed` no lleva `set` ni `update`. Si te dan ganas de escribirle uno, probablemente sea estado y no derivado.» |

---

## 0:50 · Comparten pantalla — 10 min

Dos personas. **Preguntás, no corregís.**

1. «¿Cuáles de tus valores son estado y cuáles derivados? ¿Cómo lo decidiste?»
2. «¿Qué pasa si toco este botón dos veces seguidas?»
3. «¿Dónde actualizás el contador?» — *(la respuesta correcta es «en ningún lado»)*

**Lo más probable:** alguien guardó `visible` en un signal y lo actualiza a mano en cada cambio.

> «Anda perfecto. Contá cuántos lugares tenés que tocar para agregar un filtro nuevo. Ahora contá cuántos tendrías con un `computed`. Esa diferencia es toda la clase.»

---

## 1:00 · Descanso — 10 min

---

## 1:10 · Predice y ejecuta — 15 min

**Respuestas verificadas en el navegador:** `predice-y-ejecuta/respuestas.md`. **Dos de las tres no son lo que uno diría.**

| Min | Snippet | Casi todos predicen | Pasa |
|---|---|---|---|
| 1:10 | `push` sobre el array de un signal | «no se actualiza nada» | **Se actualiza la mitad.** La lista sí, el `computed` no |
| 1:15 | `@for` sin `track` | «anda, es opcional» | **No compila.** `NG5002` |
| 1:20 | `computed` que ordena su fuente con `.sort()` | «ordena una copia» | **Ordena el original**, y otro `computed` empieza a mentir |

Cerrá con:

> «Los tres son el mismo error con tres caras: **modificar algo que otro estaba mirando**. Y ninguno de los tres te avisa en tiempo de ejecución. El único que te frena es el segundo, y te frena el compilador.»

---

## 1:25 · Misión 2, en parejas — 20 min

**Enunciado en `mision-estudiante-2.md`.**

**Tres cosas antes de largar:**

> «Uno: el punto de partida es su listado de S2. Si les quedó a medias, la corrección de S2 lo deja andando.»
>
> «Dos: **las ocho carreras no son estado.** Vienen de una constante y no cambian nunca. Un signal para algo que no cambia es ruido. Estado son tres cosas: el filtro, la búsqueda y cuál está abierta.»
>
> «Tres: guarden **el id** de la carrera abierta, no la carrera. Van a ver por qué solos, en cuanto filtren.»

---

## 1:45 · Code review en vivo — 10 min

Rúbrica del curso, y hoy el punto 2 es el que manda:

1. ¿`standalone: true` y `OnPush`?
2. **¿Actualiza el estado sin mutar?** ← hoy es el importante
3. ¿`any`, `console.log`, imports sin usar?
4. ¿El nombre dice lo que la cosa hace?
5. ¿Está en la carpeta que le toca?

Y la pregunta de la sesión:

> «Contá los signals. Si hay más de tres o cuatro en una pantalla de este tamaño, casi seguro alguno es derivado y está guardado de más. Buscá el que se actualiza a mano en dos lugares: ese es.»

Y el cierre:

> «Y algo que va a parecer un detalle y no lo es: **el panel de detalle se cierra solo cuando filtrás y la carrera abierta queda afuera.** Nadie escribió eso. Sale de haber guardado el id en vez del objeto, y de derivar el resto. Eso es lo que se gana.»

---

## 1:55 · Exit ticket y tarea — 5 min

**Exit ticket:** `exit-ticket.md`. **Tarea:** `tarea.md`, leída en voz alta.

> «`conceptos.md` tiene todo lo de hoy con los ejemplos exactos y las tres respuestas del bloque de predicciones, que son las que no se pueden adivinar.»

**Y el aviso de la próxima:**

> «La clase que viene: directivas y pipes. Vamos a sacar de los templates todos esos `toFixed(2)` que venimos arrastrando desde la primera clase.»

---

## Después de la clase

- [ ] Leer los exit tickets. Lo confuso arranca S4.
- [ ] Revisar `wayground.csv` de **esta** sesión — se corre al empezar S4.
- [ ] Aplicar la corrección de S3 al `starter/` publicado y taggear `s04`.

### Notas de la corrida real

| | |
|---|---|
| ¿Cuántos guardaron lo derivado en un signal? | |
| ¿La rotura de las 0:26 se entendió sin explicarla? | |
| ¿Qué pregunta no supe contestar? | |
| ¿Qué sacaría o agregaría? | |
