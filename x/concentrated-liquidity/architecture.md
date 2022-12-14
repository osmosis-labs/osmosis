# Concentrated Liquidity

## Background

Concentrated liquidity is a novel AMM design that allows for a more efficient use of capital.
The improvement is achieved by providing liquidity in specific ranges chosen by user.

A naive example is a pool with stable pairs such as USDC/USDT, where the price should always be near 1.
As a result, LPs can focus their capital in a small range around 1 as opposed to full range, leading
to an average of 200-300x higher capital efficiency.

At the same time, traders enjoy lower slippage as greater depth is incentives to occur around the
current price.

This design also allows for a new "range order" type that is similar to a limit order with order-books.

The introduction of concentrated liquidity creates new opportunities for providing liquidity rewards to desired strategies.
For example, it is possible to incentivize LPs based on the closeness to the current price and the time spent
within a position.

This document describes the final version of the desired product. However, the work is split into multiple phases (milestones).
See "Milestones" section for more details.

## Architecture

TODO: understand how much detail is wanted in this section

Our traditional balancer AMM relies on the following curve that tracks current reserves:
$$xy = k$$

It allows for distributing the liquidity along the xy=k curve and across the entire price
range (0, \infinity). TODO: format correctly

With the new architecture, we introduce a concept of a `position` that allows a user to
concentrate liquidity within a fixed range. A position only needs to maintain
enough reserves to satisfy trading within this range. As a result, it functions
as the traditional `xy = k` within that range.

With the new architecture, the real reserves are described by the following formula:
$$(x + L / \sqrt P_u)(y + L \sqrt P_l) = L^2$$
- `P_l` is the lower tick
- `P_u` is the upper tick

where L is the amount of liquidity provided $$L = \sqrt k$$

This formula is stemming from the original $$xy = k$$ with the range being limited.

In the traditional design, a pool's tokens `x` and `y` are tracked directly. With the concentrated design, we only
track `L` and `\sqrt P` which can be calculated with:

$$L = \sqrt (xy)$$

$$\sqrt P = y / x$$

By re-arranging the above, we get the following to track the virtual reserves:

$$x = L / \sqrt P$$

$$y = L \sqrt P$$

Note the square root around price. By tracking it this way, we can utilize
the following property that is the core of the architecture:

$$L = \Delta Y / \Delta \sqrt P$$

Since only one of the following change at a time:
- $$L$$
   * when a LP adds or removes liquidity
- $$\sqrt P$$
   * when a trader swaps

We can use the above relationship for calculating the outcome of swaps and pool joins that mint shares.

Conversely, we calculate liquidity from the other token in the pool:

$$\Delta x = \Delta \frac {1}{\sqrt P}  L$$

## Ticks

To allow for providing liquidity within certain price ranges, we will introduce the concept of a `tick`. Each tick is a function of price, allowing to partition the price
range into discrete segments (which we refer to here as ticks):

$$p(i) = 1.0001^i$$

where `p(i)` is the price at tick `i`. Taking powers of 1.0001 has a property of two ticks being 0.01% apart (1 basis point away).

Therefore, we get values like:

$$\sqrt{p(-1)} = 1.0001^{-1/2} \approx 0.99995$$

$$\sqrt{p(0)} = 1.0001^{0/2} = 1$$

$$\sqrt{p(1)} = \sqrt{1.0001} = 1.0001^{1/2} \approx 1.00005$$

TODO: tick range bounds

### User Stories

We define the feature in terms of user stories.
Each story, will be tracked as a discrete piece of work with
its own set of tasks.

The following is the list of user stories:

#### Concentrated Liquidity Module

> As an engineer, I would like the concentrated liquidity logic to exist in its own module so that I can easily reason about the concentrated liquidity abstraction that is different from the existing pools.

Therefore, we create a new module `concentrated-liquidity`. It will include all low-level
logic that is specific to minting, burning liquidity, and swapping within concentrated liquidity pools.

Under the "Liquidity Provision" user story, we will track tasks specific to defining
foundations, boilerplate, module wiring and their respective tests.

