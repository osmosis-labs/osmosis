# Pool Manager Module

The poolmanager module exists as a swap entrypoint for any pool model
that exists on the chain. The poolmanager module is responsible for routing
swaps across various pools. It also performs pool-id management for
any on-chain pool.

The user-stories for this module follow:

> As a user, I would like to have a unified entrypoint for my swaps regardless
of the underlying pool implementation so that I don't need to reason about
API complexity

> As a user, I would like the pool management to be unified so that I don't
have to reason about additional complexity stemming from divergent pool sources.

We have multiple pool-storage modules. Namely, `x/gamm` and `x/concentrated-liquidity`.

To avoid fragmenting swap and pool creation entrypoints and duplicating their boilerplate logic,
we define a `poolmanager` module. Its purpose is twofold:
1. Handle pool creation
   * Assign ids to pools
   * Store the mapping from pool id to one of the swap modules (`gamm` or `concentrated-liquidity`)
   * Propagate the execution to the appropriate module depending on the pool type.
   * Note, that pool creation messages are received by the pool model's message server.
   Each module's message server then calls the `x/poolmanager` keeper method `CreatePool`.
2. Handle swaps
   * Cover & share multihop logic
   * Propagate intra-pool swaps to the appropriate module depending on the pool type.
   * Contrary to pool creation, swap messages are received by the `x/poolmanager` message server.

Let's consider pool creation and swaps separately and in more detail.

## Pool Creation & Id Management

To make sure that the pool ids are unique across the two modules, we unify pool id management
in the `poolmanager`.

When a call to `CreatePool` keeper method is received, we get the next pool id from the module
storage, assign it to the new pool, and propagate the execution to either `gamm`
or `concentrated-liquidity` modules.

Note that we define a `CreatePoolMsg` interface:
<https://github.com/osmosis-labs/osmosis/blob/f26ceb958adaaf31510e17ed88f5eab47e2bac03/x/poolmanager/types/msg_create_pool.go#L9>

Each `balancer`, `stableswap` and `concentrated-liquidity` pool has its own implementation of `CreatePoolMsg`.

Note the `PoolType` type. This is an enumeration of all supported pool types.
We proto-generate this enumeration:

```go
// proto/osmosis/poolmanager/v1beta1/module_route.proto
// generates to x/poolmanager/types/module_route.pb.go

// PoolType is an enumeration of all supported pool types.
enum PoolType {
  option (gogoproto.goproto_enum_prefix) = false;

  // Balancer is the standard xy=k curve. Its pool model is defined in x/gamm.
  Balancer = 0;
  // Stableswap is the Solidly cfmm stable swap curve. Its pool model is defined
  // in x/gamm.
  StableSwap = 1;
  // Concentrated is the pool model specific to concentrated liquidity. It is
  // defined in x/concentrated-liquidity.
  Concentrated = 2;
}
```

Let's begin by considering the execution flow of the pool creation message.
Assume `balancer` pool is being created.

1. `CreatePoolMsg` is received by the `x/gamm` message server.

2. `CreatePool` keeper method is called from `poolmanager`, propagating
the appropriate implementation of the `CreatePoolMsg` interface.

```go
// x/poolmanager/creator.go CreatePool(...)

// CreatePool attempts to create a pool returning the newly created pool ID or
// an error upon failure. The pool creation fee is used to fund the community
// pool. It will create a dedicated module account for the pool and sends the
// initial liquidity to the created module account.
//
// After the initial liquidity is sent to the pool's account, this function calls an
// InitializePool function from the source module. That module is responsible for:
// - saving the pool into its own state
// - Minting LP shares to pool creator
// - Setting metadata for the shares
func (k Keeper) CreatePool(ctx sdk.Context, msg types.CreatePoolMsg) (uint64, error) {
    ...
}
```

3. The keeper utilizes `CreatePoolMsg` interface methods to execute the logic specific
to each pool type.

4. Lastly, `poolmanager.CreatePool` routes the execution to the appropriate module.

The propagation to the desired module is ensured by the routing table stored in memory in the `poolmanager` keeper.

```go
// x/poolmanager/keeper.go NewKeeper(...)

func NewKeeper(...) *Keeper {
    ...

	routes := map[types.PoolType]types.SwapI{
		types.Balancer:     gammKeeper,
		types.Stableswap:   gammKeeper,
		types.Concentrated: concentratedKeeper,
	}

	return &Keeper{..., routes: routes}
}
```

`MsgCreatePool` interface defines the following method: `GetPoolType() PoolType`

