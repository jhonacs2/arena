package contract

// Los montos son enteros. Monedas en unidades, cuotas nominales ×100 (340 =
// 3.40).
//
// Con float, 2.10 × 700 da 1469.9999999999998 y el saldo de alguien queda mal por
// un centavo que nadie puede explicar — y acá ese centavo es nota. Ver
// arena/CLAUDE.md §5.

// CoinsPerPoint es la conversión de decisiones.md §1: 100 monedas = 1 punto.
const CoinsPerPoint = 100

// MinNominalOdds es la cuota nominal mínima, ×100. Coincide con el CHECK
// nominal_odds_minima del esquema.
//
// «Nominal» porque su papel es el mismo en las dos economías que están sobre la
// mesa: alimentar al simulador —es su parámetro de fuerza— y decirle al alumno
// quién es favorito. Lo que está sin decidir es si además determina el pago.
const MinNominalOdds = 101

// ── Lo que NO se calcula acá, y por qué ───────────────────────────────────
//
// Este paquete NO tiene una función de pago ni una de puntos. No es un olvido:
// las dos reglas están sin confirmar, y una implementación provisoria de algo que
// se traduce a nota es peor que no tenerla, porque compila, pasa los tests y nadie
// vuelve a mirarla.
//
// **El pago de una apuesta ganadora.** Hay dos modelos en discusión y dan números
// distintos para la misma carrera:
//
//	cuotas fijas   pago = amount * oddsAtBet / 100          (paga «la casa»)
//	pari-mutuel    pago = amount * pool / winningPool       (se reparte el pool)
//
// El segundo es suma cero y necesita repartir el resto de la división; el primero
// necesita congelar la cuota al apostar, y por lo tanto una columna que el otro no
// tiene. No es un detalle de implementación: cambia el esquema y cambia lo que
// significa la nota. Mientras no esté decidido, el pago entra por
// races.SettlementRule, que este repositorio declara y deliberadamente NO
// implementa.
//
// **Los puntos.** La fórmula tiene un piso en discusión:
//
//	sin piso   puntos = floor(monedas / 100)
//	con piso   puntos = max(10, floor(monedas / 100)) + puntos regalados
//
// La diferencia es si un alumno que funde queda debiendo nota o no, que es una
// decisión pedagógica y no técnica. Además el segundo término vive en
// `point_grants`, que este paquete no conoce: los puntos los resuelve la capa que
// administra las notas, y al cliente le llegan por `GET /api/me`. Una
// `Points(balance)` acá daría un número parecido y distinto, y dos números que
// representan lo mismo se desincronizan siempre.
