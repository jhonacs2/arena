import { Injectable } from '@angular/core';
import { Subject } from 'rxjs';

import type {
  Bet,
  CreateCodesRequest,
  CreateRaceRequest,
  GiftRequest,
  GiftResponse,
  Horse,
  InviteCode,
  LedgerEntry,
  LedgerReason,
  LoginRequest,
  Me,
  Participant,
  PlaceBetRequest,
  PlaceBetResponse,
  PublicBet,
  RaceDetail,
  RaceEvent,
  RaceStatus,
  RaceSummary,
  RedeemRequest,
  ResultEntry,
  Role,
  RoomStateEvent,
  ScoreRow,
  Session,
  User,
} from '../../models';

/**
 * El backend de mentira.
 *
 * Existe por una razón concreta: el Go se está escribiendo en paralelo y el
 * frontend no puede quedarse esperando. Responde **las mismas formas y los
 * mismos códigos de error** que `arena/docs/contract/api.md`, así que un
 * componente escrito contra este mock funciona contra el backend real sin tocar
 * una línea. Se apaga con una línea en `environments/environment.ts`.
 *
 * Lo que este mock **no** es: un modelo de seguridad. Valida lo mismo que
 * validaría el servidor porque si no las pantallas no se pueden probar, pero la
 * validación que cuenta es la del handler en Go (`decisiones.md` §4).
 */

export class MockApiError extends Error {
  constructor(
    readonly status: number,
    readonly code: string,
    message: string,
  ) {
    super(message);
  }
}

interface MockUser {
  readonly id: string;
  readonly username: string;
  readonly firstName: string;
  readonly lastName: string;
  readonly role: Role;
  readonly password: string;
  balance: number;
}

interface MockBet {
  readonly id: string;
  readonly raceId: string;
  readonly userId: string;
  readonly horseId: string;
  readonly amount: number;
  readonly createdAt: string;
  status: 'placed' | 'won' | 'lost' | 'refunded';
  payout: number | null;
  settledAt: string | null;
}

interface MockRace {
  readonly id: string;
  name: string;
  status: RaceStatus;
  scheduledAt: string;
  horses: Horse[];
  readonly participants: Set<string>;
  seed: number;
  startedAt: string | null;
  results: ResultEntry[] | null;
  readonly progress: Map<string, number>;
}

interface MockCode {
  readonly code: string;
  readonly coinsGranted: number;
  readonly note: string | null;
  readonly createdAt: string;
  redeemedAt: string | null;
  redeemedBy: string | null;
}

interface MockLedgerEntry extends LedgerEntry {
  readonly userId: string;
}

/** Un evento del socket, con a quién va. `to === null` es para toda la sala. */
export interface MockRaceMessage {
  readonly raceId: string;
  readonly to: string | null;
  readonly event: RaceEvent;
}

/** Letras sin I, L, O, U. Un código se dicta en voz alta (`decisiones.md` §2). */
const CODE_LETTERS = 'ABCDEFGHJKMNPQRSTVWXYZ';
/** Dígitos sin 0 ni 1, por lo mismo. */
const CODE_DIGITS = '23456789';

const TICK_MS = 100;
const COINS_PER_POINT = 100;

/**
 * Puntos = `max(10, floor(saldo / 100))`. Una función del saldo, nunca una columna.
 *
 * **El piso de 10 no es cosmético** (`decisiones.md` §1): perder monedas saca
 * capacidad de seguir jugando, nunca calificación. Sin él, un alumno que apuesta
 * mal termina con menos nota que uno que no jugó, y el juego pasa a castigar el
 * haber participado.
 *
 * Los puntos regalados por el instructor se suman **por encima** de este piso; no
 * pasan por el juego y por eso viven aparte, en `point_grants`.
 */
const MIN_POINTS = 10;
const pointsOf = (balance: number): number =>
  Math.max(MIN_POINTS, Math.floor(balance / COINS_PER_POINT));

/**
 * Reparte el pozo entre los que acertaron — **pari-mutuel** (`decisiones.md` §1).
 *
 *     payout = amount * pool / winningPool        división entera, trunca
 *
 * El resto de esa división **se reparte**, no se descarta: si se descartara, el
 * curso perdería monedas que nadie perdió apostando y la suma dejaría de cerrar.
 * Va de a una moneda a los que más apostaron, y a igualdad de monto al más viejo.
 *
 * Es la misma regla, con el mismo desempate, que `races.PariMutuel` del backend.
 * Que estén escritas dos veces es el precio de tener un mock que sirva para
 * enseñar; que den distinto sería el bug que nadie encuentra hasta la clase.
 */
