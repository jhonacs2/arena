import { ComponentFixture, TestBed } from '@angular/core/testing';

import { S02Component } from './s02.component';

/**
 * Estos tests pasan desde el minuto cero: verifican lo que la pantalla ya hace.
 *
 * Al terminar el ejercicio la tarjeta va a ser un componente aparte, y estos
 * mismos tests **tienen que seguir pasando**: son la red que avisa si el
 * refactor cambió el comportamiento en vez de solo moverlo de lugar.
 */
describe('S2 · la carta', () => {
  let fixture: ComponentFixture<S02Component>;

  const host = (): HTMLElement => fixture.nativeElement as HTMLElement;
  const buttonWith = (label: string): HTMLButtonElement | undefined =>
    Array.from(host().querySelectorAll('button')).find((button) =>
      button.textContent?.includes(label),
    );

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [S02Component] }).compileComponents();
    fixture = TestBed.createComponent(S02Component);
    fixture.detectChanges();
  });

  it('muestra los cuatro cafés de la carta', () => {
    expect(host().textContent).toContain('Yirgacheffe');
    expect(host().textContent).toContain('Huila');
    expect(host().textContent).toContain('Cerrado');
    expect(host().textContent).toContain('Antigua');
  });

  it('marca el café del día', () => {
    expect(host().querySelector('.tag')?.textContent).toContain('Café del día');
  });

  it('el café sin stock no se puede pedir', () => {
    expect(buttonWith('Sin stock')?.disabled).toBe(true);
  });

  it('la comanda arranca vacía y se llena al pedir', () => {
    expect(host().textContent).toContain('Todavía no pidió nadie');

    buttonWith('Pedir')?.click();
    fixture.detectChanges();

    expect(host().textContent).toContain('1 × Yirgacheffe');
  });

  it('el total de una tarjeta sigue a su cantidad', () => {
    const plus = Array.from(host().querySelectorAll('button')).find((button) =>
      button.textContent?.includes('Agregar uno'),
    );

    plus?.click();
    fixture.detectChanges();

    expect(host().textContent).toContain('84');
  });
});
