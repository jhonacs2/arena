# S0 · TypeScript

**La primera clase del módulo. 2 horas, en vivo.**

El material está en los archivos de siempre — `guion.md` es el documento maestro
y todo lo demás sale de ahí. Este README tiene lo único que no entra en ninguno
de ellos: **lo que hay que tener instalado antes de entrar**.

> **Antes esto era un tema asíncrono.** Se convirtió en clase en vivo porque el
> vocabulario de tipos es lo que sostiene las once sesiones que siguen, y leerlo
> solo no es lo mismo que ver aparecer y desaparecer un error en pantalla. El
> contenido de aquel material está repartido entre `guion.md` y `conceptos.md`.

---

## Antes de la primera clase — 20 minutos

**Esto se manda por el canal del curso una semana antes, y se pregunta en el
chat el día anterior.** Los quince minutos de instalar cosas en vivo son quince
minutos de clase perdidos para todos.

### 1 · Node

Versión **20.11 o superior**, o **22**. El equipo está en 22.22.3.

```bash
node --version
```

Si no está: <https://nodejs.org> — la versión LTS.

### 2 · Un editor con TypeScript

**Visual Studio Code**: <https://code.visualstudio.com>. Cualquier editor sirve,
pero el material asume VS Code en un punto concreto: **apoyar el mouse sobre una
variable y ver qué tipo tiene**. Esa acción aparece cinco veces en la primera
clase.

### 3 · El repositorio

```bash
git clone <la URL del repo del curso>
cd <la carpeta>

cd lab/starter
npm install
npm start
```

Si eso abre <http://localhost:4200> y se ve un pizarrón de cafetería con una
tabla de precios, **estás listo**.

Y el otro proyecto, que se usa en la segunda mitad de la clase:

```bash
cd project/frontend/starter
npm install
```

### 4 · Si algo falla

| Pasa | Probá |
|---|---|
| `npm install` falla con errores de permisos | No uses `sudo`. Instalá Node desde el instalador oficial, no desde el gestor de paquetes del sistema |
| `ng: command not found` | No hace falta instalar Angular globalmente. Todos los comandos del curso van por `npm run` o `npx` |
| El puerto 4200 está ocupado | `npm start -- --port 4300` |
| Levanta pero no se ve nada | Mirá la terminal: si hay un error de compilación, está ahí con archivo y línea |

**Si después de veinte minutos no funciona, escribí al canal con el error
completo copiado.** No lo dejes para el día de la clase.

---

## Qué se lleva el alumno de esta sesión

| | |
|---|---|
| **Concepto único** | Un tipo es el conjunto de valores que algo puede tener, y alguien lo revisa antes de que el código corra |
| **Lab** | `/s00` — el menú de Café Compilado, en `sessions/s00/menu.ts` |
| **Hipódromo** | `core/models/race.model.ts` — los tipos del contrato |
| **Apunte** | `conceptos.md` |

## Qué NO se ve hoy

Ni una línea de Angular. La pantalla del lab viene escrita y el guión lo dice en
voz alta: los componentes son S1.

Tampoco se ven clases, herencia, `enum`, decoradores ni tipos condicionales.
Los decoradores aparecen en S1 con `@Component`, y no hace falta entender cómo
funcionan para usarlos.
