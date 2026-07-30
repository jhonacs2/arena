import { ChangeDetectionStrategy, Component } from '@angular/core';

/**
 * La pantalla de inicio del lab.
 *
 * Junto con S0 es lo único que viene hecho. Todo lo demás lo vas a construir
 * vos, una sesión por vez, y lo vas a ver aparecer en la barra de la izquierda.
 */
@Component({
  selector: 'app-home',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <header class="header">
      <p class="eyebrow">Módulo Angular · Talento DH</p>
      <h1>El lab</h1>
      <p class="intro">
        Acá vas a practicar el concepto de cada clase, aislado y en chiquito, antes de aplicarlo al
        proyecto del hipódromo.
      </p>
    </header>

    <section class="notice">
      <h2>Empezá por S0</h2>
      <p>
        En la barra de la izquierda hay una sola sesión, <strong>S0 · TypeScript</strong>. Esa
        pantalla viene hecha: lo que se construye ahí son los tipos, en
        <code>sessions/s00/menu.ts</code>.
      </p>
      <p>
        De S1 en adelante la barra la vas a hacer crecer vos: las rutas las declara alguien, y ese
        alguien vas a ser vos.
      </p>
      <p class="small">
        Buscá <code>TODO(S0)</code> hoy y <code>TODO(S1)</code> la clase que viene, en
        <code>src/app/</code>.
      </p>
    </section>
  `,
  styles: `
    .header {
      max-inline-size: 55ch;
      margin-block-end: var(--space-8);
    }

    .eyebrow {
      font-family: var(--font-mono);
      font-size: var(--text-2xs);
      letter-spacing: 0.14em;
      text-transform: uppercase;
      color: var(--text-muted);
    }

    .intro {
      margin-block-start: var(--space-3);
      color: var(--text-muted);
    }

    .notice {
      max-inline-size: 60ch;
      padding: var(--space-6);
      border: var(--border-width) solid var(--border);
      box-shadow: var(--shadow-hard);
      background: var(--surface-raised);
    }

    .notice h2 {
      margin-block-end: var(--space-3);
    }

    .notice p + p {
      margin-block-start: var(--space-3);
    }

    .small {
      font-size: var(--text-sm);
      color: var(--text-muted);
    }

    code {
      font-family: var(--font-mono);
      font-size: 0.9em;
      background: var(--surface-sunken);
      border: 2px solid var(--border);
      padding: 0 0.25em;
    }
  `,
})
export class HomeComponent {}
