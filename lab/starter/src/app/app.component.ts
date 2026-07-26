import { ChangeDetectionStrategy, Component } from '@angular/core';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';

import { SESIONES } from './sesiones';

/**
 * El armazón del lab.
 *
 * Acá no se practica nada: es solo la navegación entre sesiones. El ejercicio
 * de cada clase vive en `sesiones/sNN/`.
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
  protected readonly sesiones = SESIONES;
}
