<!--
order: 3
-->

# End Block

## Replenish Pool
At each `EndBlock`, the value of `OsmosisPoolDelta` is decreased depending on `PoolRecoveryPeriod` of parameter.

This allows the network to sharply increase spread fees in during acute price fluctuations, and automatically return the spread to normal after some time when the price change is long term.

```go
func (k Keeper) ReplenishPools(ctx sdk.Context) {
	delta := k.GetOsmosisPoolDelta(ctx)
	regressionAmt := delta.QuoInt64(k.PoolRecoveryPeriod(ctx))

	// Replenish osmo pool towards base pool
	// regressionAmt cannot make delta zero
	delta = delta.Sub(regressionAmt)

	k.SetOsmosisPoolDelta(ctx, delta)
}
```
