package concentrated_liquidity

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

const (
	feeAccumPrefix = "fee"
	keySeparator   = "/"
	uintBase       = 10
)

var (
	emptyCoins = sdk.DecCoins(nil)
)

// createFeeAccumulator creates an accumulator object in the store using the given poolId.
// The accumulator is initialized with the default(zero) values.
func (k Keeper) createFeeAccumulator(ctx sdk.Context, poolId uint64) error {
	err := accum.MakeAccumulator(ctx.KVStore(k.storeKey), getFeeAccumulatorName(poolId))
	if err != nil {
		return err
	}
	return nil
}

// getFeeAccumulator gets the fee accumulator object using the given poolOd
// returns error if accumulator for the given poolId does not exist.
func (k Keeper) getFeeAccumulator(ctx sdk.Context, poolId uint64) (accum.AccumulatorObject, error) {
	acc, err := accum.GetAccumulator(ctx.KVStore(k.storeKey), getFeeAccumulatorName(poolId))
	if err != nil {
		return accum.AccumulatorObject{}, err
	}

	return acc, nil
}

// chargeFee charges the given fee on the pool with the given id by updating
// the internal per-pool accumulator that tracks fee growth per one unit of
// liquidity. Returns error if fails to get accumulator.
func (k Keeper) chargeFee(ctx sdk.Context, poolId uint64, feeUpdate sdk.DecCoin) error {
	feeAccumulator, err := k.getFeeAccumulator(ctx, poolId)
	if err != nil {
		return err
	}

	feeAccumulator.AddToAccumulator(sdk.NewDecCoins(feeUpdate))

	return nil
}

// initializeFeeAccumulatorPosition initializes the pool fee accumulator with zero liquidity delta
// and zero value for the accumulator.
// Returns nil on success. Returns error if:
// - fails to get an accumulator for a given poold id
// - attempts to re-initialize an existing fee accumulator liqudity position
// - fails to create a position
func (k Keeper) initializeFeeAccumulatorPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64) error {
	// get fee accumulator for the pool
	feeAccumulator, err := k.getFeeAccumulator(ctx, poolId)
	if err != nil {
		return err
	}

	positionKey := formatFeePositionAccumulatorKey(poolId, owner, lowerTick, upperTick)

	hasPosition, err := feeAccumulator.HasPosition(positionKey)
	if err != nil {
		return err
	}

	// assure that existing position has zero liquidity
	if hasPosition {
		return fmt.Errorf("attempted to re-initialize fee accumulator position (%s) with non-zero liquidity", positionKey)
	}

	feeGrowthOutside, err := k.getFeeGrowthOutside(ctx, poolId, lowerTick, upperTick)
	if err != nil {
		return err
	}

	// initialize the owner's position with zero liquidity and accumulator set to the
	// difference between the current fee accumulator value and the fee growth outside of the tick range
	customAccumulatorValue := feeAccumulator.GetValue().Sub(feeGrowthOutside)
	if err := feeAccumulator.NewPositionCustomAcc(positionKey, sdk.ZeroDec(), customAccumulatorValue, nil); err != nil {
		return err
	}

	return nil
}

// updateFeeAccumulatorPosition updates the fee accumulator position for a given pool, owner, and tick range.
// It retrieves the current fee growth outside of the given tick range and updates the position's accumulator
// with the provided liquidity delta and the retrieved fee growth outside.
func (k Keeper) updateFeeAccumulatorPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, liquidityDelta sdk.Dec, lowerTick int64, upperTick int64) error {
	feeGrowthOutside, err := k.getFeeGrowthOutside(ctx, poolId, lowerTick, upperTick)
	if err != nil {
		return err
	}

	feeAccumulator, err := k.getFeeAccumulator(ctx, poolId)
	if err != nil {
		return err
	}

	positionKey := formatFeePositionAccumulatorKey(poolId, owner, lowerTick, upperTick)

	// replace position's accumulator before calculating unclaimed rewards
	err = preparePositionAccumulator(feeAccumulator, positionKey, feeGrowthOutside)
	if err != nil {
		return err
	}

	// determine unclaimed rewards and set the positions initialFeeAccumulatorValue to the
	// current fee accumulator value minus the fee growth outside of the tick range
	customAccumulatorValue := feeAccumulator.GetValue().Sub(feeGrowthOutside)
	err = feeAccumulator.UpdatePositionCustomAcc(positionKey, liquidityDelta, customAccumulatorValue)
	if err != nil {
		return err
	}

	return nil
}

