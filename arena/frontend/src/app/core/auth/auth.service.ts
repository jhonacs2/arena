import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Router } from '@angular/router';
import { Observable, firstValueFrom, of, shareReplay } from 'rxjs';
import { catchError, finalize, map } from 'rxjs/operators';

import { environment } from '../../../environments/environment';
import { skipAuth } from '../api/http-context';
import type { LoginRequest, RedeemRequest, Session } from '../models';
import { SessionStore } from './session.store';

interface RefreshResponse {
  readonly accessToken: string;
}

/**
 * Todo lo que habla con `/api/auth/*`, más `/api/me`.
 *
 * Las tres llamadas públicas —`checkCode`, `redeem`, `login`— llevan `skipAuth()`:
 * no hay token todavía y un 401 ahí significa «credenciales mal», no «sesión
 * vencida».
 */
@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly session = inject(SessionStore);
  private readonly router = inject(Router);
  private readonly base = environment.apiBaseUrl;

  /**
   * El refresh en vuelo, compartido.
   *
   * El refresh es **de un solo uso**: al usarlo se invalida. Si tres requests se
   * vencen juntas y cada una pide su refresh, dos se comen un token ya quemado y
   * la sesión se cae sola. Con `shareReplay` las tres esperan la misma respuesta.
   */
  private inFlightRefresh: Observable<string | null> | null = null;

  /**
   * Canjea el código y crea la cuenta.
   *
   * La comprobación previa del código **no** está acá: la hace un `httpResource`
   * en la pantalla de registro, que necesita el cuerpo de la respuesta —cuántas
   * monedas otorga— y no solo un sí o un no.
   */
  async redeem(body: RedeemRequest): Promise<Session> {
    const session = await firstValueFrom(
      this.http.post<Session>(`${this.base}/auth/redeem`, body, {
        context: skipAuth(),
        withCredentials: true,
      }),
    );
    this.session.start(session);
    return session;
  }

  async login(body: LoginRequest): Promise<Session> {
    const session = await firstValueFrom(
      this.http.post<Session>(`${this.base}/auth/login`, body, {
        context: skipAuth(),
        withCredentials: true,
      }),
    );
    this.session.start(session);
    return session;
  }

  /** Devuelve el token nuevo, o `null` si el refresh ya no vale. */
  refresh(): Observable<string | null> {
    this.inFlightRefresh ??= this.http
      .post<RefreshResponse>(
        `${this.base}/auth/refresh`,
        {},
        { context: skipAuth(), withCredentials: true },
      )
      .pipe(
        map((response) => {
          this.session.renew(response.accessToken);
          return response.accessToken;
        }),
        catchError(() => {
          // El refresh ya no vale: la sesión terminó de verdad. Se limpia y se
          // manda a iniciar sesión. Dejar a alguien en una pantalla con un cartel
          // de «tu sesión venció» y sin salida es peor que sacarlo: el botón que
          // necesita está en otra ruta.
          this.session.clear();
          void this.router.navigate(['/ingresar']);
          return of(null);
        }),
        finalize(() => {
          this.inFlightRefresh = null;
        }),
        shareReplay({ bufferSize: 1, refCount: false }),
      );

    return this.inFlightRefresh;
  }

  async logout(): Promise<void> {
    try {
      await firstValueFrom(
        this.http.post(
          `${this.base}/auth/logout`,
          {},
          { context: skipAuth(), withCredentials: true },
        ),
      );
    } catch {
      // Si el logout del servidor falla, la sesión local se cierra igual: dejar
      // al alumno adentro porque el servidor no contestó sería peor.
    }
    this.session.clear();
  }
}
