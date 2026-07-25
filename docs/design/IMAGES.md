# Imágenes a generar

Este archivo es el pedido de trabajo gráfico. Vos generás las imágenes; acá está la especificación exacta de cada una.

**La lista es corta a propósito.** Las sedas de jockey se generan como SVG desde el `id` del caballo (`docs/design/assets/silks-specimen.svg`), así que el elemento visual más repetido de la app —54 caballos, presentes en la lista, el detalle, la carrera en vivo y el historial— **no necesita ni un archivo**. Y la textura de pista que estaba planificada la saqué: los carriles ya dicen "pista".

Quedan seis piezas.

---

## Convenciones

| | |
|---|---|
| Ubicación | `project/frontend/{solution,starter}/public/img/` |
| Formato raster | **AVIF** + fallback **WebP**. Nada de PNG salvo el favicon y el OG. |
| Densidad | `@1x` y `@2x` para todo lo que se muestre a tamaño fijo |
| Paleta | tiene que convivir con `docs/design/tokens.md` — pizarra `#141a17`, tiza `#f7f7f2`, arena `#e39a3d`, bandera `#d92d20` |
| Peso | ninguna imagen por encima de **180 KB** en su versión AVIF |

**Mientras no existan**, la app usa los placeholders SVG de `docs/design/assets/`. No hay imágenes rotas en ningún momento: se puede desarrollar y dar clase con los placeholders puestos.

---

## Las seis

### 1 · `logo-mark.svg` — marca

| | |
|---|---|
| Formato | **SVG**, monocromo, un solo `path` si se puede |
| Tamaño | viewBox cuadrado, legible a **20 px** |
| Dónde | header del shell, favicon, OG, slides de las 11 sesiones |
| Estado | placeholder pendiente — lo genero yo en Fase 2 si querés |

Isotipo, no ilustración. Tiene que funcionar en un color plano sobre pizarra y sobre tiza, sin degradados ni sombras.

Tres direcciones que salen del mundo del hipódromo y no de "app genérica": el **poste de furlong** (raya vertical con bandas), la **puerta de partida** (rejilla de compartimentos) o una **seda abstracta** (el rombo/chevron del sistema, reducido a su mínima expresión). La tercera es la más coherente con el resto del sistema.

---

### 2 · `og-cover.png` — imagen de compartido

| | |
|---|---|
| Formato | **PNG** (los previews de redes no aceptan AVIF) |
| Tamaño | **1200 × 630** exacto |
| Dónde | `<meta property="og:image">`, link en WhatsApp / Slack / X |
| Peso | ≤ 300 KB |

Fondo pizarra sólido. La marca arriba a la izquierda, el nombre "Hipódromo" en display pesado, y una fila de 5–6 sedas del sistema como banda inferior. Nada de fotos de caballos reales: rompería con la identidad plana.

Texto grande, mínimo 60 px — se ve como miniatura en un chat.

---

### 3 · `favicon` — juego completo

| Archivo | Tamaño | Nota |
|---|---|---|
| `favicon.svg` | vectorial | el que usan los navegadores modernos |
| `favicon-32.png` | 32 × 32 | fallback |
| `apple-touch-icon.png` | 180 × 180 | fondo **opaco**, iOS no respeta transparencia |

Derivados directos de `logo-mark.svg`. A 16 px solo sobrevive una forma; si el isotipo tiene más de un elemento, esta versión se queda con uno.

---

### 4 · `auth-hero.avif` — panel de login y registro

| | |
|---|---|
| Formato | AVIF + WebP, `@1x` y `@2x` |
| Tamaño | **1280 × 1600** (vertical 4:5) |
| Dónde | panel lateral de `/login`, `/registro` y `/verificar` |
| Recorte | tiene que sobrevivir un `object-fit: cover` de 4:5 a 1:2 |

La única pieza con peso ilustrativo del proyecto. Escena de largada: la puerta de partida, siluetas planas de caballos y jockeys con sedas del sistema, línea de horizonte baja. Estilo **plano y geométrico**, sin degradados y sin volumen — la misma lógica que las sedas, a mayor escala.

Bordes duros y áreas de color amplias. Debajo va texto: dejá el tercio inferior tranquilo, sin detalle fino.

---

### 5 · `empty-races.svg` · `empty-bets.svg` · `empty-search.svg` — estados vacíos

| | |
|---|---|
| Formato | **SVG**, `currentColor` para la tinta |
| Tamaño | viewBox 240 × 180 |
| Dónde | los tres estados vacíos obligatorios de §8 |
| Estado | los genero yo en Fase 2 |

Van como SVG y no como raster porque tienen que seguir el tema: en oscuro la tinta se invierte a tiza. Un PNG no puede hacer eso.

Línea de 3 px, sin relleno salvo un acento de arena. Un objeto por escena: pizarra de carreras en blanco, boleto sin picar, catalejo.

---

### 6 · `slide-cover.png` — portada de las diapositivas

| | |
|---|---|
| Formato | PNG |
| Tamaño | **1920 × 1080** |
| Dónde | primera diapositiva de cada una de las 11 sesiones |
| Variante | el número de sesión va como texto en Marp, **no** dentro de la imagen |

Una sola imagen para las 11. Fondo pizarra, banda de sedas, mucho aire para que el tema de la sesión se sobreimprima en Marp. Si va texto en la imagen, hay que rehacerla 11 veces.

---

## Lo que NO hay que generar

Vale la pena tenerlo escrito, porque son justo las cosas que uno pediría por reflejo:

| | Por qué no |
|---|---|
| Retratos o siluetas de los 54 caballos | La seda es la identidad. Es SVG generado desde el `id`. |
| Avatares de los 12 usuarios | Iniciales sobre un cuadrado de color de seda, también generado. |
| Iconos de estado (`upcoming` / `live` / `finished`) | Son badges de color y texto. El color ya lo dice. |
| Textura de pista | Sacada a propósito. Los carriles con marcas de furlong ya comunican pista, y la textura solo agregaba peso y riesgo de contraste bajo el marcador. |
| Banderas, trofeos, herraduras, iconos de dinero | No están en el sistema. Los números son la tipografía mono; no necesitan un ícono al lado. |
