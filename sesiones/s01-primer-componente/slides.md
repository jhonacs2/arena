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
El guión completo está en guion.md. Las notas de cada diapositiva son el
guión de esa diapositiva: se ven con la tecla P en el HTML de Marp.
-->

---

<!-- _class: bloque -->

# 0:00

## Pregunta de apertura

<!--
5 minutos. Responden en el chat. Sin juicio, sin corregir, sin "casi".
Leer dos o tres en voz alta y seguir.
-->

---

## Pensá en una web que uses todos los días

Cuando cambia un número en pantalla —el carrito, un contador, un saldo—

# ¿quién lo cambió?

Escribilo en el chat. No hay respuesta correcta.

<!--
Va a haber de todo: "el servidor", "JavaScript", "React", "no sé".
TODAS SIRVEN. La pregunta no busca la respuesta: busca que noten que hay
alguien moviendo el DOM. Hoy ese alguien va a ser Angular.

Si nadie escribe a los 90 segundos, responder uno mismo para romper el hielo.
-->

---

<!-- _class: bloque -->

# 0:05

## Wayground

## TypeScript

<!--
sesiones/s00-typescript/wayground.csv — el material asíncrono.
Es la ÚNICA sesión donde el quiz no es de la anterior: no hay anterior.

30 segundos de explicación como máximo por pregunta que falle más de un
tercio. Si necesita más, va a la tarea. Los tres tropiezos esperables están
en el guión.
-->

---

<!-- _class: bloque -->

# 0:12

## El concepto

## Editor cerrado

<!-- A partir de acá, ni una línea de código en pantalla. Solo diagrama. -->

---

## De dónde viene esto

**2010 · AngularJS** — que el HTML pudiera tener lógica declarativa

**2016 · Angular 2** — reescritura. La idea de los *Web Components*: el navegador ya sabe de componentes

**2024 · standalone** — se va la capa de `NgModule`. El componente se declara solo

<!--
TRES MINUTOS. No más.

En la primera sesión hay que salir con algo funcionando. Quien escribió su
primer componente entiende para qué sirvió AngularJS mucho mejor que quien
escuchó veinte minutos de historia sin escribir nada.

Frase para dejar: "Angular no inventó los componentes. Los hizo tipados, con
herramientas y con un ciclo de vida."

Los NgModules vuelven en S11, con contexto, para poder leer código viejo.
-->

---

## Un componente son dos cosas

![w:860](diagramas/componente-y-template.svg)

<!--
CINCO MINUTOS sobre este diagrama.

La analogía: un mostrador de cafetería.
  · La clase es la trastienda: qué hay, cuánto cuesta, quién atiende.
  · El template es la vidriera.
  · Interpolación = escribir el precio en el cartel.
  · Property binding = apagar la luz del cartel cuando se acabó.
  · Event binding = el timbre que suena cuando alguien pide.
  · Two-way = la libreta del pedido: la escribe el cliente, la lee el mozo.

Que puedan dibujar las dos cajas y las cuatro flechas ANTES de escribir una
línea. Si arranca el código acá, copian sintaxis sin modelo mental.
-->

---

<!-- _class: ojo -->

# Ojo con esto

En Angular 18 hay que escribir

`standalone: true`

En 19 pasa a ser el valor por defecto

<!--
Es la trampa número uno al copiar código de un blog reciente.
Todo el material nuevo asume 19+; este proyecto es 18.
La sección 4 de CLAUDE.md tiene la lista completa de lo que NO existe todavía.
-->

---

<!-- _class: bloque -->

# 0:20

## Live coding

## Ustedes miran

<!--
DECIRLO EXPLÍCITO: "cierren el editor, esto lo hacemos juntos en 15 minutos
y después lo hacen ustedes".

Proyecto: lab/solution → ruta /s01
La tabla minuto a minuto está en el guión.

A las 0:37: romperlo a propósito sacando FormsModule. Leer el error en voz
alta ANTES de arreglarlo. Es el error más frecuente de la sesión: que lo vean
acá hace que a las 0:35 lo reconozcan solos.
-->

---

<!-- _class: codigo -->

## Un componente standalone

```ts
@Component({
  selector: 'app-s01',
  standalone: true,
  imports: [FormsModule],
  templateUrl: './s01.component.html',
})
export class S01Component {
  protected coffee = { name: 'Yirgacheffe', price: 42 };
}
```

<!--
"imports es de ESTE componente. Standalone quiere decir que se declara solo:
no hay ningún NgModule que lo haga por él."
-->

---

<!-- _class: codigo -->

## Los cuatro, en cuatro líneas

```html
<h2>{{ coffee.name }}</h2>

<div [class.agotado]="!coffee.available">…</div>

<button (click)="addOrder()">Pedir</button>

<input [(ngModel)]="customer" />
```

<!--
Uno por uno, con la pantalla del navegador al lado.
Cambiar el precio en el .ts en vivo y que vean la recarga.
-->

---

<!-- _class: bloque -->

# 0:35

## Misión 1

## Individual · 15 min

<!--
lab/starter → /s01. El enunciado está en mision-1.md.

ESTÁS EN SILENCIO. Disponible si preguntan, pero no circulás ofreciendo ayuda.
Los quince minutos de pelearse solo con el error SON la clase.

