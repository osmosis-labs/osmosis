package osmoutils

import (
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// This function lets you run the function f, but if theres an error or panic
// drop the state machine change and log the error.
// If there is no error, proceeds as normal (but with some slowdown due to SDK store weirdness)
// Try to avoid usage of iterators in f.
func ApplyFuncIfNoError(ctx sdk.Context, f func(ctx sdk.Context) error) (err error) {
	// Add a panic safeguard
	defer func() {
		if recoveryError := recover(); recoveryError != nil {
			PrintPanicRecoveryError(ctx, recoveryError)
			err = errors.New("panic occurred during execution")
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

// PrintPanicRecoveryError error logs the recoveryError, along with the stacktrace, if it can be parsed.
// If not emits them to stdout.
func PrintPanicRecoveryError(ctx sdk.Context, recoveryError interface{}) {
	errStackTrace := string(debug.Stack())
	switch e := recoveryError.(type) {
	case string:
		ctx.Logger().Error("Recovering from (string) panic: " + e)
	case runtime.Error:
		ctx.Logger().Error("recovered (runtime.Error) panic: " + e.Error())
	case error:
		ctx.Logger().Error("recovered (error) panic: " + e.Error())
	default:
		ctx.Logger().Error("recovered (default) panic. Could not capture logs in ctx, see stdout")
		fmt.Println("Recovering from panic ", recoveryError)
		debug.PrintStack()
		return
	}
	ctx.Logger().Error("stack trace: " + errStackTrace)
}
