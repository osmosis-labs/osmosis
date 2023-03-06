package concentrated_liquidity

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

const (
	uptimeAccumPrefix = "uptime"
)

// createUptimeAccumulators creates accumulator objects in store for each supported uptime for the given poolId.
// The accumulators are initialized with the default (zero) values.
func (k Keeper) createUptimeAccumulators(ctx sdk.Context, poolId uint64) error {
	for uptimeIndex := range types.SupportedUptimes {
		err := accum.MakeAccumulator(ctx.KVStore(k.storeKey), getUptimeAccumulatorName(poolId, uint64(uptimeIndex)))
		if err != nil {
			return err
		}
	}

	return nil
}

func getUptimeAccumulatorName(poolId uint64, uptimeIndex uint64) string {
	poolIdStr := strconv.FormatUint(poolId, uintBase)
	uptimeIndexStr := strconv.FormatUint(uptimeIndex, uintBase)
	return strings.Join([]string{uptimeAccumPrefix, poolIdStr, uptimeIndexStr}, "/")
}

// getUptimeTrackerValues extracts the values of an array of uptime trackers
func getUptimeTrackerValues(uptimeTrackers []model.UptimeTracker) []sdk.DecCoins {
	trackerValues := []sdk.DecCoins{}
	for _, uptimeTracker := range uptimeTrackers {
		trackerValues = append(trackerValues, uptimeTracker.UptimeGrowthOutside)
	}

	return trackerValues
}

// nolint: unused
// getUptimeAccumulators gets the uptime accumulator objects for the given poolId
// Returns error if accumulator for the given poolId does not exist.
func (k Keeper) getUptimeAccumulators(ctx sdk.Context, poolId uint64) ([]accum.AccumulatorObject, error) {
	accums := make([]accum.AccumulatorObject, len(types.SupportedUptimes))
	for uptimeIndex := range types.SupportedUptimes {
		acc, err := accum.GetAccumulator(ctx.KVStore(k.storeKey), getUptimeAccumulatorName(poolId, uint64(uptimeIndex)))
		if err != nil {
			return []accum.AccumulatorObject{}, err
		}

		accums[uptimeIndex] = acc
	}

	return accums, nil
}

// nolint: unused
// getUptimeAccumulatorValues gets the accumulator values for the supported uptimes for the given poolId
// Returns error if accumulator for the given poolId does not exist.
func (k Keeper) getUptimeAccumulatorValues(ctx sdk.Context, poolId uint64) ([]sdk.DecCoins, error) {
	uptimeAccums, err := k.getUptimeAccumulators(ctx, poolId)
	if err != nil {
		return []sdk.DecCoins{}, err
	}

	uptimeValues := []sdk.DecCoins{}
	for _, uptimeAccum := range uptimeAccums {
		uptimeValues = append(uptimeValues, uptimeAccum.GetValue())
	}

	return uptimeValues, nil
}

// nolint: unused
// getInitialUptimeGrowthOutsidesForTick returns an array of the initial values of uptime growth outside
// for each supported uptime for a given tick. This value depends on the tick's location relative to the current tick.
//
// uptimeGrowthOutside =
// { uptimeGrowthGlobal current tick >= tick }
// { 0                  current tick <  tick }
//
// Similar to fees, by convention the value is chosen as if all of the uptime (seconds per liquidity) to date has
// occurred below the tick.
// Returns error if the pool with the given id does not exist or if fails to get any of the uptime accumulators.
func (k Keeper) getInitialUptimeGrowthOutsidesForTick(ctx sdk.Context, poolId uint64, tick int64) ([]sdk.DecCoins, error) {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return []sdk.DecCoins{}, err
	}

	currentTick := pool.GetCurrentTick().Int64()
	if currentTick >= tick {
		uptimeAccumulatorValues, err := k.getUptimeAccumulatorValues(ctx, poolId)
		if err != nil {
			return []sdk.DecCoins{}, err
		}
		return uptimeAccumulatorValues, nil
	}

	// If currentTick < tick, we return len(SupportedUptimes) empty DecCoins
	emptyUptimeValues := []sdk.DecCoins{}
	for range types.SupportedUptimes {
		emptyUptimeValues = append(emptyUptimeValues, emptyCoins)
	}

	return emptyUptimeValues, nil
}

