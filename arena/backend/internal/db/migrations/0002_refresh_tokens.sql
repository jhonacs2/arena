-- Arena — sesiones (refresh tokens de un solo uso)
--
-- Esta tabla NO está en docs/contract/schema.sql. El contrato pide un refresh
-- «de un solo uso en cookie HttpOnly» (api.md) y eso necesita estado del lado
-- del servidor: un JWT no se puede invalidar, y «de un solo uso» es exactamente
-- invalidarlo al canjearlo.
--
-- Vive acá y no en el contrato porque es un detalle de implementación de la
-- autenticación y no una regla del negocio: no hay ninguna decisión de
-- decisiones.md que hable de sesiones. Si se decide que sí corresponde al
-- contrato, se mueve tal cual a schema.sql y este archivo queda vacío.
--
-- Idempotente, como todo lo que aplica el runner.

create table if not exists refresh_tokens (
  -- El HASH del token, no el token. Un volcado de la base no tiene que alcanzar
  -- para hacerse pasar por un alumno. SHA-256 sin salt está bien acá y no lo
  -- estaría para una contraseña: el token tiene 192 bits de entropía y no se
  -- adivina con un diccionario.
  token_hash text primary key,

  user_id    uuid not null references users (id) on delete cascade,

  issued_at  timestamptz not null default now(),
  expires_at timestamptz not null,

  -- Cuándo se canjeó. La fila NO se borra al usarla: se marca. Un intento de
  -- reusar un token ya canjeado es la señal de que alguien se lo robó, y para
  -- verla hay que poder distinguir «ya se usó» de «no existe».
  used_at    timestamptz,

  constraint expira_despues_de_emitir check (expires_at > issued_at)
);

create index if not exists refresh_tokens_user_idx on refresh_tokens (user_id);

-- Para la limpieza del arranque: borrar los vencidos sin recorrer la tabla.
create index if not exists refresh_tokens_expires_idx on refresh_tokens (expires_at);
