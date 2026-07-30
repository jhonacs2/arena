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
 │                                                                   │
 │   cloudflared ──▶ 127.0.0.1:18099 ──▶ arena-api ──▶ arena-db      │
 │                                          (contenedores, red privada) │
 │                                                                   │
 │   …y en la misma máquina, WorkAdventure con sus propios puertos    │
 └───────────────────────────────────────────────────────────────────┘
```

**Ningún puerto de entrada.** El VPS no publica nada de Arena hacia afuera:
`cloudflared` abre la conexión hacia Cloudflare desde adentro y el tráfico vuelve
por ahí. El backend sale a `127.0.0.1:18099` —solo loopback— y **la base no sale a
ninguna parte**: vive en la red privada del compose, donde el único que le habla
es el backend.

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

## ⚠️ Antes que nada: el VPS es compartido

En esta máquina ya corre **WorkAdventure**. Arena no puede comportarse como si
fuera la única cosa instalada, y eso cambia tres decisiones concretas:

| | Qué hace Arena | Por qué |
|---|---|---|
| **Puertos** | el backend en `127.0.0.1:18099`; la base, en ninguno | los 80/443 son de WorkAdventure y Arena no los necesita: el tráfico entra por el túnel. El 8080 y el 5432 son los primeros que alguien más ya ocupó |
| **Memoria** | techo explícito por contenedor, ~450 MB entre los dos | sin `mem_limit`, Postgres se cree dueño de la RAM de la máquina y el que muere por OOM es el vecino |
| **Disco** | logs con techo (`max-size: 10m`, 3 archivos por contenedor) | un contenedor sin límite de log escribe hasta llenar el disco, y el disco también es compartido |

**Comprobá qué hay ocupado antes de empezar**, así el número de puerto lo elegís
mirando y no de memoria:

```bash
sudo ss -ltnp                      # todo lo que escucha, con el proceso
docker ps --format 'table {{.Names}}\t{{.Ports}}'
free -m                            # cuánta RAM libre queda de verdad
```

Si `18099` aparece ocupado, elegí otro por encima de `10000` que no salga en esa
lista, y cambialo en **los dos lados**: `API_PORT` en `/etc/arena/arena.env` y el
`service:` de `cloudflared/config.yml`. Son los dos extremos del mismo cable.

### Cuánta memoria necesita Arena, y por qué esa

Los números están calculados para **30 alumnos**, que es el tamaño real del curso.
No son un default copiado.

| Contenedor | Techo (`mem_limit`) | Medido | Qué lo ocupa |
|---|---|---|---|
| `arena-db` | **256 MB** | 33 MB en reposo · 39 MB después de una carrera | 64 MB de `shared_buffers` reservados de a poco + un proceso por conexión (hasta 25) + autovacuum |
| `arena-api` | **192 MB** | 2 MB en reposo · 4 MB después de una carrera | un binario de Go, los WebSockets de la sala y un pool de 10 conexiones |

> Los «medido» salen de levantar este mismo compose y correrle
> `node arena/scripts/e2e.mjs` entero: tres alumnos, una carrera de 45 segundos y
> la liquidación. Con 30 alumnos en vez de 3 el backend suma sus WebSockets y la
> base sus conexiones, pero no cambia de orden de magnitud. **El techo está muy por
> encima a propósito**: lo que no querés es que el kernel mate un contenedor a
> mitad de clase por haber afinado el número al milímetro.

Los dos tienen `memswap_limit` igual al `mem_limit`: **swap prohibido**. Si le
dejás usar swap, la base no muere, se arrastra — y una carrera a 10 Hz contra una
base que swapea se ve exactamente como la app colgada, sin un solo error en los
logs.

El backend además lleva `GOMEMLIMIT=150MiB`. Sin eso, Go no sabe que está en un
cgroup, deja crecer el heap hasta el techo del contenedor y el kernel lo mata: el
backend desaparece en medio de una carrera, sin panic y sin nada en el log que lo
explique.

> **Al mirar `docker stats`, no te asustes.** Buena parte de lo que cuenta como
> memoria de `arena-db` es caché de páginas de los archivos de la base, y esa
> caché el kernel la suelta antes de matar a nadie. Un número pegado al techo no
> significa que falte memoria; lo que significaría eso es un `OOMKilled` en
> `docker inspect arena-db --format '{{.State.OOMKilled}}'`.

Si algún día son 60 en vez de 30: subí `DB_MEM_LIMIT` a `512m`, `PG_SHARED_BUFFERS`
a `128MB` (el 25%, la regla de siempre) y `DB_MAX_CONNS` a `20` — y `PG_MAX_CONNECTIONS`
tiene que quedar por encima de ese último.

---

## Lo que necesitás antes de empezar

| | |
|---|---|
| Un dominio en Cloudflare | con el DNS ya administrado por Cloudflare (nameservers apuntando ahí) |
| El VPS de Hostinger | Linux con root y systemd. El mismo donde está WorkAdventure |
| Docker con el plugin de compose | `docker compose version` tiene que responder. Si ya corrés WorkAdventure, está |
| ~1 GB de RAM libre | 450 MB son de Arena; el resto es aire para el build |

En tu máquina, para verificar: Node 22 y Go 1.26.

---

## Paso 1 · El VPS: Postgres y el backend

La base corre **en el mismo VPS que el backend**, en un contenedor al lado del
suyo (`arena/CLAUDE.md` §2). No hay una base en internet a la que conectarse ni
credenciales que ir a buscar a ningún panel.

1. **Traé el código y construí la imagen.**

   ```bash
   sudo git clone <repo> /opt/arena && cd /opt/arena
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
   | `POSTGRES_PASSWORD` | `openssl rand -hex 32` — **generala, y en hexa**: ver abajo |
   | `JWT_SECRET` | `openssl rand -base64 48` — **generalo, no lo inventes** |
   | `ADMIN_USERNAME` / `ADMIN_PASSWORD` | la cuenta de instructor que se crea al arrancar |
   | `ALLOWED_ORIGINS` | `https://arena.ejemplo.com` |
   | `API_PORT` | `18099` — el puerto de loopback, ver arriba |
   | `COOKIE_SAMESITE` | `lax` con la opción A del paso 3; `none` con la B |
   | `COOKIE_SECURE` | `true`. Siempre, fuera de localhost |
   | `ARENA_IMAGE` | `arena-api:latest` |

   `DATABASE_URL` **no hace falta tocarla**: el compose de producción la arma
   contra el servicio `db` y pisa lo que haya en el archivo. Dejala completa igual,
   para tener de dónde copiar una cadena cuando quieras un `psql`.

   > **La contraseña, en hexa.** `openssl rand -base64` emite `/` y `+`, y según
   > los bytes también `@` o `:`. Cualquiera de esos adentro de una URL de
   > conexión la parte donde no va, y el backend muere en el arranque con un error
   > que acusa a otro:
   >
   > ```
   > DATABASE_URL no se puede interpretar: cannot parse
   > `postgres://arena:xxxxxx@db:5432/arena?sslmode=disable`:
   > failed to parse as URL (invalid port ":jqrL5…" after host)
   > ```
   >
   > Habla del puerto, la contraseña ya está tapada con `xxxxxx`, y el problema es
   > la contraseña. El compose arma la conexión en forma `clave=valor` justamente
   > para no depender de eso, pero una contraseña con un espacio igual lo rompería.
   > `openssl rand -hex 32` no tiene ninguno de los dos problemas.

   El resto tiene un valor por defecto razonable, y cada variable está explicada
   en `.env.example` — incluidos los techos de memoria del paso anterior.

3. **Revisá el compose antes de levantarlo.** Con el archivo de entorno ya escrito,
   esto expande todas las variables y falla si falta alguna:

   ```bash
   cd /opt/arena/arena/deploy
   docker compose --env-file /etc/arena/arena.env -f docker-compose.prod.yml config
   ```

   Mirá dos líneas de esa salida: que el `ports` del backend empiece con
   `127.0.0.1:` y que el servicio `db` **no tenga `ports` en absoluto**.

4. **El servicio.**

   ```bash
   sudo cp arena/deploy/arena-api.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable --now arena-api
   journalctl -u arena-api -f
   ```

   La unit levanta los dos contenedores. En el primer arranque, la base crea el
   cluster —eso tarda— y recién después arranca el backend, que aplica el esquema
   solo: `MIGRATE=true` y el esquema es idempotente, así que no hay un paso de
   migración aparte que alguien pueda olvidarse de correr.

   En el log tenés que ver, en este orden: `database system is ready to accept
   connections`, después `instructor listo` y `escuchando`.

5. **Probalo desde el propio VPS**, que es el único lugar desde donde se puede:

   ```bash
   curl -si -X POST http://127.0.0.1:18099/api/auth/check-code \
     -H 'Content-Type: application/json' -d '{"code":"ZZZZ-9999"}'
   ```

   Tiene que contestar `404` con `{"error":{"code":"CODE_NOT_FOUND",…}}`. Eso
   prueba tres cosas de una: el proceso escucha, la base responde y el sobre de
   error es el del contrato. **No** intentes alcanzarlo desde afuera: no se puede,
   y está bien que no se pueda. El firewall del VPS puede quedar cerrado a todo lo
   entrante salvo SSH; el túnel del paso 2 no necesita que se habilite nada.

6. **Mirá la memoria una vez**, con todo arriba:

   ```bash
   docker stats --no-stream arena-db arena-api
   ```

   Si `arena-api` pasa de 100 MB en reposo, algo se está acumulando y conviene
   mirarlo antes de la clase, no durante.

---

## Paso 2 · Cloudflare Tunnel

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

   > Si WorkAdventure ya usa un túnel de Cloudflare, este es **otro** túnel, con
   > su propio UUID y su propio servicio. No compartas la config: el día que
   > toques una, tocás las dos.

3. **Instalá la configuración.**

   ```bash
   sudo mkdir -p /etc/cloudflared
   sudo cp ~/.cloudflared/<UUID>.json /etc/cloudflared/
   sudo cp /opt/arena/arena/deploy/cloudflared/config.yml /etc/cloudflared/
   sudo nano /etc/cloudflared/config.yml     # reemplazar los «REEMPLAZAR»
   sudo chown -R cloudflared:cloudflared /etc/cloudflared
   sudo chmod 600 /etc/cloudflared/<UUID>.json
   ```

   El `service:` del ingress ya viene apuntando a `http://127.0.0.1:18099`. Si
   elegiste otro `API_PORT`, es acá donde se cambia el otro extremo.

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