While low-level details for providing, burning liquidity, and swapping functions are to be tracked in their own user stories, we define
all messages here.

##### `MsgCreatePosition`

- **Request**

This message allows LPs to provide liquidity between `LowerTick` and `UpperTick` in a given `PoolId.
The user provides the amount of each token desired. Since LPs are only allowed to provide
liquidity proportional to the existing reserves, the actual amount of tokens used might differ from requested.
As a result, LPs may also provide the minimum amount of each token to be used so that the system fails
to create position if the desired amounts cannot be satisfied.

```go
type MsgCreatePosition struct {
	PoolId          uint64
	Sender          string
	LowerTick       int64
	UpperTick       int64
	TokenDesired0   types.Coin
	TokenDesired1   types.Coin
	TokenMinAmount0 github_com_cosmos_cosmos_sdk_types.Int
	TokenMinAmount1 github_com_cosmos_cosmos_sdk_types.Int
}
```

- **Response**

On succesful response, we receive the actual amounts of each token used to create the
liquidityCreated number of shares in the given range.

```go
type MsgCreatePositionResponse struct {
	Amount0 github_com_cosmos_cosmos_sdk_types.Int
	Amount1 github_com_cosmos_cosmos_sdk_types.Int
    LiquidityCreated github_com_cosmos_cosmos_sdk_types.Int
}
```

This message should call the `createPosition` keeper method that is introduced in the `"Liquidity Provision"` section of this document.

##### `MsgWithdrawPosition`

- **Request**

This message allows LPs to withdraw their position in a given pool and range (given by ticks), potentially in partial
amount of liquidity. It should fail if there is no position in the given tick ranges, if tick ranges are invalid,
or if attempting to withdraw an amount higher than originally provided.

```go
type MsgWithdrawPosition struct {
	PoolId          uint64
	Sender          string
	LowerTick       int64
	UpperTick       int64
	LiquidityAmount github_com_cosmos_cosmos_sdk_types.Int
}
```

- **Response**

On succesful response, we receive the amounts of each token withdrawn
for the provided share liquidity amount.

```go
type MsgWithdrawPositionResponse struct {
	Amount0 github_com_cosmos_cosmos_sdk_types.Int
	Amount1 github_com_cosmos_cosmos_sdk_types.Int
}
```

This message should call the `withdrawPosition` keeper method that is introduced in the `"Liquidity Provision"` section of this document.

##### `SwapExactAmountIn` Keeper Method

This method has the same interface as the pre-existing `SwapExactAmountIn` in the `x/gamm` module.
It takes an exact amount of coins of one denom in to return a minimum amount of tokenOutDenom.

```go
// x/concentrated-liquidity/pool.go SwapExactAmountIn(...)

func (k Keeper) SwapExactAmountIn(
	ctx sdk.Context,
	sender sdk.AccAddress,
	pool gammtypes.PoolI,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	tokenOutMinAmount sdk.Int,
	swapFee sdk.Dec,
) (tokenOutAmount sdk.Int, err error) {
    ...
}
```

This method should be called from the new `swap-router` module's `RouteExactAmountIn` initiated by the `MsgSwapExactAmountIn`.
See the next `"Swap Router Module"` section of this document for more details.

##### `SwapExactAmountOut` Keeper Method

This method is comparable to `SwapExactAmountIn`. It has the same interface as the pre-existing `SwapExactAmountOut` in the `x/gamm` module.
It takes an exact amount of coins of one denom out to return a maximum amount of tokenInDenom.

```go
// x/concentrated-liquidity/pool.go SwapExactAmountOut(...)

func (k Keeper) SwapExactAmountOut(
	ctx sdk.Context,
	sender sdk.AccAddress,
	poolI gammtypes.PoolI,
	tokenInDenom string,
	tokenInMaxAmount sdk.Int,
	tokenOut sdk.Coin,
	swapFee sdk.Dec,
) (tokenInAmount sdk.Int, err error) {
	...
}
```

This method should be called from the new `swap-router` module's `RouteExactAmountOut` initiated by the `MsgSwapExactAmountOut`.
See the next `"Swap Router Module"` section of this document for more details.

##### `InitializePool` Keeper Method

This method is part of the implementation of the `SwapI` interface in `swaprouter`
module. "Swap Router Module" section discussed the interface in more detail.

```go
// x/concentrated-liquidity/pool.go InitializePool(...)

