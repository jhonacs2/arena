import { Injectable, computed, inject, signal } from '@angular/core';

import { MIN_BET_AMOUNT, MAX_BET_AMOUNT, potentialPayout } from '../models';
import { RaceStore } from '../races/race.store';

/**
 * S5 · El simulador de apuesta.
 *
 * Este store **depende de otro**, y es lo que hay que mirar: pide `RaceStore`
 * con `inject()` igual que un componente. Un servicio puede inyectar servicios;
 * el inyector no distingue quién pregunta.
 *
 * Con eso, el pago posible se deriva de dos cosas que viven en lugares
 * distintos —el monto, que es de aquí, y el favorito de la carrera abierta, que
 * es del otro store— sin que nadie tenga que sincronizarlas.
 *
 * En S8 este mismo store va a validar el saldo, y en S7 va a mandar la apuesta
 * de verdad. Hoy solo calcula.
 */
@Injectable({ providedIn: 'root' })
export class BetStore {
  private readonly races = inject(RaceStore);

  private readonly _amount = signal(100);

  readonly amount = this._amount.asReadonly();

  /** El caballo al que se le apostaría: el favorito de la carrera abierta. */
  readonly target = computed(() => this.races.selected()?.favourite);

  readonly payout = computed(() => potentialPayout(this._amount(), this.target()?.odds ?? 0));

  /** Los límites vienen del contrato, no de aquí. */
  readonly isValid = computed(
    () => this._amount() >= MIN_BET_AMOUNT && this._amount() <= MAX_BET_AMOUNT,
  );

  setAmount(amount: number): void {
    this._amount.set(amount);
  }
}
