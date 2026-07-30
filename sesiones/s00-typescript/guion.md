# S0 · TypeScript — guión

> **Esto es un teleprompter, no un resumen.** Lo que está entre comillas se dice. Lo que está en gris se hace. Leelo de corrido antes de dar la clase, con cronómetro: si tenés que parar a pensar qué sigue, avisame y lo arreglo.

| | |
|---|---|
| **Concepto único** | Un tipo es el conjunto de valores que algo puede tener, y alguien lo revisa antes de que el código corra. |
| **Al final saben** | Anotar un tipo y saber cuándo no hace falta · escribir una `interface` y una unión de literales · hacerse cargo de `undefined` · leer un error del compilador · decir por qué `readonly` está en todo el curso. |
| **Requisito previo** | Ninguno de Angular. JavaScript básico y el entorno instalado — `README.md` de esta carpeta. |
| **Archivos** | `lab/starter/src/app/sessions/s00/menu.ts` · `project/frontend/starter/src/app/core/models/race.model.ts` |

> **Es la primera clase del módulo.** No hay Wayground de la sesión anterior porque no hay sesión anterior: el bloque de las 0:05 es un diagnóstico en vivo. De S1 en adelante, el quiz vuelve a ser lo de siempre.

---

## Glosario de la sesión

Todo lo que se nombra hoy, en el orden en que aparece. **Ninguna de estas palabras se usa antes de definirla.** Si te escuchás diciendo una que no explicaste, parás y la explicás.

| Palabra | En una frase |
|---|---|
| **Compilador** | El programa que lee tu TypeScript, lo revisa y lo convierte en JavaScript. |
| **Tipo** | El conjunto de valores que una cosa puede tener. |
| **Anotación** | Escribir el tipo a mano, después de dos puntos: `price: number`. |
| **Inferencia** | Cuando TypeScript deduce el tipo solo y no hace falta escribirlo. |
| **`interface`** | Un nombre para la forma de un objeto. |
| **Alias de tipo** | Un nombre para cualquier tipo, no solo para objetos. Se escribe con `type`. |
| **Literal** | Un tipo de un solo valor: `'grande'` es un tipo con una sola cadena adentro. |
| **Unión** | Un tipo que es «esto o esto otro»: `A \| B`. |
| **Narrowing** | Cuando el compilador achica el tipo adentro de un `if`, porque ahí ya sabe más. |
| **Opcional** | Un campo que puede no estar: `notes?: string`. |
| **`undefined`** | El valor que significa «acá no hay nada». |
| **`strict`** | El modo del compilador que no deja pasar lo dudoso. Está prendido todo el curso. |
| **Genérico** | Un tipo con un hueco, que se rellena al usarlo: `Page<Coffee>`. |
| **Utility type** | Un tipo que se arma a partir de otro: `Omit`, `Pick`, `Partial`. |
| **`readonly`** | Una marca que impide reasignar un campo o modificar un array. |
| **Módulo** | Cada archivo `.ts`. Lo que se `export`a desde uno se puede `import`ar en otro. |

---

## 0:00 · Pregunta de apertura — 5 min

**En pantalla:** diapositiva 2.

> «Pensá en la última vez que un programa te falló **usándolo**, no escribiéndolo. Se rompió con vos adelante, y cuando fuiste a ver qué había pasado era una tontería: un nombre mal escrito, un dato que no estaba, un número que en realidad era texto.»
>
> «Contame cuál fue en el chat. Con una línea alcanza.»

**Esperá 90 segundos en silencio.** Si nadie escribe, contá vos uno tuyo primero.

Van a decir «me faltó un dato», «era `undefined`», «puse mal el nombre de una propiedad», «no sé». **Todas sirven.** No corrijas ninguna.

Leé dos o tres en voz alta y cerrá así:

> «Fíjense en algo que tienen todas en común: **el error ya estaba escrito** antes de que el programa corriera. Estaba ahí, en el archivo, esperando. Nadie lo leyó hasta que fue tarde.»
>
> «Hoy vamos a ver quién puede leerlo antes.»

> ⚠️ No expliques nada todavía. Este bloque es para que hablen, no para que aprendan.

---

## 0:05 · Diagnóstico en vivo — 7 min

**En pantalla:** diapositivas 3 y 4. **No hay Wayground hoy**: es la primera clase.

> «De la clase que viene en adelante vamos a arrancar siempre con un quiz de lo anterior. Hoy no hay anterior, así que arrancamos con tres preguntas de JavaScript. No es examen, no se corrige: quiero ver qué sale.»

