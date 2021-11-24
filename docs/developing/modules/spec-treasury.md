---
sidebarDepth: 2
---

# Treasury

The Treasury module acts as the "central bank" of the Osmosis economy, measuring macroeconomic activity by [observing indicators](#observed-indicators) and adjusting [monetary policy levers](#monetary-policy-levers) to modulate miner incentives toward stable, long-term growth.

::: warning Note:
While the Treasury stabilizes miner demand by adjusting rewards, the [`Market`](spec-market.md) module is responsible for Osmosis price-stability through arbitrage and the market maker.
:::

## Concepts

### Observed Indicators

The Treasury observes three macroeconomic indicators for each epoch and keeps [indicators](#indicators) of their values during previous epochs:

- **Tax Rewards**: $T$, the income generated from transaction and stability fees during an epoch.
- **Seigniorage Rewards***: $S$, the amount of seigniorage generated from Osmo swaps to Osmosis during an epoch which is destined for ballot rewards inside the `Oracle` rewards. As of Columbus-5, all seigniorage is burned.
- **Total Staked Osmo**: $\lambda$, the total amount of Osmo staked by users and bonded to their delegated validators.

These indicators are used to derive two other values:
- **Tax Reward per unit Osmo** $\tau = T / \lambda$: this is used in [Updating Tax Rate](#k-updatetaxpolicy)
- **Total mining rewards** $R = T + S$: the sum of the Tax Rewards and the Seigniorage Rewards, used in [Updating Reward Weight](#k-updaterewardpolicy).

::: warning Note:
As of Columbus-5, all seigniorage is burned and no longer funds community or reward pools.
:::

- **Seigniorage Rewards:**: $S$, The amount of seigniorage generated from Osmo swaps to Osmosis during each epoch.

::: warning Note:
As of Columbus-5, all seigniorage is burned.   
:::

These indicators can be used to derive two other values, the **Tax Reward per unit Osmo** represented by $\tau = T / \lambda$, used in [Updating Tax Rate](#k-updatetaxpolicy), and **total mining rewards** $R = T + S$: the sum of the Tax Rewards and the Seigniorage Rewards, used in [Updating Reward Weight](#k-updaterewardpolicy).

The protocol can compute and compare the short-term ([`WindowShort`](#windowshort)) and the long-term ([`WindowLong`](#windowlong)) rolling averages of the above indicators to determine the relative direction and velocity of the Osmosis economy.

### Monetary Policy Levers

- **Tax Rate**: $r$, adjusts the amount of income gained from Osmosis transactions, limited by [_tax cap_](#tax-caps).

- **Reward Weight**: $w$, the portion of seigniorage allocated to the reward pool for [`Oracle`](spec-oracle.md) vote winners. This is given to validtors who vote within the reward band of the weighted median exchange rate.

::: warning Tip
As of Columbus-5, all seigniorage is burned and no longer funds the community pool or the oracle reward pool. Validators are rewarded for faithful oracle votes through swap fees.
:::

### Updating Policies

Both [Tax Rate](#tax-rate) and [Reward Weight](#reward-weight) are stored as values in the `KVStore` and can have their values updated through [governance proposals](#governance-proposals) after they have passed. The Treasury recalibrates each lever once per epoch to stabilize unit returns for Osmo, ensuring predictable mining rewards from staking:

- For Tax Rate, in order to make sure that unit mining rewards do not stay stagnant, the treasury adds a [`MiningIncrement`](#miningincrement) so mining rewards increase steadily over time, described [here](#kupdatetaxpolicy).

- For Reward Weight, the Treasury observes the portion of seigniorage needed to bear the overall reward profile, [`SeigniorageBurdenTarget`](#seigniorageburdentarget), and raises rates accordingly, as described [here](#k-updaterewardpolicy). The current Reward Weight is `1`.

### Probation

A probationary period specified by the [`WindowProbation`](#windowprobation) prevents the network from performing Tax Rate and Reward Weight updates during the first epochs after genesis to allow the blockchain to first obtain a critical mass of transactions and a mature, reliable history of indicators.

## Data

### PolicyConstraints

Policy updates from governance proposals and automatic calibration are constrained by the [`TaxPolicy`](#taxpolicy) and [`RewardPolicy`](#rewardpolicy) parameters, respectively. `PolicyConstraints` specifies the floor, ceiling, and max periodic changes for each variable.

```go
// PolicyConstraints defines constraints around updating a key Treasury variable
type PolicyConstraints struct {
    RateMin       sdk.Dec  `json:"rate_min"`
    RateMax       sdk.Dec  `json:"rate_max"`
    Cap           sdk.Coin `json:"cap"`
    ChangeRateMax sdk.Dec  `json:"change_max"`
}
```

The logic for constraining a policy lever update is performed by `pc.Clamp()`.

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

## Proposals

The Treasury module defines special proposals which allow the [Tax Rate](#tax-rate) and [Reward Weight](#reward-weight) values in the `KVStore` to be voted on and changed accordingly, subject to the [policy constraints](#policy-constraints) imposed by `pc.Clamp()`.

### TaxRateUpdateProposal

```go
type TaxRateUpdateProposal struct {
	Title       string  `json:"title" yaml:"title"`             // Title of the Proposal
	Description string  `json:"description" yaml:"description"` // Description of the Proposal
	TaxRate     sdk.Dec `json:"tax_rate" yaml:"tax_rate"`       // target TaxRate
}
```

::: warning Note:
As of Columbus-5, all seigniorage is burned. The Reward Weight is now set to `1`.
:::

## State

### Tax Rate

- type: `Dec`
- min: .1%
- max: 1%

The value of the Tax Rate policy lever for the current epoch.

### Reward Weight

- type: `Dec`
- default: `1`

The value of the Reward Weight policy lever for the current epoch. As of Columbus-5, the reward weight is set to `1`.

### Tax Caps

- type: `map[string]Int`

The Treasury keeps a `KVStore` that maps a denomination `denom` to an `sdk.Int` which represents the maximum income that can be generated from taxes on a transaction in that same denomination. This is updated every epoch with the equivalent value of [`TaxPolicy.Cap`](#taxpolicy) at the current exchange rate.

For example, if a transaction's value were 100 SDT with a tax rate of 5% and a tax cap of 1 SDT, the income generated would be 1 SDT, not 5 SDT.

### Tax Proceeds

- type: `Coins`

The Tax Rewards $T$ for the current epoch.

### Epoch Initial Issuance

- type: `Coins`

The total supply of Osmo at the beginning of the current epoch. This value is used in [`k.SettleSeigniorage()`](#k-settleseigniorage) to calculate the seigniorage distributed at the end of each epoch. As of Columbus 5, all seigniorage is burned.

Recording the initial issuance will automatically use the [`Supply`](spec-supply.md) module to determine the total issuance of Osmo. Peeking will return the epoch's initial issuance of µOsmo as `sdk.Int` instead of `sdk.Coins` for clarity.

### Indicators

The Treasury keeps track of following indicators for the present and previous epochs:

#### Tax Rewards

- type: `Dec`

The Tax Rewards $T$ for each `epoch`.

#### Seigniorage Rewards

- type: `Dec`

The Seigniorage Rewards $S$ for each `epoch`.

#### Total Staked Osmo

- type: `Int`

The Total Staked Osmo $\lambda$ for each `epoch`.

## Functions

### `k.UpdateIndicators()`

```go
func (k Keeper) UpdateIndicators(ctx sdk.Context)
```

At the end of each epoch $t$, this function records the current values of tax rewards $T$, seigniorage rewards $S$, and total staked Osmo $\lambda$ before moving to the next epoch $t+1$.

- $T_t$ is the current value in [`TaxProceeds`](#tax-proceeds)
- $S_t = \Sigma * w$, with epoch seigniorage $\Sigma$ and reward weight $w$.
- $\lambda_t$ is the result of `staking.TotalBondedTokens()`

### `k.UpdateTaxPolicy()`

```go
func (k Keeper) UpdateTaxPolicy(ctx sdk.Context) (newTaxRate sdk.Dec)
```

At the end of each epoch, this funtion calculates the next value of the Tax Rate monetary lever.

Using $r_t$ as the current Tax Rate and $n$ as the [`MiningIncrement`](#miningincrement) parameter:

1. Calculate the rolling average $\tau_y$ of Tax Rewards per unit Osmo over the last year `WindowLong`.

2. Calculate the rolling average $\tau_m$ of Tax Rewards per unit Osmo over the last month `WindowShort`.

3. If $\tau_m = 0$, there was no tax revenue in the last month. The Tax Rate should be set to the maximum permitted by the Tax Policy, subject to the rules of `pc.Clamp()` (see [constraints](#policy-constraints)).

4. If $\tau_m > 0$, the new Tax Rate is $r_{t+1} = (n r_t \tau_y)/\tau_m$, subject to the rules of `pc.Clamp()`. See [constraints](#policy-constraints) for more details.

When monthly tax revenues dip below the yearly average, the Treasury raises the Tax Rate. When monthly tax revenues go above the yearly average, the Treasury lowers the Tax Rate.

### `k.UpdateRewardPolicy()`

```go
func (k Keeper) UpdateRewardPolicy(ctx sdk.Context) (newRewardWeight sdk.Dec)
```

At the end of each epoch, this funtion calculates the next value of the Reward Weight monetary lever.

Using $w_t$ as the current reward weight, and $b$ as the [`SeigniorageBurdenTarget`](#seigniorageburdentarget) parameter:

1. Calculate the sum $S_m$ of seigniorage rewards over the last month `WindowShort`.

2. Calculate the sum $R_m$ of total mining rewards over the last month `WindowShort`.

3. If $R_m = 0$ and $S_m = 0$, there were no mining or seigniorage rewards in the last month. The Reward Weight should be set to the maximum permitted by the Reward Policy, subject to the rules of `pc.Clamp()`. See [constraints](#policy-constraints) for more details.

4. If $R_m > 0$ or $S_m > 0$, the new Reward Weight is $w_{t+1} = b w_t S_m / R_m$, subject to the rules of `pc.Clamp()`. See [constraints](#policy-constraints) for more details.


::: warning Note:
As of Columbus-5, all seigniorage is burned and no longer funds the community or reward pools.
:::

### `k.UpdateTaxCap()`

```go
func (k Keeper) UpdateTaxCap(ctx sdk.Context) sdk.Coins
```

This function is called at the end of an epoch to compute the Tax Caps for every denomination for the next epoch.

For every denomination in circulation, the new Tax Cap for each denomination is set to be the global Tax Cap defined in the [`TaxPolicy`](#taxpolicy) parameter, at current exchange rates.

### `k.SettleSeigniorage()`

```go
func (k Keeper) SettleSeigniorage(ctx sdk.Context)
```

This function is called at the end of an epoch to compute seigniorage and forwards the funds to the [`Oracle`](spec-oracle.md) module for ballot rewards, and the [`Distribution`](spec-distribution.md) for the community pool.

1. The seigniorage $\Sigma$ of the current epoch is calculated by taking the difference between the Osmo supply at the start of the epoch ([Epoch Initial Issuance](#epoch-initial-issuance)) and the Osmo supply at the time of calling.

   Note that $\Sigma > 0$ when the current Osmo supply is lower than at the start of the epoch, because the Osmo had been burned from Osmo swaps into Osmosis. See [here](spec-market.md#seigniorage).

2. The Reward Weight $w$ is the percentage of the seigniorage designated for ballot rewards. Amount $S$ of new Osmo is minted, and the [`Oracle`](spec-oracle.md) module receives $S = \Sigma * w$ of the seigniorage.

3. The remainder of the coins $\Sigma - S$ is sent to the [`Distribution`](spec-distribution.md) module, where it is allocated into the community pool.

::: warning Note:
As of Columbus-5, all seigniorage is burned and no longer funds the community pool or the oracle reward pool. Validators are rewarded for faithful oracle votes through swap fees.
:::

## Transitions

### End-Block

If the blockchain is at the final block of the epoch, the following procedure is run:

1. Update all the indicators with [`k.UpdateIndicators()`](#kupdateindicators)

2. If the this current block is under [probation](#probation), skip to step 6.

3. [Settle seigniorage](#ksettleseigniorage) accrued during the epoch and make funds available to ballot rewards and the community pool during the next epoch. As of Columbus-5, all seigniorage is burned.

4. Calculate the [Tax Rate](#k-updatetaxpolicy), [Reward Weight](#k-updaterewardpolicy), and [Tax Cap](#k-updatetaxcap) for the next epoch. As of Columbus-5, all seigniorage is burned, and the reward weight is set to `1`.

5. Emit the [`policy_update`](#policy_update) event, recording the new policy lever values.

6. Finally, record the Osmo issuance with [`k.RecordEpochInitialIssuance()`](#epoch-initial-issuance). This will be used in calculating the seigniorage for the next epoch.

::: details Events

| Type          | Attribute Key | Attribute Value |
| ------------- | ------------- | --------------- |
| policy_update | tax_rate      | {taxRate}       |
| policy_update | reward_weight | {rewardWeight}  |
| policy_update | tax_cap       | {taxCap}        |

:::

## Parameters

The subspace for the Treasury module is `treasury`.

```go
type Params struct {
	TaxPolicy               PolicyConstraints `json:"tax_policy" yaml:"tax_policy"`
	RewardPolicy            PolicyConstraints `json:"reward_policy" yaml:"reward_policy"`
	SeigniorageBurdenTarget sdk.Dec           `json:"seigniorage_burden_target" yaml:"seigniorage_burden_target"`
	MiningIncrement         sdk.Dec           `json:"mining_increment" yaml:"mining_increment"`
	WindowShort             int64             `json:"window_short" yaml:"window_short"`
	WindowLong              int64             `json:"window_long" yaml:"window_long"`
	WindowProbation         int64             `json:"window_probation" yaml:"window_probation"`
}
```

### TaxPolicy

- type: `PolicyConstraints`
- default:

```go
DefaultTaxPolicy = PolicyConstraints{
    RateMin:       sdk.NewDecWithPrec(5, 4), // 0.05%
    RateMax:       sdk.NewDecWithPrec(1, 2), // 1%
    Cap:           sdk.NewCoin(core.MicroSDRDenom, sdk.OneInt().MulRaw(core.MicroUnit)), // 1 SDR Tax cap
    ChangeRateMax: sdk.NewDecWithPrec(25, 5), // 0.025%
}
```

Constraints / rules for updating the [Tax Rate](#tax-rate) monetary policy lever.

### RewardPolicy

- type: `PolicyConstraints`
- default:

```go
DefaultRewardPolicy = PolicyConstraints{
    RateMin:       sdk.NewDecWithPrec(5, 2), // 5%
    RateMax:       sdk.NewDecWithPrec(90, 2), // 90%
    ChangeRateMax: sdk.NewDecWithPrec(25, 3), // 2.5%
    Cap:           sdk.NewCoin("unused", sdk.ZeroInt()), // UNUSED
}
```

Constraints / rules for updating the [Reward Weight](#reward-weight) monetary policy lever.

### SeigniorageBurdenTarget

- type: `sdk.Dec`
- default: 67%

Multiplier specifying portion of burden seigniorage needed to bear the overall reward profile for Reward Weight updates during epoch transition.

### MiningIncrement

- type: `sdk.Dec`
- default: 1.07 growth rate, 15% CAGR of $\tau$

Multiplier determining an annual growth rate for Tax Rate policy updates during epoch transition.

### WindowShort

- type: `int64`
- default: `4` (month = 4 weeks)

A number of epochs that specifies a time interval for calculating short-term moving average.

### WindowLong

- type: `int64`
- default: `52` (year = 52 weeks)

A number of epochs that specifies a time interval for calculating long-term moving average.

### WindowProbation

- type: `int64`
- default: `12` (3 months = 12 weeks)

A number of epochs that specifies a time interval for the probationary period.
