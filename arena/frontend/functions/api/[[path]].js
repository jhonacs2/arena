/**
 * Pages Function que hace que `/api/**` viva en el mismo dominio que el frontend.
 *
 * El problema que resuelve: un hostname de Cloudflare no puede ser al mismo
 * tiempo un proyecto de Pages y la punta de un túnel — el registro DNS es uno
 * solo. Así que el backend se publica aparte, en `api.arena.ejemplo.com`, y todo
 * lo que el navegador pide a `/api/**` pasa por acá y se reenvía allá.
 *
 * **El nombre del archivo es la configuración.** `functions/api/[[path]].js`
 * significa «todo lo que empiece con /api/», con `[[path]]` capturando el resto
 * incluidas las barras. No hay ruta que registrar en ningún panel: Pages arma el
 * ruteo desde la carpeta al desplegar. Por eso esto reemplaza al Worker de
 * `arena/deploy/cloudflare/`: hace lo mismo, pero viaja con el frontend y se
 * despliega con él, en vez de ser un `wrangler deploy` aparte que hay que
 * acordarse de correr.
 *
 * Por qué el salto extra, en vez de que el frontend llame directo a
 * `api.arena.ejemplo.com`:
 *
 * - **La cookie de refresh.** Es `HttpOnly` y de un solo uso (api.md). Same-origin
 *   funciona con `SameSite=Lax` y nada más. Cruzando dominios hace falta
 *   `SameSite=None; Secure`, que es exactamente la cookie que un navegador con el
 *   bloqueo de terceros activado tira a la basura — y ahí el alumno se desloguea
 *   cada 15 minutos en plena clase.
 * - **Cero CORS.** No hay preflight que depurar el día del despliegue.
 *
 * Configuración, una sola variable: en el panel de Pages → *Settings* →
 * *Environment variables*, `API_ORIGIN = https://api.arena.ejemplo.com`. Va en
 * los dos entornos, producción y preview, o los previews quedan sin backend.
 *
 * Esto **no autentica nada** y no tiene que hacerlo: reenvía tal cual, con el
 * `Authorization` y las cookies intactos. Quién puede qué lo decide el backend en
 * cada handler.
 */

/**
 * @param {{ request: Request, env: { API_ORIGIN?: string } }} context
 * @returns {Promise<Response>}
 */
export async function onRequest(context) {
  const { request, env } = context;

  // Sin la variable no hay nada que hacer, y el error tiene que decir eso. Un
  // reenvío a `undefined` da un mensaje de red que no señala a ningún lado, y es
  // el primer síntoma que vas a ver si te olvidaste de configurarla en preview.
  if (!env.API_ORIGIN) {
    return Response.json(
      {
        error: {
          code: 'API_ORIGIN_MISSING',
          message: 'Falta configurar API_ORIGIN en las variables de entorno de Pages.',
          details: {},
        },
      },
      { status: 503 },
    );
  }

  const url = new URL(request.url);
  const target = new URL(url.pathname + url.search, env.API_ORIGIN);

  // El `Request` original entero: método, cabeceras, cuerpo y la cookie de
  // refresh. Reconstruirlo a mano sería la forma de perder alguna sin notarlo.
  const upstream = new Request(target, request);

  // El WebSocket de la sala (`GET /api/ws`) pasa por acá igual que el resto: al
  // reenviar el pedido sin tocarle el cuerpo, el runtime devuelve el 101 y
  // empalma los dos sockets.
  //
  // `redirect: 'manual'` para no seguir redirecciones del origen por nuestra
  // cuenta: el navegador tiene que ver la respuesta real.
  return fetch(upstream, { redirect: 'manual' });
}
