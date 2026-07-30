import { MENU, cheapest, describe as describeCoffee, firstPage, orderTotal, priceFor, sizeFactor, withId } from './menu';
import type { Coffee, Order } from './menu';

/**
 * Estos tests pasan desde el minuto cero: verifican lo que `menu.ts` ya hace
 * bien, para que sirvan de red mientras apretás los tipos. Si alguno se pone
 * en rojo mientras trabajás, te está avisando que un cambio de tipo cambió el
 * comportamiento — andá a mirarlo antes de seguir.
 *
 * Lo que **no** verifican son los tipos. Eso lo verifica el compilador:
 *
 *   npx tsc --noEmit
 */
describe('menu (S0)', () => {
  const espresso: Coffee = { id: 'x', name: 'Prueba', origin: 'Perú', price: 100, notes: '' };

  it('cada tamaño tiene su multiplicador', () => {
    expect(sizeFactor('chico')).toBe(0.8);
    expect(sizeFactor('mediano')).toBe(1);
    expect(sizeFactor('grande')).toBe(1.3);
  });

  it('el precio sale redondeado a peso entero', () => {
    expect(priceFor(espresso, 'chico')).toBe(80);
    expect(priceFor(espresso, 'mediano')).toBe(100);
    expect(priceFor(espresso, 'grande')).toBe(130);
  });

  it('el más barato del menú es el Cerrado', () => {
    expect(cheapest(MENU)?.name).toBe('Cerrado');
  });

  it('la descripción incluye la nota de cata cuando existe', () => {
    const yirgacheffe = MENU.find((coffee) => coffee.name === 'Yirgacheffe');
    const huila = MENU.find((coffee) => coffee.name === 'Huila');

    expect(yirgacheffe).toBeDefined();
    expect(huila).toBeDefined();
    expect(describeCoffee(yirgacheffe as Coffee)).toContain('cítrico y floral');
    expect(describeCoffee(huila as Coffee)).toBe('Huila · Colombia');
  });

  it('el total del pedido suma cada línea por su cantidad', () => {
    const order: Order = {
      customer: 'Ana',
      lines: [
        { coffee: espresso, size: 'grande', quantity: 2 },
        { coffee: espresso, size: 'chico', quantity: 1 },
      ],
    };

    expect(orderTotal(order)).toBe(130 * 2 + 80);
  });

  it('la página trae los primeros y el total de la lista entera', () => {
    const page = firstPage(MENU, 3);

    expect(page.items.length).toBe(3);
    expect(page.total).toBe(MENU.length);
  });

  it('el borrador se convierte en café al recibir su id', () => {
    const saved = withId({ name: 'Nuevo', origin: 'Kenia', price: 50, notes: '' }, 'c9');

    expect(saved.id).toBe('c9');
    expect(saved.name).toBe('Nuevo');
  });
});
