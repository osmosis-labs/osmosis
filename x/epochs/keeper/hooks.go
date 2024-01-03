package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AfterEpochEnd gets called at the end of the epoch, end of epoch is the timestamp of first block produced after epoch duration.
func (k Keeper) AfterEpochEnd(ctx sdk.Context, identifier string, epochNumber int64) {
	// time := time.Now().UTC().Unix()

	// // Start CPU profiling
	// f, err := os.Create(fmt.Sprintf("/root/cpu_profile-%d.prof", time))
	// defer f.Close()
	// if err != nil {
	// 	ctx.Logger().With("filter", "epoch").Info("epoch failed", "height", ctx.BlockHeight())
	// 	panic(err)
	// }
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	// Error is not handled as AfterEpochEnd Hooks use osmoutils.ApplyFuncIfNoError()
	_ = k.hooks.AfterEpochEnd(ctx, identifier, epochNumber)

	ctx.Logger().With("filter", "epoch").Info("epoch finished", "height", ctx.BlockHeight())
}

// BeforeEpochStart new epoch is next block of epoch end block
func (k Keeper) BeforeEpochStart(ctx sdk.Context, identifier string, epochNumber int64) {
	// Error is not handled as BeforeEpochStart Hooks use osmoutils.ApplyFuncIfNoError()
	_ = k.hooks.BeforeEpochStart(ctx, identifier, epochNumber)
}
