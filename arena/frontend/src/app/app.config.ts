import { provideHttpClient, withFetch, withInterceptors } from '@angular/common/http';
import {
  provideBrowserGlobalErrorListeners,
  type ApplicationConfig,
} from '@angular/core';
import { provideRouter, withComponentInputBinding, withInMemoryScrolling } from '@angular/router';

import { environment } from '../environments/environment';
import { authInterceptor } from './core/api/auth.interceptor';
import { mockBackendInterceptor } from './core/api/mock/mock-backend.interceptor';
import { routes } from './app.routes';

/**
 * El arranque.
 *
 * `withComponentInputBinding()` hace que `carreras/:id` llegue a la pantalla como
 * un `input.required<string>()` reactivo: cuando cambia el parámetro, el
 * `httpResource` de la carrera se recarga y el socket se reconecta sin una línea
 * de código de más.
 *
 * El orden de los interceptores importa: `authInterceptor` primero para que ponga
 * el token, y el mock **al final**, así ve la request completa y puede validar el
 * rol como lo haría el servidor.
 */
export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    provideRouter(
      routes,
      withComponentInputBinding(),
      withInMemoryScrolling({ scrollPositionRestoration: 'top' }),
    ),
    provideHttpClient(
      withFetch(),
      withInterceptors(
        environment.useMockBackend ? [authInterceptor, mockBackendInterceptor] : [authInterceptor],
      ),
    ),
  ],
};
