# TWAP (Time Weighted Average Price)

The TWAP package is responsible for being able to serve TWAPs for every AMM pool.

A time weighted average price is a function that takes a sequence of `(time, price)` pairs, and returns a price representing an 'average' over the entire time period. The method of averaging can vary from the classic arithmetic mean, (such as geometric mean, harmonic mean), however we currently only implement arithmetic mean.

## Arithmetic mean TWAP

Using the arithmetic mean, the TWAP of a sequence `(t_i, p_i)`, from `t_0` to `t_n`, indexed by time in ascending order, is: $\frac{1}{t_n - t_0}\sum_{i=0}^{n-1} p_i (t_{i+1} - t_i)$. (Notice that the latest price `p_n` isn't used, as it has lasted for a time interval of `0` seconds in this range!)

To illustrate with an example, given the sequence: `(0s, $1), (2s, $5), (5s, $1)`, the arithmetic mean TWAP is: $\frac{\$1 * (2s - 0s) + \$5 * (5s - 3s)}{5s - 0s} = \frac{\$10}{5} = \$2$.

## Computation via accumulators method

The prior example for how to compute the TWAP takes linear time, which is unsuitable for use in a blockchain setting.

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

## Testing Methodology

The pre-release testing methodology planned for the twap module is:

- [ ] Using table driven unit tests to test all foreseen cases the module can be within
    - hook testing
        - All swaps correctly trigger twap record updates
        - Create pools cause records to be created
    - store
        - EndBlock triggers all relevant twaps to be saved correctly
        - Block commit wipes temporary stores
    - logic
        - Make tables of expected input / output cases for:
          - getMostRecentRecord
          - getInterpolatedRecord
          - updateRecord
          - computeArithmeticTwap
        - Test overflow handling in all relevant arithmetic
        - Complete testing code coverage (up to return err lines) for logic.go file
    - API
        - Unit tests for the public API, under foreseeable setup conditions
- [ ] Integration into the Osmosis simulator
    - The osmosis simulator, simulates building up complex state machine states, in random ways not seen before. We plan on, in a property check, maintaining expected TWAPs for short time ranges, and seeing that the keeper query will return the same value as what we get off of the raw price history for short history intervals.
- [ ] Mutation testing usage
    - integration of the TWAP module into go mutation testing: https://github.com/osmosis-labs/go-mutesting
        - The success we've seen with the tokenfactory module, is it succeeds at surfacing behavior for untested behavior.
          e.g. if you delete a line, or change the direction of a conditional, does a test catch it.
    - We expect to get this to a state, where after mutation testing is ran, the only items it mutates, that is not caught in a test, is: Deleting `return err`, or `panic` lines, in the situation where that error return or panic isn't reachable.