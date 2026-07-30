/**
 * Los datos de S2. Café Compilado otra vez, pero con dos componentes.
 *
 * Cada sesión del lab tiene sus propios tipos a propósito: así se puede hacer
 * la sesión 2 aunque la 0 haya quedado a medio terminar, y nadie rompe una
 * clase tocando otra.
 */

export interface Coffee {
  readonly id: string;
  readonly name: string;
  readonly origin: string;
  readonly price: number;
  readonly available: boolean;
}

/** Lo que el hijo le manda al padre cuando alguien pide un café. */
export interface OrderRequest {
  readonly coffee: Coffee;
  readonly quantity: number;
}

export const MENU: readonly Coffee[] = [
  { id: 'c1', name: 'Yirgacheffe', origin: 'Etiopía', price: 42, available: true },
  { id: 'c2', name: 'Huila', origin: 'Colombia', price: 38, available: true },
  { id: 'c3', name: 'Cerrado', origin: 'Brasil', price: 30, available: false },
  { id: 'c4', name: 'Antigua', origin: 'Guatemala', price: 45, available: true },
];
