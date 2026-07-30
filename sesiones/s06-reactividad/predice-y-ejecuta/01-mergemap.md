# 1 · `mergeMap` en vez de `switchMap`

**No ejecutes todavía.** Mira el código y escribe en el chat qué va a pasar.

---

```ts
private readonly results$ = this.terms.pipe(
  debounceTime(300),
  distinctUntilChanged(),
  mergeMap((term) => this.catalog.searchCounted(term)),
);
```

El catálogo tarda **1200 ms** con textos de una sola letra y **300 ms** con
textos más largos.

### La secuencia

1. Escribo `e` y espero a que salga la búsqueda.
2. Antes de que conteste, escribo `huila`.

### La pregunta

**¿Qué se ve al final, cuando ya no queda nada por llegar?**

1. Los resultados de `huila`
2. Los resultados de `e`
3. Los dos mezclados
4. Un error

Escribe tu predicción en el chat antes de que ejecutemos.

---

## Cómo reproducirlo en clase

En `lab/demo/src/app/sessions/s06/s06.component.ts`, cambia `switchMap` por
`mergeMap`.

**Hazlo despacio y en voz alta:** escribe `e`, cuenta hasta uno, escribe `huila`,
y **no toques nada más**. Lo que importa pasa un segundo después.

Después se repone el `switchMap`.
