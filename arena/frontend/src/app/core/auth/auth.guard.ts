import { inject } from '@angular/core';
import { Router, type CanActivateFn } from '@angular/router';

import { SessionStore } from './session.store';

/**
 * Guards funcionales. **No son un control de acceso**: el rol y la sesión se
 * validan en el servidor en cada endpoint (`decisiones.md` §4). Esto solo evita
 * mostrarle a alguien una pantalla que no va a poder usar.
 */
export const authGuard: CanActivateFn = (_route, state) => {
  const session = inject(SessionStore);
  if (session.isAuthenticated()) return true;

  const router = inject(Router);
  return router.createUrlTree(['/ingresar'], { queryParams: { volver: state.url } });
};

export const adminGuard: CanActivateFn = () => {
  const session = inject(SessionStore);
  if (session.isAdmin()) return true;

  const router = inject(Router);
  return router.createUrlTree([session.isAuthenticated() ? '/tablero' : '/ingresar']);
};

/** Para registro e inicio de sesión: si ya hay sesión, no tiene sentido volver ahí. */
export const guestGuard: CanActivateFn = () => {
  const session = inject(SessionStore);
  if (!session.isAuthenticated()) return true;

  const router = inject(Router);
  return router.createUrlTree([session.isAdmin() ? '/instructor' : '/tablero']);
};
