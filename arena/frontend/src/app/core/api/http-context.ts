import { HttpContext, HttpContextToken } from '@angular/common/http';

/**
 * Marca las requests que el interceptor de autenticación no tiene que tocar:
 * login, canje de código y el propio refresh.
 *
 * Sin esto, un 401 del login dispararía un refresh, que fallaría, que dispararía
 * un logout — y el alumno que escribió mal la contraseña vería «tu sesión venció»
 * en vez de «contraseña incorrecta».
 */
export const SKIP_AUTH = new HttpContextToken<boolean>(() => false);

export const skipAuth = (): HttpContext => new HttpContext().set(SKIP_AUTH, true);
