import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { CatalogService, type Coffee } from './catalog.service';

/** Los tres estados de algo que tarda. Vuelven, iguales, en S7. */
type Status = 'idle' | 'loading' | 'error';

/**
 * S6 · El buscador, suscrito a mano.
 *
 * Esto funciona… si escribes despacio. Con cuatro problemas encima, y los
 * cuatro se ven en la pantalla:
 *
 *   1. Sale **una búsqueda por cada tecla**. Mira el contador.
 *   2. Escribir lo mismo dos veces vuelve a buscar.
 *   3. Si una respuesta vieja llega tarde, **pisa a la nueva**. Escribe una
 *      letra sola, espera un poco, y escribe algo más largo.
 *   4. Nadie corta la suscripción cuando el componente se va de la pantalla.
 *
 * Los cuatro se resuelven con una línea cada uno, y son la clase de hoy.
 *
 * Los lugares que hay que tocar están marcados con `TODO(S6)`.
 */
@Component({
  selector: 'app-s06',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [FormsModule],
  templateUrl: './s06.component.html',
  styleUrl: './s06.component.css',
})
export class S06Component {
  private readonly catalog = inject(CatalogService);

  protected readonly status = signal<Status>('idle');
  protected readonly requests = signal(0);
  protected readonly results = signal<readonly Coffee[]>([]);
  protected readonly count = computed(() => this.results().length);

  protected query = '';

  /**
   * TODO(S6) · 2, 3 y 4 — Aquí está todo el problema.
   *
   * Cada tecla llama a este método, y cada llamada se suscribe. No hay nada
   * que espere, nada que compare con lo anterior, nada que cancele la búsqueda
   * vieja y nada que corte la suscripción al destruirse el componente.
   *
   * El camino es reemplazar esto por **un solo flujo**: un `Subject` al que se
   * le empuja cada texto, y una tubería de operadores que resuelva los cuatro.
   */
  protected onType(term: string): void {
    this.query = term;
    this.status.set('loading');

    this.catalog.searchCounted(term).subscribe({
      next: (coffees) => {
        this.results.set(coffees);
        this.requests.set(this.catalog.requests);
        this.status.set('idle');
      },
      error: () => {
        this.status.set('error');
        this.requests.set(this.catalog.requests);
      },
    });
  }

  protected reset(): void {
    this.query = '';
    this.catalog.resetRequests();
    this.requests.set(0);
    this.status.set('idle');
    this.results.set([]);
  }
}
