/**
 * Configuración de entorno.
 *
 * `apiBaseUrl` es relativo a propósito: en producción el frontend vive en
 * Cloudflare Pages y llama a `/api` en **su propio dominio**; Cloudflare enruta
 * al túnel del VPS. El backend no expone puerto (`arena/CLAUDE.md` §6).
 *
 * `useMockBackend` es la única línea que hay que tocar para apuntar al Go real.
 * Está en `true` mientras el backend se construye: la app entera —registro,
 * apuestas, carrera en vivo, panel— funciona contra el mock, así el frontend no
 * queda bloqueado esperando.
 */
export const environment = {
  apiBaseUrl: '/api',
  useMockBackend: false,
} as const;
