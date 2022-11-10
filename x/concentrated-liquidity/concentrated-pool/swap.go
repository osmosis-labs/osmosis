package concentrated_pool

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

// ApplySwap state of pool after swap.
// It specifically overwrites the pool's liquidity, curr tick and the curr sqrt price
func (p *Pool) ApplySwap(ctx sdk.Context, poolId uint64, newLiquidity sdk.Dec, newCurrentTick sdk.Int, newCurrentSqrtPrice sdk.Dec) error {
	// Fixed gas consumption per swap to prevent spam
	ctx.GasMeter().ConsumeGas(gammtypes.BalancerGasFeeForSwap, "cl pool swap computation")

	p.Liquidity = newLiquidity
	p.CurrentTick = newCurrentTick
	p.CurrentSqrtPrice = newCurrentSqrtPrice
	return nil
}
