import { ComponentFixture, TestBed, fakeAsync, flush, tick } from '@angular/core/testing';

import { CatalogService } from './catalog.service';
import { S06Component } from './s06.component';

/**
 * Estos tests pasan desde el minuto cero: verifican lo que el buscador ya hace.
 *
 * Al terminar el ejercicio la búsqueda va a pasar por un flujo con `debounce`,
 * `distinctUntilChanged` y `switchMap`, y estos mismos tests **tienen que
 * seguir pasando**. Lo que va a cambiar es cuántas búsquedas salen, y eso lo
 * comprueba el enunciado, no estos tests.
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

  it('arranca sin resultados y sin buscar nada', fakeAsync(() => {
    tick(2000);
    fixture.detectChanges();

    expect(catalog.requests).toBe(0);
    expect(host().textContent).toContain('Ningún café coincide');
    flush();
  }));

  it('encuentra por origen', fakeAsync(() => {
    type('etiopía');
    tick(2000);
    fixture.detectChanges();

    expect(rows()).toBe(2);
    expect(host().textContent).toContain('Yirgacheffe');
    flush();
  }));

  it('encuentra por nombre', fakeAsync(() => {
    type('huila');
    tick(2000);
    fixture.detectChanges();

    expect(rows()).toBe(1);
    flush();
  }));

  it('una búsqueda sin resultados muestra el vacío', fakeAsync(() => {
    type('zzz');
    tick(2000);
    fixture.detectChanges();

    expect(host().textContent).toContain('Ningún café coincide');
    flush();
  }));

  it('muestra el estado de error', fakeAsync(() => {
    type('error');
    tick(2000);
    fixture.detectChanges();

    expect(host().querySelector('.error')).not.toBeNull();
    flush();
  }));

  it('reiniciar deja el buscador limpio', fakeAsync(() => {
    type('huila');
    tick(2000);
    fixture.detectChanges();
    expect(rows()).toBe(1);

    const reset = Array.from(host().querySelectorAll('button')).find((button) =>
      button.textContent?.includes('Reiniciar'),
    );
    reset?.click();
    fixture.detectChanges();

    expect(rows()).toBe(0);
    expect(catalog.requests).toBe(0);
    flush();
  }));
});
