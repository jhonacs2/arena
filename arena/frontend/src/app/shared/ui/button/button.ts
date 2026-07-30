import { ChangeDetectionStrategy, Component, computed, input } from '@angular/core';

export type ButtonVariant = 'primary' | 'secondary' | 'danger' | 'ghost';
export type ButtonSize = 'sm' | 'md' | 'lg';

/**
 * El botón del sistema.
 *
 * Es el único lugar del repo donde se define el efecto de hundido: al presionar,
 * el objeto se desplaza **exactamente lo que pierde de sombra**. Es el único
 * efecto de profundidad que el neobrutalismo se permite.
 */
@Component({
  selector: 'app-button',
  changeDetection: ChangeDetectionStrategy.OnPush,
  styleUrl: './button.css',
  template: `
    <button
      [class]="classes()"
      [type]="type()"
      [disabled]="disabled() || loading()"
      [attr.aria-busy]="loading() ? 'true' : null"
    >
      @if (loading()) {
        <span class="button__spinner" aria-hidden="true"></span>
      }
      <ng-content />
    </button>
  `,
})
export class Button {
  readonly variant = input<ButtonVariant>('primary');
  readonly size = input<ButtonSize>('md');
  readonly type = input<'button' | 'submit'>('button');
  readonly disabled = input(false);
  readonly loading = input(false);

  protected readonly classes = computed(
    () => `button button--${this.variant()} button--${this.size()}`,
  );
}
