# S6 · Ejercicio 1 — De siete búsquedas a una

**Individual · 15 minutos · proyecto `lab/starter`**

---

## Enunciado

El buscador del catálogo funciona. Escribe `etiopía` letra por letra y mira el
contador: **siete búsquedas**. Escribe una sola letra, espera un segundo y
escribe otra cosa: los resultados viejos llegan tarde y pisan a los nuevos.

Ninguno de los dos es un error de escritura. Son problemas del **tiempo**, que es
lo que ninguna sesión anterior tuvo que manejar.

Convierte la búsqueda en un flujo y resuelve los cuatro.

## Estado inicial

```bash
cd lab/starter
npm start
```

La ruta `/s06` **no existe todavía**. El componente sí está, en
`src/app/sessions/s06/`, con los `TODO(S6)`.

`CatalogService` ya está escrito y **no se toca**. Devuelve observables con
retardo, y a propósito tarda más con textos de una letra.

---

## Requisitos

### 1. La pantalla es alcanzable

Declara la ruta `/s06` y hazla aparecer en la barra lateral.

### 2. Una búsqueda por persona, no por tecla

Escribir `etiopía` letra por letra tiene que producir **una** búsqueda.

- Un `Subject<string>` recibe cada texto.
- `debounceTime(300)` espera a que la persona deje de escribir.
- `distinctUntilChanged()` descarta el texto si es igual al anterior.

### 3. Gana siempre la última búsqueda

Con `switchMap`, cada texto nuevo cancela la búsqueda anterior.

**La comprobación:** escribe una sola letra, espera un segundo, escribe `huila`.
Tienen que quedar los resultados de `huila` y no volver a cambiar.

### 4. Los tres estados

| Estado | Qué se ve |
|---|---|
| Cargando | las tres barras grises |
| Vacío | `Ningún café coincide con la búsqueda.` |
| Error | el mensaje en rojo |

**Y el error no puede matar el buscador.** Después de escribir `error`, la
siguiente búsqueda tiene que funcionar.

### 5. La suscripción se corta sola

`takeUntilDestroyed()`, para que el flujo no siga vivo cuando el componente se
va de la pantalla.

### 6. El resultado es un signal

`toSignal()` convierte el flujo en un signal, y el template queda igual que en
S3: `results()`, con paréntesis.

---

## Resultado esperado

```
Buscar por nombre u origen  [ etiopía          ]

Búsquedas que salieron: 1        ← escribiendo las siete letras
Resultados: 2

  Yirgacheffe   Etiopía   4200
  Sidamo        Etiopía   4100
```

## Restricciones

- No se toca `catalog.service.ts`.
- Prohibido `mergeMap` para esto. Si no sabes por qué, está en el bloque de
  predicciones.
- Prohibido guardar la suscripción en una propiedad y cancelarla a mano: hay una
  línea para eso.
- Prohibido `any`. `OnPush`.

## Autoevaluación

- [ ] `/s06` abre y aparece en la barra lateral
- [ ] `npm test` pasa
- [ ] Escribir siete letras produce **una** búsqueda
- [ ] Escribir lo mismo dos veces **no** produce una segunda
- [ ] Una letra, esperar, y otra búsqueda: gana la segunda y no vuelve a cambiar
- [ ] Después de un error, la siguiente búsqueda funciona
- [ ] No quedó ningún `TODO(S6)`

---

## Pistas

<details>
<summary>Pista 1 — el orden que menos duele</summary>

Empieza por el `Subject` y el `switchMap`, **sin** `debounceTime` ni
`distinctUntilChanged`. Que la pantalla vuelva a funcionar.

Los tiempos después. Si pones los cinco operadores de una y no anda, no vas a
saber cuál es.
</details>

<details>
<summary>Pista 2 — no aparece nada</summary>

Un observable es **frío**: no pasa nada hasta que alguien se suscribe.

`toSignal()` se suscribe. Un `pipe()` guardado en una propiedad y nada más, no.
</details>

<details>
<summary>Pista 3 — dónde va el <code>catchError</code></summary>

```ts
switchMap((term) =>
  this.catalog.searchCounted(term).pipe(
    catchError(() => { … }),      // ← ADENTRO
  ),
),
```

Afuera, el error mata el flujo entero y el buscador deja de responder para
siempre. Adentro, solo muere esa búsqueda.
</details>

<details>
<summary>Pista 4 — <code>toSignal</code> y el tipo</summary>

```ts
protected readonly results = toSignal(this.results$, {
  initialValue: [] as readonly Coffee[],
});
```

Sin `initialValue`, el tipo es `readonly Coffee[] | undefined`, porque hasta que
llegue el primer valor el signal no tiene nada.
</details>

<details>
<summary>Solución completa</summary>

`correccion.md`, Parte A.
</details>

## Extensión

Muestra, junto al contador, **cuántas teclas se tocaron**. La relación entre los
dos números es la clase entera en una línea.

Y después contesta, en un comentario: si el debounce fuera de 2000 ms en vez de
300, ¿qué ganaría el servidor y qué perdería el usuario?

---

> **Referencia:** `conceptos.md` tiene todos los conceptos de la sesión con sus
> ejemplos.
