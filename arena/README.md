# Arena — levantarlo en local

Arena es la app **en vivo** del módulo: los alumnos canjean un código de invitación,
apuestan monedas en carreras que abrís vos, y esas monedas valen nota. Esta guía es
cómo correrla en tu máquina.

Las reglas del juego no están acá — están en
[`docs/contract/decisiones.md`](docs/contract/decisiones.md), que es la fuente de
verdad. Acá solo están los comandos.

| Pieza | Carpeta | Corre en |
|---|---|---|
| Frontend | `frontend/` | Angular 22 · `http://localhost:4200` |
| Backend | `backend/` | Go 1.26 · `http://localhost:8099` |
| Base | — | Postgres 17 en Docker · `localhost:5432` |

---

## Elegí un camino

**Hay dos, y el primero alcanza para la mayoría de las cosas.**

| | Necesitás | Sirve para |
|---|---|---|
| **A · solo el frontend** | Node | ver la app entera, cambiar pantallas, mostrarla |
| **B · todo conectado** | Node + Docker | probar el backend de verdad, apostar con datos que sobreviven |

En el camino A la app funciona **completa** —registro, apuestas, carrera en vivo a
10 Hz, panel del instructor— contra un backend de mentira que vive en el navegador.
No es una maqueta: responde con las mismas formas y los mismos códigos de error que
el Go real, porque los dos salen del mismo contrato.

Si estás tocando pantallas, quedate en A. Cambiar de A a B es **una línea**.

---

## Camino A — solo el frontend

```bash
cd arena/frontend
npm install
npm start
```

Abrí **http://localhost:4200**. Listo, no hace falta nada más.

Entrá por `/ingresar` con cualquiera de estas dos cuentas — la contraseña es
`arena1234` en las dos:

| Usuario | Rol | Qué ves |
|---|---|---|
| `profe` | instructor | el panel de `/instructor`: crear códigos y carreras |
| `anag` | alumna | el tablero con saldo, historial y carreras para apostar |

Y para probar la pantalla de registro, hay tres códigos sembrados:

| Código | Qué pasa |
|---|---|
| `AVBD-1234` | sin usar — se canjea bien |
| `TXNQ-4562` | sin usar |
| `KMPR-8827` | ya canjeado — muestra el error de «este código ya se usó» |

Los dos últimos existen para ver que los errores son **distintos**: «lo escribí mal»
y «ya me registré» no son el mismo problema para el alumno.

> **El mundo del mock vive en memoria.** Al recargar la página, `profe` y `anag`
> siguen ahí porque tienen ids fijos; las cuentas que creaste en la corrida, no.

---

## Camino B — todo conectado

Tres pasos: la base y el backend en Docker, y el frontend apuntando ahí.

### 1. El backend y la base

```bash
cd arena/deploy
cp .env.example .env       # si todavía no existe
```

> **`arena/deploy/.env` ya está creado en esta máquina**, con valores de
> desarrollo listos para usar. Si está, saltá al `docker compose` de abajo.

Abrí `.env` y completá **dos** variables — el resto ya viene con valores de
desarrollo:

```bash
JWT_SECRET=…          # generalo, no lo inventes:  openssl rand -base64 48
ADMIN_PASSWORD=…      # mínimo 8 caracteres, o el backend no arranca
```

### Con qué usuario entrás

**No hay ninguna cuenta hasta que la creás vos con estas dos variables:**

```bash
ADMIN_USERNAME=instructor
ADMIN_PASSWORD=Prueba1234!     # ← lo que está puesto hoy en el .env local
```

Esa es tu cuenta de instructor, y es la que usás en `/ingresar`. **Arena no tiene
registro abierto**: un alumno solo entra canjeando un código, los códigos los genera
un instructor, y `invite_codes.created_by` referencia a un usuario. Sin este admin
en la base nadie puede generar el primero y la app queda cerrada para todos,
incluido vos.

Para cambiar la contraseña, editala en `.env` y reiniciá con `docker compose up -d`:
el arranque hace un *upsert* que **pisa la contraseña vieja**. Es también cómo
recuperás el acceso si la olvidás a mitad de la cursada — para el instructor no hay
«olvidé mi contraseña».

