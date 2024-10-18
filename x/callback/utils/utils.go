package utils

import (
	"fmt"
	"os"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ExecuteWithGasLimit executes a function with a gas limit. Taken from: https://github.com/cosmos/cosmos-sdk/pull/18475
func ExecuteWithGasLimit(ctx sdk.Context, gasLimit uint64, f func(ctx sdk.Context) error) (gasUsed uint64, err error) {
	branchedCtx, commit := ctx.CacheContext()
	// create a new gas meter
	limitedGasMeter := storetypes.NewGasMeter(gasLimit)
	// apply gas meter with limit to branched context
	branchedCtx = branchedCtx.WithGasMeter(limitedGasMeter)
	err = catchOutOfGas(branchedCtx, f)
	// even before checking the error, we want to get the gas used
	// and apply it to the original context.
	gasUsed = limitedGasMeter.GasConsumed()
	ctx.GasMeter().ConsumeGas(gasUsed, "branch")
	// in case of errors, do not commit the branched context
	// return gas used and the error
	if err != nil {
		return gasUsed, err
	}
	// if no error, commit the branched context
	// and return gas used and no error
	commit()
	return gasUsed, nil
}

// catchOutOfGas is a helper function to catch out of gas panics and return them as errors.
func catchOutOfGas(ctx sdk.Context, f func(ctx sdk.Context) error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// we immediately check if it's an out of error gas.
			// if it is not we panic again to propagate it up.
			if _, ok := r.(storetypes.ErrorOutOfGas); !ok {
				_, _ = fmt.Fprintf(os.Stderr, "recovered: %#v", r) // log to stderr
				panic(r)
			}
			err = sdkerrors.ErrOutOfGas
		}
	}()
	return f(ctx)
}
