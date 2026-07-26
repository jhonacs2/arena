/**
 * Punto de entrada de los modelos.
 *
 * El resto de la app importa desde `core/models`, no desde cada archivo. Así
 * mover un tipo de un archivo a otro no toca veinte imports.
 */

export * from './api.model';
export * from './bet.model';
export * from './race.model';
export * from './user.model';
