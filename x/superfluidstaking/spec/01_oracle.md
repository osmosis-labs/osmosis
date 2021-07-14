# Oracle

Oracles take really important role in superfluid staking as it determines the power of LP tokens in staking.
**Note:** We can leave the exact conversion rate between GAMM share to osmo equivalent to later.

## LP token price

## Individual coins' balance per LP token

## LP token price volatility

Price array: p1, p2, p3, p4, p5 ...
Average price: p
Volatility: sum((p1-p)/p)

## How many epochs or how long will be ideal for TWAP calculation of LP token prices?

Should this be modifiable by param?

## Should we consider current price on TWAP calculation?

What if we just provide incentives just for previous values?
Previous values are verified contributions to Osmosis and we can surely provide incentives for verified contributions but for current prices, it could be used for hack and price manipulation.

How to detect current price is being manipulated for next round of epoch to hack the chain?
How to handle the sudden price dump of a coin, let's say AKT dumped for its fault?
- We don't care about current AKT price as OSMO was there last time for Osmosis chain security. This will be considered on later epochs.

## What are the params to be considered for `risk_weight` calculation?

1. The quote token that is paired with `OSMO` e.g. ATOM or AKT
2. The token weight in the pair e.g. 50:50 or 80:20 pair
3. Price volatility of quote token
4. Total liquidity amount available on the pool
5. Total liquidity amount locked for `UNBONDING` time
6. Slash amount per slash action on staking

## Reference projects that calculate LP token price on-chain for lending or other purposes

https://gist.github.com/l3wi/0871d6f3b2f9a60845a5bcdaf179ba88
https://github.com/alpaca-finance/bsc-alpaca-contract/blob/main/contracts/6/protocol/OracleMedianizer.sol