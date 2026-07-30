# S5 · Predice y ejecuta — respuestas

> **Verificado con Angular 18.2.14**, comparando instancias en un test de Karma y
> leyendo los mensajes de error del navegador. Las tres verificaciones quedaron
> como tests permanentes en `lab/solution/src/app/sessions/s05/s05.component.spec.ts`.

**El orden no se saltea:** mostrar → predecir 60 segundos → ejecutar → explicar.

---

## 1 · Declarado en los dos lados

**Opción 3, y es la peor de las tres posibles: hay dos instancias, y no hay
ningún error.**

Lo que se ve al tomar un pedido en el mostrador A:

| | Contador |
|---|---|
| Mostrador A | **1** |
| Mostrador B | **0** |
| Tablero de abajo | **0** |

### Por qué

Cuando un componente pide algo, Angular busca primero en **su propio inyector**;
si no está, sube al padre, y así hasta la raíz. Eso es la **jerarquía de
inyectores**.

`providers: [OrderService]` en el componente pone una instancia en el inyector de
ese componente. Como está más cerca, gana. Y como cada `<app-counter>` tiene el
suyo, hay **tres comandas distintas en pantalla**: la del A, la del B y la de la
raíz, que es la que lee el tablero.

> «`providedIn: 'root'` no es una orden: es un valor por defecto. Cualquiera que
> declare el servicio más abajo lo tapa.»

### Y por eso es el peligroso

No hay error, no hay advertencia, la aplicación funciona. El síntoma aparece
semanas después, y llega descrito como *«se me perdieron los datos»* o *«a veces
no se guarda»*.

**La única defensa es haberse hecho la pregunta:**

> **¿Cuántos de estos tiene que haber?**

### Cómo se detecta

Si un servicio compartido «a veces no comparte», busca su nombre en todos los
`providers` del proyecto. Si aparece en alguno además de en su `@Injectable`,
ese es.

---

## 2 · `inject()` adentro de un método

**Opción 2: compila, y falla al tocar el botón.**

```
NG0203: inject() must be called from an injection context such as a constructor,
a factory function, a field initializer, or a function used with `runInInjectionContext`.
```

### Por qué

Un **contexto de inyección** es el momento en que Angular está construyendo algo
y sabe a qué inyector preguntarle. Cuando corre un método, ya terminó de
construir el componente: no hay a quién preguntar.

**Y el mensaje es de los buenos**, porque dice exactamente dónde sí se puede:

- en el constructor
- en una factory
- **en un campo de la clase** ← lo que hacemos siempre
- dentro de `runInInjectionContext`

> **`inject()` va arriba, siempre.**

### Por qué compila

TypeScript no tiene forma de saber desde dónde se va a llamar una función. Es un
error de tiempo de ejecución, y aparece recién cuando alguien toca el botón — con
lo cual puede pasar el build, pasar la revisión y romperse con un usuario
adelante.

---

## 3 · Un servicio que pide otro servicio

**Opción 3: sí, exactamente así.**

```ts
@Injectable({ providedIn: 'root' })
export class BetStore {
  private readonly races = inject(RaceStore);
}
```

### Por qué

**El inyector no distingue quién pregunta.** Un componente, una directiva, un
pipe y otro servicio piden igual: con `inject()`, en un campo de la clase.

Y el campo de la clase de un `@Injectable` también es un contexto de inyección,
por el mismo motivo que en un componente: Angular está construyendo el servicio
en ese momento.

En el hipódromo es exactamente lo que hace `BetStore` con `RaceStore`, y es lo
que permite que el pago se derive de dos cosas que viven en lugares distintos sin
que nadie las sincronice.

### La pregunta de yapa

> «¿Y si `OrderService` inyectara a `ReceiptService` al mismo tiempo?»

Es una **dependencia circular**: A necesita B para construirse y B necesita A.
Angular lo detecta y falla al construir el primero de los dos.

Vale la pena decir qué significa cuando aparece, más allá del mensaje:

> «Casi nunca es un problema de Angular. Es la señal de que los dos servicios
> deberían ser uno solo, o de que falta un tercero que los dos usen.»

---

## La pregunta de cierre del bloque

> «De los tres, ¿cuál les habría costado más encontrar?»

> «El segundo te explota en la cara y te dice dónde. El tercero era una duda
> razonable y la respuesta es que sí.»
>
> «**El primero es el peligroso**: funciona, no avisa, y el síntoma aparece
> semanas después. Es el mismo tipo de bug que el `output()` sin escuchar de la
> clase 2 y el `push` sobre un signal de la clase 3 — **los que no fallan**.»
