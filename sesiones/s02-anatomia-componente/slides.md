---
marp: true
theme: neobrutal
paginate: true
header: 'S2 · Anatomía de un componente'
---

<!-- _class: portada -->

# S2

## Anatomía de un componente

<!--
Módulo Angular · Talento DH 8va.

El guión completo está en guion.md, y es un teleprompter. Estas notas son el
resumen de cada diapositiva. Se ven con la tecla P en el HTML de Marp.
-->

---

<!-- _class: bloque -->

# 0:00

## Pregunta de apertura

<!--
5 minutos, en el chat. Sin juicio, sin corregir. 90 segundos de silencio.
-->

---

## El listado de carreras ya está hecho

Te piden mostrar **una sola carrera** en la portada, con el mismo aspecto.

# ¿Qué hacés?

<!--
Van a decir "copio y pego el HTML". TODAS SIRVEN.

Cerrá con esto, textual:
"El que dijo copiar y pegar tiene razón: es lo único que se puede hacer hoy. Y
también es exactamente el problema — a partir del jueves hay DOS lugares que
hay que acordarse de cambiar juntos."

"Hoy vamos a ver cómo se hace para que haya uno solo."
-->

---

<!-- _class: bloque -->

# 0:05

## Wayground

## de S1

<!--
sesiones/s01-primer-componente/wayground.csv

Máximo 30 segundos por pregunta, y solo si la falló más de un tercio. Los tres
tropiezos esperables están en el guión.

El de `imports` conviene comentarlo aunque salga bien: hoy se agrega un caso
nuevo, el componente propio.
-->

---

<!-- _class: bloque -->

# 0:12

## El concepto

## Editor cerrado

<!--
Ni una línea de editor en pantalla. Solo diapositivas.
-->

---

## Veinte líneas de HTML, dos veces

El HTML repetido no es feo.

# Es un compromiso de mantenimiento que nadie firmó

<!--
DOS MINUTOS.

"Copiarlas no es difícil. El problema es el mes que viene, cuando alguien pida
agregar la distancia de la carrera. Hay que acordarse de los dos lugares. Y en
seis meses son cinco lugares, y siempre hay uno que quedó viejo."
-->

---

## Un componente tiene dos puertas

![w:900](diagramas/datos-bajan-avisos-suben.svg)

<!--
SEIS MINUTOS sobre este diagrama. Es el centro de la clase.

Flecha que baja: "LOS DATOS BAJAN. El padre le pasa al hijo lo que necesita
para dibujarse. Eso es input(), y es DE SOLO LECTURA: el hijo mira, no toca."

Flecha que sube: "LOS AVISOS SUBEN. El hijo no decide nada importante: cuando
lo tocan, avisa. Y el padre decide, porque es el único que sabe qué pasa con
las otras siete tarjetas."

El hueco punteado: "ng-content da vuelta el problema. El hijo dice ACÁ HAY UN
HUECO y el padre mete lo que quiera. El hijo no sabe qué es y no lo puede leer."

La pregunta que van a hacer: "¿y por qué el hijo no cambia el dato y listo?"
→ "Porque entonces habría dos lugares donde se decide lo mismo, y tarde o
temprano dicen cosas distintas."
-->

---

<!-- _class: ojo -->

# El hijo nunca modifica lo que le prestaron

Pide, y el padre decide.

<!--
La frase de la clase. Dejala en pantalla unos segundos.

Es la que contesta las tres preguntas difíciles del día: por qué la comanda se
queda en el padre, por qué `selected` es un input, y por qué el hijo emite en
vez de escribir.
-->

---

## Las palabras de hoy

| Palabra | Qué es |
|---|---|
| **Padre · hijo** | Quién usa a quién |
| **`input()`** | Un dato que baja |
| **`output()`** | Un aviso que sube |
| **`model()`** | Las dos cosas: `[( )]` |
| **`ng-content`** | Un hueco que llena el padre |
| **Ciclo de vida** | Se crea · cambia · se destruye |

<!--
Checkpoint de treinta segundos. "¿Alguna de estas seis no quedó clara?" y
esperá de verdad.

El glosario completo, con las quince palabras, está en el guión.
-->

---

<!-- _class: bloque -->

