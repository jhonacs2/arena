# S5 · Inyección de dependencias — guión

> **Esto es un teleprompter, no un resumen.** Lo que está entre comillas se dice. Lo que está en gris se hace. Léelo de corrido antes de dar la clase, con cronómetro.

| | |
|---|---|
| **Concepto único** | Un servicio es un objeto que alguien más crea y te presta. **Dónde se declara decide cuántos hay.** |
| **Al final saben** | Escribir un servicio con estado · pedirlo con `inject()` · decidir entre `providedIn: 'root'` y `providers` del componente · usar un `InjectionToken` para lo que no es una clase · reconocer `NG0203`. |
| **Requisito previo** | S3 (signals) y S2 (`input`/`output`). |
| **Archivos** | `lab/starter/src/app/sessions/s05/` · `project/frontend/starter/src/app/core/` |

---

## Glosario de la sesión

| Palabra | En una frase |
|---|---|
| **Servicio** | Una clase sin template, con lógica o estado, que otros piden. |
| **Inyección de dependencias** | Que un objeto no se construya a sí mismo lo que necesita, sino que se lo pidan a alguien. |
| **Inyector** | Quien guarda las instancias y las entrega cuando se las piden. |
| **`inject()`** | La forma de pedir una dependencia. |
| **Contexto de inyección** | El momento en que Angular está construyendo algo y sabe a qué inyector preguntar. |
| **`providedIn: 'root'`** | Una sola instancia para toda la aplicación. |
| **`providers` del componente** | Una instancia por cada copia de ese componente. |
| **Jerarquía de inyectores** | Que un componente busque primero en el suyo y, si no está, en el de arriba. |
| **`InjectionToken`** | Una llave con nombre y con tipo, para pedir algo que no es una clase. |
| **Store** | Un servicio que guarda estado. No es un patrón nuevo: es S3 con `providedIn: 'root'`. |
| **`asReadonly()`** | Expone un signal sin `set` ni `update`, para que solo el servicio lo cambie. |

---

## 0:00 · Pregunta de apertura — 5 min

**En pantalla:** diapositiva 2.

> «En la clase 2 vimos que los datos bajan por `input()` y los avisos suben por
> `output()`. Funciona perfecto **entre padre e hijo**.»
>
> «Ahora imagina dos mostradores, uno al lado del otro, que tienen que compartir
> la misma comanda. **No son padre e hijo: son hermanos.** ¿Cómo se hablan, con
> lo que sabes hoy?»

**Espera 90 segundos.**

Van a decir «los subo al padre», «uso una variable global», «no sé». **Todas
sirven.**

> «El que dijo subirlos al padre tiene razón, y es lo que se hace cuando no se
> conoce lo de hoy. El problema es que el padre termina guardando algo que no le
> importa, solo para que dos hijos se enteren. Y con tres niveles de por medio,
> eso se convierte en pasar un dato por cinco componentes que no lo usan.»
>
> «Hoy vamos a ver el lugar común.»

---

## 0:05 · Wayground de S4 — 7 min

**Correr:** `sesiones/s04-directivas-pipes/wayground.csv`.

| Si falla | Decir |
|---|---|
| `NG8004` | «El pipe existe; este componente no lo declaró. Tercera vez que aparece la misma regla.» |
| El pipe impuro con `OnPush` | «Cada vez que se revisa **ese** componente, no cada clic de la aplicación.» |
| `LOCALE_ID` | «Los pipes incorporados formatean en `en-US` hasta que alguien diga lo contrario.» |

---

## 0:12 · El concepto — 8 min

> **El editor está cerrado.**

### 0:12 — Qué es inyectar · 3 min

**En pantalla:** diapositiva 5.

**Los términos que se definen aquí:** *servicio*, *inyección de dependencias*,
*inyector*.

