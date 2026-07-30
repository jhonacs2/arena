import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CALL_COUNT, resetCallCount } from './call-count.pipe';
import { MoneyPipe } from './money.pipe';
import { S04Component } from './s04.component';

/**
 * Los tests de S4 son la referencia de qué tiene que lograr la Misión 1.
 *
 * El pipe se prueba **solo**, sin componente: es una función con nombre, y esa
 * es justamente la ventaja.
 */
describe('S4 · money pipe', () => {
  const pipe = new MoneyPipe();

  it('separa los miles y no muestra decimales', () => {
    expect(pipe.transform(4200)).toBe('$ 4.200');
    expect(pipe.transform(1234567)).toBe('$ 1.234.567');
  });

  it('redondea en lugar de arrastrar decimales', () => {
    expect(pipe.transform(4200.6)).toBe('$ 4.201');
  });

  it('acepta otro símbolo como parámetro', () => {
    expect(pipe.transform(4200, 'USD')).toBe('USD 4.200');
  });

  it('el cero es un importe válido', () => {
    expect(pipe.transform(0)).toBe('$ 0');
  });
});

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

  it('muestra los precios formateados por el pipe', () => {
    expect(host().textContent).toContain('$ 4.200');
    expect(host().textContent).toContain('USD 4.200');
  });

  it('el origen sale en mayúsculas sin que el componente lo toque', () => {
    expect(host().textContent).toContain('ETIOPÍA');
  });

  it('la directiva resalta solo el café del día', () => {
    const highlighted = cards().filter((card) => card.classList.contains('is-highlighted'));

    expect(highlighted.length).toBe(1);
    expect(highlighted[0]?.textContent).toContain('Yirgacheffe');
    expect(highlighted[0]?.getAttribute('data-highlight-label')).toBe('Café del día');
  });

  it('la directiva no deja el rótulo en los que no están resaltados', () => {
    const plain = cards().filter((card) => !card.classList.contains('is-highlighted'));

    expect(plain.length).toBe(3);
    expect(plain[0]?.getAttribute('data-highlight-label')).toBeNull();
  });

  it('la directiva estructural dibuja una marca por punto de puntaje', () => {
    const first = cards()[0];
    expect(first?.querySelectorAll('.bean').length).toBe(5);

    const second = cards()[1];
    expect(second?.querySelectorAll('.bean').length).toBe(4);
  });

  it('un puntaje de cero no dibuja ninguna marca', () => {
    // El Cerrado tiene 3; se comprueba el borde con el que tiene menos.
    const ratings = cards().map((card) => card.querySelectorAll('.bean').length);
    expect(Math.min(...ratings)).toBeGreaterThan(0);
    expect(ratings).toEqual([5, 4, 3, 4]);
  });

  it('el pipe impuro corre más veces que el puro', () => {
    const before = { pure: CALL_COUNT.pure, impure: CALL_COUNT.impure };

    // Con OnPush hay que provocar la revisión desde adentro: `detectChanges()`
    // a secas no alcanza, porque Angular no revisa un componente que nadie
    // marcó. Ese matiz es la mitad del ejercicio de predicción.
    const tick = Array.from(host().querySelectorAll('button')).find((button) =>
      button.textContent?.includes('Provocar'),
    );

    tick?.click();
    fixture.detectChanges();
    tick?.click();
    fixture.detectChanges();
    tick?.click();
    fixture.detectChanges();

    // El valor de entrada del pipe es siempre la misma cadena, así que el puro
    // no se vuelve a ejecutar ni una vez.
    expect(CALL_COUNT.pure).toBe(before.pure);
    expect(CALL_COUNT.impure).toBeGreaterThan(before.impure + 2);
  });
});
