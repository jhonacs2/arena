import { registerLocaleData } from '@angular/common';
import localeEs from '@angular/common/locales/es';
import { ApplicationConfig, LOCALE_ID, provideZoneChangeDetection } from '@angular/core';
import { provideRouter, withComponentInputBinding } from '@angular/router';

import { routes } from './app.routes';

/**
 * Los pipes incorporados —`date`, `number`, `percent`, `currency`— formatean
 * según el idioma de la aplicación, y ese idioma **por defecto es `en-US`**.
 *
 * Sin estas dos líneas, `{{ 0.545 | percent }}` sale `54.5%` en una pantalla
 * escrita entera en español. Aparece en S4, cuando entran los pipes.
 *
 * Son dos pasos y hacen falta los dos: registrar los datos del idioma, y
 * decirle a Angular cuál usar.
 */
registerLocaleData(localeEs);

export const appConfig: ApplicationConfig = {
  providers: [
    { provide: LOCALE_ID, useValue: 'es' },
    provideZoneChangeDetection({ eventCoalescing: true }),
    provideRouter(routes, withComponentInputBinding()),
  ],
};
