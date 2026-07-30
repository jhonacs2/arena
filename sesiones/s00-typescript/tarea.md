# S0 · Tarea asíncrona

**Entrega antes de S1.** Se lee en voz alta en clase antes de cortar: una tarea
que solo se manda por chat no se hace.

---

## Qué hacer

Terminar las dos misiones si te quedaron a medias, y después tipar **la apuesta**,
que es la parte del dominio que hoy no tocamos.

### 1 · `bet.model.ts`, leído y respondido

Abrí `project/frontend/starter/src/app/core/models/bet.model.ts`. Ya está
escrito y bien tipado — es material de lectura. Contestá estas tres, en un
archivo aparte o en un comentario al final:

1. `BetStatus` es una unión de tres literales. ¿Qué pasaría si fuera `string`?
   Escribí una línea de código que hoy no compila y que con `string` sí compilaría.
2. `payout` es `number` y el comentario dice «`0` mientras esté pendiente».
   ¿Podría el tipo decir eso en vez del comentario? ¿Cómo, y qué se ganaría?
3. `MIN_BET_AMOUNT` y `MAX_BET_AMOUNT` son constantes, no tipos. ¿Por qué los
   límites de un monto no se pueden expresar como una unión de literales?

### 2 · El borrador de una apuesta

En el mismo archivo, agregá un tipo que represente **una apuesta que todavía no se
mandó al servidor**: tiene `raceId`, `horseId` y `amount`, y nada más.

- No lo escribas a mano campo por campo. **Derivalo de `Bet`** con un utility type.
- Escribí arriba, en un comentario de una línea, por qué derivarlo es mejor que
  copiarlo.

### 3 · Una función con el tipo que dice la verdad

Agregá esta función al final del archivo:

```ts
/** La apuesta de mayor monto de la lista. */
export function biggest(bets: readonly Bet[]): /* ← el tipo va acá */ {
  // …
}
```

Con la lista vacía **no puede devolver una apuesta**. El tipo tiene que decirlo, y
adentro no puede haber ni `!` ni `as`.

## Dónde

`project/frontend/starter/src/app/core/models/bet.model.ts`

## Listo cuando

- [ ] Las tres preguntas están contestadas
- [ ] El borrador está derivado de `Bet`, no copiado
- [ ] `biggest` compila y devuelve un tipo que contempla la lista vacía
- [ ] `npx tsc --noEmit` no imprime nada
- [ ] No hay ni un `any`, ni un `!`, ni un `as`
- [ ] Está commiteado con el prefijo de la sesión: `feat(s00): tipos de la apuesta`

## Cuánto lleva

**30–45 minutos.** Si te lleva más de una hora, parás y anotás dónde te trabaste
— eso es material para el bloque de las 0:05 de la próxima.

## Pistas

Para el punto 2 hay tres utility types que sirven según qué quieras: `Pick`,
`Omit` y `Partial`. Están en `conceptos.md` §10 con un ejemplo de cada uno.

Para el punto 3, el patrón es el mismo de `cheapest` y de `favourite`: acumulador
que arranca en `undefined` y una pregunta antes de comparar. Está resuelto dos
veces en `correccion.md`.

## Material de apoyo

- **TypeScript Handbook · Everyday Types**: <https://www.typescriptlang.org/docs/handbook/2/everyday-types.html>
- **TypeScript Handbook · Narrowing**: <https://www.typescriptlang.org/docs/handbook/2/narrowing.html>
- **Playground**, para probar sin instalar nada: <https://www.typescriptlang.org/play>
- El contrato del proyecto: `docs/contract/openapi.yaml`

---

## Para la clase que viene

**Vení con el proyecto levantado.** Los quince minutos de instalar cosas te los
quiero ahorrar:

```bash
cd lab/starter && npm start
```

Si eso abre <http://localhost:4200> y se ve el pizarrón del café, estás listo.

---

## Para el instructor

No se corrige una por una. Se revisa una al azar en el code review de S1, y lo que
aparezca repetido en varias entregas se convierte en pregunta del `wayground.csv`.

**Lo que más va a aparecer:**

- **`biggest` devolviendo `Bet` y un `!` adentro.** Es exactamente el error que la
  clase entera intentó desarmar, y verlo repetido significa que el punto no quedó.
  Vale cinco minutos al empezar S1.
- **El borrador escrito a mano** «porque total son tres campos». Es la respuesta
  honesta de alguien que todavía no vio romperse una copia. Sirve mostrarles el
  diff de un `Bet` con un campo nuevo.
- **La pregunta 3 contestada con «porque son muchos números».** La respuesta que
  buscamos es que el rango es un dato de validación, no una forma; los tipos
  describen formas y las validaciones se ejecutan. Es la puerta de entrada a S8.
