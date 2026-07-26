# Sistema de diseño — Hipódromo

`docs/design/CLAUDE.md` fija el estilo: neobrutalismo disciplinado, `oklch`, bordes de 3 px, sombras duras `4px 4px 0`, cero gradientes, contraste AA sin excepciones. Eso no se discute.

Lo que este documento resuelve es lo que esas reglas dejan abierto: **de qué mundo sale la paleta, qué tipografías y cuál es el elemento que hace que la app se reconozca**.

Los valores viven en `tokens.json`, que es la fuente canónica. Este archivo explica el porqué.

---

## 1. El elemento firma: las sedas del jockey

Las *racing silks* son la casaca del jockey. No son un adorno del deporte: son un **sistema combinatorio registrado** con dos siglos de uso. Un patrón de cuerpo, dos colores, un patrón de mangas, un patrón de gorra. Color plano, borde duro, cero gradiente. Ya eran neobrutalistas antes de que existiera la palabra.

**Cada caballo recibe su seda, generada como SVG de forma determinística a partir de su `id`.** Una función pura: mismo `id`, misma seda, siempre — en la lista, en el detalle, en la carrera en vivo y en el historial.

Esa única decisión resuelve cuatro cosas a la vez:

| Problema | Cómo lo resuelve |
|---|---|
| Avatares para 54 caballos | Cero archivos de imagen. Es SVG generado. |
| Elegir colores de acento | No se eligen: la card, el badge y el carril toman los de *ese* caballo. |
| Identificar 8 caballos a 360 px | Seda + número es legible. Ocho nombres apilados, no. |
| Que el ejercicio de clase sea real | `silkFromId()` es una función pura y `<app-silk>` es el ejercicio de `input.required()` de S2 y el de pipe de S4. |

Es un sistema **de reglas**, que es exactamente la razón por la que `CLAUDE.md` eligió neobrutalismo: un dev sin ojo de diseñador lo ejecuta bien si sigue las reglas.

### La gramática

```
Silk = patrón de cuerpo × color primario × color secundario × mangas × gorra
```

| Eje | Valores |
|---|---|
| Cuerpo | `solid` `halves` `quarters` `stripes` `hoops` `chevron` `sash` `star` `diamond` `seams` |
| Mangas | `plain` `alt` `hooped` `striped` |
| Gorra | `primary` `secondary` `quartered` `striped` |
| Colores | los 10 registrados: negro, blanco, rojo, azul, amarillo, verde, naranja, violeta, rosa, celeste |

Derivación: se hashea el `id` a un entero y se toman los índices por módulo. Dos reglas de rechazo, que se aplican avanzando el hash hasta que se cumplan:

1. **Primario ≠ secundario.**
2. **ΔL ≥ 0,22 entre los dos.** Sin esto salen sedas azul-sobre-violeta que a 24 px se ven como un cuadrado sólido.

10 × 10 × 9 × 4 × 4 = **14.400 combinaciones**. Con 54 caballos, ninguna repetición.

### La regla que hace que las sedas no rompan AA

> **Ningún texto se pinta sobre una seda.**

El número del caballo va en su cuadrado aparte: `ink-900` sobre `chalk-050`, siempre, 17,49:1. La seda queda al lado, nunca debajo.

Por eso los 10 colores de seda no necesitan cumplir contraste entre sí — y por eso pueden ser los colores saturados de verdad del deporte en lugar de pasteles lavados.

---

## 2. Paleta

### La apuesta

El neobrutalismo casi siempre grita con el fondo: blanco puro y un amarillo o rosa fluorescente. Acá el **cromo va desaturado a propósito** — verde-negro pizarra, tiza tibia, ocre de arena — y las **sedas son lo único saturado en pantalla**.

Va contra el instinto del estilo. Es lo que evita que veinte cards con acento fuerte se conviertan en ruido, y es dónde se gasta toda la audacia de la página.

### Primitivas

