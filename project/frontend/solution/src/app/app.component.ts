import { ChangeDetectionStrategy, Component } from '@angular/core';

import { ShellComponent } from './layout/shell.component';

/**
 * Raíz de la aplicación. Solo monta el armazón: todo lo demás cuelga del
 * `router-outlet` que vive adentro de `<app-shell>`.
 */
@Component({
  selector: 'app-root',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [ShellComponent],
  template: '<app-shell />',
})
export class AppComponent {}
