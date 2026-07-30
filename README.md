# Hipódromo — Módulo Angular · Talento DH 8va

Workspace del **instructor**. Tiene todo: la solución de referencia, el starter del alumno, el lab, el backend y los materiales de las 11 sesiones.

> Este repo no se publica. El alumno recibe un repo derivado con solo `project/backend`, `project/frontend/starter` y `lab/starter`.

---

## Por dónde empezar

| Si querés… | Leé |
|---|---|
| Entender las reglas del proyecto | [`CLAUDE.md`](CLAUDE.md) |
| Saber qué se construye y en qué orden | [`docs/plan.md`](docs/plan.md) |
| Ver el contrato con el backend | [`docs/contract/README.md`](docs/contract/README.md) |
| Entender el sistema visual | [`docs/design/tokens.md`](docs/design/tokens.md) |
| Preparar una clase | [`docs/curriculum.md`](docs/curriculum.md) y [`sesiones/README.md`](sesiones/README.md) |
| Saber qué imágenes generar | [`docs/design/IMAGES.md`](docs/design/IMAGES.md) |

---

## Verificar

```bash
node scripts/verify.mjs            # todo
node scripts/verify.mjs --fast     # sin builds
node scripts/verify.mjs contrato   # un grupo: contrato | diseño | código
```

Corre después de cada feature. Verifica el contrato, el diseño y el código Angular; saltea con gracia lo que todavía no existe.

## Regenerar lo generado

```bash
node scripts/gen-tokens-css.mjs        # tokens.json → el bloque :root de cada styles.css
node scripts/gen-race-ticks.mjs        # el fixture de la carrera en vivo
node scripts/gen-silks-specimen.mjs    # la hoja de muestra de las sedas
node scripts/check-contrast.mjs        # informe de contraste de la paleta
```

Los cuatro son determinísticos: correrlos dos veces produce el mismo archivo.

## Diapositivas

```bash
npx @marp-team/marp-cli --theme-set theme/ sesiones/s01-*/slides.md -o slides.html
```

---

## Estructura

```
docs/contract/     ⭐ fuente de verdad — si no está acá, no existe
docs/design/       paleta, tipografía, sedas, imágenes a generar
docs/curriculum.md mapa de las 11 sesiones
project/backend/   monolito Go
project/frontend/  solution/ y starter/
lab/               app de enseñanza, una ruta por sesión
sesiones/          _plantilla/ + s00 … s10
scripts/           verificación y generadores
theme/             tema Marp de las diapositivas
```

## Requisitos

Node `^20.11 || ^22` (ver `.nvmrc`) · Go 1.26.x · Angular CLI 18.2.x

---

## Cuentas de prueba

Contraseña para todas: `Carrera123!`

| Correo | Para qué sirve |
|---|---|
| `ana@hipodromo.test` | El usuario por defecto de las demos. Historial mixto, apuestas pendientes. |
| `caro@hipodromo.test` | **Correo sin verificar.** El caso de prueba de `verified.guard` y del `403`. |
| `hugo@hipodromo.test` | Saldo 980, perdió todo. Provoca `INSUFFICIENT_BALANCE` sin armar nada. |
