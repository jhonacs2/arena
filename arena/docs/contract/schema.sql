-- Arena — esquema Postgres
--
-- Postgres corre en el **mismo VPS** que el backend. No es Supabase: su plan
-- gratuito pausa el proyecto tras 7 días sin actividad y una clase semanal cae
-- justo en esa ventana, y con el bucle de carrera a 10 Hz una base en otra red
-- paga un viaje de red por consulta. Nada de acá usa extensiones de Supabase, así
-- que mover la base es cambiar `DATABASE_URL`.
--
-- Las reglas que este esquema hace cumplir están en decisiones.md. Todo lo que se
-- pueda expresar como restricción de la base va acá y no en Go: una regla en el
-- esquema no se puede saltear por un camino que nadie revisó.
--
-- Se aplica en orden. Idempotente: se puede correr dos veces.

-- ─────────────────────────────────────────────────────────────────────────────
-- Tipos
-- ─────────────────────────────────────────────────────────────────────────────

do $$ begin
  create type user_role as enum ('student', 'admin');
exception when duplicate_object then null; end $$;

do $$ begin
  create type race_status as enum ('draft', 'open', 'running', 'finished', 'cancelled');
exception when duplicate_object then null; end $$;

do $$ begin
  create type bet_status as enum ('placed', 'won', 'lost', 'refunded');
exception when duplicate_object then null; end $$;

-- El motivo de cada movimiento del ledger. Enum y no texto libre: el panel del
-- instructor agrupa por esto y un typo silencioso arruinaría el informe.
do $$ begin
  create type ledger_reason as enum (
    'code_redeemed',   -- las 1000 monedas iniciales
    'gift',            -- regalo del instructor
    'bet_placed',      -- se descuenta al apostar
    'bet_won',         -- se acredita el pago
    'bet_refunded',    -- carrera cancelada
    'adjustment'       -- corrección manual del instructor
  );
exception when duplicate_object then null; end $$;

-- ─────────────────────────────────────────────────────────────────────────────
-- Usuarios
-- ─────────────────────────────────────────────────────────────────────────────

create table if not exists users (
  id            uuid primary key default gen_random_uuid(),
  username      text not null,
  first_name    text not null,
  last_name     text not null,
  password_hash text not null,
  role          user_role not null default 'student',

  -- Caché del ledger, mantenida en la misma transacción que el movimiento. La
  -- verdad es la suma de coin_transactions.delta; esto existe para no sumar el
  -- ledger entero en cada request. scripts/reconcile.sql compara los dos.
  balance       bigint not null default 0,

  created_at    timestamptz not null default now(),

  -- Piso en 0: ver decisiones.md §1. Si alguna vez se quiere deuda literal, esta
  -- línea y el tope del monto de la apuesta son los dos únicos lugares a tocar.
  constraint balance_no_negativo check (balance >= 0)
);

-- Case-insensitive: nadie va a recordar si se registró como "Ana" o "ana".
create unique index if not exists users_username_key on users (lower(username));

-- ─────────────────────────────────────────────────────────────────────────────
-- Códigos de invitación
-- ─────────────────────────────────────────────────────────────────────────────

create table if not exists invite_codes (
  code          text primary key,
  coins_granted bigint not null default 1000,
  note          text,                    -- para el instructor: «grupo del martes»

  redeemed_by   uuid references users (id),
  redeemed_at   timestamptz,

  created_at    timestamptz not null default now(),
  created_by    uuid not null references users (id),

  constraint code_formato check (code ~ '^[A-Z]{4}-[0-9]{4}$'),
  constraint coins_positivo check (coins_granted > 0),

  -- Un código está canjeado o no lo está. No hay medio canje: las dos columnas
  -- se llenan juntas o ninguna.
  constraint canje_consistente check (
    (redeemed_by is null and redeemed_at is null) or
    (redeemed_by is not null and redeemed_at is not null)
  )
);

-- Un usuario canjea un solo código.
create unique index if not exists invite_codes_redeemed_by_key
  on invite_codes (redeemed_by) where redeemed_by is not null;

-- ─────────────────────────────────────────────────────────────────────────────
-- Ledger de monedas — append-only
-- ─────────────────────────────────────────────────────────────────────────────

create table if not exists coin_transactions (
  id            bigserial primary key,
  user_id       uuid not null references users (id),
  delta         bigint not null,          -- positivo acredita, negativo descuenta
  reason        ledger_reason not null,

  -- A qué apunta: el id de la apuesta, o el código canjeado. Sin FK porque
  -- referencia tablas distintas según el motivo; el backend lo llena y el panel
  -- del instructor lo usa para armar el rastro.
  ref_id        text,

  balance_after bigint not null,
  created_at    timestamptz not null default now(),
  created_by    uuid references users (id),  -- el instructor, si fue regalo o ajuste

  constraint delta_no_cero check (delta <> 0),
  constraint balance_after_no_negativo check (balance_after >= 0)
);