// nolint: unused
// updateUptimeAccumulatorsToNow syncs all uptime accumulators to be up to date.
// Specifically, it gets the time elapsed since the last update and divides it
// by the qualifying liquidity for each uptime. It then adds this value to the
// respective accumulator and updates relevant time trackers accordingly.
func (k Keeper) updateUptimeAccumulatorsToNow(ctx sdk.Context, poolId uint64) error {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return err
	}

	// Get relevant pool-level values
	poolIncentiveRecords, err := k.GetAllIncentiveRecordsForPool(ctx, poolId)
	if err != nil {
		return err
	}

	uptimeAccums, err := k.getUptimeAccumulators(ctx, poolId)
	if err != nil {
		return err
	}

	// Since our base unit of time is nanoseconds, we divide with truncation by 10^9 (10e8) to get
	// time elapsed in seconds
	timeElapsedNanoSec := sdk.NewDec(int64(ctx.BlockTime().Sub(pool.GetLastLiquidityUpdate())))
	timeElapsedSec := timeElapsedNanoSec.Quo(sdk.NewDec(10e8))

	// If no time has elapsed, this function is a no-op
	if timeElapsedSec.Equal(sdk.ZeroDec()) {
		return nil
	}

	if timeElapsedSec.LT(sdk.ZeroDec()) {
		return fmt.Errorf("Time elapsed cannot be negative.")
	}

	for uptimeIndex, uptimeAccum := range uptimeAccums {
		// Get relevant uptime-level values
		curUptimeDuration := types.SupportedUptimes[uptimeIndex]

		// Qualifying liquidity is the amount of liquidity that satisfies uptime requirements
		qualifyingLiquidity, err := uptimeAccum.GetTotalShares()
		if err != nil {
			return err
		}

		// If there is no qualifying liquidity for the current uptime accumulator, we leave it unchanged
		if qualifyingLiquidity.LT(sdk.OneDec()) {
			continue
		}

		incentivesToAddToCurAccum, updatedPoolRecords, err := calcAccruedIncentivesForAccum(ctx, curUptimeDuration, qualifyingLiquidity, timeElapsedSec, poolIncentiveRecords)
		if err != nil {
			return err
		}

		// Emit incentives to current uptime accumulator
		uptimeAccum.AddToAccumulator(incentivesToAddToCurAccum)

		// Update pool records (stored in state after loop)
		poolIncentiveRecords = updatedPoolRecords
	}

	// Update pool incentive records and LastLiquidityUpdate time in state to reflect emitted incentives
	k.setMultipleIncentiveRecords(ctx, poolIncentiveRecords)
	pool.SetLastLiquidityUpdate(ctx.BlockTime())
	err = k.setPool(ctx, pool)
	if err != nil {
		return err
	}

	return nil
}

// nolint: unused
// calcAccruedIncentivesForAccum calculates IncentivesPerLiquidity to be added to an accum
// Returns the IncentivesPerLiquidity value and an updated list of IncentiveRecords that
// reflect emitted incentives
// Returns error if the qualifying liquidity/time elapsed are zero.
func calcAccruedIncentivesForAccum(ctx sdk.Context, accumUptime time.Duration, qualifyingLiquidity sdk.Dec, timeElapsed sdk.Dec, poolIncentiveRecords []types.IncentiveRecord) (sdk.DecCoins, []types.IncentiveRecord, error) {
	if !qualifyingLiquidity.IsPositive() || !timeElapsed.IsPositive() {
		return sdk.DecCoins{}, []types.IncentiveRecord{}, fmt.Errorf("Qualifying liquidity and time elapsed must both be positive.")
	}

	incentivesToAddToCurAccum := sdk.NewDecCoins()
	for incentiveIndex, incentiveRecord := range poolIncentiveRecords {
		// We consider all incentives matching the current uptime that began emitting before the current blocktime
		if incentiveRecord.StartTime.UTC().Before(ctx.BlockTime().UTC()) && incentiveRecord.MinUptime == accumUptime {
			// Total amount emitted = time elapsed * emission
			totalEmittedAmount := timeElapsed.Mul(incentiveRecord.EmissionRate)

			// Incentives to emit per unit of qualifying liquidity = total emitted / qualifying liquidity
			// Note that we truncate to ensure we do not overdistribute incentives
			incentivesPerLiquidity := totalEmittedAmount.QuoTruncate(qualifyingLiquidity)
			emittedIncentivesPerLiquidity := sdk.NewDecCoinFromDec(incentiveRecord.IncentiveDenom, incentivesPerLiquidity)

			// Ensure that we only emit if there are enough incentives remaining to be emitted
			remainingRewards := poolIncentiveRecords[incentiveIndex].RemainingAmount
			if totalEmittedAmount.LTE(remainingRewards) {
				// Add incentives to accumulator
				incentivesToAddToCurAccum = incentivesToAddToCurAccum.Add(emittedIncentivesPerLiquidity)

				// Update incentive record to reflect the incentives that were emitted
				remainingRewards = remainingRewards.Sub(totalEmittedAmount)

				// Each incentive record should only be modified once
				poolIncentiveRecords[incentiveIndex].RemainingAmount = remainingRewards
			}
		}
	}

	return incentivesToAddToCurAccum, poolIncentiveRecords, nil
}

