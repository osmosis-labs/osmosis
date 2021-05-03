<!--
order: 0
title: "Epochs Overview"
parent:
  title: "epochs"
-->

# `epochs`

## Abstract

Often in the SDK, we would like to run certain code every-so often. The purpose of `epochs` module is to allow other modules to set that they would like to be signaled once every period. So another module can specify it wants to execute code once a week, starting at UTC-time = x. `epochs` creates a generalized epoch interface to other modules so that they can easily be signalled upon such events.

## Implementation

### Keeper functions
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

### Hooks
```go
  // the first block whose timestamp is after the duration is counted as the end of the epoch
  AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64)
  // new epoch is next block of epoch end block
  BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64)
```

### How modules receive hooks

On hook receiver function of other modules, they need to filter `epochIdentifier` and only do executions for only specific epochIdentifier.
Filtering epochIdentifier could be in `Params` of other modules so that they can be modified by governance.
Governance can change epoch from `weekly` to `daily` as their need.

TODO for postlaunch: We should see if we can architect things such that the receiver doesn't have to do this filtering, and the epochs module would pre-filter for them.
