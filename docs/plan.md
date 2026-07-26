# Plan de construcción

Orden de trabajo del módulo. Una fase por vez; cada una termina con `node scripts/verify.mjs` en verde y un commit.

| | Fase | Estado |
|---|---|---|
| **0** | Contrato y esqueleto | ✅ **terminada** |
| **1** | Backend Go, completo y congelado | ✅ **terminada** |
| **2** | Baseline visual del hipódromo | ✅ **terminada** |
| **3–13** | Una sesión por fase, S1 … S11 | ⬜ siguiente |
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
- `scripts/`: `verify.mjs`, `check-contrast.mjs`, `gen-tokens-css.mjs`, `gen-race-ticks.mjs`, `gen-silks-specimen.mjs`, `sync-contract.mjs`
- `CLAUDE.md` v2

`node scripts/verify.mjs` corre hoy **20 verificaciones** sobre contrato, diseño, backend y —cuando exista— el código Angular.

## Fase 1 — Backend Go, completo y congelado ✅

Un binario, **una sola dependencia externa** (`github.com/coder/websocket`), sin base de datos, con el dataset embebido.

- Los 13 endpoints de `openapi.yaml`, con el sobre de error uniforme y los 18 códigos del catálogo
- Hub WebSocket multiplexado por sala. `race.finished` se arma **por destinatario**: difundir el mismo objeto filtraría lo que cobró cada uno
- Simulador a 10 Hz que implementa `race-simulation.md`. El test golden reproduce el fixture tick por tick, así que la app se ve igual contra el mock y contra el servidor
- Calendario que larga carreras solo: la primera a los 30 s, después el ciclo del dataset cada 2 h
- JWT HS256 escrito a mano, PBKDF2 de stdlib, refresh de un solo uso, verificación por correo con un emisor de desarrollo que imprime el enlace en consola
- Estado en memoria con copia atómica en disco; `RESET=1` vuelve al dataset limpio
- Dockerfile distroless con binario estático

Probado de punta a punta contra el servidor real: **26 verificaciones**, incluidos 22 eventos de cuenta regresiva, 419 ticks y la liquidación con el saldo consistente entre socket y REST.

**Desde acá el backend está congelado:** `CLAUDE.md` §12.

## Fase 2 — Baseline visual del hipódromo ✅

El piso sobre el que se para S1. Angular 18.2.14 con versiones exactas, standalone en todo, `strict` más `noUncheckedIndexedAccess`.

- `styles.css` con `@layer reset, tokens, base, utilities` y el bloque de tokens inyectado desde `tokens.json`
- Las tres tipografías variables auto-hospedadas en woff2 — 356 KB en el repo, cero CDN
- `silkFromId()` portado a TypeScript, `<app-silk>`, y un test que cruza el port contra la implementación de referencia en JS para los 54 caballos
- `shared/ui/`: button, skeleton, empty-state, logo. **Sin badge ni race-card**: son el ejercicio de S2
- `core/models/` (contraparte del contrato), `core/mocks/` (generado desde el seed), `core/theme/`
- `layout/shell` con salto al contenido, navegación y el interruptor de tema
- `features/sistema`: la muestra del sistema de diseño, que es la prueba visible de que todo esto funciona
- `starter/` creado con el mismo baseline — a partir de S1 los dos divergen

Build de producción en **75 kB** de transferencia inicial. Probado en navegador, en claro y oscuro, y medido de 320 a 768 px sin desbordamiento horizontal.

Un defecto real que encontró el cruce JS↔TS: había **dos mezcladores de hash distintos** —el generador de sedas usaba una mezcla de dos pasos y el simulador de carreras una de tres—. Unificados; de paso, las coincidencias de seda entre carreras bajaron de 1 a 0.

## Fases 3–13 — Una sesión por fase

Cada una entrega: slice de la solución · slice del starter con sus `// TODO(Sn)` · ruta `/sNN` del lab en ambas versiones · los 9 archivos de `sesiones/sNN-*/` · commit `feat(sNN): …` + tag en el repo del alumno.

**Recomendación:** construir S1, **darla**, ajustar la plantilla del guión con lo que salga, y recién ahí seguir con S2. El guión de 12 bloques es una hipótesis hasta que se corre con alumnos reales.

## Fase 14 — Publicación

`publish-student-repo.sh` · tags por sesión · deploy del frontend · README del alumno con setup en 5 minutos.
