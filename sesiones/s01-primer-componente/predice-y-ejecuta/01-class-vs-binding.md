# 1 · Dos formas de poner una clase

**No ejecutes todavía.** Mirá el código y escribí en el chat qué va a pasar.

---

```ts
@Component({
  selector: 'app-demo',
  standalone: true,
  template: `
    <p class="etiqueta {{ tono }}" [class.etiqueta--activa]="activo">
      Hola
    </p>
  `,
  styles: `
    .etiqueta { border: 3px solid; padding: 8px; }
    .rojo { color: red; }
    .etiqueta--activa { background: gold; }
  `,
})
export class DemoComponent {
  tono = 'rojo';
  activo = true;
}
```

### La pregunta

**¿El párrafo termina rojo, con fondo dorado, las dos cosas o ninguna?**

Escribí tu predicción en el chat antes de que ejecutemos.

---

## Cómo reproducirlo en clase

En `lab/solution/src/app/sesiones/s01/s01.component.html`, en el `<div class="producto">`, agregá temporalmente:

```html
<p class="etiqueta {{ tono }}" [class.etiqueta--activa]="activo">Hola</p>
```

y en la clase:

```ts
protected tono = 'rojo';
protected activo = true;
```

Después de ejecutar y explicar, se borra.