# 0:20

## Live coding

## Ustedes miran

<!--
DECILO TEXTUAL: "Cierren el editor. Los próximos quince minutos yo escribo y
ustedes miran. No copien."

Proyecto: lab/demo, ruta /s02. La ruta la sumás VOS antes de entrar al aula.

Secuencia de tecleo y orden de sacrificio: mision-profe.md, segundo monitor.
-->

---

<!-- _class: codigo -->

## Cortar antes de pensar · 0:20

```bash
ng generate component sessions/s02/coffee-card --flat
```

## 17 errores. Todos dicen `item` no existe.

<!--
ROTURA DELIBERADA 1, y es la más importante de la clase.

Cortá —cortá, no copies— el <article class="card"> entero y pegalo en el hijo.
Guardá SIN arreglar nada.

"Cada error es una cosa que el hijo necesita de afuera. No hay que adivinar los
inputs: la lista ya está escrita, en la terminal."

Quien intenta declarar los inputs antes de cortar se olvida de dos.
-->

---

<!-- _class: codigo -->

## `input.required()` · 0:22

```ts
readonly coffee = input.required<Coffee>();
```

```html
<h3>{{ coffee().name }}</h3>
```

## `NG8008: Required input 'coffee' … must be specified.`

<!--
"input() no devuelve el dato: devuelve una FUNCIÓN que da el dato. Por eso los
paréntesis. Es la misma forma de los signals de la clase que viene, y no es
casualidad: es lo mismo."

Usalo sin [coffee] a propósito para que salga el NG8008.

"required no documenta. EXIGE. Es el mismo tipo de promesa que los tipos de la
sesión 0."
-->

---

<!-- _class: codigo -->

## `output()` · 0:25

```ts
readonly ordered = output<OrderRequest>();
this.ordered.emit({ coffee: this.coffee(), quantity: this.quantity() });
```

```html
<app-coffee-card (ordered)="take($event)" />
```

<!--
"Los paréntesis son los mismos de (click). La diferencia es que click lo inventó
el navegador y ordered lo inventé yo hace treinta segundos."

HOVER SOBRE $event: "acá no es un evento del DOM. Es exactamente el objeto que
puse en emit(), con su tipo. Dice OrderRequest."

ROTURA DELIBERADA 2 — borrá el (ordered) y tocá Pedir. NO PASA NADA, sin ningún
error. Dejá el silencio tres segundos. Vuelve a las 1:15.
-->

---

<!-- _class: codigo -->

## `model()` · 0:29

```ts
readonly quantity = model(1);
```

```html
<app-coffee-card [(quantity)]="item.quantity" />
```

<!--
"input() y output() en la misma línea. Habilita los corchetes con paréntesis
adentro sobre un componente MÍO — lo que en S1 solo se podía con ngModel sobre
un input de HTML."

"Y miren cómo escribe el hijo: quantity.set(...), no quantity = ... Es un
signal. La clase que viene es entera sobre eso."
-->

---

<!-- _class: codigo -->

## `ng-content` · 0:31

```html
<!-- el hijo -->
<ng-content select="[card-tag]" />
<ng-content />
```

```html
<!-- el padre -->
<app-coffee-card [coffee]="item.coffee">
  <span card-tag class="tag">Café del día</span>
</app-coffee-card>
```

<!--
"Todo lo que el padre escriba entre la etiqueta que abre y la que cierra no lo
dibuja el padre: VIAJA ADENTRO DEL HIJO y cae en el hueco."

El que no lleva select es el cajón de sastre: recibe todo lo demás, y va UNO
SOLO. Qué pasa si ponés dos se ve a las 1:20.

La regla, si preguntan cuándo input y cuándo ng-content:
un DATO con el que el hijo decide → input(). MARCADO que solo muestra →
ng-content.
-->

---

## El ciclo de vida · 0:33

| Gancho | Cuándo | Para qué |
|---|---|---|
| `ngOnInit` | una vez, con los inputs llenos | preparar |
| `ngOnChanges` | cambió un input | reaccionar |
| `ngOnDestroy` | se va de la pantalla | soltar |

<!--
"Con estos tres alcanza para todo el curso. Angular los llama por vos."

