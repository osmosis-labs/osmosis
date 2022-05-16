<!--
order: 0
title: "Incentives Overview"
parent:
  title: "incentives"
-->

# `incentives`

## Abstract

Incentives module provides general interface to give yield to stakers.

The yield to be given to stakers are stored in `gauge` and it is distributed on epoch basis to the stakers who meet specific conditions.

Anyone can create gauge and add rewards to the gauge, there is no way to take it out other than distribution.

There are two kinds of `gauges`, perpetual and non-perpetual ones.

- Non perpetual ones get removed from active queue after the the distribution period finish but perpetual ones persist.
- For non perpetual ones, they distribute the tokens equally per epoch during the `gauge` is in the active period.
- For perpetual ones, it distribute all the tokens at a single time and somewhere else put the tokens regularly to distribute the tokens, it's mainly used to distribute minted OSMO tokens to LP token stakers.

## Contents

1. **[Concept](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Messages](03_messages.md)**
4. **[Events](04_events.md)**
5. **[Hooks](05_hooks.md)**  
6. **[Queries](06_queries.md)**  
7. **[Params](07_params.md)**  

## Overview 

The purpose of incentives module is to provide incentives to users who lock certain tokens for specified periods of time.

Locked tokens can be of any denomination, including LP tokens (gamm/pool/x), IBC tokens (tokens sent through IBC such as ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2), and native tokens (such as ATOM or LUNA). The incentive amount is entered by the gauge creator. Rewards for a given pool of locked up tokens are pooled into a gauge until the disbursement time. At the disbursement time, they are distributed pro-rata (proportionally) to members of the pool.

