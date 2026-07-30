# S2 · Exit ticket

**3 minutos. Anónimo si preferís.**

---

**1. Recordar** — Para cada cosa, decí si en la tarjeta de una carrera viaja por
`input()`, por `output()` o por `ng-content`:

| Qué | Por dónde |
|---|---|
| Los datos de la carrera | |
| Que alguien tocó la tarjeta | |
| La pastilla de estado, con su texto | |
| Si la tarjeta está abierta | |

**2. Aplicar** — El hijo emite y el padre no se entera. Escribí la línea que
falta en el padre:

```html
<app-coffee-card [coffee]="item.coffee" ______________ />
```

**3. ¿Qué quedó confuso?**

*(Vale «nada». Vale «todo». Vale una sola palabra.)*

---

## Para el instructor

La pregunta 3 es la que importa. Lo que aparezca ahí **arranca S3**.

**Respuestas esperadas de la 1:** `input()` · `output()` · `ng-content` ·
`input()`.

La cuarta es la que separa: quien pone «estado del hijo» no entendió por qué solo
puede haber una abierta a la vez, y vale cinco minutos al empezar S3.

**Respuesta esperada de la 2:** `(ordered)="take($event)"`.

| Lo que escriben | Qué decirles |
|---|---|
| `[ordered]="take($event)"` | Corchetes es property binding: estarías **pasándole** algo al hijo. Los avisos se escuchan con paréntesis. |
| `(ordered)="take()"` sin `$event` | Compila, y el método recibe `undefined`. El dato viaja en `$event`; si no lo pasás, se pierde. |
| `(ordered)="orders.push($event)"` | Funciona hoy y es lo que rompe la pantalla en S3. Vale nombrarlo: «acordate de esto en dos clases». |

| Señal | Qué hacer |
|---|---|
| Más de un tercio confunde `input` con `ng-content` | Abre S3, 5 minutos, con el diagrama |
| Aparece «no entendí los paréntesis de `coffee()`» | **Es lo esperado.** Se contesta entero en S3 |
| Nadie confundido | Subir el techo de la Misión 2 la próxima |
