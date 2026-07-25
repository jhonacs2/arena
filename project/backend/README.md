# Backend — Hipódromo

Monolito en Go que sirve la app ancla del módulo Angular.

Un solo binario, **una sola dependencia externa** (`github.com/coder/websocket`), sin base de datos y con el dataset embebido. `go build` produce un ejecutable que corre solo, sin archivos al lado.

> **Este servicio está congelado.** Se terminó en la Fase 1 y desde ahí no se modifica: el frontend se escribe contra él, no al revés. Si falta un endpoint, se agrega primero en `docs/contract/openapi.yaml` y se conversa — ver `CLAUDE.md` §12.

---

## Levantarlo

```bash
go run .                 # http://localhost:8080
RESET=1 go run .         # ignora la copia en disco y arranca desde el dataset limpio
go test ./...            # tests, incluidos los golden contra el contrato
```

Al arrancar, el servidor **larga carreras solo**. La primera sale a los 30 segundos; después sigue el calendario del dataset y repite el ciclo cada 2 horas. No hay que disparar nada a mano.

### Variables de entorno

| Variable | Por defecto | Qué hace |
|---|---|---|
| `PORT` | `8080` | |
| `JWT_SECRET` | *(uno de desarrollo)* | **Configuralo en producción.** El servidor avisa si no está. |
| `FRONT_URL` | `http://localhost:4200` | Base del enlace de verificación |
| `ALLOWED_ORIGINS` | *(vacío → `*`)* | Lista separada por comas |
| `RESEND_API_KEY` | *(vacío)* | Sin esto, los correos salen por el log |
| `MAIL_FROM` | `Hipódromo <no-reply@hipodromo.test>` | |
| `SNAPSHOT_PATH` | `data/snapshot.json` | Vacío desactiva la persistencia |
| `RESET` | *(vacío)* | Cualquier valor ignora la copia guardada |
| `LOG_LEVEL` | `info` | `debug` para ver el detalle del socket |

### Verificar una cuenta sin casilla de correo

Sin `RESEND_API_KEY`, el enlace de verificación se imprime en la consola:

```
  ┌─ VERIFICAR CUENTA ─────────────────────────────────────────
  │  nuevo@hipodromo.test
  │  http://localhost:4200/verificar?token=vt_8f3c1a90…
  └────────────────────────────────────────────────────────────
```

Se copia y se pega en el navegador. **No es un stub incompleto**: el flujo de verificación se puede demostrar en vivo, sin configurar nada.

---

## Cuentas del dataset

Contraseña para todas: `Carrera123!`

| Correo | Para qué sirve |
|---|---|
| `ana@hipodromo.test` | La de las demos. Historial mixto y apuestas pendientes. |
| `caro@hipodromo.test` | **Sin verificar.** Provoca `403 EMAIL_NOT_VERIFIED`. |
| `hugo@hipodromo.test` | Saldo 980. Provoca `409 INSUFFICIENT_BALANCE`. |

---

## Cómo está armado

```
main.go                  configuración, arranque y apagado ordenado
internal/
├── contract/            los tipos del contrato y el catálogo de errores
├── seed/                dataset embebido + rebase de fechas
├── store/               estado en memoria + copia en disco
├── auth/                PBKDF2, JWT HS256 y tokens opacos
├── mail/                emisor de log y emisor Resend
├── sim/                 el simulador de carrera
├── ws/                  hub de WebSocket, multiplexado por sala
├── program/             el calendario: larga carreras solo
└── api/                 rutas, middleware y handlers
```

### Decisiones que conviene conocer

**Sin base de datos.** El dataset entero pesa menos de 100 KB. El estado vive en memoria y se copia a un JSON en cada mutación, con escritura atómica. Si la ruta no se puede escribir, el servidor avisa una vez y sigue en memoria. Una app de enseñanza no debería necesitar levantar un Postgres para dar la primera clase.

**Router de la biblioteca estándar.** Desde Go 1.22, `http.ServeMux` entiende método y comodines: `GET /api/v1/races/{id}`. Un router externo solo agregaría una dependencia y una sintaxis más que explicar.

**JWT escrito a mano.** HS256 con `crypto/hmac` son unas cuarenta líneas que se leen enteras. `crypto/pbkdf2` es biblioteca estándar desde Go 1.24. Los refresh tokens son **opacos** a propósito: un JWT no se puede revocar, y estos son de un solo uso.

**El simulador está especificado, no improvisado.** `docs/contract/race-simulation.md` lo define, y `scripts/lib/race-sim.mjs` lo implementa en JavaScript para el mock del frontend. El test `TestSimulateReproducesFixture` prueba que las dos implementaciones son la misma: reproduce `race-ticks.jsonl` tick por tick. Eso es lo que permite **verificar** que la app se ve igual contra el mock y contra el servidor real.

**Las fechas se reubican al arrancar.** El dataset está anclado a un instante fijo y se desplaza por `now − ancla`. Las carreras terminadas siempre acaban de terminar y las que vienen siempre están por venir. La regla está en `docs/contract/README.md`.

---

## Tests

```bash
go test ./...              # todo
go test ./internal/sim/    # el simulador contra el fixture
go test -v ./internal/api/ # el detalle de cada caso de la API
```

Los dos que importan:

- **`TestSimulateReproducesFixture`** — Go y JavaScript producen la misma carrera, tick por tick.
- **`TestRespuestasCoincidenConLosSamples`** — la forma de cada respuesta coincide con `docs/contract/samples/`. Compara **nombres de campo**, no valores: los valores cambian con el rebase, los nombres son el contrato y son lo que rompe al frontend en silencio.

---

## Desplegar

```bash
docker build -t hipodromo-api .
docker run -p 8080:8080 -v hipodromo-data:/data \
  -e JWT_SECRET=... -e FRONT_URL=https://... -e ALLOWED_ORIGINS=https://... \
  hipodromo-api
```

Imagen distroless, binario estático, sin shell. El dataset va adentro del binario.

---

## Si algo no anda

| Síntoma | Qué mirar |
|---|---|
| El WebSocket cierra con **4001** | El access token venció (viven 15 min). Refrescar y reconectar. |
| El WebSocket cierra con **4029** | Más de 8 conexiones del mismo usuario. Suele ser un `SocketService` que no es singleton. |
| `RACE_ALREADY_STARTED` al apostar | La carrera larga a los 30 segundos del arranque. Apostar a la siguiente. |
| El servidor no arranca por la copia | La copia es de otra versión. `RESET=1 go run .` o borrar `data/snapshot.json`. |
| No llega el correo | Sin `RESEND_API_KEY` no se envía nada: el enlace está en la consola. |
