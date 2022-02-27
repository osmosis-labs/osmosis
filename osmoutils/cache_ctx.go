package osmoutils

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// This function lets you run the function f, but if theres an error or panic
// drop the state machine change and log the error.
// If there is no error, proceeds as normal (but with some slowdown due to SDK store weirdness)
// Try to avoid usage of iterators in f.
func ApplyFuncIfNoError(ctx sdk.Context, f func(ctx sdk.Context) error) (err error) {
	// Add a panic safeguard
	defer func() {
		if recovErr := recover(); recovErr != nil {
			fmt.Println(recovErr)
			err = errors.New("panic occured during execution")
		}
	}()
	// makes a new cache context, which all state changes get wrapped inside of.
	cacheCtx, write := ctx.CacheContext()
	err = f(cacheCtx)
	if err != nil {
		ctx.Logger().Error(err.Error())
	} else {
		// no error, write the output of f
		write()
	}
	return err
}
