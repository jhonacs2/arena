import { Injectable, computed, signal } from '@angular/core';

import type { Me, Role, Session, User } from '../models';

const STORAGE_KEY = 'arena:session';

/**
 * La parte de la nota que sale del saldo, con el piso de 10 (`decisiones.md` §1).
 *
 * Es la misma fórmula que la vista `user_scores` de Postgres. Que esté escrita
 * dos veces es incómodo, y la alternativa —pedir `/me` después de cada apuesta—
 * es peor: una llamada extra por apuesta para recalcular algo que ya sabemos.
 */
const pointsFromCoins = (balance: number): number => Math.max(10, Math.floor(balance / 100));

/**
 * Despeja los puntos regalados del total que informó el servidor.
 *
 * `total = pointsFromCoins(saldo) + regalados`, así que los regalados son la
 * resta. Se hace con el saldo **del mismo momento** que el total; usar el saldo
 * de después daría cualquier cosa.
 *
 * Nunca negativo: si el servidor cambiara la fórmula y esta resta diera menos
 * que cero, restarle nota a alguien sería peor que mostrar el piso.
 */
const grantedFrom = (points: number, balance: number): number =>
  Math.max(0, points - pointsFromCoins(balance));

interface StoredSession {
  readonly accessToken: string;
  readonly user: User;
  readonly balance: number;
  readonly pointsGranted: number;
}

/**
 * Dónde vive la sesión. **No hace HTTP**: eso es `AuthService`.
 *
 * Están separados porque el interceptor necesita el token en cada request y no
 * puede depender del servicio que a su vez usa `HttpClient` — se armaría un
 * círculo. Acá solo hay estado.
 *
 * El *access token* se guarda en `localStorage`. La parte que de verdad sostiene
 * la sesión es el refresh en cookie `HttpOnly`, que el JavaScript no puede leer;
 * persistir el access token solo evita el parpadeo de la pantalla de login al
 * recargar. Vence en 15 minutos y el interceptor lo renueva.
 */
@Injectable({ providedIn: 'root' })
export class SessionStore {
  private readonly stored = this.read();

  private readonly _accessToken = signal<string | null>(this.stored?.accessToken ?? null);
  private readonly _user = signal<User | null>(this.stored?.user ?? null);
  private readonly _balance = signal<number>(this.stored?.balance ?? 0);

  /**
   * Los puntos que regaló el instructor, tal como los informó el servidor.
   *
   * No se calculan: no pasan por el juego. Se despejan del total que manda el
   * servidor restándole la parte que sí sale del saldo.
   */
  private readonly _pointsGranted = signal<number>(this.stored?.pointsGranted ?? 0);

  readonly accessToken = this._accessToken.asReadonly();
  readonly user = this._user.asReadonly();
  readonly balance = this._balance.asReadonly();

  /**
   * La nota: `max(10, floor(saldo / 100)) + regalados` — `decisiones.md` §1.
   *
   * **El piso de 10 no es cosmético:** perder monedas saca capacidad de seguir
   * jugando, nunca calificación. Sin él, un alumno fundido veía «0 pts» en el
   * encabezado mientras el servidor le daba 10, y la pantalla contradecía a la
   * nota real.
   *
   * Los puntos regalados **no se pueden derivar del saldo** —no pasan por el
   * juego—, así que se guardan aparte, tomados del último dato del servidor. El
   * saldo sí manda sobre su propia parte: así el número se mueve al apostar sin
   * pedir `/me` en cada apuesta.
   */
  readonly points = computed(() => pointsFromCoins(this._balance()) + this._pointsGranted());

  readonly isAuthenticated = computed(() => this._accessToken() !== null);
  readonly role = computed<Role | null>(() => this._user()?.role ?? null);
  readonly isAdmin = computed(() => this.role() === 'admin');

  readonly fullName = computed(() => {
    const user = this._user();
    return user ? `${user.firstName} ${user.lastName}`.trim() : '';
  });

  start(session: Session): void {
    this._accessToken.set(session.accessToken);
    this._user.set(session.user);
    this._balance.set(session.balance);
    this._pointsGranted.set(grantedFrom(session.points, session.balance));
    this.persist();
  }

  /** Solo el token: es lo único que devuelve un refresh. */
  renew(accessToken: string): void {
    this._accessToken.set(accessToken);
    this.persist();
  }

  sync(me: Me): void {
    this._user.set(me.user);
    this._balance.set(me.balance);
    this._pointsGranted.set(grantedFrom(me.points, me.balance));
    this.persist();
  }

  setBalance(balance: number): void {
    this._balance.set(balance);
    this.persist();
  }

  clear(): void {
    this._accessToken.set(null);
    this._user.set(null);
    this._balance.set(0);
    this._pointsGranted.set(0);
    try {
      localStorage.removeItem(STORAGE_KEY);
    } catch {
      // Modo incógnito o almacenamiento bloqueado. La sesión igual funciona en
      // memoria; solo no sobrevive a la recarga.
    }
  }

  private persist(): void {
    const accessToken = this._accessToken();
    const user = this._user();
    if (accessToken === null || user === null) return;

    const payload: StoredSession = {
      accessToken,
      user,
      balance: this._balance(),
      pointsGranted: this._pointsGranted(),
    };
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(payload));
    } catch {
      // Ver clear().
    }
  }

  private read(): StoredSession | null {
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (raw === null) return null;
      const parsed: unknown = JSON.parse(raw);
      if (
        typeof parsed === 'object' &&
        parsed !== null &&
        'accessToken' in parsed &&
        'user' in parsed
      ) {
        return parsed as StoredSession;
      }
    } catch {
      // Un JSON corrupto no tiene que dejar la app sin arrancar.
    }
    return null;
  }
}
