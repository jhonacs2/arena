# S6 · Reactividad — guión

> **Esto es un teleprompter, no un resumen.** Lo que está entre comillas se dice. Lo que está en gris se hace. Léelo de corrido antes de dar la clase, con cronómetro.

| | |
|---|---|
| **Concepto único** | Un signal guarda un valor. Un observable describe **algo que pasa a lo largo del tiempo**. No compiten: `toSignal()` es el puente. |
| **Al final saben** | Distinguir promesa de observable · encadenar `debounceTime`, `distinctUntilChanged` y `switchMap` · explicar por qué `switchMap` y no `mergeMap` · cortar una suscripción con `takeUntilDestroyed` · manejar los tres estados. |
| **Requisito previo** | S5 (servicios) y S3 (signals). |
| **Archivos** | `lab/starter/src/app/sessions/s06/` · `project/frontend/starter/src/app/features/races/` |

---

## Glosario de la sesión

| Palabra | En una frase |
|---|---|
| **Observable** | Una descripción de valores que van a llegar a lo largo del tiempo. |
| **Suscribirse** | Pedirle a un observable que empiece de verdad, y decirle qué hacer con cada valor. |
| **Frío** | Que no pasa nada hasta que alguien se suscribe. Cada suscripción empieza de cero. |
| **Caliente** | Que pasa igual, haya o no alguien escuchando. |
| **`Subject`** | Un observable al que se le puede empujar valores desde afuera. |
| **Operador** | Una función que transforma un flujo en otro. Van dentro de `pipe()`. |
| **`debounceTime`** | Espera a que dejen de llegar valores antes de dejar pasar el último. |
| **`distinctUntilChanged`** | Descarta un valor si es igual al anterior. |
| **`switchMap`** | Por cada valor arranca algo nuevo y **cancela lo anterior**. |
| **`mergeMap`** | Igual, pero **no cancela**: deja correr todo a la vez. |
| **`catchError`** | Atrapa el error y decide con qué seguir. |
| **`takeUntilDestroyed`** | Corta la suscripción cuando el componente se destruye. |
| **`toSignal`** | Convierte un observable en un signal. |
| **Los tres estados** | Cargando, vacío y error. Desde hoy son obligatorios. |

---

## 0:00 · Pregunta de apertura — 5 min

**En pantalla:** diapositiva 2.

> «Piensa en un buscador cualquiera, de los que usas todos los días. Escribes
> siete letras.»
>
> «**¿Cuántas veces le preguntó ese buscador al servidor?** Tira un número en el
> chat.»

**Espera 90 segundos.** Van a decir 1, 7, «una por letra», «no sé». **Todas
sirven.**

> «Los dos números que están diciendo son los dos extremos, y los dos son
> reales: hay buscadores que preguntan una vez y buscadores que preguntan
> siete.»
>
> «Hoy vamos a escribir el de siete —que es el que sale solo— y después vamos a
> convertirlo en el de una, con tres líneas.»

---

## 0:05 · Wayground de S5 — 7 min

**Correr:** `sesiones/s05-inyeccion-dependencias/wayground.csv`.

| Si falla | Decir |
|---|---|
| Declarado en los dos lados | «Dos instancias, y ningún error. Es el que más cuesta encontrar.» |
| `NG0203` | «`inject()` va arriba, en el campo de la clase.» |
| `asReadonly()` | «El store es el dueño; los métodos son el contrato.» |

---

## 0:12 · El concepto — 8 min

> **El editor está cerrado.**

### 0:12 — Lo que hasta hoy no existía · 2 min

**En pantalla:** diapositiva 5.

> «Hasta la clase pasada, los datos **estaban**. Una constante, un signal, y
> listo. Desde hoy los datos **llegan**, y con eso aparece lo único que todavía
> no tuvimos que manejar: **el tiempo**.»
>
> «Y el tiempo trae problemas que no existían. Uno: llegan tarde. Dos: llegan
> desordenados. Tres: a veces no llegan y llega un error. Cuatro: llegan cuando
> ya no le importan a nadie, porque el usuario se fue a otra pantalla.»

### 0:14 — Promesa y observable · 2 min

**En pantalla:** diapositiva 6.

**Los términos que se definen aquí:** *observable*, *suscribirse*, *frío*.

| | Promesa | Observable |
|---|---|---|
| Cuántos valores | **uno** | ninguno, uno o muchos |
| Cuándo empieza | al crearla | **al suscribirse** |
| Se puede cancelar | no | **sí** |

> «Una promesa es una caja que en algún momento va a tener una cosa adentro. Un
> **observable** es una descripción de valores que van a ir llegando.»
>
> «Y la diferencia que más se paga es la segunda: un observable es **frío**. No
> pasa nada hasta que alguien se **suscribe**. Escribir la búsqueda no busca
> nada; buscar es suscribirse.»

### 0:16 — El flujo, y los cuatro problemas · 4 min

