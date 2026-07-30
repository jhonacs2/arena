# Arena — la economía y las reglas

> **Esto es el documento que manda.** El esquema, la API y el frontend salen de
> acá. Si algo de este archivo cambia, cambia el esquema — no se parchea del otro
> lado.

Arena es la app **en vivo** que usan los alumnos durante el módulo: se registran
con un código, apuestan monedas en carreras que abre el instructor, y esas monedas
valen nota. No es material didáctico y **no se les entrega el código**: es un
producto que consumen.

---

## 1. La economía

| Regla | Valor |
|---|---|
| Conversión | **100 monedas = 1 punto** |
| Saldo inicial al canjear el código | **1000 monedas = 10 puntos** |
| Piso de la nota | **no hay.** Apostar mal sí baja la nota |
| Piso del saldo | **0.** Nunca negativo |
| Monto de una apuesta | `1 ≤ monto ≤ saldo` |
| Apuestas por carrera y por alumno | **exactamente una** |
| Cuotas | ⚠️ **SIN DECIDIR** — ver abajo |
| Recarga automática | **no hay**. El instructor regala a pedido |

```
puntos = floor(monedas / 100) + puntos regalados
```

Es una función del saldo, no una columna: no hay dos números que puedan
desincronizarse. Lo verifica `schema.test.sql`.

### No hay piso, y eso es deliberado

Una versión anterior de este documento puso un piso de 10 puntos y lo llamó «la
decisión más importante del proyecto», con este argumento: sin piso, el juego es un
riesgo académico y lo racional pasa a ser no apostar nunca.

El argumento es bueno, pero **contradice la instrucción**: *«si funden las monedas
me van a deber nota y se tendrán que esforzar más»*. Con piso no deben nada y no
hay nada que compensar — la frase sólo tiene sentido sin piso.

El caso legítimo que el piso intentaba resolver —que un reconocimiento ganado no se
pueda perder en una apuesta— ya está cubierto por `point_grants`: **los puntos
regalados no pasan por el juego.** Eso protege lo que se ganó sin anular la
consecuencia de apostar mal.

### ⚠️ Cuotas fijas o pari-mutuel: SIN DECIDIR

**Nada que dependa del pago debe implementarse hasta que esto se cierre.**

| | Cómo paga | Efecto |
|---|---|---|
| **Cuotas fijas** | `monto × cuota`, la paga «la casa» | cada alumno juega contra la cuota. La masa de monedas crece |
| **Pari-mutuel** | se reparte el pool entre los que aciertan | **suma cero: para que uno gane, otro pierde** |

Con nota adentro, pari-mutuel es una curva: los alumnos se sacan calificación entre
ellos. Cuotas fijas los hace jugar contra el sistema y no entre sí, que es más
consistente con que las monedas midan participación.

El esquema hoy está escrito para **pari-mutuel** (`nominal_odds`,
`race_settlements`, sin `odds_at_bet`) porque así lo dejó una reescritura que
afirmó una decisión que no se había tomado. Queda así hasta la confirmación, pero
**no es una decisión tomada** y este aviso no se saca hasta que lo sea.

### Dos regalos distintos, que no son lo mismo

| El instructor regala… | Efecto |
|---|---|
| **monedas** | le da con qué seguir jugando. Pasa por el juego y se puede volver a perder |
| **puntos** | le sube la nota directo. No pasa por el juego y no se puede perder |

Viven en tablas separadas (`coin_transactions` y `point_grants`) porque un punto
regalado por explicar algo en el code review no debería poder perderse en una
apuesta.

### Qué significa «fundirse»

Que se queda sin con qué jugar, **no que baje de nota**. No hay recarga
automática: el que funde te pide y vos decidís. El panel del instructor muestra la
lista de saldos justamente para que puedas ver a quién le pasó sin que tenga que
levantar la mano.

> La contracara de no tener recarga automática: en cada clase vas a tener que
> mirar esa lista. Si molesta, la recarga por sesión es una tabla más y un botón
> — está aislado a propósito.

### Pari-mutuel: apuestan entre ellos, no contra la casa

Lo apostado en una carrera forma un **pool**, y el pool se reparte entre quienes
acertaron, en proporción a lo que pusieron.

```
Carrera con 800 monedas apostadas. Ganó Viento Norte, que tenía 300.
  Ana puso 100 → 100 × 800 / 300 = 266
  Bruno puso 200 → 200 × 800 / 300 = 533
```

Tres consecuencias, y las tres importan:

**Es suma cero.** El total de monedas del curso solo cambia cuando vos regalás.
Con cuotas fijas contra una casa, la masa de monedas deriva sin control y la
conversión a nota se desfasa sola.

**No hay cuota que congelar.** El pago no se puede saber al apostar: depende de
cómo apueste el resto. Por eso en `bets` no hay `odds_at_bet` — no existe un
número al momento de la apuesta que tenga sentido guardar. La cuota nominal del
caballo existe, pero solo alimenta al simulador y le indica al alumno quién es
favorito antes de que haya pool suficiente.