> «Un **servicio** es una clase sin template, con lógica o con estado, que otros
> usan.»
>
> «Hasta aquí, si un componente necesitara uno, lo lógico sería escribir
> `new OrderService()`. Y eso funciona… hasta que hay dos componentes. Cada uno
> tendría el suyo, con su propia comanda, y no habría forma de que vieran lo
> mismo.»

Y ahora la idea:

> «**Inyección de dependencias** quiere decir que un objeto no se construye lo
> que necesita: se lo pide a alguien. Ese alguien es el **inyector**, lo trae
> Angular, y su trabajo es guardar las instancias y entregarlas.»
>
> «La ventaja no es escribir menos. Es que **quien pide no decide cuántos hay**.
> Esa decisión se toma en otro lado, en una línea, y se puede cambiar sin tocar a
> nadie.»

### 0:15 — Dónde se declara decide cuántos hay · 3 min

**En pantalla:** diapositiva 6 — `diagramas/donde-vive-un-servicio.svg`.

Señala el panel izquierdo:

> «`providedIn: 'root'` quiere decir **una sola instancia para toda la
> aplicación**. Los dos mostradores piden `OrderService` y reciben exactamente
> el mismo objeto. Por eso comparten la comanda: no es que se copien los datos,
> **es el mismo objeto**.»

Señala el panel derecho:

> «Y si el servicio se declara en `providers` del componente, hay **uno por cada
> copia de ese componente**. Dos mostradores, dos cuadernos, y ninguno ve el del
> otro.»

**La pregunta que decide, y hay que dejarla en pantalla:**

> **¿Cuántos de estos tiene que haber?**

> «Una comanda: una. Un cuaderno de mostrador: uno por mostrador. La respuesta a
> esa pregunta es la línea que hay que escribir.»

### 0:18 — Lo que no es una clase · 2 min

**En pantalla:** diapositiva 7.

**El término que se define aquí:** *`InjectionToken`*.

> «Un servicio se pide por su clase: `inject(OrderService)`. ¿Y cómo se pide una
> URL? No se puede escribir `inject(string)`: habría un solo `string` para toda
> la aplicación y no querría decir nada.»
>
> «Un **`InjectionToken`** es una llave con nombre y con tipo, hecha a propósito
> para eso.»
>
> «Y la pregunta que van a hacer: *¿por qué no una constante exportada?* Porque
> una constante se importa, y quien la importa queda atado. Un token se
> reemplaza en un solo lugar. En la clase 7, cambiar entre el servidor de verdad
> y el mock va a ser exactamente esa línea.»

> **Si vas tarde:** se recorta el bloque de las 0:18. El diagrama, no.

---

## 0:20 · Live coding — 15 min

**Proyecto:** `lab/demo`, ruta `/s05`. La secuencia completa está en
**`mision-profe.md`**.

### 0:20 — El servicio compartido · 4 min

```bash
ng generate service sessions/s05/order
```

Mueve la comanda del componente al servicio, tal cual estaba:

```ts
@Injectable({ providedIn: 'root' })
export class OrderService {
  private readonly _orders = signal<readonly Order[]>([]);
  readonly orders = this._orders.asReadonly();
  readonly count = computed(() => this.orders().length);

  add(customer: string, coffee: string): void { … }
}
```

> «Fíjate que el estado se escribe igual que en la clase pasada: un `signal`
> privado y un `computed`. **Un store no es un patrón nuevo: es lo de S3 con
> `providedIn: 'root'` encima.**»
>
> «Y lo que sí es nuevo son estas dos líneas: el signal privado se escribe
> adentro, y afuera se expone `asReadonly()`. Nadie de afuera puede llamar a
> `set`. Para cambiar la comanda hay que pasar por `add()`, y los métodos son el
> contrato.»

En el componente:

```ts
protected readonly orders = inject(OrderService);
```

> «`inject()` pide la dependencia. Se llama en el campo de la clase, que es un
> **contexto de inyección**: el momento en que Angular está construyendo esto y
> sabe a quién preguntarle.»

