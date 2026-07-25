# Contrato WebSocket

Endpoint: `wss://{HOST}/ws?token={accessToken}`

El token va por **query string** porque el navegador no permite mandar headers en el handshake de WebSocket. No hay forma de usar `Authorization: Bearer` acá — es una limitación de la API `WebSocket` del navegador, no una decisión de diseño.

---

## Cliente → servidor

```jsonc
{ "type": "subscribe",   "raceId": "race_005" }   // empieza a recibir eventos de esa carrera
{ "type": "unsubscribe", "raceId": "race_005" }   // deja de recibirlos
{ "type": "ping" }                                // keepalive; el servidor responde { "type": "pong" }
```

Una sola conexión para toda la app, **multiplexada por sala**. El `SocketService` es un singleton en `core/ws/`; los componentes se suscriben a una carrera, no abren un socket.

El cliente manda `ping` cada **25 s**. Si pasan **60 s** sin recibir nada del servidor, el cliente asume conexión muerta y reconecta.

---

## Servidor → cliente

```ts
type ServerEvent =
  | { type: 'race.countdown';     raceId: string; secondsLeft: number }
  | { type: 'race.started';       raceId: string; startedAt: string }
  | { type: 'race.tick';          raceId: string; t: number;
      positions: { horseId: string; progress: number; place: number }[] }
  | { type: 'race.finished';      raceId: string; podium: string[];
      payouts: { betId: string; amount: number }[] }
  | { type: 'balance.updated';    balance: number }
  | { type: 'leaderboard.updated'; entries: LeaderboardEntry[] }
  | { type: 'pong' };
```

### Por evento

| Evento | Alcance | Frecuencia | Notas |
|---|---|---|---|
| `race.countdown` | sala | 1 Hz, últimos 60 s | `secondsLeft` cuenta 60 → 0 |
| `race.started` | sala | una vez | `startedAt` en ISO 8601 UTC |
| `race.tick` | sala | **~10 Hz** | ver abajo |
| `race.finished` | sala | una vez | `podium` son 3 `horseId` en orden; `payouts` solo trae las apuestas **del usuario conectado** |
| `balance.updated` | usuario | al liquidar o al apostar | saldo absoluto, no delta |
| `leaderboard.updated` | broadcast | al terminar una carrera | top 20 del período `all` |
| `pong` | conexión | respuesta a `ping` | |

### `race.tick` en detalle

```jsonc
{
  "type": "race.tick",
  "raceId": "race_005",
  "t": 12.4,                               // segundos desde la largada, un decimal
  "positions": [
    { "horseId": "hrs_028", "progress": 0.412, "place": 3 },
    { "horseId": "hrs_029", "progress": 0.437, "place": 1 }
  ]
}
```

- `progress` va de `0` a `1`. `1` es la línea de llegada.
- `place` es la posición **en ese instante**, empezando en 1. Se recalcula cada tick.
- `positions` trae **todos** los caballos de la carrera, siempre. No es un delta.
- El orden del array no significa nada: ordenar por `place` es tarea del cliente.

### La regla que no se negocia

**La simulación es autoridad del servidor.** El front solo pinta `positions`. No se interpolan físicas, no se predice, no se calcula el ganador en el cliente. Si un tick se pierde, el siguiente corrige — no hay estado que reconstruir.

### ⚠️ Zona y detección de cambios

Los eventos del socket llegan **fuera de la zona de Angular**: `WebSocket.onmessage` es una API que zone.js no parchea de la misma forma que `setTimeout` o los eventos del DOM, y con `OnPush` en todos los componentes la vista puede no repintar aunque el signal sí cambie.

El `SocketService` escribe en signals y el componente lee esos signals. Si la carrera **no se anima en el navegador**, el problema es la zona, no el binding. Es el contenido central de S10 y por eso está acá y no en un comentario perdido del código.

---

## Fixture grabado

`fixtures/race-ticks.jsonl` es una grabación completa de `race_005` (8 caballos): countdown, largada, ~45 s de ticks a 10 Hz y llegada. Un evento JSON por línea, con el campo extra `_offsetMs` que indica cuántos milisegundos después del inicio se emitió.

Lo usan dos cosas:

1. `MockSocketService` (frontend) lo reproduce en tiempo real. **Es el seguro de clase**: si el backend desplegado se cae en plena S10, la carrera en vivo igual corre y la lección se da igual.
2. El test golden del backend compara la forma de sus eventos contra las primeras líneas del fixture.

Se regenera con `node scripts/gen-race-ticks.mjs` — es determinístico, siempre produce el mismo archivo.
