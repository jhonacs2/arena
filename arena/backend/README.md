# Arena — backend

Go 1.26, Postgres (Supabase). Una sola dependencia externa: `github.com/jackc/pgx/v5`
(más `github.com/coder/websocket` para las salas).

Las reglas están en [`../docs/contract/decisiones.md`](../docs/contract/decisiones.md);
la API en [`../docs/contract/api.md`](../docs/contract/api.md); el esquema en
[`../docs/contract/schema.sql`](../docs/contract/schema.sql). **El contrato manda.**

## Levantarlo

```bash
export DATABASE_URL="postgres://usuario:clave@host:5432/arena?sslmode=require"
export JWT_SECRET="$(openssl rand -hex 32)"
export ADMIN_USERNAME=profe ADMIN_PASSWORD="…"
go run .
```

El esquema **se aplica en cada arranque**: es idempotente, así que no hay un paso
de migración aparte que alguien pueda olvidarse de correr.

Sin `ADMIN_USERNAME`/`ADMIN_PASSWORD` no se crea el instructor, y sin instructor
no se pueden generar códigos: `invite_codes.created_by` referencia a un usuario y
no hay registro abierto por el que llegue el primero.

### Variables de entorno

| | Default | |
|---|---|---|
| `DATABASE_URL` | — | **requerida.** Cualquier parámetro de Postgres va acá |
| `JWT_SECRET` | uno de desarrollo, con warning | firma los access tokens |
| `PORT` | `8080` | |
| `ALLOWED_ORIGINS` | vacío = refleja el origen que vino | lista separada por comas |
| `DB_MAX_CONNS` | `10` | Supabase cobra conexiones |
| `DB_SIMPLE_PROTOCOL` | `false` | **ponelo en 1 con el pooler de transacciones de Supabase** (puerto 6543): ahí no hay sentencias preparadas |
| `MIGRATE` | `true` | `0` para no aplicar el esquema |
| `COOKIE_SECURE` | `true` | |
| `COOKIE_SAMESITE` | `none` | `lax` en producción, donde `/api` está en el mismo dominio |
| `COOKIE_DOMAIN` | vacío | |
| `ADMIN_USERNAME` · `ADMIN_PASSWORD` | vacío | crea o actualiza al instructor en el arranque |
| `ADMIN_FIRST_NAME` · `ADMIN_LAST_NAME` | `Instructor Arena` | |
| `LOG_LEVEL` | `info` | `debug` para ver cada migración |

## Verificar

```bash
gofmt -l .          # tiene que no imprimir nada
go vet ./...
go build ./...
go test ./...
```

### Los tests que necesitan Postgres

Las reglas que importan de Arena las hace cumplir la **base**: un lock de fila, un
`CHECK`, un trigger. Probarlas con un doble en memoria probaría el doble. Así que
los tests del canje, del ledger y de los handlers corren contra un Postgres real.

```bash
docker run -d --rm --name arena-pg -e POSTGRES_PASSWORD=arena \
  -e POSTGRES_DB=arena -p 55433:5432 postgres:16-alpine

ARENA_TEST_DATABASE_URL="postgres://postgres:arena@localhost:55433/arena?sslmode=disable" \
  go test ./...
```

**Sin la variable se saltean** en vez de fallar: `go test ./...` tiene que poder
correr en una máquina sin Docker. Los de lógica pura (contraseñas, JWT, alfabeto
de los códigos, formato de los puntos) corren siempre.

La variable es distinta de `DATABASE_URL` a propósito: los tests borran todas las
tablas, y un `go test` con la URL de producción en el entorno sería la peor forma
posible de aprender esta lección.

Los tres que hay que mirar si algo se rompe:

- **`TestRedeemConcurrenteGanaExactamenteUno`** — dos personas con el mismo código
  en el mismo instante: una gana, la otra recibe `CODE_ALREADY_REDEEMED`.
- **`TestAdminRechazaAUnAlumno`** — un alumno con un token válido recibe 403 en
  todas las rutas de `/api/admin/`. El rol se lee de la base, no del token.
- **`testdb.Reconcile`** — lo llama todo test que mueva monedas. Es
  `../scripts/reconcile.sql` como aserción: `users.balance` tiene que ser igual a
  la suma del ledger, y cada `balance_after` el anterior más su `delta`.

## Cómo está armado

```
main.go                  configuración por entorno y cableado
internal/
  api/                   router, middleware y handlers de auth, saldo y admin
  auth/                  PBKDF2 y JWT HS256, a mano y sin dependencias
  invite/                alfabeto sin caracteres ambiguos, formato AAAA-9999
  accounts/              usuarios, códigos, sesiones y la vista de nota
  ledger/                la ÚNICA puerta por la que se mueven monedas
  db/                    pool y runner de migraciones
  testdb/                base de prueba y la reconciliación como aserción
  contract/              sobre de error y catálogo de códigos
  races/ sim/ ws/        carreras, simulador y salas
```

### Tres cosas que conviene saber antes de tocar nada

**El ledger es una sola puerta.** `ledger.Move` recibe el ejecutor —una `pgx.Tx` o
el pool— así que el descuento de una apuesta y la fila de la apuesta caen en la
misma transacción. Actualiza `users.balance` **y** escribe en `coin_transactions`,
en ese orden y nunca al revés: el `UPDATE` toma el lock de la fila del usuario, y
eso es lo que hace que dos movimientos simultáneos del mismo alumno salgan
ordenados en la secuencia de ids. Al revés, la cadena de `balance_after` se
rompería sin que nada falle.

**El rol no viaja en el token.** El JWT dice quién es; qué puede hacer lo dice
`users.role`, leído en cada petición. Bajarle el rol a alguien surte efecto en la
petición siguiente y no cuando venza su token. Además de eso, `withAdminGate`
rechaza todo `/api/admin/*` por prefijo, antes de llegar al handler: los handlers
igual chequean, y la redundancia es a propósito — el riesgo real no es que este
chequeo esté mal escrito, es que alguien agregue mañana un endpoint y se olvide.

**Los puntos salen de la vista `user_scores`, no de Go.** La fórmula
`floor(monedas/100) + regalados` vive en un solo lugar. Si viviera en dos, un día
se separarían y la nota dependería de por dónde se la mire.

Ya pagó: la fórmula tuvo un piso de 10 puntos y lo perdió mientras este backend se
escribía. No hubo una línea de Go que cambiar — sólo tests que asertaban el valor
viejo.

### Enganchar rutas de otros paquetes

`api.Server.ExtraRoutes` es el gancho, y se llama desde `main.go`:

```go
server.ExtraRoutes = append(server.ExtraRoutes, raceHandlers.Register)
```

Se registran en el mismo `mux`, así que pasan por la misma cadena de middleware
—incluido el portón de rol de `/api/admin/`—. La identidad de la petición sale de
`server.Identity(r)` y la del socket de `server.IdentityFromToken(ctx, token)`.
