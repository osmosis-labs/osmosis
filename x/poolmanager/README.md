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
   * Note, that pool creation messages are recieved by the pool model's message server.
   Each module's message server then calls the `x/poolmanager` keeper method `CreatePool`.
2. Handle swaps
   * Cover & share multihop logic
   * Propagate intra-pool swaps to the appropriate module depending on the pool type.
   * Contrary to pool creation, swap messages are recieved by the `x/poolmanager` message server.

Let's consider pool creation and swaps separately and in more detail.

## Pool Creation & Id Management

To make sure that the pool ids are unique across the two modules, we unify pool id management
in the `poolmanager`.

When a call to `CreatePool` keeper method is received, we get the next pool id from the module
storage, assign it to the new pool and propagate the execution to either `gamm`
or `concentrated-liquidity` modules.

Note that we define a `CreatePoolMsg` interface:
<https://github.com/osmosis-labs/osmosis/blob/f26ceb958adaaf31510e17ed88f5eab47e2bac03/x/poolmanager/types/msg_create_pool.go#L9>

For each of `balancer`, `stableswap` and `concentrated-liquidity` pools, we have their
own implementation of `CreatePoolMsg`.

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

2. `CreatePool` keeper method is called  from `poolmanager`, propagating
the appropriate implemenation of the `CreatePoolMsg` interface.

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

Where swapmodule is either `gamm` or `concentrated-liquidity` keeper.

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
This mapping is going to be neccessary for knowing where to route the swap messages.

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

There are 2 swap messages:

- `MsgSwapExactAmountIn`
- `MsgSwapExactAmountOut`

Their implementation of routing is similar. As a result, we only focus on `MsgSwapExactAmountIn`.

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
		swapFee sdk.Dec,
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

`tokenBalanceOut * [1 - { tokenBalanceIn / (tokenBalanceIn + (1 - swapFee) * tokenAmountIn)} ^ (tokenWeightIn / tokenWeightOut)]`

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


## Multi-Hop

All tokens are swapped using a multi-hop mechanism. That is, all swaps
are routed via the most cost-efficient way, swapping in and out from
multiple pools in the process. 
The most cost-efficient route is determined offline and the list of the pools is provided externally, by user, during the broadcasting of the swapping transaction. 
In the moment of the execution the provided route may not be the most cost efficient one anymore.

When a trade consists of just two OSMO-included routes during a single transaction,
the swap fees on each hop would be automatically halved. 
Example: for converting `ATOM -> OSMO -> LUNA` using two pools with swap fees `0.3% + 0.2%`,
instead `0.15% + 0.1%` fees will be aplied. 

[Multi-Hop](https://github.com/osmosis-labs/osmosis/blob/f26ceb958adaaf31510e17ed88f5eab47e2bac03/x/poolmanager/router.go#L16)
