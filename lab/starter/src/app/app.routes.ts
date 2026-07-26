import { Routes } from '@angular/router';

/**
 * Las rutas del lab. Una por sesión.
 *
 * Hoy hay una sola: la pantalla de inicio. El resto las vas a ir sumando vos.
 *
 * `loadComponent` es carga diferida: el navegador descarga el código de una
 * sesión recién cuando entrás. Se ve a fondo en S9; por ahora alcanza con
 * copiar la forma.
 */
export const routes: Routes = [
  // TODO(S1): sumar la ruta de tu primera sesión, así:
  //
  //   {
  //     path: 's01',
  //     title: 'S1 · Primer componente · Lab',
  //     loadComponent: () => import('./sessions/s01/s01.component').then((m) => m.S01Component),
  //   },
  //
  // Y acordate de poner `disponible: true` en sesiones.ts, o el enlace no
  // aparece en la barra lateral.

  {
    path: 'home',
    title: 'Lab · Módulo Angular',
    loadComponent: () => import('./home/home.component').then((m) => m.HomeComponent),
  },
  { path: '', redirectTo: 'home', pathMatch: 'full' },
  { path: '**', redirectTo: 'home' },
];
