# SNN · Corrección — de la pantalla en blanco a la pantalla que anda

Para tres momentos:

- **El instructor**, en el code review de las 1:45, con esto en pantalla al lado de la solución del alumno.
- **Quien se trabó** y necesita destrabarse sin que le resuelvan todo.
- **Cualquiera, después**, para comparar contra lo que hizo.

> Cada paso dice **qué se escribe**, **dónde** y **por qué ahí**.
> El «por qué» es lo único que no se puede sacar mirando `solution/`. Un paso que solo muestra código copiado no sirve.

---

# Parte A · Misión 1 — en `lab/starter`

## A1 · Crear el componente

```bash
cd lab/starter
ng generate component sesiones/sNN --flat
```

*Por qué con el CLI: los nombra igual siempre y no se olvida de ningún archivo. No hace nada mágico.*

## A2 · Que la pantalla exista

Son **dos archivos** y hay que tocar los dos:

- `app.routes.ts` — para que la dirección exista. Va antes del comodín `**`.
- `sessions.ts` — `available: true`, para que aparezca en la barra lateral.

*Por qué dos: el router resuelve direcciones; el menú es una lista que dibuja un componente. Ninguno se entera del otro.*

> **Comprobalo antes de seguir.** Lo que viene se apoya en esto.

## A3 · …

*Un paso por concepto de la sesión, en el mismo orden que el live coding. Cada uno con su «por qué», y con la comprobación concreta cuando la haya: «cambiá X y mirá que pasa Y».*

---

# Parte B · Misión 2 — en `project/frontend/starter`

## B1 · Crear el componente

```bash
cd project/frontend/starter
ng generate component features/…/… --flat
```

*Por qué en `features/`: es una pantalla del producto. `features/` puede usar `core/` y `shared/`, nunca al revés.*

## B2 · …

---

## Lo que se revisa en el bloque de las 1:45

| | Qué mirar |
|---|---|
| 1 | `standalone: true` y `changeDetection: OnPush` |
| 2 | Estado actualizado sin mutar |
| 3 | Sin `any`, sin `console.log`, sin imports que sobren |
| 4 | Los datos preparados en la clase, no calculados en el HTML |
| 5 | `<button>` para lo que se toca, no `<div>` |
| 6 | El componente en la carpeta que le toca |

**Lo que no aplica todavía:** *anotar acá lo que la rúbrica pide pero esta sesión no puede cumplir, y por qué. Por ejemplo, los tres estados antes de S7.*
