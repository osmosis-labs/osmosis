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
to a pool. 

A pool's liquidity is consisted of two assets: asset0 and asset1. In all pools, asset0 will be the lexicographically smaller of the two assets. At the current price tick, the bucket at this tick consists of a mix of both asset0 and asset1 and is called the virtual liquidity of the pool (or "L" for short). Any positions set below the current price are consisted solely of asset0 while positions above the current price only contain asset1.

Therefore in `Mint`, we can either provide liquidity above or below the current price, which would act as limit orders, or decide to provide liquidity at current price. 

As declared in the API for mint, users provide the upper and lower tick to denote the range they want to provide the liquidity for. The users are also prompted to provide the amount of token0 and token1 they desire to receive. The liquidity that needs to be provided for the token0 and token1 amount provided would be then calculated by the following methods: 

Liquidity needed for token0:
$$L = \frac{\Delta x \sqrt{P_u} \sqrt{P_l}}{\sqrt{P_u} - \sqrt{P_l}}$$

Liquidity needed for token1:
$$L = \frac{\Delta y}{\sqrt{P_u}-\sqrt{P_l}}$$

With the larger liquidity including the smaller liquidity, we take the smaller liquidity calculated for both token0 and token1 and use that as the liquidity throughout the rest of the joining process. Note that the liquidity used here does not represent an amount of a specific token, but the liquidity of the pool itself, represented in sdk.Int.

Using the provided liquidity, now we calculate the delta amount of both token0 and token1, using the following equations, where L is the liquidity calculated above:

$$\Delta x = \frac{L(\sqrt{p(i_u)} - \sqrt{p(i_c)})}{\sqrt{p(i_u)}\sqrt{p(i_c)}}$$
$$\Delta y = L(\sqrt{p(i_c)} - \sqrt{p(i_l)})$$


The deltaX and the deltaY would be the actual amount of tokens joined for the requested position. 

Given the parameters needed for calculating the tokens needed for creating a position for a given tick, the API in the msg server layer would look like the following:

```go
func CreatePosition(
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

TODO

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

