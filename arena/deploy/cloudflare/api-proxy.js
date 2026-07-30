/**
 * Worker que hace que `/api/**` viva en el mismo dominio que el frontend.
 *
 * El problema que resuelve: un hostname de Cloudflare no puede ser al mismo
 * tiempo un proyecto de Pages y la punta de un túnel — el registro DNS es uno.
 * Así que el túnel publica `api.arena.ejemplo.com` y este Worker, montado en la
 * ruta `arena.ejemplo.com/api/*`, reenvía ahí. Las rutas de Worker tienen
 * prioridad sobre el dominio de Pages, así que el resto del sitio lo sigue
 * sirviendo Pages sin enterarse.
 *
 * Por qué vale la pena el salto extra, en vez de que el frontend llame directo a
 * `api.arena.ejemplo.com`:
 *
 * - **La cookie de refresh.** Es `HttpOnly` y de un solo uso (api.md). Same-origin
 *   funciona con `SameSite=Lax` y nada más. Cruzando dominios hace falta
 *   `SameSite=None; Secure`, que es exactamente la cookie que un navegador con el
 *   bloqueo de terceros activado tira a la basura — y ahí el alumno se
 *   desloguea cada 15 minutos en plena clase.
 * - **Cero CORS.** No hay preflight que depurar el día del despliegue.
 *
 * Despliegue:
 *
 *   cd arena/deploy/cloudflare
 *   npx wrangler deploy
 *
 * Este Worker **no autentica nada** y no tiene que hacerlo: reenvía tal cual, con
 * el `Authorization` y las cookies intactos. Quién puede qué lo decide el backend
 * en cada handler.
 */

export default {
  /**
   * @param {Request} request
   * @param {{ API_ORIGIN: string }} env
   */
  async fetch(request, env) {
    const url = new URL(request.url);

    // Solo `/api/**`. Si la ruta del Worker quedó mal configurada y llega otra
    // cosa, es mejor un 404 explícito que reenviar el sitio entero al backend.
    if (!url.pathname.startsWith('/api/')) {
      return new Response('No encontrado.', { status: 404 });
    }

    const target = new URL(url.pathname + url.search, env.API_ORIGIN);

    // `redirect: 'manual'` para no seguir redirecciones del origen por nuestra
    // cuenta: el navegador tiene que ver la respuesta real.
    const upstream = new Request(target, request);

    // El WebSocket de la sala (`GET /api/ws`) pasa por acá igual que el resto:
    // Workers reenvía el 101 y el par de sockets si no se toca el body.
    return fetch(upstream, { redirect: 'manual' });
  },
};