Mostralos en este orden:
1. Recargá: cada tarjeta dice la hora en que se montó, ngOnChanges ×1.
2. Subí una cantidad: pasa a ×2. PREGUNTÁ POR QUÉ y esperá.
   → model() es entrada Y salida: el valor sube al padre y VUELVE A BAJAR por
     el mismo binding. Para el hijo es un input que cambió.
3. Sacá el Antigua de la carta: aparece el aviso de ngOnDestroy.

"Nadie llamó a ngOnDestroy. Hoy solo avisa un nombre; en la clase 6 va a ser lo
que evita que la aplicación se coma la memoria."
-->

---

<!-- _class: bloque -->

# 0:35

## Misión 1

## Partir la carta en dos

<!--
15 minutos, individual, lab/starter. Enunciado en mision-estudiante-1.md.

DECÍ ESTO ANTES DE LARGAR, que es la trampa del ejercicio:
"El orden que menos duele es: primero cortás y pegás el HTML, DEJÁS QUE SE
ROMPA TODO, y después cada error te dice qué input falta. No trates de adivinar
los inputs antes de cortar."

"Y acordate de que un componente propio también va en imports."

ESTÁS EN SILENCIO. Reloj de pistas: 0:43 y 0:47.
-->

---

## Misión 1 — los seis

1. La ruta `/s02` existe y aparece en el menú
2. La tarjeta es un componente, **con su CSS**
3. Entradas: `coffee` obligatoria · `featured` · `quantity`
4. Salida: avisa el pedido, **no lo escribe**
5. Dos huecos: con `select` y sin
6. `ngOnInit` · `ngOnChanges` · `ngOnDestroy`

<!--
Dejala en pantalla los quince minutos.

El requisito 4 es el que separa: la comanda se queda en el padre.
-->

---

<!-- _class: bloque -->

# 0:50

## Comparten pantalla

<!--
Dos personas. Una que le funciona y una que no — permiso ANTES.
PREGUNTÁS, NO CORREGÍS. Las cuatro preguntas están en el guión.

Lo más probable: alguien se llevó la comanda al hijo. NO es un error de
escritura, es de reparto, y es el más instructivo del día.
"Pedí dos cafés distintos y contame cuántas comandas hay."
-->

---

<!-- _class: bloque -->

# 1:00

## Descanso

## 10 minutos

---

<!-- _class: bloque -->

# 1:10

## Predice y ejecuta

<!--
Respuestas VERIFICADAS en predice-y-ejecuta/respuestas.md. No las improvises:
dos de las tres son contraintuitivas.

EL ORDEN NO SE SALTEA: mostrar → 60 segundos de predicción → ejecutar →
explicar.
-->

---

<!-- _class: codigo -->

## 1 · Una tarjeta sin café

```ts
readonly coffee = input.required<Coffee>();
```

```html
<app-coffee-card />
```

<!--
NO COMPILA.
NG8008: Required input 'coffee' from component CoffeeCardComponent must be
specified.

La comparación que vale decir en voz alta:
- input.required<Coffee>()          → error de compilación
- input<Coffee | undefined>(undefined) → compila y se rompe en runtime

"Es la misma decisión de la sesión 0: o el tipo dice la verdad y el compilador
te frena, o vos prometés algo y lo descubrís con un usuario adelante."
-->

---

<!-- _class: codigo -->

## 2 · Un aviso que nadie escucha

```html
<app-coffee-card [coffee]="item.coffee" />
```

## Toco «Pedir». ¿Qué pasa?

<!--
OPCIÓN 3: compila, y NO PASA ABSOLUTAMENTE NADA. Sin error de compilación, sin
advertencia, sin una línea en la consola.

ABRÍ LA CONSOLA DEL NAVEGADOR ANTES DE TOCAR. Que se vea vacía es medio
ejercicio.

"Angular no puede avisar de esto, y no es un descuido: no tiene forma de saber
si un output que nadie escucha es un olvido o es a propósito."

El primo: (click)="contador + 1" de S1. Misma familia.
-->

---

<!-- _class: codigo -->

## 3 · Dos huecos iguales

```html
<div class="card__slot"><ng-content /></div>
…
<div class="card__slot"><ng-content /></div>
```

