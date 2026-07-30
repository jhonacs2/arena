# 2 · Un aviso que nadie escucha

**No ejecutes todavía.** Mirá el código y escribí en el chat qué va a pasar.

---

El hijo emite, como siempre:

```ts
readonly ordered = output<OrderRequest>();

protected order(): void {
  this.ordered.emit({ coffee: this.coffee(), quantity: this.quantity() });
}
```

El padre lo usa **sin escuchar la salida**:

```html
<app-coffee-card [coffee]="item.coffee" [(quantity)]="item.quantity" />
```

### La pregunta

**Toco «Pedir». ¿Qué pasa?**

Tres opciones para elegir:

1. No compila.
2. Compila, y al tocar salta un error en la consola.
3. Compila, y al tocar no pasa absolutamente nada.

Escribí tu predicción en el chat antes de que ejecutemos.

---

## Cómo reproducirlo en clase

En `lab/demo/src/app/sessions/s02/s02.component.html`, borrá el
`(ordered)="take($event)"` del `<app-coffee-card>`.

**Abrí la consola del navegador antes de tocar el botón**, y dejala a la vista:
que se vea que está vacía es la mitad del ejercicio.

Después de explicar, se repone.
