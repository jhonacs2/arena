import { Injectable, signal } from '@angular/core';

export type Theme = 'claro' | 'oscuro' | 'sistema';

const STORAGE_KEY = 'hipodromo:tema';

/**
 * Modo claro y oscuro.
 *
 * El modo oscuro está **diseñado**, no invertido: reasigna tokens semánticos
 * y el borde neobrutalista pasa de tinta a tiza. Ver `docs/design/tokens.md` §2.
 *
 * `sistema` deja decidir a `prefers-color-scheme` — es el valor por defecto, y
 * la mayoría no toca nunca el interruptor.
 */
@Injectable({ providedIn: 'root' })
export class ThemeService {
  private readonly _theme = signal<Theme>(this.leerPreferencia());
  readonly theme = this._theme.asReadonly();

  constructor() {
    this.aplicar(this._theme());
  }

  set(theme: Theme): void {
    this._theme.set(theme);
    this.aplicar(theme);

    try {
      if (theme === 'sistema') {
        localStorage.removeItem(STORAGE_KEY);
      } else {
        localStorage.setItem(STORAGE_KEY, theme);
      }
    } catch {
      // Modo incógnito o almacenamiento bloqueado: el tema igual se aplica,
      // solo que no sobrevive a la recarga. No es motivo para romper nada.
    }
  }

  /** Cicla claro → oscuro → sistema. */
  ciclar(): void {
    const siguiente: Record<Theme, Theme> = { claro: 'oscuro', oscuro: 'sistema', sistema: 'claro' };
    this.set(siguiente[this._theme()]);
  }

  private aplicar(theme: Theme): void {
    const root = document.documentElement;
    if (theme === 'sistema') {
      root.removeAttribute('data-theme');
    } else {
      root.setAttribute('data-theme', theme === 'claro' ? 'light' : 'dark');
    }
  }

  private leerPreferencia(): Theme {
    try {
      const guardado = localStorage.getItem(STORAGE_KEY);
      if (guardado === 'claro' || guardado === 'oscuro') return guardado;
    } catch {
      // Igual que arriba: sin almacenamiento se cae al valor por defecto.
    }
    return 'sistema';
  }
}
