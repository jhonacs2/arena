import { bootstrapApplication } from '@angular/platform-browser';

import { App } from './app/app';
import { appConfig } from './app/app.config';

/**
 * Sin `.catch(console.error)`.
 *
 * Si el arranque falla, la promesa rechaza y `provideBrowserGlobalErrorListeners()`
 * lo reporta con el stack entero. Atraparlo acá para imprimirlo a mano solo cambia
 * dónde aparece el mismo error, y deja un `console` en el bundle.
 */
void bootstrapApplication(App, appConfig);
