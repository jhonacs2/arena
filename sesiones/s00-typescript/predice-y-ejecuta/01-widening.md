# 1 · La misma cadena, dos tipos distintos

**No ejecutes todavía.** Mirá el código y escribí en el chat qué va a pasar.

---

```ts
type CoffeeSize = 'chico' | 'mediano' | 'grande';

function sizeFactor(size: CoffeeSize): number {
  switch (size) {
    case 'chico':
      return 0.8;
    case 'mediano':
      return 1;
    case 'grande':
      return 1.3;
  }
}

// A
const size = 'grande';
sizeFactor(size);

// B
const config = { size: 'grande' };
sizeFactor(config.size);
```

### La pregunta

**¿Cuál de las dos llamadas compila: A, B, las dos o ninguna?**

Las dos cadenas dicen `'grande'`. Las dos están en un `const`.

Escribí tu predicción en el chat antes de que ejecutemos.

---

## Cómo reproducirlo en clase

En `lab/demo/src/app/sessions/s00/menu.ts`, al final del archivo:

```ts
const size = 'grande';
sizeFactor(size);

const config = { size: 'grande' };
sizeFactor(config.size);
```

Apoyá el mouse sobre `size` y después sobre `config.size` **antes** de mostrar el
error. La diferencia se ve en el tooltip.

Después de explicar, se borra.
