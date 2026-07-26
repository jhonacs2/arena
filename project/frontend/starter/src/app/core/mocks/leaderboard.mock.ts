/**
 * Tabla de posiciones. Calculada desde las apuestas liquidadas.
 *
 * GENERADO desde docs/contract/seed/ por scripts/gen-mocks.mjs — no editar a
 * mano. Es el MISMO dataset que carga el backend Go, así que un componente
 * escrito contra estos datos funciona sin cambios cuando en S7 se conecta al
 * servidor real.
 */

import type { LeaderboardEntry } from '../models';

export const LEADERBOARD_ALL: readonly LeaderboardEntry[] = [
  {
    rank: 1,
    userId: 'usr_005',
    displayName: 'Elena Quiroga',
    profit: 2100,
    bets: 3,
    wins: 2,
  },
  {
    rank: 2,
    userId: 'usr_010',
    displayName: 'Joaquín Ferrer',
    profit: 1790,
    bets: 4,
    wins: 3,
  },
  {
    rank: 3,
    userId: 'usr_009',
    displayName: 'Irene Castro',
    profit: 1700,
    bets: 2,
    wins: 1,
  },
  {
    rank: 4,
    userId: 'usr_004',
    displayName: 'Diego Paredes',
    profit: 1110,
    bets: 3,
    wins: 2,
  },
  {
    rank: 5,
    userId: 'usr_002',
    displayName: 'Bruno Salas',
    profit: 700,
    bets: 2,
    wins: 1,
  },
  {
    rank: 6,
    userId: 'usr_012',
    displayName: 'Lautaro Mendive',
    profit: 425,
    bets: 2,
    wins: 1,
  },
  {
    rank: 7,
    userId: 'usr_007',
    displayName: 'Gabriela Nieto',
    profit: 300,
    bets: 2,
    wins: 1,
  },
  {
    rank: 8,
    userId: 'usr_001',
    displayName: 'Ana Robles',
    profit: 255,
    bets: 4,
    wins: 2,
  },
  {
    rank: 9,
    userId: 'usr_006',
    displayName: 'Facundo Ibarra',
    profit: -300,
    bets: 3,
    wins: 1,
  },
  {
    rank: 10,
    userId: 'usr_011',
    displayName: 'Karina Villalba',
    profit: -350,
    bets: 2,
    wins: 0,
  },
  {
    rank: 11,
    userId: 'usr_008',
    displayName: 'Hugo Lemos',
    profit: -750,
    bets: 3,
    wins: 0,
  },
];

export const LEADERBOARD_DAILY: readonly LeaderboardEntry[] = [
  {
    rank: 1,
    userId: 'usr_009',
    displayName: 'Irene Castro',
    profit: 1700,
    bets: 2,
    wins: 1,
  },
  {
    rank: 2,
    userId: 'usr_002',
    displayName: 'Bruno Salas',
    profit: 1200,
    bets: 1,
    wins: 1,
  },
  {
    rank: 3,
    userId: 'usr_005',
    displayName: 'Elena Quiroga',
    profit: 850,
    bets: 2,
    wins: 1,
  },
  {
    rank: 4,
    userId: 'usr_007',
    displayName: 'Gabriela Nieto',
    profit: 600,
    bets: 1,
    wins: 1,
  },
  {
    rank: 5,
    userId: 'usr_001',
    displayName: 'Ana Robles',
    profit: 75,
    bets: 2,
    wins: 1,
  },
  {
    rank: 6,
    userId: 'usr_006',
    displayName: 'Facundo Ibarra',
    profit: -50,
    bets: 2,
    wins: 1,
  },
  {
    rank: 7,
    userId: 'usr_010',
    displayName: 'Joaquín Ferrer',
    profit: -50,
    bets: 2,
    wins: 1,
  },
  {
    rank: 8,
    userId: 'usr_004',
    displayName: 'Diego Paredes',
    profit: -200,
    bets: 1,
    wins: 0,
  },
  {
    rank: 9,
    userId: 'usr_011',
    displayName: 'Karina Villalba',
    profit: -200,
    bets: 1,
    wins: 0,
  },
  {
    rank: 10,
    userId: 'usr_008',
    displayName: 'Hugo Lemos',
    profit: -250,
    bets: 1,
    wins: 0,
  },
];