function settlePariMutuel(bets: readonly MockBet[], winnerId: string | null): Map<string, number> {
  const payouts = new Map<string, number>();
  const pool = bets.reduce((sum, bet) => sum + bet.amount, 0);
  const winners = bets.filter((bet) => bet.horseId === winnerId);
  const winningPool = winners.reduce((sum, bet) => sum + bet.amount, 0);

  // Nadie le pegó al ganador: se devuelve cada apuesta íntegra. Quedarse con el
  // pozo sería cobrarle al curso por una carrera que nadie ganó.
  if (winningPool === 0) {
    for (const bet of bets) payouts.set(bet.id, bet.amount);
    return payouts;
  }

  const ordered = [...winners].sort(
    (a, b) =>
      b.amount - a.amount || a.createdAt.localeCompare(b.createdAt) || a.id.localeCompare(b.id),
  );

  let truncated = 0;
  for (const bet of ordered) {
    const payout = Math.floor((bet.amount * pool) / winningPool);
    payouts.set(bet.id, payout);
    truncated += payout;
  }

  let remainder = pool - truncated;
  for (const bet of ordered) {
    if (remainder <= 0) break;
    payouts.set(bet.id, (payouts.get(bet.id) ?? 0) + 1);
    remainder--;
  }
  return payouts;
}

