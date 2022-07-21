# TWAP

We maintain TWAP entries for every gamm pool.

## Module API

```go
// GetArithmeticTwap returns an arithmetic time weighted average price.
// The returned twap is the time weighted average price (TWAP) of:
// * the base asset, in units of the quote asset (1 unit of base = x units of quote)
// * from (startTime, endTime),
// * as determined by prices from AMM pool `poolId`.
//
// The
//
// startTime and endTime do not have to be real block times that occurred,
// this function will interpolate between startTime.
// if endTime = now, we do {X}
// startTime must be in time range {X}, recommended parameterization for mainnet is {Y}
func (k Keeper) GetArithmeticTwap(ctx sdk.Context,
	poolId uint64,
	baseAssetDenom string, quoteAssetDenom string,
	startTime time.Time, endTime time.Time) (sdk.Dec, error) {
```

## File layout

**api.go** is the main file you should look at for what you should depend upon.

**logic.go** is the main file you should look at for how the TWAP implementation works.

- types/* - Implement TwapRecord, GenesisState. Define AMM interface, and methods to format keys.
- api.go - Public API, that other users / modules can/should depend on
- hook_listener.go - Defines hooks & calls to logic.go, for triggering actions on 
- keeper.go - generic SDK boilerplate (defining a wrapper for store keys + params)
- logic.go - Implements all TWAP module 'logic'. (Arithmetic, defining what to get/set where, etc.)
- module.go - SDK AppModule interface implementation.
- store.go - Managing logic for getting and setting things to underlying stores

## Store layout

Every pool has a TWAP stored in state for every asset pair.
