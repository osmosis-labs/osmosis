<!--
order: 2
-->

# State

## SuperfluidAsset

A superfluid asset is ...

It can only be updated by governance proposals. We validate at proposal creation time that the denom + pool exists.
(Are we going to ignore edge cases around a reference pool getting deleted it)

## Intermediary Accounts

Lots of questions to be answered here

## Synthetic Lockups created

At the moment, one lock can only be fully bonded to one validator.

## Osmo Equivalent Multipliers

The Osmo Equivalent Multiplier for an asset is the multiplier it has for its value relative to OSMO.

Different types of assets can have different functions for calculating their multiplier.  We currently support two asset types.

1. Native Token

The multiplier for OSMO is alway 1.

2. Gamm LP Shares

Currently we use the spot price for an asset based on a designated osmo-basepair pool of an asset. 
The multiplier is set once per epoch, at the beginning of the epoch.
In the future, we will switch this out to use a TWAP instead.
