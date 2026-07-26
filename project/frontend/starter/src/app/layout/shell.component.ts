import { ChangeDetectionStrategy, Component, computed, inject } from '@angular/core';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';

import { ThemeService } from '../core/theme/theme.service';
import { LogoComponent } from '../shared/ui/logo/logo.component';

/**
 * El armazón de la app: encabezado, navegación y el `router-outlet`.
 *
 * El widget de saldo entra en S10, cuando haya sesión y socket. Hasta entonces
 * el encabezado es marca, navegación y el interruptor de tema.
 */
@Component({
  selector: 'app-shell',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [RouterOutlet, RouterLink, RouterLinkActive, LogoComponent],
  templateUrl: './shell.component.html',
  styleUrl: './shell.component.css',
})
export class ShellComponent {
  private readonly themes = inject(ThemeService);

  protected readonly theme = this.themes.theme;

  protected readonly themeLabel = computed(
    () =>
      ({
        light: 'Tema claro. Cambiar a oscuro.',
        dark: 'Tema oscuro. Cambiar al del sistema.',
        system: 'Tema del sistema. Cambiar a claro.',
      })[this.theme()],
  );

  protected readonly themeIcon = computed(
    () => ({ light: '☀', dark: '☾', system: '◐' })[this.theme()],
  );

  protected cycleTheme(): void {
    this.themes.cycle();
  }
}
