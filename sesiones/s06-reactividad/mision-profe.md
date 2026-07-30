# S6 · Misión profe — El live coding del bloque 0:20

**Solo instructor · 15 minutos · `lab/demo`**

---

## Antes de entrar al aula

```bash
node scripts/prep-demo.mjs
cd lab/demo
npm start
```

Suma la ruta `/s06` a mano, más `available: true` en `sessions.ts`.

**En pantalla:** VS Code y el navegador en <http://localhost:4200/s06>. Se ve el
buscador ingenuo, funcionando.

**Y una recomendación concreta:** ten la pantalla del navegador con el contador
de búsquedas a la vista **todo el bloque**. Es el marcador de la clase.

---

## 0:20 — El buscador ingenuo · 3 min

**No escribas nada todavía.** Escribe `etiopía` en el campo, **letra por letra y
despacio**, y señala el contador subiendo.

> «Siete teclas, siete búsquedas.»
>
> «Y quiero que quede claro que esto **no está mal escrito**: es lo que sale
> solo, es lo que escribe cualquiera la primera vez, y hace exactamente lo que
> dice. Está mal **pensado en el tiempo**, que es otra cosa.»

Ahora el bug que importa. **Toca «Reiniciar»**, escribe **una sola letra** —`e`—,
espera un segundo entero, y escribe `huila`.

> 🔴 Deja que lo miren.

> «Apareció Huila… y después se llenó de resultados de la letra sola. La
> respuesta vieja llegó última y ganó.»
>
> «El usuario ve resultados que no corresponden a lo que escribió. Y no hay
> ningún error, ni en la consola ni en ningún lado.»

---

## 0:23 — Un flujo en vez de una llamada · 3 min

```ts
import { Subject, switchMap } from 'rxjs';
import { toSignal } from '@angular/core/rxjs-interop';

private readonly terms = new Subject<string>();

protected onType(term: string): void {
  this.query = term;
  this.terms.next(term);
}
```

> «Un `Subject` es un observable **al que se le empuja** desde afuera. Cada tecla
> ya no busca: deja caer un texto en el flujo.»

Y el flujo, **todavía sin los tiempos**:

```ts
private readonly results$ = this.terms.pipe(
  switchMap((term) => this.catalog.searchCounted(term)),
);

protected readonly results = toSignal(this.results$, { initialValue: [] });
```

Borra el `signal` de resultados que había y usa este.

> «`toSignal` es el puente: **entra un observable, sale un signal**. Y de ahí
> para abajo el template es el mismo de la clase 3 — `results()`, con
> paréntesis.»
>
> «`initialValue` está para que el tipo no sea `… | undefined`: sin eso, hasta
> que llegue el primer valor el signal no tiene nada.»

**Repite el bug de recién:** una letra, esperar, `huila`.

> «Ya no pasa. `switchMap` **cancela** la búsqueda anterior cada vez que llega un
> texto nuevo. La vieja no llega porque la cortamos.»

---

## 0:27 — `debounceTime` y `distinctUntilChanged` · 3 min

Arriba del `switchMap`:

```ts
debounceTime(300),
distinctUntilChanged(),
```

**Reinicia el contador y escribe `etiopía` letra por letra**, al ritmo normal.

> «Una. De siete a una, con dos líneas.»

Y explica cada una por separado, señalando:

> «`debounceTime` espera trescientos milisegundos **sin que llegue nada nuevo**.
> Mientras escribes, el reloj se reinicia con cada tecla.»
>
> «`distinctUntilChanged` descarta el texto si es igual al anterior. Parece que
> nunca pasa, y pasa todo el tiempo: borras una letra y la vuelves a escribir, o
> pegas el mismo texto. Sin él, esas dos son dos búsquedas.»

> «Y lo mejor es que no hay que acordarse de nada. No hay un `if` en el
> componente comprobando si el texto cambió: **está en el flujo**.»

---

## 0:30 — Los tres estados · 3 min

```ts
tap(() => this.status.set('loading')),
switchMap((term) =>
  this.catalog.searchCounted(term).pipe(
    tap(() => this.status.set('idle')),
    catchError(() => {
      this.status.set('error');
      return of([] as readonly Coffee[]);
    }),
  ),
),
```

> «Desde hoy, toda pantalla que cargue algo tiene **tres** estados: cargando,
> vacío y error. Es un punto de la definición de terminado del curso y arranca en
> esta clase — antes no aplicaba, porque no había nada que cargar.»

Escribe `error` y muestra el mensaje. **Después escribe `huila`.** Sigue
funcionando.

**Y ahora el detalle, que es lo que hay que llevarse del bloque.** Señala con el
cursor la posición del `catchError`:

> «Fíjense **dónde** está: adentro del `switchMap`, en la tubería de la búsqueda
> individual. No afuera.»
>
> «Si estuviera afuera, el error mataría el flujo entero: el buscador dejaría de
> funcionar **para siempre**, hasta recargar la página. Adentro, solo muere esa
> búsqueda y la siguiente arranca sana.»

Si hay tiempo, muéstralo: sácalo afuera, escribe `error`, y después intenta
buscar otra cosa. No responde más. Vuelve a ponerlo adentro.

---

## 0:33 — `takeUntilDestroyed` · 2 min

Al final del `pipe`:

```ts
takeUntilDestroyed(),
```

> «Un flujo vive hasta que alguien lo corta. Si el usuario se va a otra pantalla,
> el componente se destruye y **la suscripción sigue viva**, apuntando a algo que
> ya no existe.»
>
> «Con un buscador es una fuga de memoria. Con un temporizador o un socket —que
> es la clase 10— es una fuga que además **sigue trabajando**.»
>
> «Y funciona sin pasarle nada porque estamos en un campo de la clase, que es un
> contexto de inyección. Es lo de la clase pasada, otra vez.»

Si preguntan qué pasaba antes de que existiera:

> «Se guardaba la suscripción en una propiedad y se llamaba a `unsubscribe()` en
> `ngOnDestroy`. Van a ver mucho código así, y es correcto: esta línea hace lo
> mismo.»

---

## Orden de sacrificio

| | Qué se saca | Por qué se puede |
|---|---|---|
| 1.º | `takeUntilDestroyed` de **0:33** | Es una línea; se puede contar sin escribirla |
| 2.º | `distinctUntilChanged` de **0:27** | El contador ya bajó de siete a una con el debounce |
| 3.º | La demostración de sacar el `catchError` afuera | Vuelve a las 1:15 como ejercicio de predicción |

**Lo que no se sacrifica nunca:** el bug de las respuestas desordenadas de las
**0:20** y su desaparición a las **0:23**. Sin ese antes y después, `switchMap`
es una palabra.

---

## Si algo sale mal

| Pasa | Qué hacer |
|---|---|
| No aparece ningún resultado | Nadie se suscribió. `toSignal` se suscribe; un `pipe` guardado en una propiedad, no. |
| `NG0203` al usar `takeUntilDestroyed` | Está fuera de un contexto de inyección. Va en el campo de la clase, no en un método. |
| El buscador muere tras un error | El `catchError` quedó afuera del `switchMap`. |
| El bug de las respuestas desordenadas no se reproduce | Escribe **una sola letra** y espera de verdad: la latencia larga solo se dispara con textos de una letra. |
| Quedó todo hecho un desastre | `node scripts/prep-demo.mjs` y `demo/` vuelve a cero. |