Anyone can create a gauge and add rewards to the gauge. There is no way to withdraw gauge rewards other than distribution. Governance proposals can be raised to match the external incentive tokens with equivalent Osmo incentives (see for example: [proposal 47](https://www.mintscan.io/osmosis/proposals/47)).

There are two kinds of gauges: **`perpetual`** and **`non-perpetual`**:

- **`Non-perpetual`** gauges distribute their tokens equally per epoch while the gauge is in the active period. These gauges get removed from the active queue after the distribution period finishes

- **`Perpetual gauges`** distribute all their tokens at a single time and only distribute their tokens again once the gauge is refilled (this is mainly used to distribute minted OSMO tokens to LP token stakers). Perpetual gauges persist and will re-disburse tokens when refilled (there is no "active" period)



</br>
</br>

## Transactions

### create-gauge

Create a gauge to distribute rewards to users

```
osmosisd tx incentives create-gauge [lockup_denom] [reward] [flags]
```

::: details Example 1


I want to make incentives for LP tokens of pool 3, namely gamm/pool/3 that have been locked up for at least 1 day.
I want to reward 100 AKT to this pool over 2 days (2 epochs). (50 rewarded on each day)
I want the rewards to start dispersing on 21 December 2021 (1640081402 UNIX time)

```bash
osmosisd tx incentives create-gauge gamm/pool/3 10000ibc/1480B8FD20AD5FCAE81EA87584D269547DD4D436843C1D20F15E00EB64743EF4 \ 
--duration 24h  --start-time 1640081402 --epochs 2 --from WALLET_NAME --chain-id osmosis-1
```
:::

::: details Example 2

I want to make incentives for ATOM (ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2) that have been locked up for at least 1 week (164h).
I want to reward 1000 JUNO (ibc/46B44899322F3CD854D2D46DEEF881958467CDD4B3B10086DA49296BBED94BED) to ATOM holders perpetually (perpetually meaning I must add more tokens to this gauge myself every epoch). I want the reward to start dispersing immediately.

```bash
osmosisd tx incentives create-gauge ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2 \
1000000000ibc/46B44899322F3CD854D2D46DEEF881958467CDD4B3B10086DA49296BBED94BED --perpetual --duration 168h \
--from WALLET_NAME --chain-id osmosis-1
```
:::


### add-to-gauge

Add coins to a gauge previously created to distribute more rewards to users

```
osmosisd tx incentives add-to-gauge [gauge_id] [rewards] [flags]
```

::: details Example

I want to refill the gauge with 500 JUNO to a previously created gauge (gauge ID 1914) after the distribution.

```bash
osmosisd tx incentives add-to-gauge 1914 500000000ibc/46B44899322F3CD854D2D46DEEF881958467CDD4B3B10086DA49296BBED94BED \
--from WALLET_NAME --chain-id osmosis-1
```
:::


## Queries

### active-gauges           

Query active gauges

```
osmosisd query incentives active-gauges [flags]
```

::: details Example

```bash
osmosisd query incentives active-gauges
```

An example output

```
- coins: []
  distribute_to:
    denom: gamm/pool/99
    duration: 604800s
    lock_query_type: ByDuration
    timestamp: "0001-01-01T00:00:00Z"
  distributed_coins: []
  filled_epochs: "0"
  id: "297"
  is_perpetual: true
  num_epochs_paid_over: "1"
  start_time: "2021-07-03T12:27:09.323840990Z"
- coins: []
  distribute_to:
    denom: gamm/pool/99
    duration: 1209600s
    lock_query_type: ByDuration
    timestamp: "0001-01-01T00:00:00Z"
  distributed_coins: []
  filled_epochs: "0"
  id: "298"
  is_perpetual: true
  num_epochs_paid_over: "1"
  start_time: "2021-07-03T12:27:09.323840990Z"
pagination:
  next_key: BwEAAAAAAAAAHTIwMjEtMDctMDNUMTI6Mjc6MDkuMzIzODQwOTkw
  total: "0"
...
```
:::



### active-gauges-per-denom 

Query active gauges per denom

```
osmosisd query incentives active-gauges-per-denom [denom] [flags]
```

::: details Example

Query all active gauges distributing incentives to holders of gamm/pool/341

```bash
osmosisd query incentives active-gauges-per-denom gamm/pool/341
```

An example output:

```
- coins: []
  distribute_to:
    denom: gamm/pool/341
    duration: 604800s
    lock_query_type: ByDuration
    timestamp: "0001-01-01T00:00:00Z"
  distributed_coins: []
  filled_epochs: "0"
  id: "1033"
  is_perpetual: true
  num_epochs_paid_over: "1"
  start_time: "2021-09-06T22:42:52.139465318Z"
- coins: []
  distribute_to:
    denom: gamm/pool/341
    duration: 1209600s
    lock_query_type: ByDuration
    timestamp: "0001-01-01T00:00:00Z"
  distributed_coins: []
  filled_epochs: "0"
  id: "1034"
  is_perpetual: true
  num_epochs_paid_over: "1"
  start_time: "2021-09-06T22:42:52.139465318Z"
pagination:
  next_key: BwEAAAAAAAAAHTIwMjEtMDctMDNUMTI6Mjc6MDkuMzIzODQwOTkw
  total: "0"
...
```
:::



### distributed-coins       

Query coins distributed so far

```
osmosisd query incentives distributed-coins [flags]
```

::: details Example

```bash
osmosisd query incentives distributed-coins
```

An example output:

```
coins:
- amount: "27632051924"
  denom: ibc/0954E1C28EB7AF5B72D24F3BC2B47BBB2FDF91BDDFD57B74B99E133AED40972A
- amount: "3975960654"
  denom: ibc/0EF15DF2F02480ADE0BB6E85D9EBB5DAEA2836D3860E9F97F9AADE4F57A31AA0
- amount: "125999980901"
  denom: ibc/1480B8FD20AD5FCAE81EA87584D269547DD4D436843C1D20F15E00EB64743EF4
- amount: "434999992789"
  denom: ibc/1DC495FCEFDA068A3820F903EDBD78B942FBD204D7E93D3BA2B432E9669D1A59
- amount: "3001296"
  denom: ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2
- amount: "1493887986685"
  denom: ibc/3BCCC93AD5DF58D11A6F8A05FA8BC801CBA0BA61A981F57E91B8B598BF8061CB
- amount: "372218215714"
  denom: ibc/46B44899322F3CD854D2D46DEEF881958467CDD4B3B10086DA49296BBED94BED
- amount: "1049999973206"
  denom: ibc/4E5444C35610CC76FC94E7F7886B93121175C28262DDFDDE6F84E82BF2425452
- amount: "11666666665116"
  denom: ibc/7A08C6F11EF0F59EB841B9F788A87EC9F2361C7D9703157EC13D940DC53031FA
- amount: "13199999715662"
  denom: ibc/9712DBB13B9631EDFA9BF61B55F1B2D290B2ADB67E3A4EB3A875F3B6081B3B84
- amount: "1177777428443"
  denom: ibc/D805F1DA50D31B96E4282C1D4181EDDFB1A44A598BFF5666F4B43E4B8BEA95A5
- amount: "466666567747"
  denom: ibc/EA3E1640F9B1532AB129A571203A0B9F789A7F14BB66E350DCBFA18E1A1931F0
- amount: "79999999178"
  denom: ibc/F3FF7A84A73B62921538642F9797C423D2B4C4ACB3C7FCFFCE7F12AA69909C4B
- amount: "65873607694598"
  denom: uosmo
```
:::


### gauge-by-id             

Query gauge by id

```
osmosisd query incentives gauge-by-id [id] [flags]
```

::: details Example

Query the incentive distribution for gauge ID 1:

```
osmosisd query incentives gauge-by-id 1
```

```bash
gauge:
  coins:
  - amount: "16654747773959"
    denom: uosmo
  distribute_to:
    denom: gamm/pool/1
    duration: 86400s
    lock_query_type: ByDuration
    timestamp: "0001-01-01T00:00:00Z"
  distributed_coins:
  - amount: "16589795315655"
    denom: uosmo
  filled_epochs: "182"
  id: "1"
  is_perpetual: true
  num_epochs_paid_over: "1"
  start_time: "2021-06-19T04:30:19.082462364Z"
```
:::




### gauges                  

Query available gauges

```
osmosisd query incentives gauges [flags]
```

::: details Example

Query ALL gauges (by default the limit is 100, so here I will define a much larger number to output all gauges)

```bash
osmosisd query incentives gauges --limit 2000
```

An example output:

```
- coins:
  - amount: "1924196414964"
    denom: uosmo
  distribute_to:
    denom: gamm/pool/348
    duration: 604800s
    lock_query_type: ByDuration
    timestamp: "0001-01-01T00:00:00Z"
  distributed_coins: []
  filled_epochs: "0"
  id: "8"
  is_perpetual: true
  num_epochs_paid_over: "1"
  start_time: "2021-10-04T13:59:02.142175968Z"
- coins:
  - amount: "641398804181"
    denom: uosmo
  distribute_to:
    denom: gamm/pool/348
    duration: 1209600s
    lock_query_type: ByDuration
    timestamp: "0001-01-01T00:00:00Z"
  distributed_coins: []
  filled_epochs: "0"
  id: "9"
  is_perpetual: true
  num_epochs_paid_over: "1"
  start_time: "2021-10-04T13:59:02.142175968Z"
pagination:
  next_key: null
  total: "0"
...
```
:::






### rewards-estimation      

Query rewards estimation

// Error: strconv.ParseUint: parsing "": invalid syntax






### to-distribute-coins     

Query coins that is going to be distributed

```
osmosisd query incentives to-distribute-coins [flags]
```

::: details Example

```bash
osmosisd query incentives to-distribute-coins
```

An example output:

```
coins:
- amount: "20000000"
  denom: gamm/pool/87
- amount: "90791948076"
  denom: ibc/0954E1C28EB7AF5B72D24F3BC2B47BBB2FDF91BDDFD57B74B99E133AED40972A
- amount: "10000"
  denom: ibc/1480B8FD20AD5FCAE81EA87584D269547DD4D436843C1D20F15E00EB64743EF4
- amount: "1000"
  denom: ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2
- amount: "10728832013315"
  denom: ibc/3BCCC93AD5DF58D11A6F8A05FA8BC801CBA0BA61A981F57E91B8B598BF8061CB
- amount: "627782783496"
  denom: ibc/46B44899322F3CD854D2D46DEEF881958467CDD4B3B10086DA49296BBED94BED
- amount: "450000026794"
  denom: ibc/4E5444C35610CC76FC94E7F7886B93121175C28262DDFDDE6F84E82BF2425452
- amount: "38333333334884"
  denom: ibc/7A08C6F11EF0F59EB841B9F788A87EC9F2361C7D9703157EC13D940DC53031FA
- amount: "46800000284338"
  denom: ibc/9712DBB13B9631EDFA9BF61B55F1B2D290B2ADB67E3A4EB3A875F3B6081B3B84
- amount: "2822222571557"
  denom: ibc/D805F1DA50D31B96E4282C1D4181EDDFB1A44A598BFF5666F4B43E4B8BEA95A5
- amount: "2533333432253"
  denom: ibc/EA3E1640F9B1532AB129A571203A0B9F789A7F14BB66E350DCBFA18E1A1931F0
- amount: "366164843847"
  denom: uosmo
```
:::


### upcoming-gauges         

Query scheduled gauges (gauges whose `start_time` has not yet occurred)

```
osmosisd query incentives upcoming-gauges [flags]
```

::: details Example

```bash
osmosisd query incentives upcoming-gauges
```

Using this command, we will see the gauge we created earlier, among all other upcoming gauges:

```
- coins:
  - amount: "10000"
    denom: ibc/1480B8FD20AD5FCAE81EA87584D269547DD4D436843C1D20F15E00EB64743EF4
  distribute_to:
    denom: gamm/pool/3
    duration: 86400s
    lock_query_type: ByDuration
    timestamp: "1970-01-01T00:00:00Z"
  distributed_coins: []
  filled_epochs: "0"
  id: "1914"
  is_perpetual: false
  num_epochs_paid_over: "2"
  start_time: "2021-12-21T10:10:02Z"
...
```
:::