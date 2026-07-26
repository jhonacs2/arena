# S1 · Exit ticket

**3 minutos. Anónimo si preferís.**

---

**1. Recordar** — Nombrá los cuatro tipos de binding y decí, para cada uno, hacia dónde va la información: de la clase al template, del template a la clase, o en los dos sentidos.

**2. Aplicar** — Este botón tiene que quedar deshabilitado cuando `saldo` sea menor a 100. Escribí la línea:

```html
<button ______________>Apostar</button>
```

**3. ¿Qué quedó confuso?**

*(Vale «nada». Vale «todo». Vale una sola palabra.)*

---

## Para el instructor

La pregunta 3 es la que importa. Lo que aparezca ahí **arranca S2**, en el bloque de las 0:05, junto con el Wayground.

Respuesta esperada de la 2: `[disabled]="saldo < 100"`.

| Lo que escriben | Qué decirles |
|---|---|
| `disabled="saldo < 100"` | Sin corchetes es la cadena literal `"saldo < 100"`, que es *truthy*: el botón queda deshabilitado **siempre**. Es el mismo error del bloque de predicción. |
| `[disabled]="{{ saldo < 100 }}"` | Corchetes **o** llaves, nunca las dos. Los corchetes ya dicen «esto es una expresión». |
| `(disabled)="saldo < 100"` | Los paréntesis escuchan eventos. `disabled` no es un evento. |

| Señal | Qué hacer |
|---|---|
| Más de un tercio confundido con lo mismo | Abre S2, 5 minutos |
| Una sola persona trabada | Mensaje directo, no clase entera |
| Nadie confundido | La sesión fue fácil de más. Subir el techo de la Misión 2 la próxima |
