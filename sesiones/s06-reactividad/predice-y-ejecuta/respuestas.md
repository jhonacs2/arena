# S6 · Predice y ejecuta — respuestas

> **Verificado con Angular 18.2.14 y RxJS 7.8**, adelantando el reloj con
> `fakeAsync` y `tick` en un test de Karma. Los tres resultados están medidos, no
> razonados. Si cambias un snippet, vuelve a correrlo antes de dar la clase.

**El orden no se saltea:** mostrar → predecir 60 segundos → ejecutar → explicar.

---

## 1 · `mergeMap` en vez de `switchMap`

**Opción 2: al final se ven los resultados de `e`.**

Lo que se midió, paso a paso:

| Momento | Qué se ve |
|---|---|
| A los 300 ms — contesta `huila` | `Huila` |
| A los 1500 ms — contesta `e` | `Yirgacheffe, Cerrado, Antigua, Sidamo, Kiambu` |

**La pantalla termina mostrando los resultados de una búsqueda que el usuario ya
había abandonado**, y el campo de texto dice `huila`.

### Por qué

`mergeMap` deja correr todas las búsquedas a la vez y muestra cada respuesta a
medida que llega. **El orden de llegada no es el orden de salida:** la búsqueda de
una letra tarda 1200 ms y la de cinco tarda 300.

`switchMap` **cancela** la anterior cada vez que llega un valor nuevo. La de `e`
nunca llega, porque se cortó.

### La regla

| Si… | Va |
|---|---|
| solo importa el último | **`switchMap`** |
| importan todos | `mergeMap` — subir cinco archivos, registrar cinco eventos |

En un buscador, en un autocompletado y en cualquier cosa que dependa de lo último
que escribió una persona, es `switchMap`.

### Y por qué casi nadie lo ve venir

No hay ningún error. La pantalla muestra datos reales, solo que de otra pregunta.
**Y con red rápida no pasa nunca:** aparece con la red que tienen los usuarios y
no tenemos nosotros.

---

## 2 · El `catchError` un renglón más arriba

**Opción 3: no pasa nada. La pantalla se queda como estaba, para siempre.**

Lo que se midió: después del error, el flujo emitió **una** vez. Después de
escribir `huila`, siguió emitiendo… **una**. No volvió a emitir nunca.

### Por qué

**Un error en un observable es terminal:** el flujo se acaba, como si hubiera
terminado. `catchError` no lo revive; atrapa el error y decide con qué **seguir**,
pero lo que sigue reemplaza al flujo que murió.

| Dónde está | Qué muere |
|---|---|
| **adentro** del `switchMap` | solo esa búsqueda. La siguiente arranca sana |
| **afuera** | el flujo entero. **El buscador no responde nunca más** |

Adentro, el `catchError` está en la tubería de la búsqueda individual: esa es la
que muere, y el `Subject` de arriba sigue vivo esperando la próxima tecla.

### Y por eso es el peor de los tres

No hay error en la consola. El usuario escribe, no pasa nada, y no hay nada que
mirar. La única pista es que dejó de responder **después** de un problema de red
— y para cuando alguien lo reporta, ya recargó la página y no se reproduce.

---

## 3 · Una búsqueda que nadie mira

**Opción 3: el contador marca `0`.**

Cinco llamadas a `searchCounted('etiopía')`, y el contador no se movió ni una vez.

### Por qué

> **Un observable es frío: no pasa nada hasta que alguien se suscribe.**

`searchCounted()` no busca: **devuelve una receta**. La búsqueda ocurre en el
momento en que alguien se suscribe, y cada suscripción ejecuta la receta de nuevo,
desde cero.

Agregando `.subscribe()` al final de la línea, el contador pasa a 5.

### La diferencia con una promesa

```js
fetch('/api/coffees');        // ya salió, aunque nadie mire
this.catalog.search('x');     // no salió nada
```

Una promesa empieza al crearla. Un observable, al suscribirse. Es la trampa más
común al empezar con RxJS, y no da ningún error: se escribe la llamada, no pasa
nada, y no hay nada que leer.

### La pregunta de yapa

> «¿Y por qué entonces `HttpClient` a veces "funciona sin suscribirse"?»

**Nunca funciona sin suscribirse.** Lo que pasa es que `AsyncPipe`, `toSignal()` y
`firstValueFrom()` **se suscriben por ti**, y es fácil no darse cuenta de que ahí
hubo una suscripción.

---

## La pregunta de cierre del bloque

> «¿Qué tienen en común los tres?»

> «Los tres son **creer que el tiempo no existe**: que las respuestas llegan en
> orden, que un error es un episodio, y que escribir una llamada es hacerla.»
>
> «Y los tres pasan el build, pasan la revisión y solo se ven con red lenta.»
