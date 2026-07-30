# S4 · Tarea asíncrona

**Entrega antes de S5.** Se lee en voz alta en clase antes de cortar.

---

## Qué hacer

Terminar la **Misión 2** si quedó a medias, y después dos cosas que solo salen
bien si los pipes quedaron donde corresponde.

### 1 · Usar los pipes en otra pantalla

En `features/sistema`, la muestra del sistema de diseño, agrega una sección
«Formatos» que muestre:

- Tres importes con `| money`, uno de ellos sin unidad.
- Tres cuotas con `| odds`, incluyendo una entera como `9`.
- Una fila de ejemplo con la directiva `[appFavourite]` puesta y otra sin ella.

**No se toca ningún pipe.** Si tienes que tocarlos, anota qué te faltó: eso es el
material del bloque de las 0:05 de la próxima.

### 2 · Un pipe con una decisión difícil

Escribe `relativeTime`, que muestre cuánto falta para que largue una carrera:

```
{{ race.startsAt | relativeTime }}   → en 8 min · hace 2 h · largando
```

Y después contesta, por escrito, la parte que importa:

1. El valor de entrada es la fecha de largada, y **esa fecha no cambia nunca**.
   Pero lo que hay que mostrar sí cambia, porque pasa el tiempo. ¿Qué pasa si el
   pipe es puro?
2. ¿Y si lo haces impuro? ¿Cuántas veces por segundo corre? ¿Qué pasa si hay
   cuarenta carreras en pantalla?
3. La respuesta honesta es que **un pipe no es la herramienta correcta para
   esto**. ¿Con qué de lo que ya viste se resolvería mejor?

No hace falta implementar la solución buena: hace falta poder explicar por qué el
pipe no alcanza.

### 3 · Contesta una

En un comentario al final de `odds.pipe.ts`:

> `toFixed(2)` estuvo mostrando un punto decimal en una aplicación en español
> desde la primera clase, y nadie lo notó. **¿Por qué este tipo de error es tan
> difícil de ver, y qué otro lugar del proyecto podría tener el mismo problema
> ahora mismo?**

## Listo cuando

- [ ] La sección «Formatos» funciona y **ningún pipe se tocó**
- [ ] `relativeTime` compila y muestra los tres casos
- [ ] Las tres preguntas del punto 2 están contestadas
- [ ] La pregunta del punto 3 está contestada
- [ ] `npm run build` pasa
- [ ] Commiteado: `feat(s04): pipes de formato y directiva de favorito`

## Cuánto lleva

**30–45 minutos.**

## Material de apoyo

- Angular · *Pipes*: <https://angular.dev/guide/pipes>
- Angular · *Attribute directives*: <https://angular.dev/guide/directives/attribute-directives>
- MDN · `Intl.NumberFormat`: <https://developer.mozilla.org/es/docs/Web/JavaScript/Reference/Global_Objects/Intl/NumberFormat>

> Ojo con la documentación oficial: está escrita para la última versión, donde
> `standalone: true` ya no se escribe. En 18 se escribe siempre.

---

## Para el instructor

**Lo que más va a aparecer:**

- **Tocaron el pipe para la sección de formatos**, casi siempre agregándole un
  caso especial. Es la señal de que el pipe recibió el objeto en vez del valor.
- **`relativeTime` impuro y sin más análisis.** La respuesta que buscamos al 2.3
  es un `computed` que dependa de un signal de tiempo, o directamente calcularlo
  al preparar los datos. La conversación completa es de S6.
- **La respuesta al 3 en abstracto.** La que buscamos es concreta: es difícil de
  ver porque **no falla** —se ve un carácter distinto, y hay que saber que está
  mal—. Y el otro lugar es `formatTime()` de `race-list`, que sí usa `Intl` con
  `'es'`, pero cualquier `| date` que se agregue sin `LOCALE_ID` volvería a
  romperlo.
