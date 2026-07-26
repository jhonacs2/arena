import { ChangeDetectionStrategy, Component, input, output } from '@angular/core';

export type ButtonVariant = 'primario' | 'secundario' | 'fantasma' | 'peligro';
export type ButtonSize = 'sm' | 'md' | 'lg';

/**
 * El botón del sistema.
 *
 *   <app-button (accionar)="apostar()">Apostar</app-button>
 *   <app-button variant="fantasma" size="sm" [disabled]="cargando()">Cancelar</app-button>
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
      [class]="'boton boton--' + variant() + ' boton--' + size()"
      [disabled]="disabled()"
      [attr.aria-busy]="loading() ? 'true' : null"
      (click)="accionar.emit()"
    >
      <ng-content />
    </button>
  `,
  styleUrl: './button.component.css',
})
export class ButtonComponent {
  readonly variant = input<ButtonVariant>('primario');
  readonly size = input<ButtonSize>('md');
  readonly type = input<'button' | 'submit'>('button');
  readonly disabled = input(false);
  /** Marca el botón como ocupado sin deshabilitarlo: el foco no se pierde. */
  readonly loading = input(false);

  readonly accionar = output<void>();
}
