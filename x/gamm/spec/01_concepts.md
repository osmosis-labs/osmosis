<!--
order: 1
-->

# Concepts

The concept of the `gamm` module is designed to handle assets of a chain using the AMM and its concept of pool shares. 

## Pool

### Creation of Pool

At an initial creation of the pool, a fixed amount of 100 share token is minted in the pool and sent to the creator of the pool's account. Pool share denom is in the format of gamm/pool/{poolId} and is displayed in the format of GAMM-{poolId} to the user. Pool assets are sorted in alphabetical order by defualt.

### Joining Pool

When joining a pool, users provide maximum amount of tokens willing to deposit, while the front end takes care of the calculation of how many share tokens the user is eligible at the specific moment of sending the transaction. Calculation of exactly how many tokens are needed to get the designated share is done at the moment of procssing the transaction, validating that it does not exceed the maximum amount of token the user is willing to deposit. After the validation, the share of the pool is minted and sent to the user account. Joining the pool using a single asset is also possible.

### Exiting Pool

When exiting the pool, the user also probides the minimum amount of tokens they are willing to receive as they are returning the share of the pool. However, unlike joining a pool, exiting a pool requires the user to pay the exit fee, which is set as the param of the pool. The share of the user gets burnt. Exiting the pool using a single asset is also possible. 

+++[https://github.com/osmosis-labs/osmosis/blob/main/x/gamm/keeper/pool_service.go](https://github.com/osmosis-labs/osmosis/blob/main/x/gamm/keeper/pool_service.go)

## Swap

During the process of swapping a specific asset, the token user is putting into the pool is justified as `tokenIn`, while the token that would be omitted after the swap is justified as `tokenOut`  throughout the module.

Given a tokenIn, the following calculations are done to calculate how much tokens are to be swapped and ommitted from the pool.

- `tokenBalanceOut * [ 1 - { tokenBalanceIn / (tokenBalanceIn+(1-swapFee) * tokenAmountIn)}^(tokenWeightIn/tokenWeightOut)]`

The whole process is also able vice versa, the case where user provides tokenOut. The calculation  for the amount of token that the user should be putting in is done through the following formula.

- `tokenBalanceIn * [{tokenBalanceOut / (tokenBalanceOut - tokenAmountOut)}^(tokenWeightOut/tokenWeightIn)-1] / tokenAmountIn`

### Spot Price

Meanwhile, calculation of the spot price with a swap fee is done using the following formula

- `spotPrice / (1-swapFee)`

where spotPrice is 

- `(tokenBalanceIn / tokenWeightIn) / (tokenBalanceOut / tokenWeightOut)`

+++[https://github.com/osmosis-labs/osmosis/blob/main/x/gamm/keeper/swap.go](https://github.com/osmosis-labs/osmosis/blob/main/x/gamm/keeper/swap.go)

### Multihop

All tokens are swapped using multi-hop. That is, all swaps are routed via the ultimate cost-efficient way, swapping in and out from multiple pools in the process.

+++[https://github.com/osmosis-labs/osmosis/blob/main/x/gamm/keeper/multihop.go](https://github.com/osmosis-labs/osmosis/blob/main/x/gamm/keeper/multihop.go)