func (k Keeper) InitializePool(
    ctx sdk.Context,
    pool types.PoolI,
    creatorAddress sdk.AccAddress) error {
    ...
}
```

This method should be called from the new `swap-router` module's `CreatePool` initiated by the `MsgCreatePool`.
See the next `"Swap Router Module"` section of this document for more details.

#### Swap Router Module

> As a user, I would like to have a unified entrypoint for my swaps regardless of the underlying pool implementation so that I don't need to reason about API complexity

> As a user, I would like the pool management to be unified so that I don't have to reason about additional complexity stemming from divergent pool sources.

With the new `concentrated-liquidity` module, we now have a new entrypoint for swaps that is
the same with the existing `gamm` module.

To avoid fragmenting swap and pool creation entrypoints and duplicating their boilerplate logic,
we would like to define a new `swaprouter` module. Its purpose is twofold:
1. Handle pool creation messages
   * Assign ids to pools
   * Store the mapping from pool id to one of the swap modules (`gamm` or `concentrated-liquidity`)
   * Propagate the execution the appropriate module depending on the pool type.
2. Handle swap messages
   * Cover & share multihop logic
   * Propagate intra-pool swaps to the appropriate module depending on the pool type.

Therefore, we move several existing `gamm` messages and tests to the new `swap-router` module,
connecting them to the `swaprouter` keeper that propagates execution to the appropriate swap module.

The messages to move from `gamm` to `swaprouter` are:
- `CreatePoolMsg`
- `MsgSwapExactAmountIn`
- `MsgSwapExactAmountOut`

Let's consider pool creation and swaps separately and in more detail.

##### Pool Creation & Id Management

To make sure that the pool ids are unique across the two modules, we unify pool id management
in the `swaprouter`.

We migrate the store index `next_pool_id` from `gamm` to `swaprouter`. This index represents
the next available pool id to assign to a newly created pool.

When `CreatePoolMsg` is received, we get the next pool id, assign it to the new pool and propagate
the execution to either `gamm` or `concentrated-liquidity` modules.

Note that we define a `CreatePoolMsg` interface:

```go
type CreatePoolMsg interface {
	// GetPoolType returns the type of the pool to create.
	GetPoolType() PoolType
	// The creator of the pool, who pays the PoolCreationFee, provides initial liquidity,
	// and gets the initial LP shares.
	PoolCreator() sdk.AccAddress
	// A stateful validation function.
	Validate(ctx sdk.Context) error
	// Initial Liquidity for the pool that the sender is required to send to the pool account
	InitialLiquidity() sdk.Coins
	// CreatePool creates a pool implementing PoolI, using data from the message.
	CreatePool(ctx sdk.Context, poolID uint64) (gammtypes.TraditionalAmmInterface, error)
}
```

For each of `balancer`, `stableswap` and `concentrated-liquidity` pools, we have their
own implementation of `CreatePoolMsg`.

Note the `PoolType` type. This is an enumeration of all supported pool types.
We proto-generate this enumeration:

```go
// proto/osmosis/swaprouter/v1beta1/module_route.proto
// generates to x/swaprouter/types/module_route.pb.go

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

1. `CreatePoolMsg` is received by the `swaprouter` message server.

2. `CreatePool` `swaprouter` keeper method is called.

```go
// x/swaprouter/creator.go CreatePool(...)

// CreatePool attempts to create a pool returning the newly created pool ID or
// an error upon failure. The pool creation fee is used to fund the community
// pool. It will create a dedicated module account for the pool and sends the
// initial liquidity to the created module account.
//
// After the initial liquidity is sent to the pool's account, shares are minted
// and sent to the pool creator. The shares are created using a denomination in
// the form of < swap module name >/pool/{poolID}. In addition, the x/bank metadata is updated
// to reflect the newly created GAMM share denomination.
func (k Keeper) CreatePool(ctx sdk.Context, msg types.CreatePoolMsg) (uint64, error) {
    ...
}
```

