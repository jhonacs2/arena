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
ng generate component sessions/s01 --flat
```

Aparecen cuatro archivos en `src/app/sessions/s01/`: el `.ts`, el `.html`, el `.css` y el `.spec.ts`.

**Por qué con el CLI y no a mano:** los nombra igual siempre y no se olvida de ninguno. No hace nada mágico — se pueden crear a mano y funciona igual.

**Por qué `--flat`:** sin eso, el CLI crea *otra* carpeta `s01` adentro de `s01`.

## A2 · Que la pantalla exista

Son **dos archivos** y hay que tocar los dos. Este es el paso que más se olvida.

**`src/app/app.routes.ts`** — para que la dirección `/s01` lleve a algún lado:

```ts
{
  path: 's01',
  title: 'S1 · Primer componente · Lab',
  loadComponent: () => import('./sessions/s01/s01.component').then((m) => m.S01Component),
},
```

Va **antes** de `{ path: '**' }`. El router toma la primera ruta que coincide, y el comodín coincide con todo: declarado arriba se comería a las demás.

**`src/app/sessions.ts`** — para que aparezca en la barra lateral:

```ts
{ numero: 1, slug: 's01', …, available: true },
```

**Por qué en dos lugares:** el router resuelve direcciones; el menú es una lista que dibuja el componente de la barra. Son dos cosas distintas y ninguna se entera de la otra. Si solo hacés la ruta, la página existe pero nadie la encuentra; si solo hacés el índice, el enlace no lleva a ningún lado.

> **Comprobalo ahora.** `npm start` y entrá a `/s01`. Tiene que verse la página vacía que dejó el CLI y «Primer componente» en la barra. Si no, no sigas: lo que viene se apoya en esto.

## A3 · Los datos, en la clase

En `s01.component.ts`:

```ts
protected coffee = {
  name: 'Yirgacheffe',
  origin: 'Etiopía',
  price: 42,
  available: true,
};
```

**Por qué `protected`:** lo puede leer el template pero no otro componente desde afuera. `private` no serviría — el template no lo vería.

**Por qué en la clase y no en el HTML:** para que exista **un solo lugar** donde está el precio. Es todo el punto de lo que sigue.

## A4 · Interpolación — que se vea

En `s01.component.html`:

```html
<h2>{{ coffee.name }}</h2>
<p class="producto__origen">{{ coffee.origin }}</p>
<p class="producto__precio num">{{ coffee.price }}</p>
```

**La comprobación que importa:** cambiá `price: 42` por `price: 55` en el `.ts` y guardá. Si la pantalla no cambia sola, la interpolación no está puesta.

## A5 · Property binding — la clase condicional

```html
<div class="product" [class.product--soldout]="!coffee.available">
```

**Por qué los corchetes:** sin ellos, `class.product--soldout="..."` sería un atributo con texto literal. Con ellos, lo de las comillas es **una expresión de TypeScript** que Angular evalúa.

**Por qué se pueden usar `class` y `[class.x]` juntos:** no compiten. `class` pone las que van siempre; `[class.x]` es dueño de **esa** clase y de ninguna más.

Y el texto de estado:

```html
<p class="product__status">{{ coffee.available ? 'Disponible' : 'Agotado' }}</p>
```

## A6 · Event binding — el botón

En la clase:

```ts
protected toggleAvailability(): void {
  this.coffee = { ...this.coffee, available: !this.coffee.available };
}
```

En el template:

```html
<button type="button" class="button button--ghost" (click)="toggleAvailability()">
  {{ coffee.available ? 'Marcar agotado' : 'Marcar available' }}
</button>
```

**Por qué `{ ...this.coffee }` y no `this.coffee.available = !...`:** se crea un objeto nuevo en vez de modificar el que estaba. Es una regla del curso, y en S3 va a ser la diferencia entre que la vista se actualice y que no. Hoy alcanza con hacerlo.

**Por qué se actualizan el botón *y* el div:** porque hubo un clic. Angular repinta **después de que pasa algo**, y revisa todos los bindings de la pantalla, no solo el que tocaste.

## A7 · Two-way binding — y el error que hay que ver

```ts
protected customer = '';
protected quantity = 1;
```

```html
<input name="customer" type="text" autocomplete="off" [(ngModel)]="customer" />
<input name="quantity" type="number" min="1" max="20" [(ngModel)]="quantity" />
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
  return this.coffee.price * this.quantity;
}

