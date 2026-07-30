export type Role = 'student' | 'admin';

export interface User {
  readonly id: string;
  readonly username: string;
  readonly firstName: string;
  readonly lastName: string;
  readonly role: Role;
}

/**
 * Lo que devuelven `/auth/redeem`, `/auth/login` y `/auth/refresh`.
 *
 * `balance` viene en **monedas enteras** y `points` es `floor(balance / 100)`.
 * Los dos llegan del servidor: los puntos son una vista, no una columna
 * (`decisiones.md` §1), y el frontend no es quien decide la nota de nadie.
 */
export interface Session {
  readonly accessToken: string;
  readonly user: User;
  readonly balance: number;
  readonly points: number;
}

/** `GET /api/me`. Igual que `Session` pero sin token. */
export interface Me {
  readonly user: User;
  readonly balance: number;
  readonly points: number;
}

export interface RedeemRequest {
  readonly code: string;
  readonly firstName: string;
  readonly lastName: string;
  readonly username: string;
  readonly password: string;
}

export interface LoginRequest {
  readonly username: string;
  readonly password: string;
}

export interface CheckCodeResponse {
  readonly valid: boolean;
  readonly coinsGranted: number;
}
