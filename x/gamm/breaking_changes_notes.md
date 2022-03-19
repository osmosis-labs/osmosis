# List of every known client breaking change in current AMM refactor

## Queries

* PoolAssetsQuery
  * ???
* QuerySpotPrice
  * The `withswapfee` param is now removed. If this was needed for anything, please flag it. Its mainly removed due to not having a clear use, and a better query can probably be crafted for.
  * Rename TokenInDenom to QuoteAssetDenom
  * Rename TokenOutDenom to BaseAssetDenom

## Messages

* (TODO) Rename JoinPool -> JoinPoolNoSwap
* JoinPoolSwapExternAmountIn
  * Replace sdk.Coin w/ sdk.Coins
  * (TODO) Consider renaming to JoinPool, hesistant due to collison with old message
* (TODO) Update the version for all of gamm's proto files
* ExitPool
  * Before the message would fail if you had too few shares to get a single token out for any given denom. Now you can 0 of one side out, if the min amount is also not present.
* ExitSwapShareAmountIn
  * Switched to a more inefficient algorithm for now, so gas numbers will be much higher.
* Messages now have responses

## Events

## Error message

I anticipate there are lots of error messages that have changed. This is a best-attempt to log ones we know that we changed

* ExitPool when slippage was too high

## Gas numbers