**En pantalla:** diapositiva 7 — `diagramas/el-flujo.svg`.

Señala la tubería de arriba:

> «Esto se lee de arriba abajo como una frase: **escribe, espera, ¿cambió?,
> busca, guarda.** Cada paso es una línea, y cada línea resuelve exactamente uno
> de los problemas de abajo.»

Y la tabla de los cuatro:

| Problema | La línea |
|---|---|
| Una búsqueda por tecla | `debounceTime(300)` |
| La misma búsqueda dos veces | `distinctUntilChanged()` |
| Respuestas que llegan desordenadas | `switchMap()` |
| Suscripciones que nadie corta | `takeUntilDestroyed()` |

**Detente en el tercero, que es el que nadie ve venir:**

> «Escribes una letra: sale una búsqueda que va a tardar un segundo. Escribes
> tres más: sale otra que tarda trescientos milisegundos. **La segunda contesta
> primero, y después llega la primera y le pisa los resultados.**»
>
> «El usuario ve resultados que no corresponden a lo que escribió, y no hay
> ningún error. `switchMap` cancela la anterior cada vez que llega una nueva. Es
> lo que evita ese bug, y es la razón por la que casi siempre es `switchMap` y
> casi nunca `mergeMap`.»

**Y la frase que cierra el bloque:**

> «Signals y observables **no compiten**. Un signal guarda un valor; un
> observable maneja el tiempo. `toSignal()` es el puente, y lo vamos a escribir
> en dos minutos.»

> **Si vas tarde:** de este bloque no se recorta nada. La tabla de los cuatro es
> el índice del live coding.

---

## 0:20 · Live coding — 15 min

**Proyecto:** `lab/demo`, ruta `/s06`. La secuencia completa está en
**`mision-profe.md`**.

### 0:20 — El buscador ingenuo · 3 min

Está escrito, y funciona. Escribe `etiopía` **letra por letra** y señala el
contador.

> «Siete teclas, siete búsquedas. Y esto es lo que sale solo: es lo que uno
> escribe la primera vez, y no está mal escrito. Está mal **pensado en el
> tiempo**.»

Y ahora el bug que importa: escribe **una sola letra**, espera un segundo, y
escribe `huila`.

> «Miren la lista. Se llenó de resultados de la letra sola, **después** de haber
> mostrado los de huila. La respuesta vieja llegó última y ganó.»

### 0:23 — Un flujo en vez de una llamada · 3 min

```ts
private readonly terms = new Subject<string>();

protected onType(term: string): void {
  this.query = term;
  this.terms.next(term);
}
```

> «Un `Subject` es un observable **al que se le empuja**. Cada tecla ya no busca:
> deja caer un texto en el flujo.»

Y el flujo, todavía sin operadores:

```ts
private readonly results$ = this.terms.pipe(
  switchMap((term) => this.catalog.searchCounted(term)),
);

protected readonly results = toSignal(this.results$, { initialValue: [] });
```

> «`toSignal` es el puente: entra un observable, sale un signal. Y desde ahí para
> abajo el template es el mismo de la clase 3 — `results()`, con paréntesis.»

**Prueba el bug de las respuestas desordenadas otra vez.** Ya no pasa.

> «Con `switchMap`, cada texto nuevo **cancela** la búsqueda anterior. La vieja
> no llega, porque la cortamos.»

### 0:27 — `debounceTime` y `distinctUntilChanged` · 3 min

```ts
debounceTime(300),
distinctUntilChanged(),
```

Escribe `etiopía` letra por letra y señala el contador: **una búsqueda**.

> «`debounceTime` espera a que dejes de escribir. `distinctUntilChanged` descarta
> el texto si es igual al anterior — eso pasa más de lo que parece: borras una
> letra y la vuelves a escribir, y el texto final es el mismo.»
>
> «De siete a una, con dos líneas. Y no hay que acordarse de nada: está en el
> flujo.»

### 0:30 — Los tres estados · 3 min

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
> vacío y error. Es punto de la definición de terminado del curso, y arranca en
> esta clase.»

**Y el detalle que arruina buscadores enteros.** Escribe `error` y muestra el
mensaje. Después escribe `huila`: sigue funcionando.

> «Fíjense **dónde** está el `catchError`: adentro del `switchMap`, no afuera.»
>
> «Si estuviera afuera, el error mataría el flujo entero y el buscador dejaría de
> funcionar para siempre — hasta recargar. Adentro, solo muere esa búsqueda.»

### 0:33 — `takeUntilDestroyed` · 2 min

```ts
takeUntilDestroyed(),
```

> «Un flujo vive hasta que alguien lo corta. Si el usuario se va a otra pantalla,
> el componente se destruye y **la suscripción sigue viva**, apuntando a algo que
> ya no existe. Es una fuga de memoria, y con un temporizador de por medio es una
> que además sigue trabajando.»
>
> «Esta línea la corta sola. Y funciona porque estamos en un campo de la clase,
> que es un contexto de inyección — lo de la clase pasada, otra vez.»

