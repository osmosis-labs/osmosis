<!--
order: 6
-->

# Hooks

## Sending hooks

In this section we describe the "hooks" that `incentives` module provide for other modules.

If there's no usecase for this, we could ignore this.

```go
	AfterCreateGauge(ctx sdk.Context, gaugeId uint64)
	AfterAddToGauge(ctx sdk.Context, gaugeId uint64)
	AfterStartDistribution(ctx sdk.Context, gaugeId uint64)
	AfterFinishDistribution(ctx sdk.Context, gaugeId uint64)
	AfterDistribute(ctx sdk.Context, gaugeId uint64)
```

## Receiving hooks

`incentives` module receives `epoch` module's `AfterEpochEnd` hook to distribute rewards from the active gauges to lockup owners.

```go
	AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64)
```

To reduce sell pressure, specific percentage of distributed OSMO rewards are automatically staked to a validator.
If user didn't configure his preferred validator address, the amount of OSMO rewards to be staked is locked up on `lockup` module for 2 weeks. To reduce the number of locks, all the rewards that are rewarded to a single address is put on a single `PeriodLock`.
