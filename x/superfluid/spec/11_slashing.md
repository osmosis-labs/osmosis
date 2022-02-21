<!--
order: 11
-->

# Slashing

We first get a hook from the staking module, marking that a validator is about to be slashed at a slashFactor of `f`, for an infraction at height `h`.

The staking module handles slashing every delegation to that validator, which will handle slashing the delegation from every intermediary account.
However, it is up to the superfluid module to then:

* Slash every constituent superfluid staking position for this validator.
* Slash every unbonding superfluid staking position to this validator.

We do this by:

* Collect all intermediate accounts to this validator
* For each IA, iterate over every lock to the underlying native denom.
* If the lock has a synthetic lockup, it gets slashed.
* The slash works by calculating the amount of tokens to slash.
* It removes these from the underlying lock and the synthetic lock.
* These coins are moved to the community pool.

## Slashing nuances

* Slashed tokens go to the community pool, rather than being burned as in staking.
* We slash every unbonding, rather than just unbondings that started after the infraction height.
* We can "overslash" relative to the staking module. (For a slash factor of 5%, the staking module can often burn <5% of active delegation, but superfluid will always slash 5%)

We slash every unbonding, purely because lockup module tracks things by unbonding start time, whereas staking/slashing tracks things by height we begin unbonding at.
Thus we get a problem that we cannot convert between these cleanly.
Really there should be a storage of all historical block height <> block times for everything in the unbonding period, but this is not considered a near-term problem.

### Correcting overslashing

The overslashing possibility stems from a problem in the SDKs slashing module, that really is a bug there, and superfluid is doing the correct thing.
<https://github.com/cosmos/cosmos-sdk/issues/1440>

Basically, slashes to unbondings and redelegations can lower the amount that gets slashed from live delegations in the staking module today.

It turns out this edge case, where superfluid's intermediate account can have more delegation than expected from its underlying collateral, is already safely handled by the Superfluid refreshing logic.

The refreshing logic checks the total amount of tokens in locks to this denom (Reading from the lockup accumulation store), calculates how many osmo thats worth at the epochs new osmo worth for that asset, and then uses that.
Thus this safely handles this edge case, as it uses the new 'live' lockup amount.
