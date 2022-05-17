# Mint

The ```mint``` module is responsible for creating tokens in a flexible way to reward 
validators, incentivize providing pool liquidity, provide funds for Osmosis governance, and pay developers to maintain and improve Osmosis.

The module is also responsible for reducing the token creation and distribution by a set amount and a set period of time until it reaches its maximum supply (see ```reduction_factor``` and ```reduction_period_in_epochs```)

Module uses time basis epochs supported by ```epochs``` module.

## Contents

1. **[Concept](01_concepts.md)**
2. **[State](02_state.md)**
3. **[End Epoch](03_end_epoch.md)**
4. **[Parameters](04_params.md)**
5. **[Events](05_events.md)**
    
## Overview 

### Network Parameters

Below are all the network parameters for the ```mint``` module:

- **```mint_denom```** - Token type being minted
- **```genesis_epoch_provisions```** - Amount of tokens generated at epoch to the distribution categories (see distribution_proportions)
- **```epoch_identifier```** - Type of epoch that triggers token issuance (day, week, etc.)
- **```reduction_period_in_epochs```** - How many epochs must occur before implementing the reduction factor
- **```reduction_factor```** - What the total token issuance factor will reduce by after reduction period passes (if set to 66.66%, token issuance will reduce by 1/3)
- **```distribution_proportions```** - Categories in which the specified proportion of newly released tokens are distributed to
  - **```staking```** - Proportion of minted funds to incentivize staking OSMO
  - **```pool_incentives```** - Proportion of minted funds to incentivize pools on Osmosis
  - **```developer_rewards```** - Proportion of minted funds to pay developers for their past and future work
  - **```community_pool```** - Proportion of minted funds to be set aside for the community pool
- **```weighted_developer_rewards_receivers```** - Addresses that developer rewards will go to. The weight attached to an address is the percent of the developer rewards that the specific address will receive
- **```minting_rewards_distribution_start_epoch```** - What epoch will start the rewards distribution to the aforementioned distribution categories

</br>
</br>

## Queries

### params

Query all the current mint parameter values

```
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
    "developer_rewards": "0.250000000000000000",
    "community_pool": "0.050000000000000000"
  },
  "weighted_developer_rewards_receivers": [
    {
      "address": "osmo14kjcwdwcqsujkdt8n5qwpd8x8ty2rys5rjrdjj",
      "weight": "0.288700000000000000"
    },
    {
      "address": "osmo1gw445ta0aqn26suz2rg3tkqfpxnq2hs224d7gq",
      "weight": "0.229000000000000000"
    },
    {
      "address": "osmo13lt0hzc6u3htsk7z5rs6vuurmgg4hh2ecgxqkf",
      "weight": "0.162500000000000000"
    },
    {
      "address": "osmo1kvc3he93ygc0us3ycslwlv2gdqry4ta73vk9hu",
      "weight": "0.109000000000000000"
    },
    {
      "address": "osmo19qgldlsk7hdv3ddtwwpvzff30pxqe9phq9evxf",
      "weight": "0.099500000000000000"
    },
    {
      "address": "osmo19fs55cx4594een7qr8tglrjtt5h9jrxg458htd",
      "weight": "0.060000000000000000"
    },
    {
      "address": "osmo1ssp6px3fs3kwreles3ft6c07mfvj89a544yj9k",
      "weight": "0.015000000000000000"
    },
    {
      "address": "osmo1c5yu8498yzqte9cmfv5zcgtl07lhpjrj0skqdx",
      "weight": "0.010000000000000000"
    },
    {
      "address": "osmo1yhj3r9t9vw7qgeg22cehfzj7enwgklw5k5v7lj",
      "weight": "0.007500000000000000"
    },
    {
      "address": "osmo18nzmtyn5vy5y45dmcdnta8askldyvehx66lqgm",
      "weight": "0.007000000000000000"
    },
    {
      "address": "osmo1z2x9z58cg96ujvhvu6ga07yv9edq2mvkxpgwmc",
      "weight": "0.005000000000000000"
    },
    {
      "address": "osmo1tvf3373skua8e6480eyy38avv8mw3hnt8jcxg9",
      "weight": "0.002500000000000000"
    },
    {
      "address": "osmo1zs0txy03pv5crj2rvty8wemd3zhrka2ne8u05n",
      "weight": "0.002500000000000000"
    },
    {
      "address": "osmo1djgf9p53n7m5a55hcn6gg0cm5mue4r5g3fadee",
      "weight": "0.001000000000000000"
    },
    {
      "address": "osmo1488zldkrn8xcjh3z40v2mexq7d088qkna8ceze",
      "weight": "0.000800000000000000"
    }
  ],
  "minting_rewards_distribution_start_epoch": "1"
}
```
:::


### epoch-provisions

Query the current epoch provisions

```
query mint epoch-provisions
```

::: details Example

List the current epoch provisions:

```bash
osmosisd query mint epoch-provisions
```
As of this writing, this number will be equal to the ```genesis-epoch-provisions```. Once the ```reduction_period_in_epochs``` is reached, the ```reduction_factor``` will be initiated and reduce the amount of OSMO minted per epoch.
:::

## Appendix

### Current Configuration

```mint``` **module: Network Parameter effects and current configuration**

The following tables show overall effects on different configurations of the ```mint``` related network parameters:

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
<td>More epoch provisions go to developer rewards than other categories</td> 
<td>More epoch provisions go to community pool than other categories</td></tr> 
<tr><td>Lower</td> 
<td>Less epoch provisions go to pool incentives than other categories</td> 
<td>Less epoch provisions go to developer rewards than other categories</td> 
<td>Less epoch provisions go to community pool than other categories</td></tr> 
<tr><td>Constraints</td> 
<td>Value has to be less or equal to <code>1</code> and all distribution categories combined must equal <code>1</code></td> 
<td>Value has to be less or equal to <code>1</code> and all distribution categories combined must equal <code>1</code></td> 
<td>Value has to be less or equal to <code>1</code> and all distribution categories combined must equal <code>1</code></td></tr> 
<tr><td>Current configuration</td> 
<td><code>0.450000000000000000</code> (45%)</td> 
<td><code>0.250000000000000000</code> (25%)</td> 
<td><code>0.050000000000000000</code> (5%)</td></tr>
</tbody></table>