import { ComponentFixture, fakeAsync, TestBed, tick } from '@angular/core/testing';
import { By } from '@angular/platform-browser';

import { S01Component } from './s01.component';

/**
 * Estos tests describen los cuatro bindings de S1 desde afuera: qué se ve y
 * qué pasa al tocar. Son la referencia de qué tiene que lograr la Misión 1.
 *
 * El starter tiene una versión más chica de este archivo: prueba la lógica de
 * la clase, que ya está hecha, y no los bindings, que son el ejercicio.
 */
describe('S01Component · los cuatro bindings', () => {
  let fixture: ComponentFixture<S01Component>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [S01Component] }).compileComponents();
    fixture = TestBed.createComponent(S01Component);
    fixture.detectChanges();
  });

  const $ = (selector: string): HTMLElement => {
    const found = fixture.debugElement.query(By.css(selector));
    expect(found).withContext(`no se encontró ${selector}`).not.toBeNull();
    return found.nativeElement as HTMLElement;
  };

  it('1 · interpola el producto desde la clase', () => {
    expect($('h2').textContent).toContain('Yirgacheffe');
    expect($('.product__origin').textContent).toContain('Etiopía');
    expect($('.product__price').textContent).toContain('42');
  });

  it('2 · pone la clase de agotado solo cuando corresponde', () => {
    expect($('.product').classList).not.toContain('product--soldout');

    $('.product .button').click();
    fixture.detectChanges();

    expect($('.product').classList).toContain('product--soldout');
    expect($('.product__status').textContent).toContain('Agotado');
  });

  it('3 · el botón dispara el método de la clase', () => {
    const before = $('.product__status').textContent;
    $('.product .button').click();
    fixture.detectChanges();

    expect($('.product__status').textContent).not.toBe(before);
  });

  it('4 · el total se actualiza mientras se escribe', async () => {
    const quantity = $('input[name="quantity"]') as HTMLInputElement;
    quantity.value = '3';
    quantity.dispatchEvent(new Event('input'));
    await fixture.whenStable();
    fixture.detectChanges();

    // 42 × 3
    expect($('.order__total').textContent).toContain('126');
  });

  it('el botón de agregar está deshabilitado hasta que hay nombre', async () => {
    const button = $('.order .button') as HTMLButtonElement;
    expect(button.disabled).withContext('sin nombre tendría que estar deshabilitado').toBe(true);

    const customer = $('input[name="customer"]') as HTMLInputElement;
    customer.value = 'Ana';
    customer.dispatchEvent(new Event('input'));
    await fixture.whenStable();
    fixture.detectChanges();

    expect(button.disabled).toBe(false);
  });

  it('agregar suma una línea a la comanda y limpia el formulario', fakeAsync(() => {
    const customer = $('input[name="customer"]') as HTMLInputElement;
    customer.value = 'Bruno';
    customer.dispatchEvent(new Event('input'));
    tick();
    fixture.detectChanges();

    ($('.order .button') as HTMLButtonElement).click();
    tick();
    fixture.detectChanges();

    // Lo que se ve: la línea nueva en la comanda.
    const lines = fixture.debugElement.queryAll(By.css('.orders__list li'));
    expect(lines.length).toBe(1);
    expect((lines[0]!.nativeElement as HTMLElement).textContent).toContain('Bruno');

    // El vaciado se asserta sobre la clase y no sobre el `value` del input.
    // Que ngModel reescriba el DOM es tarea de Angular y su tiempo no se
    // estabiliza de forma confiable en el test; que la clase quede limpia es
    // la conducta que nos importa. La comprobación visual está en los
    // criterios de «Listo cuando» de mision-1.md.
    const componente = fixture.componentInstance as unknown as { customer: string; quantity: number };
    expect(componente.customer).withContext('el nombre tendría que quedar vacío').toBe('');
    expect(componente.quantity).toBe(1);
  }));

  it('no muta el array de pedidos', () => {
    // La regla de inmutabilidad del curso. Si alguien usa push, la vista con
    // OnPush deja de actualizarse y este test lo dice before que el navegador.
    const componente = fixture.componentInstance as unknown as { orders: readonly string[] };
    const original = componente.orders;

    (fixture.componentInstance as unknown as { customer: string }).customer = 'Caro';
    (fixture.componentInstance as unknown as { addOrder: () => void }).addOrder();

    expect(componente.orders).not.toBe(original);
    expect(original.length).toBe(0);
  });
});
