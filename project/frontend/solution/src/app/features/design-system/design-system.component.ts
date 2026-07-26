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
  selector: 'app-design-system',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [SilkComponent, ButtonComponent, SkeletonComponent, EmptyStateComponent],
  templateUrl: './design-system.component.html',
  styleUrl: './design-system.component.css',
})
export class DesignSystemComponent {
  /** Una carrera de verdad del dataset: 8 caballos, 8 sedas distintas. */
  protected readonly race = RACES.find((race) => race.id === 'race_005') ?? RACES[0]!;

  protected readonly colors = SILK_COLORS;

  protected readonly semanticTokens = [
    { token: 'surface', use: 'fondo de la página' },
    { token: 'surface-sunken', use: 'fondo hundido' },
    { token: 'text', use: 'texto de cuerpo' },
    { token: 'text-muted', use: 'texto secundario' },
    { token: 'accent', use: 'acción principal' },
    { token: 'live', use: 'carrera en vivo' },
    { token: 'success', use: 'apuesta ganada' },
    { token: 'danger', use: 'error' },
  ] as const;

  protected readonly scale = [
    { token: '2xs', use: 'etiquetas' },
    { token: 'xs', use: 'pies de card' },
    { token: 'sm', use: 'botones' },
    { token: 'base', use: 'cuerpo' },
    { token: 'lg', use: 'subtítulos' },
    { token: 'xl', use: 'nombre de carrera' },
    { token: '2xl', use: 'título de sección' },
    { token: '4xl', use: 'marcador' },
  ] as const;

  /** El saldo de ejemplo, para ver el salto tipo split-flap. */
  protected readonly balance = signal(5000);

  protected collect(): void {
    this.balance.update((actual) => actual + 1100);
  }

  protected resetBalance(): void {
    this.balance.set(5000);
  }
}
