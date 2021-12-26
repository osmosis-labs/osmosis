# Generalized F1 for lockups

## Goal

We want the goal of being able to distribute LP rewards every epoch, without the distribution having to iterate over every single user.
Every pool allows users to bond LP shares, for a time T.
There are different rewards given to users depending on how long they have bonded their LP shares.
For instance, if they have bonded for 7 days, they get more rewards than if they bond for 1 day.
Further, we must prevent the chain and users from being exposed to DOS vectors in this process.

## Desired general method of doing so

The F1 Fee Distribution idea ( https://drops.dagstuhl.de/opus/volltexte/2020/11974/pdf/OASIcs-Tokenomics-2019-10.pdf ) is basically, if rewards are to be linearly distributed to folks by share-ownership, then there is a straightforward solution using 'accumulators' to store the rewards.

The rough idea is as follows:
We want to track the rewards a user who starts owning 5 shares of something at time A, and withdraws rewards for those 5 shares at a later time B.
The way we do this is by storing an accumulator of all the rewards a single share would have gotten at time = 0, until now.
So when the user starts owning 5 shares of that something at time A, we create a state record to persist the accumulator for all rewards' a single share gets from t=0 to t=A. (Called `accum_A`)
When they go to withdraw at t=B, we read the accumulator's value at time t=B. (Called `accum_B`)
We compute the rewards per share then as `Rewards_per_share = accum_B - accum_A`.
Therefore the total rewards here is `total_rewards = Rewards_per_share * num_shares`.

This has been built out before in the cosmos SDK, check out the distribution module!

https://github.com/cosmos/cosmos-sdk/tree/master/x/distribution/spec

## More details on design

In our case, we want a similar architecture to whats implemented in staking, but with some optimizations.

#### Fewer Periods

We only need to update total share amounts right before each epoch begins, in the `BeforeEpochStart` hook. (each update is called a `period` in the cosmos SDK & F1 spec)
This is as opposed to staking, where it must be done on every bond and unbond.
This is because LP rewards only get sent at the epoch boundary, so the difference in accumulator values between each intermittent step is 0.

This also lowers the priority of implementing the refcount optimization present in the cosmos SDK, if that is a bottleneck here.

#### How to handle variable lockup lengths

Every bond should store when rewards for it were last withdrawn.

Using the [accumulation store](https://github.com/osmosis-labs/osmosis/blob/main/store/README.md) implemented, we efficiently can know how many tokens have bonded for a duration greater than length `T`.

To avoid DOS issues / simplify lots of complexity, we restrict how many different durations users can be rewarded for locking up for. E.g. your only conditions you can reward for are, e.g. `bonding >= 1 day, bonding >= 7 day, bonding >= 14 day, bonding >= 1 month`. (And no in between durations) 
The supported durations are found in [this](TBD, get link to code) state variable.

Per denomination we are rewarding lock-ups for, we store one accumulator value per bonding duration. So there is one accumulator for each (lockable denom, supported reward duration) pair.

Then at epoch time, when adding rewards for `1 day` bonds, the code finds the number of tokens that are bonded for > 1 day, by looking at the accumulation store.
Then we just increment the accumulator for 1 day bond rewards of that denomination accordingly. (Increment it by `(LP rewards for > 1 day) / (number of tokens bonded for > 1 day)`)

These rewards for `> 1 day` duration only get added to `> 1 day` accumulator, not the other accumulators. People who have locked longer get all the rewards using the method described in the next section, claiming rewards.

#### Claiming rewards

This should be done as in staking. In state, there is a record associated with what accumulator value that my unclaimed rewards begin at.

If I have a lockup for say `10 days`, last claimed rewards at t=A, then I get total rewards at time `t=B` according to:

`rewards_per_share = (accum_1_day_B - accum_1_day_A) + (accum_7_day_B - accum_7_day_A)`

`total_rewards = (my number of tokens locked) * rewards_per_share`

In practice, folks may have different numbers of tokens locked for `>1 day` and `>7 day` which would then be handled by taking a linear combination of the terms.
We can enforce that for a single pool's LP rewards, all your lockups must get their rewards withdrawn together though.
And every further bonding must withdraw existing rewards

#### Handling unbonding LP shares

If I am unbonding a 14 day LP share bond, I should still be getting the rewards for a 7 day LP shares for 6 days, until I am no longer bonded for a full 7 days remaining.

This means we need to define a 'symmetry' set for every unbonding duration.
We should truncate this to a precision of the epoch you complete your unbond during.
So what we do is when you start unbonding, we create an accumulation store for folks whose unbond is ending during {epoch N}.
Then when we distribute rewards at an epoch boundary, we can uniformly treat everyone whose unbond ends within that epoch the same way, efficiently.

Also, when beginning to unbond, we do as we do in staking, and withdraw rewards for those tokens immediately for simplicity.

### DOS vectors to be concerned about

- too many options for what bond durations to reward for
  - We handle this by limiting the number of different bond lengths you can get rewarded.
