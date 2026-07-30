import { ChangeDetectionStrategy, Component, input } from '@angular/core';

export type CalloutTone = 'danger' | 'warning' | 'success' | 'info';

/**
 * El aviso de una pantalla: el error del servidor, la confirmación de una
 * apuesta, el resultado de una carrera.
 *
 * `role` cambia con el tono: un error es `alert` —interrumpe— y el resto es
 * `status`, que se anuncia cuando el lector de pantalla termine lo que estaba
 * diciendo. Poner `alert` en todo hace que nada interrumpa.
 */
@Component({
  selector: 'app-callout',
  changeDetection: ChangeDetectionStrategy.OnPush,
  styleUrl: './callout.css',
  host: {
    '[class]': '"callout callout--" + tone()',
    '[attr.role]': 'tone() === "danger" ? "alert" : "status"',
  },
  template: `
    @if (title()) {
      <p class="callout__title">{{ title() }}</p>
    }
    <ng-content />
  `,
})
export class Callout {
  readonly tone = input<CalloutTone>('info');
  readonly title = input<string>('');
}
