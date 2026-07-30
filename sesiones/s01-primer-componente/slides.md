---
marp: true
theme: neobrutal
paginate: true
header: 'S1 · Primer componente'
---

<!-- _class: portada -->

# S1

## Primer componente

<!--
Módulo Angular · Talento DH 8va.

El guión completo está en guion.md, y es un teleprompter: lo que está entre
comillas se dice. Estas notas son el resumen de cada diapositiva; si algo no
está acá, está allá.

Se ven con la tecla P en el HTML de Marp.
-->

---

<!-- _class: bloque -->

# 0:00

## Pregunta de apertura

<!--
5 minutos. Responden en el chat. Sin juicio, sin corregir, sin "casi".
Esperá 90 segundos en silencio. Si nadie escribe, respondé vos primero.
-->

---

## Pensá en una web que uses todos los días

Cuando cambia un número en pantalla —el carrito, un contador, un saldo—

# ¿quién lo cambió?

<!--
Van a decir "el servidor", "JavaScript", "React", "no sé". TODAS SIRVEN.
No corrijas ninguna.

Cerrá con esto, textual:
"Todas tienen algo de razón. Lo que quiero que se lleven es que ALGUIEN tuvo
que agarrar ese pedazo de pantalla y cambiarlo. No se actualizó solo. Hoy
vamos a ver quién hace eso en Angular, y por qué eso es todo el asunto."

NO EXPLIQUES NADA TODAVÍA. Este bloque es para que hablen.
-->

---

<!-- _class: bloque -->

# 0:05

## Wayground

## de S0

<!--
sesiones/s00-typescript/wayground.csv — las preguntas de la clase pasada.

Decilo: "de acá en adelante arrancamos siempre así: siete minutos repasando
lo de la clase pasada. No se corrige nota, se corrigen ideas."

Máximo 30 segundos por pregunta, y solo si la falló más de un tercio. Los
cuatro tropiezos esperables, con su respuesta, están en el guión. Si algo
necesita más de 30 segundos, va a la tarea: no te enganches acá.
-->

---

<!-- _class: bloque -->

# 0:12

## El concepto

## Editor cerrado

<!--
A partir de acá, ni una línea de código en pantalla. Solo diagrama.

Si alguien pregunta "¿y el código?": "en diez minutos. Primero quiero que
puedan dibujar esto en una servilleta."
-->

---

## Un contador, sin framework

```html
<p>Llevás 0 clics</p>
<button>Sumar</button>
```

```js
let clics = 0;
boton.onclick = () => clics + 1;
```

## El número no se mueve.

<!--
DOS MINUTOS. El problema ANTES de nombrar nada.

"El HTML ya se dibujó. Ese número dejó de ser tu variable en el momento en
que se pintó. Para que cambie, tenés que ir a buscar el elemento y
escribirle el valor nuevo, a mano."

"Con un contador es molesto. Con veinte carreras, cada una con ocho caballos
y sus cuotas, es IMPOSIBLE de sostener: cada vez que cambia un dato tenés
que acordarte de todos los lugares de la pantalla que lo muestran."
-->

---

## Eso que hay que ir a tocar tiene nombre

# El DOM

El árbol de elementos que el navegador tiene en memoria y dibuja en pantalla.

Sin framework, mantenerlo sincronizado con tus datos **es trabajo tuyo**.

<!--
Primera palabra del glosario. Definila y seguí.

"Y es el trabajo que más bugs genera en cualquier interfaz: no porque sea
difícil, sino porque hay que acordarse siempre."
-->

---

## De dónde viene Angular

**2010 · AngularJS** — que el HTML pudiera decir de dónde salen sus datos

**2016 · Angular 2** — reescritura. La idea de los *Web Components*: una parte de la pantalla puede ser una pieza cerrada

**2024 · standalone** — se va la capa de `NgModule`. El componente se declara solo

<!--
TRES MINUTOS. No más.

