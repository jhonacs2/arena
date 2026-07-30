import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';
import { takeUntilDestroyed, toSignal } from '@angular/core/rxjs-interop';
import { FormsModule } from '@angular/forms';
import { Subject, catchError, debounceTime, distinctUntilChanged, of, switchMap, tap } from 'rxjs';

import { CatalogService, type Coffee } from './catalog.service';

/** Los tres estados de algo que tarda. Vuelven, iguales, en S7. */
type Status = 'idle' | 'loading' | 'error';

/**
 * S6 · El buscador.
 *
 * Es la primera vez que los datos **no están**: llegan. Y con eso aparecen
 * cuatro problemas que ninguna sesión anterior tuvo:
 *
 *   1. Una búsqueda por cada tecla        → `debounceTime`
 *   2. La misma búsqueda dos veces        → `distinctUntilChanged`
 *   3. Respuestas que llegan desordenadas → `switchMap`
 *   4. Suscripciones que nadie corta      → `takeUntilDestroyed`
 *
 * Los cuatro se resuelven con una línea cada uno, y los cuatro están abajo.
 *
 * **Signals y observables no compiten.** Un signal guarda un valor; un
 * observable describe algo que pasa a lo largo del tiempo. Aquí el tiempo lo
 * maneja RxJS y el resultado se guarda en un signal con `toSignal`, que es el
 * puente entre los dos mundos.
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

  /** Lo que el usuario escribe. Un Subject es un observable al que se le empuja. */
  private readonly terms = new Subject<string>();

  protected readonly status = signal<Status>('idle');
  protected readonly requests = signal(0);

  protected query = '';

  /**
   * El flujo entero, leído de arriba abajo:
   *
   *   escribe → espera → ¿cambió? → busca → guarda
   */
  private readonly results$ = this.terms.pipe(
    // 1 · Espera a que deje de escribir. Sin esto, una búsqueda por tecla.
    debounceTime(300),

    // 2 · Si el texto es el mismo de antes, no vuelve a buscar.
    distinctUntilChanged(),

    tap(() => this.status.set('loading')),

    // 3 · switchMap CANCELA la búsqueda anterior si llega una nueva.
    //     Con mergeMap llegarían las dos, y podría ganar la vieja.
    switchMap((term) =>
      this.catalog.searchCounted(term).pipe(
        tap(() => this.status.set('idle')),
        catchError(() => {
          this.status.set('error');
          // Devolver un observable mantiene vivo el flujo: sin esto, el error
          // mata la suscripción y el buscador deja de funcionar para siempre.
          return of([] as readonly Coffee[]);
        }),
      ),
    ),

    tap(() => this.requests.set(this.catalog.requests)),

    // Sin esto, el flujo sigue vivo después de que el componente se destruye.
    takeUntilDestroyed(),
  );

  /**
   * El puente: un observable entra, un signal sale.
   *
   * `initialValue` evita que el tipo sea `… | undefined`, que es lo que pasa
   * cuando el observable todavía no emitió nada.
   */
  protected readonly results = toSignal(this.results$, {
    initialValue: [] as readonly Coffee[],
  });

  protected readonly count = computed(() => this.results().length);

  /** Se llama desde `(input)` del campo. */
  protected onType(term: string): void {
    this.query = term;
    this.terms.next(term);
  }

  protected reset(): void {
    this.query = '';
    this.catalog.resetRequests();
    this.requests.set(0);
    this.status.set('idle');
    this.terms.next('');
  }
}
