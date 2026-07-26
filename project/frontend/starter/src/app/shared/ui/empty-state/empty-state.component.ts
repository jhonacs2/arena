import { ChangeDetectionStrategy, Component, input, output } from '@angular/core';

import { ButtonComponent } from '../button/button.component';

/**
 * Estado vacío y estado de error. Los dos son el mismo componente porque en
 * pantalla son la misma forma: un mensaje y una salida.
 *
 *   <app-empty-state titulo="Todavía no apostaste"
 *                    mensaje="Elegí una carrera y probá suerte."
 *                    accion="Ver carreras" (accionar)="irACarreras()" />
 *
 *   <app-empty-state tono="error" titulo="No pudimos cargar las carreras"
 *                    accion="Reintentar" (accionar)="recargar()" />
 *
 * Una pantalla vacía es una invitación a hacer algo, no un cartel de "no hay
 * nada". Por eso `accion` existe: si no hay salida, el estado vacío es una
 * puerta cerrada.
 */
@Component({
  selector: 'app-empty-state',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [ButtonComponent],
  template: `
    <div class="vacio" [class.vacio--error]="tono() === 'error'" [attr.role]="tono() === 'error' ? 'alert' : null">
      <p class="etiqueta">{{ tono() === 'error' ? 'Algo salió mal' : 'Sin resultados' }}</p>
      <h3 class="titulo">{{ titulo() }}</h3>

      @if (mensaje()) {
        <p class="mensaje">{{ mensaje() }}</p>
      }

      @if (accion()) {
        <app-button [variant]="tono() === 'error' ? 'peligro' : 'primario'" (accionar)="accionar.emit()">
          {{ accion() }}
        </app-button>
      }
    </div>
  `,
  styleUrl: './empty-state.component.css',
})
export class EmptyStateComponent {
  readonly titulo = input.required<string>();
  readonly mensaje = input('');
  /** Texto del botón. Vacío lo oculta. */
  readonly accion = input('');
  readonly tono = input<'vacio' | 'error'>('vacio');

  readonly accionar = output<void>();
}