| Token | `oklch` | Para qué |
|---|---|---|
| `--ink-950` | `0.13 0.018 155` | lo más hundido en oscuro |
| `--ink-900` | `0.17 0.02 155` | verde-negro pizarra · texto, borde y sombra en claro |
| `--ink-850` | `0.215 0.020 155` | **superficie en oscuro** |
| `--ink-800` | `0.27 0.021 155` | superficie elevada en oscuro |
| `--ink-700` | `0.36 0.02 155` | skeletons en oscuro |
| `--ink-500` | `0.50 0.02 155` | texto secundario en claro |
| `--chalk-000` | `1.00 0 110` | superficie elevada en claro |
| `--chalk-050` | `0.97 0.008 110` | tiza tibia · superficie en claro |
| `--chalk-100` | `0.94 0.010 110` | superficie hundida en claro |
| `--chalk-150` | `0.91 0.010 110` | **texto en oscuro** |
| `--chalk-300` | `0.80 0.013 110` | **borde en oscuro** |
| `--chalk-400` | `0.72 0.014 110` | texto secundario en oscuro |
| `--turf-500` | `0.52 0.13 152` | verde pista · `finished`, ganada |
| `--rail-500` | `0.64 0.150 55` | arena · **acento primario**, `upcoming` |
| `--flag-500` | `0.56 0.220 27` | bandera · `live`, error |

La tiza es tibia (`h 110`, `C 0.008`) pero **no es crema**: a `L 0.97` con esa croma es papel, no beige. La distinción importa — el crema con acento terracota es el fondo por defecto de medio internet.

### Modo oscuro: diseñado, no invertido

No es un filtro. En oscuro el borde neobrutalista se invierte de tinta a tiza: la card pasa de «recorte negro sobre papel» a «recorte de tiza sobre pizarra». Solo se reasignan tokens **semánticos**; las primitivas nunca cambian.

La primera versión de este modo oscuro estaba mal, y vale la pena dejar escrito por qué — son los tres errores que se repiten en cualquier tema oscuro:

| Estaba | Está | Por qué |
|---|---|---|
| Superficie en `L 0.17`, casi negro | `L 0.215` | Texto casi blanco sobre negro casi puro produce **halación**: el texto vibra y cansa. |
| `--surface`, `--surface-raised` y `--surface-sunken` **casi iguales** | ΔL de 0,055 y 0,085 | Sin diferencia de luminancia, la pantalla se ve plana y sucia, y el borde queda como único indicio de dónde termina una card. |
| Texto y borde los dos en `L 0.97` | texto `0.91`, borde `0.80` | Un borde tan brillante como la letra **compite con la letra**. El borde delimita; no tiene que gritar. |
| Sombra dura en tiza | sombra en `ink-950` | Veinte sombras blancas sobre pizarra son veinte manchas. Una sombra más oscura que la superficie da profundidad sin ruido. |

El texto bajó de 17,49:1 a **13,36:1** y el borde de 17,49:1 a **9,35:1**. Los dos siguen muy por encima de AA: acá el problema nunca fue el contraste, era el exceso.

`scripts/check-contrast.mjs` verifica ahora **las dos cosas**: el contraste y la separación entre superficies. La segunda no es una regla WCAG, es una regla de este sistema, y está mecanizada para que no vuelva a pasar.

### Contraste — medido, no afirmado

`node scripts/check-contrast.mjs` corre dentro de `verify.mjs` y falla el build si algún par baja del umbral. Estado actual:

| | claro | oscuro |
|---|---|---|
| texto de cuerpo | 17,49:1 | 13,36:1 |
| texto secundario | 5,45:1 | 7,04:1 |
| borde de 3 px | 17,49:1 | 9,35:1 |
| botón primario | 5,40:1 | 9,16:1 |
| badge EN VIVO | 4,78:1 | 7,14:1 |
| badge de ganada | 4,76:1 | 14,39:1 |

El checker ya se ganó el sueldo dos veces: `flag-500` estaba en `L 0.58` y daba **4,42:1** con tiza encima — se veía bien y no llegaba a AA. Y cuatro tokens caían **fuera del gamut sRGB**, donde el navegador los habría recortado en silencio y el contraste real habría dejado de ser el calculado.

---

## 3. Tipografía

Tres roles, tres familias, todas OFL y **self-hosted en woff2**. Nada de CDN de fuentes: el aula tiene wifi de aula, y una clase de dos horas no se cae porque Google Fonts tarde.

| Rol | Familia | Por qué esta |
|---|---|---|
| **Display** | Bricolage Grotesque *(variable: `wght` `wdth` `opsz`)* | Textura de imprenta, ligeramente incómoda. Un eje de ancho variable deja condensar los títulos largos en móvil sin cambiar de familia. No es Inter ni Space Grotesk. |
| **Cuerpo** | Public Sans | Neutra a propósito. Su trabajo es no competir con las sedas. |
| **Números** | Martian Mono *(variable)* | Cifras tabulares reales. Cuotas, saldo, cuenta regresiva y posiciones son un **tablero totalizador**, y un tablero no baila de ancho cuando cambia un dígito. |

