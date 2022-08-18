package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type GammHooks interface {
	// AfterPoolCreated is called after CreatePool
	AfterPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64)

	// AfterJoinPool is called after JoinPool, JoinSwapExternAmountIn, and JoinSwapShareAmountOut
	AfterJoinPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, enterCoins sdk.Coins, shareOutAmount sdk.Int)

	// AfterExitPool is called after ExitPool, ExitSwapShareAmountIn, and ExitSwapExternAmountOut
	AfterExitPool(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, shareInAmount sdk.Int, exitCoins sdk.Coins)

	// AfterSwap is called after SwapExactAmountIn and SwapExactAmountOut
	AfterSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins)
}

type EpochHooks interface {
	// AfterEpochEnd is called after epoch has passed
	AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64)
	// BeforeEpochStart is called before epoch starts
	BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64)
}

var _ GammHooks = MultiGammHooks{}
var _ EpochHooks = MultiEpochHooks{}

// combine multiple gamm hooks, and epoch hooks all hook functions are run in array sequence.
type MultiGammHooks []GammHooks
type MultiEpochHooks []EpochHooks

// Creates hooks for the Gamm Module.
func NewMultiGammHooks(hooks ...GammHooks) MultiGammHooks {
	return hooks
}

// Creates hooks for the Epoch Module.
func NewMultiEpochHooks(hooks ...EpochHooks) MultiEpochHooks {
	return hooks
}

func (h MultiGammHooks) AfterPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	for i := range h {
		h[i].AfterPoolCreated(ctx, sender, poolId)
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

func (h MultiGammHooks) AfterSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, input sdk.Coins, output sdk.Coins) {
	for i := range h {
		h[i].AfterSwap(ctx, sender, poolId, input, output)
	}
}

func (h MultiEpochHooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	for i := range h {
		h[i].AfterEpochEnd(ctx, epochIdentifier, epochNumber)
	}
}

func (h MultiEpochHooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	for i := range h {
		h[i].BeforeEpochStart(ctx, epochIdentifier, epochNumber)
	}
}
