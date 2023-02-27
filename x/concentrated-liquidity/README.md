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

Our traditional balancer AMM relies on the following curve that tracks current reserves:
$$xy = k$$

It allows for distributing the liquidity along the xy=k curve and across the entire price
range $(0, &infin;)$.

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
track $L$ and $\sqrt P$ which can be calculated with:

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

### Ticks

#### Context

In Uniswap V3, discrete points (called ticks) are used when providing liquidity in a concentrated liquidity pool. The price [p] corresponding to a tick [t] is defined by the equation:

$$ p(i) = 1.0001^t $$

This results in a .01% difference between adjacent tick prices. However, this does not allow for control over the specific prices that the ticks correspond to. For example, if a user wants to make a limit order at the $17,100.50 price point, they would have to interact with either tick 97473 (corresponding to price $17,099.60) or tick 97474 (price $17101.30).

Since we know what range a pair will generally trade in, how do we go about providing more granularity at that range and provide a more optimal price range between ticks instead of the "one-size-fits-all" approach explained above?

#### Geometric Tick Spacing with Additive Ranges

In Osmosis's implementation of concentrated liquidity, we will instead make use of geometric tick spacing with additive ranges.

We start by defining an exponent for the precision factor of 10 at a spot price of one - $exponentAtPriceOne$.

For instance, if $exponentAtPriceOne = -4$ , then each tick starting at 1 and ending at the first factor of 10 will represents a spot price increase of 0.0001. At this precision factor:
* $tick_0 = 1$ (tick 0 is always equal to 1 regardless of precision factor)
* $tick_1 = 1.0001$
* $tick_2 = 1.0002$
* $tick_3 = 1.0003$

This continues on until we reach a spot price of 10. At this point, since we have increased by a factor of 10, our $exponentAtCurrentTick$ increases from -4 to -3, and the ticks will increase as follows:
* $tick_{89999} =  9.9999$
* $tick_{90000} = 10.000$
* $tick_{90001} = 10.001$
* $tick_{90002} = 10.002$

For spot prices less than a dollar, the precision factor decreases at every factor of 10. For example, with a $exponentAtPriceOne$ of -4:
* $tick_{-1} = 0.9999$
* $tick_{-2} = 0.9998$
* $tick_{-5001} = 0.4999$
* $tick_{-5002} = 0.4998$

With a $exponentAtPriceOne$ of -6:
* $tick_{-1} = 0.999999$
* $tick_{-2} = 0.999998$
* $tick_{-5001} = 0.994999$
* $tick_{-5002} = 0.994998$

This goes on in the negative direction until we reach a spot price of 0.000000000000000001 or in the positive direction until we reach a spot price of 100000000000000000000000000000000000000, regardless of what the exponentAtPriceOne was. The minimum spot price was chosen as this is the smallest possible number supported by the sdk.Dec type. As for the maximum spot price, the above number was based on gamm's max spot price of 340282366920938463463374607431768211455. While these numbers are not the same, the max spot price used in concentrated liquidity utilizes the same number of significant digits as gamm's max spot price and it is less than gamm's max spot price which satisfies the requirements of the initial design requirements.

#### Formulas

After we define $exponentAtPriceOne$ (this is chosen by the pool creator based on what precision they desire the asset pair to trade at), we can then calculate how many ticks must be crossed in order for k to be incremented ( $geometricExponentIncrementDistanceInTicks$ ).

$$geometricExponentIncrementDistanceInTicks = 9 * 10^{(-exponentAtPriceOne)}$$

Since we define $exponentAtPriceOne$ and utilize this as the increment starting point instead of price zero, we must multiply the result by 9 as shown above. In other words, starting at 1, it takes 9 ticks to get to the first power of 10. Then, starting at 10, it takes 9*10 ticks to get to the next power of 10, etc.

Now that we know how many ticks must be crossed in order for our $exponentAtPriceOne$ to be incremented, we can then figure out what our change in $exponentAtPriceOne$ will be based on what tick is being traded at:

$$geometricExponentDelta = ⌊ tick / geometricExponentIncrementDistanceInTicks ⌋$$

With $geometricExponentDelta$ and $exponentAtPriceOne$, we can figure out what the $exponentAtPriceOne$ value we will be at when we reach the provided tick:

$$exponentAtCurrentTick = exponentAtPriceOne + geometricExponentDelta$$

Knowing what our $exponentAtCurrentTick$ is, we must then figure out what power of 10 this $exponentAtPriceOne$ corresponds to (by what number does the price gets incremented with each new tick):

$$currentAdditiveIncrementInTicks = 10^{(exponentAtCurrentTick)}$$

Lastly, we must determine how many ticks above the current increment we are at:

$$numAdditiveTicks = tick - (geometricExponentDelta * geometricExponentIncrementDistanceInTicks)$$

With this, we can determine the price:

$$price = (10^{geometricExponentDelta}) + (numAdditiveTicks * currentAdditiveIncrementInTicks)$$

where $(10^{geometricExponentDelta})$ is the price after $geometricExponentDelta$ increments of $exponentAtPriceOne$ (which is basically the number of decrements of difference in price between two adjacent ticks by the power of 10) and 

#### Tick Spacing Example: Tick to Price

Bob sets a limit order on the USD<>BTC pool at tick 36650010. This pool's $exponentAtPriceOne$ is -6. What price did Bob set his limit order at?


$$geometricExponentIncrementDistanceInTicks = 9 * 10^{(6)} = 9000000$$

$$geometricExponentDelta = ⌊ 36650010 / 9000000 ⌋ = 4$$

$$exponentAtCurrentTick = -6 + 4 = -2$$

$$currentAdditiveIncrementInTicks = 10^{(-2)} = 0.01$$

$$numAdditiveTicks = 36650010 - (4 * 9000000) = 650010$$

$$price = (10^{4}) + (650010 * 0.01) = 16,500.10$$

Bob set his limit order at price $16,500.10

#### Tick Spacing Example: Price to Tick

Bob sets a limit order on the USD<>BTC pool at price $16,500.10. This pool's $exponentAtPriceOne$ is -6. What tick did Bob set his limit order at?


$$geometricExponentIncrementDistanceInTicks = 9 * 10^{(6)} = 9000000$$

We must loop through increasing exponents until we find the first exponent that is greater than or equal to the desired price

$$currentPrice = 1$$

$$ticksPassed = 0$$

$$currentAdditiveIncrementInTicks = 10^{(-6)} = 0.000001$$

$$maxPriceForCurrentAdditiveIncrementInTicks = geometricExponentIncrementDistanceInTicks * currentAdditiveIncrementInTicks = 9000000 * 0.000001 = 9$$

$$ticksPassed = ticksPassed + geometricExponentIncrementDistanceInTicks = 0 + 9000000 = 9000000$$

$$totalPrice = totalPrice + maxPriceForCurrentAdditiveIncrementInTicks = 1 + 9 = 10$$

10 is less than 16,500.10, so we must increase our exponent and try again

$$currentAdditiveIncrementInTicks = 10^{(-5)} = 0.00001$$

$$maxPriceForCurrentAdditiveIncrementInTicks = geometricExponentIncrementDistanceInTicks * currentAdditiveIncrementInTicks = 9000000 * 0.00001 = 90$$

$$ticksPassed = ticksPassed + geometricExponentIncrementDistanceInTicks = 9000000 + 9000000 = 18000000$$

$$totalPrice = totalPrice + maxPriceForCurrentAdditiveIncrementInTicks = 10 + 90 = 100$$

100 is less than 16,500.10, so we must increase our exponent and try again. This goes on until...

$$currentAdditiveIncrementInTicks = 10^{(-2)} = 0.01$$

$$maxPriceForCurrentAdditiveIncrementInTicks = geometricExponentIncrementDistanceInTicks * currentAdditiveIncrementInTicks = 9000000 * 0.01 = 90000$$

