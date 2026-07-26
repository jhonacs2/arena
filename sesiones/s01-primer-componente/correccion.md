# S1 · Corrección — de la pantalla en blanco a la pantalla que anda

Para tres momentos:

- **El instructor**, en el code review de las 1:45, con esto en pantalla al lado de la solución del alumno.
- **Quien se trabó** y necesita destrabarse sin que le resuelvan todo.
- **Cualquiera, después**, para comparar contra lo que hizo.

> Cada paso dice **qué se escribe**, **dónde** y **por qué ahí**. El «por qué» es lo único que no se puede sacar mirando `solution/`.

---

# Parte A · Misión 1 — el mostrador, en `lab/starter`

## A1 · Crear el componente

```bash
cd lab/starter
ng generate component sesiones/s01 --flat
```

Aparecen cuatro archivos en `src/app/sesiones/s01/`: el `.ts`, el `.html`, el `.css` y el `.spec.ts`.

**Por qué con el CLI y no a mano:** los nombra igual siempre y no se olvida de ninguno. No hace nada mágico — se pueden crear a mano y funciona igual.

**Por qué `--flat`:** sin eso, el CLI crea *otra* carpeta `s01` adentro de `s01`.

## A2 · Que la pantalla exista

Son **dos archivos** y hay que tocar los dos. Este es el paso que más se olvida.

**`src/app/app.routes.ts`** — para que la dirección `/s01` lleve a algún lado:

```ts
{
  path: 's01',
  title: 'S1 · Primer componente · Lab',
  loadComponent: () => import('./sesiones/s01/s01.component').then((m) => m.S01Component),
},
```

Va **antes** de `{ path: '**' }`. El router toma la primera ruta que coincide, y el comodín coincide con todo: declarado arriba se comería a las demás.

**`src/app/sesiones.ts`** — para que aparezca en la barra lateral:

```ts
{ numero: 1, slug: 's01', …, disponible: true },
```

**Por qué en dos lugares:** el router resuelve direcciones; el menú es una lista que dibuja el componente de la barra. Son dos cosas distintas y ninguna se entera de la otra. Si solo hacés la ruta, la página existe pero nadie la encuentra; si solo hacés el índice, el enlace no lleva a ningún lado.

> **Comprobalo ahora.** `npm start` y entrá a `/s01`. Tiene que verse la página vacía que dejó el CLI y «Primer componente» en la barra. Si no, no sigas: lo que viene se apoya en esto.

## A3 · Los datos, en la clase

En `s01.component.ts`:

```ts
protected cafe = {
  nombre: 'Yirgacheffe',
  origen: 'Etiopía',
  precio: 42,
  disponible: true,
};
```

**Por qué `protected`:** lo puede leer el template pero no otro componente desde afuera. `private` no serviría — el template no lo vería.

**Por qué en la clase y no en el HTML:** para que exista **un solo lugar** donde está el precio. Es todo el punto de lo que sigue.

## A4 · Interpolación — que se vea

En `s01.component.html`:

```html
<h2>{{ cafe.nombre }}</h2>
<p class="producto__origen">{{ cafe.origen }}</p>
<p class="producto__precio num">{{ cafe.precio }}</p>
```

**La comprobación que importa:** cambiá `precio: 42` por `precio: 55` en el `.ts` y guardá. Si la pantalla no cambia sola, la interpolación no está puesta.

## A5 · Property binding — la clase condicional

```html
<div class="producto" [class.producto--agotado]="!cafe.disponible">
```

**Por qué los corchetes:** sin ellos, `class.producto--agotado="..."` sería un atributo con texto literal. Con ellos, lo de las comillas es **una expresión de TypeScript** que Angular evalúa.

**Por qué se pueden usar `class` y `[class.x]` juntos:** no compiten. `class` pone las que van siempre; `[class.x]` es dueño de **esa** clase y de ninguna más.

Y el texto de estado:

```html
<p class="producto__estado">{{ cafe.disponible ? 'Disponible' : 'Agotado' }}</p>
```

## A6 · Event binding — el botón

En la clase:

```ts
protected alternarDisponibilidad(): void {
  this.cafe = { ...this.cafe, disponible: !this.cafe.disponible };
}
```

En el template:

```html
<button type="button" class="boton boton--fantasma" (click)="alternarDisponibilidad()">
  {{ cafe.disponible ? 'Marcar agotado' : 'Marcar disponible' }}
</button>
```

