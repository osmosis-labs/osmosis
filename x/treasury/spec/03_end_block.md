<!--
order: 3
-->

# EndBlock

If the blockchain is at the final block of the epoch, the following procedure is run:

1. Update all the indicators with `k.UpdateIndicators()`

2. If the this current block is under [probation](./01_concepts.md#Probation), skip to step 6.

3. Settle seigniorage accrued during the epoch and make funds available to ballot rewards and the community pool during the next epoch.

4. Calculate the `Tax Rate`, `Reward Weight`, and `Tax Cap` for the next epoch.

5. Emit the `policy_update` event, recording the new policy lever values.

6. Finally, record the Luna issuance with `k.RecordEpochInitialIssuance()`. This will be used in calculating the seigniorage for the next epoch.

# Functions

## `k.UpdateIndicators()`

```go
func (k Keeper) UpdateIndicators(ctx sdk.Context)
```

This function gets run at the end of an epoch  and records the current values of tax rewards $T$, seigniorage rewards $S$, and total staked Luna $\Sigma$ as the historic indicators for epoch $t$ before moving to the next epoch $t+1$.

$T_t$ is the current value in TaxProceeds
,$S_t = \Sigma * w$ with epoch seigniorage $\Sigma$ and reward weight $w$.
$\lambda _t$ is simply the result of `staking.TotalBondedTokens()`.

## `k.UpdateTaxPolicy()`

```go
func (k Keeper) UpdateTaxPolicy(ctx sdk.Context) (newTaxRate sdk.Dec)
```

This function gets called at the end of an epoch to calculate the next value of the Tax Rate monetary lever.

Consider $\tau _t$ to be the current Tax Rate, and $n$ to be the `MiningIncrement` parameter.

1. Calculate the rolling average $\tau _y$ of Tax Rewards per unit Luna over the last year `WindowLong`.

2. Calculate the rolling average $\tau _m$` of Tax Rewards per unit Luna over the last month `WindowShort`.

3. If $\tau _m = 0$, there was no tax revenue in the last month. The Tax Rate should thus be set to the maximum permitted by the Tax Policy, subject to the rules of `pc.Clamp()`.

4. Otherwise, the new Tax Rate is $r_{t+1} = n r_t \tau _y / \tau _m$, subject to the rules of `pc.Clamp()`.

As such, the Treasury hikes up Tax Rate when tax revenues in a shorter time window is performing poorly in comparison to the longer term tax revenue average. It lowers Tax Rate when short term tax revenues are outperforming the longer term index.

## `k.UpdateRewardPolicy()`

```go
func (k Keeper) UpdateRewardPolicy(ctx sdk.Context) (newRewardWeight sdk.Dec)
```

This function gets called at the end of an epoch to calculate the next value of the Reward Weight monetary lever.

Consider $w_t$ to be the current reward weight, and $b$ to be the SeigniorageBurdenTarget parameter.

1. Calculate the sum of $S_m$ of seignorage rewards over the last month `WindowShort`.

2. Calculate the sum of $R_m$ of total mining rewards over the last month `WindowShort`.

3. If either $R_m = 0$ or $S_m = 0$, there was no mining and seigniorage rewards in the last month. The Rewards Weight should thus be set to the maximum permitted by the Reward Policy, subject to the rules of `pc.Clamp()`.

4. Otherwise, the new Reward Weight is $w_{t+1} = b w_t S_m / R_m$, subject to the rules of `pc.Clamp()`.

### `k.UpdateTaxCap()`

```go
func (k Keeper) UpdateTaxCap(ctx sdk.Context) sdk.Coins
```

This function is called at the end of an epoch to compute the Tax Caps for every denomination for the next epoch.

For each denomination in circulation, the new Tax Cap for that denomination is set to be the global Tax Cap defined in the `TaxPolicy` parameter, at current exchange rates.

### `k.SettleSeigniorage()`

```go
func (k Keeper) SettleSeigniorage(ctx sdk.Context)
```

This function is called at the end of an epoch to compute seigniorage and forwards the funds to the [`Oracle`](../../oracle/spec/README.md) module for ballot rewards, and the [`Distribution`](https://github.com/cosmos/cosmos-sdk/tree/master/x/distribution/spec/README.md) for the community pool.

1. The seigniorage $\Sigma$ of the current epoch is calculated by taking the difference between the Luna supply at the start of the epoch ([Epoch Initial Issuance](./02_state.md#EpochInitialIssuance)) and the Luna supply at the time of calling.

   Note that $\Sigma > 0$ when the current Luna supply is lower than at the start of the epoch, because the Luna had been burned from Luna swaps into Terra. See [here](../../market/spec/01_concepts.md#Seigniorage).

2. The Reward Weight $w$ is the percentage of the seigniorage designated for ballot rewards. Amount $S$ of new Luna is minted, and the [`Oracle`](../../oracle/spec/README.md) module receives $S = \Sigma * w$ of the seigniorage.

3. The remainder of the coins $\Sigma - S$ is sent to the [`Distribution`](https://github.com/cosmos/cosmos-sdk/tree/master/x/distribution/spec/README.md) module, where it is allocated into the community pool.

## PolicyConstraints

Policy updates from both governance proposals and automatic calibration are constrained by the `TaxPolicy` and `RewardPolicy` parameters, respectively. The type `PolicyConstraints` specifies the floor, ceiling, and the max periodic changes for each variable.

```go
// PolicyConstraints defines constraints around updating a key Treasury variable
type PolicyConstraints struct {
    RateMin       sdk.Dec  `json:"rate_min"`
    RateMax       sdk.Dec  `json:"rate_max"`
    Cap           sdk.Coin `json:"cap"`
    ChangeRateMax sdk.Dec  `json:"change_max"`
}
```

The logic for constraining a policy lever update is performed by `pc.Clamp()`, shown below.

```go
// Clamp constrains a policy variable update within the policy constraints
func (pc PolicyConstraints) Clamp(prevRate sdk.Dec, newRate sdk.Dec) (clampedRate sdk.Dec) {
	if newRate.LT(pc.RateMin) {
		newRate = pc.RateMin
	} else if newRate.GT(pc.RateMax) {
		newRate = pc.RateMax
	}

	delta := newRate.Sub(prevRate)
	if newRate.GT(prevRate) {
		if delta.GT(pc.ChangeRateMax) {
			newRate = prevRate.Add(pc.ChangeRateMax)
		}
	} else {
		if delta.Abs().GT(pc.ChangeRateMax) {
			newRate = prevRate.Sub(pc.ChangeRateMax)
		}
	}
	return newRate
}
```