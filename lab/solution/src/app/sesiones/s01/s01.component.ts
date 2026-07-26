import { ChangeDetectionStrategy, Component } from '@angular/core';
import { FormsModule } from '@angular/forms';

/**
 * S1 · Primer componente standalone y los cuatro tipos de binding.
 *
 * Un mostrador de cafetería. Todo pasa dentro de ESTE componente: todavía no
 * hay padres ni hijos —eso es S2—, así que acá se ve un componente solo,
 * hablando con su propio template.
 *
 * Los cuatro caminos entre la clase y el template:
 *
 *   {{ nombre }}          clase → template   interpolación
 *   [disabled]="…"        clase → template   property binding
 *   (click)="…"           template → clase   event binding
 *   [(ngModel)]="…"       en los dos sentidos
 *
 * ⚠️ `[(ngModel)]` no funciona sin importar `FormsModule`. Es un standalone:
 * los imports son de ESTE componente, no de un módulo lejano. Olvidarlo es el
 * error número uno de la sesión, y por eso está en «predice y ejecuta».
 */
@Component({
  selector: 'app-s01',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [FormsModule],
  templateUrl: './s01.component.html',
  styleUrl: './s01.component.css',
})
export class S01Component {
  /** El producto del mostrador. */
  protected cafe = {
    nombre: 'Yirgacheffe',
    origen: 'Etiopía',
    precio: 42,
    disponible: true,
  };

  /** Lo que el cliente escribe. Va y viene con `[(ngModel)]`. */
  protected cliente = '';
  protected cantidad = 1;

  /** Los pedidos ya tomados. */
  protected pedidos: readonly string[] = [];

  /** Se recalcula sola cada vez que Angular repinta. */
  protected get total(): number {
    return this.cafe.precio * this.cantidad;
  }

  protected get puedeAgregar(): boolean {
    return this.cafe.disponible && this.cantidad > 0 && this.cliente.trim().length > 0;
  }

  protected agregar(): void {
    if (!this.puedeAgregar) return;

    // Nunca `push` sobre el estado: se crea un array nuevo. Es la regla de
    // inmutabilidad de project/frontend/CLAUDE.md, y en S3 va a ser la diferencia entre que
    // la vista se actualice y que no.
    this.pedidos = [...this.pedidos, `${this.cantidad} × ${this.cafe.nombre} para ${this.cliente.trim()}`];

    this.cliente = '';
    this.cantidad = 1;
  }

  protected alternarDisponibilidad(): void {
    this.cafe = { ...this.cafe, disponible: !this.cafe.disponible };
  }

  protected limpiar(): void {
    this.pedidos = [];
  }
}