Frase para dejar: "Angular no inventó los componentes. Los hizo tipados, con
herramientas y con un ciclo de vida."

Las tres preguntas que van a hacer —NgModules, AngularJS, por qué la 18— y
sus respuestas están en el guión.

SI VAS TARDE: este bloque es lo único recortable. El diagrama y la detección
de cambios, no.
-->

---

## Un componente son dos cosas

![w:860](diagramas/componente-y-template.svg)

<!--
CINCO MINUTOS sobre este diagrama.

"A la izquierda, LA CLASE: un archivo TypeScript. Ahí viven los datos y las
decisiones."

"A la derecha, EL TEMPLATE: un archivo HTML. Y ojo, que es la primera trampa:
EL TEMPLATE NO ES HTML COMÚN. Es HTML con instrucciones adentro, y esas
instrucciones las lee Angular antes de dibujar."

La analogía: un mostrador de cafetería. La clase es la trastienda; el
template, la vidriera. Interpolación = escribir el precio en el cartel.
Property binding = apagar la luz del cartel cuando se acabó. Event binding =
el timbre. Two-way = la libreta del pedido.

Señalá cada flecha mientras la nombrás.
-->

---

## Binding quiere decir *atadura*

Atás un pedazo del template a un pedazo de la clase,

y Angular se encarga de que **no se despeguen**.

<!--
La palabra más importante del día. Definila así, en castellano, antes de
mostrar un solo corchete.

"Nunca más vas a ir a buscar un elemento para escribirle un número."
-->

---

<!-- _class: ojo -->

# ¿Cómo se entera Angular?

No hay magia y no hay nadie mirando tus variables.

**Angular repinta después de que pasó algo.**

<!--
LA PREGUNTA QUE CASI NADIE HACE Y TODOS NECESITAN.

"Angular espera a que pase ALGO: un clic, una tecla, una respuesta del
servidor. Cuando eso pasa, revisa los bindings de la pantalla y repinta lo
que haga falta. A ese momento se le dice DETECCIÓN DE CAMBIOS."

"Guardensé esa frase. En la clase 10 vamos a ver un caso donde pasa algo y
Angular NO se entera, y va a tener todo el sentido del mundo."

Sin esta diapositiva, ni el ejercicio de predicción de hoy ni el OnPush de
S10 se pueden explicar.
-->

---

## Las palabras de hoy

| Palabra | Qué es |
|---|---|
| **DOM** | El árbol que el navegador dibuja |
| **Componente** | Una clase + un template |
| **Template** | HTML con instrucciones para Angular |
| **Standalone** | Declara solo lo que usa |
| **Binding** | La atadura entre la clase y el template |
| **Detección de cambios** | Cuando Angular revisa y repinta |

<!--
Checkpoint de treinta segundos. Preguntá: "¿alguna de estas seis no quedó
clara?" y esperá de verdad.

El glosario completo, con las trece palabras, está en el guión.
-->

---

<!-- _class: bloque -->

# 0:20

## Live coding

## Ustedes miran

<!--
DECILO TEXTUAL: "Cierren el editor. Los próximos quince minutos yo escribo y
ustedes miran. No copien: van a hacer esto mismo después, y con las manos
libres se entiende mejor. Si me equivoco, avisen."

Proyecto: lab/demo, preparado antes de entrar con `node scripts/prep-demo.mjs`.
NO se trabaja en lab/solution. La tabla minuto a minuto está en el
guión. Si querés escribirlo desde cero, borrá antes src/app/sessions/s01/.
-->

---

<!-- _class: codigo -->

## El CLI crea los archivos

```bash
ng generate component sessions/s01 --flat
```

Cuatro archivos: el `.ts`, el `.html`, el `.css` y uno de tests.

<!--
"Esto es el CLI, la herramienta de línea de comandos de Angular."

"Podría haberlos creado a mano. Uso el CLI porque los nombra igual siempre y
porque no me olvido de ninguno. NO HACE NADA MÁGICO: mírenlos, están vacíos."

