# S2 · Corrección — de una pantalla a dos componentes

**Bloque 1:45 · instructor y alumno**

Las dos misiones, paso a paso, con **qué se escribe, dónde y por qué ahí**.

> **Cómo se usa.** En clase, en pantalla mientras se revisa una solución. En
> casa, para destrabarse: buscá el paso donde estás y seguí desde ahí.

---

# Parte A · La carta del lab

**Archivo:** `lab/starter/src/app/sessions/s02/`
**Referencia terminada:** `lab/solution/src/app/sessions/s02/`

## Paso 0 · Que la pantalla exista

```ts
// app.routes.ts
{
  path: 's02',
  title: 'S2 · Anatomía de un componente · Lab',
  loadComponent: () => import('./sessions/s02/s02.component').then((m) => m.S02Component),
},
```

Más `available: true` en `sessions.ts`. **Son dos archivos**, igual que en S1.

## Paso 1 · Cortar antes de pensar

```bash
ng generate component sessions/s02/coffee-card --flat
```

**Cortá** el `<article class="card">` entero del padre y pegalo en el template
del hijo. Guardá **sin arreglar nada**.

> **Por qué en este orden.** Aparecen unos quince errores, todos diciendo que
> `item` no existe. Esa lista **es** la lista de inputs que hacen falta: no hay
> que adivinarla, ya está escrita en la terminal.
>
> Quien intenta declarar los inputs antes de cortar se olvida de dos y descubre
> cuáles diez minutos después.

Y el CSS se muda con el marcado, al `coffee-card.component.css`.

> **Por qué se muda.** Un componente cuyo estilo vive en el archivo de otro no se
> puede llevar a ninguna parte — que es exactamente lo que veníamos a arreglar.

## Paso 2 · Las entradas

```ts
readonly coffee = input.required<Coffee>();
readonly featured = input(false);
readonly quantity = model(1);
```

Y en el template del hijo, `item.coffee` pasa a ser `coffee()`.

**Por qué `required` en `coffee` y no en `featured`.** Sin café no hay tarjeta que
dibujar: es un error de programación y conviene que no compile. Sin `featured`
hay una tarjeta perfectamente válida, la que no está destacada — por eso lleva
valor por defecto.

**Por qué `quantity` es `model()` y no `input()`.** Porque la cambian los dos: el
padre la inicializa y el hijo la sube con los botones. Un `input()` no puede
volver.

## Paso 3 · La salida

```ts
readonly ordered = output<OrderRequest>();

protected order(): void {
  if (!this.coffee().available) return;
  this.ordered.emit({ coffee: this.coffee(), quantity: this.quantity() });
}
```

En el padre:

```html
(ordered)="take($event)"
```

**Por qué el hijo no escribe en la comanda.** Es la decisión central de la
sesión. Si escribiera, la tarjeta necesitaría conocer la comanda — y entonces
solo serviría en pantallas que tengan una. Avisando, sirve en cualquiera.

> **El error más común:** llevarse la comanda al hijo. Se detecta pidiendo dos
> cafés distintos: aparecen dos comandas y ninguna tiene todo.

## Paso 4 · Los dos huecos

```html
<div class="card__tag">
  <ng-content select="[card-tag]" />
</div>
…
<div class="card__slot">
  <ng-content />
</div>
```

**Por qué el rótulo no es un input de texto.** Porque es marcado, no un dato: el
hijo no lo usa para decidir nada, solo lo muestra. Con `ng-content`, el día que
haya que ponerle un icono al lado no hay que tocar la tarjeta.

Y en el CSS, los dos huecos se esconden cuando no reciben nada:

```css
.card__tag:empty { display: none; }
.card__slot:empty { display: none; }
```

Sin eso, las tres tarjetas que no proyectan nada arriba quedan con un renglón en
blanco.

## Paso 5 · El ciclo de vida

```ts
export class CoffeeCardComponent implements OnInit, OnChanges, OnDestroy {
```

**Por qué la hora se guarda en `ngOnInit` y no en la declaración de la
propiedad.** Ahí funcionaría igual para la hora, pero no para nada que dependa
de un input: cuando el constructor corre, los inputs todavía están vacíos.
`ngOnInit` es el primer momento en que ya llegaron.

**Y el detalle que sorprende:** al subir la cantidad desde la tarjeta, su propio
`ngOnChanges` corre. El hijo emite → el padre guarda → el valor **vuelve a
bajar** por el mismo binding. Da toda la vuelta, y para el hijo es un input que
cambió.

---

# Parte B · La tarjeta del hipódromo

**Punto de partida:** el listado de S1.
**Referencia terminada:** `project/frontend/solution/src/app/`

## Paso 1 · `<app-badge>`, en `shared/ui/badge/`

