import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CoffeeCardComponent } from './coffee-card.component';
import { S02Component } from './s02.component';
import type { Coffee } from './menu';

/**
 * Los tests de S2 son la referencia de qué tiene que lograr la Misión 1.
 *
 * El hijo se prueba **solo**, sin su padre: es la prueba de que un componente
 * con `input()` y `output()` no depende de quién lo use.
 */
describe('S2 · coffee-card (el hijo)', () => {
  const coffee: Coffee = { id: 'x', name: 'Prueba', origin: 'Perú', price: 100, available: true };

  let fixture: ComponentFixture<CoffeeCardComponent>;

  const text = (): string => (fixture.nativeElement as HTMLElement).textContent ?? '';
  const buttons = (): HTMLButtonElement[] =>
    Array.from((fixture.nativeElement as HTMLElement).querySelectorAll('button'));

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [CoffeeCardComponent] }).compileComponents();
    fixture = TestBed.createComponent(CoffeeCardComponent);
    fixture.componentRef.setInput('coffee', coffee);
    fixture.detectChanges();
  });

  it('muestra lo que le pasaron por input', () => {
    expect(text()).toContain('Prueba');
    expect(text()).toContain('Perú');
    expect(text()).toContain('100');
  });

  it('la cantidad arranca en 1 y el total la acompaña', () => {
    expect(text()).toContain('Total:');
    expect(text()).toContain('100');

    const plus = buttons()[1];
    plus?.click();
    fixture.detectChanges();

    expect(text()).toContain('200');
  });

  it('emite el pedido con el café y la cantidad', () => {
    let received: { quantity: number; name: string } | undefined;
    fixture.componentInstance.ordered.subscribe((request) => {
      received = { quantity: request.quantity, name: request.coffee.name };
    });

    const order = buttons().find((button) => button.textContent?.includes('Pedir'));
    order?.click();

    expect(received).toEqual({ quantity: 1, name: 'Prueba' });
  });

  it('un café sin stock no se puede pedir', () => {
    fixture.componentRef.setInput('coffee', { ...coffee, available: false });
    fixture.detectChanges();

    const order = buttons().find((button) => button.textContent?.includes('Sin stock'));
    expect(order?.disabled).toBe(true);
  });

  it('cuenta los cambios de input con ngOnChanges', () => {
    fixture.componentRef.setInput('coffee', { ...coffee, price: 200 });
    fixture.detectChanges();

    expect(text()).toContain('ngOnChanges ×2');
  });

  it('avisa al destruirse', () => {
    let farewell = '';
    fixture.componentInstance.destroyed.subscribe((name) => (farewell = name));

    fixture.destroy();

    expect(farewell).toBe('Prueba');
  });
});

describe('S2 · s02 (el padre)', () => {
  let fixture: ComponentFixture<S02Component>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [S02Component] }).compileComponents();
    fixture = TestBed.createComponent(S02Component);
    fixture.detectChanges();
  });

  it('dibuja una tarjeta por café', () => {
    const cards = (fixture.nativeElement as HTMLElement).querySelectorAll('app-coffee-card');
    expect(cards.length).toBe(4);
  });

  it('proyecta el rótulo del café del día en la tarjeta destacada', () => {
    const tag = (fixture.nativeElement as HTMLElement).querySelector('.tag');
    expect(tag?.textContent).toContain('Café del día');
  });

  it('anota en la comanda lo que emite una tarjeta', () => {
    const host = fixture.nativeElement as HTMLElement;
    expect(host.textContent).toContain('Todavía no pidió nadie');

    const order = Array.from(host.querySelectorAll('button')).find((button) =>
      button.textContent?.includes('Pedir'),
    );
    order?.click();
    fixture.detectChanges();

    expect(host.textContent).toContain('1 × Yirgacheffe');
  });
});