As a result, `poolmanagerkeeper.CreatePool` can route the execution to the appropriate module in
the following way:

```go
// x/poolmanager/creator.go CreatePool(...)

swapModule := k.routes[msg.GetPoolType()]

if err := swapModule.InitializePool(ctx, pool, sender); err != nil {
    return 0, err
}
```

Where swapModule is either `gamm` or `concentrated-liquidity` keeper.

Both of these modules implement the `SwapI` interface:

```go
// x/poolmanager/types/routes.go SwapI interface

type SwapI interface {
    ...

	InitializePool(ctx sdk.Context, pool gammtypes.PoolI, creatorAddress sdk.AccAddress) error
}
```

As a result, the `poolmanager` module propagates core execution to the appropriate swap module.

Lastly, the `poolmanager` keeper stores a mapping from the pool id to the pool type.
This mapping is going to be necessary for knowing where to route the swap messages.

To achieve this, we create the following store index:

```go
// x/poolmanager/types/keys.go

var	(
    ...

    SwapModuleRouterPrefix     = []byte{0x02}
)

// N.B.: we proto-generate this struct. However, the proto
// definition is omitted for brevity.
type ModuleRoute struct {
    PoolType PoolType
}

// FormatModuleRouteKey serializes pool id with appropriate prefix into bytes.
func FormatModuleRouteKey(poolId uint64) []byte {
	return []byte(fmt.Sprintf("%s%d", SwapModuleRouterPrefix, poolId))
}

// ParseModuleRouteFromBz parses the raw bytes into ModuleRoute.
// Returns error if fails to parse or if the bytes are empty.
func ParseModuleRouteFromBz(bz []byte) (ModuleRoute, error) {
    // parsing logic
}
```

## Swaps

There are 4 swap messages:

- `MsgSwapExactAmountIn`
- `MsgSwapExactAmountOut`
- `MsgSplitRouteSwapExactAmountIn`
- `MsgSplitRouteSwapExactAmountOut`

Between, `MsgSwapExactAmountIn` and `MsgSwapExactAmountOut`, the implementation of routing is similar. We only focus on `MsgSwapExactAmountIn` below.

`MsgSplitRouteSwapExactAmountIn` and `MsgSplitRouteSwapExactAmountOut` support split routes where for each split route they call the respective
`MsgSwapExactAmountIn` or `MsgSwapExactAmountOut` message. When using the split routes, the slippage protection is disabled on the per-route basis.
For swap exact amount in, we provide zero for the min amount out. For swap exact amount out, we provide the max amount in which is 1 << 256 - 1.
Read more about route splitting in the "Route Splitting" section.

Once the message is received, it calls `RouteExactAmountIn`

```go
// x/poolmanager/router.go RouteExactAmountIn(...)

// RouteExactAmountIn defines the input denom and input amount for the first pool,
// the output of the first pool is chained as the input for the next routed pool
// transaction succeeds when final amount out is greater than tokenOutMinAmount defined.
func (k Keeper) RouteExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	routes []types.SwapAmountInRoute,
	tokenIn sdk.Coin,
	tokenOutMinAmount osmomath.Int) (tokenOutAmount osmomath.Int, err error) {
}
```

Essentially, the method iterates over the routes and calls a `SwapExactAmountIn` method
for each, subsequently updating the inter-pool swap state.

The routing works by querying the index `SwapModuleRouterPrefix`,
searching up the `poolmanagerkeeper.router` mapping, and calling
`SwapExactAmountIn` method of the appropriate module.

```go
// x/poolmanager/router.go RouteExactAmountIn(...)

moduleRouteBytes := osmoutils.MustGet(poolmanagertypes.FormatModuleRouteIndex(poolId))
moduleRoute, _ := poolmanagertypes.ModuleRouteFromBytes(moduleRouteBytes)

swapModule := k.routes[moduleRoute.PoolType]

_ := swapModule.SwapExactAmountIn(...)
```
- note that error checks and other details are omitted for brevity.

Similar to pool creation logic, we are able to call `SwapExactAmountIn` on any of the swap
modules by implementing the `SwapI` interface:

```go
// x/poolmanager/types/routes.go SwapI interface

type SwapI interface {
    ...

	SwapExactAmountIn(
		ctx sdk.Context,
		sender sdk.AccAddress,
		poolId gammtypes.PoolI,
		tokenIn sdk.Coin,
		tokenOutDenom string,
		tokenOutMinAmount osmomath.Int,
		spreadFactor osmomath.Dec,
	) (osmomath.Int, error)
}
```

