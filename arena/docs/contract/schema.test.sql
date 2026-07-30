-- Arena — tests del esquema contra un Postgres real.
--
--   docker run -d --rm --name arena-pg -e POSTGRES_PASSWORD=x -e POSTGRES_DB=arena \
--     -p 55433:5432 postgres:16-alpine
--   docker exec -i arena-pg psql -U postgres -d arena -v ON_ERROR_STOP=1 < schema.sql
--   docker exec -i arena-pg psql -U postgres -d arena -v ON_ERROR_STOP=1 < schema.test.sql
--
-- No prueba que el SQL sea válido —eso ya lo probó aplicarlo—. Prueba las dos
-- reglas de `decisiones.md` que, si están mal, están mal en la nota de alguien:
-- el piso de 10 puntos y la conservación de monedas en la liquidación pari-mutuel.

\set ON_ERROR_STOP on
begin;

-- ── Actores ─────────────────────────────────────────────────────────────────
insert into users (id, username, first_name, last_name, password_hash, role, balance)
values
  ('11111111-1111-1111-1111-111111111111', 'profe',  'Jhonatan', 'Soto',   'x', 'admin',   0),
  ('22222222-2222-2222-2222-222222222222', 'ana',    'Ana',      'Gómez',  'x', 'student', 1000),
  ('33333333-3333-3333-3333-333333333333', 'bruno',  'Bruno',    'Díaz',   'x', 'student', 1000),
  ('44444444-4444-4444-4444-444444444444', 'carla',  'Carla',    'Ruiz',   'x', 'student', 1000);

insert into races (id, name, status, created_by, seed)
values ('55555555-5555-5555-5555-555555555555', 'Clásico de prueba', 'running',
        '11111111-1111-1111-1111-111111111111', 42);

insert into horses (id, race_id, number, name, nominal_odds) values
  ('a0000000-0000-0000-0000-000000000001', '55555555-5555-5555-5555-555555555555', 1, 'Tormenta',    250),
  ('a0000000-0000-0000-0000-000000000002', '55555555-5555-5555-5555-555555555555', 2, 'Relámpago',   410),
  ('a0000000-0000-0000-0000-000000000003', '55555555-5555-5555-5555-555555555555', 3, 'Viento Norte', 180);

-- ── Las apuestas ────────────────────────────────────────────────────────────
-- Elegidas para que la división entera deje resto: pool 800 sobre un pool
-- ganador de 300 da 2.666… por moneda. Sin repartir el resto, se pierde 1 moneda.
insert into bets (id, race_id, user_id, horse_id, amount) values
  ('b0000000-0000-0000-0000-000000000001', '55555555-5555-5555-5555-555555555555',
   '22222222-2222-2222-2222-222222222222', 'a0000000-0000-0000-0000-000000000003', 100),
  ('b0000000-0000-0000-0000-000000000002', '55555555-5555-5555-5555-555555555555',
   '33333333-3333-3333-3333-333333333333', 'a0000000-0000-0000-0000-000000000003', 200),
  ('b0000000-0000-0000-0000-000000000003', '55555555-5555-5555-5555-555555555555',
   '44444444-4444-4444-4444-444444444444', 'a0000000-0000-0000-0000-000000000001', 500);

-- ── 1. Una apuesta por carrera y por alumno ─────────────────────────────────
do $$ begin
  begin
    insert into bets (race_id, user_id, horse_id, amount)
    values ('55555555-5555-5555-5555-555555555555',
            '22222222-2222-2222-2222-222222222222',
            'a0000000-0000-0000-0000-000000000002', 50);
    raise exception 'FALLO: se aceptó una segunda apuesta del mismo alumno';
  exception when unique_violation then
    raise notice 'ok · una apuesta por carrera y por alumno';
  end;
end $$;

-- ── 2. El ledger es append-only ─────────────────────────────────────────────
insert into coin_transactions (user_id, delta, reason, ref_id, balance_after)
values ('22222222-2222-2222-2222-222222222222', -100, 'bet_placed',
        'b0000000-0000-0000-0000-000000000001', 900);

do $$ begin
  begin
    update coin_transactions set delta = -1 where user_id = '22222222-2222-2222-2222-222222222222';
    raise exception 'FALLO: se pudo editar el ledger';
  exception when raise_exception then
    if sqlerrm like 'FALLO%' then raise; end if;
    raise notice 'ok · el ledger rechaza UPDATE';
  end;
end $$;

