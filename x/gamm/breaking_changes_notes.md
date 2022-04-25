# List of every known client breaking change in current AMM refactor

## Queries

* PoolAssetsQuery
  * Deleted
  * New query is TotalLiquidity
  * If you wanted the pool weights, you must now query the pool itself.
    * Please give feedback if more queries should be exposed that are pool-type specific
* QuerySpotPrice
  * The `withswapfee` param is now removed. If this was needed for anything, please flag it. Its mainly removed due to not having a clear use, and a better query can probably be crafted for.
  * Rename TokenInDenom to QuoteAssetDenom
  * Rename TokenOutDenom to BaseAssetDenom

## Messages

* JoinPoolNoSwap
  * TokenInMaxs must either contain every token in pool, or no tokens
    * Before it could just apply a max constraint on one input token.
* ExitPool
  * Before the message would fail if you had too few shares to get a single token out for any given denom. Now you can get 0 tokens of one side out, if the min amount is also not present.
* ExitSwapShareAmountIn
  * Switched to a more inefficient algorithm for now, so gas numbers will be much higher.
* MsgSwapExactAmountOut
  * Prior behavior rounded down the required AmountIn input. The logic now rounds up. Any prior test vectors will likely be off by one.
* Messages now have responses

## Events

## Error message

I anticipate there are lots of error messages that have changed. This is a best-attempt to log ones we know that we changed

* ExitPool when slippage was too high

## Gas numbers

Many are changed, need to re-review what the new normals are for each operation.

## Questions for integrators

* Would it be problematic if we renamed the message name / amino route of JoinPool to JoinPoolNoSwap
