<!--
order: 2
-->

# State

## Superfluid Asset

A superfluid asset is a alternative asset (non-OSMO) that is allowed by governance to be used for staking.

It can only be updated by governance proposals. We validate at proposal creation time that the denom + pool exists.
(Are we going to ignore edge cases around a reference pool getting deleted it)

## Intermediary Accounts

Lots of questions to be answered here

## Dedicated Gauges

Each intermediary account has has dedicated gauge where it sends the delegation rewards to.
Gauges are distributing the rewards to end users at the end of the epoch.

## Synthetic Lockups created

At the moment, one lock can only be fully bonded to one validator.

## Osmo Equivalent Multipliers

The Osmo Equivalent Multiplier for an asset is the multiplier it has for its value relative to OSMO.

Different types of assets can have different functions for calculating their multiplier. We currently support two asset types.

1. Native Token

The multiplier for OSMO is alway 1.

2. Gamm LP Shares

Currently we use the spot price for an asset based on a designated osmo-basepair pool of an asset.
The multiplier is set once per epoch, at the beginning of the epoch.
In the future, we will switch this out to use a TWAP instead.

# State changes

The state of superfluid module state modifiers are classified into below categories.

- [Proposals](07_proposals.md)
- [Messages](03_messages.md)
- [Epoch](04_epoch.md)
- [Hooks](06_hooks.md)