Abrí el .ts y leelo línea por línea: qué es un decorador, qué es el selector,
qué dice standalone, qué apuntan templateUrl y styleUrl.
-->

---

<!-- _class: codigo -->

## Hacen falta DOS cosas para que se vea

```ts
// app.routes.ts — para que la dirección exista
{ path: 's01', loadComponent: … }
```

```ts
// sessions.ts — para que aparezca en el menú
{ number: 1, slug: 's01', available: true }
```

<!--
"Son dos archivos distintos y hay que tocar los dos. Es el error más común
de la clase que viene."

El router resuelve direcciones; el menú es una lista que dibuja un
componente. Ninguno se entera del otro.

Navegá a /s01: se ve la página vacía que dejó el CLI.
"De acá para arriba, todo lo que aparezca lo vamos a poner nosotros."
-->

---

<!-- _class: codigo -->

## 1 · Interpolación

```ts
protected coffee = { name: 'Yirgacheffe', price: 42 };
```

```html
<h2>{{ coffee.name }}</h2>
<p>{{ coffee.price }}</p>
```

<!--
"Las llaves dobles dicen: Angular, poné acá el valor de esta expresión, como
texto."

AHORA LA DEMOSTRACIÓN QUE VALE POR TODA LA EXPLICACIÓN: cambiá price a 55 y
guardá.

"Cambié UN SOLO LUGAR —la clase— y la pantalla se actualizó sola. Ese es el
trato: el dato vive en un lugar, y el template lo mira."
-->

---

<!-- _class: codigo -->

## 2 · Property binding

```html
<!-- queda puesta SIEMPRE -->
<div class="product product--soldout">

<!-- depende de un dato -->
<div class="product" [class.product--soldout]="!coffee.available">
```

<!--
Escribí primero la versión de arriba, mostrá que está mal, y recién ahí la
de abajo.

"Los corchetes cambian todo. SIN corchetes, lo que va entre comillas es
TEXTO LITERAL. CON corchetes, es una expresión de TypeScript que Angular
evalúa."

"Y fíjense que puedo tener las dos cosas en el mismo elemento: class con lo
que va siempre, y [class.algo] con lo que va a veces. No se pisan, se suman."

Cambiá available a false y mostralo.
-->

---

<!-- _class: codigo -->

## 3 · Event binding

```html
<button (click)="toggleAvailability()">
  {{ coffee.available ? 'Marcar agotado' : 'Marcar disponible' }}
</button>
```

<!--
"Los paréntesis escuchan un evento del navegador. `click` no lo inventó
Angular: es el mismo de siempre. Lo que agrega Angular es que en vez de
escribir un addEventListener, decís qué método llamar."

Tocá el botón un par de veces y conectá con el concepto:

"Acá está lo de la detección de cambios: TOCARON EL BOTÓN, entonces pasó
algo, entonces Angular revisó. Por eso se actualizaron el texto del botón Y
la clase del div, sin que yo tocara nada más."
-->

---

<!-- _class: ojo -->

# Y acá va a fallar

`NG8002: Can't bind to 'ngModel'`

`since it isn't a known property of 'input'.`

<!--
ROMPELO A PROPÓSITO. Escribí [(ngModel)] SIN FormsModule y leé el error
completo en voz alta.

"Buenísimo que falle. Léanlo conmigo: no puedo enlazar a ngModel porque no
es una propiedad conocida de input."

"Y tiene razón: ngModel no existe en HTML, lo trae Angular. Y este
componente es STANDALONE, o sea que declara solo lo que usa. Yo nunca le
dije que iba a usar ngModel."

Es el error número uno de la sesión. Que lo vean acá hace que a las 0:35 lo
reconozcan solos.
-->

---

<!-- _class: codigo -->

## 4 · Two-way binding

```ts
imports: [FormsModule],   // ← sin esto, no compila
```

```html
<input [(ngModel)]="customer" />
<p>Hola, {{ customer }}</p>
```

