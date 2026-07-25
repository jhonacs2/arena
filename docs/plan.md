# Plan de construcción

Orden de trabajo del módulo. Una fase por vez; cada una termina con `node scripts/verify.mjs` en verde y un commit.

| | Fase | Estado |
|---|---|---|
| **0** | Contrato y esqueleto | ✅ **terminada** |
| **1** | Backend Go, completo y congelado | ⬜ siguiente |
| **2** | Baseline visual del hipódromo | ⬜ |
| **3–13** | Una sesión por fase, S1 … S11 | ⬜ |
| **14** | Publicación y deploy | ⬜ |

---

## Por qué este orden

El riesgo que mata este tipo de proyecto no es el orden de construcción, es la **deriva de contrato**: el frontend asume `startsAt`, el backend devuelve `start_time`, y S7 se convierte en una clase de depuración en lugar de una de `HttpClient`.

La Fase 0 congela el contrato como artefacto ejecutable **antes** de escribir cualquiera de los dos lados, y ambos asertan contra los mismos archivos JSON. Con eso, el orden de lo que viene después deja de ser crítico — pero conviene igual hacer el backend primero, porque una vez terminado no se toca más y libera toda la atención para el frontend, que es donde vive la pedagogía.

---

## Fase 0 — Contrato y esqueleto ✅

- `git init`, árbol, `.gitignore`, `.editorconfig`, `.nvmrc`
- `docs/contract/` completo: `openapi.yaml` (13 endpoints), `ws-events.md`, `error-codes.md` (18 códigos), seed (12 usuarios · 8 carreras · 54 caballos · 34 apuestas), 12 samples, fixture de 462 eventos
- `docs/design/`: `tokens.json` canónico, `tokens.md`, `IMAGES.md`, hoja de muestra de las 54 sedas
- `docs/curriculum.md` — el mapa de las 11 sesiones
- `sesiones/_plantilla/` — los 9 archivos de cada sesión
- `theme/marp-neobrutal.css`
- `scripts/`: `verify.mjs` (12 verificaciones), `check-contrast.mjs`, `gen-tokens-css.mjs`, `gen-race-ticks.mjs`, `gen-silks-specimen.mjs`
- `CLAUDE.md` v2

## Fase 1 — Backend Go, completo y congelado

Los 13 endpoints de `openapi.yaml` · hub WebSocket con multiplexado por sala · simulador de carrera a 10 Hz (autoridad del servidor) · JWT + refresh de un solo uso · verificación por correo con Resend, con sender que solo loguea en dev · carga del seed con la regla de rebase de fechas · **tests golden contra `docs/contract/samples/`** · Dockerfile y deploy.

Al terminar, el backend se congela: `CLAUDE.md` §12.

## Fase 2 — Baseline visual del hipódromo

Antes de S1 hace falta piso: `styles.css` con `@layer` y los tokens inyectados · las 3 tipografías subseteadas en woff2 · `silkFromId()` y `<app-silk>` portados desde `gen-silks-specimen.mjs` · primitivas de `shared/ui/` · `layout/shell` · `core/models/` · fixtures.

Es lo único que se construye adelantado. Sin esto, S1 no tiene dónde pararse.

## Fases 3–13 — Una sesión por fase

Cada una entrega: slice de la solución · slice del starter con sus `// TODO(Sn)` · ruta `/sNN` del lab en ambas versiones · los 9 archivos de `sesiones/sNN-*/` · commit `feat(sNN): …` + tag en el repo del alumno.

**Recomendación:** construir S1, **darla**, ajustar la plantilla del guión con lo que salga, y recién ahí seguir con S2. El guión de 12 bloques es una hipótesis hasta que se corre con alumnos reales.

## Fase 14 — Publicación

`publish-student-repo.sh` · tags por sesión · deploy del frontend · README del alumno con setup en 5 minutos.