## Paso 3 · Cloudflare Pages: el frontend

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

## Paso 4 · Verificar

Desde la raíz del repo, en tu máquina:

```bash
node scripts/verify-arena.mjs
```

Y contra la base de verdad, que es la parte que no se puede simular. **En el VPS**
la base no publica ningún puerto, así que el psql se corre adentro del propio
contenedor de Postgres:

```bash
cd /opt/arena
export ARENA_PSQL_CONTAINER=arena-db
export DATABASE_URL='postgres://arena:<POSTGRES_PASSWORD>@localhost:5432/arena?sslmode=disable'
node scripts/verify-arena.mjs esquema
```

Esa `DATABASE_URL` es la que ve el contenedor `arena-db` desde adentro —por eso
`localhost`—, no la que ve el host.

Eso aplica el esquema **dos veces** —tiene que quedar limpio las dos— y corre la
reconciliación del ledger: `sum(coin_transactions.delta)` contra `users.balance`,
por usuario, **cero diferencias**. Si sale una fila, hay una nota que el ledger no
explica y eso se arregla antes de dar la clase. A mano, sin Node:

```bash
docker exec -i arena-db psql -U arena -d arena -v ON_ERROR_STOP=1 \
  < /opt/arena/arena/scripts/reconcile.sql
```