$$ticksPassed = ticksPassed + geometricExponentIncrementDistanceInTicks = 36000000 + 9000000 = 45000000$$

$$totalPrice = totalPrice + maxPriceForCurrentAdditiveIncrementInTicks = 10000 + 90000 = 100000$$

100000 is greater than 16,500.10. This means we must now find out how many additive tick in the currentAdditiveIncrementInTicks of -2 we must pass in order to reach 16,500.10.

$$ticksToBeFulfilledByExponentAtCurrentTick = (desiredPrice - totalPrice) / currentAdditiveIncrementInTicks = (16500.10 - 100000) / 0.01 = -8349990$$

$$tickIndex = ticksPassed + ticksToBeFulfilledByExponentAtCurrentTick = 45000000 + -8349990 = 36650010$$

Bob set his limit order at tick 36650010

#### Consequences

This decision allows us to define ticks at spot prices that users actually desire to trade on, rather than arbitrarily defining ticks at .01% distance between each other. This will also make integration with UX seamless, instead of either

a) Preventing trade at a desirable spot price or
b) Having the front end round the tick's actual price to the nearest human readable/desirable spot price

One draw back of this implementation is the requirement to create many ticks that will likely never be used. For example, in order to create ticks at 10 cent increments for spot prices greater than _$10000_, a $exponentAtPriceOne$ value of -5 must be set, requiring us to traverse ticks 1-3600000 before reaching _$10,000_. This should simply be an inconvenience and should not present any valid DOS vector for the chain.

### Scope of Concentrated Liquidity

#### Concentrated Liquidity Module

