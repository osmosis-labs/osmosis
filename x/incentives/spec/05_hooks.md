<!--
order: 6
-->

# Hooks

In this section we describe the "hooks" that `incentives` module provide for other modules.

If there's no usecase for this, we could ignore this.

```go
	AfterCreateGauge(ctx sdk.Context, gaugeId uint64)
	AfterAddToGauge(ctx sdk.Context, gaugeId uint64)
	AfterStartDistribution(ctx sdk.Context, gaugeId uint64)
	AfterFinishDistribution(ctx sdk.Context, gaugeId uint64)
	AfterDistribute(ctx sdk.Context, gaugeId uint64)
```

## Distribution

Distribution is done on epochs when `epochs` module trigger an end of epoch.
Now, it is designed for only LP stakers who has not started unlocking yet can get rewards.
