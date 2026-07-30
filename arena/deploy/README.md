# Arena — despliegue

Procedimiento completo, en orden, sin adivinar. Está escrito para seguirse el día
del despliegue: cada paso dice qué se crea, dónde, y qué valor te vas a llevar al
paso siguiente.

```
   navegador
      │  https://arena.ejemplo.com
      ▼
 ┌─────────────────────────── Cloudflare ───────────────────────────┐
 │                                                                  │
 │  Pages  ──────────────────▶  el frontend Angular (estático)       │
 │                                                                  │
 │  Worker  ─── /api/** ────▶  api.arena.ejemplo.com                 │
 │                                    │                              │
 └────────────────────────────────────┼──────────────────────────────┘
                                      │  el túnel, abierto desde adentro
 ┌────────────────────────── VPS Hostinger ──────────────────────────┐
 │   cloudflared  ──▶  127.0.0.1:8080  ──▶  arena-api (contenedor)    │
 └────────────────────────────────────┼──────────────────────────────┘
                                      │  TLS
                              Supabase (Postgres)
```

**Ningún puerto de entrada.** El VPS no publica nada: `cloudflared` abre la
conexión hacia Cloudflare desde adentro y el tráfico vuelve por ahí. El firewall
del VPS puede quedar cerrado a todo lo entrante salvo SSH.

> ### El túnel oculta y protege, pero **no autentica**
>
> Que no haya puerto abierto significa que nadie puede escanear el VPS ni llegarle
> por IP: eso es real y es valioso. Lo que **no** hace es decidir quién puede
> hacer qué. Cualquiera que conozca la URL pública llega al backend igual que un
> alumno. La seguridad de verdad son los **JWT** y la **validación de rol en cada
> handler** —`arena/CLAUDE.md` §4—, y por eso `node scripts/verify-arena.mjs`
> falla si aparece una ruta `/admin/` sin chequeo de rol. Ocultar una URL no es un
> control de acceso.

---

## ⚠️ Antes que nada: Supabase se pausa a los 7 días

**El plan gratuito de Supabase pausa el proyecto tras 7 días sin actividad.** Un
curso semanal cae exactamente en esa ventana: si la clase es el martes y nadie
toca la app hasta el martes siguiente, el proyecto puede estar pausado justo
cuando lo necesitás. Reanudarlo es manual, desde el panel, y **tarda unos
minutos** — minutos que no tenés cinco antes de empezar.

Dos salidas, las dos válidas:

1. **Un ping programado.** Cualquier consulta a la base cuenta como actividad. Un
   cron en el VPS cada 12 horas alcanza:

   ```
   0 */12 * * * curl -s -o /dev/null -X POST https://api.arena.ejemplo.com/api/auth/check-code \
     -H 'Content-Type: application/json' -d '{"code":"ZZZZ-9999"}'
   ```

   o, sin depender del VPS, un Cron Trigger de Cloudflare Workers que haga la
   misma llamada una vez al día.

   **Por qué ese endpoint y no un `/health`:** tiene que *tocar la base*. Un
   handler que responde sin consultar no cuenta como actividad para Supabase, y
   entonces el ping no sirve para nada. `POST /api/auth/check-code` es público, no
   necesita token, y hace un `select` sobre `invite_codes`. Con un código
   inexistente devuelve `404 CODE_NOT_FOUND`, que es la respuesta correcta y
   confirma que la cadena entera —túnel, backend, base— está viva.

   > Si preferís un `GET /api/health` de verdad, es mejor práctica y da un
   > diagnóstico más claro. Pero **primero se agrega a
   > `arena/docs/contract/api.md`**: la regla de este repo es contrato primero, y
   > hoy ese endpoint no existe.

2. **El plan pago.** El plan Pro no pausa proyectos. Si Arena va a sostener la
   nota de una cursada entera, es la opción tranquila.

