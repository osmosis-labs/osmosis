<!--
order: 1
-->

# Concepts

## Observed Indicators

The Treasury observes three macroeconomic indicators for each epoch (set to 1 week) and keeps historical records of their values during previous epochs.

* Tax Rewards: $T$, Income generated from transaction fees (stability fee) in a during the epoch.
* Seigniorage Rewards: $S$, Amount of seignorage generated from Luna swaps to Terra during the epoch that is destined for ballot rewards inside the [Oracle](../../oracle/spec/README.md) rewards.
* Total Staked Luna: $\lambda$, total Luna that has been staked by users and bonded by their delegated validators.

These indicators can be used to derive two other values, the **Tax Reward per unit Luna** represented by $\tau = T / \lambda$, used in Updating Tax Rate, and total mining rewards $R = T + S$, simply the sum of the Tax Rewards and the Seigniorage Rewards, used in Updating Reward Weight.

The protocol can compute and compare the short-term (`WindowShort`) and long-term (`WindowLong`) rolling averages of the above indicators to determine the relative direction and velocity of the Terra economy.

## Monetary Policy Levers

> From Columbus-3, the Reward Weight lever replaces the previous lever for controlling the rate of Luna burn in seigniorage. Now, miners are compensated through burning from swap fees, and ballot rewards in the oracle.
> 
TaxRate $r$ adjusts the amount of income coming from Terra transactions, limited by `TaxCap`.

RewardWeight $w$ which is the portion of seigniorage allocated for the reward pool for the ballot winners for correctly voting within the reward band of the weighted median of exchange rate in the [Oracle](../../oracle/spec/README.md) module.

## Updating Policies

Both `TaxRate` and `RewardWeight` are stored as values in the `KVStore`, and can have their values updated through governance proposals once passed. The Treasury will also re-calibrate each lever once per epoch to stabilize unit returns for Luna, thereby ensuring predictable mining rewards from staking:

* For Tax Rate, in order to make sure that unit mining rewards do not stay stagnant, the treasury adds a `MiningIncrement` so mining rewards increase steadily over time.

* For Reward Weight, The Treasury observes the portion of burden seigniorage needed to bear the overall reward profile, `SeigniorageBurdenTarget`, and hikes up rates accordingly.

## Probation

A probationary period specified by the `WindowProbation` will prevent the network from performing updates for Tax Rate and Reward Weight during the first epochs after genesis to allow the blockchain to first obtain a critical mass of transactions and a mature and reliable history of indicators.
