<!--
order: 4
-->

# Epochs

Overall Epoch sequence:
* Superfluid
  * Claim staking rewards for every `Intermediary Account`
  * Put rewards into gauges for payout to owner's of `Synthetic Locks`
  * Update `Osmo Equivalent Multiplier` value for each LP token
    * (Currently spot price at epoch)
  * Refresh delegation amounts for all `Intermediary Accounts`
    * If expected amount > current delegation:
      * Mint new `Osmo` and `Delegate`
    * If expected amount < current delegation:
      * Use `InstantUndelegate` and burn the extra `Osmo`
* Incentives
  * Payout rewards from gagues to `Synthetic Lock` owners
  * (Also pay out regular LP incentives)
* Mint
  * Issue new Osmo, and send to various modules (distribution, incentives, etc.)
  * 25% currently goes to `x/distribution` which funds `Staking` and `Superfluid` rewards
  * Rewards for `Superfluid` are based on the just updated delegation amounts, and queued for payout in the next epoch
