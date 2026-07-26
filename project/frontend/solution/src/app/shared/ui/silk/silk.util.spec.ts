import { RACES } from '../../../core/mocks';

import { SILKS_GOLDEN } from './silks.golden';
import { silkDrawing, silkFromId, SILK_COLORS, type SilkColor } from './silk.util';

/** Luminancia de cada color, de docs/design/tokens.json. */
const LIGHTNESS: Readonly<Record<SilkColor, number>> = {
  black: 0.2,
  white: 0.98,
  red: 0.55,
  blue: 0.48,
  yellow: 0.85,
  green: 0.52,
  orange: 0.7,
  purple: 0.45,
  pink: 0.75,
  cyan: 0.78,
};

const ALL_HORSES = RACES.flatMap((race) => race.horses);

describe('silkFromId', () => {
  // ── El test que importa ────────────────────────────────────────────────
  // silks.golden.ts lo genera scripts/gen-silks-specimen.mjs, que es la
  // implementación de referencia en JavaScript y la que dibuja la hoja de
  // muestra. Si el port a TypeScript deriva, la seda de un caballo dejaría de
  // coincidir con la hoja que aprobó el instructor.
  it('coincide con la implementación de referencia para los 54 caballos', () => {
    for (const horse of ALL_HORSES) {
      expect(silkFromId(horse.id))
        .withContext(`${horse.name} (${horse.id})`)
        .toEqual(SILKS_GOLDEN[horse.id]!);
    }
  });

  it('es determinística', () => {
    expect(silkFromId('hrs_029')).toEqual(silkFromId('hrs_029'));
  });

  it('da sedas distintas para ids distintos', () => {
    expect(silkFromId('hrs_001')).not.toEqual(silkFromId('hrs_002'));
  });

  // ── Las reglas de la gramática ─────────────────────────────────────────
  it('nunca repite el color primario en el secundario', () => {
    for (const horse of ALL_HORSES) {
      const { primary, secondary } = silkFromId(horse.id);
      expect(primary).withContext(horse.name).not.toBe(secondary);
    }
  });

  it('separa los dos colores al menos 0.22 de luminancia', () => {
    // Sin esta regla salen sedas azul-sobre-violeta que a 24 px se ven como un
    // cuadrado sólido, y el sistema deja de identificar caballos.
    for (const horse of ALL_HORSES) {
      const { primary, secondary } = silkFromId(horse.id);
      const gap = Math.abs(LIGHTNESS[primary] - LIGHTNESS[secondary]);
      expect(gap).withContext(`${horse.name}: ${primary}/${secondary}`).toBeGreaterThanOrEqual(0.22);
    }
  });

  it('solo usa los diez colores registrados', () => {
    for (const horse of ALL_HORSES) {
      const { primary, secondary } = silkFromId(horse.id);
      expect(SILK_COLORS).toContain(primary);
      expect(SILK_COLORS).toContain(secondary);
    }
  });

  // ── Lo único que no se puede negociar ──────────────────────────────────
  it('no repite ninguna seda dentro de una misma carrera', () => {
    // Entre carreras distintas sí puede repetirse: en el hipódromo real las
    // sedas son del dueño, no del caballo. Dentro de una carrera, no: es donde
    // se ven lado a lado y hay que distinguirlas de un vistazo.
    for (const race of RACES) {
      const seen = new Map<string, string>();
      for (const horse of race.horses) {
        const spec = silkFromId(horse.id);
        const key = `${spec.primary}|${spec.secondary}|${spec.body}|${spec.sleeves}`;
        expect(seen.has(key))
          .withContext(`${race.name}: ${seen.get(key)} y ${horse.name} comparten seda`)
          .toBe(false);
        seen.set(key, horse.name);
      }
    }
  });
});

describe('silkDrawing', () => {
  it('devuelve figuras para todos los caballos', () => {
    for (const horse of ALL_HORSES) {
      const drawing = silkDrawing(horse.id);
      // Dos mangas como mínimo más el cuerpo: nunca puede salir vacío.
      expect(drawing.fills.length).withContext(horse.name).toBeGreaterThanOrEqual(3);
      expect(drawing.outlines.length).toBe(3);
    }
  });

  it('pinta todo con variables de seda, nunca con un color literal', () => {
    // Si se colara un hex, el color no cambiaría con el tema ni saldría de
    // docs/design/tokens.json, que es donde tiene que vivir la paleta.
    for (const horse of ALL_HORSES) {
      for (const shape of silkDrawing(horse.id).fills) {
        expect(shape.fill).withContext(`${horse.name}: ${shape.fill}`).toMatch(/^var\(--silk-[a-z]+\)$/);
      }
    }
  });
});
