# 1 · Un pipe que existe y no se declara

**No ejecutes todavía.** Mira el código y escribe en el chat qué va a pasar.

---

El pipe está escrito, en su archivo, y compila:

```ts
@Pipe({ name: 'money', standalone: true })
export class MoneyPipe implements PipeTransform {
  transform(value: number, symbol = '$'): string { … }
}
```

El componente lo usa **sin ponerlo en `imports`**:

```ts
@Component({
  selector: 'app-s04',
  standalone: true,
  imports: [],
  template: `<p>{{ 4200 | money }}</p>`,
})
export class S04Component {}
```

### La pregunta

**¿Qué pasa?**

1. Se ve `4200`, sin transformar
2. Se ve vacío
3. No compila

Escribe tu predicción en el chat antes de que ejecutemos.

---

## Cómo reproducirlo en clase

En `lab/demo/src/app/sessions/s04/s04.component.ts`, quita `MoneyPipe` de
`imports` **y también su línea de `import`** — si dejas el import sin usar, el
error que sale es otro (`TS6133`) y tapa el que interesa.

Muestra la terminal. Después se repone.
