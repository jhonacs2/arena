import { Routes } from '@angular/router';

/**
 * Las rutas del lab. Una por sesión.
 *
 * Hay dos hechas: la pantalla de inicio y la de S0. El resto las vas a ir
 * sumando vos, una por clase.
 *
 * La de S0 viene dada a propósito: declarar rutas y escribir componentes son
 * los temas de S9 y S1, y en la sesión 0 todavía no viste ninguno de los dos.
 * Lo que sí es tuyo en S0 está en `sessions/s00/menu.ts`.
 *
 * `loadComponent` es carga diferida: el navegador descarga el código de una
 * sesión recién cuando entrás. Se ve a fondo en S9; por ahora alcanza con
 * copiar la forma.
 */
export const routes: Routes = [
  {
    path: 's00',
    title: 'S0 · TypeScript · Lab',
    loadComponent: () => import('./sessions/s00/s00.component').then((m) => m.S00Component),
  },

  // TODO(S1): sumar la ruta de tu primera sesión, así:
  //
  //   {
  //     path: 's01',
  //     title: 'S1 · Primer componente · Lab',
  //     loadComponent: () => import('./sessions/s01/s01.component').then((m) => m.S01Component),
  //   },
  //
  // Y acordate de poner `available: true` en sessions.ts, o el enlace no
  // aparece en la barra lateral.

  {
    path: 'home',
    title: 'Lab · Módulo Angular',
    loadComponent: () => import('./home/home.component').then((m) => m.HomeComponent),
  },
  { path: '', redirectTo: 'home', pathMatch: 'full' },
  { path: '**', redirectTo: 'home' },
];