> **No las confundas con las del camino A.** `profe` y `anag` con `arena1234` solo
> existen en el mock del navegador; `instructor` solo existe en Postgres. Si una no
> te funciona, estás en el otro camino.

```bash
docker compose up -d --build
```

La primera vez tarda un par de minutos —compila el Go dentro de la imagen—; las
siguientes son segundos. Comprobalo:

```bash
docker compose logs api --tail 5
```

Tiene que decir `esquema al día`, `instructor listo` y `escuchando`. **El esquema se
aplica solo en cada arranque**: es idempotente, así que no hay un paso de migración
que alguien se pueda olvidar de correr.

### 2. Apuntar el frontend al backend real

En `frontend/src/environments/environment.ts`, cambiá **un booleano**:

```ts
export const environment = {
  apiBaseUrl: '/api',
  useMockBackend: false,   // ← estaba en true
} as const;
```

Es la única línea. Ni un componente sabe si atrás hay un mock o un Go.

### 3. El frontend

```bash
cd arena/frontend
npm start
```

**http://localhost:4200**, y entrás con `instructor` y la contraseña que pusiste en
`.env`.

`npm start` levanta un proxy que manda todo lo que empiece con `/api` —incluido el
WebSocket de la carrera— al backend del puerto 8099. Está en
[`frontend/proxy.conf.json`](frontend/proxy.conf.json) y **no es opcional**: la app
pide `/api` relativo y arma la URL del socket con el host de la página, así que sin
el proxy el frontend se llamaría a sí mismo. Es también lo que hace que el navegador
vea un solo origen, que es lo que necesita la cookie de sesión.

---

## La primera vuelta completa

Con el camino B andando, así se recorre entero en cinco minutos:

1. **`/instructor` → Códigos.** Generá 3 con 1000 monedas cada uno. Copiá uno.
2. **Ventana de incógnito → `/registro`.** Pegá el código, poné nombre, apellido,
   usuario y contraseña. Entrás con 1000 monedas y **10 puntos**.
3. **Volvé a `/instructor` → Carreras.** Creá una con 3 caballos y **abrila**.
4. **En la ventana del alumno**, entrá a la carrera y apostá.
5. **Largala** desde el panel. La simulación dura unos **45 segundos** y va a 10 Hz.
6. **`/instructor` → Notas.** Ahí está el saldo y el puntaje de cada uno.

Para ver la economía completa hacen falta **dos alumnos apostando a caballos
distintos** — el pago es pari-mutuel: lo que pierden unos es exactamente lo que
cobran los otros. Con un solo apostador siempre se cobra a sí mismo y no se ve nada.

---

## Comprobar que quedó bien

### La prueba que importa

```bash
ARENA_E2E_URL=http://localhost:8099 \
ARENA_E2E_USER=instructor \
ARENA_E2E_PASSWORD='la-de-tu-.env' \
  node arena/scripts/e2e.mjs
```

Recorre código → registro → apuesta → carrera → nota **contra el binario real**, y
tarda un minuto porque espera a que la carrera termine de verdad. Es lo único que
prueba que el cableado de `main.go` está enchufado: los tests de Go pasan en verde
aunque el módulo de carreras esté desconectado.

Deja tres alumnos de prueba en la base. En local no molesta; contra producción, sí.

### Todo lo demás

```bash
DATABASE_URL="postgres://arena:arena@localhost:5432/arena?sslmode=disable" \
  node scripts/verify-arena.mjs
```

Son 30 verificaciones: que el esquema aplique limpio dos veces seguidas, que el
ledger reconcilie contra los saldos, que ningún handler de `/admin/` esté sin
chequeo de rol, que no haya `float` en un monto, el contraste AA de la paleta.
**Sin `DATABASE_URL` las que tocan la base se saltean** en vez de fallar.

### Los tests de Go

