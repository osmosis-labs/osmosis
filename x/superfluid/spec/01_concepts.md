<!--
order: 1
-->

# Concepts

## SyntheticLockups

SyntheticLockups are synthetica of PeriodLocks, but different in the sense that they store suffix, which is a combination of bonding/unbonding status + validator address. This is mainly used to track whether an individual lock that has been superfluid staked has an bonding status or a unbonding status from the staking delegations.

## Intermediary Account

Intermediary Accounts establishes the connections between the superfluid staked locks and delegations to the validator. Intermediary accounts exists for every denom + validator combination, so that it would group locks with the same denom + validator selection. Superfluid staking a lock would mint equivalent amount of OSMO of the lock and send it to the intermediary account and the intermediarry accounts would be delegating to the specified validator.

## Intermediary Account Connection

Intermediary Accounts Connection serves the role of tracking the locks that an Intermediary Account is dedicated to.

---

Lets fill out this spec, and think about the entire state machine.

Its a state machine which has inputs:

- Epoch
- Slashes
- Gov props to add/remove superfluid assets
- ??? Staking weight changes? (Validators going inactive)

It has state:

- Superfluid assets
- Intermediary Accounts
- Dedicated Gauges
- Synthetic Locks
- Historical osmo equivalent multipliers?