During the process of swapping a specific asset, the token the user is
putting into the pool is denoted as `tokenIn`, while the token that
would be returned to the user, the asset that is being swapped for,
after the swap is denoted as `tokenOut` throughout the module.

For example, in the context of balancer pools, given a `tokenIn`, the
following calculations are done to calculate how many tokens are to be
swapped into and removed from the pool:

`tokenBalanceOut * [1 - { tokenBalanceIn / (tokenBalanceIn + (1 - spreadFactor) * tokenAmountIn)} ^ (tokenWeightIn / tokenWeightOut)]`

The calculation is also able to be reversed, the case where user
provides `tokenOut`. The calculation for the amount of tokens that the
user should be putting in is done through the following formula:

`tokenBalanceIn * [{tokenBalanceOut / (tokenBalanceOut - tokenAmountOut)} ^ (tokenWeightOut / tokenWeightIn) -1] / tokenAmountIn`

With the introduction of a `takerFee`, the actual amount of `tokenIn` that is used to calculate the amount of `tokenOut` is reduced by the `takerFee` amount. If governance or a governance approved DAO adds a specified trading pair to the `takerFee` module store, the fee associated with that pair is used. Otherwise, the `defaultTakerFee` defined in the poolmanger's parameters is used.

The poolmanager only concerns itself with proportionally distributing the takerFee to the respective staking rewards and community pool txfees module accounts. For swaps originating in OSMO, the poolmanger distributes these fees based on the `OsmoTakerFeeDistribution` parameter. For swaps originating in non-OSMO assets, the poolmanager distributes these fees based on the `NonOsmoTakerFeeDistribution` parameter. For taker fees generated in non whitelisted quote denoms assets, the amount that goes to the community pool (defined by the `NonOsmoTakerFeeDistribution` above) is swapped to the `community_pool_denom_to_swap_non_whitelisted_assets_to` parameter defined in poolmanager. For instance, if a taker fee is generated in BTC, the respective community pool percent is sent directly to the community pool since it is a whitelisted quote denom. If it is generated in FOO, which is not a whitelisted quote denom, the respective community pool percent is swapped to the `community_pool_denom_to_swap_non_whitelisted_assets_to` parameter defined in poolmanager and send to the community pool as that denom at epoch.

For more information on how the final distribution of these fees and how they are swapped, see the txfees module README.

Existing Swap types:
- SwapExactAmountIn
- SwapExactAmountOut

## Messages

### MsgSwapExactAmountIn

