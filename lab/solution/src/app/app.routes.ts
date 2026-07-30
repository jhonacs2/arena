import { Routes } from '@angular/router';

/**
 * Una ruta por sesión. Cada una se carga con `loadComponent`, así que el
 * navegador solo descarga la sesión que estás mirando.
 */
export const routes: Routes = [
  {
    path: 's00',
    title: 'S0 · TypeScript · Lab',
    loadComponent: () => import('./sessions/s00/s00.component').then((m) => m.S00Component),
  },
  {
    path: 's01',
    title: 'S1 · Primer componente · Lab',
    loadComponent: () => import('./sessions/s01/s01.component').then((m) => m.S01Component),
  },
  {
    path: 's02',
    title: 'S2 · Anatomía de un componente · Lab',
    loadComponent: () => import('./sessions/s02/s02.component').then((m) => m.S02Component),
  },
  {
    path: 's03',
    title: 'S3 · Signals y control flow · Lab',
    loadComponent: () => import('./sessions/s03/s03.component').then((m) => m.S03Component),
  },
  {
    path: 's04',
    title: 'S4 · Directivas y pipes · Lab',
    loadComponent: () => import('./sessions/s04/s04.component').then((m) => m.S04Component),
  },
  {
    path: 's05',
    title: 'S5 · Inyección de dependencias · Lab',
    loadComponent: () => import('./sessions/s05/s05.component').then((m) => m.S05Component),
  },
  { path: '', redirectTo: 's00', pathMatch: 'full' },
  { path: '**', redirectTo: 's00' },
];
