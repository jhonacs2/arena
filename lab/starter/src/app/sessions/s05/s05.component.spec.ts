import { ComponentFixture, TestBed } from '@angular/core/testing';

import { S05Component } from './s05.component';

/**
 * Estos tests pasan desde el minuto cero: verifican lo que el mostrador ya hace.
 *
 * Al terminar el ejercicio la comanda va a vivir en un servicio y el cuaderno en
 * otro, y estos mismos tests **tienen que seguir pasando**.
 */
describe('S5 · el mostrador', () => {
  let fixture: ComponentFixture<S05Component>;

  const host = (): HTMLElement => fixture.nativeElement as HTMLElement;
  const buttonWith = (label: string): HTMLButtonElement | undefined =>
    Array.from(host().querySelectorAll('button')).find((button) =>
      button.textContent?.trim().startsWith(label),
    );

  const type = (index: number, text: string): void => {
    const input = host().querySelectorAll<HTMLInputElement>('input')[index];
    if (!input) return;
    input.value = text;
    input.dispatchEvent(new Event('input'));
    fixture.detectChanges();
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [S05Component] }).compileComponents();
    fixture = TestBed.createComponent(S05Component);
    fixture.detectChanges();
  });

  it('arranca sin pedidos', () => {
    expect(host().textContent).toContain('Todavía no tomó nadie');
  });

  it('tomar un pedido lo suma a la comanda', () => {
    type(0, 'Ana');
    buttonWith('Tomar pedido')?.click();
    fixture.detectChanges();

    expect(host().querySelector('.orders')?.textContent).toContain('Ana');
    expect(host().textContent).toContain('Último: Ana');
  });

  it('no toma un pedido sin cliente', () => {
    expect(buttonWith('Tomar pedido')?.disabled).toBe(true);
  });

  it('quitar un pedido lo saca', () => {
    type(0, 'Beto');
    buttonWith('Tomar pedido')?.click();
    fixture.detectChanges();

    buttonWith('Quitar')?.click();
    fixture.detectChanges();

    expect(host().textContent).toContain('Todavía no tomó nadie');
  });

  it('el cuaderno anota y borra', () => {
    type(1, 'Falta leche');
    buttonWith('Anotar')?.click();
    fixture.detectChanges();

    expect(host().querySelector('.notes')?.textContent).toContain('Falta leche');

    buttonWith('Borrar')?.click();
    fixture.detectChanges();

    expect(host().textContent).toContain('Cuaderno vacío');
  });
});
