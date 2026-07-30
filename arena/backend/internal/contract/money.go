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
// Este paquete NO tiene una función de pago ni una de puntos, y eso sigue siendo
// deliberado aunque las dos reglas ya estén decididas: las dos dependen de datos
// que este paquete no ve.
//
// **El pago de una apuesta ganadora** es pari-mutuel (decisiones.md §0):
//
//	pago = amount * pool / winningPool
//
// No cabe acá porque necesita TODAS las apuestas de la carrera, no una. Vive en
// `races.PariMutuel`, detrás de `races.SettlementRule`, que es la única pieza del
// backend que conoce la economía. Una `Payout(amount, odds)` en este paquete daría
// un número plausible con la economía equivocada — la de cuotas fijas — y eso es
// exactamente el error que costó cinco reescrituras del contrato.
//
// **Los puntos.**
//
//	puntos = max(10, floor(monedas / 100)) + puntos regalados
//
// El piso de 10 está decidido (decisiones.md §0) y vive en la vista `user_scores`,
// no acá. El segundo término vive en
// `point_grants`, que este paquete no conoce: los puntos los resuelve la capa que
// administra las notas, y al cliente le llegan por `GET /api/me`. Una
// `Points(balance)` acá daría un número parecido y distinto, y dos números que
// representan lo mismo se desincronizan siempre.
