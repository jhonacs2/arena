import { ComponentFixture, TestBed, fakeAsync, flush, tick } from '@angular/core/testing';

import { CatalogService } from './catalog.service';
import { S06Component } from './s06.component';

/**
 * Los tests de S6 son la referencia de qué tiene que lograr la Misión 1.
 *
 * Todos usan `fakeAsync` y `tick`: el tiempo es el tema de la sesión, así que
 * hay que poder adelantarlo a voluntad en vez de esperarlo de verdad.
 */
describe('S6 · el buscador', () => {
  let fixture: ComponentFixture<S06Component>;
  let catalog: CatalogService;

  const host = (): HTMLElement => fixture.nativeElement as HTMLElement;
  const rows = (): number => host().querySelectorAll('.row').length;

  const type = (text: string): void => {
    const input = host().querySelector<HTMLInputElement>('input');
    if (!input) return;
    input.value = text;
    input.dispatchEvent(new Event('input'));
    fixture.detectChanges();
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [S06Component] }).compileComponents();
    catalog = TestBed.inject(CatalogService);
    catalog.resetRequests();
    fixture = TestBed.createComponent(S06Component);
    fixture.detectChanges();
  });

  it('no busca nada hasta que alguien escribe', fakeAsync(() => {
    tick(2000);
    fixture.detectChanges();

    expect(catalog.requests).toBe(0);
    flush();
  }));

  it('encuentra por origen', fakeAsync(() => {
    type('etiopía');
    tick(300); // debounce
    tick(300); // latencia
    fixture.detectChanges();

    expect(rows()).toBe(2);
    expect(host().textContent).toContain('Yirgacheffe');
    expect(host().textContent).toContain('Sidamo');
    flush();
  }));

  it('siete teclas seguidas producen UNA sola búsqueda', fakeAsync(() => {
    for (const text of ['e', 'et', 'eti', 'etio', 'etiop', 'etiopi', 'etiopía']) {
      type(text);
      tick(50); // más rápido que el debounce
    }

    tick(300);
    tick(1500);
    fixture.detectChanges();

    expect(catalog.requests).toBe(1);
    expect(rows()).toBe(2);
    flush();
  }));

  it('escribir lo mismo dos veces no vuelve a buscar', fakeAsync(() => {
    type('huila');
    tick(300);
    tick(300);
    fixture.detectChanges();
    expect(catalog.requests).toBe(1);

    type('huila');
    tick(300);
    tick(300);
    fixture.detectChanges();

    expect(catalog.requests).toBe(1);
    flush();
  }));

  it('gana la última búsqueda, aunque la anterior sea más lenta', fakeAsync(() => {
    // 'e' tarda 1200 ms; 'huila' tarda 300 ms. Sin switchMap, la respuesta de
    // 'e' llegaría después y pisaría la de 'huila'.
    type('e');
    tick(300); // pasa el debounce y sale la búsqueda lenta
    tick(100);

    type('huila');
    tick(300); // debounce
    tick(300); // la rápida responde
    fixture.detectChanges();

    expect(rows()).toBe(1);
    expect(host().textContent).toContain('Huila');

    // Y aunque pase el tiempo que le faltaba a la lenta, no vuelve.
    tick(2000);
    fixture.detectChanges();

    expect(rows()).toBe(1);
    expect(host().textContent).toContain('Huila');
    flush();
  }));

  it('muestra el estado de carga mientras espera', fakeAsync(() => {
    type('kenia');
    tick(300);
    fixture.detectChanges();

    expect(host().querySelectorAll('.skeleton').length).toBeGreaterThan(0);

    tick(300);
    fixture.detectChanges();

    expect(host().querySelectorAll('.skeleton').length).toBe(0);
    flush();
  }));

  it('un error no rompe el buscador para siempre', fakeAsync(() => {
    type('error');
    tick(300);
    tick(300);
    fixture.detectChanges();

    expect(host().querySelector('.error')).not.toBeNull();

    // Y después de un error, la siguiente búsqueda funciona.
    type('huila');
    tick(300);
    tick(300);
    fixture.detectChanges();

    expect(host().querySelector('.error')).toBeNull();
    expect(rows()).toBe(1);
    flush();
  }));

  it('una búsqueda sin resultados muestra el vacío', fakeAsync(() => {
    type('zzz');
    tick(300);
    tick(300);
    fixture.detectChanges();

    expect(host().textContent).toContain('Ningún café coincide');
    flush();
  }));
});
