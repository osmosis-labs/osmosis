<!--
order: 12
-->

# Minting


Superfluid module has the ability to arbitrarily mint and burn Osmo through the `bank` module. This is potentially dangerous so we strictly constrain it's ability to do so.
This authority is mediated through the `mintOsmoTokensAndDelegate` and `forceUndelegateAndBurnOsmoTokens` keeper methods, which are in turn called by message handlers (`SuperfluidDelegate` and `SuperfluidUndelegate`) as well as by hooks on Epoch (`RefreshIntermediaryDelegationAmounts`) and Lockup (`SuperfluidDelegateMore`)

## Invariant
Each of these mechanisms maintains a local invariant between the amount of Osmo minted and delegated by the `IntermediaryAccount`, and the quantity of the underlying asset held by locks associated to the account, modified by `OsmoEquivalentMultiplier` and `RiskAdjustment` for the underlying asset. Namely that total minted/delegated = `GetTotalSyntheticAssetsLocked` * `GetOsmoEquivalentMultiplier` * `GetRiskAdjustment`

This can be equivalently expressed as `GetExpectedDelegationAmount` being equal to the actual delegation amount.


## Message Handlers

### SuperfluidDelegate
In a `SuperfluidDelegate` transaction, we first verify that this lock is not already associated to an `IntermediaryAccount`, and then use `mintOsmoTokenAndDelegate` to properly balance the resulting change in `GetExpectedDelegationAmount` from the increase in `GetTotalSyntheticAssetsLocked`.
i.e. we mint and delegate: `GetOsmoEquivalentMultiplier` * `GetRiskAdjustment` * `lock.Coins.Amount` new Osmo tokens.

### SuperfluidUndelegate
When a user submits a transaction to unlock their asset the invariant is maintained by using `forceUndelegateAndBurnOsmoTokens` to remove an amount of Osmo equal to `lockedCoin.Amount` * `GetOsmoEquivalentMultiplier` * `GetRiskAdjustment`.

## Hooks

### RefreshIntermediaryDelegationAmounts (AfterEpochEnd Hook)
In the `RefreshIntermediaryDelegationAmounts` method, calls are made to `mintOsmoTokensAndDelegate` or `forceUndelegateAndBurnOsmoTokens` to adjust the real delegation up or down to match `GetExpectedDelegationAmount`.

### SuperfluidDelegateMore (AfterAddTokensToLock Hook)
This is called as a result of a user adding more assets to a lock that has already been associated to an `IntermediaryAccount`. The invariant is maintained by using `mintOsmoTokenAndDelegate` to match the amount of new asset locked * `GetOsmoEquivalentMultiplier` * `GetRiskAdjustment` for the underlying asset.