protected get canAddOrder(): boolean {
  return this.coffee.available && this.quantity > 0 && this.customer.trim().length > 0;
}
```

```html
<p class="pedido__total">
  Total: <span class="num">{{ total }}</span>
  @if (customer) {
    <span class="pedido__para">para {{ customer }}</span>
  }
</p>

<button type="submit" class="button" [disabled]="!canAddOrder">Agregar al order</button>
```

**Por qué un `get` y no una propiedad:** un `get` se recalcula cada vez que Angular repinta, así que siempre está al día. Una propiedad habría que acordarse de actualizarla en cada lugar que cambia el precio o la cantidad.

**Por qué `[disabled]` y no `disabled`:** `disabled="algo"` deshabilita **siempre**, porque cualquier texto que no sea vacío es verdadero para el navegador. Es exactamente el error de la pregunta 2 del exit ticket.

## A9 · La comanda

```ts
protected orders: readonly string[] = [];

protected addOrder(): void {
  if (!this.canAddOrder) return;

  this.orders = [...this.orders, `${this.quantity} × ${this.coffee.name} para ${this.customer.trim()}`];

  this.customer = '';
  this.quantity = 1;
}
```

**Por qué `[...this.orders, x]` y no `push`:** misma regla que en A6. En S3 se explica; hoy se cumple.

---

# Parte B · Misión 2 — el listado de carreras, en `project/frontend/starter`

Misma mecánica, con los datos del proyecto.

## B1 · Crear el componente

```bash
cd project/frontend/starter
ng generate component features/carreras/race-list --flat
```

**Por qué en `features/`:** es una pantalla del producto. La regla de dependencias dice que `features/` puede usar `core/` y `shared/`, y nunca al revés.

## B2 · La ruta y el enlace

**`app.routes.ts`** — antes del comodín:

```ts
{
  path: 'carreras',
  title: 'Carreras · Hipódromo',
  loadComponent: () =>
    import('./features/carreras/race-list.component').then((m) => m.RaceListComponent),
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

interface RaceView {
  readonly race: Race;
  readonly hora: string;
  readonly favourite: Horse | undefined;
  readonly statusLabel: string;
}

const STATUS_LABELS: Record<Race['status'], string> = {
  upcoming: 'Por largar',
  live: 'En vivo',
  finished: 'Terminada',
};
```

```ts
protected readonly races: readonly RaceView[] = RACES.map((race) => ({
  race,
  hora: this.formatTime(race.startsAt),
  favourite: favourite(race),
  statusLabel: STATUS_LABELS[race.status],
}));

private formatTime(iso: string): string {
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
<ul class="races" role="list">
  @for (view of races; track view.race.id) {
    <li>
      <button
        type="button"
        class="race"
        [class.race--live]="view.race.status === 'live'"
        [class.race--finished]="view.race.status === 'finished'"
        [class.race--open]="selected?.race?.id === view.race.id"
        [attr.aria-pressed]="selected?.race?.id === view.race.id"
        (click)="select(view)"
      >
        <span class="carrera__estado">{{ view.statusLabel }}</span>
        <span class="carrera__nombre">{{ view.race.name }}</span>
        <span class="carrera__hora num">{{ view.hora }}</span>
        <span class="carrera__pie">
          {{ view.race.horses.length }} competidores
          @if (view.favourite) {
            · favourite {{ view.favourite.name }} a {{ view.favourite.odds.toFixed(2) }}
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
protected selected: RaceView | null = null;

protected select(view: RaceView): void {
  this.selected = this.selected?.race.id === view.race.id ? null : view;
}
```

**Por qué se deselecciona al tocar de nuevo:** es lo que espera cualquiera y ahorra un botón de cerrar.

## B6 · El panel y el simulador

```ts
protected amount = 100;

protected get payout(): number {
  const odds = this.selected?.favourite?.odds ?? 0;
  return potentialPayout(this.amount, odds);
}
```

```html
<input name="amount" type="number" min="10" max="5000" step="10" [(ngModel)]="amount" />
<p class="simulador__pago">
  {{ selected.favourite.name }} paga <strong class="num">{{ payout }}</strong>
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