> As an engineer, I would like the concentrated liquidity logic to exist in its own module so that I can easily reason about the concentrated liquidity abstraction that is different from the existing pools.

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
	FrozenUntil     time.Time
}
```

- **Response**

On succesful response, we receive the actual amounts of each token used to create the
liquidityCreated number of shares in the given range.

```go
type MsgCreatePositionResponse struct {
	Amount0 github_com_cosmos_cosmos_sdk_types.Int
	Amount1 github_com_cosmos_cosmos_sdk_types.Int
    LiquidityCreated github_com_cosmos_cosmos_sdk_types.Dec
}
```

This message should call the `createPosition` keeper method that is introduced in the `"Liquidity Provision"` section of this document.

##### `MsgWithdrawPosition`

- **Request**

This message allows LPs to withdraw their position in a given pool and range (given by ticks), potentially in partial
amount of liquidity. It should fail if there is no position in the given tick ranges, if tick ranges are invalid,
or if attempting to withdraw an amount higher than originally provided. If an LP withdraws all of their liquidity
from a position, then the position is deleted from state. However, the fee accumulators associated with the position
are still retained until a user claims them manually.

```go
type MsgWithdrawPosition struct {
	PoolId          uint64
	Sender          string
	LowerTick       int64
	UpperTick       int64
	LiquidityAmount github_com_cosmos_cosmos_sdk_types.Dec
	FrozenUntil     time.Time
}
```

- **Response**

On successful response, we receive the amounts of each token withdrawn
for the provided share liquidity amount.

```go
type MsgWithdrawPositionResponse struct {
	Amount0 github_com_cosmos_cosmos_sdk_types.Int
	Amount1 github_com_cosmos_cosmos_sdk_types.Int
}
```

This message should call the `withdrawPosition` keeper method that is introduced in the `"Liquidity Provision"` section of this document.

##### `MsgCreatePool`

This message is responsible for creating a concentrated-liquidity pool.
It propagates the execution flow to the `x/poolmanager` module for pool id
management and for routing swaps.

```go
type MsgCreateConcentratedPool struct {
	Sender                    string
	Denom0                    string
	Denom1                    string
	TickSpacing               uint64
	PrecisionFactorAtPriceOne github_com_cosmos_cosmos_sdk_types.Int
	SwapFee                   github_com_cosmos_cosmos_sdk_types.Dec
}
```

- **Response**

On successful response, the pool id is returned.

```go
type MsgCreateConcentratedPoolResponse struct {
	PoolID uint64
}
```

##### `MsgCollectFees`

This message allows collecting fee from a position that is defined by the given
pool id, sender's address, lower tick and upper tick.

The fee collection is discussed in more detail in the "Fees" section of this document.

```go
type MsgCollectFees struct {
	PoolId    uint64
	Sender    string
	LowerTick int64
	UpperTick int64
}
```

- **Response**

On successful response, the collected tokens are returned.
The sender should also see their balance increase by the returned
amounts.

```go
type MsgCollectFeesResponse struct {
	CollectedFees []types.Coin
}
```

#### Relationship to Pool Manager Module

##### Pool Creation

As previously mentioned, the `x/poolmanager` is responsible for creating the
pool upon being called from the `x/concentrated-liquidity` module's message server.

It does so to store the mapping from pool id to concentrated-liquidity module so
that it knows where to route swaps.

Upon successful pool creation and pool id assignment, the `x/poolmanager` module
returns the execution to `x/concentrated-liquidity` module by calling `InitializePool`
on the `x/concentrated-liquidity` keeper.

The `InitializePool` method is responsible for doing concentrated-liquidity specific
initialization and storing the pool in state.

Note, that `InitializePool` is a method defined on the `SwapI` interface that is
implemented by all swap modules. For example, `x/gamm` also implements it so that
`x/pool-manager` can route pool initialization there as well.

##### Swaps

We rely on the swap messages located in `x/poolmanager`:
- `MsgSwapExactAmountIn`
- `MsgSwapExactAmountOut`

The `x/poolmanager` received the swap messages and, as long as the swap's pool id
is associated with the `concentrated-liquidity` pool, the swap is routed
into the relevant module. The routing is done via the mapping from state that was
discussed in the "Pool Creation" section.

#### Liquidity Provision

> As an LP, I want to provide liquidity in ranges so that I can achieve greater capital efficiency

This is a basic function that should allow LPs to provide liquidity in specific ranges
to a pool.

A pool's liquidity is consisted of two assets: asset0 and asset1. In all pools, asset0 will be the lexicographically smaller of the two assets. At the current tick, the bucket at this tick consists of a mix of both asset0 and asset1 and is called the virtual liquidity of the pool (or "L" for short). Any positions set below the current price are consisted solely of asset0 while positions above the current price only contain asset1.

##### Adding Liquidity

We can either provide liquidity above or below the current price, which would act as a range order, or decide to provide liquidity at the current price. 

As declared in the API for `createPosition`, users provide the upper and lower tick to denote the range they want to provide the liquidity in. The users are also prompted to provide the amount of token0 and token1 they desire to receive. The liquidity that needs to be provided for the given token0 and token1 amounts would be then calculated by the following methods: 

Liquidity needed for token0:
$$L = \frac{\Delta x \sqrt{P_u} \sqrt{P_l}}{\sqrt{P_u} - \sqrt{P_l}}$$

Liquidity needed for token1:
$$L = \frac{\Delta y}{\sqrt{P_u}-\sqrt{P_l}}$$

Then, we pick the smallest of the two values for choosing the final `L`. The reason we do that is because the new liquidity must be proportional
to the old one. By choosing the smaller value, we distribute the liqudity evenly between the two tokens. In the future steps, we will re-calculate the amount of token0 and token1 as a result the one that had higher liquidity will end up smaller than originally given by the user.

Note that the liquidity used here does not represent an amount of a specific token, but the liquidity of the pool itself, represented in `sdk.Dec`.

Using the provided liquidity, now we calculate the delta amount of both token0 and token1, using the following equations, where L is the liquidity calculated above:

$$\Delta x = \frac{L(\sqrt{p(i_u)} - \sqrt{p(i_c)})}{\sqrt{p(i_u)}\sqrt{p(i_c)}}$$
$$\Delta y = L(\sqrt{p(i_c)} - \sqrt{p(i_l)})$$

Again, by recalculating the delta amount of both tokens, we make sure that the new liquidity is proportional to the old one and the excess amount of the
token that originally computed a larger liquidity is given back to the user.

The delta X and the delta Y are the actual amounts of tokens joined for the requested position. 

Given the parameters needed for calculating the tokens needed for creating a position for a given tick, the API in the keeper layer would look like the following:

```go
ctx sdk.Context, poolId uint64, owner sdk.AccAddress, amount0Desired, amount1Desired, amount0Min, amount1Min sdk.Int, lowerTick, upperTick int64, frozenUntil time.Time
func createPosition(
    ctx sdk.Context,
    poolId uint64,
    owner sdk.AccAddress,
    amount0Desired,
    amount1Desired,
    amount0Min,
    amount1Min sdk.Int
    lowerTick,
    upperTick int64,
    frozenUntil time.Time) (amount0, amount1 sdk.Int, sdk.Dec, error) {
        ...
}
```

##### Removing Liquidity

Removing liquidity is achieved via method `withdrawPosition` which is the inverse of previously discussed `createPosition`. In fact,
the two methods share the same underlying logic, having the only difference being the sign of the liquidity. Plus signifying addition
while minus signifying subtraction.

Withdraw position also takes an additional parameter which represents the liqudity a user wants to remove. It must be less than or
equal to the available liquidity in the position to be successful.

```go
func (k Keeper) withdrawPosition(
    ctx sdk.Context,
    poolId uint64,
    owner sdk.AccAddress,
    lowerTick,
    upperTick int64,
    frozenUntil time.Time,
    requestedLiquidityAmountToWithdraw sdk.Dec) (amtDenom0, amtDenom1 sdk.Int, err error) {
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
Thus, we need two accumulators for each token.

TODO: explain the `accum` package and how it is used in CL

Temporally, these fee accumulators are accessed together from state most of the time. Therefore, we define a data structure for storing the fees of each token in the pool.

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

##### Swap Step Fees

We have a notion of `swapState.amountSpecifiedRemaining` which  is the amount of token in
remaining over all swap steps.

After performing the current swap step, the following cases are possible:

1. All amount remaining is consumed

In that case, the fee is equal to the difference between the original amount remaining
and the one actually consumed. The difference between them is the fee.

```go
feeChargeTotal = amountSpecifiedRemaining.Sub(amountIn) 
```

2. Did not consume amount remaining in-full.

The fee is charged on the amount actually consumed during a swap step.

```go
feeChargeTotal = amountIn.Mul(swapFee) 
```

3. Price impact protection makes it exit before consuming all amount remaining.

The fee is charged on the amount in actually consumed before price impact
protection got trigerred.

```go
feeChargeTotal = amountIn.Mul(swapFee) 
```

#### Liquidity Rewards

TODO

##### State

- global (per-pool)

- per-tick

- per-position

#### Placeholder

### Terminology

We will use the following terms throughout the document:

- `Virtual Reserves` - TODO

- `Real Reserves` - TODO

- `Tick` - TODO

- `FullPosition` - A single user's liquidity in a single pool spread out between two ticks with a frozenUntil timestamp. Unlike Position, FullPosition can
only describe a single instance of liquidity. If a user adds liquidity to the same pool between the same two ticks, but with a different frozenUntil timestamp, then it will be a different FullPosition.

- `Position` - A single user's liquidity in a single pool spread out between two ticks. Unlike FullPosition, position does not
take into consideration the frozenUntil timestamp. Therefore, a position can describe multiple instances of liquidity
between the same two ticks in the same pool, but with different frozenUntil timestamps.

- `Range` - TODO

### External Sources

- [Uniswap V3 Whitepaper](https://uniswap.org/whitepaper-v3.pdf)
- [Technical Note on Liquidity Math](https://atiselsts.github.io/pdfs/uniswap-v3-liquidity-math.pdf)
