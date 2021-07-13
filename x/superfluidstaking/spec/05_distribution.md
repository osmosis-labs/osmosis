# Distribution

In this stage, we start from the point of view that trustworthy table about `lp_token_price` and `risk_weight` values are set or actively updated.

The problem here is to calculate the split rating between native OSMO stakers and superfluid stakers.

## Which module codebase to update

1. If we need to manage superfluid staking associated to validators, we need to add hooks that runs before when validator's total reward is being increased and split the reward for `NATIVE_OSMO_STAKERS` and `SUPERFLUID_OSMO_STAKERS`.
2. If we need to manage superfluid staking without any validators, it's pretty easy. Just take out percentage of collected fees from `fee` pool to gauges before moving to `distribution` module.

**Note:** to make code organized for superfluid staking, distribution and staking modules will need to be modified as small as possible and superfluid staking module will have functions that will run when hooks are triggered.

## Split between native OSMO stakers and superfluid stakers (Valid only superfluid stakers are not split by validators if split by validators, each user should get different reward for which validator they delegate on)

Calculate total sum of `lp_token_price` * (1 - `risk_weight`) * `lp_bonded_amount` and call it `SUPERFLUID_STAKING_OSMO_AMOUNT`.

Here, `lp_bonded_amount` is refering to the amount of tokens bonded to `lockup` module which is locked up longer than `UNBONDING` time. And it can be calculated in a time-efficiency manner using accumulation store.

Let's say `TOTAL_STAKING_OSMO_AMOUNT` = `SUPERFLUID_STAKING_OSMO_AMOUNT` * `superfluid_multiplier` + `NATIVE_STAKING_OSMO_AMOUNT`.

Native OSMO stakers will get
`TOTAL_REWARD` * `NATIVE_STAKING_OSMO_AMOUNT` / `TOTAL_STAKING_OSMO_AMOUNT`

Each lp token superfluid stakers will get
`TOTAL_REWARD` * `lp_token_price` * (1 - `risk_weight`) * `each_lp_bonded_amount` / `TOTAL_STAKING_OSMO_AMOUNT`

**Note** native OSMO stakers and superfluid stakers are different types of staking and `superfluid_multiplier` param could be adjusted control the split rate by governance.

## How to distribute rewards (Valid only superfluid stakers are not split by validators if split by validators, each user should get different reward for which validator they delegate on)

Rewards will be distributed superfluid stakers via a perpetual gauge.

A perpetual gauge will be created per LP token for `UNBONDING` unbonding period lockups. `UNBONDING` could be 3 weeks so that it can be same as native `OSMO` unbonding period.

The superfluid staking reward amount will be added to the gauge at the end of epochs and it will be automatically distributed by `incentives` module.