### 0:24 — Dos mostradores · 3 min

Extrae el mostrador a `<app-counter>` y **ponlo dos veces**:

```html
<app-counter label="Mostrador A" />
<app-counter label="Mostrador B" />
```

**Toma un pedido en el A.** Aparece en el B.

> «No hay ni un `input()` ni un `output()` entre los dos. No se conocen. Comparten
> la comanda porque los dos le pidieron lo mismo al mismo inyector, y el inyector
> les dio el mismo objeto.»

### 0:27 — El servicio por componente · 3 min

```bash
ng generate service sessions/s05/notepad
```

**Sin `providedIn`**, y en el componente:

```ts
providers: [NotepadService],
```

Anota algo en el mostrador A.

> «El cuaderno del A no aparece en el B. Esa única línea —`providers` en el
> componente— es toda la diferencia.»

**Y ahora quítala**, para que se vea al revés:

> 🔴 **Rotura deliberada 1.** Sin `providers` y sin `providedIn`, Angular no
> sabe dónde encontrarlo, y falla al construir el componente:

```
R3InjectorError[NotepadService -> NotepadService]: NullInjectorError:
No provider for NotepadService!
```

> «Léelo de derecha a izquierda: *no hay proveedor para `NotepadService`*. El
> servicio existe y compila; lo que falta es que **alguien haya dicho dónde
> vive**.»

Ponle `providedIn: 'root'` en su lugar y anota otra vez:

> «Ahora los dos mostradores comparten el cuaderno. **No hay ningún error**: es
> una decisión de diseño, y está mal, y nada te avisa. La única forma de darse
> cuenta es haberse hecho la pregunta: *¿cuántos de estos tiene que haber?*»

Vuelve a `providers` en el componente.

### 0:31 — El token · 2 min

```ts
export const SHOP_NAME = new InjectionToken<string>('Nombre del café', {
  providedIn: 'root',
  factory: () => 'Café Compilado',
});
```

```ts
protected readonly shopName = inject(SHOP_NAME);
```

> «`factory` es el valor por defecto: con eso funciona sin configurar nada. Y
> quien quiera otro lo cambia en un solo lugar, sin tocar ni un componente.»

Muéstralo, aunque no lo dejes:

```ts
providers: [{ provide: SHOP_NAME, useValue: 'Otro café' }]
```

### 0:33 — `inject()` donde no va · 2 min

```ts
protected take(): void {
  const service = inject(OrderService);   // ← aquí no
  …
}
```

> 🔴 **Rotura deliberada 2.** Toca el botón:

```
NG0203: inject() must be called from an injection context such as a constructor,
a factory function, a field initializer, or a function used with runInInjectionContext.
```

> «Y el mensaje dice exactamente dónde se puede: en el constructor, en una
> factory, **en un campo de la clase** —que es donde lo pusimos— o adentro de
> `runInInjectionContext`.»
>
> «El motivo es simple: cuando corre un método, Angular ya terminó de construir
> el componente y no sabe a qué inyector preguntarle. `inject()` va arriba,
> siempre.»

---

## 0:35 · Misión 1 — 15 min

**Enunciado en `mision-estudiante-1.md`.**

> «El mostrador funciona y la comanda está encerrada adentro del componente. Hay
> que sacarla a un servicio compartido, sacar el cuaderno a uno por mostrador, y
> poner un segundo mostrador para comprobar que funciona.»

**Dilo antes de largar:**

> «Empieza por mover la comanda sin cambiar nada más: que la pantalla quede
> exactamente igual. El segundo mostrador al final — es la comprobación, no el
> ejercicio.»

**Reloj de pistas:**

| Min | Pista, sin resolver |
|---|---|
| 0:43 | «Si dice `No provider for`, el servicio existe pero nadie dijo dónde vive.» |
| 0:47 | «Si dice `NG0203`, `inject()` está adentro de un método. Va en el campo de la clase.» |

---

## 0:50 · Comparten pantalla — 10 min

