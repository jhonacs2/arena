import { ChangeDetectionStrategy, Component, computed, inject } from '@angular/core';
import { Router, RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';

import { AuthService } from './core/auth/auth.service';
import { SessionStore } from './core/auth/session.store';
import { ThemeService, type Theme } from './core/theme/theme.service';
import { CoinsPipe } from './shared/format/coins.pipe';

const THEME_LABEL: Readonly<Record<Theme, string>> = {
  light: 'Tema claro. Cambiar a oscuro',
  dark: 'Tema oscuro. Cambiar a automático',
  system: 'Tema automático. Cambiar a claro',
};

const THEME_MARK: Readonly<Record<Theme, string>> = {
  light: '☀',
  dark: '☾',
  system: '◐',
};

/**
 * El cascarón de la app: cabecera, saldo, tema, salida y el `<router-outlet>`.
 *
 * El saldo vive acá y no en cada pantalla porque es lo que el alumno mira todo
 * el tiempo: son sus monedas, y sus monedas son nota.
 */
@Component({
  selector: 'app-root',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [RouterOutlet, RouterLink, RouterLinkActive, CoinsPipe],
  templateUrl: './app.html',
  styleUrl: './app.css',
})
export class App {
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);
  protected readonly session = inject(SessionStore);
  protected readonly theme = inject(ThemeService);

  protected readonly themeLabel = computed(() => THEME_LABEL[this.theme.theme()]);
  protected readonly themeMark = computed(() => THEME_MARK[this.theme.theme()]);

  protected async signOut(): Promise<void> {
    await this.auth.logout();
    await this.router.navigate(['/ingresar']);
  }
}
