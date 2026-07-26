import { Routes } from '@angular/router';

/**
 * Rutas de la aplicación.
 *
 * `loadComponent` carga la vista recién cuando se navega a ella. Se usa desde
 * el principio para que en S9, cuando se explique lazy loading, el alumno lo
 * reconozca en lugar de verlo por primera vez.
 *
 * `/sistema` es la muestra del sistema de diseño: no es una pantalla del
 * producto, pero se queda como referencia.
 */
export const routes: Routes = [
  {
    path: 'carreras',
    title: 'Carreras · Hipódromo',
    loadComponent: () =>
      import('./features/races/race-list.component').then((m) => m.RaceListComponent),
  },
  {
    path: 'sistema',
    title: 'Sistema de diseño · Hipódromo',
    loadComponent: () =>
      import('./features/sistema/sistema.component').then((m) => m.SistemaComponent),
  },
  { path: '', redirectTo: 'carreras', pathMatch: 'full' },
  // La ruta comodín va SIEMPRE al final: el router toma la primera que
  // coincide, así que declarada arriba se comería todas las demás. Es uno de
  // los "predice y ejecuta" de S9.
  { path: '**', redirectTo: 'carreras' },
];
