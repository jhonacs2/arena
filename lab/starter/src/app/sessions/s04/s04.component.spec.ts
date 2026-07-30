import { ComponentFixture, TestBed } from '@angular/core/testing';

import { resetCallCount } from './call-count.pipe';
import { S04Component } from './s04.component';

/**
 * Estos tests pasan desde el minuto cero: verifican lo que la pantalla ya hace.
 *
 * Al terminar el ejercicio el formateo va a vivir en un pipe y el resaltado en
 * una directiva, y estos mismos tests **tienen que seguir pasando**: son la red
 * que avisa si el cambio movió el comportamiento en lugar de solo moverlo de
 * lugar.
 */
describe('S4 · la carta', () => {
  let fixture: ComponentFixture<S04Component>;

  const host = (): HTMLElement => fixture.nativeElement as HTMLElement;
  const cards = (): HTMLElement[] => Array.from(host().querySelectorAll('.card'));

  beforeEach(async () => {
    resetCallCount();
    await TestBed.configureTestingModule({ imports: [S04Component] }).compileComponents();
    fixture = TestBed.createComponent(S04Component);
    fixture.detectChanges();
  });

  it('muestra los cuatro cafés', () => {
    expect(cards().length).toBe(4);
  });

  it('los precios llevan separador de miles', () => {
    expect(host().textContent).toContain('$ 4.200');
    expect(host().textContent).toContain('USD 4.200');
  });

  it('solo el café del día queda resaltado', () => {
    const highlighted = cards().filter((card) => card.classList.contains('is-highlighted'));

    expect(highlighted.length).toBe(1);
    expect(highlighted[0]?.textContent).toContain('Yirgacheffe');
    expect(highlighted[0]?.getAttribute('data-highlight-label')).toBe('Café del día');
  });

  it('el resto no lleva rótulo', () => {
    const plain = cards().filter((card) => !card.classList.contains('is-highlighted'));

    expect(plain.length).toBe(3);
    expect(plain[0]?.getAttribute('data-highlight-label')).toBeNull();
  });

  it('el puntaje se dibuja con una marca por punto', () => {
    const ratings = cards().map((card) => card.querySelectorAll('.bean').length);
    expect(ratings).toEqual([5, 4, 3, 4]);
  });

  it('el café sin stock se ve apagado', () => {
    const soldout = cards().filter((card) => card.classList.contains('card--soldout'));
    expect(soldout.length).toBe(1);
    expect(soldout[0]?.textContent).toContain('Cerrado');
  });
});
