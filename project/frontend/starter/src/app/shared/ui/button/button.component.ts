import { ChangeDetectionStrategy, Component, input, output } from '@angular/core';

export type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger';
export type ButtonSize = 'sm' | 'md' | 'lg';

/**
 * El botón del sistema.
 *
 *   <app-button (pressed)="placeBet()">Apostar</app-button>
 *   <app-button variant="ghost" size="sm" [disabled]="cargando()">Cancelar</app-button>
 *
 * Envuelve un `<button>` nativo en vez de estilar un `<div>`: así el teclado,
 * el foco y los lectores de pantalla funcionan sin que haya que reimplementarlos.
 */
@Component({
  selector: 'app-button',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <button
      [type]="type()"
      [class]="'button button--' + variant() + ' button--' + size()"
      [disabled]="disabled()"
      [attr.aria-busy]="loading() ? 'true' : null"
      (click)="pressed.emit()"
    >
      <ng-content />
    </button>
  `,
  styleUrl: './button.component.css',
})
export class ButtonComponent {
  readonly variant = input<ButtonVariant>('primary');
  readonly size = input<ButtonSize>('md');
  readonly type = input<'button' | 'submit'>('button');
  readonly disabled = input(false);
  /** Marca el botón como ocupado sin deshabilitarlo: el foco no se pierde. */
  readonly loading = input(false);

  readonly pressed = output<void>();
}
