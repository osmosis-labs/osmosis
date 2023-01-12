package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
	cltypes "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

// OrderInitialPoolDenoms sets the pool denoms of a cl pool
func OrderInitialPoolDenoms(denom0, denom1 string) (string, string, error) {
	return cltypes.OrderInitialPoolDenoms(denom0, denom1)
}

func (k Keeper) CreatePosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, amount0Desired, amount1Desired, amount0Min, amount1Min sdk.Int, lowerTick, upperTick int64) (amtDenom0, amtDenom1 sdk.Int, liquidityCreated sdk.Dec, err error) {
	return k.createPosition(ctx, poolId, owner, amount0Desired, amount1Desired, amount0Min, amount1Min, lowerTick, upperTick)
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

func (k Keeper) SendCoinsBetweenPoolAndUser(ctx sdk.Context, denom0, denom1 string, amount0, amount1 sdk.Int, sender, receiver sdk.AccAddress) error {
	return k.sendCoinsBetweenPoolAndUser(ctx, denom0, denom1, amount0, amount1, sender, receiver)
}

func (k Keeper) CalcInAmtGivenOutInternal(ctx sdk.Context, desiredTokenOut sdk.Coin, tokenInDenom string, swapFee sdk.Dec, priceLimit sdk.Dec, poolId uint64) (writeCtx func(), tokenIn, tokenOut sdk.Coin, updatedTick sdk.Int, updatedLiquidity, updatedSqrtPrice sdk.Dec, err error) {
	return k.calcInAmtGivenOut(ctx, desiredTokenOut, tokenInDenom, swapFee, priceLimit, poolId)
}

func (k Keeper) CalcOutAmtGivenInInternal(ctx sdk.Context, tokenInMin sdk.Coin, tokenOutDenom string, swapFee sdk.Dec, priceLimit sdk.Dec, poolId uint64) (writeCtx func(), tokenIn, tokenOut sdk.Coin, updatedTick sdk.Int, updatedLiquidity, updatedSqrtPrice sdk.Dec, err error) {
	return k.calcOutAmtGivenIn(ctx, tokenInMin, tokenOutDenom, swapFee, priceLimit, poolId)
}

func (k Keeper) InitOrUpdateTick(ctx sdk.Context, poolId uint64, tickIndex int64, liquidityIn sdk.Dec, upper bool) (err error) {
	return k.initOrUpdateTick(ctx, poolId, tickIndex, liquidityIn, upper)
}

func (k Keeper) InitOrUpdatePosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, liquidityDelta sdk.Dec) (err error) {
	return k.initOrUpdatePosition(ctx, poolId, owner, lowerTick, upperTick, liquidityDelta)
}

func (k Keeper) PoolExists(ctx sdk.Context, poolId uint64) bool {
	return k.poolExists(ctx, poolId)
}

func (k Keeper) IsInitialPosition(initialSqrtPrice sdk.Dec, initialTick sdk.Int) bool {
	return k.isInitialPosition(initialSqrtPrice, initialTick)
}

func (k Keeper) InitializeInitialPosition(ctx sdk.Context, pool types.ConcentratedPoolExtension, amount0Desired, amount1Desired sdk.Int) error {
	return k.initializeInitialPosition(ctx, pool, amount0Desired, amount1Desired)
}

func ConvertConcentratedToPoolInterface(concentratedPool types.ConcentratedPoolExtension) (swaproutertypes.PoolI, error) {
	return convertConcentratedToPoolInterface(concentratedPool)
}

func ConvertPoolInterfaceToConcentrated(poolI swaproutertypes.PoolI) (types.ConcentratedPoolExtension, error) {
	return convertPoolInterfaceToConcentrated(poolI)
}