**Las tres van al chat, una por vez, 60 segundos cada una.** No adelantes la respuesta.

### 1

```js
const price = '42';
console.log(price * 2);
console.log(price + 2);
```

**Predicen.** Después ejecutás en la consola del navegador.

Sale **`84`** y **`'422'`**. La misma variable, dos comportamientos distintos.

> «El `*` convierte el texto a número. El `+` convierte el número a texto. Ninguna de las dos falló, y una de las dos está mal — pero JavaScript no sabe cuál, porque no sabe qué querías.»

### 2

```js
const race = { name: 'Clásico Apertura' };
console.log(race.nombre);
```

Sale **`undefined`**. Sin error, sin aviso, sin nada.

> «Escribí `nombre` en vez de `name`. Es el error más aburrido del mundo y JavaScript me deja seguir. Ese `undefined` va a viajar por tu programa hasta que rompa en un lugar que no tiene nada que ver, y ahí vas a debuggear el lugar equivocado.»

### 3

```js
function sizeFactor(size) {
  if (size === 'chico') return 0.8;
  if (size === 'grande') return 1.3;
  return 1;
}
console.log(sizeFactor('mediana'));
```

Sale **`1`**.

> «Pedí un tamaño que no existe —“mediana”, con A— y me cobró el precio del mediano. Sin error. Sin aviso. Un cliente pagó de menos y nadie se enteró nunca.»

**Cerrá el bloque así, y no te extiendas:**

> «Tres programas que **no fallaron**. Los tres están mal. Eso es lo que vamos a arreglar hoy, y no se arregla poniendo más cuidado: se arregla poniendo a alguien a revisar.»

---

## 0:12 · El concepto — 8 min

> **El editor está cerrado.** No hay VS Code en pantalla, no hay terminal. Solo diapositivas.
> Si alguien pregunta «¿y el código?», la respuesta es: «en ocho minutos. Primero quiero que puedan dibujar esto en una servilleta.»

### 0:12 — Qué es un tipo · 3 min

**En pantalla:** diapositiva 7 — `diagramas/un-tipo-es-un-conjunto.svg`.

**El término que se define acá:** *tipo*.

> «Un **tipo** es el conjunto de valores que una cosa puede tener. Nada más que eso.»

Señalá el panel izquierdo:

> «Acá está `string`. ¿Cuántas cadenas distintas hay? Infinitas. `'chico'` es una cadena. `'mediana'` también. `'asdf'` también, y la cadena vacía también. Cuando yo digo que el tamaño es un `string`, estoy diciendo *cualquiera de estas infinitas sirve*.»
>
> «Y no es lo que quiero decir. Yo quiero decir tres.»

Señalá el panel derecho:

> «Esto es el mismo dato con el tipo apretado. Se llama `CoffeeSize` y adentro tiene exactamente tres cadenas. No es una regla escrita en un comentario: **es el tipo**.»

Y ahora la línea de abajo, que es el premio:

> «Miren qué pasa con `'mediana'` en cada lado. A la izquierda devuelve 1 y nadie se entera —es exactamente lo que vimos recién—. A la derecha **no compila**, y el error te dice el archivo y la línea.»

### 0:15 — Quién revisa, y cuándo · 3 min

**En pantalla:** diapositiva 8.

**Los términos que se definen acá:** *compilador*, *`strict`*.

> «¿Quién hace esa revisión? Un programa que se llama **compilador**. Lee todo tu código, chequea que los tipos cierren, y recién ahí escribe el JavaScript que se va a ejecutar.»
>
> «Y acá está la parte que más cuesta el primer día: **TypeScript no existe en el navegador**. No hay un motor de TypeScript en ningún lado. Los tipos se borran. Lo que corre es JavaScript común, el mismo de siempre.»

Dejá que caiga y seguí:

> «Entonces, ¿para qué sirven? Para el rato en que escribís. Son un andamio: sostienen mientras construís y después no van a la obra terminada.»
>
> «Hay un modo del compilador que se llama **`strict`**, que es el que no deja pasar nada dudoso. En este curso está prendido siempre, en los cuatro proyectos. No lo vamos a apagar ni un día, y en un rato van a entender por qué es una decisión de cariño y no de maldad.»

**Preguntas que van a hacer:**

