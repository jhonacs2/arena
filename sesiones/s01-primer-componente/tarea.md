# S1 · Tarea asíncrona

**Entrega antes de S2.** Se lee en voz alta en clase antes de cortar: una tarea que solo se manda por chat no se hace.

---

## Qué hacer

Terminar la **Misión 2** si te quedó a medias, y sumarle una cosa: que cada carrera muestre **cuánto falta para que largue**.

- Una carrera que ya pasó: «hace 2 h»
- Una que falta: «en 8 min»
- Una que está corriendo: «largando»

## Dónde

`project/frontend/starter/src/app/features/carreras/race-list.component.ts`

Todo el cálculo va **en la clase**, no en el template. Agregá un campo a `RaceView` y llenalo donde se arma la lista.

## Listo cuando

- [ ] Las ocho carreras muestran su tiempo relativo
- [ ] La que está en vivo dice «largando», no un número raro
- [ ] `npm run build` pasa sin errores
- [ ] Está commiteado con el prefijo de la sesión: `feat(s01): tiempo relativo en el listado`

## Cuánto lleva

**30–45 minutos.** Si te lleva más de una hora, parás y anotás dónde te trabaste — eso es material para el bloque de las 0:05 de la próxima.

## Pistas

`Date.now()` da los milisegundos de ahora. `new Date(iso).getTime()` da los de una fecha. La resta es la diferencia en milisegundos, y de ahí a minutos y horas es aritmética.

No hace falta ninguna librería. Si te tienta instalar una, ese es el momento de parar: en S4 vas a ver la forma que Angular tiene para esto.

## Material de apoyo

- Angular · *Template syntax*: <https://angular.dev/guide/templates>
- MDN · `Date`: <https://developer.mozilla.org/es/docs/Web/JavaScript/Reference/Global_Objects/Date>

---

## Para el instructor

No se corrige una por una. Se revisa una al azar en el code review de S2, y lo que aparezca repetido en varias entregas se convierte en pregunta del `wayground.csv`.

**Lo que más va a aparecer:** el cálculo hecho en el template con un método —`{{ tiempoRelativo(race) }}`—. No está *mal*, pero se recalcula en cada detección de cambios. Es la puerta de entrada perfecta a `computed()` en S3.
