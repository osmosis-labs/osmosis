<!--
order: 2
-->

# State

## TaxRate

The value of the Tax Rate policy lever for the current epoch.

- TaxRate: `0x01 -> amino(sdk.Dec)`

## RewardWeight

The value of the Reward Weight policy lever for the current epoch.

- RewardWeight: `0x02 -> amino(sdk.Dec)`

## TaxCap

Treasury keeps a `KVStore` that maps a denomination `denom` to an `sdk.Int` that represents that maximum income that can be generated from taxes on a transaction in that denomination. This is updated every epoch with the equivalent value of `TaxPolicy.Cap` at the current exchange rate.

For instance, if a transaction's value were 100 SDT, and tax rate and tax cap 5% and 1 SDT respectively, the income generated from the transaction would be 1 SDT instead of 5 SDT, as it exceeds the tax cap.

- TaxCap: `0x03<denom_Bytes> -> amino(sdk.Int)`

## TaxProceeds

The Tax Rewards $T$ for the current epoch.

- TaxProceeds: `0x04 -> amino(sdk.Coins)`

## EpochInitialIssuance

The total supply of Luna at the beginning of the current epoch. This value is used in `k.SettleSeigniorage()` to calculate the seigniorage to distribute at the end of the epoch.

Recording the initial issuance will automatically use the `Supply` module to determine the total issuance of Luna. Peeking will return the epoch's initial issuance of µLuna as `sdk.Int` instead of `sdk.Coins` for convenience.

- EpochInitialIssuance: `0x05 -> amino(sdk.Coins)`

## Indicators
The Treasury keeps track of following indicators for the present and previous epochs:

### TaxReward
The Tax Rewards  for the `epoch`.

- TaxReward: `0x06<epoch_Bytes> -> amino(sdk.Dec)`

### SeigniorageReward
The Seigniorage Rewards $S$ for the `epoch`.

- SeigniorageReward: `0x07<epoch_Bytes> -> amino(sdk.Dec)`

### TotalStakedLuna
The Total Staked Luna $\lambda$ for the `epoch`.

- TotalStakedLuna: `0x08<epoch_Bytes> -> amino(sdk.Int)`

## CumulativeHeight

The cumulative height to keep the indicators on the hard fork.

- CumulativeHeight: `0x09 -> amino(int64)`

