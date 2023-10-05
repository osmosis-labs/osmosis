package concentrated_liquidity

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	sdkprefix "github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"golang.org/x/exp/slices"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
)

// createUptimeAccumulators creates accumulator objects in store for each supported uptime for the given poolId.
// The accumulators are initialized with the default (zero) values.
func (k Keeper) createUptimeAccumulators(ctx sdk.Context, poolId uint64) error {
	for uptimeIndex := range types.SupportedUptimes {
		err := accum.MakeAccumulator(ctx.KVStore(k.storeKey), types.KeyUptimeAccumulator(poolId, uint64(uptimeIndex)))
		if err != nil {
			return err
		}
	}

	return nil
}

// getUptimeTrackerValues extracts the values of an array of uptime trackers
func getUptimeTrackerValues(uptimeTrackers []model.UptimeTracker) []sdk.DecCoins {
	trackerValues := []sdk.DecCoins{}
	for _, uptimeTracker := range uptimeTrackers {
		trackerValues = append(trackerValues, uptimeTracker.UptimeGrowthOutside)
	}

	return trackerValues
}

// GetUptimeAccumulators gets the uptime accumulator objects for the given poolId
// Returns error if accumulator for the given poolId does not exist.
func (k Keeper) GetUptimeAccumulators(ctx sdk.Context, poolId uint64) ([]*accum.AccumulatorObject, error) {
	accums := make([]*accum.AccumulatorObject, len(types.SupportedUptimes))
	for uptimeIndex := range types.SupportedUptimes {
		acc, err := accum.GetAccumulator(ctx.KVStore(k.storeKey), types.KeyUptimeAccumulator(poolId, uint64(uptimeIndex)))
		if err != nil {
			return []*accum.AccumulatorObject{}, err
		}

		accums[uptimeIndex] = acc
	}

	return accums, nil
}

// GetUptimeAccumulatorValues gets the accumulator values for the supported uptimes for the given poolId
// Returns error if accumulator for the given poolId does not exist.
func (k Keeper) GetUptimeAccumulatorValues(ctx sdk.Context, poolId uint64) ([]sdk.DecCoins, error) {
	uptimeAccums, err := k.GetUptimeAccumulators(ctx, poolId)
	if err != nil {
		return []sdk.DecCoins{}, err
	}

	uptimeValues := []sdk.DecCoins{}
	for _, uptimeAccum := range uptimeAccums {
		uptimeValues = append(uptimeValues, uptimeAccum.GetValue())
	}

	return uptimeValues, nil
}

