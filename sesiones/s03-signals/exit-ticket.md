# S3 · Exit ticket

**3 minutos. Anónimo si preferís.**

---

**1. Recordar** — De estas cuatro cosas de una pantalla de carreras, marcá cuáles
son **estado** (van en un `signal`) y cuáles son **derivadas** (van en un
`computed`):

| | estado / derivado |
|---|---|
| El texto del buscador | |
| Cuántas carreras están en vivo | |
| Cuál carrera está abierta | |
| El pago posible de la apuesta | |

**2. Aplicar** — Este código compila y está mal. Escribí la línea corregida y
decí **qué se ve en la pantalla** si se deja así:

```ts
protected add(order: Order): void {
  this.orders().push(order);
}
```

**3. ¿Qué quedó confuso?**

---

## Para el instructor

**Respuestas de la 1:** estado · derivado · estado · derivado.

**Respuesta de la 2:** `this.orders.update((orders) => [...orders, order]);`

Y la segunda mitad es la que importa: **la lista se actualiza y los `computed` no**.
Quien conteste «no se ve nada» se llevó la mitad de la clase; quien conteste «media
pantalla queda vieja» se la llevó entera.

| Lo que escriben | Qué decirles |
|---|---|
| `this.orders.set([...this.orders(), order])` | Funciona y es correcto. `update` es más corto y evita leer y escribir en dos pasos. |
| `this.orders.update(o => o.push(order))` | `push` devuelve la longitud, no el array. El signal quedaría con un número adentro. |
| «No se actualiza la pantalla» | Es lo que casi todos creen. Vale volver a mostrarlo treinta segundos al empezar S4. |

| Señal | Qué hacer |
|---|---|
| Más de un tercio marca «cuántas están en vivo» como estado | Abre S4, 5 minutos |
| Aparece «no entendí `computed`» | Mostrar de nuevo los cuatro métodos donde **no** hay que actualizar nada |
| Nadie confundido | Subir el techo de la Misión 2 la próxima |
