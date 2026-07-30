# S6 · Tarea asíncrona

**Entrega antes de S7.** Se lee en voz alta en clase antes de cortar.

---

## Qué hacer

Terminar la **Misión 2** si quedó a medias, y después tres cosas que preparan la
clase que viene.

### 1 · Romperlo a propósito y describir el síntoma

En una rama o en un comentario, haz los tres cambios de a uno, y describe **lo que
ve el usuario** —no lo que dice el código—:

1. `switchMap` → `mergeMap`
2. `catchError` afuera del `switchMap`
3. Quitar `takeUntilDestroyed()`

Para el tercero, la forma de verlo: navega a otra sesión del lab y vuelve varias
veces mientras escribes. ¿Qué pasa con el contador de búsquedas?

**Después vuelve a dejarlo bien.**

### 2 · Los tres estados, de verdad

En el buscador del lab, mejora los tres estados:

- **Cargando**: que el esqueleto aparezca solo si la búsqueda tarda más de 150 ms.
  Un parpadeo gris en cada búsqueda rápida es peor que nada.
- **Vacío**: que el mensaje diga qué se buscó — `Ningún café coincide con
  "kenia"`.
- **Error**: un botón «Reintentar» que vuelva a lanzar la misma búsqueda.

El primero tiene truco y vale la pena pelearlo: hay más de una forma y ninguna es
obvia.

### 3 · Contesta una

En un comentario al final de `race-list.component.ts`:

> En la sesión 7 la búsqueda va a ser una petición al servidor de verdad.
> **¿Qué líneas de este archivo van a cambiar ese día?**

Piénsalo mirando el código antes de contestar. La respuesta es incómoda de creer.

## Listo cuando

- [ ] Los tres síntomas están descritos en términos de lo que ve el usuario
- [ ] El código quedó bien de nuevo
- [ ] Los tres estados están mejorados y el esqueleto no parpadea
- [ ] La pregunta del punto 3 está contestada
- [ ] `npm run build` y `npm test` pasan
- [ ] Commiteado: `feat(s06): buscador con debounce`

## Cuánto lleva

**40–60 minutos.** El punto 2 es el que puede estirarse; si te pasas de una hora,
para y anota dónde te trabaste.

## Material de apoyo

- Angular · *RxJS interop*: <https://angular.dev/ecosystem/rxjs-interop>
- RxJS · *Operator decision tree*: <https://rxjs.dev/operator-decision-tree>
- RxJS · `switchMap`: <https://rxjs.dev/api/operators/switchMap>

> Ojo con los tutoriales de RxJS: casi todos son anteriores a los signals y
> resuelven con observables cosas que hoy son un `computed`. Si algo se puede
> hacer con un signal, se hace con un signal.

---

## Para el instructor

**Lo que más va a aparecer:**

- **El síntoma del 1.3 descrito como «hay una fuga de memoria».** Es correcto y
  no es lo que se pedía: lo que **se ve** es que el contador sigue subiendo por
  búsquedas de un componente que ya no está en pantalla.
- **El esqueleto con 150 ms resuelto con un `setTimeout`.** Funciona. Vale
  mostrar la versión con RxJS al empezar S7 — es un `timer` y un `takeUntil`, y
  es tres líneas más corta.
- **La respuesta al 3 enumerando cambios que no hacen falta.** La respuesta es
  **ninguna**: el componente empuja textos a un flujo y el store contesta. De
  dónde saca el store las carreras no le importa. Que lo comprueben ahora hace
  que S7 se sienta corta, que es exactamente el objetivo del orden del curso.
