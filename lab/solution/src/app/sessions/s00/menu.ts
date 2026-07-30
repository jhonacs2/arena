/**
 * El menú de Café Compilado — los tipos de S0.
 *
 * Este archivo es el material de la sesión 0. No hay ni un componente acá: son
 * tipos y funciones de TypeScript puro. La pantalla de al lado los muestra,
 * pero lo que se aprende hoy vive en este archivo.
 *
 * El orden en que aparecen las cosas es el orden en que se dan en clase:
 * alias y uniones, interfaces, opcionales, narrowing, genéricos y utility
 * types.
 */

/**
 * Los tres tamaños que se venden.
 *
 * Es una **unión de literales**: el tipo no es «una cadena cualquiera», es
 * «una de estas tres cadenas». `'mediana'` no compila, y ese es todo el punto.
 */
export type CoffeeSize = 'chico' | 'mediano' | 'grande';

/**
 * Un café del menú.
 *
 * Todo `readonly`: los datos del menú no se editan en el lugar. Es regla del
 * curso desde el día uno, y lo que hace que el compilador la sostenga.
 */
export interface Coffee {
  readonly id: string;
  readonly name: string;
  readonly origin: string;
  /** Precio del tamaño mediano, en pesos enteros. */
  readonly price: number;
  /**
   * Nota de cata. El `?` dice que **puede no estar**: solo algunos cafés la
   * tienen, y quien la lea se tiene que hacer cargo de que sea `undefined`.
   */
  readonly notes?: string;
}

/** Una línea de un pedido: qué café, en qué tamaño, cuántos. */
export interface OrderLine {
  readonly coffee: Coffee;
  readonly size: CoffeeSize;
  readonly quantity: number;
}

export interface Order {
  readonly customer: string;
  readonly lines: readonly OrderLine[];
}

/**
 * Una porción de una lista más larga.
 *
 * `T` es un hueco: `Page<Coffee>` es una página de cafés y `Page<Order>` una
 * de pedidos, sin escribir dos interfaces casi iguales. En el hipódromo existe
 * este mismo tipo, con `Page<Race>` y `Page<Bet>`.
 */
export interface Page<T> {
  readonly items: readonly T[];
  /** Total que hay en la lista completa, no en esta página. */
  readonly total: number;
}

/**
 * Un café que todavía no se guardó: es igual a `Coffee` pero sin `id`, porque
 * el id lo asigna la caja al registrarlo.
 *
 * `Omit` lo deriva del original. Escribir los cuatro campos otra vez a mano
 * también funcionaría, hasta el día en que `Coffee` gana un campo y esta copia
 * se queda vieja sin que nadie se entere.
 */
export type CoffeeDraft = Omit<Coffee, 'id'>;

export const MENU: readonly Coffee[] = [
  { id: 'c1', name: 'Yirgacheffe', origin: 'Etiopía', price: 42, notes: 'cítrico y floral' },
  { id: 'c2', name: 'Huila', origin: 'Colombia', price: 38 },
  { id: 'c3', name: 'Cerrado', origin: 'Brasil', price: 30, notes: 'chocolate y nuez' },
  { id: 'c4', name: 'Antigua', origin: 'Guatemala', price: 45 },
];

/**
 * Cuánto multiplica cada tamaño al precio del mediano.
 *
 * El `switch` cubre los tres casos de la unión, así que TypeScript sabe que la
 * función siempre devuelve algo. Agregale un cuarto tamaño al tipo y este
 * `switch` deja de compilar: el compilador te lleva al lugar exacto donde
 * falta decidir. A eso se le dice **exhaustividad**.
 */
export function sizeFactor(size: CoffeeSize): number {
  switch (size) {
    case 'chico':
      return 0.8;
    case 'mediano':
      return 1;
    case 'grande':
      return 1.3;
  }
}

/** Lo que sale un café en un tamaño, redondeado a peso entero. */
export function priceFor(coffee: Coffee, size: CoffeeSize): number {
  return Math.round(coffee.price * sizeFactor(size));
}

/**
 * El más barato del menú.
 *
 * Devuelve `undefined` con un menú vacío. No es una cortesía: es la verdad, y
 * el tipo la dice en voz alta. Quien lo use tiene que decidir qué mostrar
 * cuando no hay nada.
 */
export function cheapest(menu: readonly Coffee[]): Coffee | undefined {
  let best: Coffee | undefined = undefined;
  for (const coffee of menu) {
    if (best === undefined || coffee.price < best.price) best = coffee;
  }
  return best;
}

/**
 * La descripción de un café para el pizarrón.
 *
 * `coffee.notes` es `string | undefined`. Adentro del `if` TypeScript ya sabe
 * que es `string` y deja concatenarla: eso es **narrowing**.
 */
export function describe(coffee: Coffee): string {
  const base = `${coffee.name} · ${coffee.origin}`;
  if (coffee.notes === undefined) {
    return base;
  }
  return `${base} — ${coffee.notes}`;
}

/** Lo que sale un pedido entero. */
export function orderTotal(order: Order): number {
  return order.lines.reduce((sum, line) => sum + priceFor(line.coffee, line.size) * line.quantity, 0);
}

/**
 * Las primeras `size` cosas de una lista, con el total de la lista completa.
 *
 * La función también es genérica: `firstPage(MENU, 3)` devuelve `Page<Coffee>`
 * sin que haga falta decirlo, porque TypeScript lo deduce del argumento.
 */
export function firstPage<T>(items: readonly T[], size: number): Page<T> {
  return { items: items.slice(0, size), total: items.length };
}

/** Registra un borrador en el menú, asignándole el id que da la caja. */
export function withId(draft: CoffeeDraft, id: string): Coffee {
  return { ...draft, id };
}
