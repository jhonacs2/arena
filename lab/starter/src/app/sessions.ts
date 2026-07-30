/**
 * Índice de las sesiones del lab.
 *
 * De acá sale la **barra lateral**. Arranca con una sola entrada, S0, y la vas
 * a ver crecer a medida que hagas cada sesión.
 *
 * Que la navegación no aparezca sola es parte de lo que se aprende: las rutas
 * las declara alguien. Ese alguien vas a ser vos, desde S1.
 *
 * Para habilitar una sesión hacen falta DOS cosas:
 *   1. Poner `available: true` acá       → aparece en la barra lateral
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
  // S0 viene hecha: su ruta y su pantalla se dan escritas, porque en la sesión
  // 0 todavía no viste ni componentes ni rutas. Es la única.
  {
    number: 0,
    slug: 's00',
    title: 'TypeScript',
    concept: 'Tipos, uniones, opcionales y genéricos.',
    available: true,
  },

  // TODO(S1): poner `available: true` cuando tu ruta /s01 funcione.
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
  { number: 10, slug: 's10', title: 'WebSockets y cierre', concept: 'Tiempo real, OnPush y NgModules como legado.', available: false },
];

export const AVAILABLE_SESSIONS = SESSIONS.filter((s) => s.available);
