# TWAP (Time Weighted Average Price)

The TWAP package is responsible for being able to serve TWAPs for every AMM pool.

A time weighted average price is a function that takes a sequence of `(time, price)` pairs, and returns a price representing an 'average' over the entire time period. The method of averaging can vary from the classic arithmetic mean, (such as geometric mean, harmonic mean), however we currently only implement arithmetic mean.

## Arithmetic mean TWAP

Using the arithmetic mean, the TWAP of a sequence `(t_i, p_i)`, from `t_0` to `t_n`, indexed by time in ascending order, is: $\frac{1}{t_n - t_0}\sum_{i=0}^{n-1} p_i (t_{i+1} - t_i)$. (Notice that the latest price `p_n` isn't used, as it has lasted for a time interval of `0` seconds in this range!)

To illustrate with an example, given the sequence: `(0s, $1), (2s, $5), (5s, $1)`, the arithmetic mean TWAP is: $\frac{\$1 * (2s - 0s) + \$5 * (5s - 3s)}{5s - 0s} = \frac{\$10}{5} = \$2$.

## Computation via accumulators method

The prior example for how to compute the TWAP takes linear time in the number of time entries in a range, which is too inefficient. We require TWAP operations to have constant time complexity (in the number of records).

This is achieved by using an accumulator. In the case of an arithmetic Twap, we can maintain an accumulator from `a_n`, representing the numerator of the Twap expression for the interval `t_0...t_n`, namely $a_n = \sum_{i=0}^{n-1} p_i (t_{i+1} - t_i)$. If we maintain such an accumulator for every pool, with `t_0 = pool_creation_time` to `t_n = current_block_time`, we can easily compute the TWAP for any interval. The twap for the time interval of price points `t_i` to `t_j` is then $twap = \frac{a_j - a_i}{t_j - t_i}$, which is constant time given the accumulator values.

In Osmosis, we maintain accumulator records for every pool, for the last 48 hours. We also maintain within each accumulator record in state, the latest spot price. This allows us to interpolate accumulation records between times. Namely, if I want the twap from `t=10s` to `t=15s`, but the time records are at `9s, 13s, 17s`, this is fine. Using the latest spot price in each record, we create the accumulator value for `t=10` by computing `a_10 = a_9 + a_9_latest_spot_price * (10s - 9s)`, and `a_15 = a_13 + a_13_latest_spot_price * (15s - 13s)`. Given these interpolated accumulation values, we can compute the TWAP as before.


## Module API

The primary intended API is `GetArithmeticTwap`, which is documented below, and has a similar cosmwasm binding.

```go
// GetArithmeticTwap returns an arithmetic time weighted average price.
// The returned twap is the time weighted average price (TWAP) of:
// * the base asset, in units of the quote asset (1 unit of base = x units of quote)
// * from (startTime, endTime),
// * as determined by prices from AMM pool `poolId`.
//
// startTime and endTime do not have to be real block times that occurred,
// the state machine will interpolate the accumulator values for those times
// from the latest Twap accumulation record prior to the provided time.
//
// startTime must be within 48 hours of ctx.BlockTime(), if you need older TWAPs,
// you will have to maintain the accumulator yourself.
//
// This function will error if:
// * startTime > endTime
// * endTime in the future
// * startTime older than 48 hours OR pool creation
// * pool with id poolId does not exist, or does not contain quoteAssetDenom, baseAssetDenom
//
// N.B. If there is a notable use case, the state machine could maintain more historical records, e.g. at one per hour.
func (k Keeper) GetArithmeticTwap(ctx sdk.Context,
	poolId uint64,
	baseAssetDenom string, quoteAssetDenom string,
	startTime time.Time, endTime time.Time) (sdk.Dec, error) {
```

There are convenience methods for `GetArithmeticTwapToNow` which sets `endTime = ctx.BlockTime()`, and has minor gas reduction.
For users who need TWAPs outside the 48 hours stored in the state machine, you can get the latest accumulation store record from `GetBeginBlockAccumulatorRecord`

## Code layout

**api.go** is the main file you should look at as a user of this module.

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
    - Not currently in scope for release blocking, but planned: Integration for gas tracking, to ensure gas of reads/writes does not grow with time.
- [ ] Mutation testing usage
    - integration of the TWAP module into go mutation testing: https://github.com/osmosis-labs/go-mutesting
        - The success we've seen with the tokenfactory module, is it succeeds at surfacing behavior for untested behavior.
          e.g. if you delete a line, or change the direction of a conditional, does a test catch it.
    - We expect to get this to a state, where after mutation testing is ran, the only items it mutates, that is not caught in a test, is: Deleting `return err`, or `panic` lines, in the situation where that error return or panic isn't reachable.