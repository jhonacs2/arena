# 2 · Una directiva sin `standalone`

**No ejecutes todavía.** Mira el código y escribe en el chat qué va a pasar.

---

```ts
@Directive({
  selector: '[appHighlight]',
  host: {
    '[class.is-highlighted]': 'appHighlight()',
  },
})
export class HighlightDirective {
  readonly appHighlight = input(false);
}
```

Y el componente, que sí es standalone, la declara:

```ts
imports: [HighlightDirective],
```

### La pregunta

**¿Qué pasa?**

1. Anda igual: `standalone` es opcional
2. Compila, pero la directiva no hace nada
3. No compila

Escribe tu predicción en el chat antes de que ejecutemos.

---

## Cómo reproducirlo en clase

En `lab/demo/src/app/sessions/s04/highlight.directive.ts`, quita la línea
`standalone: true`. Muestra la terminal.

Y la pregunta de yapa, si sobra un minuto:

> «¿Y si el pipe fuera el que no tiene `standalone`?»

La respuesta es la misma, con otro nombre de clase en el mensaje. Vale la pena
decirlo: **la regla es una sola para los tres**.