// nolint: unused
// setIncentiveRecords sets the passed in incentive records in state
func (k Keeper) setIncentiveRecord(ctx sdk.Context, incentiveRecord types.IncentiveRecord) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyIncentiveRecord(incentiveRecord.PoolId, incentiveRecord.IncentiveDenom, incentiveRecord.MinUptime)
	incentiveRecordBody := types.IncentiveRecordBody{
		RemainingAmount: incentiveRecord.RemainingAmount,
		EmissionRate:    incentiveRecord.EmissionRate,
		StartTime:       incentiveRecord.StartTime,
	}
	osmoutils.MustSet(store, key, &incentiveRecordBody)
}

// nolint: unused
// setMultipleIncentiveRecords sets multiple incentive records in state
func (k Keeper) setMultipleIncentiveRecords(ctx sdk.Context, incentiveRecords []types.IncentiveRecord) {
	for _, incentiveRecord := range incentiveRecords {
		k.setIncentiveRecord(ctx, incentiveRecord)
	}
}

// GetIncentiveRecord gets the incentive record corresponding to the passed in values from store
func (k Keeper) GetIncentiveRecord(ctx sdk.Context, poolId uint64, denom string, minUptime time.Duration) (types.IncentiveRecord, error) {
	store := ctx.KVStore(k.storeKey)
	incentiveBodyStruct := types.IncentiveRecordBody{}
	key := types.KeyIncentiveRecord(poolId, denom, minUptime)

	found, err := osmoutils.Get(store, key, &incentiveBodyStruct)
	if err != nil {
		return types.IncentiveRecord{}, err
	}

	if !found {
		return types.IncentiveRecord{}, types.IncentiveRecordNotFoundError{PoolId: poolId, IncentiveDenom: denom, MinUptime: minUptime}
	}

	return types.IncentiveRecord{
		PoolId:          poolId,
		IncentiveDenom:  denom,
		MinUptime:       minUptime,
		RemainingAmount: incentiveBodyStruct.RemainingAmount,
		EmissionRate:    incentiveBodyStruct.EmissionRate,
		StartTime:       incentiveBodyStruct.StartTime,
	}, nil
}

// GetAllIncentiveRecordsForPool gets all the incentive records for poolId
// Returns error if it is unable to retrieve
func (k Keeper) GetAllIncentiveRecordsForPool(ctx sdk.Context, poolId uint64) ([]types.IncentiveRecord, error) {
	return osmoutils.GatherValuesFromStorePrefixWithKeyParser(ctx.KVStore(k.storeKey), types.KeyPoolIncentiveRecords(poolId), ParseFullIncentiveRecordFromBz)
}

// GetUptimeGrowthInsideRange returns the uptime growth within the given tick range for all supported uptimes
func (k Keeper) GetUptimeGrowthInsideRange(ctx sdk.Context, poolId uint64, lowerTick int64, upperTick int64) ([]sdk.DecCoins, error) {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return []sdk.DecCoins{}, err
	}

	// Get global uptime accumulator values
	globalUptimeValues, err := k.getUptimeAccumulatorValues(ctx, poolId)
	if err != nil {
		return []sdk.DecCoins{}, err
	}

	// Get current, lower, and upper ticks
	currentTick := pool.GetCurrentTick().Int64()
	lowerTickInfo, err := k.getTickInfo(ctx, poolId, lowerTick)
	if err != nil {
		return []sdk.DecCoins{}, err
	}
	upperTickInfo, err := k.getTickInfo(ctx, poolId, upperTick)
	if err != nil {
		return []sdk.DecCoins{}, err
	}

	// Calculate uptime growth between lower and upper ticks
	// Note that we regard "within range" to mean [lowerTick, upperTick),
	// inclusive of lowerTick and exclusive of upperTick.
	lowerTickUptimeValues := getUptimeTrackerValues(lowerTickInfo.UptimeTrackers)
	upperTickUptimeValues := getUptimeTrackerValues(upperTickInfo.UptimeTrackers)
	if currentTick < lowerTick {
		// If current tick is below range, we subtract uptime growth of upper tick from that of lower tick
		return osmoutils.SubDecCoinArrays(lowerTickUptimeValues, upperTickUptimeValues)
	} else if currentTick < upperTick {
		// If current tick is within range, we subtract uptime growth of lower and upper tick from global growth
		globalMinusUpper, err := osmoutils.SubDecCoinArrays(globalUptimeValues, upperTickUptimeValues)
		if err != nil {
			return []sdk.DecCoins{}, err
		}

		return osmoutils.SubDecCoinArrays(globalMinusUpper, lowerTickUptimeValues)
	} else {
		// If current tick is above range, we subtract uptime growth of lower tick from that of upper tick
		return osmoutils.SubDecCoinArrays(upperTickUptimeValues, lowerTickUptimeValues)
	}
}

