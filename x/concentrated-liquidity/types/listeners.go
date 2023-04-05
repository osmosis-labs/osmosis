package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type ConcentratedLiquidityListener interface {
	// AfterConcentratedPoolCreated is called after a AfterConcentratedPoolCreated pool is created
	AfterConcentratedPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64)
}

var _ ConcentratedLiquidityListener = ConcentratedLiquidityListeners{}

// combine multiple gamm hooks, all hook functions are run in array sequence.
type ConcentratedLiquidityListeners []ConcentratedLiquidityListener

// Creates listeners for the concentrated liquidity module.
func NewConcentratedLiquidityListeners(listeners ...ConcentratedLiquidityListener) ConcentratedLiquidityListeners {
	return listeners
}

func (l ConcentratedLiquidityListeners) AfterConcentratedPoolCreated(ctx sdk.Context, sender sdk.AccAddress, poolId uint64) {
	for i := range l {
		l[i].AfterConcentratedPoolCreated(ctx, sender, poolId)
	}
}
