# Frontend — Hipódromo (solución de referencia)

Angular **18.2.14**, standalone en todo, cero `NgModule`, CSS nativo con tokens.

> Esta es la **solución de referencia**: la ve solo el instructor. Lo que recibe el alumno está en `../starter`, con los cuerpos vacíos y sus `// TODO(Sn)`.

---

## Levantarlo

```bash
npm install --save-exact
npm start          # http://localhost:4200
```

Hoy la app tiene una sola ruta: **`/sistema`**, la muestra del sistema de diseño. No es una pantalla del producto — es la prueba de que la base funciona y la referencia contra la que mirar cuando algo se ve raro. En S1 se suma `/carreras` y pasa a ser la ruta por defecto.

```bash
npm run typecheck   # tsc --noEmit
npm run build       # producción
npm test            # Karma en el navegador
```

Y desde la raíz del repo, `node scripts/verify.mjs` corre todo junto: contrato, diseño, backend y frontend.

---

## Cómo está armado

```
src/app/
├── core/
│   ├── models/      tipos del contrato: Race, Horse, Bet, User, Page, ApiError
│   ├── mocks/       GENERADO desde docs/contract/seed/ — no editar a mano
│   └── theme/       claro / oscuro / sistema
├── shared/ui/
│   ├── silk/        ⭐ las sedas de jockey
│   ├── button/      el botón del sistema
│   ├── skeleton/    estado de carga
│   ├── empty-state/ estado vacío y de error
│   └── logo/        la marca
├── layout/          el armazón: cabecera, navegación, router-outlet
└── features/
    └── sistema/     la muestra del sistema de diseño
```

**Regla de dependencias** (`CLAUDE.md` §6): `features/` puede importar de `core/` y `shared/`. `shared/` no importa de nadie. `core/` no importa de `features/`.

### Lo que NO está acá, a propósito

Cada pieza se escribe **cuando se escribe su clase**, no antes: si estuviera hecha de antemano, esa sesión se quedaría sin práctica.

Por eso `<app-badge>` (S2) y `<app-race-card>` (S2) ya están, y los pipes de S4, los stores de S5 y la pantalla en vivo de S10 todavía no. En `starter/` no está ninguna: ahí aparecen recién al cerrar cada clase.

---

## Decisiones que conviene conocer

**Las sedas de jockey son el elemento firma.** Cada caballo tiene su casaca derivada de su `id` con una función pura — `shared/ui/silk/silk.util.ts`. 54 caballos con identidad visual y **cero archivos de imagen**. El detalle está en `docs/design/tokens.md` §1.

La misma función existe dos veces: en TypeScript acá y en JavaScript en `scripts/gen-silks-specimen.mjs`, que dibuja la hoja de muestra. `silk.util.spec.ts` compara las dos contra los 54 caballos: si el port deriva, la seda de un caballo dejaría de coincidir con la hoja aprobada. Ese cruce ya encontró un bug real — dos mezcladores de hash distintos.

**Los mocks salen del seed, no de la imaginación.** `core/mocks/` lo genera `scripts/gen-mocks.mjs` desde `docs/contract/seed/`, el mismo dataset que carga el backend Go. Un componente escrito contra estos datos funciona sin cambios cuando en S7 se conecta al servidor real. Las fechas se reubican al cargar el módulo, con la misma regla de rebase que aplica el backend.

**Las tipografías están auto-hospedadas.** Tres familias variables en woff2, 356 KB, dentro del repo. Nada de CDN: un aula tiene wifi de aula, y una clase de dos horas no se puede caer porque Google Fonts tarde. Se descargan con `node scripts/fetch-fonts.mjs`, y `verify.mjs` falla si alguien enlaza al CDN.

**Los tokens no se escriben a mano.** La paleta vive en `docs/design/tokens.json` y `scripts/gen-tokens-css.mjs` la inyecta en `src/styles.css` entre los marcadores `/* @tokens:start */` y `/* @tokens:end */`. Editar los colores en el CSS hace fallar `verify.mjs`.

**`styles.css` tiene solo capas, tokens, reset y base.** Todo lo demás vive en el `.css` del componente. Es pedagógico: el temario incluye *View Encapsulation*, y si el estilo estuviera todo en global ese tema se quedaría sin nada que enseñar. Por eso el presupuesto de CSS por componente está en 8 KB y no en los 2 KB que trae el CLI.

**TypeScript va más allá de `strict`.** `noUncheckedIndexedAccess`, `noUnusedLocals` y `noUnusedParameters` están activos. El primero es el que más cuesta y el que más enseña: `horses[0]` es `Horse | undefined` y hay que hacerse cargo.

---

## Modo oscuro

Está **diseñado, no invertido**. En oscuro el borde neobrutalista se invierte de tinta a tiza, y la sombra dura también: la card pasa de "recorte negro sobre papel" a "recorte de tiza sobre pizarra".

El interruptor de la cabecera cicla claro → oscuro → sistema. Por defecto sigue a `prefers-color-scheme`.
