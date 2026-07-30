# 3 · Un `computed` que ordena

**No ejecutes todavía.** Mirá el código y escribí en el chat qué va a pasar.

---

```ts
@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <p class="crudo">{{ raw() }}</p>
    <p class="ordenado">{{ sorted() }}</p>
  `,
})
export class DemoComponent {
  readonly numbers = signal<number[]>([3, 1, 2]);

  readonly raw = computed(() => this.numbers().join(','));
  readonly sorted = computed(() => this.numbers().sort((a, b) => a - b).join(','));
}
```

### La pregunta

`numbers` se declaró `[3, 1, 2]` y **nadie llama a `set` ni a `update` en ningún
lado**.

**¿Qué muestra `raw()`?**

1. `3,1,2` — es el valor tal cual se declaró
2. `1,2,3`
3. Da error: no se puede escribir adentro de un `computed`

Escribí tu predicción en el chat antes de que ejecutemos.

---

## Cómo reproducirlo en clase

En `lab/demo/src/app/sessions/s03/s03.component.ts`, cambiá el `computed` de
`byPrice` por la versión sin la copia:

```ts
protected readonly byPrice = computed<readonly Order[]>(() =>
  this.visible().sort((a, b) => lineTotal(b) - lineTotal(a)),
);
```

Y mirá **la lista de arriba**, no el panel. Después se repone el `[...]`.