3. The keeper utilizes `CreatePoolMsg` interface methods to execute the logic specific
to each pool type.

4. Lastly, `swaprouter.CreatePool` routes the execution to the appropriate module.

The propagation to the desired module is ensured by the routing table stored in memory in the `swaprouter` keeper.

```go
// x/swaprouter/keeper.go NewKeeper(...)

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

As a result, `swaprouterkeeper.CreatePool` can route the execution to the appropriate module in
the following way:

```go
// x/swaprouter/creator.go CreatePool(...)

swapModule := k.routes[msg.GetPoolType()]

if err := swapModule.InitializePool(ctx, pool, sender); err != nil {
    return 0, err
}
```

Where swapmodule is either `gamm` or `concentrated-liquidity` keeper.

Both of these modules implement the `SwapI` interface:

```go
// x/swaprouter/types/routes.go SwapI interface

type SwapI interface {
    ...

	InitializePool(ctx sdk.Context, pool gammtypes.PoolI, creatorAddress sdk.AccAddress) error
}
```

As a result, the `swaprouter` module propagates core execution to the appropriate swap module.

Lastly, the `swaprouter` keeper stores a mapping from the pool id to the pool type.
This mapping is going to be neccessary for knowing where to route the swap messages.

To achieve this, we create the following store index:

```go
// x/swaprouter/types/keys.go

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

##### Swaps

There are 2 swap messages:

- `MsgSwapExactAmountIn`
- `MsgSwapExactAmountOut`

Their implementation of routing is similar. As a result, we only focus on `MsgSwapExactAmountIn`.

Once the message is received, it calls `RouteExactAmountIn`

```go
// x/swaprouter/router.go RouteExactAmountIn(...)

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

The bulk of its implementation is ported from `gamm`'s `MultihopSwapExactAmountIn`.
Essentially, the method iterates over the routes and calls a `SwapExactAmountIn` method
for each, subsequently updating the inter-pool swap state.

The routing works by querying the index `SwapModuleRouterPrefix`,
searching up the `swaprouterkeeper.router` mapping, and callig
the appropriate `SwapExactAmountIn` method.

```go
// x/swaprouter/router.go RouteExactAmountIn(...)

moduleRouteBytes := osmoutils.MustGet(swaproutertypes.FormatModuleRouteIndex(poolId))
moduleRoute, _ := swaproutertypes.ModuleRouteFromBytes(moduleRouteBytes)

swapModule := k.routes[moduleRoute.PoolType]

_ := swapModule.SwapExactAmountIn(...)
```
- note that error checks and other details are omitted for brevity.

Similar to pool creation logic, we are able to call `SwapExactAmountIn` on any of the swap
modules by implementing the `SwapI` interface:

```go
// x/swaprouter/types/routes.go SwapI interface

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

##### GAMM Migrations

Previously we managed and stored "next pool id" and "pool creation fee" in gamm. Now, these values
are stored in the `swaprouter` module. As a result, we perform store migration in the
upgrade handler.

Some of the queries such as `x/gamm` `NumPools depended on the "next pool id" being present in `x/gamm`.
Since it is now moved, we introduce a new "pool count" index in `x/gamm` to keep track of the number
of pools. TODO: do we even need this? Consider removing before release. Path forward TBD.

In summary, we perform the following store migrations in the upgrade handler:
- migrate "next pool id` from `x/gamm` to `x/swaprouter`
- migrate "pool creation fee" from `x/gamm` to `x/swaprouter`
- create "pool count" index in `x/gamm` TODO: do we even need this? Consider removing before release. Path forward TBD.

#### GAMM Refactor

> As an engineer, I would like the gamm module to be cohesive and only focus on the logic
related to the `TraditionalAmmInterface` pool implementations.