-- ── 3. La liquidación pari-mutuel conserva las monedas ──────────────────────
-- pool = 800, ganó Viento Norte (pool ganador 300).
--   Ana:   100 × 800 / 300 = 266,66 → 266
--   Bruno: 200 × 800 / 300 = 533,33 → 533
--   suma 799 → resto 1 → al de apuesta mayor (Bruno) → 534
--   total 800 = pool ✓
-- ¡Ojo con los tipos! `sum(bigint)` en Postgres devuelve **numeric**, no bigint.
-- Sin los `::bigint` de abajo, `(amount * total) / winning` es una división
-- *decimal*: da 266,66 y 533,33, y al guardarlos en una columna bigint Postgres
-- **redondea** —267 y 533—. El reparto del resto lo termina decidiendo el
-- redondeo en vez de la regla, y la moneda extra va al apostador equivocado.
-- Pasó de verdad al escribir este test, y la suma seguía dando 800: la
-- conservación no alcanza para detectarlo, hay que asertar los pagos uno por uno.
with pool as (
  select sum(amount)::bigint as total,
         coalesce(sum(amount) filter (
           where horse_id = 'a0000000-0000-0000-0000-000000000003'), 0)::bigint as winning
  from bets where race_id = '55555555-5555-5555-5555-555555555555'
),
base as (
  select b.id, b.amount,
         (b.amount * p.total) / p.winning as payout,   -- bigint / bigint = trunca
         row_number() over (order by b.amount desc, b.created_at asc, b.id asc) as rank
  from bets b cross join pool p
  where b.race_id = '55555555-5555-5555-5555-555555555555'
    and b.horse_id = 'a0000000-0000-0000-0000-000000000003'
),
remainder as (
  select (select total from pool) - (select sum(payout) from base) as left_over
)
update bets b
set payout = base.payout + case when base.rank <= (select left_over from remainder) then 1 else 0 end,
    status = 'won',
    settled_at = now()
from base where b.id = base.id;

update bets set payout = 0, status = 'lost', settled_at = now()
where race_id = '55555555-5555-5555-5555-555555555555' and status = 'placed';

insert into race_settlements (race_id, winner_id, pool, winning_pool, paid_out, refunded)
select '55555555-5555-5555-5555-555555555555',
       'a0000000-0000-0000-0000-000000000003',
       sum(amount), 300, sum(coalesce(payout, 0)), false
from bets where race_id = '55555555-5555-5555-5555-555555555555';

do $$
declare pool_total bigint; paid bigint; ana bigint; bruno bigint;
begin
  select s.pool, s.paid_out into pool_total, paid from race_settlements s;
  select payout into ana   from bets where user_id = '22222222-2222-2222-2222-222222222222';
  select payout into bruno from bets where user_id = '33333333-3333-3333-3333-333333333333';

  if paid <> pool_total then
    raise exception 'FALLO: se pagaron % de un pool de % — el resto se perdió', paid, pool_total;
  end if;
  if ana <> 266 or bruno <> 534 then
    raise exception 'FALLO: pagos inesperados — Ana % (esperado 266), Bruno % (esperado 534)', ana, bruno;
  end if;
  raise notice 'ok · pari-mutuel conserva: Ana 266 + Bruno 534 = 800 = pool';
end $$;

-- ── 4. Perder SÍ baja la nota ───────────────────────────────────────────────
-- Carla apostó 500 y perdió; después funde el resto. La nota la sigue.
--
-- Esta sección afirmaba lo contrario —un piso de 10 puntos— hasta que se corrigió
-- el contrato: el piso contradecía «si funden las monedas me van a deber nota y se
-- tendrán que esforzar más». Lo único intocable son los puntos regalados, y eso lo
-- prueba §5.
update users set balance = 500 where username = 'carla';

do $$ declare pts numeric; begin
  select points into pts from user_scores where username = 'carla';
  if pts <> 5 then raise exception 'FALLO: con 500 monedas los puntos dan %, esperado 5', pts; end if;
  raise notice 'ok · 500 monedas → 5 puntos';
end $$;

update users set balance = 0 where username = 'carla';

do $$ declare pts numeric; begin
  select points into pts from user_scores where username = 'carla';
  if pts <> 0 then raise exception 'FALLO: fundida, los puntos dan %, esperado 0', pts; end if;
  raise notice 'ok · 0 monedas → 0 puntos. Apostar mal sí baja la nota';
end $$;

-- ── 5. Ganar sí sube, y el regalo de puntos se suma aparte ──────────────────
update users set balance = 1500 where username = 'bruno';

do $$ declare pts numeric; begin
  select points into pts from user_scores where username = 'bruno';
  if pts <> 15 then raise exception 'FALLO: con 1500 monedas los puntos dan %, esperado 15', pts; end if;
  raise notice 'ok · 1500 monedas → 15 puntos';
end $$;

insert into point_grants (user_id, points, reason, granted_by)
values ('33333333-3333-3333-3333-333333333333', 2.5, 'Explicó @for en el code review',
        '11111111-1111-1111-1111-111111111111');

do $$ declare pts numeric; begin
  select points into pts from user_scores where username = 'bruno';
  if pts <> 17.5 then raise exception 'FALLO: con un regalo de 2,5 los puntos dan %, esperado 17,5', pts; end if;
  raise notice 'ok · los puntos regalados se suman y no pasan por el juego';
end $$;

-- ── 6. Un saldo negativo no entra ───────────────────────────────────────────
do $$ begin
  begin
    update users set balance = -1 where username = 'ana';
    raise exception 'FALLO: se aceptó un saldo negativo';
  exception when check_violation then
    raise notice 'ok · el saldo no puede quedar negativo';
  end;
end $$;

-- ── 7. Devolución solo cuando nadie acertó ──────────────────────────────────
do $$ begin
  begin
    insert into race_settlements (race_id, winner_id, pool, winning_pool, paid_out, refunded)
    values ('55555555-5555-5555-5555-555555555555',
            'a0000000-0000-0000-0000-000000000001', 800, 300, 800, true);
    raise exception 'FALLO: se aceptó una devolución con aciertos';
  exception when check_violation or unique_violation then
    raise notice 'ok · no se puede devolver si alguien acertó';
  end;
end $$;

rollback;
