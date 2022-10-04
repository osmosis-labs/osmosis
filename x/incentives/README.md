# Incentives

## Abstract

Incentives module provides general interface to give yield to stakers.

The yield to be given to stakers are stored in `gauge` and it is distributed on epoch basis to the stakers who meet specific conditions.

Anyone can create gauge and add rewards to the gauge, there is no way to take it out other than distribution.

There are two kinds of `gauges`, perpetual and non-perpetual ones.

- Non perpetual ones get removed from active queue after the the distribution period finish but perpetual ones persist.
- For non perpetual ones, they distribute the tokens equally per epoch during the `gauge` is in the active period.
- For perpetual ones, it distribute all the tokens at a single time and somewhere else put the tokens regularly to distribute the tokens, it's mainly used to distribute minted OSMO tokens to LP token stakers.

## Contents

1. **[Concept](#concepts)**
2. **[State](#state)**
3. **[Messages](#messages)**
4. **[Events](#events)**
5. **[Hooks](#hooks)**
6. **[Params](#parameters)**
7. **[Transactions](#transactions)**
8. **[Queries](#queries)**

## Concepts

The purpose of `incentives` module is to provide incentives to the users
who lock specific token for specific period of time.

Locked tokens can be of any denomination, including LP tokens (gamm/pool/x), IBC tokens (tokens sent through IBC such as ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2), and native tokens (such as ATOM or LUNA).

The incentive amount is entered by the gauge creator. Rewards for a given pool of locked up tokens are pooled into a gauge until the disbursement time. At the disbursement time, they are distributed pro-rata (proportionally) to members of the pool.

Anyone can create a gauge and add rewards to the gauge. There is no way to withdraw gauge rewards other than distribution. Governance proposals can be raised to match the external incentive tokens with equivalent Osmo incentives (see for example: [proposal 47](https://www.mintscan.io/osmosis/proposals/47)).

There are two kinds of gauges: **`perpetual`** and **`non-perpetual`**:

- **`Non-perpetual`** gauges distribute their tokens equally per epoch while the gauge is in the active period. These gauges get removed from the active queue after the distribution period finishes

- **`Perpetual gauges`** distribute all their tokens at a single time and only distribute their tokens again once the gauge is refilled (this is mainly used to distribute minted OSMO tokens to LP token stakers). Perpetual gauges persist and will re-disburse tokens when refilled (there is no "active" period)

## State

### Incentives management

All the incentives that are going to be provided are locked into
`IncentivePool` until released to the appropriate recipients after a
specific period of time.

### Gauge

Rewards to be distributed are organized by `Gauge`. The `Gauge`
describes how users can get reward, stores the amount of coins in the
gauge, the cadence at which rewards are to be distributed, and the
number of epochs to distribute the reward over.

```protobuf
enum LockQueryType {
  option (gogoproto.goproto_enum_prefix) = false;

  ByDuration = 0; // locks which has more than specific duration
  ByTime = 1; // locks which are started before specific time
}

message QueryCondition {
  LockQueryType lock_query_type = 1; // type of lock, ByLockDuration | ByLockTime
  string denom = 2; // lock denom
  google.protobuf.Duration duration = 3; // condition for lock duration, only valid if positive
  google.protobuf.Timestamp timestamp = 4; // condition for lock start time, not valid if unset value
}

message Gauge {
  uint64 id = 1; // unique ID of a Gauge
  QueryCondition distribute_to = 2; // distribute condition of a lock which meet one of these conditions
  repeated cosmos.base.v1beta1.Coin coins = 3; // can distribute multiple coins
  google.protobuf.Timestamp start_time = 4; // condition for lock start time, not valid if unset value
  uint64 num_epochs_paid_over = 5; // number of epochs distribution will be done
}
```

### Gauge queues

#### Upcoming queue

To start release `Gauges` at a specific time, we schedule distribution
start time with time key queue.

#### Active queue

Active queue has all the `Gauges` that are distributing and after
distribution period finish, it's removed from the queue.

#### Active by Denom queue

To speed up the distribution process, module introduces the active
`Gauges` by denom.

#### Finished queue

Finished queue saves the `Gauges` that has finished distribution to keep
in track.

#### Module state

The state of the module is expressed by `params`, `lockable_durations`
and `gauges`.

```protobuf
// GenesisState defines the incentives module's genesis state.
message GenesisState {
  // params defines all the parameters of the module
  Params params = 1 [ (gogoproto.nullable) = false ];
  repeated Gauge gauges = 2 [ (gogoproto.nullable) = false ];
  repeated google.protobuf.Duration lockable_durations = 3 [
    (gogoproto.nullable) = false,
    (gogoproto.stdduration) = true,
    (gogoproto.moretags) = "yaml:\"lockable_durations\""
  ];
}
```

## Messages

### Create Gauge

`MsgCreateGauge` can be submitted by any account to create a `Gauge`.

```go
type MsgCreateGauge struct {
 Owner             sdk.AccAddress
  DistributeTo      QueryCondition
  Rewards           sdk.Coins
  StartTime         time.Time // start time to start distribution
  NumEpochsPaidOver uint64 // number of epochs distribution will be done
}
```

**State modifications:**

- Validate `Owner` has enough tokens for rewards
- Generate new `Gauge` record
- Save the record inside the keeper's time basis unlock queue
- Transfer the tokens from the `Owner` to incentives `ModuleAccount`.

### Adding balance to Gauge

`MsgAddToGauge` can be submitted by any account to add more incentives
to a `Gauge`.

```go
type MsgAddToGauge struct {
 GaugeID uint64
  Rewards sdk.Coins
}
```

**State modifications:**

- Validate `Owner` has enough tokens for rewards
- Check if `Gauge` with specified `msg.GaugeID` is available
- Modify the `Gauge` record by adding `msg.Rewards`
- Transfer the tokens from the `Owner` to incentives `ModuleAccount`.

## Events

The incentives module emits the following events:

### Handlers

#### MsgCreateGauge

| Type         | Attribute Key        | Attribute Value     |
| ------------ | -------------------- | ------------------- |
| create_gauge | gauge_id             | {gaugeID}           |
| create_gauge | distribute_to        | {owner}             |
| create_gauge | rewards              | {rewards}           |
| create_gauge | start_time           | {startTime}         |
| create_gauge | num_epochs_paid_over | {numEpochsPaidOver} |
| message      | action               | create_gauge        |
| message      | sender               | {owner}             |
| transfer     | recipient            | {moduleAccount}     |
| transfer     | sender               | {owner}             |
| transfer     | amount               | {amount}            |

#### MsgAddToGauge

| Type         | Attribute Key | Attribute Value |
| ------------ | ------------- | --------------- |
| add_to_gauge | gauge_id      | {gaugeID}       |
| create_gauge | rewards       | {rewards}       |
| message      | action        | create_gauge    |
| message      | sender        | {owner}         |
| transfer     | recipient     | {moduleAccount} |
| transfer     | sender        | {owner}         |
| transfer     | amount        | {amount}        |

### EndBlockers

#### Incentives distribution

| Type         | Attribute Key | Attribute Value |
| ------------ | ------------- | --------------- |
| transfer\[\] | recipient     | {receiver}      |
| transfer\[\] | sender        | {moduleAccount} |
| transfer\[\] | amount        | {distrAmount}   |

## Hooks

In this section we describe the "hooks" that `incentives` module provide
for other modules.

If there's no usecase for this, we could ignore this.

```go
 AfterCreateGauge(ctx sdk.Context, gaugeId uint64)
 AfterAddToGauge(ctx sdk.Context, gaugeId uint64)
 AfterStartDistribution(ctx sdk.Context, gaugeId uint64)
 AfterFinishDistribution(ctx sdk.Context, gaugeId uint64)
 AfterDistribute(ctx sdk.Context, gaugeId uint64)
```

## Parameters

The incentives module contains the following parameters:

| Key                  | Type   | Example  |
| -------------------- | ------ | -------- |
| DistrEpochIdentifier | string | "weekly" |

Note: DistrEpochIdentifier is a epoch identifier, and module distribute
rewards at the end of epochs. As `epochs` module is handling multiple
epochs, the identifier is required to check if distribution should be
done at `AfterEpochEnd` hook

</br>
</br>

## Transactions

### create-gauge

Create a gauge to distribute rewards to users

```sh
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

```sh
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

In this section we describe the queries required on grpc server.

```protobuf
// Query defines the gRPC querier service.
service Query {
  // returns coins that is going to be distributed
  rpc ModuleToDistributeCoins(ModuleToDistributeCoinsRequest) returns (ModuleToDistributeCoinsResponse) {}
  // returns Gauge by id
  rpc GaugeByID(GaugeByIDRequest) returns (GaugeByIDResponse) {}
  // returns gauges both upcoming and active
  rpc Gauges(GaugesRequest) returns (GaugesResponse) {}
  // returns active gauges
  rpc ActiveGauges(ActiveGaugesRequest) returns (ActiveGaugesResponse) {}
  // returns scheduled gauges
  rpc UpcomingGauges(UpcomingGaugesRequest) returns (UpcomingGaugesResponse) {}
  // RewardsEst returns an estimate of the rewards at a future specific time.
  // The querier either provides an address or a set of locks
  // for which they want to find the associated rewards.
  rpc RewardsEst(RewardsEstRequest) returns (RewardsEstResponse) {}
  // returns lockable durations that are valid to give incentives
  rpc LockableDurations(QueryLockableDurationsRequest) returns (QueryLockableDurationsResponse) {}
}
```

### active-gauges

Query active gauges

```sh
osmosisd query incentives active-gauges [flags]
```

::: details Example

```bash
osmosisd query incentives active-gauges
```

An example output

```sh
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

```sh
osmosisd query incentives active-gauges-per-denom [denom] [flags]
```

::: details Example

Query all active gauges distributing incentives to holders of gamm/pool/341

```bash
osmosisd query incentives active-gauges-per-denom gamm/pool/341
```

An example output:

```sh
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

```sh
osmosisd query incentives distributed-coins [flags]
```

::: details Example

```bash
osmosisd query incentives distributed-coins
```

An example output:

```sh
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

```sh
osmosisd query incentives gauge-by-id [id] [flags]
```

::: details Example

Query the incentive distribution for gauge ID 1:

```sh
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

```sh
osmosisd query incentives gauges [flags]
```

::: details Example

Query ALL gauges (by default the limit is 100, so here I will define a much larger number to output all gauges)

```bash
osmosisd query incentives gauges --limit 2000
```

An example output:

```sh
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

```sh
osmosisd query incentives to-distribute-coins [flags]
```

::: details Example

```bash
osmosisd query incentives to-distribute-coins
```

An example output:

```sh
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

```sh
osmosisd query incentives upcoming-gauges [flags]
```

::: details Example

```bash
osmosisd query incentives upcoming-gauges
```

Using this command, we will see the gauge we created earlier, among all other upcoming gauges:

```sh
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
