# El contrato

Esta carpeta es la **fuente de verdad** del proyecto. El backend Go y el frontend Angular se escriben contra ella, nunca uno contra el otro.

La regla es una sola: **si un campo o un endpoint no está acá, no existe.** Cuando haga falta algo nuevo, se agrega primero en `openapi.yaml` y recién después se implementa en los dos lados.

---

## Qué hay

| Archivo | Qué es |
|---|---|
| `openapi.yaml` | Los 13 endpoints, con todos los esquemas. Especificación normativa. |
| `ws-events.md` | El contrato del WebSocket. OpenAPI no cubre sockets, por eso va aparte. |
| `error-codes.md` | Catálogo cerrado de `error.code` y cómo trata cada uno el frontend. |
| `seed/` | El dataset. Lo cargan **los dos** lados. |
| `samples/` | Una respuesta canónica por endpoint. Golden test de ambos lados. |
| `fixtures/race-ticks.jsonl` | Grabación completa de una carrera en vivo, 10 Hz. |

---

## Por qué el mecanismo anti-deriva es tan simple

Nada de codegen ni de paquetes compartidos: es tooling que después hay que explicarle al alumno en clase. En su lugar, tres piezas que un alumno de segundo día puede leer:

1. `samples/*.json` — escritos a mano una vez.
2. **Backend:** un test serializa cada handler y lo diffea contra su sample. Si la forma cambia, no compila el CI.
3. **Frontend:** un test parsea esos mismos samples contra los tipos de `core/models/`. Si falta un campo, TypeScript lo grita.

Los dos lados asertan sobre **los mismos archivos**. No hay forma de que se separen sin que algo se ponga rojo.

---

## Fechas: la regla de rebase

El seed está anclado a un instante fijo:

```
ANCHOR = 2026-03-14T12:00:00Z   ←  la largada de race_005
```

Con ese ancla, el dataset tiene siempre la misma forma: 4 carreras terminadas, 1 en vivo, 3 por venir.

Pero un `startsAt` de marzo de 2026 en una clase de agosto se ve viejo. Entonces, **al arrancar**, backend y mock aplican el mismo desplazamiento:

```
offset      = now - ANCHOR
startsAt'   = startsAt + offset
placedAt'   = placedAt + offset
finishedAt' = finishedAt + offset
```

Una línea de cada lado. El resultado: la carrera en vivo siempre está en vivo, la próxima siempre arranca en 8 minutos, y la cuenta regresiva siempre tiene algo que contar.

**Los `samples/` no se rebasean.** Guardan las fechas literales del seed, porque un golden test contra un valor que cambia cada segundo no es un test.

---

## Decisiones que tomé y que conviene que mires

`CLAUDE.md` §6 fija los endpoints, pero deja tres formas sin tipar. Las definí en `openapi.yaml`; si alguna no te cierra, se cambia acá **antes** de que exista código que dependa de ella.

| Decisión | Qué elegí | Por qué |
|---|---|---|
| `GET /races` trae `horses[]` completos | Sí, en el listado también | La interfaz `Race` de `CLAUDE.md` declara `horses` obligatorio. Si el listado no los trajera, el tipo tendría que ser opcional y el alumno pelearía con `undefined` en S1. Además la card necesita las cuotas para mostrar el favorito. |
| Forma de `/races/:id/results` | `{ raceId, finishedAt, podium[3], payouts[] }` | `podium` es público; `payouts` trae **solo las apuestas del usuario autenticado**. Sin sesión, `payouts: []`. Nadie ve lo que cobró otro. |
| Forma de `LeaderboardEntry` | `{ rank, userId, displayName, profit, bets, wins }` | `profit` = pagos − apostado. Sin `balance`, porque el saldo no dice quién apuesta bien: alguien que se registró hoy tiene 1000 sin haber acertado nada. |
| Forma de `Bet` | Denormaliza `raceName` y `horseName` | El historial se pinta sin ir a buscar la carrera ni el caballo. Evita un N+1 y una lección de joins que no toca dar. |
| `amount` y `balance` | Enteros, sin centavos | Es saldo virtual. Los flotantes en dinero son una clase entera que no está en el temario. |
| `odds` | Decimal, congelada al apostar | Pago = `amount × odds`. Si la cuota cambia después, el pago usa la guardada. |
| Rango de apuesta | `[10, 5000]`, entero | Le da a S8 un validador custom con un motivo real, además del de saldo. |
| Empates en el leaderboard | `profit` desc → `wins` desc → `displayName` asc | Un orden total. Sin esto el ranking baila entre requests y el test golden es inestable. |

---

## El dataset

| | |
|---|---|
| Usuarios | 12 · contraseña de desarrollo para todos: `Carrera123!` |
| Carreras | 8 — 4 `finished`, 1 `live`, 3 `upcoming` |
| Caballos | 54, entre 6 y 8 por carrera |
| Apuestas | 34 — 30 liquidadas, 4 pendientes |

Tres usuarios están puestos para casos específicos:

- **`ana@hipodromo.test`** — el usuario por defecto de las demos. Historial mixto, apuestas pendientes en la carrera en vivo y en una por venir.
- **`caro@hipodromo.test`** — **correo sin verificar**. Es el caso de prueba de `verified.guard` (S9) y del `403 EMAIL_NOT_VERIFIED`. No tiene apuestas, justamente porque no puede apostar.
- **`hugo@hipodromo.test`** — saldo 980, perdió todo lo que jugó. Sirve para provocar `INSUFFICIENT_BALANCE` sin armar nada.

`leaderboard.json` es el **golden esperado**, no una entrada: el backend calcula el ranking desde `bets.json` + `results.json`, y el test compara contra ese archivo.

---

## El fixture de la carrera

`fixtures/race-ticks.jsonl` — 462 eventos: 10 s de cuenta regresiva, largada, 448 ticks a 10 Hz (44,8 s de carrera) y llegada con pagos, saldo y leaderboard.

Se regenera con `node scripts/gen-race-ticks.mjs` y es determinístico — sin `Math.random`, sin `Date.now`. Correrlo dos veces da el mismo archivo, así git no acumula ruido.

La carrera está armada para que se vea algo: **Payador** (el favorito, 2.75) corre de atrás, va **séptimo a mitad de carrera** y gana rematando. Un fixture donde el favorito lidera de punta a punta no demuestra que la animación funciona.

Cada línea trae `_offsetMs`: cuántos milisegundos después del inicio se emitió. El `MockSocketService` lo reproduce con esa cadencia.

Es, además, el **seguro de clase**: si el backend desplegado se cae en plena S10, la carrera en vivo igual corre y la lección de zona y detección de cambios se da igual.
