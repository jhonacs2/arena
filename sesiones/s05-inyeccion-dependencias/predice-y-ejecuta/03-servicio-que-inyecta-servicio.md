# 3 · Un servicio que pide otro servicio

**No ejecutes todavía.** Mira el código y escribe en el chat qué va a pasar.

---

```ts
@Injectable({ providedIn: 'root' })
export class ReceiptService {
  private readonly orders = inject(OrderService);

  readonly summary = computed(() => {
    const list = this.orders.orders();
    return list.length === 0 ? 'Sin pedidos' : `${list.length} pedidos`;
  });
}
```

`ReceiptService` no es un componente: es un servicio, y le pide otro servicio con
`inject()`.

### La pregunta

**¿Se puede?**

1. No: `inject()` solo funciona en componentes
2. Sí, pero hay que pasarlo por el constructor
3. Sí, exactamente así

Escribe tu predicción en el chat antes de que ejecutemos.

---

## Cómo reproducirlo en clase

Ya está hecho en el hipódromo: `BetStore` inyecta `RaceStore`. Si prefieres
mostrarlo en el lab, crea el `ReceiptService` de arriba y muestra su `summary()`
en el tablero.

**La pregunta de yapa, si sobra un minuto:**

> «¿Y si `OrderService` inyectara a `ReceiptService` al mismo tiempo?»