TODO: describe and document all the changes in the gamm module in more detail.
- refer to previous sections ("Swap Router Module" and "Concentrated Liquidity Module")
to avoid repetition.

##### Swaps

We rely on the pre-existing swap methods located in `x/gamm/keeper/pool.go`:
- `SwapExactAmountIn`
- `SwapExactAmountOut`

Similarly to `concentrated-liquidity` module, these methods now implement the `swaprouter` `SwapI` interface.
However, the concrete implementations of the methods are unchanged from before the refactor.

##### New Functionality

##### `InitializePool` Keeper Method

This method is part of the implementation of the `SwapI` interface in `swaprouter`
module. "Swap Router Module" section discussed the interface in more detail.

This is the second implementation of the interface, the first being in the `concentrated-liquidity` module.

```go
// x/gamm/keeper/pool.go InitializePool(...)

func (k Keeper) InitializePool(
    ctx sdk.Context,
    pool types.PoolI,
    creatorAddress sdk.AccAddress) error {
    ...
}
```

This method should be called from the new `swap-router` module's `CreatePool` initiated by the `MsgCreatePool`.
See the next `"Swap Router Module"` section of this document for more details.

##### Removed Functionality

TODO:
- reiterate swap messages moved
- reiterate create pool messages moved
- reiterate state migrated and moved
- queries and CLI commands removed or ported
- any important tests removed or ported

#### Liquidity Provision

> As an LP, I want to provide liquidity in ranges so that I can achieve greater capital efficiency

This a basic function that should allow LPs to provide liquidity in specific ranges
to a pool. 

A pool's liquidity is consisted of two assets: asset0 and asset1. In all pools, asset0 will be the lexicographically smaller of the two assets. At the current price tick, the bucket at this tick consists of a mix of both asset0 and asset1 and is called the virtual liquidity of the pool (or "L" for short). Any positions set below the current price are consisted solely of asset0 while positions above the current price only contain asset1.

Therefore in `Mint`, we can either provide liquidity above or below the current price, which would act as range (limit) orders or decide to provide liquidity at the current price. 

As declared in the API for mint, users provide the upper and lower tick to denote the range they want to provide the liquidity for. The users are also prompted to provide the amount of token0 and token1 they desire to receive. The liquidity that needs to be provided for the token0 and token1 amount provided would be then calculated by the following methods: 

Liquidity needed for token0:
$$L = \frac{\Delta x \sqrt{P_u} \sqrt{P_l}}{\sqrt{P_u} - \sqrt{P_l}}$$

Liquidity needed for token1:
$$L = \frac{\Delta y}{\sqrt{P_u}-\sqrt{P_l}}$$

//TODO: what does this mean
With the larger liquidity including the smaller liquidity, we take the smaller liquidity calculated for both token0 and token1 and use that as the liquidity throughout the rest of the joining process. Note that the liquidity used here does not represent an amount of a specific token, but the liquidity of the pool itself, represented in sdk.Int.

Using the provided liquidity, now we calculate the delta amount of both token0 and token1, using the following equations, where L is the liquidity calculated above:

$$\Delta x = \frac{L(\sqrt{p(i_u)} - \sqrt{p(i_c)})}{\sqrt{p(i_u)}\sqrt{p(i_c)}}$$
$$\Delta y = L(\sqrt{p(i_c)} - \sqrt{p(i_l)})$$


The deltaX and the deltaY would be the actual amount of tokens joined for the requested position. 

Given the parameters needed for calculating the tokens needed for creating a position for a given tick, the API in the msg server layer would look like the following:

```go
func createPosition(
    ctx sdk.Context,
    poolId uint64,
    owner sdk.AccAddress,
    minAmountToken0,
    minAmountToken1 sdk.Int,
    lowerTick,
    upperTick int64) (amount0, amount1 sdk.Int, error) {
        ...
}
```

#### Swapping

> As a trader, I want to be able to swap over a concentrated liquidity pool so that my trades incur lower slippage

Unlike balancer pools where liquidity is spread out over an infinite range, concentrated liquidity pools allow for deeper liquidity at the current price, which in turn allows trades the incur less slippage.

