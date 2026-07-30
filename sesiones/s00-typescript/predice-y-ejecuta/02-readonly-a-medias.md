# 2 · Un campo `readonly` con una lista adentro

**No ejecutes todavía.** Mirá el código y escribí en el chat qué va a pasar.

---

```ts
interface OrderLine {
  readonly quantity: number;
}

interface Order {
  readonly customer: string;
  readonly lines: OrderLine[];
}

const order: Order = { customer: 'Ana', lines: [] };

// A
order.customer = 'Beto';

// B
order.lines = [];

// C
order.lines.push({ quantity: 1 });
```

### La pregunta

**Las tres líneas están marcadas `readonly`. ¿Cuáles compilan?**

Escribí tu predicción en el chat antes de que ejecutemos.

---

## Cómo reproducirlo en clase

En `lab/demo/src/app/sessions/s00/menu.ts`, al final del archivo. Las tres líneas
juntas, para que se vean los errores que hay y el que no.

Conviene mostrar el arreglo en el mismo momento:

```ts
readonly lines: readonly OrderLine[];
```

Después de explicar, se borra.
