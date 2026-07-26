/** Formas comunes de la API — paginación y errores. */

/** Respuesta paginada. `items` nunca es `null`. */
export interface Page<T> {
  readonly items: readonly T[];
  readonly page: number;
  readonly size: number;
  /** Total que matchea el filtro, no de la página. */
  readonly total: number;
}

/**
 * Catálogo **cerrado** de códigos de error — `docs/contract/error-codes.md`.
 *
 * El frontend hace `switch` sobre el código, nunca sobre el mensaje: el mensaje
 * está para mostrarlo, el código para decidir. Que sea una unión de literales
 * y no `string` es lo que hace que el `switch` sea exhaustivo y que agregar un
 * código rompa la compilación donde falte manejarlo.
 */
export type ApiErrorCode =
  | 'VALIDATION_FAILED'
  | 'INVALID_CREDENTIALS'
  | 'EMAIL_ALREADY_REGISTERED'
  | 'EMAIL_NOT_VERIFIED'
  | 'INVALID_VERIFICATION_TOKEN'
  | 'VERIFICATION_TOKEN_EXPIRED'
  | 'ALREADY_VERIFIED'
  | 'UNAUTHENTICATED'
  | 'INVALID_REFRESH_TOKEN'
  | 'FORBIDDEN'
  | 'NOT_FOUND'
  | 'RACE_ALREADY_STARTED'
  | 'HORSE_NOT_IN_RACE'
  | 'INSUFFICIENT_BALANCE'
  | 'BET_AMOUNT_OUT_OF_RANGE'
  | 'RESULTS_NOT_AVAILABLE'
  | 'RATE_LIMITED'
  | 'INTERNAL';

export interface ApiErrorBody {
  readonly code: ApiErrorCode;
  /** En español. Se muestra al usuario tal cual. */
  readonly message: string;
  /** Siempre existe: `{}` cuando no hay nada que agregar. */
  readonly details: Readonly<Record<string, unknown>>;
}

/** El sobre uniforme: `{ "error": { code, message, details } }`. */
export interface ApiError {
  readonly error: ApiErrorBody;
}

/** Errores por campo de un `VALIDATION_FAILED`. Los usa S8. */
export function fieldErrors(error: ApiErrorBody): Readonly<Record<string, string>> {
  const fields = error.details['fields'];
  return typeof fields === 'object' && fields !== null
    ? (fields as Record<string, string>)
    : {};
}
