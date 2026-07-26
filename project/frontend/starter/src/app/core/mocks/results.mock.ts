/**
 * Podios de las carreras terminadas. `payouts` se completa por usuario.
 *
 * GENERADO desde docs/contract/seed/ por scripts/gen-mocks.mjs — no editar a
 * mano. Es el MISMO dataset que carga el backend Go, así que un componente
 * escrito contra estos datos funciona sin cambios cuando en S7 se conecta al
 * servidor real.
 */

import type { RaceResult } from '../models';

/**
 * Las fechas del dataset están ancladas a un instante fijo y se desplazan al
 * cargar el módulo, igual que hace el backend: `docs/contract/README.md`,
 * regla de rebase. Así la carrera en vivo siempre está en vivo y la próxima
 * siempre está por venir, se dé la clase el mes que se dé.
 */
const ANCHOR = Date.parse('2026-03-14T12:00:00Z');
const OFFSET = Date.now() - ANCHOR;

function rebase(iso: string): string {
  return new Date(Date.parse(iso) + OFFSET).toISOString().replace(/\.\d{3}Z$/, 'Z');
}

export const RESULTS: readonly RaceResult[] = [
  {
    raceId: 'race_001',
    finishedAt: '2026-03-13T20:17:12Z',
    podium: [
      {
        place: 1,
        horseId: 'hrs_003',
        horseName: 'Viento Norte',
        number: 3,
        odds: 2.4,
      },
      {
        place: 2,
        horseId: 'hrs_001',
        horseName: 'Relámpago',
        number: 1,
        odds: 3.2,
      },
      {
        place: 3,
        horseId: 'hrs_006',
        horseName: 'Alborada',
        number: 6,
        odds: 7.25,
      },
    ],
  },
  {
    raceId: 'race_002',
    finishedAt: '2026-03-13T21:02:08Z',
    podium: [
      {
        place: 1,
        horseId: 'hrs_010',
        horseName: 'Media Luna',
        number: 4,
        odds: 3.5,
      },
      {
        place: 2,
        horseId: 'hrs_008',
        horseName: 'Trueno Manso',
        number: 2,
        odds: 2.8,
      },
      {
        place: 3,
        horseId: 'hrs_007',
        horseName: 'Centella',
        number: 1,
        odds: 4.1,
      },
    ],
  },
  {
    raceId: 'race_003',
    finishedAt: '2026-03-14T10:42:03Z',
    podium: [
      {
        place: 1,
        horseId: 'hrs_018',
        horseName: 'Guaraní',
        number: 5,
        odds: 11,
      },
      {
        place: 2,
        horseId: 'hrs_015',
        horseName: 'Pampero',
        number: 2,
        odds: 2.1,
      },
      {
        place: 3,
        horseId: 'hrs_017',
        horseName: 'Ciclón',
        number: 4,
        odds: 4.25,
      },
    ],
  },
  {
    raceId: 'race_004',
    finishedAt: '2026-03-14T11:22:19Z',
    podium: [
      {
        place: 1,
        horseId: 'hrs_022',
        horseName: 'Chasqui',
        number: 3,
        odds: 2.5,
      },
      {
        place: 2,
        horseId: 'hrs_025',
        horseName: 'Mistral',
        number: 6,
        odds: 4.4,
      },
      {
        place: 3,
        horseId: 'hrs_020',
        horseName: 'Cascabel',
        number: 1,
        odds: 3.6,
      },
    ],
  },
].map((result) => ({
  ...result,
  finishedAt: rebase(result.finishedAt),
  payouts: [],
})) as readonly RaceResult[];
