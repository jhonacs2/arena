import { Pipe, type PipeTransform } from '@angular/core';

/**
 * S4 · Dos pipes idénticos salvo por una palabra, para poder contarlos.
 *
 * Hacen exactamente lo mismo —devolver el texto tal cual— y llevan la cuenta
 * de cuántas veces los llamó Angular. Están en la pantalla para que la
 * diferencia entre puro e impuro se pueda **ver**, en lugar de creerla.
 *
 * En un proyecto de verdad esto no va: es material de clase.
 */

/** Los dos contadores, para que la pantalla pueda mostrarlos. */
export const CALL_COUNT = { pure: 0, impure: 0 };

export function resetCallCount(): void {
  CALL_COUNT.pure = 0;
  CALL_COUNT.impure = 0;
}

/**
 * PURO — el valor por defecto.
 *
 * Angular lo llama **solo cuando cambia el valor de entrada**. Si el valor es
 * el mismo, reutiliza el resultado anterior sin ejecutar nada.
 */
@Pipe({ name: 'countPure', standalone: true, pure: true })
export class CountPurePipe implements PipeTransform {
  transform(value: string): string {
    CALL_COUNT.pure += 1;
    return value;
  }
}

/**
 * IMPURO — `pure: false`.
 *
 * Angular lo llama **en cada detección de cambios**, haya cambiado algo o no.
 * Cada clic, cada tecla, cada temporizador de cualquier parte de la aplicación.
 *
 * Se usa cuando el resultado depende de algo que el valor de entrada no ve
 * —el reloj, por ejemplo—. Es caro, y casi siempre hay una forma mejor.
 */
@Pipe({ name: 'countImpure', standalone: true, pure: false })
export class CountImpurePipe implements PipeTransform {
  transform(value: string): string {
    CALL_COUNT.impure += 1;
    return value;
  }
}
