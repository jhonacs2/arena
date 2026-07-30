# S3 · Predice y ejecuta — respuestas

> **Verificado en el navegador con Angular 18.2.14**, contando elementos del DOM
> en un test de Karma. **Dos de las tres no son lo que uno diría**, y una de las
> dos contradice lo que estaba escrito en el currículo. Si cambiás un snippet,
> volvé a correrlo antes de dar la clase.

**El orden no se saltea:** mostrar → predecir 60 segundos → ejecutar → explicar.

---

## 1 · Un `push` sobre el array de un signal

**Opción 3, y es peor que la 1.**

| | Antes | Después del `push` |
|---|---|---|
| `{{ items().length }}` | 2 | **3** |
| `{{ total() }}` — un `computed` | 3 | **3** |

**La lista se actualiza. El `computed` no.** Y no se recupera nunca: probamos
tocando **otro** signal distinto para forzar un repintado, y el total siguió en 3.

### Por qué

`push` cambió lo que hay **adentro** del array. Pero el signal no guarda el
contenido: guarda **el array** — la misma dirección de memoria de siempre. Para
él no pasó nada, así que no avisó.

¿Y por qué la lista sí cambió? Porque `{{ items().length }}` se **relee** cada
vez que Angular revisa este componente, y el clic que disparó el `push` salió del
template de este mismo componente, así que Angular lo revisó.

El `computed` no se relee: está **memoizado** contra un aviso que nunca llegó.

> «Casi todos predicen que no se actualiza nada. La verdad es peor: **se
> actualiza la mitad**. Una pantalla donde todo está viejo se nota. Una donde la
> mitad está vieja se descubre cuando un cliente reclama.»

### La frase para cerrar

> «El signal no vigila el contenido. Vigila **si le pusiste otra cosa**.»

### Y por eso el `readonly` está desde la sesión 0

Para escribir este bug hubo que sacarle el `readonly` al tipo:

```
Property 'push' does not exist on type 'readonly Order[]'.
```

**Ese error es la clase entera en una línea.**

---

## 2 · Un `@for` sin `track`

**Opción 3: no compila.**

```
NG5002: @for loop must have a "track" expression
```

### Por qué

Es de las poquísimas cosas que Angular obliga a escribir, y la razón es que la
alternativa —elegir por vos— sería peor en silencio.

`track` es cómo Angular reconoce que la fila de Ana **sigue siendo la de Ana**
cuando la lista se filtra o se reordena:

| | Qué hace Angular |
|---|---|
| `track order.id` | la reconoce y la **mueve** de lugar |
| sin nada con qué reconocerla | la **destruye** y crea una nueva |

Y con la fila se va el foco del teclado, la posición del scroll, el texto a medio
escribir en un input y cualquier animación en curso.

### La segunda pregunta, si sobra un minuto

> «¿Y por qué no alcanza con `track $index`?»

Porque el índice no identifica **la fila**, identifica **la posición**. En cuanto
la lista se reordena, la posición 0 pasa a ser otra comanda y Angular cree que es
la misma de antes que cambió de contenido. Sirve solo si no hay id y la lista no
se reordena nunca.

---

## 3 · Un `computed` que ordena

**Opción 2: `raw()` muestra `1,2,3`.**

Aunque `numbers` se declaró `[3, 1, 2]` y **nadie llamó a `set` ni a `update` en
ningún lado**.

### Por qué

`sort()` ordena **en el lugar**: no devuelve una copia ordenada, devuelve **el
mismo array**, reordenado. El `computed` `sorted` le cambió el orden al array que
está adentro del signal.

Y `raw`, que solo hacía `join(',')` sobre ese mismo signal, empezó a ver los
números ordenados. Sin que nadie lo tocara.

> «Un valor que nadie cambió, cambió. Y el culpable está en otro archivo, en una
> línea que parecía de solo lectura.»

### El arreglo

```ts
readonly sorted = computed(() => [...this.numbers()].sort((a, b) => a - b).join(','));
```

Tres caracteres. `[...]` arma un array nuevo, y `sort` ordena ese.

### La familia entera

Los métodos de array que **modifican el original**, y que por eso nunca van
directo sobre un signal:

`push` · `pop` · `shift` · `unshift` · `splice` · `sort` · `reverse` · `fill`

Los que devuelven uno nuevo, y que sí: `map` · `filter` · `slice` · `concat` ·
`toSorted` · spread.

---

## La pregunta de cierre del bloque

> «¿Qué tienen en común los tres?»

> «Los tres son **modificar algo que otro estaba mirando**. En el primero, el
> array que miraba un `computed`. En el tercero, el array que miraba otro
> `computed`. Y el segundo es Angular frenándote antes de que modifiques una
> lista que el usuario estaba mirando.»
>
> «De los tres, el único que te avisa es ese: el que te frena el build. Los otros
> dos te dejan seguir.»
