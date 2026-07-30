import { HttpErrorResponse } from '@angular/common/http';

import type { ApiErrorBody, ApiErrorCode, ApiErrorEnvelope } from '../models';

/**
 * Traduce cualquier cosa que pueda salir mal a un `ApiErrorBody`.
 *
 * La regla es una sola: **el `message` del servidor se muestra tal cual**, porque
 * ya viene en castellano y escrito para el alumno (`api.md`). El frontend solo
 * inventa texto cuando no hubo respuesta —red caída, servidor apagado— porque en
 * ese caso no hay nada que mostrar.
 */
const FALLBACK: Readonly<Record<ApiErrorCode, string>> = {
  VALIDATION_FAILED: 'Revisá los datos: algo quedó incompleto.',
  CODE_NOT_FOUND: 'Ese código no existe. Revisá que esté bien escrito.',
  CODE_ALREADY_REDEEMED: 'Ese código ya fue usado.',
  USERNAME_TAKEN: 'Ese usuario ya está ocupado.',
  INVALID_CREDENTIALS: 'Usuario o contraseña incorrectos.',
  UNAUTHENTICATED: 'Tu sesión venció. Iniciá sesión de nuevo.',
  FORBIDDEN: 'No tenés permiso para hacer eso.',
  RACE_NOT_FOUND: 'Esa carrera no existe.',
  RACE_NOT_OPEN: 'La carrera ya no acepta apuestas.',
  BET_ALREADY_PLACED: 'Ya apostaste en esta carrera.',
  INSUFFICIENT_BALANCE: 'No te alcanzan las monedas.',
  INVALID_TRANSITION: 'La carrera no puede pasar a ese estado.',
  INTERNAL: 'Algo se rompió del lado del servidor.',
  NETWORK: 'No se pudo contactar al servidor. Revisá la conexión.',
};

const CODES = new Set<string>(Object.keys(FALLBACK));

function isEnvelope(value: unknown): value is ApiErrorEnvelope {
  if (typeof value !== 'object' || value === null || !('error' in value)) return false;
  const { error } = value as { error: unknown };
  return typeof error === 'object' && error !== null && 'code' in error && 'message' in error;
}

export function toApiError(cause: unknown): ApiErrorBody {
  if (cause instanceof HttpErrorResponse) {
    if (isEnvelope(cause.error)) {
      const { code, message } = cause.error.error;
      return {
        code: CODES.has(code) ? code : 'INTERNAL',
        message: message.trim() || FALLBACK[code] || FALLBACK.INTERNAL,
      };
    }
    // Sin sobre no hay mensaje del servidor: o no llegó la respuesta (status 0)
    // o algo devolvió HTML donde tendría que haber JSON.
    const code: ApiErrorCode = cause.status === 0 ? 'NETWORK' : 'INTERNAL';
    return { code, message: FALLBACK[code] };
  }

  return { code: 'INTERNAL', message: FALLBACK.INTERNAL };
}

/** El texto por defecto de un código, para cuando hay que mostrarlo sin respuesta. */
export function fallbackMessage(code: ApiErrorCode): string {
  return FALLBACK[code];
}
