/**
 * Los 12 usuarios del dataset. Contraseña de todos: Carrera123!
 *
 * GENERADO desde docs/contract/seed/ por scripts/gen-mocks.mjs — no editar a
 * mano. Es el MISMO dataset que carga el backend Go, así que un componente
 * escrito contra estos datos funciona sin cambios cuando en S7 se conecta al
 * servidor real.
 */

import type { User } from '../models';

export const USERS: readonly User[] = [
  {
    id: 'usr_001',
    email: 'ana@hipodromo.test',
    displayName: 'Ana Robles',
    balance: 5000,
    emailVerified: true,
  },
  {
    id: 'usr_002',
    email: 'bruno@hipodromo.test',
    displayName: 'Bruno Salas',
    balance: 2350,
    emailVerified: true,
  },
  {
    id: 'usr_003',
    email: 'caro@hipodromo.test',
    displayName: 'Carolina Vega',
    balance: 10000,
    emailVerified: false,
  },
  {
    id: 'usr_004',
    email: 'diego@hipodromo.test',
    displayName: 'Diego Paredes',
    balance: 3120,
    emailVerified: true,
  },
  {
    id: 'usr_005',
    email: 'elena@hipodromo.test',
    displayName: 'Elena Quiroga',
    balance: 7840,
    emailVerified: true,
  },
  {
    id: 'usr_006',
    email: 'facundo@hipodromo.test',
    displayName: 'Facundo Ibarra',
    balance: 1290,
    emailVerified: true,
  },
  {
    id: 'usr_007',
    email: 'gabriela@hipodromo.test',
    displayName: 'Gabriela Nieto',
    balance: 6475,
    emailVerified: true,
  },
  {
    id: 'usr_008',
    email: 'hugo@hipodromo.test',
    displayName: 'Hugo Lemos',
    balance: 980,
    emailVerified: true,
  },
  {
    id: 'usr_009',
    email: 'irene@hipodromo.test',
    displayName: 'Irene Castro',
    balance: 4310,
    emailVerified: true,
  },
  {
    id: 'usr_010',
    email: 'joaquin@hipodromo.test',
    displayName: 'Joaquín Ferrer',
    balance: 8905,
    emailVerified: true,
  },
  {
    id: 'usr_011',
    email: 'karina@hipodromo.test',
    displayName: 'Karina Villalba',
    balance: 2760,
    emailVerified: true,
  },
  {
    id: 'usr_012',
    email: 'lautaro@hipodromo.test',
    displayName: 'Lautaro Mendive',
    balance: 5540,
    emailVerified: true,
  },
];

/** El usuario por defecto de las demos, mientras no haya sesión real. */
export const CURRENT_USER: User = USERS[0]!;
