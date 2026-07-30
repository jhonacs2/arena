-- Arena — reconciliación del ledger
--
-- `users.balance` es una caché: la verdad es la suma de `coin_transactions.delta`
-- (ver el comentario de la columna en docs/contract/schema.sql). Esta consulta
-- compara las dos y **tiene que devolver cero filas**. Una fila acá significa que
-- alguien tiene una nota que el ledger no explica.
--
--   psql "$DATABASE_URL" -f arena/scripts/reconcile.sql
--
-- Lo corre también `node scripts/verify-arena.mjs`, que falla si sale una sola
-- fila. Si sale alguna: **no se edita el ledger** —el trigger lo impide y además
-- reescribir la historia de la nota de alguien es lo peor que se puede hacer acá—
-- se agrega un movimiento `adjustment` que compense la diferencia.

\set ON_ERROR_STOP on

-- ─────────────────────────────────────────────────────────────────────────────
-- 1. El saldo cacheado contra la suma del ledger
-- ─────────────────────────────────────────────────────────────────────────────

with ledger as (
  select
    user_id,
    sum(delta)::bigint as ledger_balance,
    count(*)           as movements
  from coin_transactions
  group by user_id
)
select
  'balance_vs_ledger'                              as problem,
  u.id                                             as user_id,
  u.username,
  u.balance                                        as cached_balance,
  coalesce(l.ledger_balance, 0)                    as ledger_balance,
  u.balance - coalesce(l.ledger_balance, 0)        as difference,
  coalesce(l.movements, 0)                         as movements
from users u
left join ledger l on l.user_id = u.id
-- Un usuario sin ningún movimiento tiene que estar en 0: el canje del código
-- acredita las 1000 monedas en el ledger, así que un saldo sin ledger detrás es
-- exactamente el bug que esto busca.
where u.balance <> coalesce(l.ledger_balance, 0)
order by abs(u.balance - coalesce(l.ledger_balance, 0)) desc;

-- ─────────────────────────────────────────────────────────────────────────────
-- 2. La cadena de balance_after
-- ─────────────────────────────────────────────────────────────────────────────
--
-- Cada movimiento guarda el saldo que dejó. Si el ledger se escribió bien, cada
-- `balance_after` es el anterior más el `delta`. Esto encuentra el movimiento
-- exacto donde se rompió, que es lo que hace falta para escribir la compensación.

with chained as (
  select
    id,
    user_id,
    delta,
    balance_after,
    lag(balance_after, 1, 0::bigint) over (partition by user_id order by id) as previous_balance
  from coin_transactions
)
select
  'balance_after_chain'                as problem,
  c.id                                 as transaction_id,
  c.user_id,
  c.previous_balance,
  c.delta,
  c.balance_after,
  c.previous_balance + c.delta         as expected_balance_after
from chained c
where c.balance_after <> c.previous_balance + c.delta
order by c.user_id, c.id;

-- ─────────────────────────────────────────────────────────────────────────────
-- 3. El último balance_after contra el saldo cacheado
-- ─────────────────────────────────────────────────────────────────────────────
--
-- Redundante con (1) si (2) pasa, y ahí está el valor: si las tres consultas dan
-- vacío, las tres representaciones del saldo coinciden.

select
  'last_movement_vs_balance'  as problem,
  u.id                        as user_id,
  u.username,
  u.balance                   as cached_balance,
  t.balance_after             as last_balance_after,
  t.id                        as last_transaction_id
from users u
join lateral (
  select id, balance_after
  from coin_transactions
  where user_id = u.id
  order by id desc
  limit 1
) t on true
where u.balance <> t.balance_after
order by u.username;