Despite this improvement, the liquidity at the current price is still finite and large single trades, times of high volume, and trades against volatile assets are eventually bound to incur some slippage.

In order to determine the depth of liquidity and subsequent amountIn/amountOut values for a given pool, we track the swap's state across multiple swap "steps". You can think of each of these steps as the current price following the original xy=k curve, with the far left bound being the next initialized tick below the current price and the far right bound being the next initialized tick above the current price. It is also important to note that we always view prices of asset1 in terms of asset0, and selling asset1 for asset0 would in turn increase it's spot price. The reciprocal is also true, where if we sell asset0 for asset1 we would decrease the pool's spot price.

When a user swaps asset0 for asset1 (can also be seen as "selling" asset0), we move left along the curve until asset1 reserves in this tick are depleted. If the tick of the current price has enough liquidity to fulfil the order without stepping to the next tick, the order is complete. If we deplete all of asset1 in the current tick, this then marks the end of the first swap "step". Since all liquidity in this tick has been depleted, we search for the next closest tick to the left of the current tick that has liquidity. Once we reach this tick, we determine how much more of asset1 is needed to complete the swap. This process continues until either the entire order is fulfilled or all liquidity is drained from the pool.

The same logic is true for swapping asset1, which is analogous to buying asset0, however instead of moving left along the set of curves, we instead search for liquidity to the right.

The core logic is run by the computeSwapStep function, where we calculate the amountIn, amountOut, and the next sqrtPrice given current price, price target, tick liquidity, and amount available to swap:

```go
func computeSwapStep(
    sqrtPriceCurrent sdk.Dec,
    sqrtPriceTarget sdk.Dec,
    liquidity sdk.Dec,
    amountRemaining sdk.Dec,
    lte bool) (sqrtPriceNext, amountIn, amountOut sdk.Dec)
{
        ...
}
```

TODO: add formulas, specific steps for calculating swaps and relation to fees.

#### Range Orders

> As a trader, I want to be able to execute ranger orders so that I have better control of the price at which I trade

TODO

#### Fees

> As a an LP, I want to earn fees on my capital so that I am incentivized to participate in the market making actively.

For our balancer style pools, fees go back directly to the pool to benefit LPs.
For a concentrated liquidity pool, this is no longer possible due to the non-fungible property
of positions. As a result, there is a different accumulator-based mechanism for keeping track
of and storing the fees.

First, note that the fees are collected in tokens themselves rather than in units of liquidity.
Thus, we need two accumulators for each token. Temporally, these fee accumulators are accessed together
from state most of the time. Therefore, we define a data structure for storing the fees of each token in the pool.

```go
// Note that this is proto-generated.

type Fee struct {
    TokenZero
    TokenOne
}
```

The only time when we need to load only one of the token fee accumulators is during swaps.
The performance overhead of loading both accumulators is negligible so we choose a better
abstraction over small performance gain.


We define the following accumulators and fee-related
fields to be stored on various layers of state:

- **Per-pool**

```go
// Note that this is proto-generated.
type Pool struct {
    ...
    SwapFee sdk.Dec
    FeeGrowthGlobalOutside Fee
}
```

Each pool is initialized with an immutable fee value `SwapFee` to be paid by
the swappers. It is denominated in units of hundredths of a basis point `0.0001%`.
// TODO: from uniswap whitepaper. What is the reason for this denomination?

`FeeGrowthGlobalOutside` represents the total amount of fees that have been earned
per unit of virtual liquidity in each token `L` from the time of the creation of the pool.

Assume that we deposited 1 unit of full-range liqudity at pool creation. A `FeeGrowthGlobalOutside.Token0`
value shows how much of token0 have been earned by that unit of liquidity up until today.

- **Per-tick**

```go
// Note that this is proto-generated.
type Tick struct {
    ...
   FeeGrowthOutside Fee
}
```

Ticks keep record of fees accumulated outside of them.
This is required for calcuating the amount of fees accrued within a range.

