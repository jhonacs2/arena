# 2 · El `catchError` un renglón más arriba

**No ejecutes todavía.** Mira el código y escribe en el chat qué va a pasar.

---

```ts
private readonly results$ = this.terms.pipe(
  debounceTime(300),
  distinctUntilChanged(),
  switchMap((term) => this.catalog.searchCounted(term)),
  catchError(() => {
    this.status.set('error');
    return of([] as readonly Coffee[]);
  }),
);
```

El `catchError` está **afuera** del `switchMap`, no adentro. Todo lo demás es
igual.

### La secuencia

1. Escribo `error`. Aparece el mensaje de error, como debe ser.
2. Escribo `huila`.

### La pregunta

**¿Qué pasa en el paso 2?**

1. Aparecen los resultados de `huila`
2. Sigue el mensaje de error
3. No pasa nada: la pantalla se queda como estaba

Escribe tu predicción en el chat antes de que ejecutemos.

---

## Cómo reproducirlo en clase

En `lab/demo/src/app/sessions/s06/s06.component.ts`, saca el `catchError` del
`pipe()` interno y ponlo en el externo.

**Deja la consola del navegador abierta**: que se vea que no hay ningún error
nuevo cuando el buscador deja de responder es la mitad del ejercicio.

Después se repone adentro.
