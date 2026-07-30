import { registerLocaleData } from '@angular/common';
import localeEs from '@angular/common/locales/es';
import { ApplicationConfig, LOCALE_ID, provideZoneChangeDetection } from '@angular/core';
import { provideRouter, withComponentInputBinding, withInMemoryScrolling } from '@angular/router';

import { routes } from './app.routes';

/**
 * Configuración de la aplicación. Cero NgModule: `bootstrapApplication` más
 * esta lista de providers es todo el arranque (`project/frontend/CLAUDE.md`).
 */
/**
 * Los pipes incorporados formatean según el idioma de la aplicación, y ese
 * idioma por defecto es `en-US`. Sin estas dos líneas, un porcentaje sale
 * `54.5%` en una aplicación escrita entera en español. Se ve en S4.
 */
registerLocaleData(localeEs);

export const appConfig: ApplicationConfig = {
  providers: [
    { provide: LOCALE_ID, useValue: 'es' },
    // El proyecto corre CON zone.js. En Angular 18 la alternativa sin zona se
    // llama `provideExperimentalZonelessChangeDetection` y no la usamos:
    // CLAUDE.md §5. `eventCoalescing` junta varios eventos del mismo tick en
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
