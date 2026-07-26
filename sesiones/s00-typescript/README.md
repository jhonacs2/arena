# Tema 0 · TypeScript — asíncrono

**Antes de S1. 100% asíncrono. Entre 90 y 120 minutos.**

No es una sesión: es el vocabulario mínimo para que S1 no se convierta en una clase de TypeScript. Si alguien llega sin esto, va a pasar la primera hora peleando con `strict` en vez de mirando Angular.

> El material usa **los tipos reales del proyecto**. El alumno tipa el dominio del hipódromo antes de la primera clase, así que llega a S1 reconociendo `Race`, `Horse` y `RaceStatus` en lugar de viéndolos por primera vez.

---

## Qué tiene que quedar

| Tema | Por qué está |
|---|---|
| Tipos primitivos y anotaciones | Todo lo demás se apoya acá |
| `interface` vs `type` | En el proyecto se usan las dos: `interface Race`, `type RaceStatus` |
| Uniones de literales | `'upcoming' \| 'live' \| 'finished'` aparece en la primera pantalla |
| Narrowing | Es lo que hace que el `switch` sobre el estado sea exhaustivo |
| Opcionales y `strictNullChecks` | `favourite()` devuelve `Horse \| undefined`, y hay que hacerse cargo |
| Genéricos básicos | `Page<Race>` y `Page<Bet>` son el mismo tipo con distinto contenido |
| Utility types | `Pick`, `Omit`, `Partial`, `Readonly` |
| `readonly` e inmutabilidad | Es regla del curso y se evalúa |
| Módulos ES | `import` / `export`, que es como está armado todo el proyecto |

---

## Qué hace el alumno

1. Lee o mira el material (enlaces abajo).
2. Abre **`docs/contract/openapi.yaml`** y escribe las interfaces de `Race`, `Horse` y `Bet` a mano, sin mirar `core/models`.
3. Compara con `project/frontend/starter/src/app/core/models/`. Las diferencias son el material de la primera pregunta de S1.
4. Contesta el Wayground — **se corre en el bloque 0:05 de S1**.

## Entrega

Un archivo `tipos.ts` con las tres interfaces. No se corrige una por una: se mira por arriba y lo que aparezca repetido se comenta en clase.

---

## Errores que van a aparecer, y está bien

Los tres que más se repiten, y que conviene tener listos para el bloque 0:05 de S1:

| Lo que escriben | Por qué no alcanza |
|---|---|
| `status: string` | Acepta `'galopando'`. La unión de literales es lo que lo impide. |
| `horses: Horse[]` sin `readonly` | El curso exige inmutabilidad; `readonly Horse[]` la hace explícita. |
| `favorito: Horse` | El contrato admite carreras sin favorito calculable. Va `Horse \| undefined`. |

---

## Material

- **TypeScript Handbook** — *Everyday Types* y *Narrowing*: <https://www.typescriptlang.org/docs/handbook/2/everyday-types.html>
- **TypeScript Playground** para probar sin instalar nada: <https://www.typescriptlang.org/play>
- El propio `docs/contract/openapi.yaml` del proyecto
