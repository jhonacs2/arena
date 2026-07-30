# S6 · Conceptos — Reactividad

> **Para qué es este archivo.** La clase es en vivo y no queda grabada. Cuando te
> sientes a hacer la tarea, esto es lo que tienes en vez de la memoria.

**Índice**

1. [El problema: aparece el tiempo](#1-el-problema-aparece-el-tiempo)
2. [Promesa y observable](#2-promesa-y-observable)
3. [Frío y caliente](#3-frío-y-caliente)
4. [`Subject`: un observable al que se le empuja](#4-subject-un-observable-al-que-se-le-empuja)
5. [Los operadores de hoy](#5-los-operadores-de-hoy)
6. [`switchMap` contra `mergeMap`](#6-switchmap-contra-mergemap)
7. [`catchError`, y dónde va](#7-catcherror-y-dónde-va)
8. [`takeUntilDestroyed` y `toSignal`](#8-takeuntildestroyed-y-tosignal)
9. [Los tres estados](#9-los-tres-estados)
10. [Glosario](#10-glosario)

---

## 1. El problema: aparece el tiempo

Hasta S5 los datos **estaban**: una constante, un signal, y listo. Desde hoy los
datos **llegan**, y eso trae cuatro problemas que no existían:

| Problema | La línea que lo resuelve |
|---|---|
| Una búsqueda por cada tecla | `debounceTime(300)` |
| La misma búsqueda dos veces | `distinctUntilChanged()` |
| Respuestas que llegan desordenadas | `switchMap()` |
| Suscripciones que nadie corta | `takeUntilDestroyed()` |

![El flujo](diagramas/el-flujo.svg)

---

## 2. Promesa y observable

| | Promesa | Observable |
|---|---|---|
| Cuántos valores | **uno** | ninguno, uno o muchos |
| Cuándo empieza | al crearla | **al suscribirse** |
| Se puede cancelar | no | **sí** |

> Una promesa es una caja que en algún momento va a tener una cosa adentro.
> Un **observable** es una descripción de valores que van a ir llegando.

---

## 3. Frío y caliente

> Un observable **frío** no hace nada hasta que alguien se **suscribe**, y cada
> suscripción empieza de cero.

```ts
const search$ = this.catalog.search('etiopía');   // no busca nada
search$.subscribe(…);                              // ahora sí
```

Fue el tercer «predice y ejecuta»: un observable que nadie usa **no se ejecuta**.
Es la trampa más común al empezar — se escribe la llamada, no aparece nada, y no
hay ningún error.

Un observable **caliente** pasa igual, haya o no alguien escuchando. Los eventos
del DOM y los sockets de S10 son calientes.

---

## 4. `Subject`: un observable al que se le empuja

```ts
private readonly terms = new Subject<string>();

protected onType(term: string): void {
  this.terms.next(term);
}
```

Cada tecla ya no busca: **deja caer un texto en el flujo**. Quién lo consume y
con qué reglas es problema del `pipe`.

---

## 5. Los operadores de hoy

Un **operador** es una función que transforma un flujo en otro. Van dentro de
`pipe()`, y se leen de arriba abajo como una frase.

```ts
this.terms.pipe(
  debounceTime(300),          // espera a que dejes de escribir
  distinctUntilChanged(),     // ¿cambió respecto al anterior?
  switchMap((t) => …),        // busca, cancelando lo anterior
  takeUntilDestroyed(),       // corta cuando el componente se va
);
```

| Operador | Qué hace |
|---|---|
| `debounceTime(ms)` | Deja pasar el último valor solo si pasaron `ms` sin que llegue otro |
| `distinctUntilChanged()` | Descarta el valor si es igual al anterior |
| `map(fn)` | Transforma cada valor |
| `filter(fn)` | Deja pasar solo los que cumplen |
| `tap(fn)` | Hace algo al pasar, sin cambiar el valor. Para efectos |
| `switchMap(fn)` | Por cada valor arranca algo nuevo y **cancela lo anterior** |
| `catchError(fn)` | Atrapa el error y decide con qué seguir |

**Lo que se midió en clase:** escribir `etiopía` letra por letra pasó de **siete
búsquedas a una** con `debounceTime` y `distinctUntilChanged`.

Y `distinctUntilChanged` parece que nunca hace falta, y hace falta todo el
tiempo: borrar una letra y volver a escribirla deja el mismo texto.

---

## 6. `switchMap` contra `mergeMap`

Este es el que hay que poder explicar.

```ts
switchMap((term) => this.catalog.search(term))   // cancela la anterior
mergeMap((term) => this.catalog.search(term))    // deja correr todas
```

**El caso que se corrió en clase:**

1. Escribes `e`. Sale una búsqueda que va a tardar 1200 ms.
2. Escribes `huila`. Sale otra que tarda 300 ms.
3. La de `huila` contesta primero y se muestran sus resultados.
4. **Con `mergeMap`**, novecientos milisegundos después llega la de `e` y **pisa
   la pantalla**. El usuario ve resultados que no corresponden a lo que escribió.
5. **Con `switchMap`**, la de `e` fue cancelada en el paso 2 y nunca llega.

> **La regla:** si solo importa el último, `switchMap`. Si importan todos —subir
> cinco archivos, por ejemplo—, `mergeMap`.

Y no hay ningún error en el caso malo. Solo hay resultados equivocados, y solo
con red lenta — que es exactamente la red que tienen los usuarios y no tenemos
nosotros.

---

## 7. `catchError`, y dónde va

```ts
switchMap((term) =>
  this.catalog.searchCounted(term).pipe(
    catchError(() => {
      this.status.set('error');
      return of([] as readonly Coffee[]);
    }),
  ),
),
```

Dos cosas, y las dos importan:

**1 · Devuelve un observable.** `catchError` tiene que contestar con algo con lo
que seguir. `of([])` dice «sigue con una lista vacía».

**2 · Va ADENTRO del `switchMap`.** Fue el segundo «predice y ejecuta»:

| Dónde | Qué pasa con un error |
|---|---|
| **adentro** del `switchMap` | muere esa búsqueda; la siguiente funciona |
| **afuera** | muere el flujo entero: **el buscador no responde nunca más**, hasta recargar |

Un error en un observable es terminal: el flujo se acaba. `catchError` adentro
solo alcanza a la búsqueda individual.

---

## 8. `takeUntilDestroyed` y `toSignal`

```ts
private readonly results$ = this.terms.pipe(…, takeUntilDestroyed());

protected readonly results = toSignal(this.results$, {
  initialValue: [] as readonly Coffee[],
});
```

> **`takeUntilDestroyed()`** corta la suscripción cuando el componente se
> destruye. Sin eso, el flujo queda vivo apuntando a algo que ya no existe.

Con un buscador es una fuga de memoria. Con un temporizador o un socket —que es
S10— es una fuga que además **sigue trabajando**.

Funciona sin pasarle nada porque se llama en un **campo de la clase**, que es un
contexto de inyección. Es lo de S5, otra vez.

> **`toSignal()`** es el puente: entra un observable, sale un signal. De ahí para
> abajo el template es el mismo de S3.

`initialValue` evita que el tipo sea `… | undefined`, que es lo que pasa mientras
el observable no emitió nada.

**Signals y observables no compiten:** un signal guarda un valor, un observable
maneja el tiempo.

---

## 9. Los tres estados

Desde esta sesión, **toda pantalla que cargue algo tiene tres**:

| Estado | Qué se ve |
|---|---|
| **Cargando** | un esqueleto, no un texto que salta |
| **Vacío** | una frase que dice qué pasó y qué hacer |
| **Error** | un mensaje, y la posibilidad de reintentar |

Es un punto de la definición de terminado del curso, y **arranca aquí**: hasta
S5 no aplicaba, porque no había nada que cargar ni nada que pudiera fallar.

---

## 10. Glosario

| Palabra | Qué es |
|---|---|
| **Observable** | Una descripción de valores que van a llegar con el tiempo |
| **Suscribirse** | Pedirle que empiece de verdad |
| **Frío** | No pasa nada hasta que alguien se suscribe |
| **Caliente** | Pasa igual, haya o no alguien escuchando |
| **`Subject`** | Un observable al que se le empuja desde afuera |
| **Operador** | Una función que transforma un flujo. Va en `pipe()` |
| **`debounceTime`** | Espera a que dejen de llegar valores |
| **`distinctUntilChanged`** | Descarta lo repetido |
| **`switchMap`** | Arranca algo nuevo y cancela lo anterior |
| **`mergeMap`** | Igual, pero sin cancelar |
| **`catchError`** | Atrapa el error y devuelve con qué seguir |
| **`takeUntilDestroyed`** | Corta al destruirse el componente |
| **`toSignal`** | Convierte un observable en un signal |

---

## Para la tarea

Lo que **no** vimos hoy: `combineLatest` y `forkJoin` —para cuando hacen falta
dos flujos a la vez—, `retry`, y `shareReplay`, que aparece cuando varios
componentes quieren la misma respuesta sin pedirla dos veces.

Y lo más importante que **no cambia**: cuando en S7 la búsqueda pase de ser un
`of()` con retardo a ser `HttpClient`, todo lo de este archivo sigue igual.
Devuelve un observable, y eso es lo único que al flujo le importa.
