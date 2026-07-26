import { ChangeDetectionStrategy, Component, computed, input } from '@angular/core';

import { silkDrawing } from './silk.util';

export type SilkSize = 'sm' | 'md' | 'lg';

/**
 * La seda del jockey de un caballo.
 *
 * Se deriva del `id`, no se elige ni se guarda en ninguna parte: mismo caballo,
 * misma casaca, en la lista, en el detalle y en la carrera en vivo.
 *
 *   <app-silk [horseId]="horse.id" [horseName]="horse.name" size="lg" />
 *
 * El `alt` no es decorativo: en la carrera en vivo hay ocho de estos seguidos y
 * un lector de pantalla tiene que poder decir de quién es cada uno.
 */
@Component({
  selector: 'app-silk',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  templateUrl: './silk.component.html',
  styleUrl: './silk.component.css',
  host: {
    '[class]': 'size()',
    role: 'img',
    '[attr.aria-label]': 'label()',
  },
})
export class SilkComponent {
  readonly horseId = input.required<string>();
  readonly horseName = input<string>('');
  readonly size = input<SilkSize>('md');

  /** El dibujo se recalcula solo si cambia el id. */
  protected readonly drawing = computed(() => silkDrawing(this.horseId()));

  protected readonly label = computed(() => {
    const name = this.horseName();
    const { primary, secondary, body } = this.drawing().spec;
    const description = `seda ${primary} y ${secondary}, patrón ${body}`;
    return name ? `${name}: ${description}` : description;
  });
}
