# CLAUDE.md — Arena

> **Acá NO vale la regla cero del repo.** El `CLAUDE.md` de la raíz dice «este
> proyecto es Angular 18» y eso aplica a `project/` y `lab/`, que es material
> didáctico congelado en esa versión. **Arena es Angular 22**, y las APIs que la
> raíz prohíbe —`resource()`, `httpResource()`, `linkedSignal()`, Signal Forms—
> acá están permitidas si conviene usarlas.

Arena es la app **en vivo** que usan los alumnos durante el módulo: se registran
con un código de invitación, apuestan monedas en carreras que abre el instructor,
y esas monedas valen nota.

**Leé primero [`docs/contract/decisiones.md`](docs/contract/decisiones.md).** Es el
documento que manda: la economía, los estados de una carrera y quién puede qué.

| | |
|---|---|
| [`README.md`](README.md) | **cómo levantarlo en local**, frontend y backend |
| [`docs/contract/decisiones.md`](docs/contract/decisiones.md) | las reglas. **Fuente de verdad** |
| [`docs/contract/schema.sql`](docs/contract/schema.sql) | Postgres. Las reglas que la base puede hacer cumplir, están ahí |
| [`docs/contract/api.md`](docs/contract/api.md) | endpoints, errores y eventos de socket |

---

## 0. El contrato no se reescribe: se reporta el desacuerdo

> **Si estás implementando y el contrato te parece equivocado, PARÁ Y DECILO. No lo edites.**

`docs/contract/` es fuente de verdad compartida entre varias personas y varios agentes trabajando en paralelo. Reescribirla desde una pieza tiene dos efectos, y el segundo es el caro:

1. Las otras piezas siguen construyendo contra lo que leyeron, y se descubre tarde.
2. **Queda por escrito, con autoridad, una decisión que nadie tomó.** Un `git log` convincente que afirma «la decisión es X» es peor que un desacuerdo abierto, porque el próximo que lo lea lo va a creer.

Ya pasó una vez: un agente reescribió la economía entera afirmando en el commit que eran «las decisiones tomadas». De sus tres cambios, dos contradecían instrucciones explícitas del usuario y el tercero era una pregunta que nunca se había respondido. Se detectó **solo porque otro agente tenía un test que toca Postgres de verdad** y el esquema le voló; con un doble en memoria, habría entregado un backend que compila, testea verde y no arranca.

**La regla, entonces:**

- Un desacuerdo con el contrato se reporta. No se resuelve editando.
- Si el contrato y el código se separaron, **manda el contrato** — y `decisiones.md` manda sobre `schema.sql` y `api.md`, que lo declara en su primera línea.
- Un cambio al contrato lo aplica **una sola mano, en una sola pasada**, y sólo con la decisión confirmada.
- **Nunca escribas «la decisión es X» sobre algo que decidiste vos.** Si lo elegiste porque el contrato no lo cubría, decí eso.

Y el corolario que vale para todo lo demás: **el test que vale es el que toca la cosa real.** El bug del reparto —`sum(bigint)` devuelve `numeric` en Postgres, así que la división redondeaba en vez de truncar y la moneda del resto le quedaba al apostador equivocado— no lo encuentra ningún doble en memoria, y verificar sólo que el total cierre tampoco: hay que asertar los pagos uno por uno.

## 1. No es material de clase

Los alumnos **usan** Arena, no leen su código. No hay starter, no hay corrección,
no se publica. Eso cambia dos cosas respecto del resto del repo:

- **No hay que enseñar con este código**, así que se puede usar lo más moderno que
  haya sin preocuparse de si ya se explicó en clase.
- **Sí hay que protegerlo.** Las monedas son nota, y donde hay nota hay incentivo
  para hacer trampa. Ver §4.

## 2. Stack

```
Angular   22.x        (última estable — la raíz está en 18, acá no)
Go        1.26.x
Postgres  16+, en el mismo VPS que el backend
```

**Comparte la línea visual con el hipódromo:** los tokens salen de
`docs/design/tokens.json` de la raíz, con `gen-tokens-css.mjs`. Neobrutalismo,
bordes de 3px, sombras duras `4px 4px 0`, cero gradientes, `oklch()`. Y **contraste
AA verificado mecánicamente**, igual que allá.

**La base es Postgres en el mismo VPS que el backend. No Supabase.** Decidido por el
usuario en un mensaje literal: *«Arena: piso de 10 en la nota, pari-mutuel, Postgres
en el VPS.»* Las dos razones que lo motivaron son reales y quedan anotadas porque
explican por qué no se vuelve: el plan gratuito de Supabase pausa el proyecto tras 7
días sin actividad —y una clase semanal cae justo en esa ventana—, y con el bucle de
carrera a 10 Hz una base en otra red paga un viaje de red por consulta.

El backend habla **Postgres plano** —sin SDK, sin extensiones— así que mover la base
sigue siendo cambiar `DATABASE_URL` en cualquier dirección. Lo que quedó escrito en
`arena/deploy/README.md` sobre el pooler en modo transaction y el ping semanal aplica
**solo** si algún día se vuelve a Supabase; con la base local no hace falta ninguno
de los dos.