Note, keeping track of the accumulators is only necessary for the ticks that have
been initialized. In other words, there is at least one position referencing that tick.

By convention, when a new tick is activated, it is set to the respective `feeGrowthOutsideX`
if the tick being initialized is below the current tick. This is equivalent to assumming that 
all fees have been accrued below the initialized tick.

In the example code snippets below, we only focus on the token0. The token1 is analogous. 

```go
tick.FeeGrowthOutside.Token0 := sdk.ZeroDec()

if initializedTickNum <= pool.CurrentTick {
    tick.FeeGrowthOutside.Token0 = pool.FeeGrowthGlobalOutside.Token0
}
```

Essentially, setting tick's `tick.FeeGrowthOutside.TokenX` to the global `pool.FeeGrowthGlobalOutside.TokenX`
represents the amount of fees collected by the pool up until the tick was activated.

Once a tick is activated again (crossed in either direction), `tick.FeeGrowthOutside.TokenX` is
updated to add the difference between `pool.FeeGrowthGlobalOutsideX` and the old value of
`tick.FeeGrowthOutside.TokenX`. 

```go
tick.FeeGrowthOutside.Token0 =  tick.FeeGrowthOutside.Token0.Add(pool.FeeGrowthGlobalOutside.Token0.Sub(tick.FeeGrowthOutside.Token0))
```

Tracking how much of fees are collected outside of a tick allows us to calculate the amount
of fees inside the position on request.

Intuitively, we update the activated tick with the amount of fees collected for
every tick lower than the tick that is being crossed.

This has two benefits:
 * We avoid updating *all* ticks
 * We can calculate a range by subtracting the top and bottom ticks for the range
 using formulas below.

Assume `FeeGrowthBelowLowerTick0` and `FeeGrowthAboveUpperTick0`.

We calculate the fee growth below the lower tick in the following way:

```go
var feeGrowthBelowLowerTick0 sdk.Dec

if pool.CurrentTick >= lowerTickNum {
    feeGrowthBelowLowerTick0 = pool.FeeGrowthOutside.Token0
} else {
    feeGrowthBelowLowerTick0 = pool.FeeGrowthGlobalOutside.Token0 - lowerTick.FeeGrowthOutside.Token0
}
```

We calculate the fee growth above the upper tick in the following way:

```go
var feeGrowthAboveUpperTick0 sdk.Dec

if pool.CurrentTick >= upperTickNum {
    feeGrowthAboveUpperTick0 = pool.FeeGrowthGlobalOutside.Token0 - upperTick.FeeGrowthOutside.Token0
} else {
    feeGrowthAboveUpperTick0 = pool.FeeGrowthOutside.Token0
}
```

Now, by having the fee growth below the lower and above the upper tick of a range,
we can calculate the fee growth inside the range by subtracting the two from the
global per-unit-of-liquidity fee growth.

```go
feeGrowthInsideRange0 := pool.FeeGrowthGlobalOutside.Token0 - feeGrowthBelowLowerTick0 - feeGrowthAboveUpperTick0
```

Note that although `tick.FeeGrowthOutside.Token0` may be initialized at a different
point in time for each tick, the comparison of these values between ticks
is not meaningful. There is also no guarantee that the values
across ticks will follow any particular pattern. 

However, this does not affect the per-position calculations since
all the position needs to know is the fee growth inside the position's
range since the position was last touched.

- **Per-position**

type Position struct {
    FeeGrowthInsideLast Fee
    UncollectedFee Fee
}

Recall that contrary to traditional pools, in a concentrated liquidity pool,
fees do not get auto re-injected into the pool. Instead, they are tracked by
`position.TokensUncollected0` and `position.TokensUncollected1` fields of each position.

The `position.FeeGrowthInside0Last` and `position.FeeGrowthInside1Last` accumulators
are used to calculate the  _uncollected fees_ to add to `position.TokensUncollected0`
and `position.TokensUncollected1`.

The amount of uncollected fees needs to be calculated every time a user modifies
their position. That is when a position is created, liquidity is added or removed.

