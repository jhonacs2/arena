# 1 · Un `push` sobre el array de un signal

**No ejecutes todavía.** Mirá el código y escribí en el chat qué va a pasar.

---

```ts
@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <p class="cuantas">{{ items().length }}</p>
    <p class="suma">{{ total() }}</p>
    <button type="button" (click)="mutate()">Agregar</button>
  `,
})
export class DemoComponent {
  readonly items = signal<number[]>([1, 2]);
  readonly total = computed(() => this.items().reduce((a, b) => a + b, 0));

  mutate(): void {
    this.items().push(10);
  }
}
```

### La pregunta

Antes de tocar el botón se ve **2** y **3**.

**Después de tocarlo, ¿qué muestra cada uno?**

1. `2` y `3` — no cambia nada
2. `3` y `13` — cambian los dos
3. otra cosa

Escribí tu predicción en el chat antes de que ejecutemos.

---

## Cómo reproducirlo en clase

Ya está pasando en el tablero: en `lab/demo/src/app/sessions/s03/s03.component.ts`
cambiá `add()` por `this.orders().push({ … })` —hay que sacarle el `readonly` al
tipo para que compile, y **ese gesto vale la pena mostrarlo**—, y mirá el panel
«Por cobrar».

**El navegador tiene que estar a la vista.** Esto no se ve en la terminal.
