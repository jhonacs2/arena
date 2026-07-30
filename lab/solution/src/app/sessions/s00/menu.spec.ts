import { MENU, cheapest, describe as describeCoffee, firstPage, orderTotal, priceFor, sizeFactor, withId } from './menu';
import type { Coffee, Order } from './menu';

/**
 * Los tests de S0 son la referencia de qué tiene que lograr la Misión 1.
 *
 * Testean **comportamiento**, no tipos: lo que verifica los tipos es
 * `npx tsc --noEmit`, y esa parte del ejercicio se comprueba con el
 * compilador, no con Jasmine.
 */
describe('menu (S0)', () => {
  const espresso: Coffee = { id: 'x', name: 'Prueba', origin: 'Perú', price: 100 };

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

  it('con un menú vacío no hay más barato', () => {
    expect(cheapest([])).toBeUndefined();
  });

  it('la descripción incluye la nota de cata solo cuando existe', () => {
    const withNotes = MENU.find((coffee) => coffee.notes !== undefined);
    const withoutNotes = MENU.find((coffee) => coffee.notes === undefined);

    expect(withNotes).toBeDefined();
    expect(withoutNotes).toBeDefined();
    expect(describeCoffee(withNotes as Coffee)).toContain('—');
    expect(describeCoffee(withoutNotes as Coffee)).not.toContain('—');
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
    const saved = withId({ name: 'Nuevo', origin: 'Kenia', price: 50 }, 'c9');

    expect(saved.id).toBe('c9');
    expect(saved.name).toBe('Nuevo');
  });
});
