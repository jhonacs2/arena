import { environment } from '../../../environments/environment';

/**
 * Arma una URL de la API.
 *
 * Existe para que ningún componente escriba `/api/...` a mano: el día que el
 * prefijo cambie —o que haga falta apuntar a otro host desde una demo— se cambia
 * en `environment.ts` y nada más.
 */
export const api = (path: string): string => `${environment.apiBaseUrl}${path}`;
