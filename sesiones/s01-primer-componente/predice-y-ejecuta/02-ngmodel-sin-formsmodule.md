# 2 · El input que no se enlaza

**No ejecutes todavía.** Mirá el código y escribí en el chat qué va a pasar.

---

```ts
import { Component } from '@angular/core';

@Component({
  selector: 'app-demo',
  standalone: true,
  imports: [],
  template: `
    <input type="text" [(ngModel)]="nombre" />
    <p>Hola, {{ nombre }}</p>
  `,
})
export class DemoComponent {
  nombre = '';
}
```

### La pregunta

**¿Qué pasa cuando escribo en el input?**

1. El saludo se actualiza mientras escribo.
2. El saludo no se actualiza, pero la app funciona.
3. La app no arranca.
4. `ng serve` no compila.

Escribí tu número en el chat antes de que ejecutemos.

---

## Cómo reproducirlo en clase

En `lab/solution/src/app/sesiones/s01/s01.component.ts`, sacá `FormsModule` de `imports` y guardá. La consola de `ng serve` va a hablar sola.
