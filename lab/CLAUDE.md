# CLAUDE.md — El lab

> Complementa el [`CLAUDE.md` de la raíz](../CLAUDE.md).

Acá se enseña **el concepto de cada sesión, aislado**. Una ruta por sesión, sin el modelo de dominio del hipódromo en el medio.

Es donde pasan tres de los doce bloques: el live coding (0:20), la Misión 1 (0:35) y «predice y ejecuta» (1:10).

---

## El mundo del lab

**El dominio es una cafetería, *Café Compilado*, y se mantiene en las once rutas.** Cambiar de dominio cada sesión obliga a re-entender el problema antes de mirar el concepto, que es justo lo que el lab existe para evitar.

**El lab no usa `shared/ui` del hipódromo.** HTML y CSS a mano, sin componentes propios salvo el de la sesión. La idea es que no haya nada entre el alumno y el concepto: si `<app-button>` ya existe, el botón deja de ser HTML y pasa a ser magia.

Lo único que el lab comparte con el hipódromo es la **paleta** (`styles.css` con el bloque de tokens generado) y las **fuentes**. Las dos vienen de `docs/design/`.

---

## Estructura

```
src/app/
├── app.component.*        el armazón: barra lateral + router-outlet
├── app.routes.ts          una ruta por sesión, con loadComponent
├── sesiones.ts            el índice: número, slug, título, concepto, disponible
└── sesiones/
    └── sNN/               s01.component.ts · .html · .css · .spec.ts
```

### Cómo sumar una sesión

Son **dos lugares** y hay que tocar los dos:

1. `sesiones.ts` — poner `disponible: true`. De ahí sale la **barra lateral**.
2. `app.routes.ts` — sumar la ruta con `loadComponent`. De ahí sale la **navegación**.

Olvidarse del segundo da una entrada de menú que no lleva a ningún lado; olvidarse del primero, una ruta que existe pero no se ve.

### La barra lateral

En **`solution/`** están las once sesiones: las dadas navegables y las futuras en gris. Sirve de mapa del curso.

En **`starter/`** la barra arranca **vacía**: el alumno la ve crecer a medida que suma sesiones. Ver aparecer su propia navegación es parte de entender que las rutas las declara alguien, no el framework.

---

## Tests

`solution/` testea **los bindings y lo que se ve**: es la referencia de qué tiene que lograr la Misión 1.

`starter/` testea **solo la lógica que ya viene hecha**, para que pase desde el minuto cero. Lo visual va en los criterios de «Listo cuando» del enunciado.

Un proyecto sin ningún `.spec.ts` hace fallar a `ng test` con `TS18003`. `verify.mjs` ya saltea esos casos, pero si agregás una sesión, agregale su spec.

```bash
npm start        # http://localhost:4200
npm test
```
