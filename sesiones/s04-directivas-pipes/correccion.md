# S4 · Corrección — el formato, en un solo lugar

**Bloque 1:45 · instructor y alumno**

---

# Parte A · La carta del lab

**Referencia terminada:** `lab/solution/src/app/sessions/s04/`

## Paso 0 · La ruta

`/s04` en `app.routes.ts` y `available: true` en `sessions.ts`.

## Paso 1 · Los pipes que ya vienen, y el idioma

```html
<p class="card__origin">{{ coffee.origin | uppercase }}</p>
<p class="card__stock">
  {{ coffee.stock | number }} en depósito · {{ share(coffee) | percent: '1.0-1' }} del total
</p>
```

Y en `app.config.ts`:

```ts
registerLocaleData(localeEs);
providers: [{ provide: LOCALE_ID, useValue: 'es' }, …]
```

**Por qué esto va primero.** Es la línea más barata de la sesión y la que más se
olvida: sin ella, cualquier pipe incorporado que se agregue después formatea en
inglés, y el síntoma —un punto donde va una coma— es fácil de no ver.

**Por qué `registerLocaleData` va afuera del objeto.** Es una llamada, no un
proveedor: se ejecuta cuando se carga el archivo. Ponerla adentro de `providers`
no compila.

## Paso 2 · `MoneyPipe`

```ts
@Pipe({ name: 'money', standalone: true, pure: true })
export class MoneyPipe implements PipeTransform {
  transform(value: number, symbol = '$'): string {
    const formatted = new Intl.NumberFormat('es', {
      maximumFractionDigits: 0,
      useGrouping: true,
    }).format(value);

    return `${symbol} ${formatted}`;
  }
}
```

**Por qué recibe un número y no el café.** Si recibiera el café, este pipe
serviría para cafés y para nada más. Recibiendo el valor sirve para un precio, un
saldo, un premio y cualquier cosa que se muestre como dinero.

Es la misma pregunta que se hace con un componente hijo en S2: **¿qué necesita
para hacer su trabajo, y qué le están dando de más?**

**Por qué `pure: true` está escrito aunque sea el valor por defecto.** Porque la
alternativa es una decisión, y una decisión que no se ve no se revisa. En el code
review se puede preguntar por qué; si no está escrito, nadie se lo pregunta.

## Paso 3 · `HighlightDirective`

```ts
@Directive({
  selector: '[appHighlight]',
  standalone: true,
  host: {
    '[class.is-highlighted]': 'appHighlight()',
    '[attr.data-highlight-label]': 'appHighlight() ? highlightLabel() : null',
  },
})
export class HighlightDirective {
  readonly appHighlight = input(false);
  readonly highlightLabel = input('Destacado');
}
```

**Por qué el input se llama igual que el selector.** Para poder escribir
`[appHighlight]="condición"` en vez de `appHighlight [highlightWhen]="condición"`.
Es una convención de Angular, no una regla, y se usa cuando la directiva tiene un
input principal.

**Por qué `null` y no `''`.** `null` **quita** el atributo del DOM; la cadena
vacía lo deja puesto y vacío. Un lector de pantalla que busque
`data-highlight-label` los distingue.

**Y lo que hay que hacer notar en el code review:** del template desapareció el
nombre de la clase CSS. La pantalla dice **cuál** está destacado; **cómo** se ve
un destacado lo decide la directiva. Ese es el corte.

## Paso 4 · `RepeatDirective`

```ts
@Directive({ selector: '[appRepeat]', standalone: true })
export class RepeatDirective {
  readonly appRepeat = input.required<number>();

  private readonly template = inject(TemplateRef<unknown>);
  private readonly container = inject(ViewContainerRef);

  constructor() {
    effect(() => {
      const times = Math.max(0, Math.trunc(this.appRepeat()));
      this.container.clear();
      for (let index = 0; index < times; index += 1) {
        this.container.createEmbeddedView(this.template);
      }
    });
  }
}
```

**Qué es cada cosa que se inyecta.** `TemplateRef` es el HTML guardado sin
dibujar —el que está adentro del `ng-template` que el asterisco creó—.
`ViewContainerRef` es el lugar del DOM donde insertarlo.

**Por qué un `effect` y no leer el input una vez.** Porque el número puede
cambiar, y entonces hay que rehacer las copias. `effect` es la pieza de signals
que S3 no llegó a ver: sirve cuando hace falta que algo **pase**, no que se
calcule.

**Y el `clear()` antes del bucle** es lo que evita que las copias se acumulen.

---

# Parte B · El hipódromo

**Referencia terminada:** `project/frontend/solution/src/app/shared/`

## Paso 1 · Los dos pipes

```ts
// money.pipe.ts
transform(value: number, unit = 'pts'): string {
  const formatted = new Intl.NumberFormat('es', {
    maximumFractionDigits: 0,
    useGrouping: true,
  }).format(value);

  return unit === '' ? formatted : `${formatted} ${unit}`;
}
```

```ts
// odds.pipe.ts
transform(value: number): string {
  return new Intl.NumberFormat('es', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value);
}
```

**Por qué no se usa `| currency`.** El saldo es virtual, entero y no es ninguna
moneda real. `currency` traería símbolos, códigos ISO y centavos que no existen.

**Y lo que hay que señalar en el code review, porque casi nadie lo nota solo:**

```js
(2.4).toFixed(2)   // '2.40'  — siempre punto, en cualquier idioma
```

`toFixed` no sabe de idiomas. Estuvo mostrando un punto decimal en una aplicación
en español desde S1, y nadie lo vio. El pipe no solo evita repetir: **arregla algo
que estaba mal**.

## Paso 2 · `FavouriteDirective`

Misma forma que la del lab. Lo que cambia es el porqué:

**El favorito de una carrera es información, no decoración.** Un color solo se lo
lleva quien ve el color. El `data-favourite-label` es lo que hace que el dato
llegue también a quien navega con lector de pantalla, y ponerlo en la directiva
es lo que evita que alguien se olvide la próxima vez que muestre una parrilla.

## Paso 3 · Que no quede ninguno suelto

```bash
grep -rn "toFixed" src/app/features/
```

**Tiene que no devolver nada.** Es la verificación del ejercicio: no alcanza con
que el pipe exista, hay que haberlo usado en todos lados.

---

## Los cinco errores que aparecen todos los años

| Lo que hacen | Por qué no alcanza | Qué decirles |
|---|---|---|
| El pipe recibe el objeto entero | Sirve para una sola entidad | «Ahora usalo para el saldo del usuario.» |
| `pure: false` «por las dudas» | Corre cada revisión, para siempre | «¿Qué depende de algo que la entrada no ve?» |
| La directiva pone la clase desde el template | No se movió nada, solo se renombró | «¿Dónde está escrito el nombre de la clase CSS?» |
| Se olvidan `LOCALE_ID` | Media pantalla en inglés | «Mira el porcentaje.» |
| Queda algún `toFixed` | El formato sigue en dos lugares | «`grep toFixed`.» |

---

## Cómo se verifica que quedó bien

```bash
cd lab/starter              && npx tsc --noEmit && npm test
cd project/frontend/starter && npx tsc --noEmit && npm run build
node scripts/verify.mjs --fast
```

Y a ojo:

- El porcentaje sale **con coma**.
- `grep -rn "toFixed" src/app/features/` no devuelve nada.
- `grep -rn "pts" src/app/` devuelve **una** línea.
- `odds.pipe.ts` no menciona ni `Race` ni `Horse`.
