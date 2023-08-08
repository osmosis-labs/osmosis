# Weighted Pool

Liquidity pools are clusters of tokens with pre-determined weights. A token's weight is how much its value accounts for the total value within the pool. For example, Uniswap pools involve two tokens with 50-50 weights. The total value of Asset A must remain equal to the total value of Asset B. Other token weights are possible, such as 90-10. It is also possible to have a liquidity pool with more than two assets.

In Osmosis, pool creators are allowed to choose the tokens within the pool and their respective weights. The parameters chosen by the pool creator cannot be changed. Other users can create separate pools with different parameters.

Weighted Pools are an extension of the classical  AMM pools popularized by Uniswap v1. Weighted Pools are great for general cases, including tokens that don't necessarily have any price correlation (ex. DAI/WETH). Unlike pools in other AMMs that only provide 50/50 weightings, Osmosis Weighted Pools enable users to build pools with more than two tokens and custom weightings, such as pools with 80/20 or 60/20/20 weightings.
