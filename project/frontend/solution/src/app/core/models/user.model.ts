/** Usuario y sesión — `docs/contract/openapi.yaml`. */

export interface User {
  readonly id: string;
  readonly email: string;
  readonly displayName: string;
  /** Saldo virtual. Entero, sin centavos: no se mueve dinero real. */
  readonly balance: number;
  readonly emailVerified: boolean;
}

/** Respuesta de `POST /auth/login`. */
export interface AuthTokens {
  /** JWT. Vive 15 minutos. */
  readonly accessToken: string;
  /** Opaco. Vive 30 días y es de **un solo uso**. */
  readonly refreshToken: string;
  readonly user: User;
}

/** Respuesta de `POST /auth/refresh`: el par rotado, sin el usuario. */
export interface RefreshedTokens {
  readonly accessToken: string;
  readonly refreshToken: string;
}

/** Respuesta de `POST /auth/verify`. */
export interface VerifiedUser {
  readonly user: User;
}

/** Saldo que se otorga al registrarse. */
export const SIGNUP_BALANCE = 1000;