```ts
export type BadgeTone = 'neutral' | 'live' | 'success' | 'accent';

@Component({
  selector: 'app-badge',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <span [class]="'badge badge--' + tone()">
      <ng-content />
    </span>
  `,
  styleUrl: './badge.component.css',
})
export class BadgeComponent {
  readonly tone = input<BadgeTone>('neutral');
}
```

**Por qué el texto entra por `ng-content` y el tono por `input()`.** El texto es
contenido y el hijo solo lo muestra. El tono es un **dato con el que el hijo
decide** —qué clase CSS poner—, y su conjunto de valores está cerrado. Esa es la
regla completa, y es la misma pregunta de S0: ¿cuáles valores, exactamente?

**Por qué va en `shared/ui/` y se exporta desde `index.ts`.** Es una primitiva:
sirve para un estado de carrera, para el resultado de una apuesta y para
cualquier cosa que venga. Si en su archivo apareciera `Race`, dejaría de serlo.

## Paso 2 · `<app-race-card>`, en `features/races/`

```ts
readonly race = input.required<Race>();
readonly time = input.required<string>();
readonly favourite = input<Horse | undefined>(undefined);
readonly selected = input(false);

readonly toggled = output<Race>();
```

**Por qué `time` es un input y no se calcula adentro.** Formatear una fecha es
una decisión de presentación de esta aplicación, no de esta tarjeta. Si la
tarjeta llamara a `Intl.DateTimeFormat`, la pantalla que quiera mostrar «en 8
min» en vez de «30 jul, 14:23» tendría que pelearse con ella.

**Por qué `selected` es un input y no estado propio.** Solo puede haber una
abierta a la vez. La tarjeta no ve a las otras siete, así que no puede sostener
esa regla. El listado sí.

**Por qué `favourite` y no calcularlo con `favourite(race)`.** Porque importar
`favourite()` ataría la tarjeta al módulo de dominio. Hoy no parece grave;
en S7, cuando el favorito venga calculado del servidor, se agradece.

## Paso 3 · Los dos huecos y el CSS que se muda

```html
<span class="race__status">
  <ng-content select="[card-status]" />
</span>
…
<span class="race__extra">
  <ng-content />
</span>
```

```css
.race__extra:empty { display: none; }
```

**Todas las reglas `.race*` se van a `race-card.component.css`.** Si quedan en el
listado, la tarjeta se ve bien únicamente adentro del listado, y el ejercicio no
sirvió de nada.

Es el paso que más se olvida y el más fácil de verificar: **en
`race-list.component.css` no puede quedar ni una regla que empiece con `.race`.**

## Paso 4 · El listado se queda con lo suyo

```ts
const STATUS_TONES: Record<Race['status'], BadgeTone> = {
  upcoming: 'neutral',
  live: 'live',
  finished: 'neutral',
};
```

```html
<app-race-card
  [race]="view.race"
  [time]="view.time"
  [favourite]="view.favourite"
  [selected]="selected?.race?.id === view.race.id"
  (toggled)="select(view)"
>
  <app-badge card-status [tone]="view.tone">{{ view.statusLabel }}</app-badge>

  @if (view.race.status === 'live') {
    <span class="live-note">Se está corriendo ahora</span>
  }
</app-race-card>
```

**La traducción de estado a tono va acá**, al lado de `STATUS_LABELS`, porque es
una decisión de esta pantalla. La pastilla no sabe qué es `live`; el listado sí.

Y el `.live-note` se queda en el CSS del **listado**, aunque se vea adentro de la
tarjeta: lo proyecta el listado, así que es suyo.

---

## Los cinco errores que aparecen todos los años

| Lo que hacen | Por qué no alcanza | Qué decirles |
|---|---|---|
| Se llevan la comanda al hijo | Cada tarjeta tiene su propia lista | «Pedí dos cafés distintos y mirá cuántas comandas hay.» |
| Dejan el CSS en el padre | La tarjeta solo se ve bien ahí | «Llevala a otra pantalla y contame qué ves.» |
| `race-card` importa `RACES` | Sirve para una sola lista | «Tapá el padre. ¿Se entiende sola?» |
| El texto de la pastilla como `input()` | Un input nuevo por cada cosa que quieran meter | «Ponele un icono al lado. ¿Cuánto tardás?» |
| `selected` como estado del hijo | Se abren dos a la vez | «Tocá dos carreras seguidas.» |

---

## Cómo se verifica que quedó bien

```bash
cd lab/starter              && npx tsc --noEmit && npm test
cd project/frontend/starter && npx tsc --noEmit && npm run build
```

Y desde la raíz:

```bash
node scripts/verify.mjs --fast
```

Y a ojo, que es lo que el verificador no puede hacer:

- La pantalla se ve **igual** que antes de empezar.
- `grep -c "^\.race" race-list.component.css` devuelve **0**.
- `badge.component.ts` no dice `Race` ni una vez.
