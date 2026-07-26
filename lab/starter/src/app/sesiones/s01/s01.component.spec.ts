import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';

import { S01Component } from './s01.component';

/**
 * Estos tests prueban la LÓGICA de la clase, que ya viene hecha. Pasan desde
 * el minuto cero.
 *
 * Los bindings del template —que son tu trabajo— no se testean acá: los
 * verificás vos con la lista de «Listo cuando» de `mision-1.md`, mirando la
 * pantalla. Aprender a comprobar en el navegador es parte del ejercicio.
 */
describe('S01Component', () => {
  let fixture: ComponentFixture<S01Component>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [S01Component] }).compileComponents();
    fixture = TestBed.createComponent(S01Component);
    fixture.detectChanges();
  });

  const instancia = () =>
    fixture.componentInstance as unknown as {
      cliente: string;
      cantidad: number;
      pedidos: readonly string[];
      readonly total: number;
      readonly puedeAgregar: boolean;
      agregar(): void;
      alternarDisponibilidad(): void;
      limpiar(): void;
    };

  it('se crea sin errores', () => {
    expect(fixture.componentInstance).toBeTruthy();
    expect(fixture.debugElement.query(By.css('.producto'))).not.toBeNull();
  });

  it('el total es el precio por la cantidad', () => {
    const componente = instancia();
    componente.cantidad = 3;
    expect(componente.total).toBe(126);
  });

  it('no deja agregar sin nombre de cliente', () => {
    const componente = instancia();
    expect(componente.puedeAgregar).toBe(false);

    componente.cliente = 'Ana';
    expect(componente.puedeAgregar).toBe(true);
  });

  it('agregar suma una línea y limpia el formulario', () => {
    const componente = instancia();
    componente.cliente = 'Bruno';
    componente.cantidad = 2;
    componente.agregar();

    expect(componente.pedidos.length).toBe(1);
    expect(componente.pedidos[0]).toContain('Bruno');
    expect(componente.cliente).toBe('');
    expect(componente.cantidad).toBe(1);
  });

  it('no muta el array de pedidos', () => {
    // La regla de inmutabilidad del curso. Si acá hubiera un push, con OnPush
    // la vista dejaría de actualizarse — y este test lo dice antes.
    const componente = instancia();
    const original = componente.pedidos;

    componente.cliente = 'Caro';
    componente.agregar();

    expect(componente.pedidos).not.toBe(original);
    expect(original.length).toBe(0);
  });

  it('alternar disponibilidad bloquea el pedido', () => {
    const componente = instancia();
    componente.cliente = 'Diego';
    expect(componente.puedeAgregar).toBe(true);

    componente.alternarDisponibilidad();
    expect(componente.puedeAgregar).toBe(false);
  });
});
