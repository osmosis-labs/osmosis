# ADR 001: Mint Truncation Refactor

## Changelog

* 22-08-2022: Initial Draft

## Status

Draft: <https://github.com/osmosis-labs/osmosis/pull/2342>

## Abstract

This ADR focuses on refactoring `x/mint` module to mitigate the discrepancies
between the actual and the projected inflation amounts.

Currently, we under-mint due to truncations. In the first year of operations,
this happens to be approximately 2600 OSMO. As a result, we cannot reach the
projected OSMO supply one-to-one.

Additionally, some of the constraints have made it difficult to refactor the
`x/mint` module. Specifically, the developer vesting provisions follow a
distinct distribution logic from the rest of the provisions. While that is
the case, these provisions are tightly coupled together, causing the
utilization of unsafe workarounds such as over-minting and later burning
developer vesting provisions. This ADR addresses these issues by decoupling
the two kinds of provisions and letting them use separate distribution logic.

## Context

### Truncations

Ref: <https://github.com/osmosis-labs/osmosis/issues/1917>

Ultimately, the major sources of the truncation issues are some SDK interfaces
such as in [`x/bank`][1] and [`x/distribution`][2].
These interfaces operate on `sdk.Coin` that uses `sdk.Int` for amounts. To use
these interfaces, we always round down to the nearest integer by
[truncating decimal provisions][3].
While we operate on amounts with the precision of 6 decimals by using
exponentiation and assuming that 1 `sdk.Int` is equal to `1 / 10^6` OSMO, this
still does not allow us to observe enough accuracy. As a result, it is possible
to under-mint. `sdk.Dec` has a precision of `18` decimals. By using it in
conjunction with the above `1 / 10^6` downscaling allows us to achieve a
precision of `6 + 18 = 24` decimals. According to tests, this precision is
sufficient to accurately represent the projected amounts and achieve the
expected supply after [30 years of operations][4].

The expected amounts have been estimated in Python by using the
following formulas:

* Total Provisions `P(n)` at yeat `n`
$$P(n)=EpochsPerPeriod*InitialRewardsPerEpoch*((1-ReductionFactor^{n+1})/(1-ReductionFactor))$$

* Total expected supply `S`
$$S=InitialSupply+EpochsPerPeriod*(InitialRewardsPerEpoch/(1-ReductionFactor))$$

Lastly, developer reward receivers suffer the most because the large source of
truncations is identified to occur in [the calculation of the proportions for
each developer account][5]:
<https://github.com/osmosis-labs/osmosis/blob/4176b287d48338870bfda3029bfa20a6e45ac126/x/mint/keeper/hooks_test.go#L601-L602>

### Additional Limitations

Next, we present the limitations that need to be eliminated to mitigate the
above truncation issues.

#### Coupled Epoch Provisions of Different Kind

Ref: <https://github.com/osmosis-labs/osmosis/issues/2025>

Our developer vesting provisions have been coupled together with the rest of
the provisions (futher referred to as "inflation provisions") despite having
a distinct distribution logic. The divergence is summarized next:

1. Developer vesting provisions are [pre-minted][6] at genesis to the developer
vesting module account.
2. Since they are pre-minted, we do not need to be minting them every epoch
contrary to the inflation provisions.
   * This has caused several issues such as having to [over-mint][7]
   by the developer vesting rewards and then [burn them later][8].
3. The developer vesting provisions are [distributed from the developer vesting
module account][9] while other rewards are [distributed from the mint module account][10].
4. We use supply offsets to [offset the unvested developer provisions][11]
since we have pre-minted the full amount at genesis. The offsets are unrelated
to the inflation provisions.

The above differences portray the distinct handling of the developer vesting
provisions. Still, their handling is highly coupled to the regular provisions,
leading to increased complexity.

#### Complicated `AfterEpochEnd` Hook

Ref: <https://github.com/osmosis-labs/osmosis/issues/1919>

Currently, `x/mint` `AfterEpochEnd` hook is focused on several goals such as:

* Determining when to start or update the provisions.

* Determining if the current epoch is the reduction epoch.

* Handling the reductions.

* Minting and distributing provisions.

As a result, it is difficult to reason about it, assert its correctness and
make new changes.

All minting and distribution logic can be encapsulated into a separate function
for better testability and readability. This encapsulation also helps to
achieve an increased separation of concerns.

## Decisions

