import { ChangeDetectionStrategy, Component, input, output } from '@angular/core';

import { ButtonComponent } from '../button/button.component';

/**
 * Estado vacío y estado de error. Los dos son el mismo componente porque en
 * pantalla son la misma forma: un mensaje y una salida.
 *
 *   <app-empty-state title="Todavía no apostaste"
 *                    mensaje="Elegí una carrera y probá suerte."
 *                    action="Ver carreras" (pressed)="goToRaces()" />
 *
 *   <app-empty-state tone="error" title="No pudimos cargar las carreras"
 *                    action="Reintentar" (pressed)="reload()" />
 *
 * Una pantalla vacía es una invitación a hacer algo, no un cartel de "no hay
 * nada". Por eso `action` existe: si no hay salida, el estado vacío es una
 * puerta cerrada.
 */
@Component({
  selector: 'app-empty-state',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [ButtonComponent],
  template: `
    <div class="empty" [class.empty--error]="tone() === 'error'" [attr.role]="tone() === 'error' ? 'alert' : null">
      <p class="label">{{ tone() === 'error' ? 'Algo salió mal' : 'Sin resultados' }}</p>
      <h3 class="title">{{ title() }}</h3>

      @if (message()) {
        <p class="message">{{ message() }}</p>
      }

      @if (action()) {
        <app-button [variant]="tone() === 'error' ? 'danger' : 'primary'" (pressed)="pressed.emit()">
          {{ action() }}
        </app-button>
      }
    </div>
  `,
  styleUrl: './empty-state.component.css',
})
export class EmptyStateComponent {
  readonly title = input.required<string>();
  readonly message = input('');
  /** Texto del botón. Vacío lo oculta. */
  readonly action = input('');
  readonly tone = input<'empty' | 'error'>('empty');

  readonly pressed = output<void>();
}
