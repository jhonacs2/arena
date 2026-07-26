# CLAUDE.md — El contrato

> Complementa el [`CLAUDE.md` de la raíz](../../CLAUDE.md). Las decisiones de forma y la regla de rebase de fechas están en [`README.md`](README.md) — **leelo antes de tocar nada**.

**Esta carpeta es la fuente de verdad.** El backend Go y el frontend Angular se escriben contra ella, nunca uno contra el otro.

> **Si un campo o un endpoint no está en `openapi.yaml`, no existe.**
> Cuando haga falta algo nuevo se agrega **primero** acá, y recién después se implementa en los dos lados. Si el spec es ambiguo o falta un endpoint, **preguntá**. No inventes rutas.

| Archivo | Qué es |
|---|---|
| `openapi.yaml` | Los 13 endpoints con todos los esquemas. Normativo. |
| `ws-events.md` | El contrato del WebSocket. OpenAPI no cubre sockets. |
| `race-simulation.md` | El simulador, implementado dos veces: Go y JavaScript. |
| `error-codes.md` | Catálogo **cerrado** de `error.code`. |
| `seed/` | El dataset. Lo cargan **los dos** lados. |
| `samples/` | Una respuesta canónica por endpoint. Golden test de ambos lados. |
| `fixtures/race-ticks.jsonl` | Grabación de una carrera en vivo a 10 Hz. |

---

## Resumen operativo

Base `/api/v1`. Autenticación por `Authorization: Bearer <accessToken>`. Error uniforme:

```json
{ "error": { "code": "INVALID_CREDENTIALS", "message": "…", "details": {} } }
```

El frontend hace `switch` sobre `code`, **nunca sobre `message`**: el mensaje está para mostrarlo, el código para decidir.

| Método | Ruta | Nota |
|---|---|---|
| POST | `/auth/register` | `201`, dispara correo de verificación |
| POST | `/auth/verify` · `/auth/resend-verification` | el correo lleva a `{FRONT_URL}/verificar?token=…` |
| POST | `/auth/login` · `/auth/refresh` · `/auth/logout` | refresh token de **un solo uso** |
| GET | `/me` | |
| GET | `/races?status=&page=&size=` | paginado `{items, page, size, total}`; el listado **sí** trae `horses[]` |
| GET | `/races/:id` · `/races/:id/results` | `payouts` trae solo las apuestas del usuario autenticado |
| POST | `/bets` | rechaza si la carrera arrancó, si no hay saldo o si el monto sale de `[10, 5000]` |
| GET | `/bets/me?page=` · `/leaderboard?period=daily\|all` | |

WebSocket: `wss://{HOST}/ws?token={accessToken}` — el token va por query string porque **el navegador no permite headers en el handshake**. Una sola conexión, multiplexada por sala. `race.tick` llega a ~10 Hz.

**La simulación de la carrera es autoridad del servidor.** El front solo pinta `positions`. No interpoles físicas ni calcules ganadores en el cliente.

> ⚠️ **Detección de cambios y WebSocket.** Los eventos del socket llegan fuera de la zona de Angular y con `OnPush` la UI puede no repintar aunque el signal cambie. El `SocketService` escribe en signals y el componente los lee. **Si la carrera no se anima en el navegador, el problema es la zona, no el binding.** Es el contenido central de S10.

---

## Cómo no se separan los dos lados

Nada de codegen: es tooling que después hay que explicarle al alumno. Tres piezas que se leen enteras:

1. `samples/*.json` — una respuesta canónica por endpoint, escrita a mano una vez.
2. **Backend:** un test serializa cada handler y compara la **forma** contra su sample.
3. **Frontend:** un test parsea esos mismos samples contra los tipos de `core/models/`.

Los dos lados asertan sobre los mismos archivos. Cambiar un nombre de campo pone algo en rojo.

Las copias de `seed/` y `fixtures/` dentro de `project/` las mantiene `scripts/sync-contract.mjs`, y `verify.mjs` corre su `--check`.