### Decision 1

#### Summary

We will **minimize truncations and use decimals for estimating distributions**.
Specifically, the [`getProportions`][12] function will take a decimal value and
return decimal result so that we can use the (non-truncated) value with
increased precision for further calculating each developer's reward and
inflation provisions. Additionally, functions that handle distributions logic
such as [`distributeDeveloperVestingProvisions`][13] will now take
`sdk.DecCoin` as opposed to `sdk.Coin` for the same reason of having to operate
on decimals with increased precision.

#### Conseqeunces

##### Positive

The truncations due to integer interfaces within the `x/mint` module are
eliminated completely. All remaining truncations are due to dependencies on the
`x/bank` and `x/distribution` modules.

##### Negative

Divergence from the original implementation of the `x/mint` module as well as
larger diff, making the review more difficult.

### Decision 2

#### Summary

We will **add 2  decimal store indexes**:

* for persisting truncation delta resulting from the mint module account across epochs
<https://github.com/osmosis-labs/osmosis/blob/0b843fcae194eb9439c3dc5fe879c47173406047/x/mint/types/keys.go#L12-L21>

* for persisting truncation delta resulting from the developer rewards module
  account across epochs
<https://github.com/osmosis-labs/osmosis/blob/0b843fcae194eb9439c3dc5fe879c47173406047/x/mint/types/keys.go#L23-L31>

#### Consequences

##### Positive

This is helpful for accounting for truncations and distributing them eventually,
not necessarily in the same epoch.

For example, assume that for some number of epochs our expected provisions are
100.6 and the actual amount distributed is 100 every epoch due to truncations.
Then, at epoch 1, we have a delta of 0.6. 0.6 cannot be distributed because it
is not an integer. So we persist it until the next epoch. Then, at epoch 2, we
get a delta of 1.2 (0.6 from epoch 1 and 0.6 from epoch 2). Now, 1 can be
distributed and 0.2 gets persisted until the next epoch.

##### Negative

Added complexity from handing 2 additional store indexes. It is mitigated by
better abstractions and extensive documentation though still present.

### Decision 3

#### Summary

We will **decouple the developer vesting provisions from the inflation provisions**.

