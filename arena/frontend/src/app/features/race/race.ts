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
import { OddsPipe } from '../../shared/format/odds.pipe';
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

  /**
   * El pago propio, ya liquidado. Sale de la apuesta —`myBet.payout`— y no de un
   * campo aparte del detalle: dos números que representan lo mismo se
   * desincronizan siempre.
   *
   * Es `linkedSignal` y no `computed` porque tiene dos fuentes legítimas: la
   * apuesta que vino en el detalle, y el `race.finished` del socket, que llega
   * antes de que nadie vuelva a pedir el detalle.
   */
  protected readonly myPayout = linkedSignal<Bet | null, number | null>({
    source: () => this.myBet(),
    computation: (bet) => bet?.payout ?? null,
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

  /**
   * El caballo propio, resuelto contra la grilla.
   *
   * La apuesta trae `horseId` y nada más: el servidor no repite el nombre en cada
   * apuesta, y está bien que no lo haga —el nombre ya viaja una vez, en `horses`—.
   */
  protected readonly myHorse = computed(() => {
    const bet = this.myBet();
    if (bet === null) return null;
    return this.horses().find((horse) => horse.id === bet.horseId) ?? null;
  });

  /**
   * La carrera se liquidó devolviendo todo: nadie le pegó al ganador.
   *
   * Es un desenlace propio y no «perdiste»: con pari-mutuel, si el pozo ganador
   * es cero no hay entre quiénes repartir, y el pozo vuelve a sus dueños. Pasa
   * seguido cuando apuesta poca gente.
   */
  protected readonly wasRefunded = computed(() => this.myBet()?.status === 'refunded');

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
      const ownHorse =
        own === null ? undefined : horses.find((candidate) => candidate.id === own.horseId);
      return {
        userId: participant.userId,
        username: participant.username,
        amount: bet?.amount ?? own?.amount ?? null,
        horseName: horse?.name ?? ownHorse?.name ?? null,
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
   * El pozo, contando la apuesta que está escrita ahora mismo.
   *
   * **No es un pago potencial, y no puede serlo.** Con pari-mutuel lo que se
   * cobra depende del pozo final y de cuántos acertaron; las dos cosas siguen
   * cambiando hasta que se cierran las apuestas, y además el servidor tapa a qué
   * caballo apostó cada uno mientras la carrera está `open` —así que ni siquiera
   * podríamos estimarlo—. Un número acá sería una promesa que nadie puede cumplir.
   *
   * `null` si el monto no es apostable: no tiene sentido mostrar un pozo que
   * incluye una apuesta que el servidor va a rechazar.
   */
  protected readonly preview = computed(() => {
    const { horseId, amount } = this.draft();
    const horse = this.horses().find((candidate) => candidate.id === horseId);
    const usable = Number.isInteger(amount) && amount >= 1 && amount <= this.session.balance();
    if (horse === undefined || !usable) return null;

    const pool = this.bets().reduce((sum, bet) => sum + bet.amount, amount);
    return { horse, pool };
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
      // La sala viene ANIDADA en `race`, y es el mismo RaceDetail del GET. Leerla
      // como si viniera plana metía `undefined` en los signals y el template se
      // caía entero al pedirle `.length`.
      case 'room.state':
        this.status.set(event.race.status);
        this.participants.set(event.race.participants);
        this.bets.set(event.race.bets);
        this.myBet.set(event.race.myBet);
        if (event.race.results !== null) this.results.set(event.race.results);
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
                  // El evento NO trae el caballo mientras la carrera está abierta,
                  // y el `race.started` revela todas juntas. Hasta entonces se
                  // muestra que apostó, no a qué.
                  horseId: null,
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
        this.bets.set(event.bets);
        if (event.myBet !== null) {
          this.myBet.set(event.myBet);
          this.myPayout.set(event.myBet.payout);
        }
        // `null` significa «tu saldo no cambió», no «tu saldo es cero».
        if (event.balance !== null) this.session.setBalance(event.balance);
        break;

      case 'race.cancelled':
        this.status.set('cancelled');
        this._failure.set(event.reason);
        if (event.balance !== null) this.session.setBalance(event.balance);
        break;
    }
  }
}