```bash
docker run -d --rm --name arena-pg-test -e POSTGRES_PASSWORD=test \
  -e POSTGRES_DB=arena_test -p 55433:5432 postgres:16-alpine

cd arena/backend
ARENA_TEST_DATABASE_URL="postgres://postgres:test@localhost:55433/arena_test?sslmode=disable" \
  go test ./...
```

**La base tiene que llamarse `arena_test`, o algo con «test» adentro.** No es una
convención: los tests truncan todas las tablas y el ayudante se niega a arrancar
si el nombre no lo tiene. La barrera existe porque el accidente ya ocurrió —
apuntar esta variable a la base del compose se llevó puestos los datos de
desarrollo, con los tests en verde y sin un solo aviso.

---

## Trabajar sobre el backend

Con Docker el ciclo es lento: cada cambio en Go implica reconstruir la imagen. Si
vas a tocar el backend, dejá **solo la base** en Docker y corré el Go a mano:

```bash
cd arena/deploy && docker compose up -d db      # solo la base

cd ../backend
PORT=8099 \
DATABASE_URL="postgres://arena:arena@localhost:5432/arena?sslmode=disable" \
JWT_SECRET="cualquier-cosa-en-local" \
ADMIN_USERNAME=instructor ADMIN_PASSWORD='Prueba1234!' \
COOKIE_SECURE=false COOKIE_SAMESITE=lax \
  go run .
```

`PORT=8099` para que el frontend lo encuentre sin tocar el proxy. `Ctrl+C`,
`go run .` otra vez, y listo.

---

## Cuando algo no anda

Estas son las que te van a pasar, con el error tal como lo vas a ver.

### `ports are not available: … 127.0.0.1:8080: bind: Only one usage…`

Algo ya ocupa ese puerto. **En tu máquina el 8080 lo tiene NVIDIA Broadcast**, y por
eso el compose publica en **8099**. Si 8099 también te queda ocupado, cambiá
`API_PORT` en `.env` **y** el `target` de `frontend/proxy.conf.json`: son los dos
extremos del mismo cable y se cambian juntos.

Para ver quién lo tiene:

```powershell
Get-NetTCPConnection -LocalPort 8099 -State Listen |
  ForEach-Object { Get-Process -Id $_.OwningProcess }
```

### El frontend tira 404 en todo lo que pide

Estás en el camino B con `useMockBackend: false` y el backend apagado, o el proxy
apunta a un puerto donde no hay nadie. `docker compose ps` y `curl
http://localhost:8099/api/races` — si eso responde `401`, el backend está vivo (te
falta el token, que es correcto).

### Entro bien pero al recargar la página me echa

Es la cookie de sesión. En local tiene que ser `COOKIE_SAMESITE=lax` con
`COOKIE_SECURE=false`. Si alguien la pone en `none` sin TLS, **el navegador descarta
la cookie en silencio**: no hay error en la consola, simplemente no se guarda.

### Cambié `schema.sql` y no se ve el cambio

El esquema se aplica al **crear** el volumen de Postgres. Para empezar de cero:

```bash
cd arena/deploy
docker compose down -v      # el -v borra la base. Se van todos los datos.
docker compose up -d --build
```

### Toqué el backend y sigue respondiendo lo de antes

Quedó un proceso viejo escuchando. En Windows `pkill` no existe:

```powershell
Get-NetTCPConnection -LocalPort 8099 -State Listen |
  ForEach-Object { Stop-Process -Id $_.OwningProcess -Force }
```

### `ADMIN_PASSWORD necesita al menos 8 caracteres`

Literal: el backend no arranca con una contraseña corta. Está en `.env`.

---

## Apagar

```bash
cd arena/deploy
docker compose down       # para todo, la base queda
docker compose down -v    # para todo y BORRA la base
```

El frontend se corta con `Ctrl+C`.

---

## Lo que no es esto

Esta guía es **local**. El despliegue de verdad —el VPS, el túnel de Cloudflare, el
servicio de systemd— está en [`deploy/README.md`](deploy/README.md), y ahí las
reglas son otras: el backend **no expone puerto**, `cloudflared` abre la conexión
hacia afuera, y las cookies van con `Secure`.
