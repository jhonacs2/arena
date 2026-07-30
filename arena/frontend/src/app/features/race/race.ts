import { httpResource } from '@angular/common/http';
import {
  ChangeDetectionStrategy,
  Component,
  computed,
  inject,
  input,
  linkedSignal,
  signal,
} from '@angular/core';
import { takeUntilDestroyed, toObservable } from '@angular/core/rxjs-interop';
import { FormField, FormRoot, form, max, min, required, submit, validate } from '@angular/forms/signals';
import { RouterLink } from '@angular/router';
import { from } from 'rxjs';
import { switchMap, tap } from 'rxjs/operators';

import { api } from '../../core/api/api';
import { toApiError } from '../../core/api/api-error';
import { SessionStore } from '../../core/auth/session.store';
import { RaceService } from '../../core/data/race.service';
import type {
  Bet,
  Participant,
  PublicBet,
  RaceDetail,
  RaceEvent,
  RaceStatus,
  ResultEntry,
} from '../../core/models';
import { RaceChannel, provideRaceChannel } from '../../core/race/race-channel';
import { CoinsPipe } from '../../shared/format/coins.pipe';
import { OddsPipe, payoutOf } from '../../shared/format/odds.pipe';
import { WhenPipe } from '../../shared/format/when.pipe';
import { StatusBadge } from '../../shared/ui/badge/status-badge';
import { Button } from '../../shared/ui/button/button';
import { Callout } from '../../shared/ui/callout/callout';
import { Silk } from '../../shared/ui/silk/silk';
import { RaceTrack } from './race-track';

interface BetDraft {
  horseId: string;
  amount: number;
}

/**
 * Una carrera: la sala, los caballos con sus cuotas, la apuesta y el vivo.
 *
 * El estado llega de dos lados y eso es a propósito: `httpResource` trae la foto
 * al entrar y el socket trae lo que pasa después. Los dos escriben sobre el mismo
 * `linkedSignal`, que es exactamente para esto —un valor derivado de una fuente
 * que además se puede sobrescribir localmente— y evita el clásico «tengo el
 * estado en dos signals y se me desincronizan».
 */
@Component({
  selector: 'app-race',
  changeDetection: ChangeDetectionStrategy.OnPush,
  providers: [provideRaceChannel()],
  imports: [
    RouterLink,
    FormField,
    FormRoot,
    RaceTrack,
    StatusBadge,
    Button,
    Callout,
    Silk,
    CoinsPipe,
    OddsPipe,
    WhenPipe,
  ],
  templateUrl: './race.html',
  styleUrls: ['../../shared/ui/surface.css', '../../shared/ui/field/form-controls.css', './race.css'],
})
export class Race {
  /** Llega de la ruta por `withComponentInputBinding()`. */
  readonly id = input.required<string>();

  private readonly channel = inject(RaceChannel);
  private readonly races = inject(RaceService);
  protected readonly session = inject(SessionStore);

  protected readonly detail = httpResource<RaceDetail>(() => api(`/races/${this.id()}`));

  // ── Estado que el socket sobrescribe ──────────────────────────────────

  protected readonly status = linkedSignal<RaceDetail | undefined, RaceStatus>({
    source: () => this.detail.value(),
    computation: (detail, previous) => detail?.status ?? previous?.value ?? 'open',
  });

  protected readonly participants = linkedSignal<RaceDetail | undefined, readonly Participant[]>({
    source: () => this.detail.value(),
    computation: (detail) => detail?.participants ?? [],
  });

  protected readonly bets = linkedSignal<RaceDetail | undefined, readonly PublicBet[]>({
    source: () => this.detail.value(),
    computation: (detail) => detail?.bets ?? [],
  });

  protected readonly myBet = linkedSignal<RaceDetail | undefined, Bet | null>({
    source: () => this.detail.value(),
    computation: (detail) => detail?.myBet ?? null,
  });

  protected readonly results = linkedSignal<RaceDetail | undefined, readonly ResultEntry[] | null>({
    source: () => this.detail.value(),
    computation: (detail) => detail?.results ?? null,
  });

  protected readonly myPayout = linkedSignal<RaceDetail | undefined, number | null>({
    source: () => this.detail.value(),
    computation: (detail) => detail?.myPayout ?? null,
  });

  private readonly _positions = signal<Readonly<Record<string, number>>>({});
  protected readonly positions = this._positions.asReadonly();

  private readonly _failure = signal<string | null>(null);
  protected readonly failure = this._failure.asReadonly();

  // ── Derivados ─────────────────────────────────────────────────────────

  protected readonly horses = computed(() => this.detail.value()?.horses ?? []);
  protected readonly isOpen = computed(() => this.status() === 'open');
  protected readonly isRunning = computed(() => this.status() === 'running');
  protected readonly isFinished = computed(() => this.status() === 'finished');
  protected readonly canBet = computed(
    () => this.isOpen() && this.myBet() === null && !this.session.isAdmin(),
  );

  protected readonly winner = computed(() => this.results()?.[0] ?? null);
  protected readonly iWon = computed(() => {
    const bet = this.myBet();
    const first = this.winner();
    return bet !== null && first !== null && bet.horseId === first.horseId;
  });

