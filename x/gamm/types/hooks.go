package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type GammHooks interface {
	// AfterCFMMPoolCreated is called after a CFMM pool is created
	AfterCFMMPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64)

	// AfterJoinPool is called after JoinPool, JoinSwapExternAmountIn, and JoinSwapShareAmountOut
	AfterJoinPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, enterCoins sdk.Coins, shareOutAmount sdk.Int)

	// AfterExitPool is called after ExitPool, ExitSwapShareAmountIn, and ExitSwapExternAmountOut
	AfterExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount sdk.Int, exitCoins sdk.Coins)

	// AfterSwap is called after SwapExactAmountIn and SwapExactAmountOut in x/gamm.
	AfterCFMMSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins)
}

var _ GammHooks = MultiGammHooks{}

// combine multiple gamm hooks, all hook functions are run in array sequence.
type MultiGammHooks []GammHooks

// Creates hooks for the Gamm Module.
func NewMultiGammHooks(hooks ...GammHooks) MultiGammHooks {
	return hooks
}

func (h MultiGammHooks) AfterCFMMPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	for i := range h {
		h[i].AfterCFMMPoolCreated(ctx, sender, poolId)
	}
}

func (h MultiGammHooks) AfterJoinPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, enterCoins sdk.Coins, shareOutAmount sdk.Int) {
	for i := range h {
		h[i].AfterJoinPool(ctx, sender, poolId, enterCoins, shareOutAmount)
	}
}

func (h MultiGammHooks) AfterExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount sdk.Int, exitCoins sdk.Coins) {
	for i := range h {
		h[i].AfterExitPool(ctx, sender, poolId, shareInAmount, exitCoins)
	}
}

func (h MultiGammHooks) AfterCFMMSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) {
	for i := range h {
		h[i].AfterCFMMSwap(ctx, sender, poolId, input, output)
	}
}
