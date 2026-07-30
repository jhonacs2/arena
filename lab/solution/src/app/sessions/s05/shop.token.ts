import { InjectionToken } from '@angular/core';

/**
 * S5 · Un valor de configuración, inyectable.
 *
 * Un servicio se pide por su clase: `inject(OrderService)`. Pero ¿cómo se pide
 * **una cadena**? No se puede escribir `inject(string)`: habría un solo `string`
 * para toda la aplicación y no querría decir nada.
 *
 * Un `InjectionToken` es una llave con nombre y con tipo, creada a propósito
 * para eso. El texto que se le pasa es solo para que los mensajes de error se
 * entiendan.
 *
 * `factory` da el valor por defecto. Con eso, esto funciona sin configurar nada
 * — y quien quiera otro valor lo cambia en un solo lugar, sin tocar ni un
 * componente:
 *
 *   providers: [{ provide: SHOP_NAME, useValue: 'Otro café' }]
 *
 * En el hipódromo, el token que hace falta es `API_URL`: en S7 va a apuntar al
 * servidor de verdad o al mock, y cambiarlo tiene que ser una línea.
 */
export const SHOP_NAME = new InjectionToken<string>('Nombre del café', {
  providedIn: 'root',
  factory: () => 'Café Compilado',
});