---

## 0:35 · Misión 1 — 15 min

**Enunciado en `mision-estudiante-1.md`.**

> «El buscador funciona y hace una búsqueda por tecla. Hay que convertirlo en un
> flujo y resolver los cuatro problemas.»

**Dilo antes de largar:**

> «Empieza por el `Subject` y el `switchMap`, sin los otros operadores: que la
> pantalla siga funcionando. Los tiempos después. Si pones los cinco de una y no
> anda, no vas a saber cuál es.»

**Reloj de pistas:**

| Min | Pista, sin resolver |
|---|---|
| 0:43 | «Si no aparece nada, revisa que alguien se haya suscrito. `toSignal` se suscribe; un `pipe` suelto, no.» |
| 0:47 | «Si el error rompe el buscador para siempre, mira dónde está el `catchError`.» |

---

## 0:50 · Comparten pantalla — 10 min

**Preguntas, no corriges.**

1. «¿Cuántas búsquedas salen si escribo siete letras? ¿Cómo lo sabes?»
2. «¿Qué pasa si el servidor tarda dos segundos y escribo otra cosa mientras?»
3. «¿Quién corta esa suscripción?»
4. «¿Qué ve el usuario mientras espera?»

**Lo más probable:** el `catchError` quedó afuera del `switchMap`.

> «Escribe `error` y después escribe otra cosa. ¿Funciona? Ese es el bug, y es
> silencioso: la primera vez que un usuario tenga un problema de red, el buscador
> se le muere hasta que recargue.»

---

## 1:00 · Descanso — 10 min

---

## 1:10 · Predice y ejecuta — 15 min

**Respuestas verificadas con `fakeAsync`:** `predice-y-ejecuta/respuestas.md`.

| Min | Snippet | Casi todos predicen | Pasa |
|---|---|---|---|
| 1:10 | `mergeMap` en vez de `switchMap` | «igual, tal vez más lento» | **Gana la respuesta vieja** y pisa a la nueva |
| 1:15 | `catchError` afuera del `switchMap` | «muestra el error y sigue» | **El buscador muere para siempre** |
| 1:20 | Un observable que nadie usa | «se ejecuta igual» | **No pasa nada.** Es frío |

Cierra con:

> «Los tres son el mismo error visto de tres formas: **creer que el tiempo no
> existe**. Y los tres pasan el build, pasan la revisión y solo se ven con red
> lenta — que es exactamente la red que tienen los usuarios y no tenemos
> nosotros.»

---

## 1:25 · Misión 2, en parejas — 20 min

**Enunciado en `mision-estudiante-2.md`.**

**Tres cosas antes de largar:**

> «Uno: hoy el buscador del hipódromo filtra una constante, así que el debounce
> "no se nota". Póngalo igual: en la clase 7 cada búsqueda va a ser una petición,
> y entonces son quince en vez de una.»
>
> «Dos: el campo pasa de `[(ngModel)]` a `(input)`. Lo que hace falta es el
> **evento**, no el valor.»
>
> «Tres: el texto que se ve en el campo y el texto con el que se filtra dejan de
> ser el mismo. Piensen bien cuál va en cada lado.»

---

## 1:45 · Code review en vivo — 10 min

Rúbrica del curso, más las preguntas de hoy:

> «¿Dónde está el `catchError`? ¿Y qué pasa si estuviera un renglón más arriba?»
>
> «¿Quién corta esa suscripción? Si la respuesta es "nadie", es una fuga.»

Y el cierre:

> «Y ahora la buena noticia. Todo lo que escribieron hoy **no cambia en la clase
> 7**: cuando `search()` deje de ser un `of()` con delay y pase a ser un
> `HttpClient`, el componente no se entera. Devuelve un observable, y eso es lo
> único que a este flujo le importa.»

---

## 1:55 · Exit ticket y tarea — 5 min

**Exit ticket:** `exit-ticket.md`. **Tarea:** `tarea.md`, leída en voz alta.

**Y el aviso de la próxima:**

> «La clase que viene conectamos el backend de verdad. Van a ver que casi no hay
> que tocar nada — y esa es toda la razón por la que hicimos las seis clases
> anteriores en este orden.»

---

## Después de la clase

- [ ] Leer los exit tickets. Lo confuso arranca S7.
- [ ] Revisar `wayground.csv` de **esta** sesión — se corre al empezar S7.
- [ ] Aplicar la corrección de S6 al `starter/` publicado y taggear `s07`.

### Notas de la corrida real

| | |
|---|---|
| ¿Cuántos dejaron el `catchError` afuera? | |
| ¿Se entendió `switchMap` sin el diagrama? | |
| ¿Qué pregunta no supe contestar? | |
| ¿Qué sacaría o agregaría? | |