| Preguntan | Respondé |
|---|---|
| «¿Entonces el navegador no entiende TypeScript?» | «No. Nadie lo ejecuta. Alguien lo traduce a JavaScript antes, y eso es lo que corre. En Angular lo hace el CLI cuando levantás el servidor.» |
| «¿Y si me equivoco igual?» | «Te podés equivocar igual, claro. Lo que no vas a poder es equivocarte **en las cosas que el tipo dice**. Un nombre de propiedad mal escrito ya no te lo deja pasar.» |
| «¿Es más lento?» | «Escribir, un poco al principio. Corregir, muchísimo menos. Y el programa corre exactamente igual, porque lo que corre es JavaScript.» |
| «¿Por qué no usamos `any` y listo?» | «Porque `any` es apagar el chequeo en ese punto. Está prohibido en el curso y el verificador lo marca. Si aparece uno, lo miramos juntos.» |

### 0:18 — Las cuatro cosas de hoy · 2 min

**En pantalla:** diapositiva 9.

> «Todo lo de hoy son cuatro ideas, y las cuatro son la misma: **decir la verdad en el tipo**.»

| | Idea | La pregunta que contesta |
|---|---|---|
| 1 | **Uniones de literales** | ¿Cuáles valores, exactamente? |
| 2 | **Opcionales y `undefined`** | ¿Puede no estar? |
| 3 | **`readonly`** | ¿Se puede cambiar? |
| 4 | **Genéricos y utility types** | ¿Cómo no repito el mismo tipo dos veces? |

> «Las tres primeras las vamos a escribir juntos ahora. La cuarta la vamos a leer, y la practican en el ejercicio y en la tarea.»

> **Si vas tarde:** lo único recortable de este bloque son los dos minutos de la lista de arriba. El diagrama y «los tipos se borran», no.

---

## 0:20 · Live coding — 15 min

**Decilo textual antes de empezar:**

> «Cierren el editor. Los próximos quince minutos yo escribo y ustedes miran. No copien: van a hacer esto mismo después, y con las manos libres se entiende mejor. Si me equivoco, avisen.»

**En pantalla:** VS Code y el navegador lado a lado. Proyecto: **`lab/demo`**, archivo `src/app/sessions/s00/menu.ts`.

> **Antes de entrar al aula:** `node scripts/prep-demo.mjs` y después `cd lab/demo && npm start`.
>
> `lab/demo` es una copia descartable de `lab/starter`: arranca en el mismo estado en que están los alumnos, con los tipos flojos. **No le borres nada a `lab/solution`** — tu solución de referencia queda intacta.
>
> La secuencia de tecleo completa, con las tres roturas deliberadas, está en **`mision-profe.md`**. Ponela en el segundo monitor: este guión lleva la clase, ese archivo lleva el teclado.

### 0:20 — Inferencia y anotación · 2 min

Abrí `menu.ts` y pasá el mouse por encima de `MENU`, sin escribir nada.

> «Miren lo que dice el editor cuando apoyo el mouse: ya sabe que esto es una lista de cafés. **Nadie se lo dijo.** Lo dedujo del valor. A eso se le dice **inferencia**, y es la razón por la que TypeScript no es escribir el doble.»

Ahora escribí, en cualquier lado:

```ts
let price = 42;
price = 'cuarenta y dos';
```

**Rotura deliberada 1.** Leé el error completo en voz alta:

```
Type 'string' is not assignable to type 'number'.
```

> «*El tipo `string` no se puede asignar al tipo `number`*. Yo nunca escribí “number” en ningún lado: lo dedujo de que puse 42. Y ahora me lo sostiene.»

Borrá las dos líneas.

> «Cuando el valor está ahí, no se anota. Se anota cuando **todavía no está**: los parámetros de una función, lo que devuelve, lo que viene de afuera.»

### 0:22 — La unión de literales · 3 min

Mostrá `sizeFactor` tal como está, y llamala desde abajo:

```ts
sizeFactor('mediana');
```

> «Esto compila. Es el mismo error que vimos a las 0:05, ahora con TypeScript prendido. ¿Por qué pasa? Porque el tipo del parámetro es `CoffeeSize`, y `CoffeeSize` hoy es… —mostralo— **`string`**. Y `'mediana'` es un `string` perfectamente válido.»

Ahora apretalo:

```ts
export type CoffeeSize = 'chico' | 'mediano' | 'grande';
```

**Guardá.** El error aparece solo, sin tocar nada más:

```
Argument of type '"mediana"' is not assignable to parameter of type 'CoffeeSize'.
```

