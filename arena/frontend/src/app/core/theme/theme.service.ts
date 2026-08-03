import { Injectable, signal } from '@angular/core';

export type Theme = 'light' | 'dark' | 'system';

const STORAGE_KEY = 'arena:tema';

/**
 * Modo claro y oscuro.
 *
 * El oscuro está **diseñado**, no invertido: reasigna tokens semánticos y la
 * sombra dura pasa de tinta a tiza, porque en oscuro la profundidad se dibuja
 * con luz (`docs/design/CLAUDE.md`).
 *
 * `system` es el valor por defecto y la mayoría no toca nunca el interruptor.
 */
@Injectable({ providedIn: 'root' })
export class ThemeService {
  private readonly _theme = signal<Theme>(this.readPreference());
  readonly theme = this._theme.asReadonly();

  constructor() {
    this.apply(this._theme());
  }

  set(theme: Theme): void {
    this._theme.set(theme);
    this.apply(theme);

    try {
      if (theme === 'system') localStorage.removeItem(STORAGE_KEY);
      else localStorage.setItem(STORAGE_KEY, theme);
    } catch {
      // Almacenamiento bloqueado: el tema igual se aplica, solo que no sobrevive
      // a la recarga. No es motivo para romper nada.
    }
  }

  /** Cicla claro → oscuro → sistema. */
  cycle(): void {
    const next: Record<Theme, Theme> = { light: 'dark', dark: 'system', system: 'light' };
    this.set(next[this._theme()]);
  }

  private apply(theme: Theme): void {
    const root = document.documentElement;
    if (theme === 'system') root.removeAttribute('data-theme');
    else root.setAttribute('data-theme', theme);
  }

  private readPreference(): Theme {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored === 'light' || stored === 'dark') return stored;
    } catch {
      // Igual que arriba.
    }
    return 'system';
  }
}
