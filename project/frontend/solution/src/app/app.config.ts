import { ApplicationConfig, provideZoneChangeDetection } from '@angular/core';
import { provideRouter, withComponentInputBinding, withInMemoryScrolling } from '@angular/router';

import { routes } from './app.routes';

/**
 * Configuración de la aplicación. Cero NgModule: `bootstrapApplication` más
 * esta lista de providers es todo el arranque (`CLAUDE.md` §5).
 */
export const appConfig: ApplicationConfig = {
  providers: [
    // El proyecto corre CON zone.js. En Angular 18 la alternativa sin zona se
    // llama `provideExperimentalZonelessChangeDetection` y no la usamos:
    // CLAUDE.md §4. `eventCoalescing` junta varios eventos del mismo tick en
    // una sola detección de cambios.
    provideZoneChangeDetection({ eventCoalescing: true }),

    provideRouter(
      routes,
      // Hace que los parámetros de ruta lleguen como `input()`, sin inyectar
      // ActivatedRoute. Es lo que se enseña en S9 y por eso está desde ahora.
      withComponentInputBinding(),
      withInMemoryScrolling({ scrollPositionRestoration: 'enabled', anchorScrolling: 'enabled' }),
    ),
  ],
};
