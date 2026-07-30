/**
 * El sobre de error de `arena/docs/contract/api.md`.
 *
 * ```json
 * { "error": { "code": "CODE_ALREADY_REDEEMED", "message": "Ese código ya fue usado." } }
 * ```
 *
 * `message` viene **en castellano** desde el servidor y se muestra tal cual. El
 * frontend no lo reescribe: si mañana el backend afina la redacción, la app la
 * refleja sin tocar una línea. `code` es para decidir *qué* mostrar.
 */
export type ApiErrorCode =
  | 'VALIDATION_FAILED'
  | 'CODE_NOT_FOUND'
  | 'CODE_ALREADY_REDEEMED'
  | 'USERNAME_TAKEN'
  | 'INVALID_CREDENTIALS'
  | 'UNAUTHENTICATED'
  | 'FORBIDDEN'
  | 'RACE_NOT_FOUND'
  | 'RACE_NOT_OPEN'
  | 'BET_ALREADY_PLACED'
  | 'INSUFFICIENT_BALANCE'
  | 'INVALID_TRANSITION'
  | 'INTERNAL'
  /** No está en el catálogo: es lo que ponemos cuando ni siquiera hubo respuesta. */
  | 'NETWORK';

export interface ApiErrorBody {
  readonly code: ApiErrorCode;
  readonly message: string;
}

export interface ApiErrorEnvelope {
  readonly error: ApiErrorBody;
}
