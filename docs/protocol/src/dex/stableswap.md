# Stableswap Pool

Stableswap pools offer low slippage for assets that are intended to be tightly correlated, such as stablecoins. These pools are designed around a specific price ratio where assets are expected to trade, providing minimal slippage around this target price while still maintaining price impact for each trade.

Stableswap pools implement the Solidly stableswap curve with the invariant: $f(x, y) = xy(x^2 + y^2) = k$. This is generalized to multi-asset pools as $f(a_1, ..., a_n) = a_1 * ... * a_n (a_1^2 + ... + a_n^2)$.

Unlike weighted pools that work well for uncorrelated assets, stableswap pools are optimized for assets that should maintain a relatively stable price relationship, making them ideal for stablecoin-to-stablecoin trading pairs.
