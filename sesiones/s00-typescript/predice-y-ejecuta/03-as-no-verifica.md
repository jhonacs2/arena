# 3 · Un dato que jura ser una carrera

**No ejecutes todavía.** Mirá el código y escribí en el chat qué va a pasar.

---

```ts
interface Horse {
  readonly name: string;
}

interface Race {
  readonly name: string;
  readonly horses: readonly Horse[];
}

const text = '{"name":"Clásico Apertura"}';
const race = JSON.parse(text) as Race;

console.log(race.name);
console.log(race.horses.length);
```

### La pregunta

**¿Compila? Y si compila, ¿qué imprime cada `console.log`?**

Fijate bien en el JSON: tiene el nombre de la carrera. Nada más.

Escribí tu predicción en el chat antes de que ejecutemos.

---

## Cómo reproducirlo en clase

En la consola del navegador alcanza, porque el punto es que el compilador **no
participa**. Pero es más fuerte hacerlo en el editor, en
`lab/demo/src/app/sessions/s00/menu.ts`: se ve que no hay ni un subrayado rojo
antes de ejecutar.

Después de explicar, se borra.
