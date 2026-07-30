# 2 · `inject()` adentro de un método

**No ejecutes todavía.** Mira el código y escribe en el chat qué va a pasar.

---

```ts
export class CounterComponent {
  protected take(): void {
    const orders = inject(OrderService);
    orders.add(this.customer(), this.coffee());
    this.customer.set('');
  }
}
```

Se pide el servicio justo donde se lo usa, en vez de arriba en un campo.

### La pregunta

**¿Qué pasa?**

1. No compila
2. Compila, y al tocar «Tomar pedido» falla
3. Anda igual: es lo mismo, solo que más cerca de donde se usa

Escribe tu predicción en el chat antes de que ejecutemos.

---

## Cómo reproducirlo en clase

En `lab/demo/src/app/sessions/s05/counter.component.ts`, mueve el `inject()`
adentro de `take()`.

**Muestra la terminal primero** —para que se vea que compiló sin una queja— y
recién después toca el botón con la consola del navegador abierta.

Después se repone.