**El resto de la división no se descarta.** `800 / 300` no es exacto: los pagos
truncados suman 799 y sobra 1 moneda. Esa moneda se reparte de a una entre los
ganadores por orden de monto, hasta agotarla. Si se descartara, el pari-mutuel
dejaría de ser suma cero y el curso perdería monedas que nadie perdió apostando.

> Al escribir el test de esto apareció un bug que vale documentar: **`sum(bigint)`
> en Postgres devuelve `numeric`**, así que la división salía decimal y al
> guardarla en una columna `bigint` se **redondeaba** en vez de truncar. La suma
> seguía dando 800 —la conservación se cumplía— pero la moneda del resto le
> quedaba al apostador equivocado. Verificar solo la conservación no lo detecta:
> hay que asertar los pagos uno por uno.

### Una apuesta por carrera, y por qué

Con pari-mutuel, cubrir todos los caballos **no** es un arbitraje: te devuelve el
pool menos la parte de los demás, que es pérdida esperada. Así que la regla no
está para tapar un agujero, está por dos motivos propios:

1. **Que la decisión importe.** Elegir un caballo es el juego; comprar todos es no
   jugar.
2. **Que el pool no lo domine quien más boletos compre**, sino quien acierte.

---

## 2. El registro por código de invitación

No hay registro abierto ni verificación por correo. El instructor genera códigos y
los reparte.

**Formato:** `AAAA-9999` — cuatro letras, guion, cuatro dígitos. Ejemplo:
`AVBD-1234`.

Se generan con un alfabeto **sin caracteres ambiguos**: sin `I`, `L`, `O`, `U`
en las letras y sin `0` ni `1` en los dígitos. El código se dicta en voz alta o se
copia de un chat, y `AVBD-1O34` es una llamada de soporte garantizada.

**El flujo, en una sola pantalla:**

1. El alumno escribe el código.
2. Si es válido y no fue canjeado, se habilita el resto del formulario: nombre,
   apellido, usuario y contraseña.
3. Al enviar: se crea el usuario, se marca el código como canjeado, se acreditan
   las 1000 monedas en el ledger, y queda con sesión iniciada.

Los tres pasos son **una transacción**. Un código a medio canjear —usuario creado
sin monedas, o código quemado sin usuario— es el peor estado posible.

**Un código, un uso.** El canje es atómico: si dos personas envían el mismo código
en el mismo instante, una gana y la otra recibe `CODE_ALREADY_REDEEMED`.

---

## 3. Las carreras y las salas

El instructor **es** el operador de la carrera. Nada arranca solo.

```
draft ──▶ open ──▶ running ──▶ finished
  │         │
  └─────────┴──▶ cancelled
```

| Estado | Qué se puede hacer |
|---|---|
| `draft` | el instructor la arma: nombre, caballos, cuotas. No la ven los alumnos |
| `open` | los alumnos la ven, se unen a la sala y **apuestan** |
| `running` | la carrera corre. **No se aceptan apuestas** |
| `finished` | resultados publicados, apuestas liquidadas |
| `cancelled` | se devuelve cada apuesta al saldo, íntegra |

**«Sala»** es el conjunto de alumnos conectados a una carrera. Se ve quién está,
qué apostó cada uno una vez que la carrera arranca, y el desarrollo en vivo.

Reglas duras:

- Las apuestas se cierran **en el servidor**, al pasar a `running`. El botón
  deshabilitado en el frontend es una cortesía, no un control.
- **La liquidación se calcula una sola vez**, al pasar a `finished`, y queda
  escrita en `race_settlements` con el pool y el pool ganador. No se recalcula
  nunca: recalcular con datos que cambiaron es cómo se paga dos veces.
- **Si nadie acertó, se devuelve cada apuesta íntegra.** No hay a quién pagarle, y
  quedarse el pool sería inventar una casa que este modelo no tiene.
- La simulación es **autoritativa del servidor**. El cliente dibuja lo que recibe.
- El ledger es **append-only**. No se borra ni se edita una transacción; se
  compensa con otra.

---

## 4. Quién puede qué

| | Alumno | Instructor |
|---|---|---|
| Canjear código | sí | — |
| Ver carreras `open`/`running`/`finished` | sí | sí |
| Ver carreras `draft` | **no** | sí |
| Apostar | sí, una por carrera | no |
| Ver su saldo y su historial | sí | — |
| Ver el saldo de todos | **no** | sí |
| Crear códigos, carreras, caballos | no | sí |
| Abrir, largar, cancelar una carrera | no | sí |
| Regalar monedas | no | sí |
| Ver el ledger completo | no | sí |

El rol vive en `users.role` y se valida **en el servidor en cada endpoint**. Un
alumno que edite su token no obtiene nada.

---

## 5. Lo que Arena NO es

- **No es material de clase.** No hay starter, no hay corrección, no se publica el
  código. Los alumnos usan la app, no la leen.
- **No usa Angular 18.** Es Angular 22. La regla cero del repo aplica a
  `project/` y `lab/`, no acá — ver `arena/CLAUDE.md`.
- **No comparte backend con el hipódromo.** Son dos productos; lo único que
  comparten es la **línea visual** (`docs/design/tokens.json`).
- **No maneja dinero real.** Monedas simbólicas que se traducen a nota.
