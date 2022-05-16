# Concepts

The `x/gamm` module implements an AMM using Balancer style pools with
varying amounts and weights of assets in pools.

## Pool

### Creation of Pool

At an initial creation of the pool, a fixed amount of 100 share token is
minted in the pool and sent to the creator of the pool's account. The
pool share denom is in the format of `gamm/pool/{poolID}` and is
displayed in the format of `GAMM-{poolID}` to the user. Pool assets are
sorted in alphabetical order by default.

### Joining Pool

When joining a pool, a user provides the maximum amount of tokens
they're willing to deposit, while the front end takes care of the
calculation of how many share tokens the user is eligible for at the
specific moment of sending the transaction.

Calculation of exactly how many tokens are needed to get the designated
share is done at the moment of processing the transaction, validating
that it does not exceed the maximum amount of tokens the user is willing
to deposit. After the validation, GAMM share tokens of the pool are
minted and sent to the user's account. Joining the pool using a single
asset is also possible.

### Exiting Pool

When exiting a pool, the user provides the minimum amount of tokens they
are willing to receive as they are returning their shares of the pool.
However, unlike joining a pool, exiting a pool requires the user to pay
the exit fee, which is set as a param of the pool. The user's share
tokens burnt as result. Exiting the pool using a single asset is also
possible.

+++<https://github.com/osmosis-labs/osmosis/blob/main/x/gamm/keeper/pool_service.go>

## Swap

During the process of swapping a specific asset, the token the user is
putting into the pool is denoted as `tokenIn`, while the token that
would be returned to the user, the asset that is being swapped for,
after the swap is denoted as `tokenOut` throughout the module.

Given a `tokenIn`, the following calculations are done to calculate how
many tokens are to be swapped into and removed from the pool:

`tokenBalanceOut * [1 - { tokenBalanceIn / (tokenBalanceIn + (1 - swapFee) * tokenAmountIn)} ^ (tokenWeightIn / tokenWeightOut)]`

The calculation is also able to be reversed, the case where user
provides `tokenOut`. The calculation for the amount of tokens that the
user should be putting in is done through the following formula:

`tokenBalanceIn * [{tokenBalanceOut / (tokenBalanceOut - tokenAmountOut)} ^ (tokenWeightOut / tokenWeightIn) -1] / tokenAmountIn`

### Spot Price

Meanwhile, calculation of the spot price with a swap fee is done using
the following formula:

`spotPrice / (1 - swapFee)`, where `spotPrice` is defined as:

`(tokenBalanceIn / tokenWeightIn) / (tokenBalanceOut / tokenWeightOut)`

+++<https://github.com/osmosis-labs/osmosis/blob/main/x/gamm/keeper/swap.go>

### Multi-Hop

All tokens are swapped using a multi-hop mechanism. That is, all swaps
are routed via the most cost-efficient way, swapping in and out from
multiple pools in the process.

+++<https://github.com/osmosis-labs/osmosis/blob/main/x/gamm/keeper/multihop.go>
