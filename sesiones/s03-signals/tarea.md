# S3 · Tarea asíncrona

**Entrega antes de S4.** Se lee en voz alta en clase antes de cortar.

---

## Qué hacer

Terminar la **Misión 2** si te quedó a medias, y después dos cosas que solo se
pueden hacer bien si el estado quedó ordenado.

### 1 · Ordenar el programa

Agregá dos botones al listado: **por hora de largada** y **por cuota del
favorito**.

- El criterio elegido es **un signal más**.
- La lista ordenada es **un `computed` más**.
- El orden del dataset no cambia. Comprobalo: ordená por cuota, recargá, y fijate
  que el programa vuelve a estar como estaba.

Si te sale agregando un tercer signal que guarda la lista ya ordenada, borralo y
volvé a intentarlo: ese es exactamente el hábito que la clase vino a sacar.

### 2 · Romperlo a propósito, y describir el síntoma

En una rama o en un comentario, cambiá el `computed` de la parrilla por la
versión sin la copia:

```ts
return this.selected()?.race.horses.sort((a, b) => a.odds - b.odds) ?? [];
```

Abrí una carrera, cerrala, y abrí otra. Después contestá por escrito:

1. ¿Qué se rompió, exactamente? Describí lo que ve el usuario, no lo que dice el
   código.
2. ¿Por qué no hay ningún error en la consola?
3. ¿Qué otra pantalla del proyecto podría verse afectada por esto, aunque nadie
   la haya tocado?

**Después volvé a poner el `[...]`.**

### 3 · Contestá una

En un comentario al final de `race-list.component.ts`:

> Las ocho carreras están en una constante y no son un signal. En la sesión 7 van
> a venir de un servidor. **¿Qué líneas de este archivo van a tener que cambiar
> ese día, y cuáles no?**

## Listo cuando

- [ ] Los dos órdenes funcionan y el dataset no queda reordenado
- [ ] Hay **un** signal nuevo y **un** computed nuevo, no más
- [ ] Las tres preguntas del punto 2 están contestadas describiendo síntomas
- [ ] La pregunta del punto 3 está contestada
- [ ] `npm run build` pasa
- [ ] Commiteado: `feat(s03): orden configurable del programa`

## Cuánto lleva

**30–45 minutos.**

## Material de apoyo

- Angular · *Signals*: <https://angular.dev/guide/signals>
- Angular · *Control flow*: <https://angular.dev/guide/templates/control-flow>
- MDN · métodos de array que **mutan** y que no:
  <https://developer.mozilla.org/es/docs/Web/JavaScript/Reference/Global_Objects/Array>

> Ojo con la documentación oficial: está escrita para la última versión. Si ves
> `linkedSignal()` o `resource()`, son de v19 y v20 y **no existen en 18**. La
> lista completa está en el `CLAUDE.md` del repo.

---

## Para el instructor

**Lo que más va a aparecer:**

- **Un tercer signal con la lista ordenada**, actualizado a mano en el `set` del
  criterio. Funciona, y es el hábito viejo con ropa nueva. Cinco minutos al
  empezar S4.
- **La respuesta al 2.1 en abstracto** («se rompe el orden»). La que buscamos es
  el síntoma: *abrís una carrera, la cerrás, abrís otra, y la primera quedó con
  los caballos en otro orden para siempre*.
- **La respuesta al 3 subestimando cuánto no cambia.** Casi todo el archivo
  sobrevive: cambia la línea de `all` y aparece un estado de carga. Que lo vean
  ahora hace que S7 se sienta corta.
