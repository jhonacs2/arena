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

Tres familias, y cada una es un material: **`ink-`** es la tinta con la que se escribe, **`chalk-`** es la tiza, **`slate-`** es la pizarra sobre la que se escribe en oscuro.

| Token | `oklch` | Para qué |
|---|---|---|
| `--ink-900` | `0.17 0.02 155` | verde-negro pizarra · texto, borde y sombra en claro; tinta sobre acentos en oscuro |
| `--ink-500` | `0.50 0.02 155` | texto secundario en claro |
| `--slate-950` | `0.13 0.008 250` | lo más hundido en oscuro |
| `--slate-850` | `0.235 0.010 250` | **superficie en oscuro** |
| `--slate-800` | `0.30 0.011 250` | superficie elevada en oscuro |
| `--slate-700` | `0.36 0.011 250` | skeletons en oscuro |
| `--chalk-000` | `1.00 0 110` | superficie elevada en claro |
| `--chalk-050` | `0.97 0.008 110` | tiza tibia · superficie en claro |
| `--chalk-100` | `0.94 0.010 110` | superficie hundida en claro |
| `--chalk-150` | `0.91 0.010 110` | **texto en oscuro** |
| `--chalk-300` | `0.80 0.013 110` | **sombra dura en oscuro** |
| `--chalk-400` | `0.72 0.014 110` | texto secundario en oscuro |
| `--chalk-600` | `0.61 0.014 110` | **borde en oscuro** |
| `--turf-500` | `0.52 0.13 152` | verde pista · `finished`, ganada |
| `--turf-300` | `0.62 0.125 152` | ganada en oscuro |
| `--rail-500` | `0.64 0.150 55` | arena · **acento primario en los dos temas**, `upcoming` |
| `--flag-500` | `0.56 0.220 27` | bandera · `live`, error |
| `--flag-300` | `0.65 0.160 27` | `live`, error en oscuro |

La tiza es tibia (`h 110`, `C 0.008`) pero **no es crema**: a `L 0.97` con esa croma es papel, no beige. La distinción importa — el crema con acento terracota es el fondo por defecto de medio internet.

**La pizarra es fría y casi acromática** (`h 250`, `C 0.010`). Antes las superficies del oscuro eran `ink-850/800/950` —matiz 155, croma 0,02, el mismo verde que `turf`—, y ese es el segundo error clásico del modo oscuro después de la halación: un matiz que en un trazo de texto es la pizarra de la marca, extendido a pantalla completa no lee como pizarra, lee como **pantano**. Y el fondo competía de matiz con el verde semántico de «ganada». A `L 0.235` la pizarra nueva es `#1b1f23`: seis puntos entre el rojo y el azul, un carbón neutro con un susurro frío, no un color.

Eso deja el tema con una lógica de una sola línea: **todo lo que es información es tibio —la tiza, el ámbar, el coral, el verde— y el suelo es frío.** Es de noche en el hipódromo, y lo que está iluminado es lo que hay que leer.

### Modo oscuro: diseñado, no invertido

No es un filtro. En oscuro la card pasa de «recorte negro sobre papel» a «recorte de tiza sobre pizarra». Solo se reasignan tokens **semánticos**.

Este modo oscuro se rehizo dos veces, y las dos vale la pena dejarlas escritas.

**La primera pasada** arregló tres errores que se repiten en cualquier tema oscuro:

| Estaba | Está | Por qué |
|---|---|---|
| Superficie casi negra | `L 0.235` | Texto casi blanco sobre negro casi puro produce **halación**: el texto vibra y cansa. |
| `--surface`, `--surface-raised` y `--surface-sunken` **casi iguales** | ΔL de 0,065 y 0,105 | Sin diferencia de luminancia la pantalla se ve plana y sucia, y el borde queda como único indicio de dónde termina una card. |
| Texto y borde los dos en `L 0.97` | texto `0.91`, borde `0.61` | Un borde tan brillante como la letra **compite con la letra**. El borde delimita; no tiene que gritar. |

**La segunda pasada** revirtió una decisión de la primera. Queda acá con su razón original, porque el que la tomó no estaba siendo tonto — le faltaba el número:

> | Decía | Ahora | |
> |---|---|---|
> | Sombra dura en `ink-950`, más oscura que la superficie | Sombra dura en `chalk-300`, **tiza** | «Veinte sombras blancas sobre pizarra son veinte manchas. Una sombra más oscura que la superficie da profundidad sin ruido.» |
>
> La segunda mitad de esa frase es falsa y nunca se midió: `ink-950` sobre `ink-850` daba **1,15:1**. No daba profundidad sin ruido, no daba profundidad. Y no se arregla oscureciéndola — a esa altura de la curva sRGB ya no queda luminancia que sacar. El techo con una sombra **negra pura** es 1,32:1 con el fondo en `L 0.25` y 1,55:1 con el fondo en `L 0.30`; para llegar apenas a 2:1 el fondo tendría que irse a `L 0.365`, que ya es gris y no oscuro. **En oscuro la profundidad se dibuja con luz.**
>
> La primera mitad, en cambio, era una advertencia buena, y está pagada: el borde bajó de `chalk-300` a `chalk-600`. La tiza no se sumó a la pantalla, **se mudó** — del contorno al bloque desplazado. Área brillante total: menos que antes, no más.

Jerarquía resultante en oscuro, de más fuerte a más débil: **texto 12,77 → sombra 8,94 → texto secundario 6,73 → borde 4,41**. Y `check-contrast.mjs` verifica ahora la sombra como par propio, así que la regresión que causó todo esto falla el build en vez de pasar desapercibida.

### Los bloques de color saltan lo mismo en los dos temas

El tercer error, y el más caro visualmente: **un acento no se aclara porque el fondo se oscureció.** Eso duplica el salto y lo convierte en una linterna. El acento es ahora el **mismo** `rail-500` en claro y en oscuro — en claro queda más oscuro que la página, en oscuro más claro.

| Salto contra el fondo | claro | oscuro antes | oscuro ahora |
|---|---|---|---|
| tarjeta de saldo (acento) | 3,24:1 | 8,38:1 | **4,72:1** |
| badge de ganada | 4,76:1 | 13,16:1 | **4,84:1** |
| badge EN VIVO | 4,78:1 | 6,53:1 | **4,79:1** |

`scripts/check-contrast.mjs` verifica ahora **las dos cosas**: el contraste y la separación entre superficies. La segunda no es una regla WCAG, es una regla de este sistema, y está mecanizada para que no vuelva a pasar.

### Contraste — medido, no afirmado

`node scripts/check-contrast.mjs` corre dentro de `verify.mjs` y falla el build si algún par baja del umbral. Estado actual:

| | claro | oscuro |
|---|---|---|
| texto de cuerpo | 17,49:1 | 12,77:1 |
| texto secundario | 5,45:1 | 6,73:1 |
| borde de 3 px | 17,49:1 | 4,41:1 |
| sombra dura | 17,49:1 | 8,94:1 |
| botón primario | 5,40:1 | 5,40:1 |
| badge EN VIVO | 4,78:1 | 5,47:1 |
| badge de ganada | 4,76:1 | 5,53:1 |

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
