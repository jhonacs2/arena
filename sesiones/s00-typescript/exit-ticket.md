# S0 · Exit ticket

**3 minutos. Anónimo si preferís.**

---

**1. Recordar** — Un tipo es el conjunto de valores que algo puede tener. Escribí
el tipo que corresponde a cada caso:

| Qué representa | El tipo |
|---|---|
| El estado de una carrera: `upcoming`, `live` o `finished` | |
| El lugar en el podio: 1, 2 o 3 | |
| El caballo favorito, sabiendo que puede no haber ninguno | |

**2. Aplicar** — Este tipo tiene un problema. La lista de caballos se puede
modificar desde cualquier parte del programa. Escribí la línea corregida:

```ts
interface Race {
  readonly horses: Horse[];
}
```

**3. ¿Qué quedó confuso?**

*(Vale «nada». Vale «todo». Vale una sola palabra.)*

---

## Para el instructor

La pregunta 3 es la que importa. Lo que aparezca ahí **arranca S1**, en el bloque
de las 0:05, junto con el Wayground.

**Respuestas esperadas de la 1:**

| | |
|---|---|
| Estado | `'upcoming' \| 'live' \| 'finished'` |
| Podio | `1 \| 2 \| 3` |
| Favorito | `Horse \| undefined` |

**Respuesta esperada de la 2:** `readonly horses: readonly Horse[];`

| Lo que escriben | Qué decirles |
|---|---|
| `readonly horses: Horse[]` sin cambiar nada | Es lo que ya estaba. El `readonly` que falta es el de adentro, el del array. |
| `horses: readonly Horse[]` | Cierra la puerta correcta, pero deja abierta la otra: `race.horses = []` vuelve a compilar. Van los dos. |
| `Readonly<Horse[]>` | Funciona, y es una forma válida. Preguntarle si sabe qué hace con un objeto anidado — la respuesta es «nada», y es el mismo tema. |
| `const horses` | Confunde `const` con `readonly`. `const` es sobre la variable; `readonly`, sobre la propiedad. Vale la pena aclararlo en clase. |

| Señal | Qué hacer |
|---|---|
| Más de un tercio confundido con lo mismo | Abre S1, 5 minutos |
| Una sola persona trabada | Mensaje directo, no clase entera |
| Nadie confundido | La sesión fue fácil de más. Subir el techo de la Misión 2 la próxima |
