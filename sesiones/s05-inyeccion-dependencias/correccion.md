# S5 · Corrección — el estado encuentra su dueño

**Bloque 1:45 · instructor y alumno**

---

# Parte A · El mostrador del lab

**Referencia terminada:** `lab/solution/src/app/sessions/s05/`

## Paso 1 · La comanda, a un servicio

```ts
@Injectable({ providedIn: 'root' })
export class OrderService {
  private readonly _orders = signal<readonly Order[]>([]);
  private nextId = 1;

  readonly orders = this._orders.asReadonly();
  readonly count = computed(() => this.orders().length);
  readonly lastCustomer = computed(() => this.orders().at(-1)?.customer ?? '');

  add(customer: string, coffee: string): void {
    const name = customer.trim();
    if (name === '') return;
    this._orders.update((orders) => [...orders, { id: this.nextId++, customer: name, coffee }]);
  }
}
```

**Es una mudanza, no una reescritura.** El `signal` y el `computed` son los mismos
de S3; lo único que cambia es de quién son.

**Por qué el signal es privado y afuera va `asReadonly()`.** Si el componente
pudiera llamar a `set`, el servicio dejaría de ser el dueño de su estado y pasaría
a ser un lugar donde cualquiera escribe. Con `asReadonly()`, la única forma de
cambiar la comanda es `add()` — y entonces la validación del nombre vacío está
garantizada, porque no hay otro camino.

**Por qué la validación vive en el servicio y no en el componente.** Porque si
vive en el componente, el segundo componente que llame a `add()` no la tiene.

## Paso 2 · El cuaderno, uno por mostrador

```ts
@Injectable()                            // ← SIN providedIn
export class NotepadService { … }
```

```ts
@Component({
  …
  providers: [NotepadService],           // ← una por cada <app-counter>
})
export class CounterComponent { … }
```

**Por qué este no va en `root`.** Un cuaderno compartido entre dos cajas no es una
comodidad: es un error. La pregunta que decide es siempre la misma —*¿cuántos de
estos tiene que haber?*— y aquí la respuesta es «uno por mostrador».

**Y cómo se comprueba:** anotar en el A y mirar el B. Con un solo mostrador en
pantalla, los dos servicios se comportan igual y no se prueba nada.

## Paso 3 · El token

```ts
export const SHOP_NAME = new InjectionToken<string>('Nombre del café', {
  providedIn: 'root',
  factory: () => 'Café Compilado',
});
```

**Por qué un token y no `export const SHOP_NAME = 'Café Compilado'`.** La
constante se importa, y quien la importa queda atado. El token se reemplaza en un
solo lugar, sin tocar ni un componente — y eso es lo que hace que se pueda probar
con otro valor.

---

# Parte B · Los stores del hipódromo

**Referencia terminada:** `project/frontend/solution/src/app/core/`

## Paso 1 · `RaceStore`

Los `computed` de S3 se copian tal cual. Lo que cambia:

```ts
private readonly _filter = signal<RaceFilter>('all');
readonly filter = this._filter.asReadonly();

setFilter(filter: RaceFilter): void {
  this._filter.set(filter);
}
```

**Por qué el programa entero sigue sin ser un signal.** Viene de una constante y
no cambia nunca. En S7 lo va a traer `HttpClient` y **ahí sí**: va a cambiar esa
línea, y ninguna otra, porque todo lo demás ya se deriva.

**Por qué la carrera abierta se sigue derivando del id.** Es la decisión de S3, y
se muda intacta. Mover el estado a un servicio no cambia cómo se piensa: cambia
quién es el dueño.

## Paso 2 · `BetStore` inyecta a `RaceStore`

```ts
@Injectable({ providedIn: 'root' })
export class BetStore {
  private readonly races = inject(RaceStore);

  readonly target = computed(() => this.races.selected()?.favourite);
  readonly payout = computed(() => potentialPayout(this._amount(), this.target()?.odds ?? 0));
}
```

**Por qué el pago no se calcula en la pantalla.** Porque depende de dos cosas que
viven en lugares distintos, y en cuanto haya una segunda pantalla que muestre un
pago —la de la carrera en vivo, en S10— habría que repetir la fórmula.

**Y por qué los límites salen de `core/models`** y no están escritos aquí: son del
contrato. `MIN_BET_AMOUNT` y `MAX_BET_AMOUNT` ya existían desde S0.

## Paso 3 · Lo que se queda en la pantalla

```ts
const STATUS_LABELS: Record<RaceStatus, string> = { … };
const STATUS_TONES: Record<RaceStatus, BadgeTone> = { … };
const TIME_FORMAT = new Intl.DateTimeFormat('es', { … });
```

**Este es el corte, y es el que se discute en el code review.** Otra vista de las
mismas carreras podría querer «Finalizada» en vez de «Terminada», o un icono en
vez de una palabra, o la hora en formato relativo. Eso es presentación y es de la
pantalla.

El filtro y la carrera abierta, no: cualquier vista que muestre carreras necesita
exactamente los mismos.

## Paso 4 · El monto y `[(ngModel)]`

```ts
protected get amount(): number {
  return this.bets.amount();
}

protected set amount(value: number) {
  this.bets.setAmount(value);
}
```

**Por qué no se expone el signal de escritura para que `ngModel` lo use.** Porque
sería abrir el store entero para resolver un detalle de un formulario. El
componente hace de traductor, que es su trabajo.

---

## Los cinco errores que aparecen todos los años

| Lo que hacen | Por qué no alcanza | Qué decirles |
|---|---|---|
| `providedIn: 'root'` en el cuaderno | Los dos mostradores comparten anotaciones | «Anota en el A y mira el B.» |
| Probar con **un** mostrador | Con una copia, los dos servicios se ven iguales | «Pon el segundo antes de decir que anda.» |
| Exponen el signal de escritura | El store deja de ser dueño de nada | «¿Quién más puede vaciar la comanda?» |
| `inject()` adentro de un método | `NG0203` en tiempo de ejecución | «Léelo: te dice dónde sí se puede.» |
| Se llevan `STATUS_LABELS` al store | Otra vista querría otras etiquetas | «¿Y si mañana esto es un icono?» |

---

## Cómo se verifica que quedó bien

```bash
cd lab/starter              && npx tsc --noEmit && npm test
cd project/frontend/starter && npx tsc --noEmit && npm run build
node scripts/verify.mjs --fast
```

Y a ojo:

- Tomar un pedido en un mostrador lo muestra **en el otro**.
- Anotar en un mostrador **no** se ve en el otro.
- `grep -n "signal(" race-list.component.ts` no devuelve **ninguna** línea de estado.
- `STATUS_LABELS` **sí** sigue en `race-list.component.ts`.
