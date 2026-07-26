/**
 * Índice de las sesiones del lab.
 *
 * De acá sale la **barra lateral**. Arranca todo en `disponible: false`: la
 * barra está vacía a propósito, y la vas a ver crecer a medida que hagas cada
 * sesión.
 *
 * Que la navegación no aparezca sola es parte de lo que se aprende: las rutas
 * las declara alguien. Ese alguien vas a ser vos.
 *
 * Para habilitar una sesión hacen falta DOS cosas:
 *   1. Poner `disponible: true` acá      → aparece en la barra lateral
 *   2. Sumar la ruta en `app.routes.ts`  → el enlace lleva a algún lado
 *
 * Si hacés solo la primera, vas a tener un enlace que no va a ninguna parte.
 * Si hacés solo la segunda, una página que existe pero que no se ve.
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
  // TODO(S1): poner `disponible: true` cuando tu ruta /s01 funcione.
  {
    number: 1,
    slug: 's01',
    title: 'Primer componente',
    concept: 'Un componente standalone y los cuatro tipos de binding.',
    available: false,
  },
  { number: 2, slug: 's02', title: 'Anatomía de un componente', concept: 'input, output y ng-content.', available: false },
  { number: 3, slug: 's03', title: 'Signals y control flow', concept: 'signal, computed, @if y @for.', available: false },
  { number: 4, slug: 's04', title: 'Directivas y pipes', concept: 'Transformar sin tocar el componente.', available: false },
  { number: 5, slug: 's05', title: 'Inyección de dependencias', concept: 'Servicios, inject() y tokens.', available: false },
  { number: 6, slug: 's06', title: 'Reactividad', concept: 'Observables, operadores y debounce.', available: false },
  { number: 7, slug: 's07', title: 'HttpClient', concept: 'Pedir datos y manejar los tres estados.', available: false },
  { number: 8, slug: 's08', title: 'Reactive Forms', concept: 'Formularios con validación de verdad.', available: false },
  { number: 9, slug: 's09', title: 'Routing', concept: 'Rutas, parámetros y guards.', available: false },
  { number: 10, slug: 's10', title: 'WebSockets', concept: 'Tiempo real, zona y OnPush.', available: false },
  { number: 11, slug: 's11', title: 'Producción', concept: 'NgModules como legado y build de producción.', available: false },
];

export const AVAILABLE_SESSIONS = SESSIONS.filter((s) => s.available);
