import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ChangeDetectionStrategy, Component, Injectable, inject } from '@angular/core';

import { SHOP_NAME } from './shop.token';
import { OrderService } from './order.service';
import { S05Component } from './s05.component';

/**
 * Los tests de S5 son la referencia de qué tiene que lograr la Misión 1.
 *
 * Los dos primeros son el corazón de la sesión: comprobar desde la pantalla que
 * un servicio es compartido y el otro no.
 */
describe('S5 · dos mostradores y una comanda', () => {
  let fixture: ComponentFixture<S05Component>;

  const host = (): HTMLElement => fixture.nativeElement as HTMLElement;
  const counters = (): HTMLElement[] => Array.from(host().querySelectorAll('app-counter'));

  /** Escribe en un input y dispara el evento que ngModel escucha. */
  const type = (input: HTMLInputElement | null | undefined, text: string): void => {
    if (!input) return;
    input.value = text;
    input.dispatchEvent(new Event('input'));
    fixture.detectChanges();
  };

  const buttonIn = (root: HTMLElement, label: string): HTMLButtonElement | undefined =>
    Array.from(root.querySelectorAll('button')).find((button) =>
      button.textContent?.trim().startsWith(label),
    );

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [S05Component] }).compileComponents();
    fixture = TestBed.createComponent(S05Component);
    fixture.detectChanges();
  });

  it('dibuja los dos mostradores', () => {
    expect(counters().length).toBe(2);
  });

  it('un pedido tomado en un mostrador aparece en el otro y en la comanda', () => {
    const [first, second] = counters();
    expect(first && second).toBeTruthy();

    type(first?.querySelector<HTMLInputElement>('input'), 'Ana');
    buttonIn(first as HTMLElement, 'Tomar pedido')?.click();
    fixture.detectChanges();

    // El mostrador que no tomó el pedido también ve el contador en 1.
    expect(first?.textContent).toContain('1');
    expect(second?.textContent).toContain('1');
    // Y la comanda de la pantalla lo tiene.
    expect(host().querySelector('.orders')?.textContent).toContain('Ana');
  });

  it('el cuaderno de cada mostrador es distinto', () => {
    const [first, second] = counters();

    const noteInput = first?.querySelectorAll<HTMLInputElement>('input')[1];
    type(noteInput, 'Falta leche');
    buttonIn(first as HTMLElement, 'Anotar')?.click();
    fixture.detectChanges();

    expect(first?.querySelector('.notes')?.textContent).toContain('Falta leche');
    expect(second?.querySelector('.notes')).toBeNull();
    expect(second?.textContent).toContain('Cuaderno vacío');
  });

  it('quitar un pedido lo saca de la comanda', () => {
    const [first] = counters();
    type(first?.querySelector<HTMLInputElement>('input'), 'Beto');
    buttonIn(first as HTMLElement, 'Tomar pedido')?.click();
    fixture.detectChanges();

    buttonIn(host(), 'Quitar')?.click();
    fixture.detectChanges();

    expect(host().textContent).toContain('Todavía no tomó nadie');
  });

  it('el token trae el nombre por defecto', () => {
    expect(host().querySelector('h1')?.textContent).toContain('Café Compilado');
  });
});

describe('S5 · el servicio, sin componente', () => {
  it('es la misma instancia cada vez que se lo pide', () => {
    TestBed.configureTestingModule({});
    const first = TestBed.inject(OrderService);
    const second = TestBed.inject(OrderService);

    expect(first).toBe(second);
  });

  it('no deja escribir el estado desde afuera', () => {
    TestBed.configureTestingModule({});
    const service = TestBed.inject(OrderService);

    // `orders` es de solo lectura: no tiene `set` ni `update`.
    expect('set' in service.orders).toBe(false);
    expect('update' in service.orders).toBe(false);
  });

  it('el token se puede reemplazar sin tocar ningún componente', () => {
    TestBed.configureTestingModule({
      providers: [{ provide: SHOP_NAME, useValue: 'Otro café' }],
    });

    expect(TestBed.inject(SHOP_NAME)).toBe('Otro café');
  });

  it('providers en un componente crea OTRA instancia, sin ningún error', () => {
    // Es el primer «predice y ejecuta»: el servicio dice providedIn: 'root' y
    // además está declarado en el componente. Hay dos, y nada avisa.
    @Component({
      selector: 'app-shadowed',
      standalone: true,
      changeDetection: ChangeDetectionStrategy.OnPush,
      providers: [OrderService],
      template: `<p>{{ orders.count() }}</p>`,
    })
    class ShadowedComponent {
      readonly orders = inject(OrderService);
    }

    TestBed.configureTestingModule({ imports: [ShadowedComponent] });
    const root = TestBed.inject(OrderService);
    const fixture = TestBed.createComponent(ShadowedComponent);

    expect(fixture.componentInstance.orders).not.toBe(root);
  });

  it('un servicio puede inyectar otro servicio', () => {
    @Injectable({ providedIn: 'root' })
    class ReceiptService {
      readonly orders = inject(OrderService);
      readonly summary = (): string => `${this.orders.orders().length} pedidos`;
    }

    TestBed.configureTestingModule({});
    const receipt = TestBed.inject(ReceiptService);

    expect(receipt.orders).toBe(TestBed.inject(OrderService));
    expect(receipt.summary()).toBe('0 pedidos');
  });

  it('inject() fuera de un contexto de inyección falla', () => {
    // Una función suelta, llamada desde ningún lado en particular.
    const outside = (): OrderService => inject(OrderService);

    expect(() => outside()).toThrowError(/NG0203|injection context/i);
  });
});
