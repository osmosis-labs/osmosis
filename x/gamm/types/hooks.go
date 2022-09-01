package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/osmoutils"
)

type GammHooks interface {
	// AfterPoolCreated is called after CreatePool
	AfterPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) error

	// AfterJoinPool is called after JoinPool, JoinSwapExternAmountIn, and JoinSwapShareAmountOut
	AfterJoinPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, enterCoins sdk.Coins, shareOutAmount sdk.Int) error

	// AfterExitPool is called after ExitPool, ExitSwapShareAmountIn, and ExitSwapExternAmountOut
	AfterExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount sdk.Int, exitCoins sdk.Coins) error

	// AfterSwap is called after SwapExactAmountIn and SwapExactAmountOut
	AfterSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) error
}

var _ GammHooks = MultiGammHooks{}

// combine multiple gamm hooks, all hook functions are run in array sequence.
type MultiGammHooks []GammHooks

// Creates hooks for the Gamm Module.
func NewMultiGammHooks(hooks ...GammHooks) MultiGammHooks {
	return hooks
}

func (h MultiGammHooks) AfterPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) error {
	for i := range h {
		wrappedHookFn := func(ctx sdk.Context) error {
			return h[i].AfterPoolCreated(ctx, sender, poolId)
		}
		err := osmoutils.ApplyFuncIfNoError(ctx, wrappedHookFn)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("error in gamm hook %v", err))
		}
	}
	return nil
}

func (h MultiGammHooks) AfterJoinPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, enterCoins sdk.Coins, shareOutAmount sdk.Int) error {
	for i := range h {
		wrappedHookFn := func(ctx sdk.Context) error {
			return h[i].AfterJoinPool(ctx, sender, poolId, enterCoins, shareOutAmount)
		}
		err := osmoutils.ApplyFuncIfNoError(ctx, wrappedHookFn)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("error in gamm hook %v", err))
		}
	}
	return nil
}

func (h MultiGammHooks) AfterExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount sdk.Int, exitCoins sdk.Coins) error {
	for i := range h {
		wrappedHookFn := func(ctx sdk.Context) error {
			return h[i].AfterExitPool(ctx, sender, poolId, shareInAmount, exitCoins)
		}
		err := osmoutils.ApplyFuncIfNoError(ctx, wrappedHookFn)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("error in gamm hook %v", err))
		}
	}
	return nil
}

func (h MultiGammHooks) AfterSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) error {
	for i := range h {
		wrappedHookFn := func(ctx sdk.Context) error {
			return h[i].AfterSwap(ctx, sender, poolId, input, output)
		}
		err := osmoutils.ApplyFuncIfNoError(ctx, wrappedHookFn)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("error in gamm hook %v", err))
		}
	}
	return nil
}
