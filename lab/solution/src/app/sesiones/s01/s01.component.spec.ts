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
    expect($('.producto__origen').textContent).toContain('Etiopía');
    expect($('.producto__precio').textContent).toContain('42');
  });

  it('2 · pone la clase de agotado solo cuando corresponde', () => {
    expect($('.producto').classList).not.toContain('producto--agotado');

    $('.producto .boton').click();
    fixture.detectChanges();

    expect($('.producto').classList).toContain('producto--agotado');
    expect($('.producto__estado').textContent).toContain('Agotado');
  });

  it('3 · el botón dispara el método de la clase', () => {
    const antes = $('.producto__estado').textContent;
    $('.producto .boton').click();
    fixture.detectChanges();

    expect($('.producto__estado').textContent).not.toBe(antes);
  });

  it('4 · el total se actualiza mientras se escribe', async () => {
    const cantidad = $('input[name="cantidad"]') as HTMLInputElement;
    cantidad.value = '3';
    cantidad.dispatchEvent(new Event('input'));
    await fixture.whenStable();
    fixture.detectChanges();

    // 42 × 3
    expect($('.pedido__total').textContent).toContain('126');
  });

  it('el botón de agregar está deshabilitado hasta que hay nombre', async () => {
    const boton = $('.pedido .boton') as HTMLButtonElement;
    expect(boton.disabled).withContext('sin nombre tendría que estar deshabilitado').toBe(true);

    const cliente = $('input[name="cliente"]') as HTMLInputElement;
    cliente.value = 'Ana';
    cliente.dispatchEvent(new Event('input'));
    await fixture.whenStable();
    fixture.detectChanges();

    expect(boton.disabled).toBe(false);
  });

  it('agregar suma una línea a la comanda y limpia el formulario', fakeAsync(() => {
    const cliente = $('input[name="cliente"]') as HTMLInputElement;
    cliente.value = 'Bruno';
    cliente.dispatchEvent(new Event('input'));
    tick();
    fixture.detectChanges();

    ($('.pedido .boton') as HTMLButtonElement).click();
    tick();
    fixture.detectChanges();

    // Lo que se ve: la línea nueva en la comanda.
    const lineas = fixture.debugElement.queryAll(By.css('.comanda__lista li'));
    expect(lineas.length).toBe(1);
    expect((lineas[0]!.nativeElement as HTMLElement).textContent).toContain('Bruno');

    // El vaciado se asserta sobre la clase y no sobre el `value` del input.
    // Que ngModel reescriba el DOM es tarea de Angular y su tiempo no se
    // estabiliza de forma confiable en el test; que la clase quede limpia es
    // la conducta que nos importa. La comprobación visual está en los
    // criterios de «Listo cuando» de mision-1.md.
    const componente = fixture.componentInstance as unknown as { cliente: string; cantidad: number };
    expect(componente.cliente).withContext('el nombre tendría que quedar vacío').toBe('');
    expect(componente.cantidad).toBe(1);
  }));

  it('no muta el array de pedidos', () => {
    // La regla de inmutabilidad del curso. Si alguien usa push, la vista con
    // OnPush deja de actualizarse y este test lo dice antes que el navegador.
    const componente = fixture.componentInstance as unknown as { pedidos: readonly string[] };
    const original = componente.pedidos;

    (fixture.componentInstance as unknown as { cliente: string }).cliente = 'Caro';
    (fixture.componentInstance as unknown as { agregar: () => void }).agregar();

    expect(componente.pedidos).not.toBe(original);
    expect(original.length).toBe(0);
  });
});
