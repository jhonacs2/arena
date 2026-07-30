import {
  HttpErrorResponse,
  HttpResponse,
  type HttpInterceptorFn,
  type HttpRequest,
} from '@angular/common/http';
import { inject } from '@angular/core';
import { Observable, of, throwError } from 'rxjs';
import { delay } from 'rxjs/operators';

import { environment } from '../../../../environments/environment';
import { MockApiError, MockWorld } from './mock-world';

/** Suficiente para que se vean los estados de carga; poco para que moleste. */
const LATENCY_MS = 180;

const bearer = (request: HttpRequest<unknown>): string | null =>
  request.headers.get('Authorization')?.replace(/^Bearer\s+/i, '') ?? null;

function body<T>(request: HttpRequest<T>): Record<string, unknown> {
  const raw = request.body;
  return typeof raw === 'object' && raw !== null ? (raw as Record<string, unknown>) : {};
}

const text = (value: unknown): string => (typeof value === 'string' ? value : '');
const int = (value: unknown): number => (typeof value === 'number' ? Math.trunc(value) : NaN);

/**
 * Enruta `/api/**` al backend de mentira.
 *
 * Va **último** en la cadena a propósito: así ve la request con el header
 * `Authorization` ya puesto por `authInterceptor` y puede validar el rol como lo
 * haría el servidor.
 */
export const mockBackendInterceptor: HttpInterceptorFn = (request, next) => {
  const base = environment.apiBaseUrl;
  if (!request.url.startsWith(base)) return next(request);

  const world = inject(MockWorld);
  const route = request.url.slice(base.length).split('?')[0] ?? '';
  const token = bearer(request);

  try {
    const payload = resolve(world, request, route, token);
    if (payload === undefined) return next(request);
    return of(new HttpResponse({ status: 200, body: payload })).pipe(delay(LATENCY_MS));
  } catch (cause) {
    if (cause instanceof MockApiError) {
      const response = new HttpErrorResponse({
        status: cause.status,
        url: request.url,
        error: { error: { code: cause.code, message: cause.message } },
      });
      return throwError(() => response).pipe(delay(LATENCY_MS)) as Observable<never>;
    }
    throw cause;
  }
};

/** `undefined` significa «esta ruta no la maneja el mock». */
function resolve(
  world: MockWorld,
  request: HttpRequest<unknown>,
  route: string,
  token: string | null,
): unknown {
  const data = body(request);
  const post = request.method === 'POST';
  const get = request.method === 'GET';

  if (post && route === '/auth/check-code') return world.checkCode(text(data['code']));
  if (post && route === '/auth/redeem') {
    return world.redeem({
      code: text(data['code']),
      firstName: text(data['firstName']),
      lastName: text(data['lastName']),
      username: text(data['username']),
      password: text(data['password']),
    });
  }
  if (post && route === '/auth/login') {
    return world.login({ username: text(data['username']), password: text(data['password']) });
  }
  if (post && route === '/auth/refresh') return world.refresh();
  if (post && route === '/auth/logout') {
    world.logout(token);
    return { ok: true };
  }

  if (get && route === '/me') return world.me(token);
  if (get && route === '/me/transactions') return world.transactions(token);
  if (get && route === '/races') return world.listRaces(token);

  const raceMatch = /^\/races\/([^/]+)(\/join|\/bet)?$/.exec(route);
  if (raceMatch !== null) {
    const raceId = raceMatch[1] ?? '';
    const action = raceMatch[2];
    if (get && action === undefined) return world.raceDetail(token, raceId);
    if (post && action === '/join') return world.join(token, raceId);
    if (post && action === '/bet') {
      return world.placeBet(token, raceId, {
        horseId: text(data['horseId']),
        amount: int(data['amount']),
      });
    }
  }

  if (post && route === '/admin/codes') {
    return world.createCodes(token, {
      count: int(data['count']),
      coinsGranted: int(data['coinsGranted']),
      note: text(data['note']),
    });
  }
  if (get && route === '/admin/codes') return world.listCodes(token);
  if (get && route === '/admin/scores') return world.scores(token);

  const giftMatch = /^\/admin\/users\/([^/]+)\/gift$/.exec(route);
  if (post && giftMatch !== null) {
    return world.gift(token, giftMatch[1] ?? '', {
      coins: int(data['coins']),
      note: text(data['note']),
    });
  }

  if (post && route === '/admin/races') {
    const horses = Array.isArray(data['horses']) ? (data['horses'] as unknown[]) : [];
    return world.createRace(token, {
      name: text(data['name']),
      scheduledAt: text(data['scheduledAt']),
      horses: horses.map((raw, index) => {
        const horse = typeof raw === 'object' && raw !== null ? (raw as Record<string, unknown>) : {};
        return {
          number: int(horse['number']) || index + 1,
          name: text(horse['name']),
          nominalOdds: int(horse['nominalOdds']),
        };
      }),
    });
  }

  const adminRaceMatch = /^\/admin\/races\/([^/]+)\/(open|start|cancel)$/.exec(route);
  if (post && adminRaceMatch !== null) {
    const raceId = adminRaceMatch[1] ?? '';
    switch (adminRaceMatch[2]) {
      case 'open':
        return world.openRace(token, raceId);
      case 'start':
        return world.startRace(token, raceId);
      case 'cancel':
        return world.cancelRace(token, raceId, text(data['reason']) || 'Cancelada por el instructor');
    }
  }

  return undefined;
}
