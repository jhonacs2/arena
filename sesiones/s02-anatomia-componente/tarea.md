# S2 · Tarea asíncrona

**Entrega antes de S3.** Se lee en voz alta en clase antes de cortar.

---

## Qué hacer

Terminar la **Misión 2** si te quedó a medias, y después usar la tarjeta en un
segundo lugar. Ese es el examen de verdad de lo de hoy: **si el corte quedó bien,
esto lleva diez minutos y no hay que tocar el componente.**

### 1 · Una carrera destacada en la portada

Arriba del listado, mostrá **solo la carrera en vivo** —o la próxima por largar,
si no hay ninguna en vivo— usando el mismo `<app-race-card>`.

- Con una `<app-badge>` de tono `accent` que diga `DESTACADA`.
- Sin nada proyectado en el hueco de abajo.
- Sin que abrirla o cerrarla afecte al listado.

**No se toca `race-card.component.ts`.** Si tenés que tocarlo, anotá qué te faltó:
eso es el material del bloque de las 0:05 de la próxima.

### 2 · Un tercer tono para la pastilla

Agregá el tono `success` a `<app-badge>` y usalo para las carreras terminadas.

Fijate cuántos archivos tuviste que tocar. Si fueron más de dos, la pastilla
quedó sabiendo cosas que no le corresponden.

### 3 · Contestá por escrito

En un comentario al final de `race-list.component.ts`, dos preguntas:

1. `selected` es un `input()` de la tarjeta. ¿Qué se rompería si en vez de eso
   cada tarjeta se guardara **por su cuenta** si está abierta? Describí el
   síntoma exacto que vería el usuario.
2. `time` llega ya formateado. ¿Qué habría que cambiar, y en qué archivos, para
   que el listado mostrara «en 8 min» en vez de «30 jul, 14:23»? ¿Y si el
   formateo estuviera adentro de la tarjeta?

## Listo cuando

- [ ] La carrera destacada aparece arriba y usa el mismo componente
- [ ] `race-card.component.ts` **no se tocó**
- [ ] El tono `success` funciona en las carreras terminadas
- [ ] Las dos preguntas están contestadas
- [ ] `npm run build` pasa
- [ ] Commiteado con el prefijo de la sesión: `feat(s02): carrera destacada`

## Cuánto lleva

**30–45 minutos.** Si te lleva más de una hora, parás y anotás dónde te trabaste.

## Pistas

Para el punto 1, la carrera destacada sale de los mismos datos que ya preparaste:
buscá en `races` la que tenga `status === 'live'`, y si no hay, la primera
`upcoming`. Todo el cálculo va en la clase, no en el template.

Para el punto 3.1, probá romperlo de verdad antes de contestar: es más rápido
verlo que imaginarlo.

## Material de apoyo

- Angular · *Component communication*: <https://angular.dev/guide/components/inputs>
- Angular · *Content projection*: <https://angular.dev/guide/components/content-projection>
- Angular · *Lifecycle*: <https://angular.dev/guide/components/lifecycle>

> Ojo con la documentación oficial: está escrita para la última versión y a veces
> muestra APIs que en 18 no existen. Si algo no compila, la lista de lo prohibido
> está en el `CLAUDE.md` del repo.

---

## Para el instructor

No se corrige una por una. Se revisa una al azar en el code review de S3.

**Lo que más va a aparecer:**

- **Tocaron `race-card` para la destacada**, casi siempre agregándole un input
  `variant`. Es la señal de que el corte quedó atado a la pantalla, y es
  material de cinco minutos al empezar S3.
- **La respuesta a la 3.1 en abstracto** («se rompería el estado»). La respuesta
  que buscamos es el síntoma: *se abren dos a la vez y ninguna se cierra*.
- **El cálculo de la destacada hecho en el template.** No está mal, pero se
  recalcula en cada detección de cambios — y es la puerta de entrada perfecta a
  `computed()`, que es lo primero de S3.
