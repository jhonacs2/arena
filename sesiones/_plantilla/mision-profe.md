# SNN · Misión profe — El live coding del bloque 0:20

**Solo instructor · 15 minutos · `lab/demo`**

La secuencia exacta que se escribe en vivo, en orden, con lo que se dice en cada
paso y las roturas deliberadas marcadas. Está pensado para el segundo monitor: el
`guion.md` lleva la clase, esto lleva el teclado.

---

## Antes de entrar al aula

```bash
node scripts/prep-demo.mjs
cd lab/demo
npm start
```

`lab/demo` es una copia descartable de `lab/starter`, así que arranca en el mismo
estado en que están los alumnos.

> **No trabajes en `lab/solution`.** No hay que borrarle nada a la solución de
> referencia para dar la clase. Y si el bloque sale mal,
> `node scripts/prep-demo.mjs` devuelve el lienzo limpio en un segundo, así que se
> puede ensayar tres veces.

**En pantalla:** VS Code y el navegador lado a lado.

**Decilo textual antes de empezar:**

> «Cierren el editor. Los próximos quince minutos yo escribo y ustedes miran. No
> copien: van a hacer esto mismo después, y con las manos libres se entiende
> mejor. Si me equivoco, avisen.»

---

## 0:20 — … · N min

El código exacto, en el orden en que se escribe.

```ts
```

> «La frase literal que se dice mientras se escribe.»

## 0:2N — … · N min · **rotura deliberada**

Escribí primero el error, a propósito. Después el arreglo. El error concreto que
van a cometer en el ejercicio tiene que verse **acá primero**.

```
El mensaje de error literal, copiado de la terminal o del navegador.
```

> «Leélo conmigo: …»

---

## Al terminar el bloque

El resultado queda **en pantalla** durante el ejercicio 1: es la referencia visual
de lo que van a construir. No cierres el navegador.

## Qué hacer si te quedás sin tiempo

El orden de sacrificio, de lo primero que se recorta a lo último:

1. …
2. …
3. **Nunca recortes** la rotura deliberada del error más común. Verlo acá primero
   es la mitad del valor del bloque.
