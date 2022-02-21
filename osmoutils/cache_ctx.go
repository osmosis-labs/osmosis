package osmoutils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// This function lets you run the function f, but if theres an error
// drop the state machine change and log the error.
// If there is no error, proceeds as normal (but with some slowdown due to SDK store weirdness)
// Try to avoid usage of iterators in f.
func ApplyFuncIfNoError(ctx sdk.Context, f func(ctx sdk.Context) error) error {
	// makes a new cache context, which all state changes get wrapped inside of.
	cacheCtx, write := ctx.CacheContext()
	err := f(cacheCtx)
	if err != nil {
		ctx.Logger().Error(err.Error())
		return err
	} else {
		// no error, write the output of f
		write()
		return nil
	}
}
