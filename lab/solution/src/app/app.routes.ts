import { Routes } from '@angular/router';

/**
 * Una ruta por sesión. Cada una se carga con `loadComponent`, así que el
 * navegador solo descarga la sesión que estás mirando.
 */
export const routes: Routes = [
  {
    path: 's01',
    title: 'S1 · Primer componente · Lab',
    loadComponent: () => import('./sesiones/s01/s01.component').then((m) => m.S01Component),
  },
  { path: '', redirectTo: 's01', pathMatch: 'full' },
  { path: '**', redirectTo: 's01' },
];
