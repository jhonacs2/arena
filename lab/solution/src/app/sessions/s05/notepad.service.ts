import { Injectable, signal } from '@angular/core';

/**
 * S5 · El cuaderno de cada mostrador.
 *
 * **No lleva `providedIn`.** Se provee en el componente que lo usa, con
 * `providers: [NotepadService]`, y eso cambia todo:
 *
 *   providedIn: 'root'        → UNA instancia para toda la aplicación
 *   providers: [] en un       → UNA instancia POR CADA copia de ese
 *   componente                  componente en pantalla
 *
 * Aquí es lo que queremos: cada mostrador anota lo suyo y no ve lo del otro.
 * Un cuaderno compartido entre dos cajas sería un error, no una comodidad.
 *
 * La pregunta que decide dónde va un servicio es siempre la misma:
 *
 *   **¿Cuántos de estos tiene que haber?**
 */
@Injectable()
export class NotepadService {
  private readonly _notes = signal<readonly string[]>([]);

  readonly notes = this._notes.asReadonly();

  write(note: string): void {
    const text = note.trim();
    if (text === '') return;

    this._notes.update((notes) => [...notes, text]);
  }

  erase(): void {
    this._notes.set([]);
  }
}
