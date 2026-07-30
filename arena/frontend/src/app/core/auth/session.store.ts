import { Injectable, computed, signal } from '@angular/core';

import type { Me, Role, Session, User } from '../models';

const STORAGE_KEY = 'arena:session';

interface StoredSession {
  readonly accessToken: string;
  readonly user: User;
  readonly balance: number;
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

  readonly accessToken = this._accessToken.asReadonly();
  readonly user = this._user.asReadonly();
  readonly balance = this._balance.asReadonly();

  /**
   * **Los puntos son una función del saldo**, no un dato aparte: `decisiones.md`
   * §1. Dos números que representan lo mismo se desincronizan siempre, así que
   * acá tampoco se guardan dos.
   */
  readonly points = computed(() => Math.floor(this._balance() / 100));

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

    const payload: StoredSession = { accessToken, user, balance: this._balance() };
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
