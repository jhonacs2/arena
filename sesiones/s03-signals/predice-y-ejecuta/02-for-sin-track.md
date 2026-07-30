# 2 · Un `@for` sin `track`

**No ejecutes todavía.** Mirá el código y escribí en el chat qué va a pasar.

---

```html
<ul>
  @for (order of visible()) {
    <li>{{ order.customer }}</li>
  }
</ul>
```

### La pregunta

**¿Compila?**

1. Sí: `track` es opcional y Angular usa el índice
2. Sí, pero tira una advertencia en la consola
3. No compila

Escribí tu predicción en el chat antes de que ejecutemos.

---

## Cómo reproducirlo en clase

En `lab/demo/src/app/sessions/s03/s03.component.html`, sacale el
`; track order.id` al `@for` del tablero. **Mostrá la terminal**: el error sale
ahí, antes de que el navegador tenga nada que decir.

Después se repone.

Y si sobra un minuto, la segunda pregunta:

> «¿Y por qué no alcanza con `track $index`?»
