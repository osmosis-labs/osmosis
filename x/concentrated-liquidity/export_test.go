package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/math"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/model"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
	cltypes "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
)

// OrderInitialPoolDenoms sets the pool denoms of a cl pool
func OrderInitialPoolDenoms(denom0, denom1 string) (string, string, error) {
	return cltypes.OrderInitialPoolDenoms(denom0, denom1)
}

func (k Keeper) CreatePosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, amount0Desired, amount1Desired, amount0Min, amount1Min sdk.Int, lowerTick, upperTick int64) (amtDenom0, amtDenom1 sdk.Int, liquidityCreated sdk.Dec, err error) {
	return k.createPosition(ctx, poolId, owner, amount0Desired, amount1Desired, amount0Min, amount1Min, lowerTick, upperTick)
}

func (k Keeper) SetTickInfo(ctx sdk.Context, poolId uint64, tickIndex int64, tickInfo model.TickInfo) {
	k.setTickInfo(ctx, poolId, tickIndex, tickInfo)
}

func (k Keeper) SetPool(ctx sdk.Context, pool types.ConcentratedPoolExtension) error {
	return k.setPool(ctx, pool)
}

func (k Keeper) WithdrawPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, liquidityAmount sdk.Dec) (amtDenom0, amtDenom1 sdk.Int, err error) {
	return k.withdrawPosition(ctx, poolId, owner, lowerTick, upperTick, liquidityAmount)
}

func (k Keeper) GetPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64) (*model.Position, error) {
	return k.getPosition(ctx, poolId, owner, lowerTick, upperTick)
}

func ComputeSwapStep(sqrtPriceCurrent, sqrtPriceTarget, liquidity, amountRemaining sdk.Dec, zeroForOne bool) (sqrtPriceNext, amountIn, amountOut sdk.Dec) {
	swapStrategy := math.NewSwapStrategy(zeroForOne)
	return swapStrategy.ComputeSwapStep(sqrtPriceCurrent, sqrtPriceTarget, liquidity, amountRemaining)
}

func (k Keeper) HasPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64) bool {
	return k.hasPosition(ctx, poolId, owner, lowerTick, upperTick)
}

func (k Keeper) GetPoolById(ctx sdk.Context, poolId uint64) (types.ConcentratedPoolExtension, error) {
	return k.getPoolById(ctx, poolId)
}

func (k Keeper) CrossTick(ctx sdk.Context, poolId uint64, tickIndex int64) (liquidityDelta sdk.Dec, err error) {
	return k.crossTick(ctx, poolId, tickIndex)
}

func (k Keeper) GetTickInfo(ctx sdk.Context, poolId uint64, tickIndex int64) (tickInfo model.TickInfo, err error) {
	return k.getTickInfo(ctx, poolId, tickIndex)
}
