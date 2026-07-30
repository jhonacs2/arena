import { ComponentFixture, TestBed } from '@angular/core/testing';

import { S03Component } from './s03.component';

/**
 * Estos tests pasan desde el minuto cero: verifican lo que el tablero ya hace.
 *
 * Al terminar el ejercicio la comanda va a vivir en un `signal` y lo que se ve
 * va a salir de un `computed`, y estos mismos tests **tienen que seguir
 * pasando**: son la red que avisa si el cambio rompió el comportamiento en vez
 * de solo cambiar cómo está escrito.
 */
describe('S3 · el tablero de la comanda', () => {
  let fixture: ComponentFixture<S03Component>;

  const host = (): HTMLElement => fixture.nativeElement as HTMLElement;
  const rows = (): HTMLElement[] => Array.from(host().querySelectorAll('.order'));
  const buttonWith = (label: string): HTMLButtonElement | undefined =>
    Array.from(host().querySelectorAll('button')).find((button) =>
      button.textContent?.trim().startsWith(label),
    );

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [S03Component] }).compileComponents();
    fixture = TestBed.createComponent(S03Component);
    fixture.detectChanges();
  });

  it('muestra las cinco comandas iniciales', () => {
    expect(rows().length).toBe(5);
    expect(host().textContent).toContain('Ana');
    expect(host().textContent).toContain('Eva');
  });

  it('la entregada no se puede avanzar', () => {
    const served = rows().find((row) => row.textContent?.includes('Carla'));
    const button = served?.querySelector('button');
    expect(button?.disabled).toBe(true);
  });

  it('quitar una comanda la saca de la pantalla', () => {
    const before = rows().length;

    buttonWith('Quitar')?.click();
    fixture.detectChanges();

    expect(rows().length).toBe(before - 1);
  });

  it('agregar una comanda la muestra', () => {
    buttonWith('Agregar comanda')?.click();
    fixture.detectChanges();

    expect(rows().length).toBe(6);
    expect(host().textContent).toContain('Cliente 6');
  });

  it('reiniciar vuelve al estado inicial', () => {
    buttonWith('Quitar')?.click();
    fixture.detectChanges();
    expect(rows().length).toBe(4);

    buttonWith('Reiniciar')?.click();
    fixture.detectChanges();
    expect(rows().length).toBe(5);
  });

  it('el total de una línea es cantidad por precio', () => {
    // Ana: 2 × 42 = 84
    expect(host().textContent).toContain('84');
  });
});