We must recalculate the values for any modification because with more liquidity
added to the position, the amount of fees collected by the position increases.

Let `feeGrowthInside0` be the amount of fee growth per unit of liquidity within
the position's ticks. We use the same strategy for computing fees between two ticks (in-range) that
was described in the previous section. Once we have `feeGrowthInside0` computed, we update the
`position.UncollectedFee.Token0` and `position.FeeGrowthInsideLast.Token0`.

```go
// Note that tokensUncollected0 is 0 when a position is created.
if !isPositionNew {
    uncollectefFeeAddition := pool.Liquidity.Mul(feeGrowthInside.Token0.Sub(position.FeeGrowthInsideLast.Token0))
    position.UncollectedFee.Token0 = position.UncollectedFee.Token0.Add(uncollectefFeeAddition)
}

position.FeeGrowthInsideLast.Token0 = feeGrowthInside.Token0
```

##### Collecting Fees

Collecting fees is as simple as transferring the requested amount
from the pool address to the position's owner.

After every epoch, the system iterates over all positions to call
`collectFees` for each and auto-collects fees.

Currently, there is no ability to collect manually to prevent spam.

```go
func (k Keeper) collectFees(
    owner sdk.AccAddress,
    lowerTick, upperTick int64) error {
    // validate ticks

    // get position if exists

    // bank send position.TokensUncollected0 and position.TokensUncollected1
    //from pool address to position owner
    // TODO: revisit to make sure truncations are handled correctly.
}
```

##### Swaps

Swapping within a single tick works as the regular `xy = k` curve. For swaps
across ticks to work, we simply apply the same fee calculation logic for every swap step.

Consider data structures defined above. Let `tokenInAmt` be the amount of token being
swapped in.

Then, to calculate the fee within a single tick, we perform the following steps:

1. Calculate an updated `tokenInAmtAfterFee` by charging the `pool.SwapFee` on `tokenInAmt`.

```go
// Update global fee accumulator tracking fees for denom of tokenInAmt.
// TODO: revisit to make sure if truncations need to happen.
pool.FeeGrowthGlobalOutside.TokenX = pool.FeeGrowthGlobalOutside.TokenX.Add(tokenInAmt.Mul(pool.SwapFee))

// Update tokenInAmt to account for fees.
fee = tokenInAmt.Mul(pool.SwapFee).Ceil()
tokenInAmtAfterFee = tokenInAmt.Sub(fee)

k.bankKeeper.SendCoins(ctx, swapper, pool.GetAddress(), ...) // send tokenInAmtAfterFee
```

2. Proceed to calculating the next square root price by utilizing the updated `tokenInAmtAfterFee.

Depending on which of the tokens in `tokenIn`,

If token1 is being swapped in:
$$\Delta \sqrt P = \Delta y / L$$

Here, `tokenInAmtAfterFee` is delta y.

If token0 is being swapped in:
$$\Delta \sqrt P = L / \Delta x$$

Here, `tokenInAmtAfterFee` is delta x.

Once we have the updated square root price, we can calculate the amount of `tokenOut` to be returned.
The returned `tokenOut` is computed with fees accounted for given that we used `tokenInAmtAfterFee`.

#### Liquidity Rewards

> As an LP, I want to earn liquidity rewards so that I am more incentivized to provide liquidity in the ranges closer to the price.

TODO



##### State

- global (per-pool)

- per-tick

- per-position

### Additional Requirements

#### GAMM Refactor

TODO

### Risks

TODO

### Milestones

#### Milestone 1 - Swap Within a Single Tick

TODO

#### Placeholder

### Terminology

We will use the following terms throughout the document:

- `Virtual Reserves` - TODO

- `Real Reserves` - TODO

- `Tick` - TODO

- `Position` - TODO

- `Range` - TODO

### External Sources

- [Uniswap V3 Whitepaper](https://uniswap.org/whitepaper-v3.pdf)
- [Technical Note on Liquidity Math](https://atiselsts.github.io/pdfs/uniswap-v3-liquidity-math.pdf)
