# Catálogo de errores

Formato uniforme para **toda** respuesta de error del backend (`CLAUDE.md` §6):

```json
{ "error": { "code": "INVALID_CREDENTIALS", "message": "Correo o contraseña incorrectos.", "details": {} } }
```

Reglas:

- `code` es un **catálogo cerrado**. El frontend hace `switch` sobre él; nunca sobre `message`.
- `message` está en **español** y se muestra tal cual al usuario. Si el `error.interceptor` no reconoce el `code`, cae al `message` del servidor.
- `details` siempre existe. Es `{}` cuando no hay nada que agregar.
- Un solo error por respuesta. Los errores de validación por campo van dentro de `details.fields`.

---

## Catálogo

| `code` | HTTP | Cuándo | `details` |
|---|---|---|---|
| `VALIDATION_FAILED` | 422 | El body no pasa validación | `{ "fields": { "email": "Formato inválido" } }` |
| `INVALID_CREDENTIALS` | 401 | Login con correo o contraseña incorrectos | `{}` |
| `EMAIL_ALREADY_REGISTERED` | 409 | Registro con un correo que ya existe | `{}` |
| `EMAIL_NOT_VERIFIED` | 403 | Acción que exige correo verificado | `{}` |
| `INVALID_VERIFICATION_TOKEN` | 400 | Token de verificación mal formado o inexistente | `{}` |
| `VERIFICATION_TOKEN_EXPIRED` | 410 | Token de verificación vencido (24 h) | `{}` |
| `ALREADY_VERIFIED` | 409 | Reenvío o verificación sobre una cuenta ya verificada | `{}` |
| `UNAUTHENTICATED` | 401 | Falta el `Authorization`, o el access token venció | `{}` |
| `INVALID_REFRESH_TOKEN` | 401 | Refresh token inválido, vencido o ya rotado | `{}` |
| `FORBIDDEN` | 403 | Autenticado, pero el recurso no es suyo | `{}` |
| `NOT_FOUND` | 404 | Carrera, caballo o apuesta inexistente | `{ "resource": "race", "id": "race_099" }` |
| `RACE_ALREADY_STARTED` | 409 | Apuesta sobre una carrera `live` o `finished` | `{ "raceId": "race_005", "status": "live" }` |
| `HORSE_NOT_IN_RACE` | 422 | El `horseId` no corre en ese `raceId` | `{ "raceId": "race_006", "horseId": "hrs_001" }` |
| `INSUFFICIENT_BALANCE` | 409 | El monto supera el saldo | `{ "balance": 120, "amount": 500 }` |
| `BET_AMOUNT_OUT_OF_RANGE` | 422 | Monto fuera de `[10, 5000]` o no entero | `{ "min": 10, "max": 5000 }` |
| `RESULTS_NOT_AVAILABLE` | 409 | `/races/:id/results` sobre una carrera que no terminó | `{ "status": "upcoming" }` |
| `RATE_LIMITED` | 429 | Demasiadas peticiones (login y reenvío de correo) | `{ "retryAfterSeconds": 60 }` |
| `INTERNAL` | 500 | Cualquier cosa no prevista | `{}` |

---

## Cómo los trata el frontend

| Manejo | Códigos | Quién |
|---|---|---|
| Refrescar el token y reintentar **una** vez | `UNAUTHENTICATED` | `auth.interceptor` (S7) |
| Cerrar sesión y mandar a `/login` | `INVALID_REFRESH_TOKEN`, o segundo `UNAUTHENTICATED` seguido | `auth.interceptor` |
| Redirigir a `/verificar` | `EMAIL_NOT_VERIFIED` | `verified.guard` (S9) |
| Pintar error bajo el campo | `VALIDATION_FAILED` con `details.fields` | formularios (S8) |
| Mostrar mensaje en el formulario de apuesta | `INSUFFICIENT_BALANCE`, `RACE_ALREADY_STARTED`, `BET_AMOUNT_OUT_OF_RANGE` | `bet-form` (S8) |
| Vista de error con botón de reintento | el resto | cada feature |

## Errores del WebSocket

El socket **no** usa este formato. Cierra la conexión con un código y el cliente reconecta con backoff:

| Código de cierre | Significado | Acción del cliente |
|---|---|---|
| `1000` | Cierre limpio | no reconectar |
| `4001` | Token ausente, inválido o vencido | refrescar el token y reconectar una vez |
| `4029` | Demasiadas conexiones para ese usuario | backoff largo |
| otro | Caída de red o del servidor | backoff exponencial: 1 s, 2 s, 4 s, 8 s, tope 30 s |
