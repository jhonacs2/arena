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
export interface Sesion {
  readonly numero: number;
  readonly slug: string;
  readonly titulo: string;
  /** El concepto de la sesión, en una frase. */
  readonly concepto: string;
  readonly disponible: boolean;
}

export const SESIONES: readonly Sesion[] = [
  // TODO(S1): poner `disponible: true` cuando tu ruta /s01 funcione.
  {
    numero: 1,
    slug: 's01',
    titulo: 'Primer componente',
    concepto: 'Un componente standalone y los cuatro tipos de binding.',
    disponible: false,
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
