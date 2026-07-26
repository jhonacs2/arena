/**
 * Índice de las sesiones del lab.
 *
 * Una entrada por sesión. La navegación y las rutas salen de acá, así que
 * sumar una sesión es agregar una línea y su carpeta — no hay que tocar dos
 * lugares y olvidarse de uno.
 */
export interface Sesion {
  readonly numero: number;
  readonly slug: string;
  readonly titulo: string;
  /** El concepto de la sesión, en una frase. */
  readonly concepto: string;
  readonly disponible: boolean;
}

export const SESIONES: readonly Sesion[] = [
  {
    numero: 1,
    slug: 's01',
    titulo: 'Primer componente',
    concepto: 'Un componente standalone y los cuatro tipos de binding.',
    disponible: true,
  },
  { numero: 2, slug: 's02', titulo: 'Anatomía de un componente', concepto: 'input, output y ng-content.', disponible: false },
  { numero: 3, slug: 's03', titulo: 'Signals y control flow', concepto: 'signal, computed, @if y @for.', disponible: false },
  { numero: 4, slug: 's04', titulo: 'Directivas y pipes', concepto: 'Transformar sin tocar el componente.', disponible: false },
  { numero: 5, slug: 's05', titulo: 'Inyección de dependencias', concepto: 'Servicios, inject() y tokens.', disponible: false },
  { numero: 6, slug: 's06', titulo: 'Reactividad', concepto: 'Observables, operadores y debounce.', disponible: false },
  { numero: 7, slug: 's07', titulo: 'HttpClient', concepto: 'Pedir datos y manejar los tres estados.', disponible: false },
  { numero: 8, slug: 's08', titulo: 'Reactive Forms', concepto: 'Formularios con validación de verdad.', disponible: false },
  { numero: 9, slug: 's09', titulo: 'Routing', concepto: 'Rutas, parámetros y guards.', disponible: false },
  { numero: 10, slug: 's10', titulo: 'WebSockets', concepto: 'Tiempo real, zona y OnPush.', disponible: false },
  { numero: 11, slug: 's11', titulo: 'Producción', concepto: 'NgModules como legado y build de producción.', disponible: false },
];

export const SESIONES_DISPONIBLES = SESIONES.filter((s) => s.disponible);