**Por qué `{ ...this.cafe }` y no `this.cafe.disponible = !...`:** se crea un objeto nuevo en vez de modificar el que estaba. Es una regla del curso, y en S3 va a ser la diferencia entre que la vista se actualice y que no. Hoy alcanza con hacerlo.

**Por qué se actualizan el botón *y* el div:** porque hubo un clic. Angular repinta **después de que pasa algo**, y revisa todos los bindings de la pantalla, no solo el que tocaste.

## A7 · Two-way binding — y el error que hay que ver

```ts
protected cliente = '';
protected cantidad = 1;
```

```html
<input name="cliente" type="text" autocomplete="off" [(ngModel)]="cliente" />
<input name="cantidad" type="number" min="1" max="20" [(ngModel)]="cantidad" />
```

**Va a fallar**, y está bien:

```
NG8002: Can't bind to 'ngModel' since it isn't a known property of 'input'.
```

Se arregla en el `@Component`:

```ts
import { FormsModule } from '@angular/forms';

@Component({
  // …
  imports: [FormsModule],
})
```

**Por qué:** `ngModel` no existe en HTML, lo trae Angular. Y este componente es **standalone**: declara solo lo que usa. Si no está en `imports`, Angular no lo conoce.

**La regla completa:** si el template usa algo de Angular, tiene que estar en `imports`. Vale para el router y, desde S2, para cada componente que se use adentro de otro.

## A8 · Lo que se calcula solo

```ts
protected get total(): number {
  return this.cafe.precio * this.cantidad;
}

protected get puedeAgregar(): boolean {
  return this.cafe.disponible && this.cantidad > 0 && this.cliente.trim().length > 0;
}
```

```html
<p class="pedido__total">
  Total: <span class="num">{{ total }}</span>
  @if (cliente) {
    <span class="pedido__para">para {{ cliente }}</span>
  }
</p>

<button type="submit" class="boton" [disabled]="!puedeAgregar">Agregar al pedido</button>
```

**Por qué un `get` y no una propiedad:** un `get` se recalcula cada vez que Angular repinta, así que siempre está al día. Una propiedad habría que acordarse de actualizarla en cada lugar que cambia el precio o la cantidad.

**Por qué `[disabled]` y no `disabled`:** `disabled="algo"` deshabilita **siempre**, porque cualquier texto que no sea vacío es verdadero para el navegador. Es exactamente el error de la pregunta 2 del exit ticket.

## A9 · La comanda

```ts
protected pedidos: readonly string[] = [];

protected agregar(): void {
  if (!this.puedeAgregar) return;

  this.pedidos = [...this.pedidos, `${this.cantidad} × ${this.cafe.nombre} para ${this.cliente.trim()}`];

  this.cliente = '';
  this.cantidad = 1;
}
```

**Por qué `[...this.pedidos, x]` y no `push`:** misma regla que en A6. En S3 se explica; hoy se cumple.

---

# Parte B · Misión 2 — el listado de carreras, en `project/frontend/starter`

Misma mecánica, con los datos del proyecto.

## B1 · Crear el componente

```bash
cd project/frontend/starter
ng generate component features/races/race-list --flat
```

**Por qué en `features/`:** es una pantalla del producto. La regla de dependencias dice que `features/` puede usar `core/` y `shared/`, y nunca al revés.

## B2 · La ruta y el enlace

**`app.routes.ts`** — antes del comodín:

```ts
{
  path: 'carreras',
  title: 'Carreras · Hipódromo',
  loadComponent: () =>
    import('./features/races/race-list.component').then((m) => m.RaceListComponent),
},
```

Y cambiá los dos `redirectTo: 'sistema'` por `'carreras'`.

**`layout/shell.component.html`** — el enlace en el encabezado:

```html
<a routerLink="/carreras" routerLinkActive="activo">Carreras</a>
```

## B3 · Preparar los datos, en la clase

```ts
import { RACES } from '../../core/mocks';
import { favourite, potentialPayout, type Horse, type Race } from '../../core/models';

interface CarreraVista {
  readonly carrera: Race;
  readonly hora: string;
  readonly favorito: Horse | undefined;
  readonly etiquetaEstado: string;
}

const ETIQUETAS: Record<Race['status'], string> = {
  upcoming: 'Por largar',
  live: 'En vivo',
  finished: 'Terminada',
};
```

