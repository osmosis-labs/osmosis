package concentrated_liquidity

import (
	db "github.com/tendermint/tm-db"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/internal/math"
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/model"
	types "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
)

// initOrUpdateTick retrieves the tickInfo from the specified tickIndex and updates both the liquidityNet and LiquidityGross.
// if we are initializing or updating an upper tick, we subtract the liquidityIn from the LiquidityNet
// if we are initializing or updating an lower tick, we add the liquidityIn from the LiquidityNet
func (k Keeper) initOrUpdateTick(ctx sdk.Context, poolId uint64, tickIndex int64, liquidityIn sdk.Dec, upper bool) (err error) {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return err
	}

	currentTick := pool.GetCurrentTick().Int64()

	tickInfo, err := k.getTickInfo(ctx, poolId, tickIndex)
	if err != nil {
		return err
	}

	// calculate liquidityGross, which does not care about whether liquidityIn is positive or negative
	liquidityBefore := tickInfo.LiquidityGross

	// note that liquidityIn can be either positive or negative.
	// If negative, this would work as a subtraction from liquidityBefore
	liquidityAfter := math.AddLiquidity(liquidityBefore, liquidityIn)

	tickInfo.LiquidityGross = liquidityAfter

	// calculate liquidityNet, which we take into account and track depending on whether liquidityIn is positive or negative
	if upper {
		tickInfo.LiquidityNet = tickInfo.LiquidityNet.Sub(liquidityIn)
	} else {
		tickInfo.LiquidityNet = tickInfo.LiquidityNet.Add(liquidityIn)
	}

	// if given tickIndex is LTE to current tick, tick's fee growth outside is set as fee accumulator's value
	if tickIndex <= currentTick {
		accum, err := k.getFeeAccumulator(ctx, poolId)
		if err != nil {
			return err
		}

		tickInfo.FeeGrowthOutside = accum.GetValue()
	}

	k.SetTickInfo(ctx, poolId, tickIndex, tickInfo)

	return nil
}

func (k Keeper) crossTick(ctx sdk.Context, poolId uint64, tickIndex int64) (liquidityDelta sdk.Dec, err error) {
	tickInfo, err := k.getTickInfo(ctx, poolId, tickIndex)
	if err != nil {
		return sdk.Dec{}, err
	}

	accum, err := k.getFeeAccumulator(ctx, poolId)
	if err != nil {
		return sdk.Dec{}, err
	}

	// subtract tick's fee growth outside from current fee accumulator
	tickInfo.FeeGrowthOutside = accum.GetValue().Sub(tickInfo.FeeGrowthOutside)
	k.SetTickInfo(ctx, poolId, tickIndex, tickInfo)

	return tickInfo.LiquidityNet, nil
}

// getTickInfo gets tickInfo given poolId and tickIndex. Returns a boolean field that returns true if value is found for given key.
func (k Keeper) getTickInfo(ctx sdk.Context, poolId uint64, tickIndex int64) (tickInfo model.TickInfo, err error) {
	store := ctx.KVStore(k.storeKey)
	tickStruct := model.TickInfo{}
	key := types.KeyTick(poolId, tickIndex)
	if !k.poolExists(ctx, poolId) {
		return model.TickInfo{}, types.PoolNotFoundError{PoolId: poolId}
	}

	found, err := osmoutils.Get(store, key, &tickStruct)
	// return 0 values if key has not been initialized
	if !found {
		// If tick has not yet been initialized, we create a new one and initialize
		// the fee growth outside.
		initialFeeGrowthOutside, err := k.getInitialFeeGrowthOutsideForTick(ctx, poolId, tickIndex)
		if err != nil {
			return tickStruct, err
		}

		return model.TickInfo{LiquidityGross: sdk.ZeroDec(), LiquidityNet: sdk.ZeroDec(), FeeGrowthOutside: initialFeeGrowthOutside}, nil
	}
	if err != nil {
		return tickStruct, err
	}

	return tickStruct, nil
}

func (k Keeper) SetTickInfo(ctx sdk.Context, poolId uint64, tickIndex int64, tickInfo model.TickInfo) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyTick(poolId, tickIndex)
	osmoutils.MustSet(store, key, &tickInfo)
}

// iteratorBetweenTicks returns a store iterator that ranges between the given lowerTick and upperTick in the pool.
func (k Keeper) iteratorBetweenTicks(ctx sdk.Context, poolId uint64, lowerTick, upperTick int64) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	lowerKey := types.KeyTick(poolId, lowerTick)
	upperKey := types.KeyTick(poolId, upperTick)

	return store.Iterator(storetypes.InclusiveEndBytes(lowerKey), storetypes.InclusiveEndBytes(upperKey))
}

// getTickLiquidityDepthFromIterator uses the given iterator, iterates over ticks, creates and returns LiquidityDepth array.
func (k Keeper) getTickLiquidityDepthFromIterator(ctx sdk.Context, poolId uint64, iterator db.Iterator) []types.LiquidityDepth {
	liquidityDepths := []types.LiquidityDepth{}
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		tickIndex := sdk.BigEndianToUint64(iterator.Value())

		tickInfo, err := k.getTickInfo(ctx, poolId, int64(tickIndex))
		if err != nil {
			panic(err)
		}
		liquidityDepth := types.LiquidityDepth{
			TickIndex:    sdk.NewInt(int64(tickIndex)),
			LiquidityNet: tickInfo.LiquidityNet,
		}

		liquidityDepths = append(liquidityDepths, liquidityDepth)
	}

	return liquidityDepths
}

// validateTickInRangeIsValid validates that given ticks are valid.
// That is, both lower and upper ticks are within MinTick and MaxTick range for the given exponentAtPriceOne.
// Also, lower tick must be less than upper tick.
// Returns error if validation fails. Otherwise, nil.
// TODO: test
func validateTickRangeIsValid(tickSpacing uint64, exponentAtPriceOne sdk.Int, lowerTick int64, upperTick int64) error {
	minTick, maxTick := GetMinAndMaxTicksFromExponentAtPriceOne(exponentAtPriceOne)
	// Check if the lower and upper tick values are divisible by the tick spacing.
	if lowerTick%int64(tickSpacing) != 0 || upperTick%int64(tickSpacing) != 0 {
		return types.TickSpacingError{LowerTick: lowerTick, UpperTick: upperTick, TickSpacing: tickSpacing}
	}

	// Check if the lower tick value is within the valid range of MinTick to MaxTick.
	if lowerTick < minTick || lowerTick >= maxTick {
		return types.InvalidTickError{Tick: lowerTick, IsLower: true, MinTick: minTick, MaxTick: maxTick}
	}

	// Check if the upper tick value is within the valid range of MinTick to MaxTick.
	if upperTick > maxTick || upperTick <= minTick {
		return types.InvalidTickError{Tick: upperTick, IsLower: false, MinTick: minTick, MaxTick: maxTick}
	}

	// Check if the lower tick value is greater than or equal to the upper tick value.
	if lowerTick >= upperTick {
		return types.InvalidLowerUpperTickError{LowerTick: lowerTick, UpperTick: upperTick}
	}
	return nil
}

// GetMinAndMaxTicksFromExponentAtPriceOne determines min and max ticks allowed for a given exponentAtPriceOne value
// This allows for a min spot price of 0.000000000000000001 and a max spot price of 100000000000000000000000000000000000000 for every exponentAtPriceOne value
func GetMinAndMaxTicksFromExponentAtPriceOne(exponentAtPriceOne sdk.Int) (minTick, maxTick int64) {
	return math.GetMinAndMaxTicksFromExponentAtPriceOneInternal(exponentAtPriceOne)
}
