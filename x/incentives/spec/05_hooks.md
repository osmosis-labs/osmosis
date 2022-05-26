# Hooks

In this section we describe the "hooks" that `incentives` module provide
for other modules.

If there's no usecase for this, we could ignore this.

``` {.go}
 AfterCreateGauge(ctx sdk.Context, gaugeId uint64)
 AfterAddToGauge(ctx sdk.Context, gaugeId uint64)
 AfterStartDistribution(ctx sdk.Context, gaugeId uint64)
 AfterFinishDistribution(ctx sdk.Context, gaugeId uint64)
 AfterDistribute(ctx sdk.Context, gaugeId uint64)
```
