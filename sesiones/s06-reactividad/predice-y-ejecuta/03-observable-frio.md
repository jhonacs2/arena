# 3 · Una búsqueda que nadie mira

**No ejecutes todavía.** Mira el código y escribe en el chat qué va a pasar.

---

```ts
protected precargar(): void {
  this.catalog.searchCounted('etiopía');
}
```

El método se llama al tocar un botón. `searchCounted` es exactamente el mismo que
usa el buscador, y suma uno al contador de búsquedas **cada vez que alguien se
suscribe**.

### La pregunta

**Toco el botón cinco veces. ¿Cuánto marca el contador?**

1. `5`
2. `1`
3. `0`

Escribe tu predicción en el chat antes de que ejecutemos.

---

## Cómo reproducirlo en clase

En `lab/demo/src/app/sessions/s06/s06.component.ts`, agrega el método de arriba y
un botón que lo llame.

Toca el botón varias veces y señala el contador. Después, en la misma línea,
agrega `.subscribe()` al final y vuelve a tocarlo.

Después se borra el botón.

**La pregunta de yapa, si sobra un minuto:**

> «¿Y por qué entonces `HttpClient` a veces "funciona sin suscribirse"?»

Nunca funciona sin suscribirse. Lo que pasa es que `AsyncPipe`, `toSignal` y
`firstValueFrom` **se suscriben por ti**, y es fácil no darse cuenta.
