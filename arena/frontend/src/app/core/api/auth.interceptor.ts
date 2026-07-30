import { HttpErrorResponse, type HttpInterceptorFn, type HttpRequest } from '@angular/common/http';
import { inject } from '@angular/core';
import { throwError } from 'rxjs';
import { catchError, switchMap } from 'rxjs/operators';

import { AuthService } from '../auth/auth.service';
import { SessionStore } from '../auth/session.store';
import { SKIP_AUTH } from './http-context';

const withBearer = <T>(request: HttpRequest<T>, token: string): HttpRequest<T> =>
  request.clone({ setHeaders: { Authorization: `Bearer ${token}` } });

/**
 * Pone el token en cada request y renueva la sesión cuando vence.
 *
 * El access token dura 15 minutos, así que en una clase de dos horas vence ocho
 * veces. Que el alumno tenga que volver a entrar en medio de una carrera no es
 * una opción: acá se reintenta una vez con el token nuevo y no se nota.
 *
 * **Un solo reintento.** Si el 401 vuelve con el token fresco, el problema no es
 * el token y reintentar en bucle solo esconde el error real.
 */
export const authInterceptor: HttpInterceptorFn = (request, next) => {
  if (request.context.get(SKIP_AUTH)) return next(request);

  const session = inject(SessionStore);
  const auth = inject(AuthService);

  const token = session.accessToken();
  const authorized = token === null ? request : withBearer(request, token);

  return next(authorized).pipe(
    catchError((error: unknown) => {
      const isExpired = error instanceof HttpErrorResponse && error.status === 401;
      if (!isExpired || token === null) return throwError(() => error);

      return auth.refresh().pipe(
        switchMap((renewed) =>
          renewed === null ? throwError(() => error) : next(withBearer(request, renewed)),
        ),
      );
    }),
  );
};
