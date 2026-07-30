import { ChangeDetectionStrategy, Component, input } from '@angular/core';

/**
 * El estado vacío.
 *
 * Existe como componente para que ninguna pantalla se olvide de tenerlo: una
 * lista vacía sin explicación se lee como una app rota, y en clase eso son diez
 * minutos de «a mí no me carga».
 */
@Component({
  selector: 'app-empty-state',
  changeDetection: ChangeDetectionStrategy.OnPush,
  styleUrl: './empty-state.css',
  template: `
    <p class="empty__mark" aria-hidden="true">{{ mark() }}</p>
    <h3 class="empty__title">{{ title() }}</h3>
    @if (detail()) {
      <p class="empty__detail">{{ detail() }}</p>
    }
    <ng-content />
  `,
})
export class EmptyState {
  readonly title = input.required<string>();
  readonly detail = input<string>('');
  readonly mark = input<string>('—');
}
