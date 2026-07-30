import { InjectionToken } from '@angular/core';

/**
 * S5 · A dónde apunta la aplicación.
 *
 * Un servicio se pide por su clase. Una cadena no: no se puede escribir
 * `inject(string)`, porque habría un solo `string` para toda la aplicación y no
 * querría decir nada.
 *
 * Un `InjectionToken` es una llave con nombre y con tipo, hecha a propósito
 * para eso. El texto que recibe es solo para que los errores se entiendan.
 *
 * **Por qué esto no es una constante exportada.** Una constante se importa, y
 * quien la importa queda atado a ella: para probar contra otro servidor habría
 * que tocar cada archivo que la usa. Un token se reemplaza en un solo lugar:
 *
 *   providers: [{ provide: API_URL, useValue: 'http://localhost:8080' }]
 *
 * En S7, cambiar entre el backend Go real y el mock va a ser exactamente esa
 * línea, y ni un componente se entera.
 */
export const API_URL = new InjectionToken<string>('URL base de la API', {
  providedIn: 'root',
  factory: () => '/api',
});
