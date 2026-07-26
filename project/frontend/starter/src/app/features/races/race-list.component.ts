import { ChangeDetectionStrategy, Component } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { RACES } from '../../core/mocks';
import { favourite, potentialPayout, type Horse, type Race } from '../../core/models';
import { SilkComponent } from '../../shared/ui/silk/silk.component';

/**
 * S1 · Listado de carreras.
 *
 * Los datos ya están: `RACES` sale de `core/mocks` y tiene las 8 carreras con
 * sus 54 caballos, preparados en `carreras`.
 *
 * Buscá `TODO(S1)` para encontrar tu trabajo, acá y en el HTML. El CSS ya
 * está listo: no hace falta que lo toques.
 */

/** Lo que la vista necesita, ya preparado. */
interface CarreraVista {
  readonly carrera: Race;
  readonly hora: string;
  readonly favorito: Horse | undefined;
  readonly etiquetaEstado: string;
}

const ETIQUETAS: Record<Race['status'], string> = {
  upcoming: 'Por largar',
  live: 'En vivo',
  finished: 'Terminada',
};

@Component({
  selector: 'app-race-list',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [FormsModule, SilkComponent],
  templateUrl: './race-list.component.html',
  styleUrl: './race-list.component.css',
})
export class RaceListComponent {
  /**
   * Ya está hecho. Los datos se preparan UNA vez, acá en la clase, y no en el
   * template: si lo calculás en el HTML, se recalcula en cada detección de
   * cambios.
   */
  protected readonly carreras: readonly CarreraVista[] = RACES.map((carrera) => ({
    carrera,
    hora: this.formatearHora(carrera.startsAt),
    favorito: favourite(carrera),
    etiquetaEstado: ETIQUETAS[carrera.status],
  }));

  /** La carrera que el usuario tocó. `null` = ninguna. */
  protected seleccionada: CarreraVista | null = null;

  /** Lo que se escribe en el simulador. Va y viene con `[(ngModel)]`. */
  protected monto = 100;

  /**
   * TODO(S1): esto siempre devuelve 0 porque la cuota está clavada en 0.
   *
   * Reemplazá ese `0` por la cuota del favorito de la carrera seleccionada.
   * Ojo: puede no haber ninguna carrera seleccionada.
   */
  protected get pagoPotencial(): number {
    return potentialPayout(this.monto, 0);
  }

  /**
   * TODO(S1): funciona, pero a medias.
   *
   * Ahora selecciona siempre. Hacé que tocar la MISMA carrera dos veces la
   * deseleccione (`seleccionada = null`): es lo que espera cualquiera y
   * ahorra un botón de cerrar.
   */
  protected seleccionar(vista: CarreraVista): void {
    this.seleccionada = vista;
  }

  /**
   * Ya está hecho. `Intl.DateTimeFormat` es del navegador, no de Angular: los
   * pipes de Angular llegan en S4.
   */
  private formatearHora(iso: string): string {
    return new Intl.DateTimeFormat('es', {
      day: '2-digit',
      month: 'short',
      hour: '2-digit',
      minute: '2-digit',
    }).format(new Date(iso));
  }
}
