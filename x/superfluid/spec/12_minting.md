<!--
order: 12
-->

# Minting


Superfluid module has the ability to arbitrarily mint and burn Osmo through the `bank` module. This is potentially dangerous so we strictly constrain it's ability to do so.
This authority is mediated through the `mintOsmoTokensAndDelegate` and `forceUndelegateAndBurnOsmoTokens` keeper methods, which are in turn called by message handlers (`SuperfluidDelegate`, `SuperfluidUndelegate`, and `SuperfluidDelegateMore`) as well as by `RefreshIntermediaryDelegationAmounts` as part of an `Epoch` hook.

## Invariant
Each of these mechanisms maintains a local invariant between the amount of Osmo minted and delegated by the `IntermediaryAccount`, and the quantity of the underlying asset held by locks associated to the account, modified by `OsmoEquivalentMultiplier` and `RiskAdjustment` for the underlying asset. Namely that total minted/delegated = `GetTotalSyntheticAssetsLocked` * `GetOsmoEquivalentMultiplier` * `GetRiskAdjustment`

This can be equivalently expressed as `GetExpectedDelegationAmount` being equal to the actual delegation amount.

## RefreshIntermediaryDelegationAmounts
In the `RefreshIntermediaryDelegationAmounts` method, calls are made to `mintOsmoTokensAndDelegate` or `forceUndelegateAndBurnOsmoTokens` to adjust the real delegation up or down to match `GetExpectedDelegationAmount`.

## Message Handlers

### SuperfluidDelegate
In a `SuperfluidDelegate` transaction, we first verify that this lock is not already associated to an `IntermediaryAccount`, and then use `mintOsmoTokenAndDelegate` to properly balance the resulting change in `GetExpectedDelegationAmount` from the increase in `GetTotalSyntheticAssetsLocked`.
i.e. we mint and delegate: `GetOsmoEquivalentMultiplier` * `GetRiskAdjustment` * `lock.Coins.Amount` new Osmo tokens.

### SuperfluidDelegateMore

### SuperfluidUndelegate