/** mulberry32 — PRNG determinístico. Misma semilla, misma carrera. */
function prng(seed: number): () => number {
  let state = seed >>> 0;
  return () => {
    state = (state + 0x6d2b79f5) >>> 0;
    let t = state;
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

@Injectable({ providedIn: 'root' })
export class MockWorld {
  private readonly users = new Map<string, MockUser>();
  private readonly codes = new Map<string, MockCode>();
  private readonly races = new Map<string, MockRace>();
  private readonly bets: MockBet[] = [];
  private readonly ledger: MockLedgerEntry[] = [];
  private readonly tokens = new Map<string, string>();
  private readonly refreshTokens = new Map<string, string>();

  private sequence = 1;
  private timer: ReturnType<typeof setInterval> | null = null;

  /** Lo que consume el canal de socket de mentira. */
  readonly messages = new Subject<MockRaceMessage>();

  constructor() {
    this.seed();
  }

  // ── Autenticación ───────────────────────────────────────────────────────

  checkCode(code: string): { valid: true; coinsGranted: number } {
    const found = this.requireCode(code);
    return { valid: true, coinsGranted: found.coinsGranted };
  }

  redeem(body: RedeemRequest): Session {
    const code = this.requireCode(body.code);

    if (body.password.length < 8) {
      throw new MockApiError(400, 'VALIDATION_FAILED', 'La contraseña necesita 8 caracteres o más.');
    }
    if (!body.firstName.trim() || !body.lastName.trim() || !body.username.trim()) {
      throw new MockApiError(400, 'VALIDATION_FAILED', 'Faltan datos para crear la cuenta.');
    }
    if (this.findByUsername(body.username) !== undefined) {
      throw new MockApiError(409, 'USERNAME_TAKEN', 'Ese usuario ya está ocupado.');
    }

    // Los tres pasos son UNA transacción: usuario + código quemado + monedas
    // acreditadas, o nada. Un código a medio canjear es el peor estado posible.
    const user: MockUser = {
      id: this.nextId('user'),
      username: body.username.trim(),
      firstName: body.firstName.trim(),
      lastName: body.lastName.trim(),
      role: 'student',
      password: body.password,
      balance: 0,
    };
    this.users.set(user.id, user);
    code.redeemedAt = new Date().toISOString();
    code.redeemedBy = user.username;
    this.credit(user, code.coinsGranted, 'code_redeemed', { note: `Código ${code.code}` });

    return this.sessionFor(user);
  }

  login(body: LoginRequest): Session {
    const user = this.findByUsername(body.username);
    if (user === undefined || user.password !== body.password) {
      throw new MockApiError(401, 'INVALID_CREDENTIALS', 'Usuario o contraseña incorrectos.');
    }
    return this.sessionFor(user);
  }

  refresh(): { accessToken: string } {
    // El refresh real vive en una cookie HttpOnly que el JS no ve. Acá se toma
    // el último emitido, que alcanza para probar el reintento del interceptor.
    const [pair] = [...this.refreshTokens.entries()].slice(-1);
    if (pair === undefined) {
      throw new MockApiError(401, 'UNAUTHENTICATED', 'Tu sesión venció. Iniciá sesión de nuevo.');
    }
    const [oldToken, userId] = pair;
    this.refreshTokens.delete(oldToken);
    const user = this.requireUser(userId);
    const accessToken = this.issue(user);
    return { accessToken };
  }

  logout(token: string | null): void {
    if (token !== null) this.tokens.delete(token);
  }

  me(token: string | null): Me {
    const user = this.authenticate(token);
    return { user: this.publicUser(user), balance: user.balance, points: pointsOf(user.balance) };
  }

  transactions(token: string | null): { items: LedgerEntry[] } {
    const user = this.authenticate(token);
    const items = this.ledger
      .filter((entry) => entry.userId === user.id)
      .map(({ userId: _userId, ...entry }) => entry)
      .reverse();
    return { items };
  }

  // ── Carreras, lado alumno ───────────────────────────────────────────────

  listRaces(token: string | null): { items: RaceSummary[] } {
    const user = this.authenticate(token);
    const visible: readonly RaceStatus[] =
      user.role === 'admin'
        ? ['draft', 'open', 'running', 'finished', 'cancelled']
        : ['open', 'running', 'finished', 'cancelled'];

    const items = [...this.races.values()]
      .filter((race) => visible.includes(race.status))
      .sort((a, b) => a.scheduledAt.localeCompare(b.scheduledAt))
      .map((race) => ({
        id: race.id,
        name: race.name,
        status: race.status,
        scheduledAt: race.scheduledAt,
        horseCount: race.horses.length,
        participantCount: race.participants.size,
        myBet: this.betOf(race, user.id),
      }));

    return { items };
  }

  raceDetail(token: string | null, raceId: string): RaceDetail {
    const user = this.authenticate(token);
    const race = this.requireRace(raceId);
    if (race.status === 'draft' && user.role !== 'admin') {
      throw new MockApiError(404, 'RACE_NOT_FOUND', 'Esa carrera no existe.');
    }
    return this.detailFor(race, user.id);
  }

  join(token: string | null, raceId: string): RaceDetail {
    const user = this.authenticate(token);
    const race = this.requireRace(raceId);
    const wasIn = race.participants.has(user.id);
    race.participants.add(user.id);

    if (!wasIn) {
      this.publish(race.id, null, {
        type: 'room.joined',
        raceId: race.id,
        userId: user.id,
        username: user.username,
        participantCount: race.participants.size,
      });
    }
    return this.detailFor(race, user.id);
  }

  placeBet(token: string | null, raceId: string, body: PlaceBetRequest): PlaceBetResponse {
    const user = this.authenticate(token);
    const race = this.requireRace(raceId);

    // Exactamente el orden en que lo validaría el handler en Go, dentro de una
    // transacción. Un botón deshabilitado en la pantalla es cortesía.
    if (user.role === 'admin') {
      throw new MockApiError(403, 'FORBIDDEN', 'El instructor no apuesta.');
    }
    if (race.status !== 'open') {
      throw new MockApiError(409, 'RACE_NOT_OPEN', 'La carrera ya no acepta apuestas.');
    }
    const horse = race.horses.find((candidate) => candidate.id === body.horseId);
    if (horse === undefined) {
      throw new MockApiError(400, 'VALIDATION_FAILED', 'Ese caballo no corre en esta carrera.');
    }
    if (this.bets.some((bet) => bet.raceId === race.id && bet.userId === user.id)) {
      throw new MockApiError(409, 'BET_ALREADY_PLACED', 'Ya apostaste en esta carrera.');
    }
    if (!Number.isInteger(body.amount) || body.amount < 1) {
      throw new MockApiError(400, 'VALIDATION_FAILED', 'El monto tiene que ser un entero de 1 o más.');
    }
    if (body.amount > user.balance) {
      throw new MockApiError(409, 'INSUFFICIENT_BALANCE', 'No te alcanzan las monedas.');
    }

    const bet: MockBet = {
      id: this.nextId('bet'),
      raceId: race.id,
      userId: user.id,
      horseId: horse.id,
      amount: body.amount,
      // No se congela ninguna cuota: con pari-mutuel no hay una cuota que
      // congelar. Lo que se cobra sale del pozo, y el pozo todavía está creciendo.
      createdAt: new Date().toISOString(),
      status: 'placed',
      payout: null,
      settledAt: null,
    };
    this.bets.push(bet);
    race.participants.add(user.id);
    this.credit(user, -body.amount, 'bet_placed', { raceName: race.name });

    this.publish(race.id, null, {
      type: 'bet.placed',
      raceId: race.id,
      userId: user.id,
      username: user.username,
      amount: bet.amount,
      // El caballo NO viaja: si se revelara, los últimos copiarían a los primeros
      // y la apuesta dejaría de medir criterio. Se revelan todas al largar.
      betCount: this.bets.filter((candidate) => candidate.raceId === race.id).length,
    });

    return { bet: this.toBet(bet), balance: user.balance };
  }

  /** Lo primero que recibe alguien que se conecta al socket: la sala completa. */
  roomState(userId: string, raceId: string): RoomStateEvent {
    this.requireUser(userId);
    const race = this.requireRace(raceId);
    // La sala va ANIDADA en `race`, y es el mismo RaceDetail del GET. Mandarla
    // plana es lo que rompió la pantalla de la carrera contra el backend real.
    return { type: 'room.state', raceId: race.id, race: this.detailFor(race, userId) };
  }

  /** El socket llega con el token en el query: acá se traduce a un usuario. */
  userIdFromToken(token: string | null): string | null {
    if (token === null) return null;
    const userId = this.tokens.get(token) ?? /^mock\.(.+)\.\d+$/.exec(token)?.[1];
    return userId !== undefined && this.users.has(userId) ? userId : null;
  }

  // ── Instructor ──────────────────────────────────────────────────────────

  createCodes(token: string | null, body: CreateCodesRequest): { codes: string[] } {
    this.requireAdmin(token);
    if (!Number.isInteger(body.count) || body.count < 1 || body.count > 200) {
      throw new MockApiError(400, 'VALIDATION_FAILED', 'La cantidad tiene que estar entre 1 y 200.');
    }
    if (!Number.isInteger(body.coinsGranted) || body.coinsGranted < 1) {
      throw new MockApiError(400, 'VALIDATION_FAILED', 'Las monedas tienen que ser un entero positivo.');
    }

    const created: string[] = [];
    while (created.length < body.count) {
      const code = this.randomCode();
      if (this.codes.has(code)) continue;
      this.codes.set(code, {
        code,
        coinsGranted: body.coinsGranted,
        note: body.note?.trim() || null,
        createdAt: new Date().toISOString(),
        redeemedAt: null,
        redeemedBy: null,
      });
      created.push(code);
    }
    return { codes: created };
  }

  listCodes(token: string | null): { items: InviteCode[] } {
    this.requireAdmin(token);
    const items = [...this.codes.values()]
      .sort((a, b) => b.createdAt.localeCompare(a.createdAt))
      .map((code) => ({ ...code }));
    return { items };
  }

  scores(token: string | null): { items: ScoreRow[] } {
    this.requireAdmin(token);
    const items = [...this.users.values()]
      .filter((user) => user.role === 'student')
      .map((user) => {
        const own = this.bets.filter((bet) => bet.userId === user.id);
        const won = own.filter((bet) => {
          const race = this.races.get(bet.raceId);
          return race?.results?.[0]?.horseId === bet.horseId;
        });
        return {
          userId: user.id,
          username: user.username,
          firstName: user.firstName,
          lastName: user.lastName,
          balance: user.balance,
          points: pointsOf(user.balance),
          betsPlaced: own.length,
          betsWon: won.length,
        };
      })
      .sort((a, b) => b.balance - a.balance);
    return { items };
  }

  gift(token: string | null, userId: string, body: GiftRequest): GiftResponse {
    const admin = this.requireAdmin(token);
    const target = this.requireUser(userId);
    if (!Number.isInteger(body.coins) || body.coins === 0) {
      throw new MockApiError(400, 'VALIDATION_FAILED', 'Poné un entero distinto de cero.');
    }
    // Piso en 0: el saldo nunca queda negativo (`decisiones.md` §1).
    if (target.balance + body.coins < 0) {
      throw new MockApiError(409, 'INSUFFICIENT_BALANCE', 'El ajuste dejaría el saldo en negativo.');
    }

    this.credit(target, body.coins, body.coins > 0 ? 'gift' : 'adjustment', {
      note: body.note?.trim() || `Por ${admin.username}`,
    });
    return { balance: target.balance, points: pointsOf(target.balance) };
  }

  createRace(token: string | null, body: CreateRaceRequest): { race: RaceDetail } {
    const admin = this.requireAdmin(token);
    if (!body.name.trim()) {
      throw new MockApiError(400, 'VALIDATION_FAILED', 'La carrera necesita un nombre.');
    }
    if (body.horses.length < 2) {
      throw new MockApiError(400, 'VALIDATION_FAILED', 'Una carrera necesita al menos dos caballos.');
    }
    if (body.horses.some((h) => !Number.isInteger(h.nominalOdds) || h.nominalOdds < 101)) {
      throw new MockApiError(400, 'VALIDATION_FAILED', 'Las cuotas van ×100 y tienen que ser mayores a 1,00.');
    }

    const race: MockRace = {
      id: this.nextId('race'),
      name: body.name.trim(),
      status: 'draft',
      scheduledAt: body.scheduledAt,
      horses: body.horses.map((horse, index) => ({
        id: `${this.nextId('horse')}`,
        number: horse.number || index + 1,
        name: horse.name.trim(),
        nominalOdds: horse.nominalOdds,
      })),
      participants: new Set<string>(),
      seed: 0,
      startedAt: null,
      results: null,
      progress: new Map<string, number>(),
    };
    this.races.set(race.id, race);
    return { race: this.detailFor(race, admin.id) };
  }

  openRace(token: string | null, raceId: string): { race: RaceDetail } {
    const admin = this.requireAdmin(token);
    const race = this.requireRace(raceId);
    if (race.status !== 'draft') {
      throw new MockApiError(409, 'INVALID_TRANSITION', 'Solo se abre una carrera en borrador.');
    }
    race.status = 'open';
    return { race: this.detailFor(race, admin.id) };
  }

  startRace(token: string | null, raceId: string): { race: RaceDetail } {
    const admin = this.requireAdmin(token);
    const race = this.requireRace(raceId);
    if (race.status !== 'open') {
      throw new MockApiError(409, 'INVALID_TRANSITION', 'Solo se larga una carrera abierta.');
    }

    // Cerrar las apuestas es lo primero, y pasa en el servidor.
    race.status = 'running';
    race.startedAt = new Date().toISOString();
    race.seed = Math.floor(Math.random() * 0xffffffff);
    for (const horse of race.horses) race.progress.set(horse.id, 0);

    this.publish(race.id, null, {
      type: 'race.started',
      raceId: race.id,
      startedAt: race.startedAt,
      // Al pasar a `running` se revelan todas las apuestas juntas.
      bets: this.publicBets(race),
    });
    this.run(race);

    return { race: this.detailFor(race, admin.id) };
  }

  cancelRace(token: string | null, raceId: string, reason: string): { race: RaceDetail } {
    const admin = this.requireAdmin(token);
    const race = this.requireRace(raceId);
    if (race.status === 'finished' || race.status === 'cancelled') {
      throw new MockApiError(409, 'INVALID_TRANSITION', 'Esa carrera ya está cerrada.');
    }

    this.stopTimer();
    race.status = 'cancelled';

    // Se devuelve cada apuesta ÍNTEGRA, y queda en el ledger como bet_refunded.
    const refunds = new Map<string, number>();
    for (const bet of this.bets) {
      if (bet.raceId !== race.id || bet.status !== 'placed') continue;
      bet.status = 'refunded';
      bet.payout = bet.amount;
      bet.settledAt = new Date().toISOString();
      const user = this.users.get(bet.userId);
      if (user === undefined) continue;
      this.credit(user, bet.amount, 'bet_refunded', { raceName: race.name });
      refunds.set(user.id, bet.amount);
    }

    // POR DESTINATARIO, igual que `race.finished`: el sobre lleva la devolución
    // de quien lo recibe, y difundir el mismo objeto la filtraría al resto.
    for (const userId of race.participants) {
      const user = this.users.get(userId);
      if (user === undefined) continue;
      const refund = refunds.get(userId) ?? null;
      this.publish(race.id, userId, {
        type: 'race.cancelled',
        raceId: race.id,
        reason,
        myRefund: refund,
        balance: refund === null ? null : user.balance,
      });
    }
    return { race: this.detailFor(race, admin.id) };
  }

  // ── La simulación ───────────────────────────────────────────────────────

  /**
   * La simulación es **autoritativa del servidor**: el cliente dibuja lo que
   * recibe. Acá el «servidor» es este mock, y la semilla queda guardada para que
   * la misma carrera se pueda volver a correr igual.
   */
  private run(race: MockRace): void {
    this.stopTimer();
    const random = prng(race.seed);

    // La cuota es la probabilidad implícita: menos cuota, más favorito, más
    // velocidad de base. La dispersión es lo que hace que igual pueda perder.
    const speed = new Map<string, number>();
    for (const horse of race.horses) {
      speed.set(horse.id, (100 / horse.nominalOdds) * 0.02 + 0.004 + random() * 0.004);
    }

    let tick = 0;
    this.timer = setInterval(() => {
      tick += 1;
      for (const horse of race.horses) {
        const base = speed.get(horse.id) ?? 0.01;
        const jitter = 0.6 + random() * 0.8;
        const next = Math.min(1, (race.progress.get(horse.id) ?? 0) + base * jitter);
        race.progress.set(horse.id, next);
      }

      const ranked = [...race.horses].sort(
        (a, b) => (race.progress.get(b.id) ?? 0) - (race.progress.get(a.id) ?? 0),
      );
      this.publish(race.id, null, {
        type: 'race.tick',
        raceId: race.id,
        t: tick / 10,
        positions: race.horses.map((horse) => ({
          horseId: horse.id,
          progress: race.progress.get(horse.id) ?? 0,
          place: ranked.findIndex((candidate) => candidate.id === horse.id) + 1,
        })),
      });

      const done = race.horses.some((horse) => (race.progress.get(horse.id) ?? 0) >= 1);
      if (done) this.finish(race);
    }, TICK_MS);
  }

  private finish(race: MockRace): void {
    this.stopTimer();
    race.status = 'finished';

    const ordered = [...race.horses].sort(
      (a, b) => (race.progress.get(b.id) ?? 0) - (race.progress.get(a.id) ?? 0),
    );
    race.results = ordered.map((horse, index) => ({
      position: index + 1,
      horseId: horse.id,
      horseName: horse.name,
    }));
    const winnerId = race.results[0]?.horseId ?? null;

    // El sobre de `race.finished` se arma POR DESTINATARIO: difundir el mismo
    // objeto filtraría cuánto cobró cada uno.
    // El pozo se reparte UNA vez, sobre las apuestas que siguen en pie, y recién
    // después se le paga a cada uno. Calcularlo adentro del bucle de destinatarios
    // sería recalcularlo con datos ya movidos: así es como se paga dos veces.
    const pending = this.bets.filter(
      (candidate) => candidate.raceId === race.id && candidate.status === 'placed',
    );
    const payouts = settlePariMutuel(pending, winnerId);
    const settledAt = new Date().toISOString();
    const refunded = pending.every((bet) => bet.horseId !== winnerId);

    for (const bet of pending) {
      bet.payout = payouts.get(bet.id) ?? 0;
      bet.settledAt = settledAt;
      bet.status = refunded ? 'refunded' : bet.horseId === winnerId ? 'won' : 'lost';
    }

    for (const userId of race.participants) {
      const user = this.users.get(userId);
      if (user === undefined) continue;

      const bet = pending.find((candidate) => candidate.userId === userId);
      const payout = bet?.payout ?? 0;
      if (bet !== undefined && payout > 0) {
        this.credit(user, payout, refunded ? 'bet_refunded' : 'bet_won', { raceName: race.name });
      }

      // POR DESTINATARIO: difundir el mismo objeto filtraría cuánto cobró cada
      // uno. `balance` va en null cuando no cambió, para no tocar el widget.
      this.publish(race.id, userId, {
        type: 'race.finished',
        raceId: race.id,
        results: race.results,
        bets: this.publicBets(race),
        myBet: bet === undefined ? null : this.toBet(bet),
        balance: payout > 0 ? user.balance : null,
      });
    }
  }

  private stopTimer(): void {
    if (this.timer !== null) {
      clearInterval(this.timer);
      this.timer = null;
    }
  }

  // ── Interno ─────────────────────────────────────────────────────────────

  private publish(raceId: string, to: string | null, event: RaceEvent): void {
    this.messages.next({ raceId, to, event });
  }

  private credit(
    user: MockUser,
    delta: number,
    reason: LedgerReason,
    extra: { raceName?: string; note?: string; at?: string } = {},
  ): void {
    const { at, ...rest } = extra;
    user.balance += delta;
    this.ledger.push({
      id: this.ledger.length + 1,
      userId: user.id,
      delta,
      reason,
      balanceAfter: user.balance,
      // `at` existe solo para la semilla: sin él, los cinco movimientos de
      // ejemplo tendrían la hora de arranque de la app y el historial se vería
      // como si todo hubiera pasado en el mismo segundo.
      createdAt: at ?? new Date().toISOString(),
      ...rest,
    });
  }

  private sessionFor(user: MockUser): Session {
    const accessToken = this.issue(user);
    return {
      accessToken,
      user: this.publicUser(user),
      balance: user.balance,
      points: pointsOf(user.balance),
    };
  }

  private issue(user: MockUser): string {
    const accessToken = `mock.${user.id}.${this.sequence++}`;
    this.tokens.set(accessToken, user.id);
    this.refreshTokens.set(`refresh.${user.id}.${this.sequence++}`, user.id);
    return accessToken;
  }

  private publicUser(user: MockUser): User {
    const { password: _password, balance: _balance, ...rest } = user;
    return rest;
  }

  /**
   * El token se **parsea**, no se busca en una tabla.
   *
   * Es lo que hace un JWT: el servidor no guarda sesiones, lee lo que el token
   * dice y verifica la firma. Acá no hay firma que verificar, pero sí importa la
   * consecuencia: al recargar la página, este mock arranca de cero y una tabla de
   * tokens en memoria dejaría afuera a quien ya estaba adentro. Con el id
   * embebido, recargar no cierra la sesión — igual que con el backend real.
   *
   * Los usuarios creados en esta corrida sí se pierden al recargar: el mundo
   * entero vive en memoria. Los de la semilla (`anag`, `profe`) tienen id fijo y
   * sobreviven.
   */
  private authenticate(token: string | null): MockUser {
    if (token === null) {
      throw new MockApiError(401, 'UNAUTHENTICATED', 'Tu sesión venció. Iniciá sesión de nuevo.');
    }
    const userId = this.tokens.get(token) ?? /^mock\.(.+)\.\d+$/.exec(token)?.[1];
    if (userId === undefined) {
      throw new MockApiError(401, 'UNAUTHENTICATED', 'Tu sesión venció. Iniciá sesión de nuevo.');
    }
    return this.requireUser(userId);
  }

  private requireAdmin(token: string | null): MockUser {
    const user = this.authenticate(token);
    if (user.role !== 'admin') {
      throw new MockApiError(403, 'FORBIDDEN', 'No tenés permiso para hacer eso.');
    }
    return user;
  }

  private requireUser(userId: string): MockUser {
    const user = this.users.get(userId);
    if (user === undefined) {
      throw new MockApiError(401, 'UNAUTHENTICATED', 'Tu sesión venció. Iniciá sesión de nuevo.');
    }
    return user;
  }

  private requireRace(raceId: string): MockRace {
    const race = this.races.get(raceId);
    if (race === undefined) {
      throw new MockApiError(404, 'RACE_NOT_FOUND', 'Esa carrera no existe.');
    }
    return race;
  }

  private requireCode(raw: string): MockCode {
    const code = raw.trim().toUpperCase();
    const found = this.codes.get(code);
    // Los dos errores son distintos a propósito: el alumno tiene que poder
    // distinguir «lo escribí mal» de «ya me registré».
    if (found === undefined) {
      throw new MockApiError(404, 'CODE_NOT_FOUND', 'Ese código no existe. Revisá que esté bien escrito.');
    }
    if (found.redeemedAt !== null) {
      throw new MockApiError(409, 'CODE_ALREADY_REDEEMED', 'Ese código ya fue usado.');
    }
    return found;
  }

  private findByUsername(username: string): MockUser | undefined {
    const wanted = username.trim().toLowerCase();
    return [...this.users.values()].find((user) => user.username.toLowerCase() === wanted);
  }

  private betOf(race: MockRace, userId: string): Bet | null {
    const bet = this.bets.find(
      (candidate) => candidate.raceId === race.id && candidate.userId === userId,
    );
    return bet === undefined ? null : this.toBet(bet);
  }

  private toBet(bet: MockBet): Bet {
    return {
      id: bet.id,
      horseId: bet.horseId,
      amount: bet.amount,
      status: bet.status,
      payout: bet.payout,
      createdAt: bet.createdAt,
      settledAt: bet.settledAt,
    };
  }

  private participantsOf(race: MockRace): Participant[] {
    return [...race.participants].flatMap((userId) => {
      const user = this.users.get(userId);
      return user === undefined ? [] : [{ userId: user.id, username: user.username }];
    });
  }

  private publicBets(race: MockRace): PublicBet[] {
    const hide = race.status === 'open';
    return this.bets
      .filter((bet) => bet.raceId === race.id)
      .flatMap((bet) => {
        const user = this.users.get(bet.userId);
        if (user === undefined) return [];
        return [
          {
            userId: bet.userId,
            username: user.username,
            amount: bet.amount,
            horseId: hide ? null : bet.horseId,
          },
        ];
      });
  }

  private detailFor(race: MockRace, userId: string): RaceDetail {
    return {
      id: race.id,
      name: race.name,
      status: race.status,
      scheduledAt: race.scheduledAt,
      horses: race.horses,
      participants: this.participantsOf(race),
      bets: this.publicBets(race),
      // El pago propio sale de `myBet.payout`, ya liquidado. No es un campo
      // aparte: dos números para lo mismo se desincronizan siempre.
      myBet: this.betOf(race, userId),
      results: race.results,
    };
  }

  private nextId(prefix: string): string {
    return `${prefix}-${this.sequence++}`;
  }

  private randomCode(): string {
    const pick = (alphabet: string): string =>
      alphabet[Math.floor(Math.random() * alphabet.length)] ?? 'A';
    const letters = Array.from({ length: 4 }, () => pick(CODE_LETTERS)).join('');
    const digits = Array.from({ length: 4 }, () => pick(CODE_DIGITS)).join('');
    return `${letters}-${digits}`;
  }

  // ── La semilla ──────────────────────────────────────────────────────────

  private seed(): void {
    const instructor: MockUser = {
      id: 'user-admin',
      username: 'profe',
      firstName: 'Jhonatan',
      lastName: 'Soto',
      role: 'admin',
      password: 'arena1234',
      balance: 0,
    };
    this.users.set(instructor.id, instructor);

    const ana: MockUser = {
      id: 'user-ana',
      username: 'anag',
      firstName: 'Ana',
      lastName: 'Gómez',
      role: 'student',
      password: 'arena1234',
      balance: 0,
    };
    this.users.set(ana.id, ana);
    this.credit(ana, 1000, 'code_redeemed', {
      note: 'Código KMPR-8827',
      at: '2026-07-22T18:11:00.000Z',
    });
    this.credit(ana, 300, 'gift', {
      note: 'Participación en clase',
      at: '2026-07-23T21:40:00.000Z',
    });

    const codes: readonly MockCode[] = [
      {
        code: 'AVBD-1234',
        coinsGranted: 1000,
        note: 'grupo del martes',
        createdAt: '2026-07-20T13:00:00.000Z',
        redeemedAt: null,
        redeemedBy: null,
      },
      {
        code: 'KMPR-8827',
        coinsGranted: 1000,
        note: 'grupo del martes',
        createdAt: '2026-07-20T13:00:00.000Z',
        redeemedAt: '2026-07-22T18:11:00.000Z',
        redeemedBy: 'anag',
      },
      {
        code: 'TXNQ-4562',
        coinsGranted: 1000,
        note: 'grupo del jueves',
        createdAt: '2026-07-21T13:00:00.000Z',
        redeemedAt: null,
        redeemedBy: null,
      },
    ];
    for (const code of codes) this.codes.set(code.code, { ...code });

    const open: MockRace = {
      id: 'race-open',
      name: 'Clásico del Recuerdo',
      status: 'open',
      scheduledAt: '2026-07-29T22:30:00.000Z',
      horses: [
        { id: 'horse-1', number: 1, name: 'Viento Norte', nominalOdds: 340 },
        { id: 'horse-2', number: 2, name: 'Tinta China', nominalOdds: 210 },
        { id: 'horse-3', number: 3, name: 'Farol de Niebla', nominalOdds: 750 },
        { id: 'horse-4', number: 4, name: 'Última Curva', nominalOdds: 480 },
        { id: 'horse-5', number: 5, name: 'Pampa Seca', nominalOdds: 1250 },
        { id: 'horse-6', number: 6, name: 'Rayo de Tiza', nominalOdds: 620 },
      ],
      participants: new Set([ana.id]),
      seed: 0,
      startedAt: null,
      results: null,
      progress: new Map<string, number>(),
    };
    this.races.set(open.id, open);
    this.bets.push({
      id: 'bet-seed-1',
      raceId: open.id,
      userId: ana.id,
      horseId: 'horse-2',
      amount: 200,
      createdAt: '2026-07-29T22:05:00.000Z',
      status: 'placed',
      payout: null,
      settledAt: null,
    });
    this.credit(ana, -200, 'bet_placed', {
      raceName: open.name,
      at: '2026-07-29T22:05:00.000Z',
    });

    const finished: MockRace = {
      id: 'race-finished',
      name: 'Gran Premio de la Tiza',
      status: 'finished',
      scheduledAt: '2026-07-24T21:00:00.000Z',
      horses: [
        { id: 'horse-11', number: 1, name: 'Doña Estampa', nominalOdds: 260 },
        { id: 'horse-12', number: 2, name: 'Bandera Roja', nominalOdds: 430 },
        { id: 'horse-13', number: 3, name: 'Sombra Larga', nominalOdds: 910 },
        { id: 'horse-14', number: 4, name: 'Cabo Suelto', nominalOdds: 380 },
      ],
      participants: new Set([ana.id]),
      seed: 42,
      startedAt: '2026-07-24T21:02:00.000Z',
      results: [
        { position: 1, horseId: 'horse-11', horseName: 'Doña Estampa' },
        { position: 2, horseId: 'horse-14', horseName: 'Cabo Suelto' },
        { position: 3, horseId: 'horse-12', horseName: 'Bandera Roja' },
        { position: 4, horseId: 'horse-13', horseName: 'Sombra Larga' },
      ],
      progress: new Map([
        ['horse-11', 1],
        ['horse-12', 0.94],
        ['horse-13', 0.9],
        ['horse-14', 0.97],
      ]),
    };
    this.races.set(finished.id, finished);
    this.bets.push({
      id: 'bet-seed-2',
      raceId: finished.id,
      userId: ana.id,
      horseId: 'horse-11',
      amount: 150,
      createdAt: '2026-07-24T20:58:00.000Z',
      // Ana fue la única que le pegó al ganador, así que se llevó el pozo entero:
      // su apuesta de 150 más las de los demás. Con cuota fija habría cobrado 390.
      status: 'won',
      payout: 150,
      settledAt: '2026-07-24T21:03:30.000Z',
    });
    this.credit(ana, -150, 'bet_placed', {
      raceName: finished.name,
      at: '2026-07-24T20:58:00.000Z',
    });
    this.credit(ana, 150, 'bet_won', {
      raceName: finished.name,
      at: '2026-07-24T21:06:00.000Z',
    });

    const draft: MockRace = {
      id: 'race-draft',
      name: 'Handicap de Cierre',
      status: 'draft',
      scheduledAt: '2026-07-31T22:00:00.000Z',
      horses: [
        { id: 'horse-21', number: 1, name: 'Cuarta Raya', nominalOdds: 300 },
        { id: 'horse-22', number: 2, name: 'Malón', nominalOdds: 520 },
        { id: 'horse-23', number: 3, name: 'Punto y Banca', nominalOdds: 700 },
      ],
      participants: new Set<string>(),
      seed: 0,
      startedAt: null,
      results: null,
      progress: new Map<string, number>(),
    };
    this.races.set(draft.id, draft);
  }
}