create index if not exists coin_transactions_user_idx
  on coin_transactions (user_id, id desc);

-- Append-only de verdad, no por convención: sin esto, un UPDATE mal escrito
-- reescribe la historia de la nota de alguien y nada queda registrado.
create or replace function ledger_solo_insert() returns trigger as $$
begin
  raise exception 'coin_transactions es append-only: compensá con otro movimiento';
end $$ language plpgsql;

drop trigger if exists ledger_sin_update on coin_transactions;
create trigger ledger_sin_update before update or delete on coin_transactions
  for each row execute function ledger_solo_insert();

-- ─────────────────────────────────────────────────────────────────────────────
-- Carreras y caballos
-- ─────────────────────────────────────────────────────────────────────────────

create table if not exists races (
  id           uuid primary key default gen_random_uuid(),
  name         text not null,
  status       race_status not null default 'draft',
  scheduled_at timestamptz,

  created_by   uuid not null references users (id),
  created_at   timestamptz not null default now(),
  opened_at    timestamptz,
  started_at   timestamptz,
  finished_at  timestamptz,

  -- Semilla de la simulación. Se fija al largar, y con ella la carrera es
  -- reproducible: si alguien reclama el resultado, se vuelve a correr igual.
  seed         bigint,

  constraint nombre_no_vacio check (length(trim(name)) > 0)
);

create index if not exists races_status_idx on races (status, scheduled_at);

create table if not exists horses (
  id      uuid primary key default gen_random_uuid(),
  race_id uuid not null references races (id) on delete cascade,
  number  int not null,
  name    text not null,

  -- Cuota **nominal**, ×100 en entero: 3.40 se guarda 340.
  --
  -- Con pari-mutuel esta cuota NO determina el pago: el pago sale del pool. Sirve
  -- para dos cosas: alimentar al simulador —es su parámetro de fuerza, el mismo
  -- que usa `docs/contract/race-simulation.md`— y mostrarle al alumno qué caballo
  -- es favorito antes de que haya apuestas suficientes para que el pool diga algo.
  nominal_odds int not null,

  constraint nominal_odds_minima check (nominal_odds >= 101),
  constraint number_positivo check (number >= 1),
  unique (race_id, number)
);

-- ─────────────────────────────────────────────────────────────────────────────
-- Salas
-- ─────────────────────────────────────────────────────────────────────────────

create table if not exists race_participants (
  race_id   uuid not null references races (id) on delete cascade,
  user_id   uuid not null references users (id),
  joined_at timestamptz not null default now(),
  primary key (race_id, user_id)
);

-- ─────────────────────────────────────────────────────────────────────────────
-- Apuestas
-- ─────────────────────────────────────────────────────────────────────────────

create table if not exists bets (
  id           uuid primary key default gen_random_uuid(),
  race_id      uuid not null references races (id),
  user_id      uuid not null references users (id),
  horse_id     uuid not null references horses (id),

  amount       bigint not null,
  status       bet_status not null default 'placed',

  -- Se llena al liquidar. En pari-mutuel el pago **no se puede saber al apostar**:
  -- depende de cómo apostó el resto. Por eso acá no hay cuota congelada — no
  -- existe una cuota al momento de la apuesta que tenga sentido guardar.
  payout       bigint,

  created_at   timestamptz not null default now(),
  settled_at   timestamptz,

  constraint amount_positivo check (amount >= 1),

  -- Una apuesta por carrera y por alumno. En pari-mutuel esto no impide un
  -- arbitraje —cubrir todos los caballos devuelve el pool menos la parte de los
  -- demás, que es una pérdida esperada, no una ganancia—. Está por dos razones
  -- distintas: que la decisión importe, y que el pool no lo domine quien más
  -- apueste en cantidad de boletos.
  unique (race_id, user_id)
);

create index if not exists bets_race_idx on bets (race_id);
create index if not exists bets_user_idx on bets (user_id, created_at desc);

-- ─────────────────────────────────────────────────────────────────────────────
-- Resultados
-- ─────────────────────────────────────────────────────────────────────────────

create table if not exists race_results (
  race_id  uuid not null references races (id) on delete cascade,
  horse_id uuid not null references horses (id),
  position int not null,
  primary key (race_id, horse_id),

  constraint position_positiva check (position >= 1),
  unique (race_id, position)
);