// GetUptimeGrowthOutsideRange returns the uptime growth outside the given tick range for all supported uptimes
func (k Keeper) GetUptimeGrowthOutsideRange(ctx sdk.Context, poolId uint64, lowerTick int64, upperTick int64) ([]sdk.DecCoins, error) {
	globalUptimeValues, err := k.getUptimeAccumulatorValues(ctx, poolId)
	if err != nil {
		return []sdk.DecCoins{}, err
	}

	uptimeGrowthInside, err := k.GetUptimeGrowthInsideRange(ctx, poolId, lowerTick, upperTick)
	if err != nil {
		return []sdk.DecCoins{}, err
	}

	return osmoutils.SubDecCoinArrays(globalUptimeValues, uptimeGrowthInside)
}

// initOrUpdatePositionUptime either adds or updates records for all uptime accumulators `position` qualifies for
func (k Keeper) initOrUpdatePositionUptime(ctx sdk.Context, poolId uint64, position *model.Position, owner sdk.AccAddress, lowerTick, upperTick int64, liquidityDelta sdk.Dec, joinTime time.Time, freezeDuration time.Duration) error {
	// We update accumulators _prior_ to any position-related updates to ensure
	// past rewards aren't distributed to new liquidity. We also update pool's
	// LastLiquidityUpdate here.
	err := k.updateUptimeAccumulatorsToNow(ctx, poolId)
	if err != nil {
		return err
	}

	// Create records for relevant uptime accumulators here.
	uptimeAccumulators, err := k.getUptimeAccumulators(ctx, poolId)
	if err != nil {
		return err
	}

	globalUptimeGrowthInsideRange, err := k.GetUptimeGrowthInsideRange(ctx, poolId, lowerTick, upperTick)
	if err != nil {
		return err
	}

	globalUptimeGrowthOutsideRange, err := k.GetUptimeGrowthOutsideRange(ctx, poolId, lowerTick, upperTick)
	if err != nil {
		return err
	}

	// Loop through uptime accums for all supported uptimes on the pool and init or update position's records
	positionName := string(types.KeyFullPosition(poolId, owner, lowerTick, upperTick, joinTime, freezeDuration))
	for uptimeIndex, uptime := range types.SupportedUptimes {
		// We assume every position update requires the position to be frozen for the
		// min uptime again. Thus, the difference between the position's `freezeDuration`
		// and the blocktime when the update happens should be greater than or equal
		// to the required uptime.

		// TODO: consider replacing BlockTime with a new field, JoinTime, so that post-join updates are not skipped
		if freezeDuration >= uptime {
			curUptimeAccum := uptimeAccumulators[uptimeIndex]

			// If a record does not exist for this uptime accumulator, create a new position.
			// Otherwise, add to existing record.
			recordExists, err := curUptimeAccum.HasPosition(positionName)
			if err != nil {
				return err
			}

			if !recordExists {
				// Since the position should only be entitled to uptime growth within its range, we checkpoint globalUptimeGrowthInsideRange as
				// its accumulator's init value. During the claiming (or, equivalently, position updating) process, we ensure that incentives are
				// not overpaid.
				err = curUptimeAccum.NewPositionCustomAcc(positionName, position.Liquidity, globalUptimeGrowthInsideRange[uptimeIndex], emptyOptions)
				if err != nil {
					return err
				}
			} else {
				// Prep accum since we claim rewards first under the hood before any update (otherwise we would overpay)
				err = preparePositionAccumulator(curUptimeAccum, positionName, globalUptimeGrowthOutsideRange[uptimeIndex])
				if err != nil {
					return err
				}

				// Note that even though "unclaimed rewards" accrue in the accumulator prior to freezeDuration, since position withdrawal
				// and incentive collection are only allowed when current time is past freezeDuration these rewards are not accessible until then.
				err = curUptimeAccum.UpdatePositionCustomAcc(positionName, liquidityDelta, globalUptimeGrowthInsideRange[uptimeIndex])
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
