import { Routes } from '@angular/router';

/**
 * Las rutas de la aplicación.
 *
 * Hoy hay una sola: `/sistema`, la muestra del sistema de diseño. No es una
 * pantalla del producto — es tu referencia: están todos los colores, las
 * tipografías, los botones y las sedas de los caballos. Cuando algo se vea
 * raro, mirá ahí primero.
 *
 * `loadComponent` es carga diferida: el navegador descarga el código de una
 * pantalla recién cuando entrás. Se ve a fondo en S9.
 */
export const routes: Routes = [
  // TODO(S1): sumar la ruta del listado de carreras y hacerla la principal.
  //
  //   {
  //     path: 'carreras',
  //     title: 'Carreras · Hipódromo',
  //     loadComponent: () =>
  //       import('./features/races/race-list.component').then((m) => m.RaceListComponent),
  //   },
  //
  // Después cambiá los dos redirectTo de abajo para que apunten a 'carreras',
  // y sumá el enlace en el encabezado (layout/shell.component.html).

  {
    path: 'sistema',
    title: 'Sistema de diseño · Hipódromo',
    loadComponent: () =>
      import('./features/sistema/sistema.component').then((m) => m.SistemaComponent),
  },
  { path: '', redirectTo: 'sistema', pathMatch: 'full' },
  // La ruta comodín va SIEMPRE al final: el router toma la primera que
  // coincide, así que declarada arriba se comería todas las demás.
  { path: '**', redirectTo: 'sistema' },
];
