import { Directive, input } from '@angular/core';

/**
 * S4 · Marca la fila del caballo favorito.
 *
 *   <li [appFavourite]="horse.id === view.favourite?.id">…</li>
 *
 * Una directiva de atributo: no dibuja nada, le agrega comportamiento a un
 * elemento que ya existe.
 *
 * Lo que aporta además del color es lo que **no se ve**: el favorito de una
 * carrera es información, no decoración, y quien navega con lector de pantalla
 * también tiene que enterarse. Poner esa regla en un solo lugar es la razón por
 * la que esto es una directiva y no dos atributos repetidos en el template.
 */
@Directive({
  selector: '[appFavourite]',
  standalone: true,
  host: {
    '[class.is-favourite]': 'appFavourite()',
    '[attr.data-favourite-label]': 'appFavourite() ? label() : null',
  },
})
export class FavouriteDirective {
  readonly appFavourite = input(false);

  /** Lo que se anuncia. Se puede cambiar sin tocar la directiva. */
  readonly label = input('Favorito');
}
