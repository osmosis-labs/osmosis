# Mint

The `mint` module is responsible for creating tokens in a flexible way to reward
validators, incentivize providing pool liquidity and fund the community pool.

The module is also responsible for reducing the token creation and distribution by a set period
until it reaches its maximum supply (see `reduction_factor` and `reduction_period_in_epochs`)

The module uses time basis epochs supported by the `epochs` module.

## Contents

1. **[Concept](#concepts)**
2. **[State](#state)**
3. **[Begin Epoch](#begin-epoch)**
4. **[Parameters](#network-parameters)**
5. **[Events](#events)**
6. **[Queries](#queries)**

## Concepts

The `x/mint` module is designed to handle the regular printing of new
tokens within a chain. The design taken within Osmosis is to

- Mint new tokens once per epoch (default one day)
- To have a "Reductioning factor" every period, which reduces the number of
  rewards per epoch. (default: period is 3 years, where a
  year is 52 epochs. The next period's rewards are 2/3 of the prior
  period's rewards)

### Reduction factor

This is a generalization over the Bitcoin-style halvenings. Every year, the number
of rewards issued per week will reduce by a governance-specified
factor, instead of a fixed `1/2`. So
`RewardsPerEpochNextPeriod = ReductionFactor * CurrentRewardsPerEpoch)`.
When `ReductionFactor = 1/2`, the Bitcoin halvenings are recreated. We
default to having a reduction factor of `2/3` and thus reduce rewards
at the end of every year by `33%`.

The implication of this is that the total supply is finite, according to
the following formula:

`Total Supply = InitialSupply + EpochsPerPeriod * { {InitialRewardsPerEpoch} / {1 - ReductionFactor} }`

## State

### Minter

The [`Minter`](https://github.com/osmosis-labs/osmosis/blob/cbb683e8395655042b4421355cd54a8c96bfa507/x/mint/types/mint.pb.go#L30) is an abstraction for holding current rewards information.

```go
type Minter struct {
    EpochProvisions osmomath.Dec   // Rewards for the current epoch
}
```

### Params

Minting [`Params`](https://github.com/osmosis-labs/osmosis/blob/cbb683e8395655042b4421355cd54a8c96bfa507/x/mint/types/mint.pb.go#L168) are held in the global params store.

### LastReductionEpoch

Last reduction epoch stores the epoch number when the last reduction of
coin mint amount per epoch has happened.

## Begin-Epoch

Minting parameters are recalculated and inflation is paid at the beginning
of each epoch. An epoch is signaled by x/epochs

### NextEpochProvisions

The target epoch provision is recalculated on each reduction period
(default 3 years). At the time of the reduction, the current provision is
multiplied by the reduction factor (default `2/3`), to calculate the
provisions for the next epoch. Consequently, the rewards of the next
period will be lowered by a `1` - reduction factor.

### EpochProvision

Calculate the provisions generated for each epoch based on current epoch
provisions. The provisions are then minted by the `mint` module's
`ModuleMinterAccount`. These rewards are transferred to a
`FeeCollector`, which handles distributing the rewards per the chain's needs.
This fee collector is specified as the `auth` module's `FeeCollector` `ModuleAccount`.

## Network Parameters

The minting module contains the following parameters:

| Key                                        | Type         | Example                                |
| ------------------------------------------ | ------------ | -------------------------------------- |
| mint_denom                                 | string       | "uosmo"                                |
| genesis_epoch_provisions                   | string (dec) | "500000000"                            |
| epoch_identifier                           | string       | "weekly"                               |
| reduction_period_in_epochs                 | int64        | 156                                    |
| reduction_factor                           | string (dec) | "0.6666666666666"                      |
| distribution_proportions.staking           | string (dec) | "0.4"                                  |
| distribution_proportions.pool_incentives   | string (dec) | "0.3"                                  |
| distribution_proportions.developer_rewards | string (dec) | "0" (deprecated)                       |
| distribution_proportions.community_pool    | string (dec) | "0.3"                                  |
| weighted_developer_rewards_receivers       | array        | [] (deprecated)                        |
| minting_rewards_distribution_start_epoch   | int64        | 10                                     |

Below are all the network parameters for the `mint` module:

- **`mint_denom`** - Token type being minted
- **`genesis_epoch_provisions`** - Amount of tokens generated at the epoch to the distribution categories (see distribution_proportions)
- **`epoch_identifier`** - Type of epoch that triggers token issuance (day, week, etc.)
- **`reduction_period_in_epochs`** - How many epochs must occur before implementing the reduction factor
- **`reduction_factor`** - What the total token issuance factor will reduce by after the reduction period passes (if set to 66.66%, token issuance will reduce by 1/3)
- **`distribution_proportions`** - Categories in which the specified proportion of newly released tokens are distributed to
  - **`staking`** - Proportion of minted funds to incentivize staking OSMO
  - **`pool_incentives`** - Proportion of minted funds to incentivize pools on Osmosis
  - **`developer_rewards`** - Deprecated, must be `0`. Developer vesting is not part of mint distribution
  - **`community_pool`** - Proportion of minted funds to be set aside for the community pool
- **`weighted_developer_rewards_receivers`** - Deprecated and unused
- **`minting_rewards_distribution_start_epoch`** - What epoch will start the rewards distribution to the aforementioned distribution categories

### Notes

1. `mint_denom` defines denom for minting token - uosmo
2. `genesis_epoch_provisions` provides minting tokens per epoch at genesis.
3. `epoch_identifier` defines the epoch identifier to be used for the mint module e.g. "weekly"
4. `reduction_period_in_epochs` defines the number of epochs to pass to reduce the mint amount
5. `reduction_factor` defines the reduction factor of tokens at every `reduction_period_in_epochs`
6. `distribution_proportions` defines distribution rules for minted tokens. The community pool
   receives the remainder after rounding. `developer_rewards` is deprecated and must be `0`.
7. `weighted_developer_rewards_receivers` is deprecated and unused. Developer vesting is not
   part of mint distribution.
8. `minting_rewards_distribution_start_epoch` defines the start epoch of minting to make sure
   minting start after initial pools are set

## Events

The minting module emits the following events:

### End of Epoch

| Type | Attribute Key    | Attribute Value   |
| ---- | ---------------- | ----------------- |
| mint | epoch_number     | {epochNumber}     |
| mint | epoch_provisions | {epochProvisions} |
| mint | amount           | {amount}          |

</br>
</br>

## Queries

### params

Query all the current mint parameter values

```sh
query mint params
```

::: details Example

List all current min parameters in json format by:

```bash
osmosisd query mint params -o json | jq
```

An example of the output:

```json
{
  "mint_denom": "uosmo",
  "genesis_epoch_provisions": "821917808219.178082191780821917",
  "epoch_identifier": "day",
  "reduction_period_in_epochs": "365",
  "reduction_factor": "0.666666666666666666",
  "distribution_proportions": {
    "staking": "0.250000000000000000",
    "pool_incentives": "0.450000000000000000",
    "developer_rewards": "0.000000000000000000",
    "community_pool": "0.300000000000000000"
  },
  "weighted_developer_rewards_receivers": [],
  "minting_rewards_distribution_start_epoch": "1"
}
```

:::

### epoch-provisions

Query the current epoch provisions

```sh
query mint epoch-provisions
```

::: details Example

List the current epoch provisions:

```bash
osmosisd query mint epoch-provisions
```

As of this writing, this number will be equal to the `genesis-epoch-provisions`. Once the `reduction_period_in_epochs` is reached, the `reduction_factor` will be initiated and reduce the amount of OSMO minted per epoch.
:::

### inflation

Query the current inflation rate

```sh
query mint inflation
```

::: details Example

Get the current inflation rate:

```bash
osmosisd query mint inflation
```

This returns the current annualized inflation rate affecting the circulating supply. The calculation is based on the formula:

`((Epoch provisions * (1 - Community Pool distribution proportion)) * 365) / Total Supply`

This gives the inflation rate that applies to tokens in circulation, excluding those allocated to the community pool.
:::

## Appendix

### Current Configuration

`mint` **module: Network Parameter effects and current configuration**

The following tables show overall effects on different configurations of the `mint` related network parameters:

<table><thead><tr><th></th> 
<th><code>mint_denom</code></th> 
<th><code>epoch_provisions</code></th> 
<th><code>epoch_identifier</code></th></tr></thead> <tbody>
<tr><td>Type</td> 
<td>string</td> 
<td>string (dec)</td> 
<td>string</td></tr> 
<tr><td>Higher</td> 
<td>N/A</td> 
<td>Higher inflation rate</td> 
<td>Increases time to <code>reduction_period</code></td></tr> 
<tr><td>Lower</td> 
<td>N/A</td> 
<td>Lower inflation rate</td> 
<td>Decreases time to <code>reduction_period</code></td></tr> 
<tr><td>Constraints</td> 
<td>N/A</td> 
<td>Value has to be a positive integer</td> 
<td>String must be <code>day</code>, <code>week</code>, <code>month</code>, or <code>year</code></td></tr> 
<tr><td>Current configuration</td> 
<td><code>uosmo</code></td> 
<td><code>821917808219.178</code> (821,9178 OSMO)</td> 
<td><code>day</code></td></tr>
</tbody></table>

<table><thead><tr><th></th> 
<th><code>reduction_period_in_epochs</code></th> 
<th><code>reduction_factor</code></th> 
<th><code>staking</code></th></tr></thead> 
<tbody><tr><td>Type</td> 
<td>string</td> 
<td>string (dec)</td> 
<td>string (dec)</td></tr> 
<tr><td>Higher</td> 
<td>Longer period of time until <code>reduction_factor</code> implemented</td> 
<td>Reduces time until maximum supply is reached</td> 
<td>More epoch provisions go to staking rewards than other categories</td></tr> 
<tr><td>Lower</td> 
<td>Shorter period of time until <code>reduction_factor</code> implemented</td> 
<td>Increases time until maximum supply is reached</td> 
<td>Less epoch provisions go to staking rewards than other categories</td></tr> 
<tr><td>Constraints</td> 
<td>Value has to be a whole number greater than or equal to <code>1</code></td> 
<td>Value has to be less or equal to <code>1</code></td> 
<td>Value has to be less or equal to <code>1</code> and all distribution categories combined must equal <code>1</code></td></tr> 
<tr><td>Current configuration</td> 
<td><code>365</code> (epochs)</td> 
<td><code>0.666666666666666666</code> (66.66%)</td> 
<td><code>0.250000000000000000</code> (25%)</td></tr>
</tbody></table>

<table><thead><tr><th></th> 
<th><code>pool_incentives</code></th> 
<th><code>developer_rewards</code></th> 
<th><code>community_pool</code></th></tr></thead> 
<tbody><tr><td>Type</td> 
<td>string (dec)</td> 
<td>string (dec)</td> 
<td>string (dec)</td></tr> 
<tr><td>Higher</td> 
<td>More epoch provisions go to pool incentives than other categories</td> 
<td>Deprecated. Must be <code>0</code></td> 
<td>More epoch provisions go to community pool than other categories</td></tr> 
<tr><td>Lower</td> 
<td>Less epoch provisions go to pool incentives than other categories</td> 
<td>Deprecated. Must be <code>0</code></td> 
<td>Less epoch provisions go to community pool than other categories</td></tr> 
<tr><td>Constraints</td> 
<td>Value has to be less or equal to <code>1</code> and all distribution categories combined must equal <code>1</code></td> 
<td>Must be <code>0</code></td>
<td>Value has to be less or equal to <code>1</code> and all distribution categories combined must equal <code>1</code></td></tr> 
<tr><td>Current configuration</td> 
<td><code>0.450000000000000000</code> (45%)</td> 
<td><code>0</code> (deprecated)</td>
<td><code>0.050000000000000000</code> (5%)</td></tr>
</tbody></table>
