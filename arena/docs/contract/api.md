# Arena — API

Base: `/api`. Todo JSON. Todas las respuestas de error usan el mismo sobre.

**Autenticación:** `Authorization: Bearer <access>`. JWT HS256, igual que el
backend del hipódromo: 15 minutos de vida, refresh de un solo uso en cookie
`HttpOnly`.

**Los montos son enteros.** Monedas en unidades; cuotas ×100 (`340` = 3.40). Nunca
`float` en el cable — ver el comentario de `horses.nominal_odds` en `schema.sql`.

> **Las cuotas son nominales y no determinan el pago.** La liquidación es
> **pari-mutuel** (`decisiones.md` §1): el pozo se reparte entre los que aciertan.
> Por eso ninguna apuesta lleva cuota congelada ni pago potencial — al apostar,
> cuánto se cobra todavía no existe. Este documento decía lo contrario mientras la
> economía era de cuota fija; `decisiones.md` manda, y esto se alineó a él.

---

## Sobre de error

```json
{ "error": { "code": "CODE_ALREADY_REDEEMED", "message": "Ese código ya fue usado." } }
```

`message` está **en castellano** y se muestra tal cual al usuario. `code` es para
el frontend.

| Código | HTTP | Cuándo |
|---|---|---|
| `VALIDATION_FAILED` | 400 | campo faltante o mal formado |
| `CODE_NOT_FOUND` | 404 | el código no existe |
| `CODE_ALREADY_REDEEMED` | 409 | ya lo usó alguien |
| `USERNAME_TAKEN` | 409 | usuario ocupado |
| `INVALID_CREDENTIALS` | 401 | usuario o contraseña incorrectos |
| `UNAUTHENTICATED` | 401 | falta el token o venció |
| `FORBIDDEN` | 403 | el rol no alcanza |
| `USER_NOT_FOUND` | 404 | el alumno al que se le quiso regalar no existe |
| `RATE_LIMITED` | 429 | demasiados intentos seguidos |
| `RACE_NOT_FOUND` | 404 | |
| `RACE_NOT_OPEN` | 409 | apostar en una carrera que no está `open` |
| `BET_ALREADY_PLACED` | 409 | ya apostó en esta carrera |
| `INSUFFICIENT_BALANCE` | 409 | el monto supera el saldo |
| `INVALID_TRANSITION` | 409 | transición de estado no permitida |
| `INTERNAL` | 500 | |

---

## Público

### `POST /api/auth/check-code`

Valida el código **sin canjearlo**, para habilitar el resto del formulario.

```json
→ { "code": "AVBD-1234" }
← { "valid": true, "coinsGranted": 1000 }
```

Un código inexistente y uno ya canjeado devuelven códigos de error distintos: el
alumno tiene que poder distinguir «lo escribí mal» de «ya me registré».

### `POST /api/auth/redeem`

Canjea el código y crea la cuenta. **Una transacción**: usuario + código marcado +
monedas acreditadas, o nada.

```json
→ {
    "code": "AVBD-1234",
    "firstName": "Ana",
    "lastName": "Gómez",
    "username": "anag",
    "password": "……"
  }
← { "accessToken": "…", "user": { … }, "balance": 1000, "points": 10 }
```

Contraseña: mínimo 8 caracteres. Se hashea con PBKDF2 (stdlib), como el hipódromo.

### `POST /api/auth/login`

```json
→ { "username": "anag", "password": "……" }
← { "accessToken": "…", "user": { … }, "balance": 1000, "points": 10 }
```

### `POST /api/auth/refresh` · `POST /api/auth/logout`

Refresh de un solo uso: al usarlo se invalida y se emite uno nuevo.

---

## Alumno

### `GET /api/me`

```json
← {
    "user": { "id": "…", "username": "anag", "firstName": "Ana", "lastName": "Gómez" },
    "balance": 1350,
    "points": 13
  }
```

`points = max(10, floor(balance / 100)) + puntos regalados`. **El piso de 10 es
parte del contrato** (`decisiones.md` §1): apostar mal saca monedas, nunca nota.

### `GET /api/me/transactions`

El historial del ledger, más nuevo primero. Es lo que le permite al alumno
entender por qué tiene la nota que tiene.

```json
← { "items": [
    { "id": 42, "delta": -200, "reason": "bet_placed", "balanceAfter": 800,
      "createdAt": "…", "raceName": "Clásico del Recuerdo" },
    { "id": 41, "delta": 1000, "reason": "code_redeemed", "balanceAfter": 1000, "createdAt": "…" }
  ] }
```

### `GET /api/races`

Las carreras visibles: `open`, `running`, `finished`. **Nunca `draft`.**

```json
← { "items": [
    { "id": "…", "name": "Clásico del Recuerdo", "status": "open",
      "scheduledAt": "…", "horseCount": 6, "participantCount": 12,
      "myBet": null }
  ] }
```

`myBet` viene resuelto: sin él, el frontend haría una llamada por carrera.

### `GET /api/races/:id`

Detalle con caballos, participantes de la sala, y —si terminó— resultados.

### `POST /api/races/:id/join`

Entra a la sala. Idempotente.

### `POST /api/races/:id/bet`

```json
→ { "horseId": "…", "amount": 200 }
← { "bet": { "id": "…", "horseId": "…", "amount": 200, "status": "placed",
             "payout": null, "createdAt": "…", "settledAt": null },
    "balance": 800 }
```

**Se valida en el servidor, en una transacción:** la carrera está `open`, el
caballo es de esa carrera, `1 ≤ amount ≤ balance`, y no hay apuesta previa.

