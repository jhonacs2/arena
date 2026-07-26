import { ChangeDetectionStrategy, Component, input } from '@angular/core';

/**
 * La marca.
 *
 * Es una casaca de jockey reducida a su mínima expresión: la silueta con
 * mangas y un chevron. Sale del mismo mundo que las sedas, así que el logo y
 * los 54 caballos hablan el mismo idioma sin que haya que explicarlo.
 *
 * Monocromo y de una sola forma: a 16 px, en el favicon, sobrevive la silueta.
 * Especificación en `docs/design/IMAGES.md` §1 — cuando exista el archivo
 * definitivo, se reemplaza este componente.
 */
@Component({
  selector: 'app-logo',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <svg viewBox="0 0 40 38" [attr.aria-label]="etiqueta()" role="img">
      <g fill="none" stroke="currentColor" stroke-width="3" stroke-linejoin="round">
        <rect x="3.5" y="6.5" width="8" height="13" />
        <rect x="28.5" y="6.5" width="8" height="13" />
        <rect x="12.5" y="4.5" width="15" height="29" />
      </g>
      <path
        class="chevron"
        d="M12.5,12 L20,17 L27.5,12 L27.5,17 L20,22 L12.5,17 Z"
        fill="currentColor"
      />
      <path d="M17,4.5 L23,4.5 L21.5,8 L18.5,8 Z" fill="currentColor" />
    </svg>
  `,
  styles: `
    :host {
      display: inline-block;
      inline-size: 2rem;
      line-height: 0;
    }

    svg {
      inline-size: 100%;
      block-size: auto;
    }

    /* El único elemento con color de la marca. El resto es la tinta del tema. */
    .chevron {
      color: var(--accent);
    }
  `,
})
export class LogoComponent {
  readonly etiqueta = input('Hipódromo');
}
