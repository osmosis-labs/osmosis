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

var emptyCoins = sdk.DecCoins(nil)

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

// initializeFeeAccumulatorPosition initializes the fee accumulator for a given position in a pool
// by creating a new accumulator for the position with zero liquidity and an accumulator value
// equal to the difference between the current fee accumulator value and the fee growth outside of the tick range.
//
// Returns nil on success. Returns error if:
// - fails to get an accumulator for a given pool id
// - attempts to re-initialize an existing fee accumulator liquidity position
// - fails to create a position
func (k Keeper) initializeFeeAccumulatorPosition(ctx sdk.Context, poolId uint64, lowerTick, upperTick int64, positionId uint64) error {
	// Get the fee accumulator for the given pool.
	feeAccumulator, err := k.getFeeAccumulator(ctx, poolId)
	if err != nil {
		return err
	}

	// Get the key for the position's accumulator in the fee accumulator.
	positionKey := cltypes.KeyFeePositionAccumulator(positionId)

	// Check if the position already exists in the fee accumulator and has non-zero liquidity.
	hasPosition, err := feeAccumulator.HasPosition(positionKey)
	if err != nil {
		return err
	}
	if hasPosition {
		return fmt.Errorf("attempted to re-initialize fee accumulator position (%s) with non-zero liquidity", positionKey)
	}

	// Get the fee growth outside of the tick range for the position's pool and ticks.
	feeGrowthOutside, err := k.getFeeGrowthOutside(ctx, poolId, lowerTick, upperTick)
	if err != nil {
		return err
	}

	// Initialize the owner's position with zero liquidity and an accumulator value
	// equal to the difference between the current fee accumulator value and the fee growth outside of the tick range.
	customAccumulatorValue := feeAccumulator.GetValue().Sub(feeGrowthOutside)
	if err := feeAccumulator.NewPositionCustomAcc(positionKey, sdk.ZeroDec(), customAccumulatorValue, nil); err != nil {
		return err
	}

	return nil
}

// updateFeeAccumulatorPosition updates the fee accumulator for a given position
// by calculating the unclaimed rewards and setting the position's initialFeeAccumulatorValue
// to the current fee accumulator value minus the fee growth outside of the position's tick range.
func (k Keeper) updateFeeAccumulatorPosition(ctx sdk.Context, liquidityDelta sdk.Dec, positionId uint64) error {
	// Get the position with the given ID.
	position, err := k.GetPosition(ctx, positionId)
	if err != nil {
		return err
	}

	// Get the fee growth outside of the tick range for the position's pool and ticks.
	feeGrowthOutside, err := k.getFeeGrowthOutside(ctx, position.PoolId, position.LowerTick, position.UpperTick)
	if err != nil {
		return err
	}

	// Get the fee accumulator for the position's pool.
	feeAccumulator, err := k.getFeeAccumulator(ctx, position.PoolId)
	if err != nil {
		return err
	}

	// Get the key for the position's accumulator in the fee accumulator.
	positionKey := cltypes.KeyFeePositionAccumulator(positionId)

	// Replace the position's accumulator in the fee accumulator with a new one
	// that has the latest fee growth outside of the tick range.
	err = preparePositionAccumulator(feeAccumulator, positionKey, feeGrowthOutside)
	if err != nil {
		return err
	}

	// Calculate the unclaimed rewards for the position by subtracting the fee growth outside
	// of the tick range from the current fee accumulator value.
	customAccumulatorValue := feeAccumulator.GetValue().Sub(feeGrowthOutside)

	// Update the position's initialFeeAccumulatorValue in the fee accumulator with the calculated value,
	// taking into account the change in liquidity of the position.
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

// collectFees collects the fees earned by a position and sends them to the owner's account.
// Returns error if the position with the given id does not exist or if fails to get the fee accumulator.
func (k Keeper) collectFees(ctx sdk.Context, owner sdk.AccAddress, positionId uint64) (sdk.Coins, error) {
	// Get the position with the given ID.
	position, err := k.GetPosition(ctx, positionId)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Get the fee accumulator for the position's pool.
	feeAccumulator, err := k.getFeeAccumulator(ctx, position.PoolId)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Get the key for the position's accumulator in the fee accumulator.
	positionKey := cltypes.KeyFeePositionAccumulator(positionId)

	// Check if the position exists in the fee accumulator.
	hasPosition, err := feeAccumulator.HasPosition(positionKey)
	if err != nil {
		return sdk.Coins{}, err
	}
	if !hasPosition {
		return sdk.Coins{}, cltypes.PositionIdNotFoundError{PositionId: positionId}
	}

	// Compute the fee growth outside of the range between the position's lower and upper ticks.
	feeGrowthOutside, err := k.getFeeGrowthOutside(ctx, position.PoolId, position.LowerTick, position.UpperTick)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Prepare the position's accumulator for claiming rewards and claim the rewards.
	feesClaimed, err := prepareAccumAndClaimRewards(feeAccumulator, positionKey, feeGrowthOutside)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Send the claimed fees from the pool's address to the owner's address.
	pool, err := k.getPoolById(ctx, position.PoolId)
	if err != nil {
		return sdk.Coins{}, err
	}
	if err := k.bankKeeper.SendCoins(ctx, pool.GetAddress(), owner, feesClaimed); err != nil {
		return sdk.Coins{}, err
	}

	// Emit an event for the fees collected.
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			cltypes.TypeEvtCollectFees,
			sdk.NewAttribute(sdk.AttributeKeyModule, cltypes.AttributeValueCategory),
			sdk.NewAttribute(cltypes.AttributeKeyPositionId, strconv.FormatUint(positionId, 10)),
			sdk.NewAttribute(cltypes.AttributeKeyTokensOut, feesClaimed.String()),
		),
	})

	return feesClaimed, nil
}

// queryClaimableFees returns the amount of fees that a position is eligible to claim.
//
// Returns error if:
// - pool with the given id does not exist
// - position given by pool id, owner, lower tick and upper tick does not exist
// - other internal database or math errors.
func (k Keeper) queryClaimableFees(ctx sdk.Context, positionId uint64) (sdk.Coins, error) {
	// Since this is a query, we don't want to modify the state and therefore use a cache context.
	cacheCtx, _ := ctx.CacheContext()

	// Get the position with the given ID.
	position, err := k.GetPosition(cacheCtx, positionId)
	if err != nil {
		return nil, err
	}

	// Get the fee accumulator for the position's pool.
	feeAccumulator, err := k.getFeeAccumulator(cacheCtx, position.PoolId)
	if err != nil {
		return nil, err
	}

	// Get the key for the position's accumulator in the fee accumulator.
	positionKey := cltypes.KeyFeePositionAccumulator(positionId)

	// Check if the position exists in the fee accumulator.
	hasPosition, err := feeAccumulator.HasPosition(positionKey)
	if err != nil {
		return nil, err
	}
	if !hasPosition {
		return nil, cltypes.PositionIdNotFoundError{PositionId: positionId}
	}

	// Compute the fee growth outside of the range between the position's lower and upper ticks.
	feeGrowthOutside, err := k.getFeeGrowthOutside(cacheCtx, position.PoolId, position.LowerTick, position.UpperTick)
	if err != nil {
		return nil, err
	}

	// Replace the position's accumulator before calculating unclaimed rewards.
	err = preparePositionAccumulator(feeAccumulator, positionKey, feeGrowthOutside)
	if err != nil {
		return nil, err
	}

	// Claim the position's fees.
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
