/**
 * Los datos de S3. El tablero de la comanda de Café Compilado.
 *
 * Cada sesión del lab tiene sus propios tipos: así se puede hacer la sesión 3
 * aunque la 2 haya quedado a medias.
 */

export type OrderStatus = 'pending' | 'ready' | 'served';

/** El filtro incluye un valor que no es un estado: «todas». */
export type OrderFilter = OrderStatus | 'all';

export interface Order {
  readonly id: string;
  readonly customer: string;
  readonly coffee: string;
  readonly quantity: number;
  readonly price: number;
  readonly status: OrderStatus;
}

export const STATUS_LABELS: Record<OrderStatus, string> = {
  pending: 'Pendiente',
  ready: 'Lista',
  served: 'Entregada',
};

export const FILTER_LABELS: Record<OrderFilter, string> = {
  all: 'Todas',
  pending: 'Pendientes',
  ready: 'Listas',
  served: 'Entregadas',
};

export const INITIAL_ORDERS: readonly Order[] = [
  { id: 'o1', customer: 'Ana', coffee: 'Yirgacheffe', quantity: 2, price: 42, status: 'pending' },
  { id: 'o2', customer: 'Beto', coffee: 'Huila', quantity: 1, price: 38, status: 'ready' },
  { id: 'o3', customer: 'Carla', coffee: 'Cerrado', quantity: 3, price: 30, status: 'served' },
  { id: 'o4', customer: 'Dante', coffee: 'Antigua', quantity: 1, price: 45, status: 'pending' },
  { id: 'o5', customer: 'Eva', coffee: 'Yirgacheffe', quantity: 1, price: 42, status: 'ready' },
];

/** Lo que sale una línea de la comanda. */
export function lineTotal(order: Order): number {
  return order.quantity * order.price;
}

/** El siguiente estado del ciclo. Entregada no vuelve. */
export function nextStatus(status: OrderStatus): OrderStatus {
  switch (status) {
    case 'pending':
      return 'ready';
    case 'ready':
      return 'served';
    case 'served':
      return 'served';
  }
}
