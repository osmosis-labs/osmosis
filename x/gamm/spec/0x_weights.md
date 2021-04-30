<!--
order: 2
-->

# Weights

Weights refer to the how we weight the reserves of assets within a pool.
Its often convenient to think of weights in terms of ratios,
so a 1:1 pool between "ATOM" and "BTC" is a pool where the spot price is
`#ATOM in pool / #BTC in pool`.

A 2:1 pool is one where the spot price is
`2*(#ATOM in pool) / #BTC in pool`.
This weights allows one to get the same spot price in the pool,
with fewer ATOM reserves.
(This comes at the cost of having higher slippage,
e.g. buying 10 atoms moves the price more than a 1:1 pool with the same BTC liquidity).

Within the state machine, we represent weights as numbers, and the ratios are computed internally.
So you could specify a pool between three assets, with weights 100:30:50,
which is equivalent to a 10:3:5 pool.

The ratios provided in a CreatePool message, or governance proposal are capped at 2^20.
However, within the state machine they are stored with an extra 30 bits of precision,
allowing for smooth changes between two weights to happen with sufficient granularity.

(Note, these docs are intended to get shuffled around as we write more of the spec for x/gamm.
I just wanted to document this along with the PR, to save work for our future selves)