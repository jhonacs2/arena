/**
 * Índice de las sesiones del lab.
 *
 * Una entrada por sesión. La navegación y las rutas salen de acá, así que
 * sumar una sesión es agregar una línea y su carpeta — no hay que tocar dos
 * lugares y olvidarse de uno.
 */
export interface Session {
  readonly number: number;
  readonly slug: string;
  readonly title: string;
  /** El concepto de la sesión, en una frase. */
  readonly concept: string;
  readonly available: boolean;
}

export const SESSIONS: readonly Session[] = [
  {
    number: 0,
    slug: 's00',
    title: 'TypeScript',
    concept: 'Tipos, uniones, opcionales y genéricos.',
    available: true,
  },
  {
    number: 1,
    slug: 's01',
    title: 'Primer componente',
    concept: 'Un componente standalone y los cuatro tipos de binding.',
    available: true,
  },
  {
    number: 2,
    slug: 's02',
    title: 'Anatomía de un componente',
    concept: 'input, output, model y ng-content.',
    available: true,
  },
  {
    number: 3,
    slug: 's03',
    title: 'Signals y control flow',
    concept: 'signal, computed, @if, @for y @switch.',
    available: true,
  },
  { number: 4, slug: 's04', title: 'Directivas y pipes', concept: 'Transformar sin tocar el componente.', available: false },
  { number: 5, slug: 's05', title: 'Inyección de dependencias', concept: 'Servicios, inject() y tokens.', available: false },
  { number: 6, slug: 's06', title: 'Reactividad', concept: 'Observables, operadores y debounce.', available: false },
  { number: 7, slug: 's07', title: 'HttpClient', concept: 'Pedir datos y manejar los tres estados.', available: false },
  { number: 8, slug: 's08', title: 'Reactive Forms', concept: 'Formularios con validación de verdad.', available: false },
  { number: 9, slug: 's09', title: 'Routing', concept: 'Rutas, parámetros y guards.', available: false },
  { number: 10, slug: 's10', title: 'WebSockets y cierre', concept: 'Tiempo real, OnPush y NgModules como legado.', available: false },
];

export const AVAILABLE_SESSIONS = SESSIONS.filter((s) => s.available);