La autenticación es propia: JWT HS256 y PBKDF2 de stdlib, el mismo patrón que el
backend del hipódromo. El registro es por código de invitación, que no encaja con
ningún proveedor de auth de estantería.

## 3. El idioma, igual que en todo el repo

> **El texto que ve el usuario, en español. El código, en inglés.**

Incluye los `message` del sobre de error: se muestran tal cual al usuario, así que
van en castellano. Los `code` son para el frontend y van en inglés.

## 4. Las monedas son nota: lo que eso obliga

Esta es la sección que distingue Arena de un juego.

- **Toda validación pasa en el servidor.** Un botón deshabilitado es cortesía. El
  rol, el estado de la carrera, el monto y el «una apuesta por carrera» se
  verifican en el handler, en una transacción.
- **La liquidación se calcula una vez** y queda escrita en `race_settlements`.
  Recalcular un pago con datos que cambiaron es cómo se paga dos veces.
- **El ledger es append-only**, y lo impone un trigger, no una convención. Un
  error se compensa con otro movimiento; no se edita la historia de la nota de
  alguien.
- **Los puntos son una vista**, no una columna. Dos números que representan lo
  mismo se desincronizan siempre. La vista lleva el **piso de 10**: perder monedas
  saca capacidad de seguir jugando, nunca calificación. Y los puntos regalados se
  suman aparte, en `point_grants`, porque no pasan por el juego.
- **La simulación es autoritativa del servidor** y la semilla queda guardada: si
  alguien reclama un resultado, se vuelve a correr igual.
- **`bet.placed` no revela el caballo** mientras la carrera está `open`.

## 5. Los montos son enteros

Monedas en unidades. **Cuotas nominales ×100 en entero:** `340` es 3.40.

Con `float`, `2.10 × 700` da `1469.9999999999998` y el saldo de alguien queda mal
por un centavo que nadie puede explicar — y acá ese centavo es nota. En Go es
`int64`, en Postgres `bigint`, en TypeScript `number` pero **siempre entero**.

El pago es pari-mutuel, no cuota fija:

```
payout = amount * pool / winningPool      división entera, trunca
```

y el resto de esa división **se reparte**, no se descarta — ver
[`docs/contract/decisiones.md`](docs/contract/decisiones.md) §1.

> **La trampa que ya nos mordió una vez:** en Postgres `sum(bigint)` devuelve
> `numeric`. Sin castear, la división sale decimal y al guardarla en `bigint` se
> **redondea** en vez de truncar; la suma total sigue cerrando, pero la moneda del
> resto le queda al apostador equivocado. `schema.test.sql` asierta los pagos uno
> por uno justamente porque verificar solo el total no lo detecta.

## 6. Despliegue

| | Dónde |
|---|---|
| Frontend | Cloudflare Pages |
| Backend | VPS de Hostinger |
| Base | Postgres en el mismo VPS — ver §2 |

El backend **no expone puerto**: `cloudflared` corre como servicio en el VPS y abre
la conexión hacia afuera (Cloudflare Tunnel). El frontend llama a `/api` en su
propio dominio y Cloudflare enruta al túnel.

> Eso oculta el origen **y** lo protege: no hay puerto de entrada que escanear. Lo
> que **no** hace es autenticar — la seguridad real son los JWT y la validación por
> rol de §4. Ocultar una URL no es un control de acceso.

## 7. Verificación

`node scripts/verify-arena.mjs`. Igual que el `verify.mjs` de la raíz: si falla, se
arregla antes de seguir.

Verifica, además de lo obvio: que el esquema aplique limpio dos veces seguidas, que
la reconciliación del ledger contra `users.balance` dé cero diferencias, que ningún
handler de `/admin/` esté sin chequeo de rol, que no haya `float` en montos, y el
contraste AA de la paleta.

Los tests de Go piden Postgres. Sin él se saltean, así que hay que dárselo:

```bash
docker run -d --rm --name arena-pg-test -e POSTGRES_PASSWORD=x -e POSTGRES_DB=arena_test \
  -p 55440:5432 postgres:16-alpine
export ARENA_TEST_DATABASE_URL="postgres://postgres:x@localhost:55440/arena_test?sslmode=disable"
export DATABASE_URL="$ARENA_TEST_DATABASE_URL"
node scripts/verify-arena.mjs
```

**El «test» en el nombre de la base no es decorativo:** los tests truncan todas
las tablas y `testdb.Pool` se niega a correr contra una base que no lo tenga.
Apuntar esa variable a la base de desarrollo ya borró los datos de alguien, con
los tests en verde y sin un aviso.

### Y el cableado, que ningún test de Go cubre

```bash
node arena/scripts/e2e.mjs      # con el backend corriendo
```

Los 12 paquetes en verde **no** prueban que `main.go` haya enchufado la regla de
liquidación, el hub y el runner. Un backend con `Rule` nula compila, pasa todos los
tests y no liquida una sola carrera: es el agujero que este script tapa. Recorre
código → registro → apuesta → carrera → nota contra el binario real, y asierta el
piso de 10 fundiendo a un alumno de verdad.

**Después de tocar `main.go`, corrélo.** Es el único lugar donde un cable
desconectado se nota.
