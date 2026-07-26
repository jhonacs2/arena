import { ChangeDetectionStrategy, Component } from '@angular/core';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';

import { SESIONES_DISPONIBLES } from './sesiones';

/**
 * El armazón del lab: la barra lateral y el `router-outlet`.
 *
 * La barra muestra **solo las sesiones que ya hiciste**. Arranca vacía y crece
 * con vos: cada vez que habilitás una en `sesiones.ts` y le sumás su ruta,
 * aparece acá.
 */
@Component({
  selector: 'app-root',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [RouterOutlet, RouterLink, RouterLinkActive],
  templateUrl: './app.component.html',
  styleUrl: './app.component.css',
})
export class AppComponent {
  protected readonly sesiones = SESIONES_DISPONIBLES;
}
