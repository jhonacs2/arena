import { Pipe, type PipeTransform } from '@angular/core';

const pad = (value: number): string => String(value).padStart(2, '0');

/**
 * Una fecha ISO como la lee alguien en un aula: `29/07 22:30`.
 *
 * Sin año a propósito: todo lo que se ve en Arena pasó o va a pasar dentro del
 * mismo módulo de once sesiones, y el año solo ocupa ancho en una tabla.
 *
 * Formateado a mano y no con `Intl`. Es una decisión, no pereza: `Intl` con
 * `es-AR` devuelve «29/7 06:30 p. m.», que ocupa más, no alinea en una columna de
 * cifras tabulares y depende del ICU de la máquina. Acá siempre son cinco
 * caracteres de fecha y cinco de hora, en 24 horas, en cualquier navegador.
 */
@Pipe({ name: 'when' })
export class WhenPipe implements PipeTransform {
  transform(value: string | null | undefined): string {
    if (!value) return '—';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return '—';
    return (
      `${pad(date.getDate())}/${pad(date.getMonth() + 1)} ` +
      `${pad(date.getHours())}:${pad(date.getMinutes())}`
    );
  }
}