```ts
protected readonly carreras: readonly CarreraVista[] = RACES.map((carrera) => ({
  carrera,
  hora: this.formatearHora(carrera.startsAt),
  favorito: favourite(carrera),
  etiquetaEstado: ETIQUETAS[carrera.status],
}));

private formatearHora(iso: string): string {
  return new Intl.DateTimeFormat('es', {
    day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit',
  }).format(new Date(iso));
}
```

**Por qué se prepara acá y no en el template:** lo que se calcula en el HTML se recalcula en **cada** detección de cambios. Acá se calcula una vez.

**Por qué `Horse | undefined`:** `favourite()` no puede garantizar que haya un caballo. TypeScript obliga a decidir qué pasa si no hay.

**Por qué `Intl.DateTimeFormat` y no un pipe:** los pipes son S4. `Intl` es del navegador y conviene verlo una vez a mano antes de que un pipe lo esconda.

## B4 · La lista

```html
<ul class="carreras" role="list">
  @for (vista of carreras; track vista.carrera.id) {
    <li>
      <button
        type="button"
        class="carrera"
        [class.carrera--viva]="vista.carrera.status === 'live'"
        [class.carrera--terminada]="vista.carrera.status === 'finished'"
        [class.carrera--abierta]="seleccionada?.carrera?.id === vista.carrera.id"
        [attr.aria-pressed]="seleccionada?.carrera?.id === vista.carrera.id"
        (click)="seleccionar(vista)"
      >
        <span class="carrera__estado">{{ vista.etiquetaEstado }}</span>
        <span class="carrera__nombre">{{ vista.carrera.name }}</span>
        <span class="carrera__hora num">{{ vista.hora }}</span>
        <span class="carrera__pie">
          {{ vista.carrera.horses.length }} competidores
          @if (vista.favorito) {
            · favorito {{ vista.favorito.name }} a {{ vista.favorito.odds.toFixed(2) }}
          }
        </span>
      </button>
    </li>
  }
</ul>
```

**Por qué `<button>` y no `<div>`:** un botón se tabula, se activa con Enter y con la barra espaciadora, y los lectores de pantalla lo anuncian como control. Un `div` con `(click)` no hace nada de eso, y hay que reimplementarlo todo a mano.

**Por qué `[attr.aria-pressed]` y no `[aria-pressed]`:** `aria-*` son **atributos**, no propiedades del elemento. `[attr.…]` es la forma de escribir un atributo.

**Sobre el `@for`:** recorre la lista. `track` le dice a Angular cómo reconocer cada elemento entre repintados. Se ve a fondo en S3.

## B5 · La selección

```ts
protected seleccionada: CarreraVista | null = null;

protected seleccionar(vista: CarreraVista): void {
  this.seleccionada = this.seleccionada?.carrera.id === vista.carrera.id ? null : vista;
}
```

**Por qué se deselecciona al tocar de nuevo:** es lo que espera cualquiera y ahorra un botón de cerrar.

## B6 · El panel y el simulador

```ts
protected monto = 100;

protected get pagoPotencial(): number {
  const cuota = this.seleccionada?.favorito?.odds ?? 0;
  return potentialPayout(this.monto, cuota);
}
```

```html
<input name="monto" type="number" min="10" max="5000" step="10" [(ngModel)]="monto" />
<p class="simulador__pago">
  {{ seleccionada.favorito.name }} paga <strong class="num">{{ pagoPotencial }}</strong>
</p>
```

Y `FormsModule` en los `imports`, igual que en el lab.

**Por qué `?? 0`:** puede no haber carrera seleccionada, y puede no haber favorito. Los dos `?.` cortan la cadena y el `?? 0` da un valor por defecto.

---

## Lo que se revisa en el bloque de las 1:45

| | Qué mirar |
|---|---|
| 1 | `standalone: true` y `changeDetection: OnPush` en el `@Component` |
| 2 | Estado actualizado sin mutar: `{ ...obj }` y `[...arr, x]` |
| 3 | Sin `any`, sin `console.log`, sin imports que sobren |
| 4 | Los datos preparados en la clase, no calculados en el HTML |
| 5 | `<button>` para lo que se toca, no `<div>` |
| 6 | El componente en `features/`, no en `shared/` ni en `core/` |

**No aplica hoy:** los estados de carga, vacío y error. No hay nada que cargar — los datos están en el código. Vuelven en S7 y ahí son obligatorios.
