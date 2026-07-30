import { httpResource } from '@angular/common/http';
import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';
import {
  FormField,
  FormRoot,
  disabled,
  form,
  minLength,
  pattern,
  required,
  submit,
  validate,
} from '@angular/forms/signals';
import { Router, RouterLink } from '@angular/router';

import { api } from '../../core/api/api';
import { toApiError } from '../../core/api/api-error';
import { AuthService } from '../../core/auth/auth.service';
import type { CheckCodeResponse } from '../../core/models';
import { Button } from '../../shared/ui/button/button';
import { Callout } from '../../shared/ui/callout/callout';
import { FieldErrors } from '../../shared/ui/field/field-errors';

/**
 * El código es `AAAA-9999`: cuatro letras, guion, cuatro dígitos.
 *
 * El guion es opcional al escribir. Se dicta en voz alta y se copia de un chat,
 * y pelearle a alguien porque no puso el guion es una llamada de soporte que no
 * hacía falta: se normaliza y listo.
 */
const TYPED = /^[A-Za-z]{4}-?[0-9]{4}$/;

function normalizeCode(raw: string): string {
  const clean = raw
    .toUpperCase()
    .replace(/[^A-Z0-9]/g, '')
    .slice(0, 8);
  return clean.length > 4 ? `${clean.slice(0, 4)}-${clean.slice(4)}` : clean;
}

interface Profile {
  firstName: string;
  lastName: string;
  username: string;
  password: string;
}

interface Registration {
  code: string;
  profile: Profile;
}

/**
 * Registro por código de invitación. **Es la primera pantalla que ve un alumno.**
 *
 * Un solo formulario, dos tiempos: primero el código, y cuando el servidor
 * confirma que sirve, aparece el resto. El resto está `disabled` de verdad —no
 * escondido— así se ve desde el primer segundo qué va a haber que completar.
 *
 * Los dos errores de código se muestran **distinto a propósito**: «no existe» es
 * un error de tipeo y lleva a revisar las letras; «ya fue usado» no es un error,
 * es alguien que ya se registró, y lleva a iniciar sesión. Tratarlos igual manda
 * a la mitad del aula a pedir un código nuevo que no necesitan.
 */
@Component({
  selector: 'app-register',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [FormField, FormRoot, RouterLink, Button, Callout, FieldErrors],
  templateUrl: './register.html',
  styleUrls: ['../../shared/ui/field/form-controls.css', './register.css'],
})
export class Register {
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  private readonly model = signal<Registration>({
    code: '',
    profile: { firstName: '', lastName: '', username: '', password: '' },
  });

  /** El código normalizado, o `null` si todavía no está completo. */
  private readonly candidate = computed(() => {
    const typed = this.model().code;
    return TYPED.test(typed) ? normalizeCode(typed) : null;
  });

  /**
   * La comprobación del código, con `httpResource`.
   *
   * Devolver `undefined` en la request deja el recurso en `idle`: no hay una
   * llamada por tecla, solo una cuando el código está completo. Con un formato
   * fijo de ocho caracteres eso es exactamente una llamada por intento, así que
   * no hace falta debounce.
   */
  protected readonly codeCheck = httpResource<CheckCodeResponse>(() => {
    const code = this.candidate();
    return code === null
      ? undefined
      : { url: api('/auth/check-code'), method: 'POST', body: { code } };
  });

  /** El error del servidor, ya traducido al sobre del contrato. */
  protected readonly codeError = computed(() => {
    if (this.codeCheck.status() !== 'error') return null;
    return toApiError(this.codeCheck.error());
  });

  protected readonly codeAccepted = computed(
    () => this.codeCheck.status() === 'resolved' && this.codeCheck.value()?.valid === true,
  );

  protected readonly coinsGranted = computed(() => this.codeCheck.value()?.coinsGranted ?? 0);
  /** 100 monedas = 1 punto. La cuenta la hace la app, no el alumno. */
  protected readonly pointsGranted = computed(() => Math.floor(this.coinsGranted() / 100));

  protected readonly registration = form(this.model, (path) => {
    required(path.code, { message: 'Escribí el código que te dieron.' });
    pattern(path.code, TYPED, { message: 'El formato es AAAA-9999: cuatro letras y cuatro números.' });

    /**
     * El resultado del servidor entra al formulario como un error del campo.
     *
     * Es un `validate` y no un `validateHttp` porque de la respuesta hace falta
     * algo más que «sirve o no sirve»: las monedas que otorga, para poder
     * mostrarlas. `validateHttp` no expone el cuerpo de la respuesta.
     */
    validate(path.code, () => {
      const failure = this.codeError();
      if (failure === null) return undefined;
      return { kind: failure.code, message: failure.message };
    });

    // Un solo `disabled` sobre el subárbol: el código habilita el resto.
    disabled(path.profile, ({ stateOf }) => !stateOf(path.code).valid() || !this.codeAccepted(), );

    required(path.profile.firstName, { message: 'Falta el nombre.' });
    required(path.profile.lastName, { message: 'Falta el apellido.' });
    required(path.profile.username, { message: 'Elegí un usuario.' });
    minLength(path.profile.username, 3, { message: 'El usuario necesita 3 caracteres o más.' });
    required(path.profile.password, { message: 'Elegí una contraseña.' });
    minLength(path.profile.password, 8, { message: 'La contraseña necesita 8 caracteres o más.' });
  });

  private readonly _failure = signal<string | null>(null);
  protected readonly failure = this._failure.asReadonly();

  protected async send(): Promise<void> {
    this._failure.set(null);

    await submit(this.registration, {
      action: async () => {
        const { code, profile } = this.model();
        try {
          const session = await this.auth.redeem({
            code: normalizeCode(code),
            firstName: profile.firstName.trim(),
            lastName: profile.lastName.trim(),
            username: profile.username.trim(),
            password: profile.password,
          });
          await this.router.navigate([session.user.role === 'admin' ? '/instructor' : '/tablero']);
          return undefined;
        } catch (cause) {
          const failure = toApiError(cause);
          // El error va al campo que lo causó cuando se sabe cuál es. Un cartel
          // arriba de todo obliga a buscar; un error en el campo, no.
          if (failure.code === 'USERNAME_TAKEN') {
            return [
              {
                fieldTree: this.registration.profile.username,
                kind: failure.code,
                message: failure.message,
              },
            ];
          }
          this._failure.set(failure.message);
          return undefined;
        }
      },
    });
  }
}
