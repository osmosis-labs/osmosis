# TWAP (Time Weighted Average Price)

The TWAP package is responsible for being able to serve TWAPs for every AMM pool.

A time weighted average price is a function that takes a sequence of `(time, price)` pairs, and returns a price representing an 'average' over the entire time period. The method of averaging can vary from the classic arithmetic mean, (such as geometric mean, harmonic mean), however we currently only implement arithmetic mean.

## Arithmetic mean TWAP

Using the arithmetic mean, the TWAP of a sequence `(t_i, p_i)`, from `t_0` to `t_n`, indexed by time in ascending order, is: $$\frac{1}{t_n - t_0}\sum_{i=0}^{n-1} p_i (t_{i+1} - t_i)$$
Notice that the latest price `p_n` isn't used, as it has lasted for a time interval of `0` seconds in this range!

To illustrate with an example, given the sequence: `(0s, $1), (4s, $6), (5s, $1)`, the arithmetic mean TWAP is: 
$$\frac{\$1 * (4s - 0s) + \$6 * (5s - 4s)}{5s - 0s} = \frac{\$10}{5} = \$2$$

## Computation via accumulators method

The prior example for how to compute the TWAP takes linear time in the number of time entries in a range, which is too inefficient. We require TWAP operations to have constant time complexity (in the number of records).

This is achieved by using an accumulator. In the case of an arithmetic TWAP, we can maintain an accumulator from `a_n`, representing the numerator of the TWAP expression for the interval `t_0...t_n`, namely 
$$a_n = \sum_{i=0}^{n-1} p_i (t_{i+1} - t_i)$$
If we maintain such an accumulator for every pool, with `t_0 = pool_creation_time` to `t_n = current_block_time`, we can easily compute the TWAP for any interval. The TWAP for the time interval of price points `t_i` to `t_j` is then $twap = \frac{a_j - a_i}{t_j - t_i}$, which is constant time given the accumulator values.

In Osmosis, we maintain accumulator records for every pool, for the last 48 hours.
We also maintain within each accumulator record in state, the latest spot price.
This allows us to interpolate accumulation records between times.
Namely, if I want the twap from `t=10s` to `t=15s`, but the time records are at `9s, 13s, 17s`, this is fine.
Using the latest spot price in each record, we create the accumulator value for `t=10` by computing
`a_10 = a_9 + a_9_latest_spot_price * (10s - 9s)`, and `a_15 = a_13 + a_13_latest_spot_price * (15s - 13s)`. 
Given these interpolated accumulation values, we can compute the TWAP as before.


## Module API

The primary intended API is `GetArithmeticTwap`, which is documented below, and has a similar cosmwasm binding.

```go
// GetArithmeticTwap returns an arithmetic time weighted average price.
// The returned twap is the time weighted average price (TWAP), using the arithmetic mean, of:
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
	startTime time.Time, endTime time.Time) (sdk.Dec, error) { ... }
```

There are convenience methods for `GetArithmeticTwapToNow` which sets `endTime = ctx.BlockTime()`, and has minor gas reduction.
For users who need TWAPs outside the 48 hours stored in the state machine, you can get the latest accumulation store record from `GetBeginBlockAccumulatorRecord`.

## Code layout

**api.go** is the main file you should look at as a user of this module.

**logic.go** is the main file you should look at for how the TWAP implementation works.

- client/* - Implementation of GRPC and CLI queries
- types/* - Implement TwapRecord, GenesisState. Define AMM interface, and methods to format keys.
- module/module.go - SDK AppModule interface implementation.
- api.go - Public API, that other users / modules can/should depend on
- hook_listener.go - Defines hooks & calls to logic.go, for triggering actions on 
- keeper.go - generic SDK boilerplate (defining a wrapper for store keys + params)
- logic.go - Implements all TWAP module 'logic'. (Arithmetic, defining what to get/set where, etc.)
- store.go - Managing logic for getting and setting things to underlying stores

## Store layout

We maintain TWAP accumulation records for every AMM pool on Osmosis. 

Because Osmosis supports multi-asset pools, a complicating factor is that we have to store a record for every asset pair in the pool.
For every pool, at a given point in time, we make one twap record entry per unique pair of denoms in the pool. If a pool has `k` denoms, the number of unique pairs is `k * (k - 1) / 2`.
All public API's for the module will sort the input denoms to the canonical representation, so the caller does not need to worry about this. (The canonical representation is the denoms in lexicographical order)

Each twap record stores [(source)](https://github.com/osmosis-labs/osmosis/tree/main/proto/osmosis/gamm/twap):
* last spot price of base asset A in terms of quote asset B
* last spot price of base asset B in terms of quote asset A
* Accumulation value of base asset A in terms of quote asset B
* Accumulation value of base asset B in terms of quote asset A

All TWAP records are indexed in state by the time of write.

A new TWAP record is created in two situations:
* When a pool is created
* In the `EndBlock`, if the block contains any potentially price changing event for the pool. (Swap, LP, Exit)

When a pool is created, records are created with the current spot price of the pool.

During `EndBlock`, new records are created, with:
* The accumulator's value is updated based upon the most recent prior accumulator's stored last spot price
* The `LastSpotPrice` value is equal to the EndBlock spot price.

In the event that a pool is created, and has a swap in the same block, the record entries are over written with the end block price.

### Tracking spot-price changing events in a block

The flow by which we currently track spot price changing events in a block is as follows:
* AMM hook triggers for Swapping, LPing or Exiting a pool
* TWAP listens for this hook, and adds this pool ID to a local tracker
* In end block, TWAP iterates over every changed pool in that block, based on the local tracker, and updates their TWAP records
* In end block, TWAP clears the changed pool list, so it is blank by the next block.

The mechanism by which we maintain this changed pool list, is the SDK `Transient Store`.
The transient store is a KV store in the SDK, that stores entries in memory, for the duration of a block,
and then clears on the block committing. This is done to save on gas (and I/O for the state machine).

## Pruning

To avoid infinite growth of the state with the TWAP records, we attempt to delete some old records after every epoch.
Essentially, records younger than a configurable parameter are pruned away. Currently, this parameter is set to 48 hours.
Therefore, at the end of an epoch records younger than 48 hours before the current block time are pruned away.

## Testing Methodology

The pre-release testing methodology planned for the twap module is:

- [ ] Using table driven unit tests to test all foreseen states of the module
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
- [ ] End to end migration tests
  - Tests that migration of Osmosis pools created prior to the TWAP upgrade, get TWAPs recorded starting at the v11 upgrade.
- [ ] Integration into the Osmosis simulator
  - The osmosis simulator, simulates building up complex state machine states, in random ways not seen before. We plan on, in a property check, maintaining expected TWAPs for short time ranges, and seeing that the keeper query will return the same value as what we get off of the raw price history for short history intervals.
  - Not currently deemed release blocking, but planned: Integration for gas tracking, to ensure gas of reads/writes does not grow with time.
- [ ] Mutation testing usage
  - integration of the TWAP module into [go mutation testing](https://github.com/osmosis-labs/go-mutesting): 
    - We've seen with the `tokenfactory` module that it succeeds at surfacing behavior for untested logic.
        e.g. if you delete a line, or change the direction of a conditional, mutation tests show if regular Go tests catch it.
    - We expect to get this to a state, where after mutation testing is ran, the only items it mutates, that is not caught in a test, is: Deleting `return err`, or `panic` lines, in the situation where that error return or panic isn't reachable.