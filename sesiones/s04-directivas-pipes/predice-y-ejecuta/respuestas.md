# S4 · Predice y ejecuta — respuestas

> **Verificado con Angular 18.2.14.** Los dos mensajes de error están copiados de
> la salida del compilador, y el comportamiento del tercero se midió contando
> llamadas en un test de Karma. Si cambias un snippet, vuelve a correrlo antes de
> dar la clase.

**El orden no se saltea:** mostrar → predecir 60 segundos → ejecutar → explicar.

---

## 1 · Un pipe que existe y no se declara

**Opción 3: no compila.**

```
NG8004: No pipe found with name 'money'.
```

### Por qué

«No encontré ningún pipe que se llame `money`». El pipe existe, está exportado y
compila solo — pero **este componente** no lo declaró.

Es la tercera vez que aparece la misma regla, con tres caras distintas:

| Sesión | Qué faltaba en `imports` |
|---|---|
| S1 | `FormsModule`, para `[(ngModel)]` |
| S2 | el componente hijo |
| S4 | el pipe |

> **Si el template lo usa, se declara.** Vale para componentes, directivas y
> pipes, y no hay excepciones — salvo `@if` y `@for`, que no son directivas sino
> sintaxis del template.

### El error hermano, que cuesta más

Si el pipe **sí** está en `imports` y el error sigue apareciendo: el `name` del
`@Pipe` no coincide con lo que se escribió en el template.

```ts
@Pipe({ name: 'money' })    // esto
{{ x | Money }}             // no es esto
```

**Se comparan cadenas, no clases.** El nombre de la clase no importa.

---

## 2 · Una directiva sin `standalone`

**Opción 3: no compila.**

```
TS-992011: The directive 'HighlightDirective' appears in 'imports', but is not standalone
and cannot be imported directly. It must be imported via an NgModule.
```

### Por qué

En Angular 18, `standalone` **no es el valor por defecto**: hay que escribirlo. Una
directiva sin él es una directiva de NgModule, y un componente standalone no puede
importar eso directamente.

Es exactamente el mismo error que sale con un componente, y con un pipe. **La regla
es una sola para los tres.**

> En versiones posteriores a la 18, `standalone: true` pasó a ser el
> comportamiento por defecto y dejó de escribirse. En este curso se escribe
> siempre — está en `CLAUDE.md` §5 y el verificador lo revisa.

---

## 3 · Un pipe impuro en un componente `OnPush`

**Opción 3: solo cuando Angular revisa ese componente.**

Casi todos eligen la 2, y la diferencia entre la 2 y la 3 es la mitad del tema.

### Lo que se mide en la pantalla

Después de seis clics en el botón:

| | Llamadas |
|---|---|
| `pure: true` | **1** |
| `pure: false` | **7** |

El puro se llamó **una vez** y no volvió a correr: su valor de entrada es siempre
la misma cadena.

### Por qué la opción 2 está cerca y no es exacta

Un pipe impuro corre cada vez que Angular **revisa el componente en el que está**.
Cuántas veces es eso depende de la estrategia de detección de cambios:

| | Cuándo se revisa |
|---|---|
| `OnPush` | cuando pasa algo **adentro** del componente, o cambia un input, o avisa un signal |
| `Default` | prácticamente en cada evento de **cualquier parte** de la aplicación |

Con `Default`, la opción 2 sería la correcta.

**Se puede demostrar en vivo:** cambia `OnPush` por `Default` y el contador
empieza a subir con cualquier clic de la página.

### Y por eso este es el peor de los tres

> «Los dos primeros te los frena el compilador: se arreglan en diez segundos. El
> tercero **no falla nunca**. Funciona, y es lento.»
>
> «Y lo lento no se ve en el proyecto de la clase, con cuatro cafés. Se ve con
> cuatro mil filas y un usuario que dice que la aplicación se traba.»

### Cuándo sí va impuro

Cuando el resultado depende de algo que el valor de entrada no ve — el reloj es el
ejemplo clásico: un pipe de «hace 3 minutos» tiene una entrada que no cambia
nunca y una salida que cambia sola.

Y aun así, casi siempre hay una forma mejor: en este curso, un `computed` de S3.