Mientras se decide, **queda escrito acá** y hay que revisarlo el día del
despliegue. La otra consecuencia del plan gratuito: los backups son limitados, así
que conviene un `pg_dump` propio después de cada clase (ver el final).

---

## Lo que necesitás antes de empezar

| | |
|---|---|
| Un dominio en Cloudflare | con el DNS ya administrado por Cloudflare (nameservers apuntando ahí) |
| Un VPS de Hostinger | Linux con root y systemd. Un plan chico alcanza |
| Una cuenta de Supabase | el proyecto se crea en el paso 1 |
| `git` y Docker en el VPS | `apt install docker.io` en Ubuntu |

En tu máquina, para verificar: Node 22, Go 1.26 y —opcional pero recomendado—
`psql`.

---

## Paso 1 · Supabase: la base

1. **Creá el proyecto.** Panel de Supabase → *New project*. Elegí la región **más
   cercana al VPS**: cada consulta paga ese viaje, y con la simulación corriendo a
   10 Hz la latencia se nota. Guardá la contraseña de la base que te muestra: **no
   se vuelve a mostrar.**

2. **Aplicá el esquema.** Del panel, *SQL Editor* → pegá
   `arena/docs/contract/schema.sql` → *Run*. O desde tu máquina, con la cadena de
   conexión del paso 3:

   ```bash
   psql "$DATABASE_URL_SESSION" -v ON_ERROR_STOP=1 -f arena/docs/contract/schema.sql
   ```

   **Corrélo dos veces.** El esquema es idempotente y la segunda pasada tiene que
   terminar sin un solo error. Si falla, el esquema está mal, no la base — eso es
   exactamente lo que verifica `node scripts/verify-arena.mjs esquema`.

3. **La cadena de conexión.** Panel → *Connect* (arriba, junto al nombre del
   proyecto). Vas a ver tres:

   | | Forma | Para qué |
   |---|---|---|
   | **Pooler, transaction** | `…pooler.supabase.com:6543` | **el backend** |
   | Pooler, session | `…pooler.supabase.com:5432` | migraciones, `psql` a mano |
   | Directa | `db.<ref>.supabase.co:5432` | solo si el VPS tiene IPv6 |

   Reemplazá `[YOUR-PASSWORD]` por la contraseña del paso 1 y agregá
   `?sslmode=require`. Eso es `DATABASE_URL`. La de *session* guardala aparte como
   `DATABASE_URL_SESSION`: es la que usás para aplicar el esquema y para correr
   `reconcile.sql`.

### Por qué el pooler en modo *transaction*, y no la conexión directa

Tres razones, en orden de cuánto te van a arruinar el día:

- **La conexión directa es IPv6 en los proyectos nuevos.** `db.<ref>.supabase.co`
  resuelve a un AAAA y nada más; el IPv4 es un add-on pago. Un VPS de Hostinger
  puede no tener IPv6 de salida, y ahí la conexión directa simplemente no se
  establece. El pooler es alcanzable por IPv4. Este solo punto ya decide.
- **Postgres cobra caro cada conexión.** Un proyecto chico de Supabase tolera
  pocas decenas de conexiones simultáneas. En modo *transaction* el pooler
  devuelve la conexión al montón **al terminar cada transacción**, así que las
  conexiones que tu pool tiene abiertas e inactivas no inmovilizan un backend de
  Postgres del otro lado. En modo *session* la conexión queda tomada mientras el
  cliente la tenga abierta, que para un backend con pool propio es *siempre*.
- **Tu backend ya tiene su propio pool.** No necesita que el pooler le mantenga
  estado de sesión: abre transacción, hace lo suyo, cierra. Es el caso de uso
  exacto del modo *transaction*.

**Lo que cuesta**, y hay que saberlo antes de que se manifieste como un error
raro: en modo *transaction* no hay *prepared statements* cacheados entre
transacciones, ni `LISTEN`/`NOTIFY`, ni `SET` que sobreviva, ni advisory locks de
sesión. Por eso, **con el pooler en modo transaction hay que poner**:

```
DB_SIMPLE_PROTOCOL=true
```

Sin eso, la primera consulta que se repita falla con `prepared statement
"lrupsc_1" already exists` — un error que no dice nada sobre su causa. Con
conexión directa o pooler de sesión se deja en `false`, que es más rápido.

Nada de lo demás limita a Arena: la difusión a la sala es **en proceso** —la
simulación es autoritativa del servidor y vive en el backend, `decisiones.md`
§3—, así que `LISTEN`/`NOTIFY` no hace falta. Y `DB_MAX_CONNS=10` es de sobra
para 30 alumnos: el cuello está en la latencia, no en el paralelismo.

Para el esquema y para `reconcile.sql`, usá la de **session** o el SQL Editor.

---

## Paso 2 · El VPS: el backend

1. **Traé el código y construí la imagen.**

   ```bash
   git clone <repo> /opt/arena && cd /opt/arena
   docker build -f arena/deploy/Dockerfile -t arena-api:latest arena
   ```

   La imagen final es `distroless`: no tiene shell ni gestor de paquetes, y el
   proceso corre como `nonroot`.

2. **El archivo de entorno.**

   ```bash
   sudo mkdir -p /etc/arena
   sudo cp arena/deploy/.env.example /etc/arena/arena.env
   sudo chmod 600 /etc/arena/arena.env
   sudo nano /etc/arena/arena.env
   ```

   Completá, como mínimo:

   | Variable | Valor |
   |---|---|
   | `DATABASE_URL` | la del pooler en modo *transaction*, del paso 1 |
   | `DB_SIMPLE_PROTOCOL` | **`true`** con ese pooler. Ver el paso 1 |
   | `JWT_SECRET` | `openssl rand -base64 48` — **generalo, no lo inventes** |
   | `ADMIN_USERNAME` / `ADMIN_PASSWORD` | la cuenta de instructor que se crea al arrancar |
   | `ALLOWED_ORIGINS` | `https://arena.ejemplo.com` |
   | `COOKIE_SAMESITE` | `lax` con la opción A del paso 4; `none` con la B |
   | `COOKIE_SECURE` | `true`. Siempre, fuera de localhost |
   | `ARENA_IMAGE` | `arena-api:latest` |

   El resto tiene un valor por defecto razonable, y cada variable está explicada
   en `.env.example`.

3. **El servicio.**

   ```bash
   sudo cp arena/deploy/arena-api.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable --now arena-api
   journalctl -u arena-api -f
   ```

4. **Probalo desde el propio VPS**, que es el único lugar desde donde se puede:

   ```bash
   curl -si -X POST http://127.0.0.1:8080/api/auth/check-code \
     -H 'Content-Type: application/json' -d '{"code":"ZZZZ-9999"}'
   ```

   Tiene que contestar `404` con `{"error":{"code":"CODE_NOT_FOUND",…}}`. Eso
   prueba tres cosas de una: el proceso escucha, la base responde y el sobre de
   error es el del contrato. **No** intentes alcanzarlo desde
   afuera: no se puede, y está bien que no se pueda. El firewall del VPS puede
   quedar `default deny incoming` con SSH como única excepción; el túnel del paso
   3 no necesita que se habilite nada entrante.

---

## Paso 3 · Cloudflare Tunnel

Todo esto se hace **en el VPS**.

1. **Instalá `cloudflared`.**

   ```bash
   curl -L --output /tmp/cloudflared.deb \
     https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
   sudo dpkg -i /tmp/cloudflared.deb
   sudo useradd --system --no-create-home --shell /usr/sbin/nologin cloudflared
   ```

2. **Autenticá y creá el túnel.**

   ```bash
   cloudflared tunnel login      # abre una URL: elegí tu dominio
   cloudflared tunnel create arena
   ```

   El segundo comando imprime **el UUID del túnel** y deja un
   `~/.cloudflared/<UUID>.json` — las credenciales. Los dos valores van en la
   config del paso siguiente.