<!--
"La regla completa: SI EL TEMPLATE USA ALGO DE ANGULAR, TIENE QUE ESTAR EN
IMPORTS. Vale para esto, para el router, y desde la clase que viene, para
cada componente que usen adentro de otro."

Escribí en el input y mostrá cómo el <p> cambia mientras escribís.

"[(ngModel)] es literalmente [ngModel] más (ngModelChange): property binding
y event binding juntos, con azúcar sintáctica."
-->

---

<!-- _class: bloque -->

# 0:35

## Misión 1

## Individual · 15 min

<!--
lab/starter. El enunciado completo está en mision-estudiante-1.md.

ESTÁS EN SILENCIO. Disponible si preguntan, pero no circulás ofreciendo
ayuda. Los quince minutos de pelearse con el error SON la clase.

Reloj de pistas, solo si más de la mitad está trabada en lo mismo:
  0:43 — "¿A alguien le apareció un error con ngModel? Vuelvan a leerlo
         entero: dice dónde está el problema."
  0:47 — "Acordate: para que la ruta aparezca en el menú hay que tocar DOS
         archivos."
-->

---

## Misión 1 · El mostrador

**El componente no existe. Lo creás vos.**

1. Que la pantalla exista — CLI, ruta e índice
2. **Interpolación** — nombre, origen y precio
3. **Property binding** — la clase de agotado
4. **Event binding** — el botón de disponibilidad
5. **Two-way** — cliente y cantidad

**Listo cuando** cambiar el precio en el `.ts` cambia lo que se ve, y el total se actualiza mientras escribís.

<!--
Recalcá el "no existe": arrancan de una barra lateral vacía y una pantalla
de inicio, no de un archivo con huecos.

Si alguien se queda muy atrás, correccion.md Parte A tiene el paso a paso.
Para destrabarse, no para copiar.
-->

---

<!-- _class: bloque -->

# 0:50

## Comparten pantalla

<!--
Dos personas: una que le funciona y una que no. A la segunda, pedile permiso
antes.

PREGUNTÁS, NO CORREGÍS. Aunque veas el error en el primer segundo.
  ¿Qué esperabas que pasara? · ¿Qué pasó? · ¿Cómo te diste cuenta?
  ¿Cómo se lo explicarías a alguien que no vino hoy?

Lo más probable: alguien escribió class="product--soldout" en vez de
[class.product--soldout]="…". Es la mejor pantalla posible para compartir —
no la arregles vos.
-->

---

<!-- _class: bloque -->

# 1:00

## Descanso

## 10 minutos

<!-- "Vuelvan puntuales, que lo que viene es lo mejor de la clase." -->

---

<!-- _class: bloque -->

# 1:10

## Predice y ejecuta

<!--
EL ORDEN NO SE SALTEA:
  1. Mostrás el código. NO lo ejecutás.
  2. "¿Qué va a pasar? Escribilo en el chat." — 60 SEGUNDOS DE RELOJ.
  3. Recién ahí, ejecutás.
  4. Explicás la diferencia entre lo que dijeron y lo que pasó.

El paso 2 es todo el ejercicio. Si ejecutás primero, esto es una demo.

Las respuestas están en predice-y-ejecuta/respuestas.md y están verificadas
en el navegador. No las improvises.
-->

---

<!-- _class: codigo -->

## 1 · ¿Qué clases quedan?

```html
<p class="label {{ tone }}"
   [class.label--active]="isActive">Hola</p>
```

```ts
tone = 'red';
isActive = true;
```

<!--
No ejecutar. 60 segundos.

Casi todos dicen que se pisan. NO: quedan LAS TRES.
  class="label red label--active"

class y [class.x] no compiten: se combinan. El binding específico es dueño
de SU clase y de ninguna otra.

Por eso en la Misión 2 van a usar las dos juntas.
-->

---

<!-- _class: codigo -->

## 2 · ¿Qué pasa al escribir?