Si a los 8 minutos más de la mitad está trabada en lo mismo, una pista para
todos en voz alta, sin resolver:
"¿Alguien ya vio el error de ngModel? ¿Dónde declara un componente lo que usa?"
-->

---

## Misión 1 · El mostrador

Buscá `TODO(S1)` en `lab/starter`. Están numerados del 1 al 4.

1. **Interpolación** — nombre, origen y precio
2. **Property binding** — la clase `product--soldout`
3. **Event binding** — el botón de disponibilidad
4. **Two-way** — cliente y cantidad

**Listo cuando:** cambiar el precio en el `.ts` cambia lo que se ve, y el total se actualiza mientras escribís.

---

<!-- _class: bloque -->

# 0:50

## Comparten pantalla

<!--
Dos alumnos. Uno que funciona y uno que no — con permiso.

PREGUNTÁS, NO CORREGÍS. Aunque esté mal. Aunque duela.
  ¿Qué esperabas que pasara? · ¿Qué pasó? · ¿Cómo lo averiguaste?
  ¿Cómo se lo explicarías a alguien que no estuvo hoy?

Lo más probable: alguien escribió class="producto--agotado" en vez de
[class.producto--agotado]="…". Es la mejor pantalla posible para compartir.
-->

---

<!-- _class: bloque -->

# 1:00

## Descanso

## 10 minutos

<!-- Volver puntual: los quince minutos de después son los más densos. -->

---

<!-- _class: bloque -->

# 1:10

## Predice y ejecuta

<!--
Mostrar el código. NO EJECUTAR.
"¿Qué va a pasar? Escribilo en el chat." — 60 segundos.
Recién ahí ejecutar.

El paso de predecir es TODO el ejercicio. Ejecutar primero lo convierte en
una demo y no aprende nadie.

Los tres están verificados en el navegador. Las respuestas, en
predice-y-ejecuta/respuestas.md.
-->

---

<!-- _class: codigo -->

## 1 · ¿Qué clases quedan?

```html
<p class="label {{ tone }}"
   [class.etiqueta--activa]="activo">Hola</p>
```

```ts
tone = 'rojo';
activo = true;
```

<!--
No ejecutar. 60 segundos de predicciones.

Casi todos dicen que se pisan. NO: quedan las TRES.
  class="etiqueta rojo etiqueta--activa"

class y [class.x] no compiten: se combinan. El binding específico es dueño de
SU clase y nada más. Por eso en la Misión 2 van a usar las dos juntas.
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

Leer el mensaje completo en voz alta. Es de los mejores que da Angular.
-->

---

<!-- _class: codigo -->

## 3 · ¿Cuánto marca después de 3 clics?

```html
<button (click)="contador + 1">Sumar</button>
<p>Llevo {{ contador }} clics</p>
```

<!--
Se queda en 0. Y NO HAY NINGÚN ERROR — ni en consola, ni en el build.

La expresión se evalúa y el resultado se tira. Nadie lo asignó.

La regla: un binding que no falla no quiere decir que ande. El silencio es el
peor síntoma que puede tener un bug.

Si sobra tiempo, preguntar: "de los tres, ¿cuál les habría costado más
encontrar en un proyecto de verdad?". Casi siempre eligen este, y tienen razón.
-->

---

<!-- _class: bloque -->

# 1:25

## Misión 2

## Al hipódromo

<!--
En parejas, 20 min, project/frontend/starter.
Conducción por turnos: 10 min escribe uno y dicta el otro, después se invierte.

DOS COSAS PARA DECIR ANTES DE LARGAR:
  · El @for se los damos hecho. El control flow es S3; hoy son los bindings.
  · El starter YA FUNCIONA A MEDIAS. No arrancan de una hoja en blanco:
    arrancan de algo que anda mal, que es más parecido al trabajo real.

Circular. Escuchar más que hablar. La pareja que termina recibe la extensión.
-->

---

## Misión 2 · El programa de carreras

Las ocho carreras ya están. Falta conectarlas.

- **Interpolación** — estado, nombre, hora, competidores
- **Property binding** — `race--live`, `race--finished`, `race--open`
- **Event binding** — `(click)` → `select(view)`
- **Two-way** — el monto del simulador

Y en el `.ts`: `payout` devuelve siempre 0, y `select()` nunca deselecciona.

<!--
Si alguien se queja del orden de la lista: BIEN. Que les moleste. Ordenar y
filtrar es S3. Hoy se pintan los datos como vienen.
-->

---

<!-- _class: bloque -->

# 1:45

## Code review en vivo

<!--
Una solución de la Misión 2, con permiso. Rúbrica de docs/curriculum.md,
en voz alta y en orden.

En S1 el punto 4 —cargando/vacío/error— casi no aplica, y conviene decirlo:
"hoy no hay estado de carga porque no hay nada que cargar. En S7 vuelve, y
va a ser obligatorio."

Empezar por algo que está bien hecho. Siempre hay algo.
-->

---

<!-- _class: bloque -->

# 1:55

## Exit ticket

<!--
Tres preguntas, 3 minutos.
La tercera —"¿qué quedó confuso?"— arranca S2.

Leer la tarea en voz alta antes de cortar.
-->

---

<!-- _class: portada -->

# Nos vemos en S2

## Tarea en `tarea.md`