Nunca se corrige editando el ledger —el trigger lo impide— sino agregando un
movimiento `adjustment` que compense.

### Checklist del día

- [ ] `systemctl status arena-api cloudflared` — los dos `active (running)`
- [ ] `docker ps` muestra `arena-db` y `arena-api` en `Up`, y el de la base `(healthy)`
- [ ] `docker stats --no-stream arena-db arena-api` — ninguno pegado a su techo
- [ ] `free -m` — queda RAM libre con WorkAdventure también arriba
- [ ] Un `POST` a `https://arena.ejemplo.com/api/auth/check-code` contesta `CODE_NOT_FOUND`, desde una red que no sea la del VPS
- [ ] `node scripts/verify-arena.mjs` en verde
- [ ] La reconciliación del ledger da cero filas
- [ ] Entrás con la cuenta de instructor y podés generar un código
- [ ] Un código de prueba se canjea y acredita 1000 monedas

---

## Local, con Postgres descartable

Sin túnel y sin tocar el VPS. Este es el **otro** compose, el de desarrollo:

```bash
cd arena/deploy
cp .env.example .env      # completar JWT_SECRET
docker compose up --build
```

Levanta Postgres con el esquema ya aplicado y el backend en
`http://127.0.0.1:8099`. El frontend, aparte, con `npm start` en `arena/frontend`.

La diferencia con producción, en una línea: **el de desarrollo publica la base en
`127.0.0.1:5432`** para poder abrirle un psql, y el de producción no la publica en
ningún lado.

Para volver a aplicar el esquema después de cambiarlo —sin borrar los datos, que
es la prueba que importa:

```bash
docker compose exec -T db psql -U arena -d arena -v ON_ERROR_STOP=1 \
  < ../docs/contract/schema.sql
```

Y para empezar de cero: `docker compose down -v`.

---

## Mantenimiento

**Backup después de cada clase.** Las monedas son nota, y el volumen vive en el
mismo disco que todo lo demás:

```bash
docker exec arena-db pg_dump -U arena -d arena --no-owner --no-privileges -Fc \
  > "/root/backups/arena-$(date +%F).dump"
```

Para restaurar en una base vacía:

```bash
docker exec -i arena-db pg_restore -U arena -d arena --clean --if-exists \
  < /root/backups/arena-2026-08-04.dump
```

**Un psql cuando hace falta mirar algo:**

```bash
docker exec -it arena-db psql -U arena -d arena
```

**Actualizar el backend:**

