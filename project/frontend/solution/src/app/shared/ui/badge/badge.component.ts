import { ChangeDetectionStrategy, Component, input } from '@angular/core';

export type BadgeTone = 'neutral' | 'live' | 'success' | 'accent';

/**
 * S2 · La pastilla de estado.
 *
 *   <app-badge tone="live">En vivo</app-badge>
 *   <app-badge>{{ view.statusLabel }}</app-badge>
 *
 * El texto **no** es un `input()`: entra por `<ng-content>`. La diferencia no
 * es de estilo, es de responsabilidad — con `ng-content` el padre puede meter
 * lo que quiera adentro (un texto, un número, un icono y un texto) sin que la
 * pastilla tenga que prever cada caso con un input nuevo.
 *
 * `tone` sí es un `input()`, porque no es contenido: es una decisión sobre
 * cómo se ve, y el conjunto de valores posibles está cerrado.
 */
@Component({
  selector: 'app-badge',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <span [class]="'badge badge--' + tone()">
      <ng-content />
    </span>
  `,
  styleUrl: './badge.component.css',
})
export class BadgeComponent {
  readonly tone = input<BadgeTone>('neutral');
}
