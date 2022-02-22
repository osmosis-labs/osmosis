<!--
order: 4
-->

# Epochs

Overall Epoch sequence

* Epoch N ends:
  * Distribute gauge rewards for all non-superfluid gauges
  * Mint new tokens
    * Issue new Osmo, and send to various modules (distribution, incentives, etc.)
    * 25% currently goes to `x/distribution` which funds `Staking` and `Superfluid` rewards
    * Rewards for `Superfluid` are based on the just updated delegation amounts, and queued for payout in the next epoch
* First block after epoch:
  * Claim staking rewards for every `Intermediary Account`, put them into gauges.
  * Distribute Superfluid staking rewards from gauges to bonded Synthetic Lock owners
  * Update `Osmo Equivalent Multiplier` value for each LP token
    * (Currently spot price at epoch)
  * Refresh delegation amounts for all `Intermediary Accounts`
    * Calcultate the expected delegation for this account as `Osmo Equivalent Multipler` *`# LP Shares`* `Risk adjustment`
      * If this is less than 0.000001 `Osmo` it will be rounded to 0
    * Lookup current delegation amount for `Intermediary Account`
      * If there is no delegation, treat the current delegation as 0
    * If expected amount > current delegation:
      * Mint new `Osmo` and `Delegate` to `Validator`
    * If expected amount < current delegation:
      * Use `InstantUndelegate` and burn the received `Osmo`
