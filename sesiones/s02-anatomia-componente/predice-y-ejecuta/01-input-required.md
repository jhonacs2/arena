# 1 · Una tarjeta sin café

**No ejecutes todavía.** Mirá el código y escribí en el chat qué va a pasar.

---

El hijo:

```ts
@Component({ selector: 'app-coffee-card', standalone: true, /* … */ })
export class CoffeeCardComponent {
  readonly coffee = input.required<Coffee>();
}
```

El padre:

```html
<app-coffee-card />
```

### La pregunta

**¿Qué pasa? ¿Anda con el café en `undefined`, se ve una tarjeta vacía, o algo
más?**

Escribí tu predicción en el chat antes de que ejecutemos.

---

## Cómo reproducirlo en clase

En `lab/demo/src/app/sessions/s02/s02.component.html`, al final del archivo,
fuera del `@for`. Después de explicar, se borra.
