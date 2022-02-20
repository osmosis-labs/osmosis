<!--
order: 4
-->

# Epochs

At the Osmosis rewards distribution epoch time, all superfluid staking rewards get distributed.
The envisioned flow of how this works is as follows:

* (Epochs) AfterEpochEnd hook runs for epoch N
* (Mint) distributes rewards to all stakers at the epoch that just endeds prices
* (Superfluid) Claim all staking rewards to every intermediary module accounts
* (Superfluid) Update all TWAP values [updateEpochEnd][./../keeper/hooks.go]
  * Here we are setting the TWAP value for epoch N+1, as the TWAP from the duration of epoch N.
  * Currently using spot price at epoch time.
* (Superfluid) Update all the intermediary accounts staked amounts. (Mint/Burn coins as needed as well)
  * TODO: Consider if this is wrong for burn, should that come before reward distrbution?
