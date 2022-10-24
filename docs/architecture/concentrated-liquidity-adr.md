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

Providing, burning liquidity, and swapping functions are to be tracked in their own stories.

#### Swap Router Module

> As a user, I would like to have a unified entrypoint for my swaps regardless of the underlying pool implementation so that I don't need to reason about API complexity

> As a user, I would like the pool management to be unified so that I don't have to reason about additional complexity stemming from divergent pool sources.

With the new `concentrated-liquidity` module, we now have a new entrypoint for swaps that is
the same with the existing `gamm` module.

To avoid fragmenting swap entrypoints and duplicating boilerplate logic, we would like to define
a new `swap-router` module. For now, its only purpose is to receive swap messages and propagate them
either to the `gamm` or `concentrated-liquidity` modules.

Therefore, we should move the existing `gamm` swap messages and tests to the new `swap-router` module, connecting to the `swap-router` keeper that simply propagates swaps to `gamm` or `concentrated-liquidity` modules.

The messages to move are:
- `MsgSwapExactAmountIn`
- `MsgSwapExactAmountOut`

TODO: figure out routing logic:
- should we use an id?
- should have a new pool type field?

#### Liquidity Provision

> As an LP, I want to provide liquidity in ranges so that I can achieve greater capital efficiency

This a basic function that should allow LPs to provide liquidity in specific ranges
to a pool. We will disucss each of these functions in more detail in the following sections.

##### Msg

```go
func Mint(
    ctx sdk.Context,
    poolId uint64,
    owner sdk.AccAddress,
    liquidityIn sdk.Int,
    lowerTick,
    upperTick int64) {
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

#### Range Orders

> As a trader, I want to be able to execute ranger orders so that I have better control of the price at which I trade

TODO

#### Fees

> As a an LP, I want to earn fees on my capital so that I am incentivized to participate in the market making actively.

TODO

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

####

### Terminology

We will use the following terms throughout the document:

- `Virtual Reserves` - TODO

- `Real Reserves` - TODO

- `Tick` - TODO

- `Position` - TODO

- `Range` - TODO

