import { bootstrapApplication } from '@angular/platform-browser';

import { AppComponent } from './app/app.component';
import { appConfig } from './app/app.config';

bootstrapApplication(AppComponent, appConfig).catch((error: unknown) => {
  // Sin este catch, un error de arranque queda como una promesa rechazada sin
  // manejar y la pantalla se ve en blanco, sin ninguna pista de por qué.
  console.error('No se pudo iniciar la aplicación', error);
});