```bash
cd /opt/arena && sudo git pull
docker build -f arena/deploy/Dockerfile -t arena-api:latest arena
sudo systemctl restart arena-api
```

El volumen de la base no se toca en ese ciclo. `cloudflared` tampoco hace falta
reiniciarlo: reconecta solo. Hay unos segundos de corte, así que no lo hagas con
una carrera `running`.

**Rotar `JWT_SECRET`** invalida todas las sesiones abiertas. Se puede, pero no en
medio de una clase.

**Lo que nunca se corre en el VPS:** `docker compose -f docker-compose.prod.yml
down -v`. Esa `-v` borra el volumen con las monedas de todos, sin preguntar. Para
parar, `sudo systemctl stop arena-api`.

---

## Cuando algo no anda

| Síntoma | Dónde mirar |
|---|---|
| `curl` al hostname público da 502 | `journalctl -u cloudflared -n 50`. Casi siempre el `service:` del ingress no coincide con `API_PORT` |
| Un 404 **sin** el sobre JSON de error | ese 404 lo puso el catch-all del túnel, no el backend: revisá el `hostname:` del ingress. El 404 del backend siempre trae `{"error":{"code":…}}` |
| El frontend carga pero `/api` da 404 | la ruta del Worker (opción A) o el `_routes` de Pages |
| `port is already allocated` al levantar | otra cosa del VPS —WorkAdventure— tiene ese puerto. `sudo ss -ltnp`, elegí otro `API_PORT` y cambiá también el `service:` del túnel |
| El backend no arranca | `journalctl -u arena-api -n 50`. Casi siempre `JWT_SECRET` o `POSTGRES_PASSWORD` sin completar |
| `DATABASE_URL no se puede interpretar: … invalid port` | no es el puerto: es la contraseña, que trae un carácter que rompe la cadena. Regenerala con `openssl rand -hex 32` — ver el paso 1 |
| `password authentication failed for user "arena"` | cambiaste `POSTGRES_PASSWORD` **después** del primer arranque. La imagen solo la usa al crear el cluster: hay que cambiarla también en la base, con `docker exec arena-db psql -U arena -d arena -c "alter user arena password '…'"` |
| El backend arranca y se muere solo, sin panic | lo mató el kernel por memoria: `docker inspect arena-api --format '{{.State.OOMKilled}}'`. Subí `API_MEM_LIMIT` **y** `GOMEMLIMIT` |
| La base responde lenta y el VPS swapea | `free -m`. Con WorkAdventure arriba puede no haber lugar para los dos: bajá `PG_MAX_CONNECTIONS` o sumá RAM |
| `connection refused` a la base desde el host | es lo esperado: en producción la base no publica puerto. Usá `docker exec arena-db psql …` |
| El esquema falla en la segunda pasada | un `create` sin `if not exists`. `node scripts/verify-arena.mjs esquema` te dice cuál |
| El alumno se desloguea cada 15 minutos | la cookie de refresh: mirá la opción A del paso 3 |
| El disco se llena | `docker system df`. Los logs de Arena están topeados; revisá las imágenes viejas con `docker image prune` |

---

## Apéndice · Si algún día la base vuelve a estar afuera (Supabase)

No es el despliegue de hoy: la base es Postgres en el VPS, decidido y confirmado
(`arena/CLAUDE.md` §2). Queda anotado porque el backend habla **Postgres plano**
—sin SDK y sin extensiones—, así que mudarla es cambiar `DATABASE_URL`, y estas
tres cosas son las que muerden al hacerlo:

1. **El plan gratuito de Supabase pausa el proyecto tras 7 días sin actividad.**
   Un curso semanal cae exactamente en esa ventana, y reanudarlo es manual y tarda
   varios minutos que no tenés cinco antes de empezar. Las dos salidas: un **ping**
   programado —un cron cada 12 horas contra un endpoint que *toque la base*, como
   `POST /api/auth/check-code`, porque un handler que responde sin consultar no
   cuenta como actividad— o el **plan pago**, que no pausa proyectos. Y como los
   backups del plan gratuito son limitados, el `pg_dump` propio después de cada
   clase deja de ser opcional.

2. **Hay que usar el pooler en modo *transaction*** (puerto 6543), no la conexión
   directa: en los proyectos nuevos `db.<ref>.supabase.co` resuelve solo a IPv6 y
   un VPS sin IPv6 de salida no llega. Con ese pooler, **`DB_SIMPLE_PROTOCOL=true`
   no es opcional**: las sentencias preparadas no sobreviven a la transacción y el
   driver falla con `prepared statement "lrupsc_1" already exists`, un error que no
   dice una palabra sobre su causa.

3. **`sslmode=require`**, porque ahí la base sí está en internet — a diferencia de
   hoy, donde la conexión no sale del bridge de Docker.