[MsgSwapExactAmountIn](https://github.com/osmosis-labs/osmosis/blob/f26ceb958adaaf31510e17ed88f5eab47e2bac03/proto/osmosis/gamm/v1beta1/tx.proto#L79)

### MsgSwapExactAmountOut

[MsgSwapExactAmountOut](https://github.com/osmosis-labs/osmosis/blob/f26ceb958adaaf31510e17ed88f5eab47e2bac03/proto/osmosis/gamm/v1beta1/tx.proto#L102)

### MsgSplitRouteSwapExactAmountIn

[MsgSplitRouteSwapExactAmountIn](https://github.com/osmosis-labs/osmosis/blob/46e6a0c2051a3a5ef8cdd4ecebfff7305b13ab98/proto/osmosis/poolmanager/v1beta1/tx.proto#L41)

## MsgSplitRouteSwapExactAmountOut

[MsgSplitRouteSwapExactAmountOut](https://github.com/osmosis-labs/osmosis/blob/46e6a0c2051a3a5ef8cdd4ecebfff7305b13ab98/proto/osmosis/poolmanager/v1beta1/tx.proto#L85)

## MsgSetDenomPairTakerFee

[MsgSplitRouteSwapExactAmountOut](https://github.com/osmosis-labs/osmosis/blob/d129ea37f5490d8a212932a78cd35cb864c799c7/proto/osmosis/poolmanager/v1beta1/tx.proto#L121)

## Multi-Hop

All tokens are swapped using a multi-hop mechanism. That is, all swaps
are routed via the most cost-efficient way, swapping in and out from
multiple pools in the process.
The most cost-efficient route is determined offline and the list of the pools is provided externally, by user, during the broadcasting of the swapping transaction.
At the moment of execution, the provided route may not be the most cost-efficient one anymore.

[Multi-Hop](https://github.com/osmosis-labs/osmosis/blob/f26ceb958adaaf31510e17ed88f5eab47e2bac03/x/poolmanager/router.go#L16)

## Route Splitting

Each route can be thought of as a separate multi-hop swap.

Splitting swaps across multiple pools for the same token pair can be beneficial for several reasons,
primarily relating to reduced slippage, price impact, and potentially lower spreads.

Here's a detailed explanation of these advantages:

- **Reduced slippage**: When a large trade is executed in a single pool, it can be significantly affected if someone else executes a large swap against that pool.

- **Lower price impact**: When executing a large trade in a single pool, the price impact can be substantial, leading to a less favorable exchange rate for the trader.
By splitting the swap across multiple pools, the price impact in each pool is minimized, resulting in a better overall exchange rate.

- **Improved liquidity utilization**: Different pools may have varying levels of liquidity, spreads, and price curves. By splitting swaps across multiple pools,
the router can utilize liquidity from various sources, allowing for more efficient execution of trades. This is particularly useful when the liquidity in
a single pool is not sufficient to handle a large trade or when the price curve of one pool becomes less favorable as the trade size increases.

- **Potentially lower spreads**: In some cases, splitting swaps across multiple pools may result in lower overall spreads. This can happen when different pools
have different spread structures, or when the total spread paid across multiple pools is lower than the spread for executing the entire trade in a single pool with
higher slippage.

Note, that the actual split happens off-chain. The router is only responsible for executing the swaps in the order and quantities of token in provided
by the routes.

## EstimateTradeBasedOnPriceImpact Query

The `EstimateTradeBasedOnPriceImpact` query allows users to estimate a trade for all pool types given the following parameters are provided for this request `EstimateTradeBasedOnPriceImpactRequest`:

- **FromCoin**: (`sdk.Coin`): is the total amount of tokens one wants to sell.
- **ToCoinDenom**: (`string`): is the denom they want to buy with the tokens being sold.
- **PoolId**: (`uint64`): is the identifier of the pool that the trade will happen on.
- **MaxPriceImpact**: (`sdk.Dec`): is the maximum percentage that the user is willing to affect the price of the pool.
- **ExternalPrice**: (`sdk.Dec`) is an external price that the user can optionally enter to have the `MaxPriceImpact` adjusted as the `SpotPrice` of a pool could be changed at any time.

The response would be `EstimateTradeBasedOnPriceImpactResponse` which contains the following data:

- **InputCoin**: (`sdk.Coin`): the actual input amount that would be tradeable under that price impact (might be the full amount).
- **OutputCoin**: (`sdk.Coin`): the amount of the `ToCoinDenom` tokens being received for the actual `InputCoin` trade.

With that data it is easier for any entity to fill in the `MsgSwapExactAmountIn` details. The response could be filled with a valid trade or an empty one(InputCoin = 0, OutputCoin = 0), an empty one indicates that no trade could be estimated.
It will not error if a trade cannot be estimated.

### Process

The following is the process in which the query finds a trade that will stay below the `MaxPriceImpact` value.

1. Verify `PoolId`, `FromCoin`, `ToCoinDenom` are not empty.
2. Return the specific `swapModule` based on the `PoolId`.
3. Return the specific `PoolI` interface from the `swapModule` based on the `PoolId`.
4. Calculate the `SpotPrice` in terms of the token being bought, therefore if it's an `OSMO/ATOM` pool and `OSMO` is being sold we need to calculate the `SpotPrice` in terms of `ATOM` being the base asset and `OSMO` being the quote asset.
5. If we have a `ExternalPrice` specified in the request we need to adjust the `MaxPriceImpact` into a new variable `adjustedMaxPriceImpact` which would either increase if the `SpotPrice` is cheaper than the `ExternalPrice` or decrease if the `SpotPrice` is more expensive leaving less room to estimate a trade.
   1. If the `adjustedMaxPriceImpact` was calculated to be `0` or negative it means that the `SpotPrice` is more expensive than the `ExternalPrice` and has already exceeded the possible `MaxPriceImpact`. We return a `osmomath.ZeroInt()` input and output for the input and output coins indicating that no trade is viable.
6. Then according to the pool type we attempt to find a viable trade, we must process each pool type differently as they return different results for different scenarios. The sections below explain the different pool types and how they each handle input.

#### Balancer Pool Type Process

The following is the example input/output when executing `CalcOutAmtGivenIn` on balancer pools:

- If the input is greater than the total liquidity of the pool, the output will be the total liquidity of the target token.
- If the input is an amount that is reasonably within the range of liquidity of the pool, the output will be a tolerable slippage amount based on pool data.
- If the input is a small amount for which the pool cannot calculate a viable swap output e.g `1`, the output will be a small value which can be either positive (greater or equal to 1) or zero, depending on the pool's weights. In the latter case an `ErrInvalidMathApprox` is returned.

Here is the following process for the `EstimateTradeBasedOnPriceImpactBalancerPool` function:

1. The function initially calculates the output amount (`tokenOut`) using the input amount (`FromCoin`) without including a swap fee using the `CalcOutAmtGivenIn` function.

   1. If `tokenOut` is zero or an `ErrInvalidMathApprox` is returned, the function returns zero for both the input and output coin, signifying that trading a negligible amount yields no output.

2. The function calculates the current trade price (`currTradePrice`) using the initially estimated `tokenOut`. Following that, it calculates the deviation of this price from the spot price (`priceDeviation`).

   1. If the `priceDeviation` is within the acceptable range (`adjustedMaxPriceImpact`), the function recalculates `tokenOut` but this time includes the swap fee. The estimated trade is then returned.

3. In case the initial `priceDeviation` was not within the acceptable range, the function starts a binary search loop. It initializes `lowAmount`, `highAmount`, and `currFromCoin` to perform this search.

4. Within the binary search loop, the function recalculates the middle amount (`midAmount`) to try estimate `CalcOutAmtGivenIn` again. It performs new trade estimations until it either finds an acceptable `priceDeviation` or exhausts the search range.

5. If the loop exhausts the search range without finding a viable trade, it returns zero for both the input and output coin.

6. If a viable trade is found that respects the `adjustedMaxPriceImpact`, the function performs a final recalculation, this time including the swap fee, and returns the estimated trade.

#### StableSwap Pool Type Process

The following is the example input/output when executing `CalcOutAmtGivenIn` on stableswap pools:

- If the input is greater than the total liquidity of the pool, the function will `panic`.
- If the input is an amount that is reasonably within the range of liquidity of the pool, the output will be a tolerable slippage amount based on pool data.
- If the input is a small amount for which the pool cannot calculate a viable swap output e.g `1`, the function will throw an error.

Here is the following process for the `EstimateTradeBasedOnPriceImpactStableSwapPool` function:

1. The function begins by attempting to estimate the output amount (`tokenOut`) for a given input amount (`req.FromCoin`). This calculation is done without accounting for the swap fee.

   1. If an error occurs, and it's not a panic, the function returns zero coins for both the input and output, signifying an error due to an amount that's too small for the trade to proceed.

   2. If a panic occurs during the calculation, the function sets the output coin (`tokenOut`) to zero and proceeds to find a smaller acceptable trade amount.

2. When there is no error or panic, the function calculates the current trade price (`currTradePrice`) and checks if the price deviation (`priceDeviation`) from the spot price is within acceptable limits (`adjustedMaxPriceImpact`).

   1. If the `priceDeviation` is acceptable, the function re-estimates the output amount (`tokenOut`) considering the swap fee. If successful, this trade estimate is returned.

3. The function initializes variables `lowAmount` and `highAmount` to search for an acceptable trade amount if the initial amount is too large or too small.

4. Within a loop, the function performs a binary search to find an acceptable trade amount. It attempts a new trade with the middle amount (`midAmount`) between `lowAmount` and `highAmount`.

5. If the new trade amount leads to an error without a panic, the function returns zero coins, indicating the amount has become too small.

6. If the new trade amount leads to a panic, the function adjusts the `highAmount` downwards to continue the search.

7. If the new trade amount does not cause an error or panic, and its `priceDeviation` is within limits, the function adjusts the `lowAmount` upwards to continue the search.

8. If the loop completes without finding an acceptable trade amount, the function returns zero coins for both the input and the output.

9. If a viable trade is found, the function performs a final recalculation considering the swap fee and returns the estimated trade.

#### Concentrated Liquidity Pool Type Process

The following is the example input/output when executing `CalcOutAmtGivenIn` on concentrated liquidity pools:

- If the input is greater than the total liquidity of the pool, the function will error.
- If the input is an amount that is reasonably within the range of liquidity of the pool, the output will be a tolerable slippage amount based on pool data.
- f the input is a small amount for which the pool cannot calculate a viable swap output e.g `1`, the function will return a zero.

Here is the following process for the `EstimateTradeBasedOnPriceImpactConcentratedLiquidity` function:

1. The function starts by attempting to estimate the output amount (`tokenOut`) for a given input amount (`req.FromCoin`), using the `CalcOutAmtGivenIn` method of the `swapModule`.

   1. If `tokenOut` is zero, it means the amount being traded is too small. The function returns zero coins for both input and output.

   2. If an error occurs but `tokenOut` is not zero, the function ignores the error. The function assumes the error could mean the input is too large and proceeds to the next steps to find a suitable trade amount.

2. If there is no error in estimating `tokenOut`, the function calculates the current trade price (`currTradePrice`). It then checks if the price deviation (`priceDeviation`) from the spot price is within acceptable limits (`adjustedMaxPriceImpact`).

   1. If the `priceDeviation` is acceptable, the function re-estimates the `tokenOut` considering the swap fee. If successful, this trade estimate is returned.

3. The function initializes `lowAmount` and `highAmount` variables to search for an acceptable trade amount if the initial amount is unsuitable.

4. Within a loop, the function performs a binary search for an acceptable trade amount. It calculates a middle amount (`midAmount`) between `lowAmount` and `highAmount` to attempt a new trade.

5. If the new `tokenOut` is zero, the function returns zero coins for both input and output, indicating the trade amount is too small.

6. If no error occurs with the new trade amount and the `priceDeviation` is within limits, the function adjusts `lowAmount` upwards.

7. If an error occurs with the new trade amount, the function adjusts `highAmount` downwards to continue the search.

8. If the loop completes without finding an acceptable trade amount, the function returns zero coins for both input and output.

9. If a viable trade amount is found, the function performs a final estimation of `tokenOut` considering the swap fee and returns the estimated trade.

## Taker Fees

Taker fee distribution is defined in the poolmanager module’s param store:

```proto
type TakerFeeParams struct {
    DefaultTakerFee cosmossdk_io_math.LegacyDec `protobuf:"bytes,1,opt,name=default_taker_fee,json=defaultTakerFee,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"default_taker_fee"`

    OsmoTakerFeeDistribution TakerFeeDistributionPercentage `protobuf:"bytes,2,opt,name=osmo_taker_fee_distribution,json=osmoTakerFeeDistribution,proto3" json:"osmo_taker_fee_distribution"`

    NonOsmoTakerFeeDistribution TakerFeeDistributionPercentage `protobuf:"bytes,3,opt,name=non_osmo_taker_fee_distribution,json=nonOsmoTakerFeeDistribution,proto3" json:"non_osmo_taker_fee_distribution"`

    CommunityPoolDenomToSwapNonWhitelistedAssetsTo string `protobuf:"bytes,5,opt,name=community_pool_denom_to_swap_non_whitelisted_assets_to,json=communityPoolDenomToSwapNonWhitelistedAssetsTo,proto3" json:"community_pool_denom_to_swap_non_whitelisted_assets_to,omitempty" yaml:"community_pool_denom_to_swap_non_whitelisted_assets_to"`
}
```

Not shown here is a separate KVStore, which holds overrides for the defaultTakerFee.

The Osmosis protocol now supports setting up taker fee agreements with specific denoms to share a certain percentage of taker fees generated in any route containing those denoms:

```go
func (k *Keeper) SetTakerFeeShareAgreementForDenom(ctx sdk.Context, takerFeeShare types.TakerFeeShareAgreement) error

type TakerFeeShareAgreement struct {
    Denom       string `json:"denom"`
    SkimAddress string `json:"skim_address"`
    SkimPercent string `json:"skim_percent"`
}
```

For example, if the agreement specifies a 10% `skim_percent`, then 10% of all taker fees generated in a swap route containing the specified denom will be sent to the `skim_address` at the end of each epoch. These percentages are additive, so if there are agreements with skim percents of 10%, 20%, and 30%, the total skim percent for the route will be 60%. If the skim percent exceeds 100%, the swap fails to go through.

The protocol can also register alloyed asset pools, which are pools containing denoms with taker fee share agreements:

```go
func (k *Keeper) SetRegisteredAlloyedPool(ctx sdk.Context, poolId uint64) error
```

If a swap route contains an alloyed asset pool but no individual taker fee share agreement denoms, the taker fees will be skimmed based on the underlying denoms in the pool that have taker fee share agreements, adjusted by their respective weights. For example, if an alloyed asset pool contains 50% nBTC (10% skim percent) and 50% iBTC (5% skim percent), the total skim percent for the pool will be 7.5%, with 5% of the taker fees going to the nBTC skim address and 2.5% to the iBTC skim address. Just like the taker fee share agreements, if the skim percent exceeds 100%, the swap fails to go through.

NOTE: We cache the alloyed pool asset weight composition and recalculate in the poolmanager EndBlock every 700 blocks (approx 30 minutes at 2.6s blocks) to avoid recalculating the weights for every swap.

At the time of swap, all taker fees (including those skimmed via taker fee share agreements) are sent to the `taker_fee_collector` module account. At the time of sending to the `taker_fee_collector`, we track the amount of fees to be skimmed via the `KeyTakerFeeShareDenomAccrualForTakerFeeChargedDenom` key. At the end of each epoch, all the fees tracked are distributed as follows:

- Skimmed taker fee:
    - Sent directly to the `skim_address` specified in the taker fee agreement.
- Non native taker fees
    - For Community Pool: Sent to `non_native_fee_collector_community_pool` module account, swapped to `CommunityPoolDenomToSwapNonWhitelistedAssetsTo`, then sent to community pool
    - For Stakers: Sent to `non_native_fee_collector_stakers` module account, swapped to OSMO, then sent to auth module account, which distributes it to stakers
    - The sub-module accounts here are used so that, if a swap fails, the tokens that fail to swap are not grouped back into the wrong taker fee category in the next epoch
- OSMO taker fees
    - For Community Pool: Sent directly to community pool
    - For Stakers: Sent directly to auth module account, which distributes it to stakers

Lets go through the lifecycle to better understand how taker fee works in a variety of situations, and how the module account and distribution parameters are used depending on the input token.

### Example 1: Non OSMO taker fee

A user makes a swap of USDC to OSMO. First, the protocol checks the KVStore to determine if the denom pair has a taker fee override. If the pair exists in the KVStore, the taker fee override is used. If the pair does not exist, the defaultTakerFee is used. It is important to note that the order of the denom pair now matters, so if a denom pair taker fee override exists for OSMO to USDC but not USDC to OSMO, the default taker fee will instead be used.

In this example, defaultTakerFee is 0.02%. A USDC->OSMO KVStore exists with an override of 0.01%. Therefore, 0.01% is used.

Now, imagine the amount in is 10000 USDC. This means that the amount of takerFee utilized is 0.01% of 10000, which is 1 USDC.

In the takerFee params, there are two distribution categories:
1. Taker fees generated in OSMO
2. Taker fees generated in non-OSMO

Since USDC is non-OSMO, we look at category 2. In both categories, the fees are distributed to a combo of staking rewards and community pool:

```proto
type TakerFeeDistributionPercentage struct {
    StakingRewards cosmossdk_io_math.LegacyDec `protobuf:"bytes,1,opt,name=staking_rewards,json=stakingRewards,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"staking_rewards" yaml:"staking_rewards"`
    CommunityPool  cosmossdk_io_math.LegacyDec `protobuf:"bytes,2,opt,name=community_pool,json=communityPool,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"community_pool" yaml:"community_pool"`
}
```

For simplicity sake, let’s say staking rewards are 40% and community pool is 60%. This means that out of the 1 USDC taken, 0.4 USDC is meant for staking rewards and 0.6 USDC is meant for community pool.

At time of swap, all 1 USDC is sent to the `taker_fee_collector` module account. Nothing is done with any taker fee funds until epoch.

Starting with the community pool funds, at epoch, the protocol checks if the token is a whitelisted fee token. If it is, it is sent directly to the community pool. If it is not, the funds are sent to the `non_native_fee_collector_community_pool`, swapped to the `CommunityPoolDenomToSwapNonWhitelistedAssetsTo` defined in the `poolmanager` params above, and then sent all at once to the community pool after all swaps at epoch have taken place.

Next, for staking rewards, since this is a non-OSMO token, it is swapped to OSMO and sent to the auth module account, which distributes it to stakers.

### Example 2: OSMO taker fee

This example does not differ much from the previous example. In this example, a user is swapping 1000 OSMO for USDC.

We search for a KVStore taker fee override before utilizing the default taker fee. Just as before (order does not matter), a KVStore entry for OSMO<>USDC exists, so we utilize a 0.01% taker fee instead of the 0.02% default taker fee. 0.01% of 10000 OSMO is 1 OSMO.

At time of swap, all 1 OSMO is sent to the `taker_fee_collector` module account. Again, nothing is done with any taker fee funds until epoch.

At epoch, we check the `OsmoTakerFeeDistribution`. In this example, let’s say its 20% to community pool and 80% to stakers. This means that 0.2 OSMO is set for community pool and 0.8 is set for stakers.

For community pool, this is just a direct send of the OSMO to the community pool.

For staking, the OSMO is directly sent to the auth "fee collector" module account, which distributes it to stakers.

### Example 3: Multi hop swap through a taker fee share denom and alloyed asset pool

The swap route is as follows:

OSMO -> nBTC -> allBTC -> iBTC

The Osmosis protocol has a taker fee share agreement with nBTC and iBTC. The skim percent for nBTC is 10% and for iBTC is 5%.

This swap generates 5 OSMO, 10 nBTC, and 20 allBTC in taker fees.

Because this route is made up of two taker fee share denoms, the total skim percent for the route is 15% (10% from nBTC and 5% from iBTC). 10% of the taker fees are noted to go to the nBTC `skim_address` (0.5 OSMO, 1 nBTC, 2 allBTC) and 5% to the iBTC `skim_address` (0.25 OSMO, 0.5 iBTC, 1 allBTC). All funds, including those noted to go to the `skim_address`, are sent to the `taker_fee_collector` module account at time of swap. At epoch, the noted skim amounts are sent to the respective `skim_address`.

Notice, no logic touches the alloyed asset pool due to the swap route containing taker fee share denoms, which trumps the alloyed asset pool share (the next example will show how the alloyed asset pool is used when the route does not contain taker fee share denoms). The rate of update in blocks is determined by the `alloyedAssetCompositionUpdateRate` variable.

### Example 4: Multi hop swap through an alloyed asset pool

The swap route is as follows:

OSMO -> wBTC -> allBTC -> xBTC

The Osmosis protocol has a taker fee share agreement with nBTC and iBTC. The skim percent for nBTC is 10% and for iBTC is 5%.

allBTC is made up of four denoms:
- nBTC (10% skim percent)
- iBTC (5% skim percent)
- WBTC (no taker fee share agreement)
- xBTC (no taker fee share agreement)

allBTC's composition is 25% nBTC, 25% iBTC, 25% WBTC, and 25% xBTC.

This swap generates 5 OSMO, 10 wBTC, and 20 allBTC in taker fees.

Because this route does not contain taker fee share denoms, but does contain a registered alloyed pool, each skim percent must be calculated as the weighted skim percents of the underlying denoms with taker fee share agreements. The total skim percent for the route is 3.75%, 2.5% to nBTC (10% * 25%) and 1.25% to iBTC (5% * 25%). 2.5% of the taker fees are noted to go to the nBTC `skim_address` (0.125 OSMO, 0.25 wBTC, 0.5 allBTC) and 1.25% to the iBTC `skim_address` (0.0625 OSMO, 0.125 wBTC, 0.25 allBTC). All funds, including those noted to go to the `skim_address`, are sent to the `taker_fee_collector` module account at time of swap. At epoch, the noted skim amounts are sent to the respective `skim_address`.

### Cached Maps: `cachedTakerFeeShareAgreementMap` and `cachedRegisteredAlloyPoolByAlloyDenomMap`

#### Overview

The `cachedTakerFeeShareAgreementMap` and `cachedRegisteredAlloyPoolByAlloyDenomMap` are two in-memory caches used to optimize the retrieval of taker fee share agreements and registered alloyed pools, respectively. These caches are designed to reduce the number of reads from the persistent KVStore, thereby improving the performance of the poolmanager module.

The initialization of these caches occurs during the `BeginBlock` method to ensure they are populated at the start of the block if they are empty. This means that the cache population logic in `BeginBlock` is always executed when the node is first started up, and then is a no-op for the rest of the block processing.

By using these caches, the pool manager module can efficiently handle frequent read operations, thereby improving overall performance and reducing latency.

#### `cachedTakerFeeShareAgreementMap`

- **Purpose**: This map caches taker fee share agreements, which define the percentage of taker fees to be skimmed and sent to specific addresses for various denoms.
- **Type**: `map[string]types.TakerFeeShareAgreement`
- **Setter**: The cache is initialized by calling the `setTakerFeeShareAgreementsMapCached` method, which populates the map with all taker fee share agreements from the KVStore.
- **Usage**: The cache is used to quickly retrieve taker fee share agreements during swaps and other operations that involve taker fees.

#### `cachedRegisteredAlloyPoolByAlloyDenomMap`

- **Purpose**: This map caches the state of registered alloyed pools, which are pools containing denoms with taker fee share agreements.
- **Type**: `map[string]types.AlloyContractTakerFeeShareState`
- **Setter**: The cache is initialized by calling the `setAllRegisteredAlloyedPoolsByDenomCached` method, which populates the map with all registered alloyed pools from the KVStore.
- **Usage**: The cache is used to quickly retrieve the state of registered alloyed pools during swaps and other operations that involve alloyed assets.
