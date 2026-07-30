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
 * `balance` viene en **monedas enteras** y `points` es
 * `max(10, floor(balance / 100)) + regalados` — el piso de 10 y los puntos que
 * regala el instructor son parte de la fórmula (`decisiones.md` §1).
 *
 * Los dos llegan del servidor: los puntos son una vista, no una columna, y el
 * frontend no es quien decide la nota de nadie. Puede ser **fraccionario**
 * (12,5), porque los regalos viajan ×100 en entero y se muestran con decimal.
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
