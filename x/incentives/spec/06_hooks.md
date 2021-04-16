<!--
order: 6
-->

# Hooks

In this section we describe the "hooks" that `incentives` module provide for other modules.

If there's no usecase for this, we could ignore this.

```go
	AfterCreatePot(ctx sdk.Context, potId uint64)
	AfterAddToPot(ctx sdk.Context, potId uint64)
	AfterStartDistribution(ctx sdk.Context, potId uint64)
	AfterFinishDistribution(ctx sdk.Context, potId uint64)
	AfterDistribute(ctx sdk.Context, potId uint64)
```
