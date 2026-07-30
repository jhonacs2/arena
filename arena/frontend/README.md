# Arena · frontend

Angular **22.1.0**, standalone, zoneless, sin SSR. Las versiones están fijadas
exactas: `package.json` no tiene ni un `^`.

Las reglas están en [`arena/CLAUDE.md`](../CLAUDE.md) y la economía en
[`arena/docs/contract/decisiones.md`](../docs/contract/decisiones.md). La línea
visual sale de `docs/design/tokens.json` de la raíz — el bloque de tokens de
`src/styles.css` lo genera `node scripts/gen-tokens-css.mjs` y **no se edita a
mano**.

```bash
npm start          # http://localhost:4200
npm run build      # producción
npm run typecheck  # tsc --noEmit
```

## El backend de mentira

`src/environments/environment.ts` tiene `useMockBackend: true`. Con eso, un
interceptor responde `/api/**` desde `core/api/mock/mock-world.ts` con **las
mismas formas y los mismos códigos de error** de `docs/contract/api.md`, e incluye
la simulación de la carrera a 10 Hz. Apuntar al Go real es cambiar ese booleano a
`false`: no hay una sola línea de componente que dependa del mock.

Cuentas de la semilla (contraseña `arena1234` en las dos):

| Usuario | Rol |
|---|---|
| `profe` | instructor |
| `anag` | alumna, con saldo e historial |

Códigos sembrados: `AVBD-1234` y `TXNQ-4562` sin usar, `KMPR-8827` ya canjeado —
sirven para ver los dos errores distintos de la pantalla de registro.

> El mundo del mock vive en memoria. Al recargar, la sesión de `profe` y `anag`
> sobrevive (sus ids son fijos); las cuentas creadas durante la corrida, no.

## Mapa

```
src/app/
├── core/
│   ├── api/        api-error · auth.interceptor · http-context · mock/
│   ├── auth/       session.store · auth.service · auth.guard
│   ├── data/       race.service · admin.service
│   ├── race/       race-channel (WebSocket real y de mentira)
│   ├── theme/      claro · oscuro · sistema
│   └── models/     los tipos del contrato
├── shared/
│   ├── format/     coins · odds · when   (los montos son ENTEROS)
│   └── ui/         button · badge · callout · empty-state · field · silk
└── features/       register · login · dashboard · race · admin
```

**Los montos son enteros.** Monedas en unidades, cuotas ×100 (`340` es 3,40).
`shared/format/` formatea con aritmética de enteros y concatenación: el `number`
con coma no existe en ningún momento.
