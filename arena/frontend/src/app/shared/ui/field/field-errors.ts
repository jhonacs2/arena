import { ChangeDetectionStrategy, Component, computed, input } from '@angular/core';

/** Lo mínimo que tiene un error de Signal Forms. */
export interface FieldError {
  readonly kind: string;
  readonly message?: string;
}

/**
 * Los errores de un campo.
 *
 * Solo se muestran cuando el campo ya fue tocado: gritarle «esto está mal» a
 * alguien que todavía no terminó de escribir la primera letra es hostil, y hace
 * que el formulario se vea roto antes de que exista un error de verdad.
 *
 * El contenedor existe siempre y es `aria-live`: si apareciera recién con el
 * error, el lector de pantalla no anunciaría nada.
 */
@Component({
  selector: 'app-field-errors',
  changeDetection: ChangeDetectionStrategy.OnPush,
  styleUrl: './field-errors.css',
  host: { 'aria-live': 'polite' },
  template: `
    @for (error of visible(); track error.kind) {
      <p class="field-error">{{ error.message ?? 'Ese valor no es válido.' }}</p>
    }
  `,
})
export class FieldErrors {
  readonly errors = input<readonly FieldError[]>([]);
  readonly touched = input(false);

  protected readonly visible = computed(() => (this.touched() ? this.errors() : []));
}
