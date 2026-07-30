import { ComponentFixture, TestBed } from '@angular/core/testing';

import { S03Component } from './s03.component';

/**
 * Los tests de S3 son la referencia de qué tiene que lograr la Misión 1.
 *
 * Todos miran **la pantalla**, no los signals: lo que hay que lograr es que la
 * vista siga al estado, y eso solo se comprueba desde afuera.
 */
describe('S3 · el tablero de la comanda', () => {
  let fixture: ComponentFixture<S03Component>;

  const host = (): HTMLElement => fixture.nativeElement as HTMLElement;
  const rows = (): HTMLElement[] => Array.from(host().querySelectorAll('.order'));
  const buttonWith = (label: string): HTMLButtonElement | undefined =>
    Array.from(host().querySelectorAll('button')).find((button) =>
      button.textContent?.trim().startsWith(label),
    );
  const tab = (label: string): HTMLButtonElement | undefined =>
    Array.from(host().querySelectorAll<HTMLButtonElement>('.tab')).find((button) =>
      button.textContent?.includes(label),
    );

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [S03Component] }).compileComponents();
    fixture = TestBed.createComponent(S03Component);
    fixture.detectChanges();
  });

  it('arranca mostrando las cinco comandas', () => {
    expect(rows().length).toBe(5);
  });

  it('el filtro deja solo las de ese estado', () => {
    tab('Pendientes')?.click();
    fixture.detectChanges();

    expect(rows().length).toBe(2);
    expect(host().textContent).toContain('Ana');
    expect(host().textContent).not.toContain('Carla');
  });

  it('el contador de cada filtro sale de la comanda, no de lo que se ve', () => {
    tab('Pendientes')?.click();
    fixture.detectChanges();

    expect(tab('Todas')?.textContent).toContain('5');
    expect(tab('Entregadas')?.textContent).toContain('1');
  });

  it('avanzar una comanda cambia su estado y repinta la pantalla', () => {
    expect(host().textContent).toContain('Pendiente');

    buttonWith('Marcar lista')?.click();
    fixture.detectChanges();

    const ready = host().textContent?.match(/Lista/g) ?? [];
    expect(ready.length).toBeGreaterThan(2);
  });

  it('quitar una comanda la saca de la pantalla', () => {
    const before = rows().length;

    buttonWith('Quitar')?.click();
    fixture.detectChanges();

    expect(rows().length).toBe(before - 1);
  });

  it('agregar una comanda la muestra al final', () => {
    buttonWith('Agregar comanda')?.click();
    fixture.detectChanges();

    expect(rows().length).toBe(6);
    expect(host().textContent).toContain('Cliente 6');
  });

  it('el total por cobrar deja afuera lo entregado', () => {
    // pendientes y listas: 2×42 + 1×38 + 1×45 + 1×42 = 209
    expect(host().textContent).toContain('209');
  });

  it('el mensaje de vacío depende del filtro', () => {
    buttonWith('Marcar lista')?.click();
    fixture.detectChanges();
    buttonWith('Marcar lista')?.click();
    fixture.detectChanges();

    tab('Pendientes')?.click();
    fixture.detectChanges();

    expect(host().textContent).toContain('La barra está al día');
  });

  it('ordenar por precio no reordena la comanda', () => {
    const listedFirst = rows()[0]?.textContent ?? '';
    const panel = host().querySelector('.panel__list')?.textContent ?? '';

    // La más cara es Carla (3 × 30 = 90); la comanda sigue empezando por Ana.
    expect(panel.trim().startsWith('Carla')).toBe(true);
    expect(listedFirst).toContain('Ana');
  });

  it('reiniciar vuelve al estado inicial', () => {
    buttonWith('Quitar')?.click();
    fixture.detectChanges();
    expect(rows().length).toBe(4);

    buttonWith('Reiniciar')?.click();
    fixture.detectChanges();
    expect(rows().length).toBe(5);
  });
});