// getFeeGrowthOutside returns fee growth upper tick - fee growth lower tick
func (k Keeper) getFeeGrowthOutside(ctx sdk.Context, poolId uint64, lowerTick, upperTick int64) (sdk.DecCoins, error) {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return sdk.DecCoins{}, err
	}
	currentTick := pool.GetCurrentTick().Int64()

	// get lower, upper tick info
	lowerTickInfo, err := k.getTickInfo(ctx, poolId, lowerTick)
	if err != nil {
		return sdk.DecCoins{}, err
	}
	upperTickInfo, err := k.getTickInfo(ctx, poolId, upperTick)
	if err != nil {
		return sdk.DecCoins{}, err
	}

	poolFeeAccumulator, err := k.getFeeAccumulator(ctx, poolId)
	if err != nil {
		return sdk.DecCoins{}, err
	}
	poolFeeGrowth := poolFeeAccumulator.GetValue()

	// calculate fee growth for upper tick and lower tick
	feeGrowthAboveUpperTick := calculateFeeGrowth(upperTick, upperTickInfo.FeeGrowthOutside, currentTick, poolFeeGrowth, true)
	feeGrowthBelowLowerTick := calculateFeeGrowth(lowerTick, lowerTickInfo.FeeGrowthOutside, currentTick, poolFeeGrowth, false)

	return feeGrowthAboveUpperTick.Add(feeGrowthBelowLowerTick...), nil
}

// getInitialFeeGrowthOutsideForTick returns the initial value of fee growth outside for a given tick.
// This value depends on the tick's location relative to the current tick.
//
// feeGrowthOutside =
// { feeGrowthGlobal current tick >= tick }
// { 0               current tick <  tick }
//
// The value is chosen as if all of the fees earned to date had occurrd below the tick.
// Returns error if the pool with the given id does exist or if fails to get the fee accumulator.
func (k Keeper) getInitialFeeGrowthOutsideForTick(ctx sdk.Context, poolId uint64, tick int64) (sdk.DecCoins, error) {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return sdk.DecCoins{}, err
	}

	currentTick := pool.GetCurrentTick().Int64()
	if currentTick >= tick {
		feeAccumulator, err := k.getFeeAccumulator(ctx, poolId)
		if err != nil {
			return sdk.DecCoins{}, err
		}
		return feeAccumulator.GetValue(), nil
	}

	return emptyCoins, nil
}

// collectFees collects fees from the fee accumulator for the position given by pool id, owner, lower tick and upper tick.
// Upon successful collection, it bank sends the fees from the pool address to the owner and returns the collected coins.
// Returns error if:
// - pool with the given id does not exist
// - position given by pool id, owner, lower tick and upper tick does not exist
// - other internal database or math errors.
func (k Keeper) collectFees(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick int64, upperTick int64) (sdk.Coins, error) {
	feeAccumulator, err := k.getFeeAccumulator(ctx, poolId)
	if err != nil {
		return sdk.Coins{}, err
	}

	positionKey := formatFeePositionAccumulatorKey(poolId, owner, lowerTick, upperTick)

	hasPosition, err := feeAccumulator.HasPosition(positionKey)
	if err != nil {
		return sdk.Coins{}, err
	}

	if !hasPosition {
		return sdk.Coins{}, cltypes.PositionNotFoundError{PoolId: poolId, LowerTick: lowerTick, UpperTick: upperTick}
	}

	// compute fee growth outside of the range between lower tick and upper tick.
	feeGrowthOutside, err := k.getFeeGrowthOutside(ctx, poolId, lowerTick, upperTick)
	if err != nil {
		return sdk.Coins{}, err
	}

	// replace position's accumulator before calculating unclaimed rewards
	err = preparePositionAccumulator(feeAccumulator, positionKey, feeGrowthOutside)
	if err != nil {
		return sdk.Coins{}, err
	}

	// claim fees.
	feesClaimed, err := feeAccumulator.ClaimRewards(positionKey)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Check if feeAccumulator was deleted after claiming rewards. If not, we update the custom accumulator value.
	hasPosition, err = feeAccumulator.HasPosition(positionKey)
	if err != nil {
		return sdk.Coins{}, err
	}

	if hasPosition {
		customAccumulatorValue := feeAccumulator.GetValue().Sub(feeGrowthOutside)
		err := feeAccumulator.SetPositionCustomAcc(positionKey, customAccumulatorValue)
		if err != nil {
			return sdk.Coins{}, err
		}
	}

	// Once we have iterated through all the positions, we do a single bank send from the pool to the owner.
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return sdk.Coins{}, err
	}
	if err := k.bankKeeper.SendCoins(ctx, pool.GetAddress(), owner, feesClaimed); err != nil {
		return sdk.Coins{}, err
	}
	return feesClaimed, nil
}

