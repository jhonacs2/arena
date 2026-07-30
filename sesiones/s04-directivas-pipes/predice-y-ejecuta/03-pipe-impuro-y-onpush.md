# 3 · Un pipe impuro en un componente `OnPush`

**No ejecutes todavía.** Mira el código y escribe en el chat qué va a pasar.

---

```ts
@Pipe({ name: 'countImpure', standalone: true, pure: false })
export class CountImpurePipe implements PipeTransform {
  transform(value: string): string {
    CALL_COUNT.impure += 1;
    return value;
  }
}
```

```ts
@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `{{ 'café' | countImpure }}`,
})
export class DemoComponent { … }
```

El valor de entrada es **siempre la misma cadena**, `'café'`, y nunca cambia.

### La pregunta

**El contador arranca en 1. ¿Cuándo sube?**

1. Nunca: la entrada no cambia
2. Con cada clic en **cualquier parte** de la aplicación
3. Solo cuando Angular revisa **ese** componente

Escribe tu predicción en el chat antes de que ejecutemos.

---

## Cómo reproducirlo en clase

Ya está en la pantalla, en el panel «Puro contra impuro».

**Toca el botón varias veces** y mira los dos contadores. Después, si sobra
tiempo, la demostración que cierra el tema: cambia `ChangeDetectionStrategy.OnPush`
por `Default` en el componente y vuelve a probar.

Después se repone el `OnPush`.
