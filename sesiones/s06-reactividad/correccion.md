# S6 · Corrección — de una llamada a un flujo

**Bloque 1:45 · instructor y alumno**

---

# Parte A · El buscador del lab

**Referencia terminada:** `lab/solution/src/app/sessions/s06/`

## Paso 1 · Un `Subject` en vez de una llamada

```ts
private readonly terms = new Subject<string>();

protected onType(term: string): void {
  this.query = term;
  this.terms.next(term);
}
```

**Por qué el campo usa `(input)` y no `[(ngModel)]`.** Lo que hace falta es el
**evento**, no el valor: hay que poder empujar cada pulsación al flujo. Con
`[(ngModel)]` el valor se guarda y nadie se entera de cuándo cambió.

## Paso 2 · El flujo, en el orden en que se lee

```ts
private readonly results$ = this.terms.pipe(
  debounceTime(300),
  distinctUntilChanged(),
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
  takeUntilDestroyed(),
);
```

**Por qué en ese orden.** Se lee como una frase: *espera, ¿cambió?, avisa que
está cargando, busca, guarda*. Y el orden importa de verdad en un punto: el
`debounceTime` va **antes** del `switchMap`, porque lo que hay que evitar es que
la búsqueda **salga**, no cancelarla después de haber salido.

**Por qué el `tap` de `loading` está afuera y el de `idle` adentro.** El «estoy
buscando» se enciende cuando se decide buscar; el «ya está» se apaga cuando
contesta esa búsqueda en particular.

**Por qué `catchError` va adentro.** Es el segundo «predice y ejecuta», y se
midió: con el `catchError` afuera, después de un error el flujo emitió **una** vez
y no volvió a emitir nunca. Un error en un observable es terminal — mata el flujo
donde esté. Adentro, solo muere esa búsqueda.

**Por qué `takeUntilDestroyed` va al final y sin argumentos.** Al final para que
alcance a todo lo de arriba; sin argumentos porque se llama en un campo de la
clase, que es un contexto de inyección. Es lo de S5, otra vez.

## Paso 3 · El puente

```ts
protected readonly results = toSignal(this.results$, {
  initialValue: [] as readonly Coffee[],
});
```

**Por qué `initialValue`.** Sin él, el tipo es `readonly Coffee[] | undefined`,
porque hasta que el observable emita algo el signal no tiene nada. Con él, el
template no necesita preguntar.

**Y una cosa que conviene decir en voz alta:** `toSignal` **se suscribe**. Si en
vez de esto se guardara `results$` en una propiedad y no se hiciera nada más, no
pasaría absolutamente nada — es el tercer «predice y ejecuta».

---

# Parte B · El buscador del hipódromo

**Referencia terminada:** `project/frontend/solution/src/app/features/races/race-list.component.ts`

## Paso 1 · Dos textos, no uno

```ts
private readonly typed = new Subject<string>();

/** Lo que se ve en el campo. El store recibe el valor más tarde. */
protected readonly draft = signal('');

protected onSearch(value: string): void {
  this.draft.set(value);
  this.typed.next(value);
}
```

**Este es el paso que se piensa mal la primera vez.** El texto que se ve y el
texto que filtra dejaron de ser el mismo:

| Qué | Quién lo tiene | Cuándo cambia |
|---|---|---|
| lo que se ve en el campo | `draft`, en el componente | en cada tecla |
| lo que filtra la lista | `query`, en el store | 250 ms después de la última |

Si el campo leyera del store, la letra recién escrita desaparecería del input
mientras el debounce espera. Se puede probar: es la Pista 1 del enunciado.

## Paso 2 · La suscripción, en el constructor

```ts
constructor() {
  this.typed
    .pipe(debounceTime(250), distinctUntilChanged(), takeUntilDestroyed())
    .subscribe((value) => this.races.setQuery(value));
}
```

**Por qué en el constructor y no en `ngOnInit`.** Porque `takeUntilDestroyed()`
sin argumentos necesita un contexto de inyección, y `ngOnInit` no lo es. En
`ngOnInit` habría que pasarle un `DestroyRef` a mano.

**Por qué `RaceStore` no cambia.** El debounce es una decisión de **esta
pantalla**: cuánto esperar antes de filtrar es cuestión de interfaz. Otra vista
podría querer filtrar al instante, o al apretar Enter. El store solo sabe filtrar.

## Paso 3 · «Ver todas» limpia las dos cosas

```ts
protected clearSearch(): void {
  this.draft.set('');
  this.typed.next('');
  this.races.clearFilters();
}
```

Si solo limpia el store, el campo queda con texto y la lista muestra todo. Si solo
limpia el campo, la lista sigue filtrada. Es el síntoma de haber separado los dos
textos y no haberse acordado.

---

## Los cinco errores que aparecen todos los años

| Lo que hacen | Por qué no alcanza | Qué decirles |
|---|---|---|
| `catchError` afuera del `switchMap` | El buscador muere tras el primer error | «Escribe `error` y después otra cosa.» |
| `mergeMap` en vez de `switchMap` | Gana la respuesta vieja | «Una letra, espera, y escribe otra cosa.» |
| El `pipe` guardado y nadie suscrito | No pasa nada, y no hay error | «¿Quién se suscribió?» |
| El campo lee del store | La letra desaparece mientras se escribe | «Escribe despacio y mira el campo.» |
| Guardan la suscripción y la cortan en `ngOnDestroy` | Funciona, y hay una línea | «Está bien. Ahora hazlo con `takeUntilDestroyed`.» |

> El último **no es un error**: es lo que se hacía antes de que existiera, y hay
> muchísimo código así. Conviene tratarlo como lo que es — una versión más larga
> de lo mismo.

---

## Cómo se verifica que quedó bien

```bash
cd lab/starter              && npx tsc --noEmit && npm test
cd project/frontend/starter && npx tsc --noEmit && npm run build
node scripts/verify.mjs --fast
```

Y a ojo, que es lo que el verificador no puede hacer:

- Escribir siete letras deja el contador en **1**.
- Una letra, esperar, y otra búsqueda: gana la segunda y **no vuelve a cambiar**.
- Después de escribir `error`, la siguiente búsqueda funciona.
- Escribir rápido en el hipódromo **no** hace parpadear la lista en cada tecla.