`status` es `placed` · `won` · `lost` · `refunded`. **`payout` llega en `null`
hasta que la carrera se liquida**, y ahí es el pago real. No hay pago potencial:
con pari-mutuel depende del pozo final y de cuántos aciertan, y las dos cosas
siguen cambiando hasta que se cierran las apuestas.

La respuesta **no trae `points`**. El saldo sí, porque cambió; los puntos salen de
la vista `user_scores` y se piden en `GET /api/me`.

---

## Instructor (`role = admin`)

Todos devuelven `403 FORBIDDEN` si el rol no alcanza. **El rol se verifica en el
servidor en cada endpoint**, no en el frontend.

### `POST /api/admin/codes`

```json
→ { "count": 25, "coinsGranted": 1000, "note": "grupo del martes" }
← { "codes": ["AVBD-1234", "KMPR-8827", …] }
```

Alfabeto sin caracteres ambiguos: letras sin `I`, `L`, `O`, `U`; dígitos sin `0`
ni `1`. Un código se dicta en voz alta.

### `GET /api/admin/codes`

Todos, con su estado de canje y quién lo usó.

### `GET /api/admin/scores`

El panel de nota: la vista `user_scores`. Monedas, puntos, apuestas hechas y
ganadas por alumno.

### `POST /api/admin/users/:id/gift`

```json
→ { "coins": 300, "note": "participación en clase" }
← { "balance": 1650, "points": 16 }
```

Acepta `coins` negativo para un ajuste. Va al ledger como `gift` o `adjustment`,
con `created_by`, así que queda el rastro de quién lo hizo.

### `POST /api/admin/users/:id/grant-points`

Puntos de nota directos, **por fuera del juego**: no pasan por el saldo ni por el
ledger de monedas, y se suman **por encima** del piso de 10.

```json
→ { "points": 250, "reason": "explicó @for en el code review" }
← { "points": 12.5 }
```

`points` viaja **×100 en entero** en el pedido —`250` es 2,5 puntos— por la misma
razón que todo lo demás: en el cable no hay decimales. La respuesta sí trae el
total en puntos, que es lo que se muestra.

### `POST /api/admin/races`

```json
→ { "name": "Clásico del Recuerdo", "scheduledAt": "…",
    "horses": [ { "number": 1, "name": "Viento Norte", "nominalOdds": 340 }, … ] }
← { "race": { "id": "…", "status": "draft", … } }
```

### `PATCH /api/admin/races/:id` · `POST /api/admin/races/:id/horses`

Editar mientras está en `draft`.

### `POST /api/admin/races/:id/open`

`draft → open`. Desde acá los alumnos la ven y pueden apostar.

### `POST /api/admin/races/:id/start`

`open → running`. **Cierra las apuestas en el servidor** y arranca la simulación.
Fija `seed`, así que la carrera es reproducible.

### `POST /api/admin/races/:id/cancel`

A `cancelled` desde `draft`, `open` o `running`. **Devuelve cada apuesta íntegra**
al saldo (`bet_refunded`), en una transacción.

---

## WebSocket

`GET /api/ws?raceId=…`, con el token en el primer mensaje o en el query.

Los eventos son **servidor → cliente**. El cliente no manda nada salvo el
handshake: apostar es un `POST`, no un mensaje de socket.

### El sobre

**Todo evento es un objeto plano con `type` y `raceId`, y su carga al lado.** Esto
estaba sin escribir y el frontend lo asumió mal: leía la sala como si `room.state`
trajera los campos sueltos, cuando vienen anidados en `race`. El resultado fue
`undefined` metido en el estado y la pantalla de la carrera rota entera — sin un
error de compilación, porque un campo que no llega es `undefined` y sigue.

```json
{ "type": "race.tick", "raceId": "…", "t": 4.2,
  "positions": [{ "horseId": "…", "progress": 0.61, "place": 1 }] }
```

| Evento | Cuándo | Carga, además de `type` y `raceId` |
|---|---|---|
| `room.state` | al conectarse | `{ race }` — **el mismo objeto que `GET /races/:id`**, armado por destinatario porque incluye `myBet` |
| `room.joined` | alguien entra | `{ userId, username, participantCount }` |
| `bet.placed` | alguien apuesta | `{ userId, username, amount, betCount }` — **sin el caballo** |
| `race.started` | el instructor larga | `{ startedAt, bets }` — acá se revelan todas juntas |
| `race.tick` | 10 Hz mientras corre | `{ t, positions: [{ horseId, progress, place }] }`. `t` en segundos |
| `race.finished` | terminó | `{ results, bets, myBet, balance }` — **por destinatario** |
| `race.cancelled` | cancelada | `{ reason, myRefund, balance }` — **por destinatario** |

`balance` y `myRefund` llegan en **`null` cuando no cambiaron**, que no es lo mismo
que cero: `null` significa «no toques el saldo que ya tenías».

Ningún evento trae `points`: la nota suma los puntos regalados, que el módulo de
carreras no conoce. Salen de `GET /api/me`.

Dos detalles que no son cosméticos:

- **`bet.placed` no revela a qué caballo** mientras la carrera está `open`. Si lo
  revelara, los últimos en apostar copiarían a los primeros y la apuesta dejaría
  de medir criterio. Al pasar a `running` se revelan todas juntas.
- **`race.finished` se arma por destinatario.** Difundir el mismo objeto filtraría
  cuánto cobró cada uno — mismo motivo que en el backend del hipódromo.
