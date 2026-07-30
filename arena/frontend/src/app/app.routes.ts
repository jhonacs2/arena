import type { Routes } from '@angular/router';

import { adminGuard, authGuard, guestGuard } from './core/auth/auth.guard';

/**
 * Las rutas.
 *
 * **Las URLs van en castellano**: son navegación que el usuario lee, igual que el
 * texto de la pantalla. El código que las declara, en inglés.
 *
 * Todo con `loadComponent`: el alumno que entra a registrarse no necesita bajar el
 * panel del instructor, y en el aula el ancho de banda se reparte entre veinte
 * personas al mismo tiempo.
 */
export const routes: Routes = [
  {
    path: 'registro',
    title: 'Canjeá tu código · Arena',
    canActivate: [guestGuard],
    loadComponent: () => import('./features/register/register').then((m) => m.Register),
  },
  {
    path: 'ingresar',
    title: 'Iniciar sesión · Arena',
    canActivate: [guestGuard],
    loadComponent: () => import('./features/login/login').then((m) => m.Login),
  },
  {
    path: 'tablero',
    title: 'Tu tablero · Arena',
    canActivate: [authGuard],
    loadComponent: () => import('./features/dashboard/dashboard').then((m) => m.Dashboard),
  },
  {
    path: 'carreras/:id',
    title: 'Carrera · Arena',
    canActivate: [authGuard],
    loadComponent: () => import('./features/race/race').then((m) => m.Race),
  },
  {
    path: 'instructor',
    title: 'Panel del instructor · Arena',
    canActivate: [adminGuard],
    loadComponent: () => import('./features/admin/admin').then((m) => m.Admin),
  },
  { path: '', pathMatch: 'full', redirectTo: 'tablero' },
  // Cualquier otra cosa cae al registro: es la puerta de entrada, y alguien que
  // llegó con una URL vieja tiene que terminar en algún lado con sentido.
  { path: '**', redirectTo: 'registro' },
];