  /**
   * La sala, ya resuelta: cada participante con lo que apostó.
   *
   * `horseName` queda en `null` mientras la carrera está `open` porque el
   * servidor no manda el caballo. Que se vea «caballo oculto» y no un hueco es
   * parte de la regla: se sabe que apostó, no a qué.
   */
  protected readonly room = computed(() => {
    const bets = this.bets();
    const horses = this.horses();
    const meId = this.session.user()?.id ?? null;
    const mine = this.myBet();

    return this.participants().map((participant) => {
      const bet = bets.find((candidate) => candidate.userId === participant.userId);
      const horse =
        bet?.horseId == null ? undefined : horses.find((candidate) => candidate.id === bet.horseId);
      // La fila propia sí muestra el caballo: el servidor lo tapa para el resto
      // de la sala, no para el dueño de la apuesta, que obviamente ya lo sabe.
      const isMine = meId !== null && participant.userId === meId;
      const own = isMine ? mine : null;
      return {
        userId: participant.userId,
        username: participant.username,
        amount: bet?.amount ?? own?.amount ?? null,
        horseName: horse?.name ?? own?.horseName ?? null,
        isMine,
      };
    });
  });

  protected readonly loadError = computed(() =>
    this.detail.status() === 'error' ? toApiError(this.detail.error()).message : null,
  );

  // ── El formulario de apuesta ──────────────────────────────────────────

  /**
   * El monto arranca en 100 —una moneda de un punto— o en el saldo si es menos,
   * y desde ahí el alumno lo cambia. `linkedSignal` es lo que permite las dos
   * cosas: que se recalcule cuando cambia el saldo y que se pueda escribir.
   */
  private readonly draft = linkedSignal<number, BetDraft>({
    source: () => this.session.balance(),
    computation: (balance, previous) => ({
      horseId: previous?.value.horseId ?? '',
      amount: Math.min(100, balance),
    }),
  });

  protected readonly betForm = form(this.draft, (path) => {
    required(path.horseId, { message: 'Elegí un caballo.' });
    min(path.amount, 1, { message: 'El mínimo es 1 moneda.' });
    // El tope es una función, no un número: el saldo cambia —una carrera que
    // termina, un regalo del instructor— y el validador tiene que seguirlo.
    max(path.amount, () => this.session.balance(), {
      message: 'No podés apostar más monedas de las que tenés.',
    });
    validate(path.amount, ({ value }) =>
      Number.isInteger(value()) ? undefined : { kind: 'integer', message: 'Las monedas son enteras.' },
    );
  });

  /**
   * Lo que pagaría la apuesta que está escrita ahora mismo.
   *
   * `null` si el monto no es apostable. Mostrar «cobrás 339.996» debajo de un
   * monto que supera el saldo es prometer algo que el servidor va a rechazar.
   */
  protected readonly preview = computed(() => {
    const { horseId, amount } = this.draft();
    const horse = this.horses().find((candidate) => candidate.id === horseId);
    const usable =
      Number.isInteger(amount) && amount >= 1 && amount <= this.session.balance();
    if (horse === undefined || !usable) return null;
    return { horse, payout: payoutOf(amount, horse.odds) };
  });

  constructor() {
    // Entrar a la sala y quedarse escuchando. `switchMap` sobre el id: si se
    // navega a otra carrera, la suscripción anterior se cierra sola.
    toObservable(this.id)
      .pipe(
        tap(() => this.reset()),
        switchMap((raceId) =>
          from(this.races.join(raceId).catch(() => null)).pipe(
            switchMap(() => this.channel.connect(raceId)),
          ),
        ),
        takeUntilDestroyed(),
      )
      .subscribe((event) => this.apply(event));
  }

  protected chooseHorse(horseId: string): void {
    this.draft.update((draft) => ({ ...draft, horseId }));
    this.betForm.horseId().markAsTouched();
  }

  protected async send(): Promise<void> {
    this._failure.set(null);

    await submit(this.betForm, {
      action: async () => {
        const { horseId, amount } = this.draft();
        try {
          const response = await this.races.placeBet(this.id(), { horseId, amount });
          this.myBet.set(response.bet);
        } catch (cause) {
          // Los errores que importan son de servidor: la carrera se cerró justo,
          // o alguien apostó desde otra pestaña. El botón deshabilitado no los
          // previene, por eso se muestran acá tal cual vienen.
          this._failure.set(toApiError(cause).message);
        }
        return undefined;
      },
    });
  }

  private reset(): void {
    this._positions.set({});
    this._failure.set(null);
  }

  /** Un evento del socket. Es el único lugar donde el vivo escribe estado. */
  private apply(event: RaceEvent): void {
    switch (event.type) {
      case 'room.state':
        this.status.set(event.status);
        this.participants.set(event.participants);
        this.bets.set(event.bets);
        break;

      case 'room.joined':
        this.participants.update((current) =>
          current.some((participant) => participant.userId === event.userId)
            ? current
            : [...current, { userId: event.userId, username: event.username }],
        );
        break;

      case 'bet.placed':
        this.bets.update((current) =>
          current.some((bet) => bet.userId === event.userId)
            ? current
            : [
                ...current,
                {
                  userId: event.userId,
                  username: event.username,
                  amount: event.amount,
                  horseId: event.horseId,
                },
              ],
        );
        break;

      case 'race.started':
        this.status.set('running');
        // Al largar se revelan todas las apuestas juntas.
        this.bets.set(event.bets);
        break;

      case 'race.tick': {
        // Objeto nuevo en cada tick: nunca se muta el estado en su lugar.
        const next: Record<string, number> = {};
        for (const position of event.positions) next[position.horseId] = position.progress;
        this._positions.set(next);
        break;
      }

      case 'race.finished':
        this.status.set('finished');
        this.results.set(event.results);
        this.myPayout.set(event.payout);
        this.session.setBalance(event.balance);
        break;

      case 'race.cancelled':
        this.status.set('cancelled');
        this._failure.set(event.reason);
        // La devolución la hizo el servidor; el saldo se vuelve a leer al salir.
        break;
    }
  }
}
