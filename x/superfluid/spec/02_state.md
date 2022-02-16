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

## TWAP

** Clarify what we mean by TWAP. Namely the time-weighted average of the osmo backing of an LP share.

The TWAP for an epoch is set at the beginning of the epoch. We refer to the TWAP for use in epoch N as the 24hr TWAP for the entire prior epoch (N-1). This is then set at the beginning of epoch N.

Note, that we don't actually use the TWAP. We just use the spot price at the start of epoch N for now.

## Latest OSMO equivalent TWAP

Snapshot of OSMO equivalent that is updated at the end of epoch, can be queried by `GetLatestEpochOsmoEquivalentTWAP`.

Note: For now, it is updated with latest OSMO equivalent per LP token.
