import { ChangeDetectionStrategy, Component } from '@angular/core';

/**
 * La pantalla de inicio del lab.
 *
 * Es lo único que viene hecho. Todo lo demás lo vas a construir vos, una
 * sesión por vez, y lo vas a ver aparecer en la barra de la izquierda.
 */
@Component({
  selector: 'app-inicio',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <header class="cabecera">
      <p class="eyebrow">Módulo Angular · Talento DH</p>
      <h1>El lab</h1>
      <p class="entrada">
        Acá vas a practicar el concepto de cada clase, aislado y en chiquito, antes de aplicarlo al
        proyecto del hipódromo.
      </p>
    </header>

    <section class="aviso">
      <h2>Todavía no hay ninguna sesión</h2>
      <p>
        La barra de la izquierda está vacía porque las rutas las declara alguien, y ese alguien vas
        a ser vos. Al terminar la primera clase vas a ver aparecer <strong>S1</strong> ahí.
      </p>
      <p class="chico">
        Buscá <code>TODO(S1)</code> en <code>src/app/</code> para saber dónde empezar.
      </p>
    </section>
  `,
  styles: `
    .cabecera {
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

    .entrada {
      margin-block-start: var(--space-3);
      color: var(--text-muted);
    }

    .aviso {
      max-inline-size: 60ch;
      padding: var(--space-6);
      border: var(--border-width) solid var(--border);
      box-shadow: var(--shadow-hard);
      background: var(--surface-raised);
    }

    .aviso h2 {
      margin-block-end: var(--space-3);
    }

    .aviso p + p {
      margin-block-start: var(--space-3);
    }

    .chico {
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
export class InicioComponent {}