// queryClaimableFees queries the fee accumulator for the position given by pool id, owner, lower tick and upper tick.
// It returns the outstanding fees that can be claimed by the owner.
// Returns error if:
// - pool with the given id does not exist
// - position given by pool id, owner, lower tick and upper tick does not exist
// - other internal database or math errors.
func (k Keeper) queryClaimableFees(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick int64, upperTick int64) (sdk.Coins, error) {
	cacheCtx, _ := ctx.CacheContext()
	feeAccumulator, err := k.getFeeAccumulator(cacheCtx, poolId)
	if err != nil {
		return nil, err
	}

	positionKey := formatFeePositionAccumulatorKey(poolId, owner, lowerTick, upperTick)

	hasPosition, err := feeAccumulator.HasPosition(positionKey)
	if err != nil {
		return nil, err
	}

	if !hasPosition {
		return nil, cltypes.PositionNotFoundError{PoolId: poolId, LowerTick: lowerTick, UpperTick: upperTick}
	}

	// compute fee growth outside of the range between lower tick and upper tick.
	feeGrowthOutside, err := k.getFeeGrowthOutside(cacheCtx, poolId, lowerTick, upperTick)
	if err != nil {
		return nil, err
	}

	// replace position's accumulator before calculating unclaimed rewards
	err = preparePositionAccumulator(feeAccumulator, positionKey, feeGrowthOutside)
	if err != nil {
		return nil, err
	}

	// claim fees.
	feesClaimed, err := feeAccumulator.ClaimRewards(positionKey)
	if err != nil {
		return nil, err
	}

	return feesClaimed, nil
}

func getFeeAccumulatorName(poolId uint64) string {
	poolIdStr := strconv.FormatUint(poolId, uintBase)
	return strings.Join([]string{feeAccumPrefix, poolIdStr}, "/")
}

// calculateFeeGrowth for the given targetTicks.
// If calculating fee growth for an upper tick, we consider the following two cases
// 1. currentTick >= upperTick: If current Tick is GTE than the upper Tick, the fee growth would be pool fee growth - uppertick's fee growth outside
// 2. currentTick < upperTick: If current tick is smaller than upper tick, fee growth would be the upper tick's fee growth outside
// this goes vice versa for calculating fee growth for lower tick.
func calculateFeeGrowth(targetTick int64, feeGrowthOutside sdk.DecCoins, currentTick int64, feesGrowthGlobal sdk.DecCoins, isUpperTick bool) sdk.DecCoins {
	if (isUpperTick && currentTick >= targetTick) || (!isUpperTick && currentTick < targetTick) {
		return feesGrowthGlobal.Sub(feeGrowthOutside)
	}
	return feeGrowthOutside
}

// formatFeePositionAccumulatorKey formats the position's fee accumulator key prefixed by pool id, owner, lower tick
// and upper tick with a key separator in-between.
func formatFeePositionAccumulatorKey(poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64) string {
	return strings.Join([]string{feeAccumPrefix, strconv.FormatUint(poolId, uintBase), owner.String(), strconv.FormatInt(lowerTick, uintBase), strconv.FormatInt(upperTick, uintBase)}, keySeparator)
}

// preparePositionAccumulator is called prior to updating unclaimed rewards,
// as we must set the position's accumulator value to the sum of
// - the fee/uptime growth inside at position creation time (position.InitAccumValue)
// - fee/uptime growth outside at the current block time (feeGrowthOutside/uptimeGrowthOutside)
func preparePositionAccumulator(accumulator accum.AccumulatorObject, positionKey string, growthOutside sdk.DecCoins) error {
	position, err := accum.GetPosition(accumulator, positionKey)
	if err != nil {
		return err
	}

	customAccumulatorValue := position.InitAccumValue.Add(growthOutside...)
	err = accumulator.SetPositionCustomAcc(positionKey, customAccumulatorValue)
	if err != nil {
		return err
	}
	return nil
}
