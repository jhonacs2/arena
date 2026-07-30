# 1 · Declarado en los dos lados

**No ejecutes todavía.** Mira el código y escribe en el chat qué va a pasar.

---

El servicio dice que vive en la raíz:

```ts
@Injectable({ providedIn: 'root' })
export class OrderService {
  private readonly _orders = signal<readonly Order[]>([]);
  readonly orders = this._orders.asReadonly();
  readonly count = computed(() => this.orders().length);
  add(…): void { … }
}
```

Y **además** el mostrador lo declara en sus `providers`:

```ts
@Component({
  selector: 'app-counter',
  standalone: true,
  providers: [OrderService],
  …
})
export class CounterComponent {
  protected readonly orders = inject(OrderService);
}
```

En la pantalla hay dos mostradores y un tablero que también pide `OrderService`.

### La pregunta

**Tomo un pedido en el mostrador A. ¿Qué se ve?**

1. Aparece en el B y en el tablero
2. Da error: está declarado dos veces
3. Otra cosa

Escribe tu predicción en el chat antes de que ejecutemos.

---

## Cómo reproducirlo en clase

En `lab/demo/src/app/sessions/s05/counter.component.ts`, agrega `OrderService`
al array `providers` que ya tiene `NotepadService`.

**Toma un pedido en el A** y mira los tres contadores: el del A, el del B y el
del tablero. Después se quita.
