import { ChangeDetectionStrategy, Component, computed, input } from '@angular/core';

import { silkDrawing } from './silk.util';

export type SilkSize = 'sm' | 'md' | 'lg';

/**
 * La seda del jockey de un caballo — el elemento firma del sistema visual.
 *
 * Se deriva del `id` con una función pura: mismo caballo, misma casaca, en la
 * lista, en la sala y en la carrera en vivo. Es lo que le da identidad visual a
 * cada caballo sin un solo archivo de imagen.
 *
 *   <app-silk [horseId]="horse.id" [horseName]="horse.name" size="lg" />
 *
 * El `aria-label` no es decorativo: en la carrera en vivo hay seis de estos
 * seguidos y un lector de pantalla tiene que poder decir de quién es cada uno.
 */
@Component({
  selector: 'app-silk',
  changeDetection: ChangeDetectionStrategy.OnPush,
  templateUrl: './silk.html',
  styleUrl: './silk.css',
  host: {
    '[class]': 'size()',
    role: 'img',
    '[attr.aria-label]': 'label()',
  },
})
export class Silk {
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
