<!--
order: 4
-->

# Epochs

At the Osmosis rewards distribution epoch time, all superfluid staking rewards get distributed.
The envisioned flow of how this works is as follows:

* (Epochs) AfterEpochEnd hook runs
* (Mint) distributes rewards to all stakers at the epoch that just endeds prices
* (Superfluid) Update all TWAP values [updateEpochEnd][./../keeper/hooks.go]
* (Superfluid) Claim all staking rewards to every intermediary module accounts
* (Superfluid) Update all the intermediary accounts staked amounts. (Mint/Burn coins as needed as well)
