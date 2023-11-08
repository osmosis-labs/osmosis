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
   1. If the `adjustedMaxPriceImpact` was calculated to be `0` or negative it means that the `SpotPrice` is more expensive than the `ExternalPrice` and has already exceeded the possible `MaxPriceImpact`. We return a `sdk.ZeroInt()` input and output for the input and output coins indicating that no trade is viable.
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

## Take Fee

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

There are also two module accounts involved:

```proto
non_native_fee_collector: osmo1g7ajkk295vactngp74shkfrprvjrdwn662dg26
non_native_fee_collector_community_pool: osmo1f3xhl0gqmyhnu49c8k3j7fkdv75ug0xjtaqu09
```

Lets go through the lifecycle to better understand how taker fee works in a variety of situations, and how each of these parameters and module accounts are used.

### Example 1: Non OSMO taker fee

A user makes a swap of USDC to OSMO. First, the protocol checks the KVStore to determine if the the denom pair has a taker fee override. If the pair exists in the KVStore, the taker fee override is used. If the pair does not exist, the defaultTakerFee is used.

In this example, defaultTakerFee is 0.02%. A USDC<>OSMO KVStore exists with an override of 0.01%. Therefore, 0.01% is used.

Now, imagine the amount in is 1000 USDC. This means that the amount of takerFee utilized is 0.01% of 1000, which is 1 USDC. 

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

For simplicity sake, let’s say staking rewards is 40% and community pool is 60%. This means that out of the 1 USDC taken, 0.4 USDC is meant for staking rewards and 0.6 USDC is meant for community pool.

Starting with the community pool funds, the protocol checks if the fee is a whitelisted fee token. If it is, it is sent directly to the community pool. If it is not, it is sent to the `non_native_fee_collector_community_pool` module address. At epoch, the funds in this account are swapped to the `CommunityPoolDenomToSwapNonWhitelistedAssetsTo` defined in the `poolmanger` params above, and then sent all at once to the community pool at that time.

Next, for staking rewards, since this is a non-OSMO token, it is sent directly to the `non_native_fee_collector`. At epoch, all funds in this module account are swapped to OSMO and distributed directly to OSMO stakers.

### Example 2: OSMO taker fee

This example does not differ much from the previous example. In this example, a user is swapping 1000 OSMO for USDC.

Just as before, we search for a KVStore taker fee override before utilizing the default taker fee. Just as before (order does not matter), a KVStore entry for OSMO<>USDC exists, so we utilize a 0.01% taker fee instead of the 0.02% default taker fee. 0.01% of 1000 OSMO is 1 OSMO.

We now check the `OsmoTakerFeeDistribution`. In this example, let’s say its 20% to community pool and 80% to stakers. This means that 0.2 OSMO is set for community pool and 0.8 is set for stakers.

For community pool, this is just a direct send to community pool.

For staking, we actually ALSO send this to the `non_native_fee_collector`. At epoch time, this OSMO is just skipped over, while everything else is swapped to OSMO. At the very end, it takes the OSMO directly sent to the `non_native_fee_collector` along with the non native tokens that were just swapped to OSMO and distributes it to stakers.

### Important Note: How to extract the data

If one were to take the total amount of tokens in the two module accounts (`non_native_fee_collector` and `non_native_fee_collector_community_pool`), this would be slightly over exaggerating the amount that is generated from taker fees. This is because, when a user uses a non-native token as a FEE TOKEN, this is also sent to the `non_native_fee_collector`. So there are two options here to extract the info:

1. Track the delta of the module accounts 1 block after epoch X and 1 block before epoch X+1. Also, track the total non osmo txfees generated in this period. Subtract the total non osmo txfees generated in this period from the delta of the `non_native_fee_collector`. Add this to the delta of `non_native_fee_collector_community_pool` values. This is the taker fees generated
2. This is the less better way, but you can track the `SendCoinsFromAccountToModule` events from each block. The problem with this is, imagine I swap USDC to OSMO and use USDC as txfee. This would generate three `SendCoinsFromAccountToModule` events:
   1. Txfee gets sent to `non_native_fee_collector`
   2. Part of taker fee gets sent to `non_native_fee_collector`
   3. Other part of taker fee gets sent to `non_native_fee_collector_community_pool`
   You could make an assumption here that the txfee is going to be the smaller of the two that gets sent to the `non_native_fee_collector`, or better the order of operations is going to always be the same, so you can figure out if the first or second send to `non_native_fee_collector` is the txfee and not track that value