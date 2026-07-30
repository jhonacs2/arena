# S5 · Exit ticket

**3 minutos. Anónimo si prefieres.**

---

**1. Recordar** — Para cada caso, escribe si va `providedIn: 'root'` o
`providers` del componente:

| Qué | Dónde se declara |
|---|---|
| La sesión del usuario | |
| El texto que el usuario escribió en **este** buscador | |
| El carrito de compras | |
| El estado de apertura de **este** menú desplegable | |

**2. Aplicar** — Este código compila y falla al tocar el botón. Escribe la
versión corregida y di **por qué** falla:

```ts
export class BetComponent {
  protected place(): void {
    const bets = inject(BetStore);
    bets.submit();
  }
}
```

**3. ¿Qué quedó confuso?**

---

## Para el instructor

**Respuestas de la 1:** root · componente · root · componente.

La pregunta que las contesta todas es una sola: **¿cuántos de estos tiene que
haber?** Quien conteste bien las cuatro entendió la sesión, aunque no recuerde la
sintaxis.

**Respuesta de la 2:**

```ts
export class BetComponent {
  private readonly bets = inject(BetStore);

  protected place(): void {
    this.bets.submit();
  }
}
```

Falla porque cuando corre un método, Angular ya terminó de construir el
componente y no sabe a qué inyector preguntarle. `NG0203`.

| Lo que escriben | Qué decirles |
|---|---|
| Mueven el `inject()` pero no explican por qué | La mitad. Vale preguntar en voz alta qué es un contexto de inyección. |
| «Falla porque `inject` es asíncrono» | No lo es. Es una cuestión de **cuándo**, no de esperar. |
| `new BetStore()` | Compila y crea un store nuevo, vacío, que no comparte nada. Es el error del primer «predice y ejecuta» con otra cara. |

| Señal | Qué hacer |
|---|---|
| Confunden root con componente en más de un caso | Abre S6 con el diagrama, 5 minutos |
| Aparece «no entendí la jerarquía» | Es lo más abstracto del día. Se contesta con el ejemplo de los dos cuadernos |
| Nadie confundido | Subir el techo de la Misión 2 la próxima |
