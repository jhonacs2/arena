import { Injectable } from '@angular/core';
import { Observable, delay, of, throwError } from 'rxjs';

export interface Coffee {
  readonly id: string;
  readonly name: string;
  readonly origin: string;
  readonly price: number;
}

const CATALOG: readonly Coffee[] = [
  { id: 'c1', name: 'Yirgacheffe', origin: 'Etiopía', price: 4200 },
  { id: 'c2', name: 'Huila', origin: 'Colombia', price: 3800 },
  { id: 'c3', name: 'Cerrado', origin: 'Brasil', price: 3000 },
  { id: 'c4', name: 'Antigua', origin: 'Guatemala', price: 4500 },
  { id: 'c5', name: 'Sidamo', origin: 'Etiopía', price: 4100 },
  { id: 'c6', name: 'Nariño', origin: 'Colombia', price: 3900 },
  { id: 'c7', name: 'Kiambu', origin: 'Kenia', price: 5200 },
  { id: 'c8', name: 'Tarrazú', origin: 'Costa Rica', price: 4800 },
];

/**
 * S6 · El catálogo, que ahora tarda.
 *
 * Hasta S5 los datos estaban ahí: una constante, y listo. Aquí aparece lo
 * único que todavía no tuvimos que manejar — **el tiempo**.
 *
 * `of(...)` crea un observable que emite un valor y termina. El `delay` finge
 * la red. En S7 esto va a ser `HttpClient` y no va a cambiar nada de lo que
 * escriban los componentes: **la forma es la misma**.
 *
 * Que la búsqueda tarde distinto según el texto no es un capricho: es lo que
 * hace que se vea el problema de las respuestas que llegan fuera de orden, que
 * es la mitad de la clase.
 */
@Injectable({ providedIn: 'root' })
export class CatalogService {
  /** Cuántas búsquedas salieron de verdad. La pantalla lo muestra. */
  private _requests = 0;

  get requests(): number {
    return this._requests;
  }

  resetRequests(): void {
    this._requests = 0;
  }

  /**
   * Busca cafés por nombre o por origen.
   *
   * **Un observable es frío:** esta función no busca nada. Devuelve una receta,
   * y la búsqueda ocurre cuando alguien se suscribe. Si nadie lo hace, el
   * contador de arriba no se mueve — y eso es el tercer «predice y ejecuta».
   */
  search(term: string): Observable<readonly Coffee[]> {
    const text = term.trim().toLowerCase();

    if (text === 'error') {
      return throwError(() => new Error('El catálogo no responde')).pipe(delay(300));
    }

    const results = CATALOG.filter((coffee) =>
      `${coffee.name} ${coffee.origin}`.toLowerCase().includes(text),
    );

    // Cuanto más corto el texto, más lenta la respuesta: así una búsqueda vieja
    // puede llegar DESPUÉS de una nueva.
    const latency = text.length <= 1 ? 1200 : 300;

    return of(results).pipe(
      // El contador se toca al suscribirse, no al llamar a `search`.
      delay(latency),
    );
  }

  /** Igual que `search`, pero contando las suscripciones de verdad. */
  searchCounted(term: string): Observable<readonly Coffee[]> {
    return new Observable<readonly Coffee[]>((subscriber) => {
      this._requests += 1;
      const inner = this.search(term).subscribe(subscriber);
      return () => inner.unsubscribe();
    });
  }
}