**Preguntas, no corriges.**

1. «¿Cuántas instancias de este servicio hay ahora mismo en la pantalla?»
2. «¿Cómo lo sabes?»
3. «Si mañana hay un tercer mostrador, ¿qué tocas?»
4. «¿Puede alguien de afuera vaciar tu comanda sin pasar por un método?»

**Lo más probable:** alguien puso `providedIn: 'root'` en el cuaderno y no notó
que los dos mostradores comparten anotaciones — porque solo probó con uno.

> «Probar con una copia esconde exactamente el error que esta clase enseña.»

---

## 1:00 · Descanso — 10 min

---

## 1:10 · Predice y ejecuta — 15 min

**Respuestas verificadas:** `predice-y-ejecuta/respuestas.md`.

| Min | Snippet | Casi todos predicen | Pasa |
|---|---|---|---|
| 1:10 | `providedIn: 'root'` **y** `providers` en el componente | «gana uno de los dos» | **Hay dos instancias**, y no hay error |
| 1:15 | `inject()` adentro de un método | «anda» | **`NG0203`**, en tiempo de ejecución |
| 1:20 | Un servicio que inyecta a otro servicio | «no se puede» | **Se puede.** El inyector no distingue quién pregunta |

Cierra con:

> «El segundo te explota en la cara. El tercero era una duda razonable y la
> respuesta es que sí. **El primero es el peligroso**: funciona, no avisa, y el
> síntoma aparece semanas después como “se me perdieron los datos”.»

---

## 1:25 · Misión 2, en parejas — 20 min

**Enunciado en `mision-estudiante-2.md`.**

**Tres cosas antes de largar:**

> «Uno: el estado del listado se muda **tal cual** a un `RaceStore`. No es una
> reescritura: es un cambio de dueño. Los `computed` que ya escribieron en la
> clase 3 se copian y pegan.»
>
> «Dos: lo que **se queda** en la pantalla es lo de presentación — las etiquetas
> de los estados, los tonos y el formato de la hora. Otra vista de las mismas
> carreras podría querer otras.»
>
> «Tres: `BetStore` va a inyectar a `RaceStore`. Un servicio puede pedir
> servicios; es lo tercero del bloque de predicciones.»

---

## 1:45 · Code review en vivo — 10 min

Rúbrica del curso, más las dos preguntas de hoy:

> «¿Cuántas instancias de este servicio hay? ¿Y cómo lo sabrías sin ejecutarlo?»
>
> «¿Alguien de afuera del store puede escribir su estado? Si la respuesta es sí,
> el store no es el dueño de nada.»

Y el cierre:

> «Miren el componente ahora. En la clase 1 tenía todo. Hoy tiene las etiquetas,
> los tonos y el formato de la hora — **lo que de verdad es de esta pantalla**.
> Todo lo demás encontró su lugar, y cada mudanza fue por el mismo motivo:
> alguien más lo iba a necesitar.»

---

## 1:55 · Exit ticket y tarea — 5 min

**Exit ticket:** `exit-ticket.md`. **Tarea:** `tarea.md`, leída en voz alta.

**Y el aviso de la próxima:**

> «La clase que viene: reactividad y observables. Los datos van a dejar de estar
> ahí y van a empezar a llegar — y con eso aparece el tiempo, que es lo único que
> todavía no tuvimos que manejar.»

---

## Después de la clase

- [ ] Leer los exit tickets. Lo confuso arranca S6.
- [ ] Revisar `wayground.csv` de **esta** sesión — se corre al empezar S6.
- [ ] Aplicar la corrección de S5 al `starter/` publicado y taggear `s06`.

### Notas de la corrida real

| | |
|---|---|
| ¿Cuántos probaron con un solo mostrador? | |
| ¿Apareció `NG0203` sin que yo lo provocara? | |
| ¿Qué pregunta no supe contestar? | |
| ¿Qué sacaría o agregaría? | |
