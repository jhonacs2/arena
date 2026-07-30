import { ChangeDetectionStrategy, Component, inject, signal } from '@angular/core';
import { FormField, FormRoot, form, required, submit } from '@angular/forms/signals';
import { Router, RouterLink } from '@angular/router';

import { toApiError } from '../../core/api/api-error';
import { AuthService } from '../../core/auth/auth.service';
import { Button } from '../../shared/ui/button/button';
import { Callout } from '../../shared/ui/callout/callout';
import { FieldErrors } from '../../shared/ui/field/field-errors';

interface Credentials {
  username: string;
  password: string;
}

/**
 * Inicio de sesión, para quien ya canjeó su código.
 *
 * No hay «olvidé mi contraseña»: no hay correo verificado con el que mandar nada
 * (`decisiones.md` §2). Si alguien la pierde, el instructor le regenera la cuenta
 * —y el ledger, que es lo que vale, sigue estando.
 */
@Component({
  selector: 'app-login',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [FormField, FormRoot, RouterLink, Button, Callout, FieldErrors],
  templateUrl: './login.html',
  styleUrls: ['../../shared/ui/field/form-controls.css', './login.css'],
})
export class Login {
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  private readonly model = signal<Credentials>({ username: '', password: '' });

  protected readonly credentials = form(this.model, (path) => {
    required(path.username, { message: 'Escribí tu usuario.' });
    required(path.password, { message: 'Escribí tu contraseña.' });
  });

  private readonly _failure = signal<string | null>(null);
  protected readonly failure = this._failure.asReadonly();

  protected async send(): Promise<void> {
    this._failure.set(null);

    await submit(this.credentials, {
      action: async () => {
        try {
          const session = await this.auth.login(this.model());
          await this.router.navigate([session.user.role === 'admin' ? '/instructor' : '/tablero']);
        } catch (cause) {
          // `INVALID_CREDENTIALS` no dice cuál de los dos campos está mal, y está
          // bien que no lo diga: decirlo confirmaría qué usuarios existen.
          this._failure.set(toApiError(cause).message);
        }
        return undefined;
      },
    });
  }
}