> «Cambié **una línea** y el compilador encontró el problema que estaba a cincuenta líneas de distancia. Eso es lo que compra apretar un tipo.»
>
> «`type` le pone nombre a un tipo. La barra vertical es una **unión**: “esto o esto o esto”. Y cada una de esas tres cadenas entre comillas es un **literal**: un tipo que tiene un solo valor adentro.»

### 0:25 — El `switch` que se queja solo · 3 min

Reescribí `sizeFactor`:

```ts
export function sizeFactor(size: CoffeeSize): number {
  switch (size) {
    case 'chico':
      return 0.8;
    case 'mediano':
      return 1;
    case 'grande':
      return 1.3;
  }
}
```

> «Fíjense que no hay `default`, y sin embargo compila. TypeScript sabe que `CoffeeSize` tiene tres valores y que los cubrí a los tres, así que sabe que esta función siempre devuelve algo.»

**Ahora la demostración que vale por todo el bloque.** Agregale un cuarto tamaño al tipo:

```ts
export type CoffeeSize = 'chico' | 'mediano' | 'grande' | 'jumbo';
```

**Rotura deliberada 2.** El error aparece en `sizeFactor`, no en el tipo:

```
Function lacks ending return statement and return type does not include 'undefined'.
```

> «Agregué un tamaño arriba y el compilador me llevó al único lugar del programa donde falta decidir qué hacer con él. **Eso** es lo que un comentario nunca va a hacer por ustedes.»

Sacá `| 'jumbo'`.

### 0:28 — Opcional y narrowing · 3 min

Mostrá el `MENU` y señalá los dos `notes: ''`.

> «Estos dos cafés no tienen nota de cata. Les pusimos cadena vacía porque el tipo nos obligó a poner algo. Es una mentira chiquita, y las mentiras chiquitas se propagan: ahora, en todo el programa, “no tiene nota” y “tiene una nota vacía” son lo mismo.»

Apretalo:

```ts
notes?: string;
```

> «El signo de pregunta dice **puede no estar**. Y ahora el tipo de `coffee.notes` ya no es `string`: es `string | undefined`.»

Borrá los dos `notes: ''` del `MENU` y guardá. **Mirá la pantalla del navegador**: el pizarrón ahora dice `Huila · Colombia — undefined`.

**Rotura deliberada 3.** No la arregles todavía.

> «Nadie se quejó. Compila. Y está mal, porque `describe` sigue preguntando si la nota es la cadena vacía, y ya no hay ninguna cadena vacía.»

Arreglalo:

```ts
if (coffee.notes === undefined) {
  return base;
}
return `${base} — ${coffee.notes}`;
```

> «Y acá pasa algo que quiero que vean. Apoyen el mouse sobre `coffee.notes` **arriba** del `if`: dice `string | undefined`. Ahora abajo: dice `string`, a secas.»
>
> «El compilador entendió que si llegaste hasta acá, no puede ser `undefined`. A eso se le dice **narrowing**: adentro del `if` el tipo se achica, porque ahí ya sabe más que afuera.»

### 0:31 — Lo que puede no estar · 2 min

Abrí la consola del navegador y ejecutá el caso vacío, o mostralo en el test:

> «`cheapest` promete devolver un café. Miren qué pasa si el menú está vacío.»

```
TypeError: Cannot read properties of undefined (reading 'price')
```

Mostrá la línea culpable:

```ts
let best = menu[0]!;
```

> «Ese signo de exclamación es una promesa: le estoy diciendo al compilador *confiá en mí, siempre hay al menos uno*. Y no puedo cumplirla.»
>
> «La regla es corta: **si te tienta poner un `!`, casi siempre lo que falta es decir la verdad en el tipo.**»

Cambiá la firma a `Coffee | undefined` y sacá el `!`. Mostrá que **el pizarrón sigue funcionando** porque el template ya tiene el `@else`.

### 0:33 — `readonly` · 2 min

Escribí:

```ts
MENU.push({ id: 'c9', name: 'Trucho', origin: 'Ninguno', price: 1 });
```

> «Esto compila. Estoy modificando el menú del negocio desde cualquier parte del programa.»

Poné `readonly Coffee[]` y guardá:

```
Property 'push' does not exist on type 'readonly Coffee[]'.
```

> «`readonly` no congela nada en tiempo de ejecución: es una marca para el compilador. Y alcanza, porque el que se equivoca es el que escribe.»
>
> «Esto va a estar en **todo** el curso, y en la clase 3 va a ser la diferencia entre que la pantalla se actualice y que no. Por ahora: los datos que vienen de afuera no se editan en el lugar.»

