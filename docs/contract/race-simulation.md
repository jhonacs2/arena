# Simulación de la carrera — especificación

Este algoritmo se implementa **dos veces**: en Go (`project/backend/internal/sim/`) y en JavaScript (`scripts/lib/race-sim.mjs`, que alimenta el fixture y el mock del frontend).

Que las dos produzcan la misma carrera no es un capricho de simetría: es lo que permite **verificar** el punto 5 de la definición de terminado — *"se probó contra el mock y contra el backend real, y se ve igual en los dos"* — en lugar de prometerlo. El test golden de Go reproduce `fixtures/race-ticks.jsonl` tick por tick.

Si cambiás una constante, cambiala acá primero y regenerá el fixture.

---

## Entradas y salidas

```
simulate(raceId, runIndex, horses[]) → { duration, runners[], ticks[], podium[3] }
```

`runIndex` es el número de corrida de esa carrera: `0` la primera vez que el servidor la larga, `1` la segunda. Mismas entradas → misma carrera, siempre. **Sin `rand`, sin reloj.**

---

## Constantes

| Nombre | Valor | Qué controla |
|---|---|---|
| `TICK_HZ` | `10` | frecuencia de `race.tick` |
| `COUNTDOWN_SECONDS` | `60` | cuenta regresiva del servidor. El fixture graba solo los últimos 10: nadie mira un minuto de cuenta atrás en una demo |
| `BASE_DURATION` | `42.0` s | lo que tarda el favorito nominal |
| `ODDS_SPREAD` | `4.5` s | cuánto más tarda, de base, el de cuota más alta |
| `JITTER` | `5.0` s | ventana aleatoria. **Es lo que permite el batacazo** |
| `SHAPES` | `[0.82, 1.00, 1.22]` | exponente por estilo: `front`, `even`, `closer` |

Con estos números el favorito gana **alrededor de la mitad** de las corridas — 51 % medido sobre las cuatro carreras del programa. Lo suficiente para que apostarle sea razonable frente al 14 % del azar puro, y no tanto como para que sea la única respuesta. Con los valores iniciales (`5.5` y `3.0`) ganaba el 73 %, y apostar dejaba de ser una decisión.

---

## Aleatoriedad determinística

FNV-1a de 32 bits sobre la clave completa, y una mezcla final para que ids parecidos no den valores parecidos.

```
fnv1a(s):  h = 0x811C9DC5
           por cada byte c de s:  h ^= c;  h = (h * 0x01000193) mod 2³²
           devolver h

mix32(h):  h ^= h >> 15
           h  = (h * 0x2545F491) mod 2³²
           h ^= h >> 13
           devolver h

rnd(raceId, runIndex, horseId, salt) = mix32(fnv1a("raceId/runIndex/horseId/salt")) / 2³²
```

Tres sales: `"t"` tiempo de llegada, `"s"` estilo, `"p"` fase de la ondulación.

> La clave se arma con `/` como separador y el `runIndex` en decimal sin ceros a la izquierda. Un byte de diferencia da otra carrera, así que las dos implementaciones tienen que armar la cadena igual.

---

## Preparación

```
# 1. Habilidad esperada, desde la cuota. Menor cuota = favorito.
#    El empate se rompe por número de partida, para no depender del orden del array.
byOdds  = horses ordenados por (odds asc, number asc)
skill_i = posición_en_byOdds(i) / (n - 1)          # 0 = favorito, 1 = el de cuota más alta

# 2. Tiempo de llegada
finishTime_i = BASE_DURATION + ODDS_SPREAD × skill_i + JITTER × (rnd(…,"t") − 0.5)

# 3. Estilo de carrera
styleIndex_i = piso( rnd(…,"s") × 3 )              # 0 front, 1 even, 2 closer
shape_i      = SHAPES[styleIndex_i]

# 4. Fase de la ondulación
phase_i = rnd(…,"p") × 2π

# 5. La carrera termina cuando cruza el primero, no cuando llegan todos
duration = techo1( mín(finishTime_i) )   # hacia ARRIBA a la décima, ver abajo
```

Un `front` sale fuerte y se apaga; un `closer` sale atrás y remata. El estilo es **independiente de la cuota**: por eso a veces el favorito viene de atrás y a veces se apaga.

---

## Progreso

Para cada tick `i = 1 … duration × TICK_HZ`, con `t = redondear1(i / TICK_HZ)`:

```
u_i        = mín(t / finishTime_i, 1)
wobble(p,t)= 0.0045 × sen(p + 0.9·t) + 0.0022 × sen(2.7·p + 2.3·t)
crudo_i    = u_i ^ shape_i + wobble(phase_i, t) × (1 − u_i)
progress_i = acotar(crudo_i, anterior_i, 1)        # nunca retrocede, nunca pasa de 1
```

La ondulación se multiplica por `(1 − u)`: se apaga cerca del disco. Nadie zigzaguea en la llegada.

**Puestos:** ordenar por `progress` descendente; empate por `horseId` ascendente. El primero es `place: 1`.

**Redondeo:** `progress` a 3 decimales, `t` a 1 decimal, *después* de calcular los puestos.

**Orden del array:** `positions` sale en el orden de `horses`, no por puesto. Ordenar por `place` es tarea del cliente — está dicho en `ws-events.md` y es un detalle que se evalúa en S10.

---

## Podio y liquidación

`podium` son los tres primeros del **último tick**. La carrera termina cuando el ganador cruza; los demás quedan con `progress < 1`, que es lo que pasa en una carrera de verdad.

Liquidación, en este orden:

1. Cada apuesta `pending` de esa carrera: gana si su `horseId` es el del puesto 1.
2. `payout = redondear(amount × odds)` con la cuota **congelada al apostar**. Si perdió, `0`.
3. Se acredita el pago al saldo.
4. Se emite `balance.updated` a cada usuario afectado, `race.finished` a la sala (con `payouts` filtrados por usuario) y `leaderboard.updated` a todos.

---

## Sobre la igualdad entre las dos implementaciones

`sen` y `pow` son IEEE-754 en los dos lenguajes, pero **no está garantizado que coincidan al último bit**. Con el redondeo a 3 decimales, una diferencia de 1 ULP solo cambiaría el resultado si el valor cayera a menos de 1e-16 de un límite de redondeo.

El test golden de Go compara contra el fixture con tolerancia `1e-9` en `progress` y **exactitud en los puestos**. Si alguna vez divergen, salta ahí.

---

## El fixture

`fixtures/race-ticks.jsonl` es la grabación de **`race_005`, corrida 164**, desde la sesión de `usr_001` (Ana Robles).

Esa corrida no se eligió al azar: Payador, el favorito, corre de `closer`, va **sexto a mitad de carrera** y gana por **0,003** de progreso — un photo finish. Un fixture donde el favorito lidera de punta a punta no prueba que la animación funcione.

Se regenera con `node scripts/gen-race-ticks.mjs`.