// getInitialUptimeGrowthOppositeDirectionOfLastTraversalForTick returns an array of the initial values
// of uptime growth opposite the direction of last traversal for each supported uptime for a given tick.
// This value depends on the provided tick's location relative to the current tick. If the provided tick
// is greater than the current tick, then the value is zero. Otherwise, the value is the value of the
// current global spread reward growth.
// TODO: Explain that this is true iff this is being called for consecutive ticks, not if we were jump around.
//
// Similar to spread factors, by convention the value is chosen as if all of the uptime (seconds per liquidity) to date has
// occurred below the tick.
// Returns error if the pool with the given id does not exist or if fails to get any of the uptime accumulators.
func (k Keeper) getInitialUptimeGrowthOppositeDirectionOfLastTraversalForTick(ctx sdk.Context, pool types.ConcentratedPoolExtension, tick int64) ([]sdk.DecCoins, error) {
	currentTick := pool.GetCurrentTick()
	if currentTick >= tick {
		uptimeAccumulatorValues, err := k.GetUptimeAccumulatorValues(ctx, pool.GetId())
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

// UpdatePoolUptimeAccumulatorsToNow syncs all uptime accumulators that are refetched from state for the given
// poold id to be up to date for the given pool. Updates the pool last liquidity update time with
// the current block time and writes the updated pool to state.
// Specifically, it gets the time elapsed since the last update and divides it
// by the qualifying liquidity on the active tick. It then adds this value to the
// respective accumulator and updates relevant time trackers accordingly.
// WARNING: this method may mutate the pool, make sure to refetch the pool after calling this method.
// Note: the following are the differences of this function from updateGivenPoolUptimeAccumulatorsToNow:
// * this function fetches the uptime accumulators from state.
// * this function fetches a pool from state by id.
// updateGivenPoolUptimeAccumulatorsToNow is used in swaps for performance reasons to minimize state reads.
// UpdatePoolUptimeAccumulatorsToNow is used in all other cases.
func (k Keeper) UpdatePoolUptimeAccumulatorsToNow(ctx sdk.Context, poolId uint64) error {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return err
	}

	return k.updatePoolUptimeAccumulatorsToNowWithPool(ctx, pool)
}

func (k Keeper) updatePoolUptimeAccumulatorsToNowWithPool(ctx sdk.Context, pool types.ConcentratedPoolExtension) error {
	uptimeAccums, err := k.GetUptimeAccumulators(ctx, pool.GetId())
	if err != nil {
		return err
	}

	if err := k.updateGivenPoolUptimeAccumulatorsToNow(ctx, pool, uptimeAccums); err != nil {
		return err
	}

	return nil
}

var dec1e9 = osmomath.NewDec(1e9)

// updateGivenPoolUptimeAccumulatorsToNow syncs all given uptime accumulators for a given pool id
// Updates the pool last liquidity update time with the current block time and writes the updated pool to state.
// If last liquidity update happened in the current block, this function is a no-op.
// Specifically, it gets the time elapsed since the last update and divides it
// by the qualifying liquidity for each uptime. It then adds this value to the
// respective accumulator and updates relevant time trackers accordingly.
// This method also serves the purpose of correctly splitting rewards between the linked balancer pool and the cl pool.
// CONTRACT: the caller validates that the pool with the given id exists.
// CONTRACT: given uptimeAccums are associated with the given pool id.
// CONTRACT: caller is responsible for the uptimeAccums to be up-to-date.
// WARNING: this method may mutate the pool, make sure to refetch the pool after calling this method.
// Note: the following are the differences of this function from updatePoolUptimeAccumulatorsToNow:
// * this function does not refetch the uptime accumulators from state.
// * this function operates on the given pool directly, instead of fetching it from state.
// This is to avoid unnecessary state reads during swaps for performance reasons.
func (k Keeper) updateGivenPoolUptimeAccumulatorsToNow(ctx sdk.Context, pool types.ConcentratedPoolExtension, uptimeAccums []*accum.AccumulatorObject) error {
	if pool == nil {
		return types.ErrPoolNil
	}

	// Since our base unit of time is nanoseconds, we divide with truncation by 10^9 to get
	// time elapsed in seconds
	timeElapsedNanoSec := osmomath.NewDec(int64(ctx.BlockTime().Sub(pool.GetLastLiquidityUpdate())))
	timeElapsedSec := timeElapsedNanoSec.Quo(dec1e9)

	// If no time has elapsed, this function is a no-op
	if timeElapsedSec.IsZero() {
		return nil
	}

	if timeElapsedSec.IsNegative() {
		return types.TimeElapsedNotPositiveError{TimeElapsed: timeElapsedSec}
	}

	poolId := pool.GetId()

	// Get relevant pool-level values
	poolIncentiveRecords, err := k.GetAllIncentiveRecordsForPool(ctx, poolId)
	if err != nil {
		return err
	}

	// We optimistically assume that all liquidity on the active tick qualifies and handle
	// uptime-related checks in forfeiting logic.

	// If there is no share to be incentivized for the current uptime accumulator, we leave it unchanged
	qualifyingLiquidity := pool.GetLiquidity()
	if !qualifyingLiquidity.LT(osmomath.OneDec()) {
		for uptimeIndex := range uptimeAccums {
			// Get relevant uptime-level values
			curUptimeDuration := types.SupportedUptimes[uptimeIndex]
			incentivesToAddToCurAccum, updatedPoolRecords, err := calcAccruedIncentivesForAccum(ctx, curUptimeDuration, qualifyingLiquidity, timeElapsedSec, poolIncentiveRecords)
			if err != nil {
				return err
			}

			// Emit incentives to current uptime accumulator
			uptimeAccums[uptimeIndex].AddToAccumulator(incentivesToAddToCurAccum)

			// Update pool records (stored in state after loop)
			poolIncentiveRecords = updatedPoolRecords
		}
	}

	// Update pool incentive records and LastLiquidityUpdate time in state to reflect emitted incentives
	err = k.setMultipleIncentiveRecords(ctx, poolIncentiveRecords)
	if err != nil {
		return err
	}

	pool.SetLastLiquidityUpdate(ctx.BlockTime())
	err = k.setPool(ctx, pool)
	if err != nil {
		return err
	}

	return nil
}

// calcAccruedIncentivesForAccum calculates IncentivesPerLiquidity to be added to an accum.
// This function is non-mutative. It operates on and returns an updated _copy_ of the passed in incentives records.
// Returns the IncentivesPerLiquidity value and an updated list of IncentiveRecords that
// reflect emitted incentives
// Returns error if the qualifying liquidity/time elapsed are zero.
func calcAccruedIncentivesForAccum(ctx sdk.Context, accumUptime time.Duration, liquidityInAccum osmomath.Dec, timeElapsed osmomath.Dec, poolIncentiveRecords []types.IncentiveRecord) (sdk.DecCoins, []types.IncentiveRecord, error) {
	if !liquidityInAccum.IsPositive() || !timeElapsed.IsPositive() {
		return sdk.DecCoins{}, []types.IncentiveRecord{}, types.QualifyingLiquidityOrTimeElapsedNotPositiveError{QualifyingLiquidity: liquidityInAccum, TimeElapsed: timeElapsed}
	}

	copyPoolIncentiveRecords := make([]types.IncentiveRecord, len(poolIncentiveRecords))
	copy(copyPoolIncentiveRecords, poolIncentiveRecords)
	incentivesToAddToCurAccum := sdk.NewDecCoins()
	for incentiveIndex, incentiveRecord := range copyPoolIncentiveRecords {
		// We consider all incentives matching the current uptime that began emitting before the current blocktime
		incentiveRecordBody := incentiveRecord.IncentiveRecordBody
		if !incentiveRecordBody.StartTime.UTC().Before(ctx.BlockTime().UTC()) || incentiveRecord.MinUptime != accumUptime {
			// If the incentive does not match the current uptime or has not started emitting, we skip it
			continue
		}

		// Total amount emitted = time elapsed * emission
		totalEmittedAmount := timeElapsed.Mul(incentiveRecordBody.EmissionRate)

		// Incentives to emit per unit of qualifying liquidity = total emitted / liquidityInAccum
		// Note that we truncate to ensure we do not overdistribute incentives
		incentivesPerLiquidity := totalEmittedAmount.QuoTruncate(liquidityInAccum)
		emittedIncentivesPerLiquidity := sdk.NewDecCoinFromDec(incentiveRecordBody.RemainingCoin.Denom, incentivesPerLiquidity)

		// Ensure that we only emit if there are enough incentives remaining to be emitted
		remainingRewards := poolIncentiveRecords[incentiveIndex].IncentiveRecordBody.RemainingCoin.Amount

		// if total amount emitted does not exceed remaining rewards,
		if totalEmittedAmount.LTE(remainingRewards) {
			incentivesToAddToCurAccum = incentivesToAddToCurAccum.Add(emittedIncentivesPerLiquidity)

			// Update incentive record to reflect the incentives that were emitted
			remainingRewards = remainingRewards.Sub(totalEmittedAmount)

			// Each incentive record should only be modified once
			copyPoolIncentiveRecords[incentiveIndex].IncentiveRecordBody.RemainingCoin.Amount = remainingRewards
		} else {
			// If there are not enough incentives remaining to be emitted, we emit the remaining rewards.
			// When the returned records are set in state, all records with remaining rewards of zero will be cleared.
			remainingIncentivesPerLiquidity := remainingRewards.QuoTruncate(liquidityInAccum)
			emittedIncentivesPerLiquidity = sdk.NewDecCoinFromDec(incentiveRecordBody.RemainingCoin.Denom, remainingIncentivesPerLiquidity)
			incentivesToAddToCurAccum = incentivesToAddToCurAccum.Add(emittedIncentivesPerLiquidity)

			copyPoolIncentiveRecords[incentiveIndex].IncentiveRecordBody.RemainingCoin.Amount = osmomath.ZeroDec()
		}
	}

	return incentivesToAddToCurAccum, copyPoolIncentiveRecords, nil
}

// findUptimeIndex finds the uptime index for the passed in min uptime.
// Returns error if uptime index cannot be found.
func findUptimeIndex(uptime time.Duration) (int, error) {
	index := slices.IndexFunc(types.SupportedUptimes, func(e time.Duration) bool { return e == uptime })

	if index == -1 {
		return index, types.InvalidUptimeIndexError{MinUptime: uptime, SupportedUptimes: types.SupportedUptimes}
	}

	return index, nil
}

// setIncentiveRecords sets the passed in incentive records in state
// Errors if the incentive record has an unsupported min uptime.
func (k Keeper) setIncentiveRecord(ctx sdk.Context, incentiveRecord types.IncentiveRecord) error {
	store := ctx.KVStore(k.storeKey)

	uptimeIndex, err := findUptimeIndex(incentiveRecord.MinUptime)
	if err != nil {
		return err
	}

	key := types.KeyIncentiveRecord(incentiveRecord.PoolId, uptimeIndex, incentiveRecord.IncentiveId)
	incentiveRecordBody := incentiveRecord.IncentiveRecordBody

	// If the remaining amount is zero and the record already exists in state, we delete the record from state.
	// If the remaining amount is zero and the record doesn't exist in state, we do a no-op.
	// In all other cases, we update the record in state
	if store.Has(key) && incentiveRecordBody.RemainingCoin.IsZero() {
		store.Delete(key)
	} else if incentiveRecordBody.RemainingCoin.Amount.IsPositive() {
		osmoutils.MustSet(store, key, &incentiveRecordBody)
	}

	return nil
}

// setMultipleIncentiveRecords sets multiple incentive records in state
func (k Keeper) setMultipleIncentiveRecords(ctx sdk.Context, incentiveRecords []types.IncentiveRecord) error {
	for _, incentiveRecord := range incentiveRecords {
		err := k.setIncentiveRecord(ctx, incentiveRecord)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetIncentiveRecord gets the incentive record corresponding to the passed in values from store
func (k Keeper) GetIncentiveRecord(ctx sdk.Context, poolId uint64, minUptime time.Duration, incentiveRecordId uint64) (types.IncentiveRecord, error) {
	store := ctx.KVStore(k.storeKey)
	incentiveBodyStruct := types.IncentiveRecordBody{}

	uptimeIndex, err := findUptimeIndex(minUptime)
	if err != nil {
		return types.IncentiveRecord{}, err
	}

	key := types.KeyIncentiveRecord(poolId, uptimeIndex, incentiveRecordId)

	found, err := osmoutils.Get(store, key, &incentiveBodyStruct)
	if err != nil {
		return types.IncentiveRecord{}, err
	}

	if !found {
		return types.IncentiveRecord{}, types.IncentiveRecordNotFoundError{PoolId: poolId, MinUptime: minUptime, IncentiveRecordId: incentiveRecordId}
	}

	return types.IncentiveRecord{
		PoolId:              poolId,
		MinUptime:           minUptime,
		IncentiveId:         incentiveRecordId,
		IncentiveRecordBody: incentiveBodyStruct,
	}, nil
}

// GetAllIncentiveRecordsForPool gets all the incentive records for poolId
// Returns error if it is unable to retrieve records.
func (k Keeper) GetAllIncentiveRecordsForPool(ctx sdk.Context, poolId uint64) ([]types.IncentiveRecord, error) {
	return osmoutils.GatherValuesFromStorePrefixWithKeyParser(ctx.KVStore(k.storeKey), types.KeyPoolIncentiveRecords(poolId), ParseFullIncentiveRecordFromBz)
}

// GetIncentiveRecordSerialized gets incentive records based on limit set by pagination request.
func (k Keeper) GetIncentiveRecordSerialized(ctx sdk.Context, poolId uint64, pagination *query.PageRequest) ([]types.IncentiveRecord, *query.PageResponse, error) {
	incentivesRecordStore := sdkprefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPoolIncentiveRecords(poolId))

	incentiveRecords := []types.IncentiveRecord{}
	pageRes, err := query.Paginate(incentivesRecordStore, pagination, func(key, value []byte) error {
		parts := bytes.Split(key, []byte(types.KeySeparator))

		minUptimeIndex, err := strconv.ParseUint(string(parts[0]), 10, 64)
		if err != nil {
			return err
		}

		incentiveRecordId, err := strconv.ParseUint(string(parts[1]), 10, 64)
		if err != nil {
			return err
		}

		incRecord, err := k.GetIncentiveRecord(ctx, poolId, types.SupportedUptimes[minUptimeIndex], incentiveRecordId)
		if err != nil {
			return err
		}

		incentiveRecords = append(incentiveRecords, incRecord)

		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return incentiveRecords, pageRes, err
}

// getAllIncentiveRecordsForUptime gets all the incentive records for the given poolId and minUptime
// Returns error if the passed in uptime is not supported or it is unable to retrieve records.
func (k Keeper) getAllIncentiveRecordsForUptime(ctx sdk.Context, poolId uint64, minUptime time.Duration) ([]types.IncentiveRecord, error) {
	// Ensure pool exists in state
	_, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return []types.IncentiveRecord{}, err
	}

	uptimeIndex, err := findUptimeIndex(minUptime)
	if err != nil {
		return []types.IncentiveRecord{}, err
	}

	return osmoutils.GatherValuesFromStorePrefixWithKeyParser(ctx.KVStore(k.storeKey), types.KeyUptimeIncentiveRecords(poolId, uptimeIndex), ParseFullIncentiveRecordFromBz)
}

// GetUptimeGrowthInsideRange returns the uptime growth within the given tick range for all supported uptimes.
// UptimeGrowthInside tracks the incentives accured by a specific LP within a pool. It keeps track of the cumulative amount of incentives
// collected by a specific LP within a pool. This function also measures the growth of incentives accured by a particular LP since the last
// time incentives were collected.
// WARNING: this method may mutate the pool, make sure to refetch the pool after calling this method.
// The mutation occurs in the call to GetTickInfo().
func (k Keeper) GetUptimeGrowthInsideRange(ctx sdk.Context, poolId uint64, lowerTick int64, upperTick int64) ([]sdk.DecCoins, error) {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return []sdk.DecCoins{}, err
	}

	// Get global uptime accumulator values
	globalUptimeValues, err := k.GetUptimeAccumulatorValues(ctx, poolId)
	if err != nil {
		return []sdk.DecCoins{}, err
	}

	// Get current, lower, and upper ticks
	currentTick := pool.GetCurrentTick()
	lowerTickInfo, err := k.GetTickInfo(ctx, poolId, lowerTick)
	if err != nil {
		return []sdk.DecCoins{}, err
	}
	upperTickInfo, err := k.GetTickInfo(ctx, poolId, upperTick)
	if err != nil {
		return []sdk.DecCoins{}, err
	}

	// Calculate uptime growth between lower and upper ticks
	// Note that we regard "within range" to mean [lowerTick, upperTick),
	// inclusive of lowerTick and exclusive of upperTick.
	lowerTickUptimeValues := getUptimeTrackerValues(lowerTickInfo.UptimeTrackers.List)
	upperTickUptimeValues := getUptimeTrackerValues(upperTickInfo.UptimeTrackers.List)
	// If current tick is below range, we subtract uptime growth of upper tick from that of lower tick
	if currentTick < lowerTick {
		// Note: SafeSub with negative accumulation is possible if upper tick is initialized first
		// while current tick > upper tick. Then, the current tick under the lower tick. The lower
		// tick then gets initialized to zero.
		// Therefore, we allow for negative result.
		return osmoutils.SafeSubDecCoinArrays(lowerTickUptimeValues, upperTickUptimeValues)
	} else if currentTick < upperTick {
		// If current tick is within range, we subtract uptime growth of lower and upper tick from global growth
		// Note: each individual tick snapshot never be greater than the global uptime accumulator.
		// Therefore, we do not allow for negative result.
		globalMinusUpper, err := osmoutils.SubDecCoinArrays(globalUptimeValues, upperTickUptimeValues)
		if err != nil {
			return []sdk.DecCoins{}, err
		}

		// Note: SafeSub with negative accumulation is possible if lower tick is initialized after upper tick
		// and the current tick is between the two.
		return osmoutils.SafeSubDecCoinArrays(globalMinusUpper, lowerTickUptimeValues)
	} else {
		// If current tick is above range, we subtract uptime growth of lower tick from that of upper tick
		// Note: SafeSub with negative accumulation is possible if lower tick is initialized after upper tick
		// and the current tick is above the two.
		return osmoutils.SafeSubDecCoinArrays(upperTickUptimeValues, lowerTickUptimeValues)
	}
}

// GetUptimeGrowthOutsideRange returns the uptime growth outside the given tick range for all supported uptimes.
// UptimeGrowthOutside tracks the incentive accured by the entire pool. It keeps track of the cumulative amount of incentives collected
// by a specific pool since the last time incentives were accured.
// We use this function to calculate the total amount of incentives owed to the LPs when they withdraw their liquidity or when they
// attempt to claim their incentives.
// When LPs are ready to claim their incentives we calculate it using: (shares of # of LP) * (uptimeGrowthOutside - uptimeGrowthInside)
func (k Keeper) GetUptimeGrowthOutsideRange(ctx sdk.Context, poolId uint64, lowerTick int64, upperTick int64) ([]sdk.DecCoins, error) {
	globalUptimeValues, err := k.GetUptimeAccumulatorValues(ctx, poolId)
	if err != nil {
		return []sdk.DecCoins{}, err
	}

	uptimeGrowthInside, err := k.GetUptimeGrowthInsideRange(ctx, poolId, lowerTick, upperTick)
	if err != nil {
		return []sdk.DecCoins{}, err
	}

	return osmoutils.SubDecCoinArrays(globalUptimeValues, uptimeGrowthInside)
}

// initOrUpdatePositionUptimeAccumulators either initializes or updates liquidity for uptime position accumulators for every supported uptime.
// It syncs the uptime accumulators to the current block time. If this is a new position, it creates a new position accumulator for every supported uptime accumulator.
// If this is an existing position, it updates the existing position accumulator for every supported uptime accumulator.
// Returns error if:
// - fails to update global uptime accumulators
// - fails to get global uptime accumulators
// - fails to calculate uptime growth inside range
// - fails to calculate uptime growth outside range
// - fails to determine if position accumulator is new or existing
// - fails to create/update position uptime accumulators
// WARNING: this method may mutate the pool, make sure to refetch the pool after calling this method.
func (k Keeper) initOrUpdatePositionUptimeAccumulators(ctx sdk.Context, poolId uint64, liquidity osmomath.Dec, lowerTick, upperTick int64, liquidityDelta osmomath.Dec, positionId uint64) error {
	// We update accumulators _prior_ to any position-related updates to ensure
	// past rewards aren't distributed to new liquidity. We also update pool's
	// LastLiquidityUpdate here.
	err := k.UpdatePoolUptimeAccumulatorsToNow(ctx, poolId)
	if err != nil {
		return err
	}

	// Get uptime accumulators for every supported uptime.
	uptimeAccumulators, err := k.GetUptimeAccumulators(ctx, poolId)
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
	positionName := string(types.KeyPositionId(positionId))
	for uptimeIndex, curUptimeAccum := range uptimeAccumulators {
		// If a record does not exist for this uptime accumulator, create a new position.
		// Otherwise, add to existing record.
		recordExists := curUptimeAccum.HasPosition(positionName)

		if !recordExists {
			// Liquidity cannot be negative for a new position
			if !liquidityDelta.IsPositive() {
				return types.NonPositiveLiquidityForNewPositionError{LiquidityDelta: liquidityDelta, PositionId: positionId}
			}

			// Since the position should only be entitled to uptime growth within its range, we checkpoint globalUptimeGrowthInsideRange as
			// its accumulator's init value. During the claiming (or, equivalently, position updating) process, we ensure that incentives are
			// not overpaid.
			err = curUptimeAccum.NewPositionIntervalAccumulation(positionName, liquidity, globalUptimeGrowthInsideRange[uptimeIndex], emptyOptions)
			if err != nil {
				return err
			}
		} else {
			// Prep accum since we claim rewards first under the hood before any update (otherwise we would overpay)
			err := updatePositionToInitValuePlusGrowthOutside(curUptimeAccum, positionName, globalUptimeGrowthOutsideRange[uptimeIndex])
			if err != nil {
				return err
			}

			// Note that even though "unclaimed rewards" accrue in the accumulator prior to reaching minUptime, since position withdrawal
			// and incentive collection are only allowed when current time is past minUptime these rewards are not accessible until then.
			err = curUptimeAccum.UpdatePositionIntervalAccumulation(positionName, liquidityDelta, globalUptimeGrowthInsideRange[uptimeIndex])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// updateAccumAndClaimRewards claims and returns the rewards that `positionKey` is entitled to, updating the accumulator's value before
// and after claiming to ensure that rewards are never overdistributed.
// CONTRACT: position accumulator value prior to this call is equal to the growth inside the position at the time of last update.
// Returns error if:
// - fails to prepare position accumulator
// - fails to claim rewards
// - fails to check if position record exists
// - fails to update position accumulator with the current growth inside the position
func updateAccumAndClaimRewards(accum *accum.AccumulatorObject, positionKey string, growthOutside sdk.DecCoins) (sdk.Coins, sdk.DecCoins, error) {
	// Set the position's accumulator value to it's initial value at creation time plus the growth outside at this moment.
	err := updatePositionToInitValuePlusGrowthOutside(accum, positionKey, growthOutside)
	if err != nil {
		return sdk.Coins{}, sdk.DecCoins{}, err
	}

	// Claim rewards, set the unclaimed rewards to zero, and update the position's accumulator value to reflect the current accumulator value.
	// Removes the position state from accum if remaining liquidity is zero for the position.
	incentivesClaimedCurrAccum, dust, err := accum.ClaimRewards(positionKey)
	if err != nil {
		return sdk.Coins{}, sdk.DecCoins{}, err
	}

	// Check if position record was deleted after claiming rewards.
	hasPosition := accum.HasPosition(positionKey)

	// If position still exists, we update the position's accumulator value to be the current accumulator value minus the growth outside.
	if hasPosition {
		// The position accumulator value must always equal to the growth inside at the time of last update.
		// Since this is the time we update the accumulator, we must subtract the growth outside from the global accumulator value
		// to get growth inside at the current block time.
		// Note: this is SafeSub because interval accumulation is allowed to be negative.
		currentGrowthInsideForPosition, _ := accum.GetValue().SafeSub(growthOutside)
		err := accum.SetPositionIntervalAccumulation(positionKey, currentGrowthInsideForPosition)
		if err != nil {
			return sdk.Coins{}, sdk.DecCoins{}, err
		}
	}

	return incentivesClaimedCurrAccum, dust, nil
}

// moveRewardsToNewPositionAndDeleteOldAcc claims the rewards from the old position and moves them to the new position.
// Deletes the position tracker associated with the old position name.
// The positions must be associated with the given accumulator.
// The given growth outside the positions range is used for claim rewards accounting.
// The rewards are moved as "unclaimed rewards" to the new position.
// Returns nil on success. Error otherwise.
// NOTE: It is only used by fungifyChargedPosition which we disabled for launch.
// nolint: unused
func moveRewardsToNewPositionAndDeleteOldAcc(accum *accum.AccumulatorObject, oldPositionName, newPositionName string, growthOutside sdk.DecCoins) error {
	if oldPositionName == newPositionName {
		return types.ModifySamePositionAccumulatorError{PositionAccName: oldPositionName}
	}

	hasPosition := accum.HasPosition(oldPositionName)
	if !hasPosition {
		return fmt.Errorf("position %s does not exist", oldPositionName)
	}

	if err := updatePositionToInitValuePlusGrowthOutside(accum, oldPositionName, growthOutside); err != nil {
		return err
	}

	unclaimedRewards, err := accum.DeletePosition(oldPositionName)
	if err != nil {
		return err
	}

	err = accum.AddToUnclaimedRewards(newPositionName, unclaimedRewards)
	if err != nil {
		return err
	}

	// Ensure that the new position's accumulator value is the growth inside.
	currentGrowthInsideForPosition := accum.GetValue().Sub(growthOutside)
	err = accum.SetPositionIntervalAccumulation(newPositionName, currentGrowthInsideForPosition)
	if err != nil {
		return err
	}

	return nil
}

// prepareClaimAllIncentivesForPosition updates accumulators to the current time and returns all the incentives for a given position.
// It claims all the incentives that the position is eligible for and determines if those incentives should be forfeited or not.
// The parent function (collectIncentives) does the actual bank sends for both the collected and forfeited incentives.
//
// Returns error if the position/uptime accumulators don't exist, or if there is an issue that arises while claiming.
func (k Keeper) prepareClaimAllIncentivesForPosition(ctx sdk.Context, positionId uint64) (sdk.Coins, sdk.Coins, error) {
	// Retrieve the position with the given ID.
	position, err := k.GetPosition(ctx, positionId)
	if err != nil {
		return sdk.Coins{}, sdk.Coins{}, err
	}

	err = k.UpdatePoolUptimeAccumulatorsToNow(ctx, position.PoolId)
	if err != nil {
		return sdk.Coins{}, sdk.Coins{}, err
	}

	// Compute the age of the position.
	positionAge := ctx.BlockTime().Sub(position.JoinTime)

	// Should never happen, defense in depth.
	if positionAge < 0 {
		return sdk.Coins{}, sdk.Coins{}, types.NegativeDurationError{Duration: positionAge}
	}

	// Retrieve the uptime accumulators for the position's pool.
	uptimeAccumulators, err := k.GetUptimeAccumulators(ctx, position.PoolId)
	if err != nil {
		return sdk.Coins{}, sdk.Coins{}, err
	}

	// Compute uptime growth outside of the range between lower tick and upper tick
	uptimeGrowthOutside, err := k.GetUptimeGrowthOutsideRange(ctx, position.PoolId, position.LowerTick, position.UpperTick)
	if err != nil {
		return sdk.Coins{}, sdk.Coins{}, err
	}

	// Create a variable to hold the name of the position.
	positionName := string(types.KeyPositionId(positionId))

	// Create variables to hold the total collected and forfeited incentives for the position.
	collectedIncentivesForPosition := sdk.Coins{}
	forfeitedIncentivesForPosition := sdk.Coins{}

	supportedUptimes := types.SupportedUptimes

	// Loop through each uptime accumulator for the pool.
	for uptimeIndex, uptimeAccum := range uptimeAccumulators {
		// Check if the accumulator contains the position.
		// There should never be a case where you can have a position for 1 accumulator, and not the rest.
		hasPosition := uptimeAccum.HasPosition(positionName)

		// If the accumulator contains the position, claim the position's incentives.
		if hasPosition {
			collectedIncentivesForUptime, _, err := updateAccumAndClaimRewards(uptimeAccum, positionName, uptimeGrowthOutside[uptimeIndex])
			if err != nil {
				return sdk.Coins{}, sdk.Coins{}, err
			}

			if positionAge < supportedUptimes[uptimeIndex] {
				// If the age of the position is less than the current uptime we are iterating through, then the position's
				// incentives are forfeited to the community pool. The parent function does the actual bank send.
				forfeitedIncentivesForPosition = forfeitedIncentivesForPosition.Add(collectedIncentivesForUptime...)
			} else {
				// If the age of the position is greater than or equal to the current uptime we are iterating through, then the
				// position's incentives are collected by the position owner. The parent function does the actual bank send.
				collectedIncentivesForPosition = collectedIncentivesForPosition.Add(collectedIncentivesForUptime...)
			}
		}
	}

	return collectedIncentivesForPosition, forfeitedIncentivesForPosition, nil
}

func (k Keeper) GetClaimableIncentives(ctx sdk.Context, positionId uint64) (sdk.Coins, sdk.Coins, error) {
	// Since this is a query, we don't want to modify the state and therefore use a cache context.
	cacheCtx, _ := ctx.CacheContext()
	return k.prepareClaimAllIncentivesForPosition(cacheCtx, positionId)
}

// collectIncentives collects incentives for all uptime accumulators for the specified position id.
//
// Upon successful collection, it bank sends the incentives from the pool address to the owner and returns the collected coins.
// Returns error if:
// - position with the given id does not exist
// - other internal database or math errors.
func (k Keeper) collectIncentives(ctx sdk.Context, sender sdk.AccAddress, positionId uint64) (sdk.Coins, sdk.Coins, error) {
	// Retrieve the position with the given ID.
	position, err := k.GetPosition(ctx, positionId)
	if err != nil {
		return sdk.Coins{}, sdk.Coins{}, err
	}

	if sender.String() != position.Address {
		return sdk.Coins{}, sdk.Coins{}, types.NotPositionOwnerError{
			PositionId: positionId,
			Address:    sender.String(),
		}
	}

	// Claim all incentives for the position.
	collectedIncentivesForPosition, forfeitedIncentivesForPosition, err := k.prepareClaimAllIncentivesForPosition(ctx, position.PositionId)
	if err != nil {
		return sdk.Coins{}, sdk.Coins{}, err
	}

	// If no incentives were collected, return an empty coin set.
	if collectedIncentivesForPosition.IsZero() && forfeitedIncentivesForPosition.IsZero() {
		return collectedIncentivesForPosition, forfeitedIncentivesForPosition, nil
	}

	// Send the collected incentives to the position's owner.
	pool, err := k.getPoolById(ctx, position.PoolId)
	if err != nil {
		return sdk.Coins{}, sdk.Coins{}, err
	}

	// Send the collected incentives to the position's owner from the pool's address.
	if err := k.bankKeeper.SendCoins(ctx, pool.GetIncentivesAddress(), sender, collectedIncentivesForPosition); err != nil {
		return sdk.Coins{}, sdk.Coins{}, err
	}

	// Send the forfeited incentives to the community pool from the pool's address.
	err = k.communityPoolKeeper.FundCommunityPool(ctx, forfeitedIncentivesForPosition, pool.GetIncentivesAddress())
	if err != nil {
		return sdk.Coins{}, sdk.Coins{}, err
	}

	// Emit an event indicating that incentives were collected.
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtCollectIncentives,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyPositionId, strconv.FormatUint(positionId, 10)),
			sdk.NewAttribute(types.AttributeKeyTokensOut, collectedIncentivesForPosition.String()),
			sdk.NewAttribute(types.AttributeKeyForfeitedTokens, forfeitedIncentivesForPosition.String()),
		),
	})

	return collectedIncentivesForPosition, forfeitedIncentivesForPosition, nil
}

// createIncentive creates an incentive record in state for the given pool.
//
// Upon successful creation, it bank sends the incentives from the owner address to the pool address and returns the incentives record.
// Returns error if:
// - poolId is invalid
// - incentiveAmount is invalid (zero or negative).
// - emissionRate is invalid (zero or negative)
// - startTime is < blockTime.
// - minUptime is not an authorizedUptime.
// - other internal database or math errors.
// WARNING: this method may mutate the pool, make sure to refetch the pool after calling this method.
func (k Keeper) CreateIncentive(ctx sdk.Context, poolId uint64, sender sdk.AccAddress, incentiveCoin sdk.Coin, emissionRate osmomath.Dec, startTime time.Time, minUptime time.Duration) (types.IncentiveRecord, error) {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return types.IncentiveRecord{}, err
	}

	// checks if the Coin has a non-negative amount and the denom is valid.
	if !incentiveCoin.IsValid() || incentiveCoin.IsZero() {
		return types.IncentiveRecord{}, types.InvalidIncentiveCoinError{PoolId: poolId, IncentiveCoin: incentiveCoin}
	}

	// Ensure start time is >= current blocktime
	if startTime.Before(ctx.BlockTime()) {
		return types.IncentiveRecord{}, types.StartTimeTooEarlyError{PoolId: poolId, CurrentBlockTime: ctx.BlockTime(), StartTime: startTime}
	}

	// Ensure emission rate is nonzero and nonnegative
	if !emissionRate.IsPositive() {
		return types.IncentiveRecord{}, types.NonPositiveEmissionRateError{PoolId: poolId, EmissionRate: emissionRate}
	}

	// Ensure min uptime is one of the authorized uptimes.
	// Note that this is distinct from the supported uptimes – while we set up pools and positions to
	// accommodate all supported uptimes, we only allow incentives to be created for uptimes that are
	// authorized by governance.
	authorizedUptimes := k.GetParams(ctx).AuthorizedUptimes
	osmoutils.SortSlice(authorizedUptimes)

	validUptime := false
	for _, authorizedUptime := range authorizedUptimes {
		if minUptime == authorizedUptime {
			validUptime = true

			// We break here to save on iterations
			break
		}
	}
	if !validUptime {
		return types.IncentiveRecord{}, types.InvalidMinUptimeError{PoolId: poolId, MinUptime: minUptime, AuthorizedUptimes: authorizedUptimes}
	}

	senderHasBalance := k.bankKeeper.HasBalance(ctx, sender, incentiveCoin)
	if !senderHasBalance {
		return types.IncentiveRecord{}, types.IncentiveInsufficientBalanceError{PoolId: poolId, IncentiveDenom: incentiveCoin.Denom, IncentiveAmount: incentiveCoin.Amount}
	}

	// Sync global uptime accumulators to current blocktime to ensure consistency in reward emissions
	err = k.UpdatePoolUptimeAccumulatorsToNow(ctx, poolId)
	if err != nil {
		return types.IncentiveRecord{}, err
	}

	// Get an ID unique to this incentive record
	incentiveRecordId := k.GetNextIncentiveRecordId(ctx)
	k.SetNextIncentiveRecordId(ctx, incentiveRecordId+1)

	incentiveRecordBody := types.IncentiveRecordBody{
		RemainingCoin: sdk.NewDecCoinFromCoin(incentiveCoin),
		EmissionRate:  emissionRate,
		StartTime:     startTime,
	}

	// Set up incentive record to put in state
	incentiveRecord := types.IncentiveRecord{
		PoolId:              poolId,
		IncentiveRecordBody: incentiveRecordBody,
		MinUptime:           minUptime,
		IncentiveId:         incentiveRecordId,
	}

	// Get all incentive records for uptime
	existingRecordsForUptime, err := k.getAllIncentiveRecordsForUptime(ctx, poolId, minUptime)
	if err != nil {
		return types.IncentiveRecord{}, err
	}

	// Fixed gas consumption per incentive creation to prevent spam
	ctx.GasMeter().ConsumeGas(uint64(types.BaseGasFeeForNewIncentive*len(existingRecordsForUptime)), "cl incentive creation fee")

	// Set incentive record in state
	err = k.setIncentiveRecord(ctx, incentiveRecord)
	if err != nil {
		return types.IncentiveRecord{}, err
	}

	// Transfer tokens from sender to the pool's incentive address
	if err := k.bankKeeper.SendCoins(ctx, sender, pool.GetIncentivesAddress(), sdk.NewCoins(incentiveCoin)); err != nil {
		return types.IncentiveRecord{}, err
	}

	return incentiveRecord, nil
}

// nolint: unused
// getLargestDuration retrieves the largest duration from the given slice.
func getLargestDuration(durations []time.Duration) time.Duration {
	var largest time.Duration
	for _, duration := range durations {
		if duration > largest {
			largest = duration
		}
	}
	return largest
}

// getLargestAuthorizedUptimeDuration retrieves the largest authorized uptime duration from the params.
// NOTE: It is only used by fungifyChargedPosition which we disabled for launch.
// nolint: unused
func (k Keeper) getLargestAuthorizedUptimeDuration(ctx sdk.Context) time.Duration {
	return getLargestDuration(k.GetParams(ctx).AuthorizedUptimes)
}

// nolint: unused
// getLargestSupportedUptimeDuration retrieves the largest supported uptime duration from the preset constant slice.
func (k Keeper) getLargestSupportedUptimeDuration() time.Duration {
	return getLargestDuration(types.SupportedUptimes)
}