Martian Mono no es una decisión estética suelta: en la carrera en vivo el marcador se actualiza 10 veces por segundo. Con cifras proporcionales, cada tick corre el layout.

### Escala

Modular, razón 1,2. `--text-base: 1rem`. Los tamaños de display usan `clamp()` para ser fluidos entre 360 px y 1440 px sin media queries.

```
2xs .694  ·  xs .833  ·  sm .9  ·  base 1  ·  lg 1.2
xl 1.44   ·  2xl 1.728  ·  3xl 2.074  ·  4xl 2.488  ·  5xl 2.986   (rem)
```

Pesos: cuerpo 400/600. Display 700/800 — el neobrutalismo pide grosor, y en Bricolage el peso alto es donde la familia tiene carácter.

---

## 4. Forma

| Token | Valor | Nota |
|---|---|---|
| `--border-width` | `3px` | fijo, §8 |
| `--shadow-hard` | `4px 4px 0 var(--shadow)` | fijo, §8 · sin blur, sin alpha |
| `--shadow-hard-sm` | `2px 2px 0` | elementos chicos: badges, el cuadrado del número |
| `--shadow-hard-lg` | `6px 6px 0` | modales y la card destacada del bento |
| `--radius` | `0` | **cero, sin excepciones** |

Radio cero en todo. No porque el borde redondeado esté mal, sino porque un sistema con dos radios necesita un criterio para elegir entre ellos, y ese criterio es gusto. Cero es una regla; "6 px en cards y 0 en badges" es una opinión que se degrada.

Espaciado: base 4 px, escala `1 2 3 4 6 8 12 16 24` → `0.25rem … 6rem`.

**Estado presionado:** el botón se desplaza `2px 2px` y la sombra baja a `2px 2px 0`. El objeto se hunde de verdad. Es el único efecto de profundidad del sistema.

---

## 5. Movimiento

Tres momentos. Los tres comunican estado. **Nada más se mueve.**

| Momento | Qué hace | Por qué así |
|---|---|---|
| `race.tick` | El carril traslada su marcador. Solo `transform`. | 10 Hz sobre 8 carriles. Cualquier propiedad que dispare layout tira los cuadros al piso. |
| `balance.updated` | El número **salta** como un split-flap. | Vernáculo del tablero totalizador. Un fade diría "esto cambió despacio"; el saldo cambia de golpe. |
| Cuenta regresiva | Dígitos mono que reemplazan. Sin rebote, sin escala. | Un reloj que rebota es un reloj en el que no confiás. |

```css
--ease-snap: cubic-bezier(0.2, 0, 0, 1);
--dur-instant: 90ms;   /* hover, foco */
--dur-quick:  160ms;   /* badges, entrada de card */
--dur-flip:   260ms;   /* el salto del saldo */
```

Todo dentro de `@media (prefers-reduced-motion: reduce)` cae a cambio instantáneo. La carrera **sigue mostrando posiciones** — se saca la transición, no la información.

---

## 6. Layout

- **Mobile-first de verdad.** La carrera en vivo funciona a 360 px: ocho carriles horizontales, seda + número a la izquierda, barra de progreso a la derecha.
- **Container queries en `<app-race-card>`.** La card responde a *su contenedor*, no al viewport. Es lo que la hace servir igual en la grilla de la lista y en la columna angosta del detalle, sin variantes ni un `@Input() compact`.
- **Bento en el dashboard.** Carreras destacadas, saldo y leaderboard en módulos asimétricos. La jerarquía la da el tamaño del módulo, no un título que diga "importante".
- **`:has()` para estado condicional.** `.bet-form:has(:invalid)` apaga el botón sin una clase extra en el template ni un `[class.disabled]` en el TypeScript.

### En toda vista con datos, sin excepción (§8)

Cargando con skeleton · vacío con una acción · error con reintento. Labels asociados a sus inputs. Foco visible siempre. `aria-live="polite"` en el marcador de la carrera.

---

## 7. Lo que saqué

> *Antes de salir, mirate al espejo y sacate un accesorio.* — Chanel

Estaba planificada una **textura de pista** de fondo en la vista en vivo. La saqué. Los ocho carriles con sus marcas de furlong ya dicen "esto es una pista"; la textura solo agregaba una imagen que generar, un pedido de red y un riesgo de contraste bajo el marcador.

Un accesorio menos, y una imagen menos en `IMAGES.md`.
