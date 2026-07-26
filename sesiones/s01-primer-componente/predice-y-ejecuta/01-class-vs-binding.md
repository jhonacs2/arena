# 1 · Dos formas de poner una clase

**No ejecutes todavía.** Mirá el código y escribí en el chat qué va a pasar.

---

```ts
@Component({
  selector: 'app-demo',
  standalone: true,
  template: `
    <p class="label {{ tone }}" [class.etiqueta--activa]="activo">
      Hola
    </p>
  `,
  styles: `
    .label { border: 3px solid; padding: 8px; }
    .rojo { color: red; }
    .etiqueta--activa { background: gold; }
  `,
})
export class DemoComponent {
  tone = 'rojo';
  activo = true;
}
```

### La pregunta

**¿El párrafo termina rojo, con fondo dorado, las dos cosas o ninguna?**

Escribí tu predicción en el chat antes de que ejecutemos.

---

## Cómo reproducirlo en clase

En `lab/solution/src/app/sessions/s01/s01.component.html`, en el `<div class="product">`, agregá temporalmente:

```html
<p class="label {{ tone }}" [class.etiqueta--activa]="activo">Hola</p>
```

y en la clase:

```ts
protected tone = 'rojo';
protected activo = true;
```

Después de ejecutar y explicar, se borra.
