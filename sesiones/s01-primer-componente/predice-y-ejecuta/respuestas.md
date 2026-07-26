# S1 · Predice y ejecuta — respuestas

**Solo para el instructor.** No va al repo del alumno.

El orden importa y no se saltea ningún paso:

1. Mostrar el código. **No ejecutar.**
2. *«¿Qué va a pasar? Escribilo en el chat.»* — 60 segundos.
3. Ejecutar.
4. Explicar la diferencia entre lo que dijeron y lo que pasó.

El paso 2 es todo el ejercicio. Ejecutar primero lo convierte en una demo y no aprende nadie.

> Los tres se verificaron en el navegador. El resultado del 1 no es el que uno supone de memoria.

---

## 1 · `01-class-vs-binding.md`

**Qué está roto:** nada. Ese es el punto.

**Qué predice casi todo el curso:** que se pisan. «Gana el `[class.…]`», «gana el `class`», «se rompe».

**Qué pasa de verdad:** quedan **las tres clases**. El atributo renderizado es exactamente:

```
class="etiqueta rojo etiqueta--activa"
```

**Por qué:** `class="…"` y `[class.x]="…"` **no compiten, se combinan**. Angular trata el binding específico —`[class.x]`— como dueño de esa clase y nada más; el resto del atributo `class` queda intacto. Son dos mecanismos que escriben en lugares distintos del mismo atributo.

**La regla que se llevan:** *podés usar `class` y `[class.x]` en el mismo elemento sin miedo.* Es exactamente lo que van a hacer en la Misión 2: `class="carrera"` fija, más `[class.carrera--viva]` condicional.

> Si alguien pregunta por `[class]="objeto"` —la forma que reemplaza todo el conjunto—: existe, y ahí sí hay reglas de precedencia. Se ve en S4 con directivas. Hoy no.

---

## 2 · `02-ngmodel-sin-formsmodule.md`

**Qué está roto:** falta `FormsModule` en los `imports` del componente.

**Qué predice casi todo el curso:** la 1 o la 2 — «anda», o «no se actualiza pero la app funciona». Casi nadie dice que no compila.

**Qué pasa de verdad:** la **4**. `ng serve` no compila y dice:

```
NG8002: Can't bind to 'ngModel' since it isn't a known property of 'input'.
```

**Por qué:** `ngModel` es una directiva, no algo del navegador. Angular solo reconoce las directivas que el componente **declaró** en sus `imports`. Sin declararla, `[(ngModel)]` es un intento de escribir una propiedad `ngModel` en un `<input>`, y esa propiedad no existe.

Esto es lo que quiere decir **standalone**: el componente se declara solo, sin un `NgModule` que lo haga por él. Es más código para escribir y muchísimo menos código para buscar cuando algo no anda.

**La regla que se llevan:** *si el template usa algo de Angular, tiene que estar en `imports`.* Vale para `FormsModule`, `RouterLink`, y para cada componente que usen adentro de otro a partir de S2.

> Vale la pena leer el mensaje de error completo en voz alta. Es de los mejores que da Angular, y aprender a leerlos ahorra más tiempo que cualquier atajo.

---

## 3 · `03-expresion-sin-asignar.md`

**Qué está roto:** `(click)="contador + 1"` calcula, pero no guarda.

**Qué predice casi todo el curso:** que suma. La expresión *se parece* a algo que suma.

**Qué pasa de verdad:** el contador se queda en **0**, para siempre. Y —esto es lo importante— **no hay ningún error**. Ni en la consola, ni en el build.

**Por qué:** lo que va entre comillas en un event binding es una **expresión que Angular evalúa**. `contador + 1` se evalúa, da `1`, y el resultado se descarta. Nadie lo asignó a nada. Para que cambie el estado hay que asignarlo: `contador = contador + 1`, o llamar a un método que lo haga.

**La regla que se llevan:** *un binding que no falla no quiere decir que ande.* El silencio es el peor síntoma que puede tener un bug, y este es el primero de muchos que van a ver.

> Es también el mejor argumento a favor de poner la lógica en métodos de la clase en vez de en el template: un método se puede leer, testear y depurar. Una expresión en el HTML, no.

---

## Si sobra tiempo

Preguntar: *«de los tres, ¿cuál les habría costado más encontrar en un proyecto de verdad?»*

Casi siempre eligen el 3, y tienen razón: los otros dos te frenan el build. Ese es el punto de todo el bloque.
