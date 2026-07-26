import { ChangeDetectionStrategy, Component, input } from '@angular/core';

/**
 * Bloque de carga.
 *
 *   <app-skeleton [lines]="3" />
 *   <app-skeleton shape="card" />
 *
 * Es uno de los tres estados obligatorios de toda view con datos
 * (`CLAUDE.md` §11). Ocupa aproximadamente el espacio del contenido real para
 * que la página no salte cuando llegan los datos.
 */
@Component({
  selector: 'app-skeleton',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <div class="bone hueso--{{ shape() }}" [attr.aria-hidden]="true">
      @for (line of placeholders(); track $index) {
        <span class="bar"></span>
      }
    </div>
    <span class="sr-only" role="status">{{ label() }}</span>
  `,
  styleUrl: './skeleton.component.css',
})
export class SkeletonComponent {
  readonly shape = input<'text' | 'card' | 'row'>('text');
  readonly lines = input(3);
  readonly label = input('Cargando…');

  protected placeholders(): readonly number[] {
    return Array.from({ length: Math.max(1, this.lines()) }, (_, i) => i);
  }
}