Borrá el `push`.

---

## 0:35 · Misión 1 — 15 min

**En pantalla:** diapositiva 18 con el enunciado. Enunciado completo en `mision-estudiante-1.md`.

> «Ahora ustedes. Mismo menú, en `lab/starter`, archivo `sessions/s00/menu.ts`. Hay cinco `TODO(S0)` y son casi lo mismo que acabamos de hacer. Quince minutos.»

**Decí estas dos cosas antes de largar, porque si no las van a preguntar quince veces:**

> «Uno: **el archivo del componente no se toca**. Los componentes son la clase que viene. Si algo se rompe ahí, el arreglo está en `menu.ts`.»
>
> «Dos: la terminal donde corre `npm start` es su segundo par de ojos. Tenganla a la vista: los errores de tipo aparecen ahí, con archivo y línea.»

> **Estás en silencio.** Disponible si preguntan, pero no circulás ofreciendo ayuda.

**Reloj de pistas** — solo si más de la mitad está trabada en lo mismo:

| Min | Pista, en voz alta, sin resolver |
|---|---|
| 0:43 | «Si borraron los `notes: ''` y todavía no tocaron la interfaz, el error que ven es correcto. Vayan al TODO 2 primero.» |
| 0:47 | «El orden que menos duele es: 1, 2, 3, después 4 y 5. El 2 y el 3 tocan el mismo lugar.» |

---

## 0:50 · Comparten pantalla — 10 min

Dos personas. **Una que le funciona y una que no** — a la segunda pedile permiso antes.

> **Preguntás, no corregís.** Aunque veas el error en el primer segundo.

1. «¿Qué esperabas que pasara acá?»
2. «¿Qué dice el error, con tus palabras?»
3. «¿Cómo te diste cuenta de dónde estaba?»
4. «Si tuvieras que explicarle esta línea a alguien que no vino hoy, ¿qué le decís?»

**Lo más probable que aparezca:** alguien resolvió el error de `notes` poniendo `notes: string | undefined` en vez de `notes?: string`. **No lo marques como error**, porque casi no lo es:

> «Fijate que no es lo mismo pero se parece mucho. `notes?: string` dice *el campo puede no venir*. `notes: string | undefined` dice *el campo viene siempre, y a veces trae undefined adentro*. Con `strict` prendido, el segundo te obliga a escribir `notes: undefined` en cada café que no tenga. Probalo y contame cuál te resulta más cómodo.»

---

## 1:00 · Descanso — 10 min

> «Diez minutos. Vuelvan puntuales, que lo que viene es lo mejor de la clase.»

---

## 1:10 · Predice y ejecuta — 15 min

**Los archivos:** `predice-y-ejecuta/`. **Las respuestas:** `predice-y-ejecuta/respuestas.md`, y están verificadas — no las improvises.

**El orden no se saltea:**

1. Mostrás el código. **No lo ejecutás.**
2. «¿Qué va a pasar? Escribilo en el chat.» — **60 segundos de reloj.**
3. Recién ahí, ejecutás.
4. Explicás la diferencia entre lo que dijeron y lo que pasó.

> El paso 2 es todo el ejercicio. Si ejecutás primero, esto es una demo y no aprende nadie.

| Min | Snippet | Casi todos predicen | Pasa |
|---|---|---|---|
| 1:10 | `const config = { size: 'grande' }` pasado a `sizeFactor` | «compila, es la misma cadena» | **No compila.** El tipo es `string`, no `'grande'` |
| 1:15 | `readonly lines: OrderLine[]` y después `order.lines.push(…)` | «no compila, dice readonly» | **Compila.** `readonly` protege el campo, no el array |
| 1:20 | `JSON.parse(texto) as Race` y después `race.horses.length` | «no compila» o «anda» | **Compila y explota** en tiempo de ejecución |

Cerrá el bloque con esta pregunta:

> «De los tres, ¿cuál les parece más peligroso en un proyecto de verdad?»

Casi siempre eligen el tercero, y tienen razón:

> «Exacto. Los dos primeros los descubrís escribiendo. El tercero es **vos jurándole algo al compilador**: `as` no verifica nada, solo hace que se calle. Cada vez que escriban `as` o `!`, la pregunta es una sola: *¿puedo cumplir esta promesa?*»

---

## 1:25 · Misión 2, en parejas — 20 min

