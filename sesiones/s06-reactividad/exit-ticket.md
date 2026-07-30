# S6 · Exit ticket

**3 minutos. Anónimo si prefieres.**

---

**1. Recordar** — Une cada problema con la línea que lo resuelve:

| Problema | Operador |
|---|---|
| Sale una búsqueda por cada tecla | |
| Llega una respuesta vieja y pisa a la nueva | |
| La suscripción sigue viva después de salir de la pantalla | |
| Escribir lo mismo dos veces vuelve a buscar | |

**2. Aplicar** — Este buscador deja de funcionar para siempre después del primer
error de red. Escribe la versión corregida:

```ts
this.terms.pipe(
  debounceTime(300),
  switchMap((term) => this.api.search(term)),
  catchError(() => of([])),
);
```

**3. ¿Qué quedó confuso?**

---

## Para el instructor

**Respuestas de la 1:** `debounceTime` · `switchMap` · `takeUntilDestroyed` ·
`distinctUntilChanged`.

**Respuesta de la 2:** el `catchError` va **adentro** del `switchMap`:

```ts
switchMap((term) =>
  this.api.search(term).pipe(catchError(() => of([]))),
),
```

Y la segunda mitad —el porqué— es la que separa: **un error en un observable es
terminal**. Afuera mata el flujo entero; adentro, solo esa búsqueda.

| Lo que escriben | Qué decirles |
|---|---|
| Agregan `retry()` | Reintenta, y cuando se agoten los intentos el flujo muere igual. No es la respuesta. |
| Ponen `catchError` en los dos lados | Funciona. Vale preguntar cuál de los dos se ejecutó. |
| Cambian `switchMap` por `mergeMap` | No tiene nada que ver, y trae el bug del primer ejercicio de predicción. |

| Señal | Qué hacer |
|---|---|
| Más de un tercio deja el `catchError` afuera | Abre S7 con eso: ahí los errores son de verdad |
| Aparece «no entendí `switchMap`» | Volver al caso de las dos latencias, 3 minutos |
| Nadie confundido | Subir el techo de la Misión 2 la próxima |
