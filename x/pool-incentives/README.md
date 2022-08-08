# Pool Incentives

## Abstract
The `pool-incentives` module is separate but related to the `incentives` module. When a pool is created using the `GAMM` module, the `pool-incentives` module automatically creates individual gauges in the `incentives` module for every lock duration that exists in that pool. 
The `pool-incentives` module also takes the `pool_incentives` distributed from the `gov` module and distributes it to the various incentivized gauges.

## Contents

1. **[Concept](#concepts)**
2. **[State](#state)**
3. **[Governance](#gov)**
4. **[Transactions](#transactions)**
5. **[Queries](#queries)**

## Concepts

The purpose of the `pool incentives` module is to distribute incentives
to a pool's LPs. This assumes that pool's follow the interface from the
`x/gamm` module

`Pool incentives` module doesn't directly distribute the rewards to the
LPs. When a pool is created, the `pool incentives` module creates a
`gauge` in the `incentives` module for every lock duration that exists.
Also, the `pool incentives` module takes a part of the minted inflation
from the mint module, and automatically distributes it to the various
selected gauges.

## State

### Genesis states

```go
type GenesisState struct {
 // params defines all the paramaters of the module.
 Params            Params          
 LockableDurations []time.Duration 
 DistrInfo         *DistrInfo      
}

type Params struct {
 // minted_denom is the denomination of the coin expected to be minted
 //  by the minting module.
 // Pool-incentives module doesn’t actually mint the coin itself, 
 // but rather manages the distribution of coins that matches the defined minted_denom.
 MintedDenom string 
 // allocation_ratio defines the proportion of the minted minted_denom 
 // that is to be allocated as pool incentives.
 AllocationRatio github_com_cosmos_cosmos_sdk_types.Dec 
}
```

Lockable durations can be set to the pool incentives module at genesis.
Every time a pool is created, the `pool incentives` module creates the
same amount of 'gauge' as there are lockable durations for the pool.

Also in regards to the `Params`, when the mint module mints new tokens
to the fee collector at Begin Block, the `pool incentives` module takes
the token which matches the 'minted denom' from the fee collector.
Tokens are taken according to the 'allocationRatio', and are distributed
to each `DistrRecord` of the DistrInfo. For example, if the fee
collector holds 1000uatom and 2000 uosmo at Begin Block, and Params'
mintedDenom is set to uosmo, and AllocationRatio is set to 0.1, 200uosmo
will be taken from the fee collector and distributed to the
`DistrRecord`s.

## Gov

`Pool Incentives` module uses the values set at genesis or values added
by chain governance to distribute part of the inflation minted by the
mint module to specified gauges.

```go
type DistrInfo struct {
 TotalWeight github_com_cosmos_cosmos_sdk_types.Int 
 Records     []DistrRecord                          
}

type DistrRecord struct {
 GaugeId  uint64                                 
 Weight github_com_cosmos_cosmos_sdk_types.Int 
}
```

`DistrInfo` internally manages the `DistrRecord` and total weight of all
`DistrRecord`. Governance can modify DistrInfo via
`UpdatePoolIncentivesProposal` proposal.

### UpdatePoolIncentivesProposal

```go
type UpdatePoolIncentivesProposal struct {
 Title       string       
 Description string      
 Records     []DistrRecord 
}
```

`UpdatePoolIncentivesProposal` can be used by governance to update
`DistrRecord`s.

```shell
osmosisd tx gov submit-proposal update-pool-incentives [gaugeIds] [weights]
```

Proposals can be proposed in using the CLI command format above.\
For example, to designate 100 weight to gauge id 2 and 200 weight to
gauge id 3, the following command can be used.

```shell
osmosisd tx gov submit-proposal update-pool-incentives 2,3 100,200
```


## Transactions

### replace-pool-incentives 

```sh
osmosisd tx poolincentives replace-pool-incentives [gaugeIds] [weights] [flags]
```

::: details Example 

Fully replace records for pool incentives:

```bash
osmosisd tx poolincentives replace-pool-incentives proposal.json --from --chain-id
```

The proposal.json would look as follows:

```json
{
  "title": "Pool Incentive Adjustment",
  "description": "Adjust pool incentives",
  "records": [
    {
      "gauge_id": "0",
      "weight": "100000"
    },
    {
      "gauge_id": "1",
      "weight": "1766249"
    },
    {
      "gauge_id": "XXX",
      "weight": "XXXXXXXX"
    },
    ...
  ]
}
```
:::




### update-pool-incentives  

Update the weight of specified pool gauges in regards to their share of incentives (by creating a proposal)

```sh
osmosisd tx poolincentives update-pool-incentives [gaugeIds] [weights] [flags] --from --chain-id
```

::: details Example

Update the pool incentives for `gauge_id` 0 and 1:

```bash
osmosisd tx gov submit-proposal update-pool-incentives proposal.json --from WALLET_NAME --chain-id CHAIN_ID
```

The proposal.json would look as follows:

```json
{
  "title": "Pool Incentive Adjustment",
  "description": "Adjust pool incentives",
  "records": [
    {
      "gauge_id": "0",
      "weight": "100000"
    },
    {
      "gauge_id": "1",
      "weight": "1766249"
    },
  ]
}
```
:::


## Queries

### distr-info                   

Query distribution info for all pool gauges

```sh
osmosisd query poolincentives distr-info
```

::: details Example

```bash
osmosisd query poolincentives distr-info
```

An example output:

```bash
  - gauge_id: "1877"
    weight: "60707"
  - gauge_id: "1878"
    weight: "40471"
  - gauge_id: "1897"
    weight: "1448"
  - gauge_id: "1898"
    weight: "869"
  - gauge_id: "1899"
    weight: "579"
...
```
:::


### external-incentivized-gauges 

Query externally incentivized gauges (gauges distributing rewards on top of the normal OSMO rewards)

```sh
osmosisd query pool-incentives external-incentivized-gauges
```

::: details Example

```bash
osmosisd query pool-incentives external-incentivized-gauges
```

An example output:

```bash
- coins:
  - amount: "596400000"
    denom: ibc/0EF15DF2F02480ADE0BB6E85D9EBB5DAEA2836D3860E9F97F9AADE4F57A31AA0
  distribute_to:
    denom: gamm/pool/562
    duration: 604800s
    lock_query_type: ByDuration
    timestamp: "1970-01-01T00:00:00Z"
  distributed_coins:
  - amount: "596398318"
    denom: ibc/0EF15DF2F02480ADE0BB6E85D9EBB5DAEA2836D3860E9F97F9AADE4F57A31AA0
  filled_epochs: "28"
  id: "1791"
  is_perpetual: false
  num_epochs_paid_over: "28"
  start_time: "1970-01-01T00:00:00Z"
- coins:
  - amount: "1000000"
    denom: ibc/46B44899322F3CD854D2D46DEEF881958467CDD4B3B10086DA49296BBED94BED
  distribute_to:
    denom: gamm/pool/498
    duration: 86400s
    lock_query_type: ByDuration
    timestamp: "1970-01-01T00:00:00Z"
  distributed_coins:
  - amount: "999210"
    denom: ibc/46B44899322F3CD854D2D46DEEF881958467CDD4B3B10086DA49296BBED94BED
  filled_epochs: "2"
  id: "1660"
  is_perpetual: false
  num_epochs_paid_over: "2"
  start_time: "2021-10-14T16:00:00Z"
...
```
:::

### gauge-ids                    

Query the gauge ids (by duration) by pool id

```sh
osmosisd query poolincentives gauge-ids [pool-id] [flags]
```

::: details Example

Find out what the gauge IDs are for pool 1:

```bash
osmosisd query poolincentives gauge-ids 1
```

An example output:

```bash
gauge_ids_with_duration:
- duration: 86400s
  gauge_id: "1"
- duration: 604800s
  gauge_id: "2"
- duration: 1209600s
  gauge_id: "3"
```

In this example, we see that gauge IDs 1,2, and 3 are for the one day, one week, and two week lockup periods respectively for the OSMO/ATOM pool.
:::



### incentivized-pools           

Query all incentivized pools with their respective gauge IDs and lockup durations

```sh
osmosisd query poolincentives incentivized-pools [flags]
```

::: details Example

```bash
osmosisd query poolincentives incentivized-pools
```

An example output:

```bash
- gauge_id: "1897"
  lockable_duration: 86400s
  pool_id: "602"
- gauge_id: "1898"
  lockable_duration: 604800s
  pool_id: "602"
- gauge_id: "1899"
  lockable_duration: 1209600s
  pool_id: "602"
...
```
:::




### lockable-durations           

Query incentivized lockup durations

```sh
osmosisd query poolincentives lockable-durations [flags]
```

::: details Example

```bash
osmosisd query poolincentives lockable-durations
```

An example output:

```bash
lockable_durations:
- 86400s
- 604800s
- 1209600s
```
:::



### params                       

Query pool-incentives module parameters

```sh
osmosisd query poolincentives params [flags]
```

::: details Example

```bash
osmosisd query poolincentives params
```

An example output:

```bash
params:
  minted_denom: uosmo
```
:::
