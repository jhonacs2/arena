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
| [`docs/contract/decisiones.md`](docs/contract/decisiones.md) | las reglas. **Fuente de verdad** |
| [`docs/contract/schema.sql`](docs/contract/schema.sql) | Postgres. Las reglas que la base puede hacer cumplir, están ahí |
| [`docs/contract/api.md`](docs/contract/api.md) | endpoints, errores y eventos de socket |

---

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

**Nada de Supabase.** Se evaluó y se descartó por dos razones: su plan gratuito
pausa el proyecto tras 7 días sin actividad y una clase semanal cae justo en esa
ventana, y con el bucle de carrera a 10 Hz una base en otra red paga un viaje de
red por consulta. El backend habla **Postgres plano** —sin SDK, sin extensiones—,
así que mover la base es cambiar `DATABASE_URL`.

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
  mismo se desincronizan siempre. Y la vista lleva el **piso de 10**: apostar mal
  no baja la nota.
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
| Base | Postgres en el mismo VPS |

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
