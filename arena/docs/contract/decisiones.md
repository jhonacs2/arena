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
| Piso del saldo | **0.** Nunca negativo |
| Monto de una apuesta | `1 ≤ monto ≤ saldo` |
| Apuestas por carrera y por alumno | **exactamente una** |
| Regalos del instructor | sí, en cualquier momento |

**Puntos = `floor(saldo / 100)`.** Es una función del saldo, no una columna: no hay
dos números que puedan desincronizarse.

### Qué significa «fundirse»

El alumno arranca con 10 puntos ya acreditados. Si apuesta mal y llega a 0
monedas, **no queda en negativo: queda en 0 puntos**, y esos 10 puntos que tenía
los tiene que compensar con el resto de la cursada. Eso es lo que se decidió con
«me van a deber nota y se tendrán que esforzar más».

> **Asunción explícita, marcada porque es la única lectura que tomé sin
> confirmar:** el saldo tiene piso en 0 y una apuesta nunca puede superar el
> saldo. La deuda es *pedagógica* —puntos que faltan— no un saldo negativo en la
> base. Si querés deuda literal, es un cambio de una línea: sacar el `CHECK
> (balance >= 0)` y el tope del monto. Está aislado a propósito.

### Una apuesta por carrera, y por qué no es un detalle

Con cuotas fijas, si un alumno pudiera apostar a **todos** los caballos de una
carrera se garantizaría ganancia siempre que la suma de las probabilidades
implícitas quede por debajo de 1. Sería una máquina de imprimir nota.

**Una sola apuesta por carrera lo elimina de raíz**, y además hace que la decisión
importe, que es lo interesante.

### La masa de monedas crece, y está bien

Las cuotas son fijas y paga «la casa», así que el total de monedas en circulación
sube con el tiempo. Es deliberado: **participar da monedas, y las monedas son
nota**. El panel del instructor muestra monedas y puntos, y el mapeo final a la
calificación lo decide el instructor — no lo decide la app.

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
- La cuota se **congela en la apuesta** (`odds_at_bet`). Nunca se recalcula desde
  la cuota actual del caballo.
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
