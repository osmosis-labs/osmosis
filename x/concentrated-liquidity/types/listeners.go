package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type ConcentratedLiquidityListener interface {
	// AfterConcentratedPoolCreated runs after a concentrated liquidity poos is initialized.
	AfterConcentratedPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64)
	// AfterInitialPoolPositionCreated is called after the first position is created in a concentrated
	// liquidity pool.
	AfterInitialPoolPositionCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64)
	// AfterLastPoolPositionRemoved is called after the last position is removed in a concentrated
	// liquidity pool.
	AfterLastPoolPositionRemoved(ctx sdk.Context, sender sdk.AccAddress, poolId uint64)
	// AfterConcentratedPoolSwap is called after a swap in a concentrated liquidity pool.
	AfterConcentratedPoolSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64)
}

type ConcentratedLiquidityListeners []ConcentratedLiquidityListener

var _ ConcentratedLiquidityListener = &ConcentratedLiquidityListeners{}

func (l ConcentratedLiquidityListeners) AfterConcentratedPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	for i := range l {
		l[i].AfterConcentratedPoolCreated(ctx, sender, poolId)
	}
}

func (l ConcentratedLiquidityListeners) AfterInitialPoolPositionCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	for i := range l {
		l[i].AfterInitialPoolPositionCreated(ctx, sender, poolId)
	}
}

func (l ConcentratedLiquidityListeners) AfterLastPoolPositionRemoved(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	for i := range l {
		l[i].AfterLastPoolPositionRemoved(ctx, sender, poolId)
	}
}

func (l ConcentratedLiquidityListeners) AfterConcentratedPoolSwap(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	for i := range l {
		l[i].AfterConcentratedPoolSwap(ctx, sender, poolId)
	}
}

// Creates hooks for the x/concentrated-liquidity module.
func NewConcentratedLiquidityListeners(listeners ...ConcentratedLiquidityListener) ConcentratedLiquidityListeners {
	return listeners
}
