# S5 · Misión profe — El live coding del bloque 0:20

**Solo instructor · 15 minutos · `lab/demo`**

---

## Antes de entrar al aula

```bash
node scripts/prep-demo.mjs
cd lab/demo
npm start
```

Suma la ruta `/s05` a mano, más `available: true` en `sessions.ts`.

**En pantalla:** VS Code y el navegador en <http://localhost:4200/s05>. Se ve un
mostrador con la comanda adentro del componente, funcionando.

---

## 0:20 — El servicio compartido · 4 min

```bash
ng generate service sessions/s05/order
```

**Mueve la comanda tal cual está.** No es una reescritura: es cortar y pegar.

```ts
@Injectable({ providedIn: 'root' })
export class OrderService {
  private readonly _orders = signal<readonly Order[]>([]);
  private nextId = 1;

  readonly orders = this._orders.asReadonly();
  readonly count = computed(() => this.orders().length);

  add(customer: string, coffee: string): void {
    const name = customer.trim();
    if (name === '') return;
    this._orders.update((orders) => [...orders, { id: this.nextId++, customer: name, coffee }]);
  }
}
```

> «Fíjate que el estado se escribe **igual** que la clase pasada: un `signal`
> privado y un `computed`. **Un store no es un patrón nuevo: es lo de S3 con
> `providedIn: 'root'` encima.**»

Y detente en las dos líneas que sí son nuevas:

```ts
private readonly _orders = signal(…);
readonly orders = this._orders.asReadonly();
```

> «El signal se escribe adentro; afuera se expone de solo lectura. Nadie puede
> llamar a `set` desde un componente. Para cambiar la comanda hay que pasar por
> `add()`, y **los métodos son el contrato del servicio**.»

En el componente:

```ts
protected readonly orders = inject(OrderService);
```

> «`inject()` pide la dependencia. Se llama en el **campo de la clase**, y eso
> importa: es un contexto de inyección, el momento en que Angular está
> construyendo esto y sabe a quién preguntarle. En un rato vamos a ver qué pasa
> si se llama en otro lado.»

**La pantalla queda igual.** Dilo:

> «No ganamos nada todavía. Lo que ganamos aparece en treinta segundos.»

---

## 0:24 — Dos mostradores · 3 min

Extrae el mostrador a `<app-counter>` —es el ejercicio de S2, va rápido— y
**ponlo dos veces**:

```html
<app-counter label="Mostrador A" />
<app-counter label="Mostrador B" />
```

**Toma un pedido en el mostrador A** y espera un segundo antes de hablar.

> «Apareció en el B. Y en el tablero de abajo.»
>
> «No hay ni un `input()` ni un `output()` entre los dos. **No se conocen.**
> Comparten la comanda porque los dos le pidieron lo mismo al mismo inyector, y
> el inyector les dio **el mismo objeto**. No se copiaron los datos: es el mismo.»

Y la conexión con S2:

> «Esto es lo que en la clase 2 no se podía hacer. `input()` y `output()`
> conectan padre e hijo. Un servicio conecta a cualquiera con cualquiera.»

---

## 0:27 — El servicio por componente · 3 min

```bash
ng generate service sessions/s05/notepad
```

**Quítale el `providedIn`** que puso el CLI, y decláralo en el componente:

```ts
@Injectable()
export class NotepadService { … }
```

```ts
@Component({
  …
  providers: [NotepadService],
})
export class CounterComponent { … }
```

Anota algo en el mostrador A.

> «El cuaderno del A no aparece en el B. Esa única línea —`providers` en el
> componente— es toda la diferencia.»

**Ahora quítala**, sin poner nada en su lugar:

> 🔴 **Rotura deliberada 1.**

```
R3InjectorError[NotepadService -> NotepadService]: NullInjectorError:
No provider for NotepadService!
```

> «Léelo de derecha a izquierda: *no hay proveedor para `NotepadService`*. El
> servicio existe y compila. Lo que falta es que **alguien haya dicho dónde
> vive**.»

Y ahora la parte que importa: ponle `providedIn: 'root'` y anota otra vez.

> «Ahora los dos mostradores comparten el cuaderno.»

**Deja que lo miren dos segundos.**

> «Y esto **no es un error**. No hay mensaje, no hay advertencia, la aplicación
> funciona. Es una decisión de diseño equivocada, y la única forma de darse
> cuenta es haberse hecho la pregunta: *¿cuántos de estos tiene que haber?*»

Vuelve a `providers` en el componente.

---

## 0:31 — El token · 2 min

```ts
export const SHOP_NAME = new InjectionToken<string>('Nombre del café', {
  providedIn: 'root',
  factory: () => 'Café Compilado',
});
```

```ts
protected readonly shopName = inject(SHOP_NAME);
```

> «Un servicio se pide por su clase. Una cadena no se puede pedir así: habría un
> solo `string` en toda la aplicación. Un `InjectionToken` es una llave con
> nombre y con tipo, hecha para esto.»
>
> «`factory` es el valor por defecto: con eso funciona sin configurar nada.»

Muéstralo aunque no lo dejes puesto:

```ts
providers: [{ provide: SHOP_NAME, useValue: 'Otro café' }]
```

> «Y esta es la razón de que sea un token y no una constante exportada: una
> constante se importa, y quien la importa queda atado. Esto se reemplaza en un
> solo lugar, sin tocar ni un componente. En la clase 7, cambiar entre el
> servidor de verdad y el mock va a ser exactamente esta línea.»

---

## 0:33 — `inject()` donde no va · 2 min

En el componente, adentro de un método:

```ts
protected take(): void {
  const service = inject(OrderService);
  …
}
```

**Compila.** Toca el botón.

> 🔴 **Rotura deliberada 2.**

```
NG0203: inject() must be called from an injection context such as a constructor,
a factory function, a field initializer, or a function used with `runInInjectionContext`.
```

> «Y el mensaje es de los buenos: te dice **exactamente dónde sí se puede**. En
> el constructor, en una factory, en un campo de la clase —que es donde lo
> pusimos al principio— o adentro de `runInInjectionContext`.»
>
> «El motivo es simple: cuando corre un método, Angular ya terminó de construir
> el componente y no tiene forma de saber a qué inyector preguntarle. **`inject()`
> va arriba, siempre.**»

Vuelve a dejarlo en el campo.

---

## Orden de sacrificio

| | Qué se saca | Por qué se puede |
|---|---|---|
| 1.º | El token de **0:31** | Está entero en `conceptos.md` §5 y es requisito del enunciado |
| 2.º | El `NG0203` de **0:33** | Va a aparecer solo durante la Misión 1, garantizado |
| 3.º | La extracción a `<app-counter>` de **0:24** | Se puede dejar el segundo mostrador escrito de antemano |

**Lo que no se sacrifica nunca:** el momento de las **0:24** —tomar un pedido en
un mostrador y verlo en el otro— y la comparación de las **0:27**. Los dos juntos
son la sesión.

---

## Si algo sale mal

| Pasa | Qué hacer |
|---|---|
| `No provider for` | El servicio no tiene `providedIn` ni está en ningún `providers`. |
| `NG0203` sin buscarlo | `inject()` quedó adentro de un método o de un callback. Va en el campo. |
| Los dos cuadernos se comparten y no querías | `providedIn: 'root'` sigue puesto además del `providers`. Es el primer «predice y ejecuta». |
| Quedó todo hecho un desastre | `node scripts/prep-demo.mjs` y `demo/` vuelve a cero. |
