# S5 · Tarea asíncrona

**Entrega antes de S6.** Se lee en voz alta en clase antes de cortar.

---

## Qué hacer

Terminar la **Misión 2** si quedó a medias, y después tres cosas que solo salen
bien si el estado quedó donde corresponde.

### 1 · Testear el store sin ningún componente

Escribe `race.store.spec.ts`:

```ts
TestBed.configureTestingModule({});
const store = TestBed.inject(RaceStore);

store.setFilter('live');
expect(store.visible().length).toBe(1);

store.setQuery('payador');
expect(store.visible().length).toBe(1);

store.clearFilters();
expect(store.visible().length).toBe(8);
```

Al menos cinco casos, incluido este, que es el que más importa:

> Abrir una carrera, filtrarla hacia afuera, y comprobar que `selected()` queda
> en `undefined`.

Y después contesta, en un comentario: **¿por qué esto no se podía hacer cuando el
estado vivía en `race-list`?**

### 2 · Una segunda pantalla que use el mismo store

Agrega a `features/sistema` un panel «Estado del programa» que muestre:

- Los cuatro contadores de `RaceStore`.
- El nombre de la carrera abierta, o «ninguna».

**No recibe nada por `input()` y no es hijo del listado.** Después abre las dos
pantallas por turnos y comprueba que el filtro se conserva al ir y volver.

Eso último es lo que no se podía tener antes, y conviene que lo veas: el estado
ya no se destruye con el componente.

### 3 · Contesta una

En un comentario al final de `api.config.ts`:

> `API_URL` no lo usa nadie todavía. **¿Qué se gana con haberlo escrito ahora, y
> qué habría costado esperar a S7 y usar una constante importada mientras
> tanto?**

## Listo cuando

- [ ] `race.store.spec.ts` tiene al menos cinco casos y pasan
- [ ] El panel de «Estado del programa» funciona sin recibir nada
- [ ] El filtro se conserva al cambiar de pantalla y volver
- [ ] Las dos preguntas están contestadas
- [ ] `npm run build` y `npm test` pasan
- [ ] Commiteado: `feat(s05): stores de carreras y apuestas`

## Cuánto lleva

**30–45 minutos.**

## Material de apoyo

- Angular · *Dependency injection*: <https://angular.dev/guide/di>
- Angular · *Injection context*: <https://angular.dev/guide/di/dependency-injection-context>
- Angular · *Hierarchical injectors*: <https://angular.dev/guide/di/hierarchical-dependency-injection>

---

## Para el instructor

**Lo que más va a aparecer:**

- **El test del store creando el componente igual**, «por las dudas». Es la
  señal de que no se terminó de creer que el store existe solo. Vale mostrarlo
  al empezar S6: dos líneas, sin `fixture`, sin `detectChanges`.
- **La respuesta al 1 en abstracto** («porque estaba acoplado»). La que buscamos
  es concreta: había que **crear el componente, dibujarlo y buscar botones en el
  DOM** para probar un filtro que es tres líneas de lógica.
- **La respuesta al 3 subestimando el costo.** Cambiar de constante a token más
  adelante es fácil en un proyecto de cinco archivos; en S7 hay interceptores y
  servicios que la usarían, y cada uno queda atado a la importación.
