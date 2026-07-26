import { ChangeDetectionStrategy, Component, signal } from '@angular/core';

import { RACES } from '../../core/mocks';
import { ButtonComponent } from '../../shared/ui/button/button.component';
import { EmptyStateComponent } from '../../shared/ui/empty-state/empty-state.component';
import { SilkComponent } from '../../shared/ui/silk/silk.component';
import { SkeletonComponent } from '../../shared/ui/skeleton/skeleton.component';
import { SILK_COLORS } from '../../shared/ui/silk/silk.util';

/**
 * Muestra del sistema de diseño.
 *
 * No es una pantalla del producto: es la prueba de que la base funciona —
 * tokens, tipografías, sedas y primitivas— y la referencia contra la que
 * mirar cuando algo se ve raro.
 *
 * En S1 esta ruta deja de ser la principal, pero se queda: es más rápido
 * revisar acá que abrir cinco pantallas del producto.
 */
@Component({
  selector: 'app-sistema',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [SilkComponent, ButtonComponent, SkeletonComponent, EmptyStateComponent],
  templateUrl: './sistema.component.html',
  styleUrl: './sistema.component.css',
})
export class SistemaComponent {
  /** Una carrera de verdad del dataset: 8 caballos, 8 sedas distintas. */
  protected readonly carrera = RACES.find((race) => race.id === 'race_005') ?? RACES[0]!;

  protected readonly colores = SILK_COLORS;

  protected readonly semanticos = [
    { token: 'surface', uso: 'fondo de la página' },
    { token: 'surface-sunken', uso: 'fondo hundido' },
    { token: 'text', uso: 'texto de cuerpo' },
    { token: 'text-muted', uso: 'texto secundario' },
    { token: 'accent', uso: 'acción principal' },
    { token: 'live', uso: 'carrera en vivo' },
    { token: 'success', uso: 'apuesta ganada' },
    { token: 'danger', uso: 'error' },
  ] as const;

  protected readonly escala = [
    { token: '2xs', uso: 'etiquetas' },
    { token: 'xs', uso: 'pies de card' },
    { token: 'sm', uso: 'botones' },
    { token: 'base', uso: 'cuerpo' },
    { token: 'lg', uso: 'subtítulos' },
    { token: 'xl', uso: 'nombre de carrera' },
    { token: '2xl', uso: 'título de sección' },
    { token: '4xl', uso: 'marcador' },
  ] as const;

  /** El saldo de ejemplo, para ver el salto tipo split-flap. */
  protected readonly saldo = signal(5000);

  protected cobrar(): void {
    this.saldo.update((actual) => actual + 1100);
  }

  protected reiniciarSaldo(): void {
    this.saldo.set(5000);
  }
}
