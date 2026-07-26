# 3 · El contador que no cuenta

**No ejecutes todavía.** Mirá el código y escribí en el chat qué va a pasar.

---

```ts
@Component({
  selector: 'app-demo',
  standalone: true,
  template: `
    <button (click)="contador + 1">Sumar</button>
    <p>Llevo {{ contador }} clics</p>
  `,
})
export class DemoComponent {
  contador = 0;
}
```

### La pregunta

**¿Qué muestra el párrafo después de tocar el botón tres veces?**

Y la de verdad: **¿por qué?**

Escribí tu predicción en el chat antes de que ejecutemos.

---

## Cómo reproducirlo en clase

En `lab/solution/src/app/sessions/s01/s01.component.html`, cambiá temporalmente:

```html
<button type="button" class="button button--ghost" (click)="toggleAvailability()">
```

por:

```html
<button type="button" class="button button--ghost" (click)="coffee.available">
```

El botón deja de hacer nada. **No hay ningún error.** Eso es lo interesante.