## ¿Cero, una o dos veces?

<!--
COMPILA SIN UNA QUEJA, y aparece UNA SOLA VEZ. El segundo hueco queda vacío
para siempre.

POR QUÉ: el contenido proyectado SE MUEVE, no se copia. Son los mismos nodos
del DOM, que Angular saca del padre y mete en el hueco. Un nodo no puede estar
en dos lugares a la vez.

"No es que Angular decida no duplicarlo: es que NO HAY QUÉ DUPLICAR."

Mostrá la terminal primero, para que se vea que compiló.

Cierre del bloque: "¿cuál les habría costado más encontrar?" → el segundo.
-->

---

<!-- _class: bloque -->

# 1:25

## Misión 2

## En parejas

<!--
20 minutos. project/frontend/starter. Enunciado en mision-estudiante-2.md.

TRES COSAS ANTES DE LARGAR:
1. "El punto de partida es el listado que ustedes escribieron la clase pasada.
   Si les quedó a medias, la corrección de S1 lo deja andando en diez minutos."
2. "Son DOS componentes en lugares distintos. <app-badge> es una primitiva y va
   en shared/ui/ — no sabe qué es una carrera. <app-race-card> sí sabe, así que
   va en features/races/."
3. "Cuando muevan el marcado, MUEVAN EL CSS CON ÉL."
-->

---

## Misión 2 — el reparto

| Qué | Dónde | Por qué |
|---|---|---|
| `<app-badge>` | `shared/ui/` | no sabe qué es una carrera |
| `<app-race-card>` | `features/races/` | sí sabe |
| Cuál está abierta | **el listado** | solo puede haber una |
| Formatear la hora | **el listado** | la tarjeta no decide eso |

<!--
Dejala en pantalla los veinte minutos.

shared/ NUNCA importa de features/. Es la regla de dependencias del proyecto y
hoy es la mitad del ejercicio.

La pareja que termina recibe la extensión: usar race-card una segunda vez para
la carrera destacada, sin tocar el componente.
-->

---

<!-- _class: bloque -->

# 1:45

## Code review

<!--
10 minutos. Una solución de la Misión 2, con permiso. correccion.md al lado.

Rúbrica del curso, completa, en este orden:
1. standalone y OnPush
2. estado sin mutar
3. any, console.log, imports sin usar
4. el nombre dice lo que hace
5. está en la carpeta que le toca

EMPEZÁ POR ALGO QUE ESTÁ BIEN HECHO.

Los cinco errores de todos los años, con qué decirles, están en correccion.md.
-->

---

<!-- _class: ojo -->

# Tapá el archivo del padre

Leyendo **solo** el hijo, ¿podés decir qué hace y qué necesita?

<!--
LA PREGUNTA QUE HACE LA SESIÓN. Es un test que pueden correr solos, para
siempre.

Si la respuesta es sí, el corte está bien hecho. Si tenés que ir a mirar quién
lo usa, quedó atado a una sola pantalla.

Y el cierre:
"Se van a dar cuenta de que la pantalla se ve EXACTAMENTE IGUAL que la clase
pasada. Es correcto: hoy no agregamos ni una función. Lo que cambió es que
ahora hay una pieza que se puede usar en otro lado, y en la clase 10 el
leaderboard la va a usar sin tocarle una línea."
-->

---

<!-- _class: bloque -->

# 1:55

## Exit ticket

## y tarea

<!--
Exit ticket: tres preguntas, tres minutos. La tercera arranca S3.

Tarea: LEELA EN VOZ ALTA antes de cortar. El punto 1 —usar la tarjeta en otra
pantalla sin tocarla— es el examen de verdad de lo de hoy.

Y el apunte:
"conceptos.md tiene todo lo de hoy con los ejemplos exactos que corrimos, y los
tres errores con su mensaje literal."
-->

---

<!-- _class: portada -->

# Hasta la próxima

## S3 · Signals y control flow

<!--
El anzuelo, textual:
"La clase que viene: signals. Van a entender por qué coffee() se lee con
paréntesis, y vamos a hacer que el listado se filtre solo."

Y anotá, con la clase fresca, las cuatro preguntas del final del guión.
-->