The [**Draft Implementation**](https://github.com/osmosis-labs/osmosis/pull/2342)
makes the distinction between the developer provisions and the inflation provisions
clearer by:

* [Distinctly splitting the two provisions in minter](https://github.com/osmosis-labs/osmosis/blob/0b843fcae194eb9439c3dc5fe879c47173406047/x/mint/types/minter.go#L54-L63)

* Decoupling and distinctly handling each provision type separately:

   * [inflation provisions](https://github.com/osmosis-labs/osmosis/blob/0b843fcae194eb9439c3dc5fe879c47173406047/x/mint/keeper/keeper.go#L167)

   * [dev reward provisions](https://github.com/osmosis-labs/osmosis/blob/0b843fcae194eb9439c3dc5fe879c47173406047/x/mint/keeper/keeper.go#L283)

#### Consequences

##### Positive

The above decoupling makes the abstractions clearer, reduces complexity, and
fixes [#2025][14]. This change also makes it more intuitive to apply
the [Decision 2](#decision_2) above since the 2 truncation store indexes are
split into 2 to separate the distribution from separate module accounts.

##### Negative

This is a large change to the core logic of the `x/mint` module, requiring more
thorough testing and quality assurance.

### Decision 4

#### Summary

We will **encapsulate the minting and distribution logic from `AfterEpochEnd` hook**
into a separate function.

In the [**Draft Implementation**][15], the logic for distributing all epoch
provisions in the `AfterEpochEnd` hook has been moved to the
[distributeEpochProvisions][16] function.

It handles distributing both [inflation provisions][17]
and [developer vesting provisions][18].

#### Consequences

##### Positive

This change allows for better encapsulation of all distribution logic and for
the ability to more thoroughly unit test it.

The enhanced encapsulation in turn leads to a more modular mint keeper where
each method strives to focus on a single concern. 

## Backwards Compatibility

This change is not backward compatible with any of the previous Osmosis
versions and requires a major upgrade to be deployed.

## Further Discussions

### Distributing Truncation Delta

For any truncation deltas occurring at epochs before the proposed
implementation is live, it will have to be manually estimated and minted/burned
in the upgrade handler.

The implementation proposed in this ADR is independent of distributing the old
truncation deltas. It only ensures that there are no more discrepancies after
the proposed proof-of-concept is deployed.

The work for estimating and isolating the old truncation deltas has been
performed in:

* <https://github.com/osmosis-labs/osmosis/pull/1874>

* <https://github.com/osmosis-labs/osmosis/tree/roman/mint-rounding-year2-isolation>

As a result, as long as the next upgrade height and epoch are known, the old
truncation deltas up until the last epoch before the upgrade can be estimated
and applied in the upgrade handler.

## References

* Draft POC: <https://github.com/osmosis-labs/osmosis/pull/2342>
* Projected Inflation: <https://medium.com/osmosis/osmo-token-distribution-ae27ea2bb4db>
* Isolating sources of the truncations:
   * Logically: <https://github.com/osmosis-labs/osmosis/pull/1874>
   * By module account: <https://github.com/osmosis-labs/osmosis/tree/roman/mint-rounding-year2-isolation>
* Truncatins Issue: <https://github.com/osmosis-labs/osmosis/issues/1919>
* Coupling Developer Vesting with Inflation Provisions: <https://github.com/osmosis-labs/osmosis/issues/2025>
* Refactoring `x/mint` `AfterEpochEnd` Hook: <https://github.com/osmosis-labs/osmosis/issues/1919>

[1]:https://github.com/osmosis-labs/osmosis/blob/86bdbebd3cffc16586d0d0c25f751321436d7a44/x/mint/keeper/keeper.go#L266-L267
[2]:https://github.com/osmosis-labs/osmosis/blob/86bdbebd3cffc16586d0d0c25f751321436d7a44/x/mint/keeper/keeper.go#L255-L256
[3]:https://github.com/osmosis-labs/osmosis/blob/86bdbebd3cffc16586d0d0c25f751321436d7a44/x/mint/keeper/keeper.go#L290
[4]:https://github.com/osmosis-labs/osmosis/blob/724d2cacb38596919c29dd3f9173c1ce0c58804d/x/mint/keeper/hooks_test.go#L453
[5]:https://github.com/osmosis-labs/osmosis/blob/4176b287d48338870bfda3029bfa20a6e45ac126/x/mint/keeper/keeper.go#L265
[6]:https://github.com/osmosis-labs/osmosis/blob/86bdbebd3cffc16586d0d0c25f751321436d7a44/x/mint/keeper/genesis.go#L30
[7]:https://github.com/osmosis-labs/osmosis/blob/86bdbebd3cffc16586d0d0c25f751321436d7a44/x/mint/keeper/hooks.go#L54-L55
[8]:https://github.com/osmosis-labs/osmosis/blob/86bdbebd3cffc16586d0d0c25f751321436d7a44/x/mint/keeper/keeper.go#L230
[9]:https://github.com/osmosis-labs/osmosis/blob/86bdbebd3cffc16586d0d0c25f751321436d7a44/x/mint/keeper/keeper.go#L266-L267
[10]:https://github.com/osmosis-labs/osmosis/blob/86bdbebd3cffc16586d0d0c25f751321436d7a44/x/mint/keeper/keeper.go#L194
[11]:https://github.com/osmosis-labs/osmosis/blob/86bdbebd3cffc16586d0d0c25f751321436d7a44/x/mint/keeper/keeper.go#L275-L277
[12]:https://github.com/osmosis-labs/osmosis/blob/724d2cacb38596919c29dd3f9173c1ce0c58804d/x/mint/keeper/keeper.go#L477
[13]:https://github.com/osmosis-labs/osmosis/blob/724d2cacb38596919c29dd3f9173c1ce0c58804d/x/mint/keeper/keeper.go#L286
[14]:https://github.com/osmosis-labs/osmosis/issues/2025
[15]:https://github.com/osmosis-labs/osmosis/pull/2342
[16]:https://github.com/osmosis-labs/osmosis/blob/0b843fcae194eb9439c3dc5fe879c47173406047/x/mint/keeper/hooks.go#L49
[17]:https://github.com/osmosis-labs/osmosis/blob/0b843fcae194eb9439c3dc5fe879c47173406047/x/mint/keeper/keeper.go#L147
[18]:https://github.com/osmosis-labs/osmosis/blob/0b843fcae194eb9439c3dc5fe879c47173406047/x/mint/keeper/keeper.go#L153
