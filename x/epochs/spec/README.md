<!--
order: 0
title: "Epochs Overview"
parent:
  title: "epochs"
-->

# `epochs`

## Abstract

The purpose of `epochs` module is to provide generalized epoch interface to other modules so that they can easily implement epochs without keeping own code for epochs.

## Implementation

### Keeper functions
```go
// Keeper is the interface for lockup module keeper
type Keeper interface {
	// CreateEpochCounter Returns full balance of the module
  // All of these epoch counters could be set on epochs' genesis
  // if startTime is not set, we could use genesisTime - ctx.BlockTime at the time of InitChain
	CreateEpochCounter(ctx sdk.Context, epochIdentifier string, duration time.Duration, startTime time.Time)
}
```

### Hooks
```go
  onEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64)
  onEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64)
```

### How modules receive hooks

On hook receiver function of other modules, they need to filter `epochIdentifier` and only do executions for only specific epochIdentifier.
Filtering epochIdentifier could be in `Params` of other modules so that they can be modified by governance.
Governance can change epoch from `weekly` to `daily` as their need.