3. **Instalá la configuración.**

   ```bash
   sudo mkdir -p /etc/cloudflared
   sudo cp ~/.cloudflared/<UUID>.json /etc/cloudflared/
   sudo cp /opt/arena/arena/deploy/cloudflared/config.yml /etc/cloudflared/
   sudo nano /etc/cloudflared/config.yml     # reemplazar los tres «REEMPLAZAR»
   sudo chown -R cloudflared:cloudflared /etc/cloudflared
   sudo chmod 600 /etc/cloudflared/<UUID>.json
   ```

4. **Publicá el hostname.** Esto crea el registro DNS por vos:

   ```bash
   cloudflared tunnel route dns arena api.arena.ejemplo.com
   ```

5. **Levantá el servicio.**

   ```bash
   sudo cp /opt/arena/arena/deploy/cloudflared/cloudflared.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable --now cloudflared
   systemctl status cloudflared
   ```

6. **Verificá desde tu máquina**, no desde el VPS:

   ```bash
   curl -si -X POST https://api.arena.ejemplo.com/api/auth/check-code \
     -H 'Content-Type: application/json' -d '{"code":"ZZZZ-9999"}'
   ```

   El mismo `404 CODE_NOT_FOUND` del paso anterior, ahora entrando por Cloudflare.
   Si responde, el túnel está andando y el backend sigue sin tener un solo puerto
   de entrada.

---

## Paso 4 · Cloudflare Pages: el frontend

1. **Construí.**

   ```bash
   cd arena/frontend && npm ci && npm run build
   ```

2. **Creá el proyecto.** Panel de Cloudflare → *Workers & Pages* → *Create* →
   *Pages*. Conectá el repo o subí el directorio con `npx wrangler pages deploy`.

   | | |
   |---|---|
   | Build command | `npm ci && npm run build` |
   | Output directory | `dist/arena/browser` |
   | Root directory | `arena/frontend` |

3. **El dominio.** *Custom domains* → `arena.ejemplo.com`.

### Y ahora `/api` en el mismo dominio

Un hostname **no puede ser al mismo tiempo** un proyecto de Pages y la punta de un
túnel: el registro DNS es uno solo. Por eso el túnel publica
`api.arena.ejemplo.com`. Para que el frontend pueda llamar a `/api` en su propio
dominio hay dos caminos:

**Opción A — un Worker que reenvía `/api/**` (recomendada).**

```bash
cd arena/deploy/cloudflare
# reemplazar los hostnames en wrangler.toml
npx wrangler deploy
```

La ruta del Worker gana sobre el dominio de Pages, así que Pages sigue sirviendo
todo lo que no sea `/api/*`. Se elige esta por una razón concreta: **la cookie de
refresh**. Es `HttpOnly` y de un solo uso (`api.md`). Same-origin funciona con
`SameSite=Lax` y listo. Cruzando dominios necesitás `SameSite=None; Secure`, que
es precisamente la cookie que un navegador con bloqueo de terceros descarta — y
ahí el alumno se desloguea cada 15 minutos en medio de la clase. Además: cero
CORS que depurar.

**Opción B — el frontend llama a `api.arena.ejemplo.com` directo.** Un archivo
menos, pero te comprás el CORS (`ALLOWED_ORIGINS` bien puesto) y la cookie de
tercera parte con todo lo que eso implica. Si vas por acá, probá el refresh en
Safari y en Chrome con el bloqueo de terceros activado **antes** de la primera
clase.

El Worker no autentica: reenvía tal cual, con `Authorization` y cookies intactas.
Quién puede qué lo sigue decidiendo el backend.

---

## Paso 5 · Verificar

Desde la raíz del repo:

```bash
node scripts/verify-arena.mjs
```

Y contra la base de verdad, que es la parte que no se puede simular:

```bash
DATABASE_URL="$DATABASE_URL_SESSION" node scripts/verify-arena.mjs esquema
```

Eso aplica el esquema **dos veces** —tiene que quedar limpio las dos— y corre la
reconciliación del ledger: `sum(coin_transactions.delta)` contra `users.balance`,
por usuario, **cero diferencias**. Si sale una fila, hay una nota que el ledger no
explica y eso se arregla antes de dar la clase. A mano:

```bash
psql "$DATABASE_URL_SESSION" -f arena/scripts/reconcile.sql
```

Nunca se corrige editando el ledger —el trigger lo impide— sino agregando un
movimiento `adjustment` que compense.

### Checklist del día

- [ ] El proyecto de Supabase **no está pausado** (mirarlo el día anterior, no el mismo día)
- [ ] `systemctl status arena-api cloudflared` — los dos `active (running)`
- [ ] Un `POST` a `https://arena.ejemplo.com/api/auth/check-code` contesta `CODE_NOT_FOUND`, desde una red que no sea la del VPS
- [ ] `node scripts/verify-arena.mjs` en verde
- [ ] La reconciliación del ledger da cero filas
- [ ] Entrás con la cuenta de instructor y podés generar un código
- [ ] Un código de prueba se canjea y acredita 1000 monedas

---

## Local, con Postgres descartable

Sin Supabase y sin túnel:

```bash
cd arena/deploy
cp .env.example .env      # completar JWT_SECRET
docker compose up --build
```

Levanta Postgres con el esquema ya aplicado y el backend en
`http://127.0.0.1:8080`. El frontend, aparte, con `npm start` en
`arena/frontend`.

Para volver a aplicar el esquema después de cambiarlo —sin borrar los datos, que
es la prueba que importa:

```bash
docker compose exec -T db psql -U arena -d arena -v ON_ERROR_STOP=1 \
  < ../docs/contract/schema.sql
```

Y para empezar de cero: `docker compose down -v`.

---

## Mantenimiento

**Backup después de cada clase.** Los backups del plan gratuito son limitados y
las monedas son nota:

```bash
pg_dump "$DATABASE_URL_SESSION" --no-owner --no-privileges -Fc \
  -f "arena-$(date +%F).dump"
```

**Actualizar el backend:**

```bash
cd /opt/arena && git pull
docker build -f arena/deploy/Dockerfile -t arena-api:latest arena
sudo systemctl restart arena-api
```

`cloudflared` no hace falta reiniciarlo: reconecta solo. Hay unos segundos de
corte, así que no lo hagas con una carrera `running`.

**Rotar `JWT_SECRET`** invalida todas las sesiones abiertas. Se puede, pero no en
medio de una clase.

---

## Cuando algo no anda

| Síntoma | Dónde mirar |
|---|---|
| `curl` al hostname público da 502 | `journalctl -u cloudflared -n 50`. Casi siempre el `service:` del ingress no coincide con `PORT` |
| Un 404 **sin** el sobre JSON de error | ese 404 lo puso el catch-all del túnel, no el backend: revisá el `hostname:` del ingress. El 404 del backend siempre trae `{"error":{"code":…}}` |
| El frontend carga pero `/api` da 404 | la ruta del Worker (opción A) o el `_routes` de Pages |
| El backend no arranca | `journalctl -u arena-api -n 50`. Casi siempre `DATABASE_URL` o `JWT_SECRET` |
| `connection refused` a la base | ¿proyecto pausado? ¿usaste el pooler y no la directa? |
| `prepared statement "lrupsc_1" already exists` | estás en el pooler en modo *transaction* sin `DB_SIMPLE_PROTOCOL=true` |
| El alumno se desloguea cada 15 minutos | la cookie de refresh: mirá la opción A del paso 4 |
| El esquema falla en la segunda pasada | un `create` sin `if not exists`. `node scripts/verify-arena.mjs esquema` te dice cuál |