```ts
@Component({
  standalone: true,
  imports: [],
  template: `<input [(ngModel)]="name" />
             <p>Hola, {{ name }}</p>`,
})
export class DemoComponent { name = ''; }
```

<!--
Opciones: 1) anda · 2) no se actualiza pero funciona · 3) no arranca ·
4) no compila.

Casi nadie dice la 4, y es la 4:
  NG8002: Can't bind to 'ngModel' since it isn't a known property of 'input'.

Ya lo vieron en el live coding. Si alguien lo reconoce, festejalo.
-->

---

<!-- _class: codigo -->

## 3 · ¿Cuánto marca después de 3 clics?

```html
<button (click)="counter + 1">Sumar</button>
<p>Llevo {{ counter }} clics</p>
```

<!--
Se queda en 0. Y NO HAY NINGÚN ERROR — ni en consola, ni en el build.

La expresión se evalúa y el resultado se tira. Nadie lo asignó.

CERRÁ EL BLOQUE CON ESTA PREGUNTA:
"De los tres, ¿cuál les habría costado más encontrar en un proyecto de
verdad?"

Casi siempre eligen este, y tienen razón: los otros dos te frenan el build.
"Un binding que no falla no quiere decir que ande. El silencio es el peor
síntoma que puede tener un bug."
-->

---

<!-- _class: bloque -->

# 1:25

## Misión 2

## Al hipódromo

<!--
En parejas, 20 min, project/frontend/starter.
Conducción por turnos: 10 min escribe uno y dicta el otro, después se
invierte. El que dicta no toca el teclado; el que escribe no decide.

TRES COSAS PARA DECIR ANTES DE LARGAR:
  1. La pantalla NO EXISTE. La crean ustedes, con el CLI.
  2. Los datos ya están: las 8 carreras reales en core/mocks. No inventen.
  3. El @for se los damos escrito en la corrección. Recorrer listas es S3;
     hoy el trabajo son los bindings.

Circulá. Escuchá más de lo que hablás.
-->

---

## Misión 2 · El programa de carreras

**La pantalla no existe. La crean ustedes.**

- **Interpolación** — estado, nombre, hora, competidores
- **Property binding** — `race--live`, `race--finished`, `race--open`
- **Event binding** — abrir y cerrar la parrilla
- **Two-way** — el monto del simulador

Los datos ya están en `core/mocks`. El CSS también.

<!--
Si alguien se queja del orden de la lista: BIEN, que les moleste. Ordenar y
filtrar es S3. Hoy se pintan los datos como vienen.

La pareja que termina recibe la extensión del final del enunciado.
-->

---

<!-- _class: bloque -->

# 1:45

## Code review en vivo

<!--
Una solución de la Misión 2, con permiso. Tené correccion.md abierto al lado.

Rúbrica, en voz alta y en orden:
  1. ¿standalone y OnPush? · 2. ¿actualiza sin mutar? · 3. ¿any, console.log,
  imports de más? · 4. ¿el nombre dice lo que hace? · 5. ¿está en la carpeta
  que le toca?

EMPEZÁ POR ALGO QUE ESTÁ BIEN HECHO. Siempre hay algo.

Y decí esto textual, porque si no lo van a esperar:
"Se van a dar cuenta de que no hay estado de carga ni de error. Es correcto:
hoy no hay NADA que cargar, los datos ya están en el código. En la clase 7,
cuando los pidamos a un servidor, los tres estados van a ser obligatorios."
-->

---

<!-- _class: bloque -->

# 1:55

## Exit ticket

<!--
Tres preguntas, 3 minutos.

"La tercera es la que más me sirve: qué quedó confuso. Vale nada, vale todo,
vale una palabra. Con eso arranco la clase que viene."

Leé la tarea EN VOZ ALTA antes de cortar. Una tarea que solo se manda por
chat no se hace.
-->

---

<!-- _class: portada -->

# Nos vemos en S2

## `input()`, `output()` y `ng-content`
