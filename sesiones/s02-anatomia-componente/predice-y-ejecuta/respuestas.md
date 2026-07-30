# S2 · Predice y ejecuta — respuestas

> **Verificado con Angular 18.2.14 y el `tsconfig.json` del curso**, con
> `strictTemplates`. Los mensajes están copiados de la salida del compilador y
> el comportamiento del tercero se midió contando elementos en el DOM. Si
> cambiás un snippet, volvé a correrlo antes de dar la clase.

**El orden no se saltea:** mostrar → predecir 60 segundos → ejecutar → explicar.

---

## 1 · Una tarjeta sin café

**No compila.**

```
NG8008: Required input 'coffee' from component CoffeeCardComponent must be specified.
```

### Por qué

`input.required()` no es una anotación amable: es una exigencia que verifica el
compilador, igual que los tipos de S0.

Y la comparación que vale la pena hacer en voz alta:

| | Qué pasa si falta |
|---|---|
| `input.required<Coffee>()` | **error de compilación**, con archivo y línea |
| `input<Coffee \| undefined>(undefined)` | compila, y el hijo se rompe en tiempo de ejecución |

> «Es la misma decisión de la sesión 0: o el tipo dice la verdad y el compilador
> te frena, o vos prometés algo y lo descubrís con un usuario adelante.»

### La frase para cerrar

> «`required` no documenta. **Exige.**»

---

## 2 · Un aviso que nadie escucha

**Opción 3: compila, y al tocar no pasa absolutamente nada.**

Sin error de compilación. Sin advertencia en la terminal. **Sin una sola línea en
la consola del navegador.** El botón se aprieta, se hunde, y no ocurre nada.

### Por qué

Un `output()` es un emisor. Emite exista o no alguien del otro lado — igual que
un `addEventListener` que nadie registró, salvo que acá ni siquiera hay un
listener que falte: el padre simplemente no escribió el binding.

Angular no puede avisar de esto, y no es un descuido: **no tiene forma de saber
si un `output()` que nadie escucha es un olvido o es a propósito.** Un componente
reutilizable puede tener seis salidas de las que una pantalla usa dos.

### El primo de este bug

```html
<button (click)="contador + 1">   <!-- S1 -->
```

Es la misma familia: compila, se ejecuta, no hace nada, no avisa.

> **Un binding que no falla no quiere decir que ande.** El silencio es el peor
> síntoma que puede tener un bug.

### Cómo se encuentra en la vida real

Empezando por el hijo, no por el padre: `emit()` está ahí, así que el problema
está del otro lado. En el editor, buscar el nombre del output — si aparece una
sola vez en todo el proyecto, nadie lo escucha.

---

## 3 · Dos huecos iguales

**Compila sin una queja, y «Vuelve el jueves.» aparece UNA sola vez.**

El segundo `<ng-content>` queda vacío para siempre.

### Por qué

El contenido que el padre proyecta **se mueve**, no se copia: son los mismos
nodos del DOM, que Angular saca del padre y mete en el hueco del hijo. Un nodo
del DOM no puede estar en dos lugares a la vez, así que el segundo hueco se queda
sin nada.

> «No es que Angular decida no duplicarlo: es que **no hay qué duplicar**. Es un
> solo `<p>`, y ya está adentro del primer hueco.»

### La regla

| | |
|---|---|
| `<ng-content>` **sin `select`** | va **uno solo** por componente |
| Dos lugares distintos | se distinguen con `select` |

```html
<ng-content select="[card-tag]" />   <!-- arriba -->
<ng-content />                        <!-- todo lo demás -->
```

### Por qué no da error

Es una decisión discutible de Angular y conviene decirlo así: hay casos raros
—`ng-content` adentro de un `@if`— donde tener dos tiene sentido, y el compilador
no distingue ese caso del error. Así que no avisa.

---

## La pregunta de cierre del bloque

> «De los tres, ¿cuál les habría costado más encontrar en un proyecto de verdad?»

Casi siempre eligen el segundo, y tienen razón:

> «El primero te frena el build: lo arreglás en diez segundos. El tercero se ve
> raro en la pantalla, y mirando lo encontrás. **El segundo no se ve por ningún
> lado**: no hay error, no hay log, el botón está y se puede tocar.»
>
> «Guárdense la sensación, porque vuelve dos veces más: en la clase 6 con una
> suscripción que nadie hizo, y en la clase 10 con un evento que llega y no
> repinta nada.»
