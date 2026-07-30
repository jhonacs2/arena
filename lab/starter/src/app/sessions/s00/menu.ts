/**
 * El menú de Café Compilado — el ejercicio de S0.
 *
 * Este archivo **compila y funciona**. Ese es justamente el problema: también
 * compila cuando está mal. Un tamaño escrito `'mediana'`, un café sin nota de
 * cata al que le inventamos una cadena vacía, un menú vacío que rompe en
 * tiempo de ejecución — nada de eso lo frena nadie hoy.
 *
 * Tu trabajo es apretar los tipos hasta que el compilador se haga cargo. Cada
 * lugar donde falta está marcado con `TODO(S0)`.
 *
 * La pantalla de al lado (`s00.component.ts`) usa estas funciones y **no hay
 * que tocarla**: los componentes son el tema de la clase que viene.
 */

/**
 * Los tres tamaños que se venden.
 *
 * TODO(S0) · 1 — `string` acepta `'mediana'`, `'venti'` y `'asdf'`. El tipo
 * tiene que decir «una de estas tres», no «una cadena cualquiera».
 */
export type CoffeeSize = string;

/**
 * Un café del menú.
 *
 * TODO(S0) · 2 — Faltan dos cosas:
 *   a) los campos se pueden reasignar; el menú no se edita en el lugar
 *   b) `notes` no la tienen todos, y el tipo dice que sí
 */
export interface Coffee {
  id: string;
  name: string;
  origin: string;
  /** Precio del tamaño mediano, en pesos enteros. */
  price: number;
  /** Nota de cata. Los que no tienen llevan cadena vacía. Por ahora. */
  notes: string;
}

/** Una línea de un pedido: qué café, en qué tamaño, cuántos. */
export interface OrderLine {
  coffee: Coffee;
  size: CoffeeSize;
  quantity: number;
}

export interface Order {
  customer: string;
  lines: OrderLine[];
}

/**
 * Una porción de una lista más larga.
 *
 * TODO(S0) · 6 (extensión) — Sirve solo para cafés. El día que haya que
 * paginar pedidos hay que escribir `OrderPage`, igual pero con otro contenido.
 * Hacelo genérico y que `firstPage` lo acompañe.
 */
export interface CoffeePage {
  items: Coffee[];
  /** Total que hay en la lista completa, no en esta página. */
  total: number;
}

/**
 * Un café que todavía no se guardó: el id lo asigna la caja al registrarlo.
 *
 * TODO(S0) · 7 (extensión) — Está copiado a mano de `Coffee`. El día que
 * `Coffee` gane un campo, esta copia se queda vieja y nadie se entera.
 * Derivalo del original.
 */
export interface CoffeeDraft {
  name: string;
  origin: string;
  price: number;
  notes: string;
}

/**
 * TODO(S0) · 3 — La lista se puede reordenar y se le puede hacer `push`.
 * Y fijate en los dos `notes: ''`: son cafés sin nota de cata a los que les
 * inventamos una. Cuando `notes` sea opcional, esos dos campos se borran.
 */
export const MENU: Coffee[] = [
  { id: 'c1', name: 'Yirgacheffe', origin: 'Etiopía', price: 42, notes: 'cítrico y floral' },
  { id: 'c2', name: 'Huila', origin: 'Colombia', price: 38, notes: '' },
  { id: 'c3', name: 'Cerrado', origin: 'Brasil', price: 30, notes: 'chocolate y nuez' },
  { id: 'c4', name: 'Antigua', origin: 'Guatemala', price: 45, notes: '' },
];

/**
 * Cuánto multiplica cada tamaño al precio del mediano.
 *
 * TODO(S0) · 4 — Ese `return 1` del final se traga cualquier cosa: pedile un
 * `'mediana'` y te cobra el precio del mediano sin decir nada. Cuando
 * `CoffeeSize` sea una unión, esto se puede escribir con un `switch` que cubra
 * los tres casos y sin ningún caso por defecto.
 */
export function sizeFactor(size: CoffeeSize): number {
  if (size === 'chico') return 0.8;
  if (size === 'grande') return 1.3;
  return 1;
}

/** Lo que sale un café en un tamaño, redondeado a peso entero. */
export function priceFor(coffee: Coffee, size: CoffeeSize): number {
  return Math.round(coffee.price * sizeFactor(size));
}

/**
 * El más barato del menú.
 *
 * TODO(S0) · 5 — Ese `!` le dice al compilador «confiá en mí, siempre hay al
 * menos uno». Es mentira: `cheapest([])` explota en tiempo de ejecución.
 * Decí la verdad en el tipo y sacá el `!`.
 */
export function cheapest(menu: readonly Coffee[]): Coffee {
  let best = menu[0]!;
  for (const coffee of menu) {
    if (coffee.price < best.price) best = coffee;
  }
  return best;
}

/**
 * La descripción de un café para el pizarrón.
 *
 * Cuando `notes` sea opcional, la comparación contra la cadena vacía deja de
 * tener sentido: hay que preguntar si existe.
 */
export function describe(coffee: Coffee): string {
  const base = `${coffee.name} · ${coffee.origin}`;
  if (coffee.notes === '') {
    return base;
  }
  return `${base} — ${coffee.notes}`;
}

/** Lo que sale un pedido entero. */
export function orderTotal(order: Order): number {
  return order.lines.reduce((sum, line) => sum + priceFor(line.coffee, line.size) * line.quantity, 0);
}

/** Las primeras `size` cosas de una lista, con el total de la lista completa. */
export function firstPage(items: readonly Coffee[], size: number): CoffeePage {
  return { items: items.slice(0, size), total: items.length };
}

/** Registra un borrador en el menú, asignándole el id que da la caja. */
export function withId(draft: CoffeeDraft, id: string): Coffee {
  return { ...draft, id };
}
