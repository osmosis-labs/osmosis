# Epochs

## Abstract

Often in the SDK, we would like to run certain code every-so often. The
purpose of `epochs` module is to allow other modules to set that they
would like to be signaled once every period. So another module can
specify it wants to execute code once a week, starting at UTC-time = x.
`epochs` creates a generalized epoch interface to other modules so that
they can easily be signalled upon such events.

## Contents

1. **[Concept](#concepts)**
2. **[State](#state)**
3. **[Events](#events)**
4. **[Keeper](#keeper)**
5. **[Hooks](#hooks)**
6. **[Queries](#queries)**
7. **[Downtime Recovery](#downtime-recovery)**

## Concepts

The epochs module defines on-chain timers, that execute at fixed time intervals.
Other SDK modules can then register logic to be executed at the timer ticks.
We refer to the period in between two timer ticks as an "epoch".

Every timer has a unique identifier.
Every epoch will have a start time, and an end time, where `end time = start time + timer interval`.
On Osmosis mainnet, we only utilize one identifier, with a time interval of `one day`.

The timer will tick at the first block whose blocktime is greater than the timer end time,
and set the start as the prior timer end time. (Notably, its not set to the block time!)
This means that if the chain has been down for awhile, you will get one timer tick per block,
until the timer has caught up.

## State

Epochs module keeps a single [`EpochInfo`](https://github.com/osmosis-labs/osmosis/blob/b4befe4f3eb97ebb477323234b910c4afafab9b7/proto/osmosis/epochs/genesis.proto#L12) per identifier.
Its fields are modified at every timer tick. 
EpochInfos are initialized as part of genesis initialization or upgrade logic,
and are only modified on begin blockers.

```protobuf
message EpochInfo {
  // Identifier is a unique reference to this particular timer.
  string identifier = 1;
  // Start time is the time at which the timer first ever ticks.
  // If start time is in the future, the epoch will not begin until the start
  // time.
  google.protobuf.Timestamp start_time = 2 [
    (gogoproto.stdtime) = true,
    (gogoproto.nullable) = false,
    (gogoproto.moretags) = "yaml:\"start_time\""
  ];
  // Duration is the time in between epoch ticks.
  // In order for intended behavior to be met, duration should
  // be greater than the chains expected block time.
  // Duration must be non-zero.
  google.protobuf.Duration duration = 3 [
    (gogoproto.nullable) = false,
    (gogoproto.stdduration) = true,
    (gogoproto.jsontag) = "duration,omitempty",
    (gogoproto.moretags) = "yaml:\"duration\""
  ];
  // current_epoch is the current epoch number, or in other words,
  // how many times has the timer 'ticked'.
  // The first tick (current_epoch=1) is defined as
  // the first block whose blocktime is greater than the EpochInfo start_time.
  int64 current_epoch = 4;
  // Describes the start time of the current timer interval.
  // The interval is (current_epoch_start_time, current_epoch_start_time +
  // duration] When the timer ticks, this is set to current_epoch_start_time =
  // last_epoch_start_time + duration only one timer tick for a given identifier
  // can occur per block.
  //
  // NOTE! The current_epoch_start_time may diverge significantly from the
  // wall-clock time the epoch began at. Wall-clock time of epoch start may be
  // >> current_epoch_start_time. Suppose current_epoch_start_time = 10,
  // duration = 5. Suppose the chain goes offline at t=14, and comes back online
  // at t=30, and produces blocks at every successive time. (t=31, 32, etc.)
  // * The t=30 block will start the epoch for (10, 15]
  // * The t=31 block will start the epoch for (15, 20]
  // * The t=32 block will start the epoch for (20, 25]
  // * The t=33 block will start the epoch for (25, 30]
  // * The t=34 block will start the epoch for (30, 35]
  // * The **t=36** block will start the epoch for (35, 40]
  google.protobuf.Timestamp current_epoch_start_time = 5 [
    (gogoproto.stdtime) = true,
    (gogoproto.nullable) = false,
    (gogoproto.moretags) = "yaml:\"current_epoch_start_time\""
  ];
  // epoch_counting_started is a boolean, that indicates whether this
  // epoch timer has began yet.
  bool epoch_counting_started = 6;
  reserved 7;
  // This is the block height at which the current epoch started. (The block
  // height at which the timer last ticked)
  int64 current_epoch_start_height = 8;
}
```

## Events

The `epochs` module emits the following events:

### BeginBlocker

|  Type          | Attribute Key |  Attribute Value |
|  --------------| ---------------| -----------------|
|  epoch\_start |  epoch\_number |  {epoch\_number} |
|  epoch\_start |  start\_time   |  {start\_time} |

### EndBlocker

|  Type        | Attribute Key  | Attribute Value |
|  ------------| ---------------| -----------------|
|  epoch\_end  | epoch\_number  | {epoch\_number} |

## Keepers

### Keeper functions

Epochs keeper module provides utility functions to manage epochs.

```go
// Keeper is the interface for lockup module keeper
type Keeper interface {
  // GetEpochInfo returns epoch info by identifier
  GetEpochInfo(ctx sdk.Context, identifier string) types.EpochInfo
  // SetEpochInfo set epoch info
  SetEpochInfo(ctx sdk.Context, epoch types.EpochInfo) 
  // DeleteEpochInfo delete epoch info
  DeleteEpochInfo(ctx sdk.Context, identifier string)
  // IterateEpochInfo iterate through epochs
  IterateEpochInfo(ctx sdk.Context, fn func(index int64, epochInfo types.EpochInfo) (stop bool))
  // Get all epoch infos
  AllEpochInfos(ctx sdk.Context) []types.EpochInfo
}
```

## Hooks

```go
  // the first block whose timestamp is after the duration is counted as the end of the epoch
  AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64)
  // new epoch is next block of epoch end block
  BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64)
```

### How modules receive hooks

On hook receiver function of other modules, they need to filter
`epochIdentifier` and only do executions for only specific
epochIdentifier. Filtering epochIdentifier could be in `Params` of other
modules so that they can be modified by governance.

This is the standard dev UX of this:
```golang
func (k MyModuleKeeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
    params := k.GetParams(ctx)
    if epochIdentifier == params.DistrEpochIdentifier {
    // my logic
  }
}
```

### Panic isolation

If a given epoch hook panics, its state update is reverted, but we keep
proceeding through the remaining hooks. This allows more advanced epoch
logic to be used, without concern over state machine halting, or halting
subsequent modules.

This does mean that if there is behavior you expect from a prior epoch
hook, and that epoch hook reverted, your hook may also have an issue. So
do keep in mind "what if a prior hook didn't get executed" in the safety
checks you consider for a new epoch hook.

## Queries

Epochs module is providing below queries to check the module's state.

```protobuf
service Query {
  // EpochInfos provide running epochInfos
  rpc EpochInfos(QueryEpochsInfoRequest) returns (QueryEpochsInfoResponse) {}
  // CurrentEpoch provide current epoch of specified identifier
  rpc CurrentEpoch(QueryCurrentEpochRequest) returns (QueryCurrentEpochResponse) {}
}
```

### Epoch Infos

Query the currently running epochInfos

```sh
osmosisd query epochs epoch-infos
```
::: details Example

An example output:

```sh
epochs:
- current_epoch: "183"
  current_epoch_start_height: "2438409"
  current_epoch_start_time: "2021-12-18T17:16:09.898160996Z"
  duration: 86400s
  epoch_counting_started: true
  identifier: day
  start_time: "2021-06-18T17:00:00Z"
- current_epoch: "26"
  current_epoch_start_height: "2424854"
  current_epoch_start_time: "2021-12-17T17:02:07.229632445Z"
  duration: 604800s
  epoch_counting_started: true
  identifier: week
  start_time: "2021-06-18T17:00:00Z"
```
:::

### Current Epoch


Query the current epoch by the specified identifier

```sh
osmosisd query epochs current-epoch [identifier]
```

::: details Example

Query the current `day` epoch:

```sh
osmosisd query epochs current-epoch day
```

Which in this example outputs:

```sh
current_epoch: "183"
```

### Downtime Recovery

When the chain is recovering from downtime, and multiple epochs for a given identifier should have been executed,
the chain will