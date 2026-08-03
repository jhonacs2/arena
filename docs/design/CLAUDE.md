# CLAUDE.md — Sistema visual

> Complementa el [`CLAUDE.md` de la raíz](../../CLAUDE.md). El porqué de cada decisión está en [`tokens.md`](tokens.md).

**Sin Tailwind. Sin Sass. Sin librería de componentes.** CSS nativo con tokens. La razón es pedagógica: el temario incluye *View Encapsulation*, y con utilidades atómicas ese tema se queda sin nada que enseñar.

---

## Reglas duras

- **Los tokens no se escriben a mano.** Se editan en `tokens.json` y se corre `node scripts/gen-tokens-css.mjs`, que inyecta el bloque entre `/* @tokens:start */` y `/* @tokens:end */` de cada `styles.css`. Editar el CSS a mano hace fallar a `verify.mjs`.
- **Color en `oklch()`, nunca hex.**
- **Contraste AA (4,5:1) en todo texto**, verificado por `scripts/check-contrast.mjs`. El neobrutalismo falla exactamente acá.
- `styles.css` global **solo** para capas, tokens, reset y base. Todo lo demás vive en el `.css` del componente, encapsulado, y **fuera de capas** a propósito: así gana sin `!important`.
- Bordes sólidos de 3 px, sombras duras `4px 4px 0` sin blur, **cero gradientes**, **radio 0 sin excepciones**.
- Nada de neumorfismo ni glassmorphism.
- Movimiento con propósito: la animación comunica estado —la carrera avanzando, el saldo cambiando—, nunca decora. Respetar `prefers-reduced-motion`.

---

## Modo oscuro: se diseña, no se invierte

Los dos temas se diseñan por separado y **los dos se miran antes de dar algo por terminado** (punto 5 de la definición de terminado).

Reglas del oscuro:

- Se reasignan **tokens semánticos** (`--surface`, `--text`, `--border`), **jamás primitivos**.
- **Las tres superficies tienen que distinguirse a simple vista.** `--surface`, `--surface-raised` y `--surface-sunken` con la misma luminancia es el error más común y el que hace que una pantalla oscura se vea plana y sucia.
- **Nada de negro puro ni de blanco puro.** Texto casi blanco sobre fondo casi negro produce halación: el texto vibra. El texto oscuro se apoya en una tiza levemente bajada, no en `#fff`.
- **La que se invierte a tiza es la SOMBRA, no el borde.** En oscuro la profundidad se dibuja con luz: una sombra de tinta topea en 1,55:1 aunque la pintes de negro puro, porque la curva sRGB ya está en el piso. El gesto no cambia —`4px 4px 0`, sin blur, radio 0—, cambia de qué color es el bloque. El detalle y los números están en `$shadowComment` de [`tokens.json`](tokens.json).
- **Y se paga apagando el borde.** Si la sombra brilla y el borde también, la pantalla es una reja iluminada. Jerarquía en oscuro, de más fuerte a más débil: **texto → sombra → texto secundario → borde**. El borde llega justo por encima del piso de 3:1 y ahí se queda.
- **Los bloques de color saltan lo mismo que en claro.** Un acento no se aclara porque el fondo se oscureció: eso duplica el salto y lo vuelve una linterna. El acento es el **mismo** `rail-500` en los dos temas — en claro queda más oscuro que la página, en oscuro más claro.

---

## Las sedas de jockey son el elemento firma

Cada caballo tiene su casaca, **generada como SVG determinístico desde su `id`** — patrón de cuerpo × 2 colores × mangas, con rechazo si los dos colores no separan ΔL ≥ 0,22.

Implementación de referencia: `scripts/gen-silks-specimen.mjs`. Port a TypeScript: `shared/ui/silk/silk.util.ts`, con un test que compara los dos contra los 54 caballos.

> **Ningún texto se pinta sobre una seda.** El número del caballo va en su cuadrado aparte, tinta sobre tiza. Es lo que permite que los diez colores de seda sean los saturados de verdad sin romper AA.

---

## Layout

Mobile-first real: la carrera en vivo funciona a 360 px · **container queries** en `<app-race-card>` para que responda a su contenedor y no al viewport · grid bento en el dashboard · `:has()` para estado condicional sin clases extra en el template.

---

## Imágenes

Las sedas y los avatares son SVG generado: **no hay archivos de imagen para el 80 % de la UI**. Lo que sí hace falta está en [`IMAGES.md`](IMAGES.md) — seis piezas con dimensiones, formato y prompt. Las genera el instructor; mientras tanto la app usa los placeholders de `assets/` y **nunca hay una imagen rota**.

Si necesitás una imagen que no está en `IMAGES.md`, **agregala ahí primero** con su especificación. No metas un `<img>` apuntando a un archivo que no existe.