**En pantalla:** diapositiva 24. Enunciado en `mision-estudiante-2.md`.

> «Ahora al proyecto de verdad. En parejas, veinte minutos: diez escribe uno y dicta el otro, y a los diez se invierten. El que dicta no toca el teclado; el que escribe no decide.»

**Tres cosas para decir antes de largar:**

> «Uno: el archivo es `core/models/race.model.ts`, y los nombres de los campos **ya están bien**. No se cambia ni uno: esos nombres son el contrato con el backend, y hay código Go del otro lado que espera exactamente esos.»
>
> «Dos: lo que hay que apretar sale de `docs/contract/openapi.yaml`. No se adivina, se abre. Si el contrato dice tres estados, son tres.»
>
> «Tres, y es la parte importante: cuando terminen, el ejercicio **no es que compile**. Es que **cinco líneas que hoy compilan dejen de compilar**. Están escritas al final del enunciado. Péguenlas, mírenlas fallar, y bórrenlas.»

Circulás entre las parejas. **Escuchás más de lo que hablás.** La pareja que termina recibe la extensión del final del enunciado.

---

## 1:45 · Code review en vivo — 10 min

Una solución de la Misión 2, con permiso. **En pantalla, al lado, `correccion.md`.**

**Decí esto primero, textual, porque si no van a buscar la rúbrica del curso y no van a encontrar la mitad:**

> «La rúbrica que vamos a usar todo el módulo arranca con “¿es standalone?” y “¿es OnPush?”. Hoy esas dos preguntas no aplican: no hay un solo componente en juego. Desde la clase que viene sí. Hoy revisamos otras cinco.»

Rúbrica de S0, en voz alta y en este orden:

1. ¿El tipo dice **exactamente** lo que el contrato garantiza — ni más ni menos?
2. ¿Quedó algún `!` o algún `as` que sea una promesa que el código no puede cumplir?
3. ¿Lo que no se edita está `readonly`?
4. ¿Los nombres están en inglés y dicen lo que la cosa es?
5. ¿Se abrió el contrato, o se adivinó?

**Empezá por algo que está bien hecho.** Siempre hay algo.

Y cerrá con esto:

> «Ninguno de estos cambios se ve en la pantalla. La app hace exactamente lo mismo que hacía hace veinte minutos. Lo que cambió es que ahora hay una clase entera de errores que ya no se puede escribir — y eso lo van a agradecer en la clase 7, cuando estos mismos tipos vengan de un servidor de verdad.»

---

## 1:55 · Exit ticket y tarea — 5 min

**Exit ticket:** `exit-ticket.md`. Tres preguntas, tres minutos.

> «La tercera es la que más me sirve: qué quedó confuso. Vale “nada”, vale “todo”, vale una palabra. Con eso arranco la clase que viene.»

**Tarea:** `tarea.md`. **Leela en voz alta antes de cortar.** Una tarea que solo se manda por chat no se hace.

**Y decí esto, que es lo que más se olvida:**

> «Última cosa, y es importante. En el repo les dejé un archivo que se llama **`conceptos.md`**. Está todo lo de hoy: cada concepto con su definición y **los ejemplos exactos que corrimos acá**, incluidos los errores con su mensaje literal. La clase es en vivo y no queda grabada, así que cuando se sienten a hacer la tarea, eso es lo que tienen en vez de acordarse. Ténganlo abierto al lado del editor.»

> ⚠️ No lo saltees por falta de tiempo. Si nadie sabe que el apunte existe, es como si no existiera — y es lo único que tienen entre esta clase y la que viene.

**Y el aviso de la próxima:**

> «La clase que viene arrancamos con Angular de verdad: el primer componente. Vengan con el proyecto levantado, porque los quince minutos de instalar cosas se los quiero ahorrar.»

---

## Después de la clase

- [ ] Leer los exit tickets. Lo confuso arranca S1.
- [ ] Revisar `wayground.csv` de **esta** sesión con lo que más falló — se corre al empezar S1.
- [ ] Completar las notas de abajo.

### Notas de la corrida real

*Completá después de dar la clase. Es lo que hace que S1 salga mejor.*

| | |
|---|---|
| ¿Qué bloque se pasó de tiempo? | |
| ¿Qué pregunta no supe contestar? | |
| ¿Qué error apareció que no estaba previsto? | |
| ¿Cuántos llegaron con el entorno instalado? | |
| ¿Qué sacaría o agregaría? | |
