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
	tokenOutMinAmount sdk.Int) (tokenOutAmount sdk.Int, err error) {
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
		tokenOutMinAmount sdk.Int,
		spreadFactor sdk.Dec,
	) (sdk.Int, error)
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

## Multi-Hop

All tokens are swapped using a multi-hop mechanism. That is, all swaps
are routed via the most cost-efficient way, swapping in and out from
multiple pools in the process.
The most cost-efficient route is determined offline and the list of the pools is provided externally, by user, during the broadcasting of the swapping transaction.
At the moment of execution, the provided route may not be the most cost-efficient one anymore.

When a trade consists of just two OSMO-included routes during a single transaction,
the spread factors on each hop would be automatically halved.
Example: for converting `ATOM -> OSMO -> LUNA` using two pools with spread factors `0.3% + 0.2%`,
instead `0.15% + 0.1%` spread factors will be applied.

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


- **FromCoin**: (`sdk.Coin`): the total amount of tokens one wants to sell.
- **ToCoinDenom**: (`string`): is the denom they want to buy with the tokens being sold.
- **PoolId**: (`uint64`): is the identifier of the pool that the trade will happen on.
- **MaxPriceImpact**: (`sdk.Dec`): is the maximum percentage that the user is willing to affect the price of the pool.
- **TwapPrice**: (`sdk.Dec`) is an external price that the user can optionally enter to have the `MaxPriceImpact` adjusted as the `SpotPrice` of a pool could be changed at any time.

The response would be `EstimateTradeBasedOnPriceImpactResponse` which contains the following data:

- **InputCoin**: (`sdk.Coin`): the actual input amount that would be tradeable under that price impact (might be the full amount).
- **OutputCoin**: (`sdk.Coin`): the amount of the `ToCoinDenom` tokens being received for the actual `InputCoin` trade.

With that data it is easier for any entity to fill in the `MsgSwapExactAmountIn` details.

### Process

The following is the process in which the query finds a trade that will stay below the `MaxPriceImpact` value.

1. Verify `PoolId`, `FromCoin`, `ToCoinDenom` are not empty
2. Return the specific `swapModule` based on the `PoolId`
3. Return the specific `PoolI` interface from the `swapModule` based on the `PoolId`
4. Calculate the `SpotPrice` in terms of the token being bought, therefore if it's an OSMO/ATOM pool and OSMO is being sold we need to calculate the `SpotPrice` in terms of `ATOM` being the base asset and OSMO being the quote asset.
5. If we have a `TwapPrice` specified in the request we need to adjust the `MaxPriceImpact` into a new variable `adjustedMaxPriceImpact` which would either increase if the `SpotPrice` is cheaper than the `TwapPrice` or decrease if the `SpotPrice` is more expensive leaving less room to estimate a trade.
   1. If the `adjustedMaxPriceImpact` was calculated to be `0` or negative it means that the `SpotPrice` is more expensive than the `TwapPrice` and has already exceeded the possible `MaxPriceImpact`. We return a `sdk.ZeroInt()` input and output for the input and output coins indicating that no trade is viable.
6. Firstly we execute `CalcOutAmtGivenIn` using the initial amount given from `FromCoin` without a swap fee as we're only interested in the price impact of the trade which doesn't involve swap fees.
   1. If this errors there is very little liquidity for the pool to estimate anything.
   2. If the amount of tokens returned is `0` it means we've entered a very small amount and we are trying to estimate a trade on dust. We should return the full input amount `FromCoin` as the `InputCoin` and `OutputCoin` of `sdk.ZeroInt()` to indicate we will get `0` from our trade.
7. We calculate the `currTradePrice` by dividing the `FromCoin.Amount` by the `tokenOut.Amount` received from the calculation.
8. Using the `currTradePrice` we get the `priceDeviation` calculated by (`currTradePrice` - `SpotPrice`)/`SpotPrice`.
   1. If the `priceDeviation` is less than or equal to the `adjustedMaxPriceImpact` it means that the first trade of `FromCoin` does not exceed the `MaxPriceImpact`. Therefore we can return the amount of tokens we will get out of this after re-executing `CalcOutAmtGivenIn` with the swap-fee included.
9. If the `priceDeviation` is greater than the `adjustedMaxPriceImpact` we begin a binary search loop halfing the input amount in `FromCoin` each time until a trade is estimated within the `adjustedMaxPriceImpact` or it fails.
10. We initialized three variables, `lowAmount=1`, `highAmount=FromCoin.Amount`, `currFromCoin=FromCoin`
11. While the `lowAmount` is less than or equal to the `highAmount`:
    1.  Calculate the `midAmount` by adding the `lowAmount` to the `highAmount` and dividing by 2.
    2.  The new `currFromCoin` now has the `midAmount` as it's amount of tokens.
    3.  Calculate the `tokenOut` using the new input value `currFromCoin` and `CalcOutAmtGivenIn` without a swap fee.
        1.  If there is an `err` or the `tokenOut` is `sdk.ZeroInt()` we need to return `InputCoin` as the full `FromCoin` and the `OutputCoin` as `sdk.ZeroInt()` to indicate we are trying to trade dust.
    4.  Re-calculate the `currTradePrice` and the `priceDeviation` for the newest input/output amounts.
        1.  If the `priceDeviation` is less that or equal to `adjustedmaxPriceImpact` the `lowAmount` is set to the `midAmount + 1`, this will cause the process to re-loop at a higher `currFromCoin` amount.
        2.  Otherwise we set the `highAmount` to the `midAmount-1`, this will cause the process to re-loop at a lower `currFromCoin` amount.
    5.  Once the `lowAmount` is less than or equal to `highAmount` the loop breaks.
12. If the `highAmount` looped all the way to `sdk.Zero()` it means no amount of tokens could have been trade to get a trade that respects the `adjustedPriceImpact`, this is usually because of low-liquidity pools.
13. Otherwise we have successfully found a trade that respects the `adjustedPriceImpact` therefore we need to re-calculate the `currFromCoin` trade this time with the swap fee of the pool to give an accurate `OutputCoin`.
14. Return the `InputCoin`, `OutputCoin` trade estimation.