-- ─────────────────────────────────────────────────────────────────────────────
-- Liquidación pari-mutuel — una fila por carrera
-- ─────────────────────────────────────────────────────────────────────────────

-- Lo apostado se reparte entre quienes acertaron; no hay casa. El pago de cada
-- ganador es `amount * pool / winning_pool`, división entera.
--
-- La división entera deja un resto (en las carreras reales se llama *breakage*).
-- Ese resto **no se descarta**: se reparte de a una moneda entre los ganadores
-- ordenados por (amount desc, created_at asc, id asc) hasta agotarlo. Si se
-- descartara, el pari-mutuel dejaría de ser suma cero y el total de monedas del
-- curso bajaría sin que nadie lo haya perdido apostando.
--
-- Si nadie acertó, `winning_pool = 0` y **se devuelve cada apuesta íntegra**: no
-- hay a quién pagarle y quedarse el pool sería inventar una casa.
create table if not exists race_settlements (
  race_id      uuid primary key references races (id) on delete cascade,
  winner_id    uuid not null references horses (id),

  pool         bigint not null,   -- suma de todo lo apostado en la carrera
  winning_pool bigint not null,   -- suma de lo apostado al ganador
  paid_out     bigint not null,   -- suma de los pagos efectivos
  refunded     boolean not null default false,

  settled_at   timestamptz not null default now(),

  constraint pool_no_negativo check (pool >= 0 and winning_pool >= 0),
  constraint winning_pool_cabe check (winning_pool <= pool),

  -- La conservación, como restricción de la base y no como esperanza. Vale en los
  -- dos casos por el mismo motivo: si se devolvió, la suma de devoluciones es el
  -- pool; si se pagó, el reparto del resto agota el pool hasta la última moneda.
  constraint conserva_monedas check (paid_out = pool),

  -- Devolución y ganador conviven: hubo un ganador de la carrera, pero nadie le
  -- había apostado. `refunded` sin `winning_pool = 0` sería un reparto perdido.
  constraint devolucion_solo_sin_aciertos check (refunded = (winning_pool = 0))
);

-- ─────────────────────────────────────────────────────────────────────────────
-- Regalos de puntos — append-only, aparte de las monedas
-- ─────────────────────────────────────────────────────────────────────────────

-- Dos regalos distintos que el instructor puede hacer, y no son lo mismo:
--   regalar MONEDAS   → le da con qué seguir jugando
--   regalar PUNTOS    → le sube la nota directo, sin pasar por el juego
--
-- Separados porque tienen efectos distintos y porque un punto regalado no debería
-- poder perderse en una apuesta.
create table if not exists point_grants (
  id         bigserial primary key,
  user_id    uuid not null references users (id),
  points     numeric(6, 2) not null,
  reason     text not null,
  granted_by uuid not null references users (id),
  created_at timestamptz not null default now(),

  constraint points_no_cero check (points <> 0),
  constraint reason_no_vacio check (length(trim(reason)) > 0)
);

create index if not exists point_grants_user_idx on point_grants (user_id);

create or replace function grants_solo_insert() returns trigger as $$
begin
  raise exception 'point_grants es append-only: compensá con otro registro';
end $$ language plpgsql;

drop trigger if exists grants_sin_update on point_grants;
create trigger grants_sin_update before update or delete on point_grants
  for each row execute function grants_solo_insert();

-- ─────────────────────────────────────────────────────────────────────────────
-- Vista de puntos
-- ─────────────────────────────────────────────────────────────────────────────

-- Los puntos son una función del saldo, no una columna: dos números que
-- representan lo mismo se desincronizan siempre.
--
-- **El piso de 10 es la regla que más importa de todo el esquema.** Son los 10
-- puntos que el alumno recibe al canjear el código, y apostar mal no los puede
-- tocar: una racha perdedora le saca monedas —y con eso, capacidad de seguir
-- jugando— pero nunca calificación. Sin este `greatest`, el juego se convierte en
-- un riesgo académico y nadie se anima a apostar, que es lo contrario de lo que se
-- busca.
create or replace view user_scores as
select
  u.id,
  u.username,
  u.first_name,
  u.last_name,
  u.balance,
  greatest(10, u.balance / 100)::int as points_from_coins,
  coalesce((select sum(g.points) from point_grants g where g.user_id = u.id), 0) as points_granted,
  greatest(10, u.balance / 100)::numeric
    + coalesce((select sum(g.points) from point_grants g where g.user_id = u.id), 0) as points,
  (select count(*) from bets b where b.user_id = u.id) as bets_placed,
  (select count(*) from bets b where b.user_id = u.id and b.status = 'won') as bets_won
from users u
where u.role = 'student';
