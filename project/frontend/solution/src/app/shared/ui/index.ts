/**
 * Primitivas del sistema de diseño.
 *
 * `<app-race-card>` no está acá: no es una primitiva, es una pieza del
 * dominio y vive en `features/races/`. La regla es la de siempre —
 * `shared/` no sabe qué es una carrera.
 */

export * from './badge/badge.component';
export * from './button/button.component';
export * from './empty-state/empty-state.component';
export * from './logo/logo.component';
export * from './silk/silk.component';
export * from './silk/silk.util';
export * from './skeleton/skeleton.component';
