import { ChangeDetectionStrategy, Component, computed, input } from '@angular/core';

import type { Horse, RaceStatus, ResultEntry } from '../../core/models';
import { OddsPipe } from '../../shared/format/odds.pipe';
import { Silk } from '../../shared/ui/silk/silk';

/**
 * La pista.
 *
 * Dibuja lo que llega por el socket y nada más: **la simulación es autoritativa
 * del servidor** (`decisiones.md` §4). Este componente no adivina posiciones ni
 * interpola por su cuenta; el único suavizado es una transición CSS de 100 ms,
 * que es exactamente lo que dura un tick a 10 Hz.
 *
 * Se mueve **solo con `transform`**. A 10 Hz y con seis caballos, animar `left` o
 * `width` serían 60 reflows por segundo: el navegador tendría que recalcular el
 * layout de la página entera sesenta veces, y en una máquina de aula eso se ve.
 * `transform` se resuelve en el compositor y no toca el layout.
 */
@Component({
  selector: 'app-race-track',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Silk, OddsPipe],
  templateUrl: './race-track.html',
  styleUrl: './race-track.css',
})
export class RaceTrack {
  readonly horses = input.required<readonly Horse[]>();
  /** Avance por caballo, 0 en la largada y 1 en el disco. */
  readonly positions = input<Readonly<Record<string, number>>>({});
  readonly status = input.required<RaceStatus>();
  readonly results = input<readonly ResultEntry[] | null>(null);
  /** El caballo al que apostó quien mira. Se marca en su andarivel. */
  readonly myHorseId = input<string | null>(null);

  protected readonly running = computed(() => this.status() === 'running');

  /**
   * Dónde dibujar cada caballo.
   *
   * Normalmente son los ticks. Pero a una carrera terminada se puede entrar desde
   * el tablero **sin haberla visto correr**: no hay ticks, y todos los caballos
   * quedarían pegados a la largada con un «1°» al lado, que se lee como un error.
   * Cuando no hay ticks y hay resultados, la posición se deriva del puesto.
   */
  private readonly drawn = computed<Readonly<Record<string, number>>>(() => {
    const live = this.positions();
    if (Object.keys(live).length > 0) return live;

    const results = this.results();
    if (results === null) return {};

    const derived: Record<string, number> = {};
    for (const entry of results) {
      // El ganador en el disco y cada puesto un escalón atrás. Es una
      // reconstrucción, no el desarrollo real — pero el orden sí es el real.
      derived[entry.horseId] = Math.max(0.12, 1 - (entry.position - 1) * 0.08);
    }
    return derived;
  });

  /** El orden actual. Es el marcador visible, y se redibuja con cada tick. */
  protected readonly standings = computed(() => {
    const positions = this.drawn();
    return [...this.horses()]
      .map((horse) => ({ horse, progress: positions[horse.id] ?? 0 }))
      .sort((a, b) => b.progress - a.progress);
  });

  protected readonly leader = computed(() => this.standings()[0]?.horse.name ?? null);

  /**
   * El texto que se anuncia por voz.
   *
   * **Solo el puntero, no el detalle.** El marcador cambia diez veces por
   * segundo: si se anunciara cada tick, un lector de pantalla quedaría leyendo
   * números para siempre y no se entendería nada. Este `computed` solo cambia de
   * valor cuando cambia el que va adelante, así que `aria-live` habla cuando pasa
   * algo.
   */
  protected readonly announcement = computed(() => {
    if (this.status() === 'finished') {
      const winner = this.results()?.[0]?.horseName;
      return winner === undefined ? 'La carrera terminó.' : `Ganó ${winner}.`;
    }
    if (!this.running()) return '';
    const leader = this.leader();
    return leader === null ? '' : `Va adelante ${leader}.`;
  });

  protected progressOf(horseId: string): number {
    return this.drawn()[horseId] ?? 0;
  }

  protected placeOf(horseId: string): number | null {
    return this.results()?.find((entry) => entry.horseId === horseId)?.position ?? null;
  }
}
