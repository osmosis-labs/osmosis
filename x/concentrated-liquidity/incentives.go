package concentrated_liquidity

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	sdkprefix "cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"golang.org/x/exp/slices"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

// We choose 10^27 to allow sufficient buffer before the accumulator starts getting truncated again.
// Internally, we multiply the number of seconds passed since the last liquidity update by the emission rate per second
// Then, we scale that value by 10^27 to avoid truncation to zero when dividing by the liquidity in the accumulator.
// We do not go for a higher scaling factor to allow for enough room before hitting the maximum integer value of 2^256
// in the intermediary multiplications.
//
// More analysis on the choice of scaling factor can be found here:
// https://hackmd.io/o3oqT8VhSPKAiqNl_mlxXQ
var (
	perUnitLiqScalingFactor = osmomath.NewDec(1e15).MulMut(osmomath.NewDec(1e12))
	oneDecScalingFactor     = osmomath.OneDec()
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
var oneDec = osmomath.OneDec()

// updateGivenPoolUptimeAccumulatorsToNow syncs all given uptime accumulators for a given pool id.
// Updates the pool last liquidity update time with the current block time and writes the updated pool to state.
// If last liquidity update happened in the current block, this function is a no-op.
// Specifically, it gets the time elapsed since the last update and divides it
// by the qualifying liquidity for each uptime. It then adds this value to the
// respective accumulator and updates relevant time trackers accordingly.
// This method also serves the purpose of correctly splitting rewards between the linked balancer pool and the cl pool.
// CONTRACT: the caller validates that the pool with the given id exists.
// CONTRACT: given uptimeAccums are associated with the given pool id.
// CONTRACT: caller is responsible for the uptimeAccums to be up-to-date.
//
// WARNING: this method may mutate the pool, make sure to refetch the pool after calling this method.
// Note: the following are the differences of this function from updatePoolUptimeAccumulatorsToNow:
//
// * this function does not refetch the uptime accumulators from state.
//
// * this function operates on the given pool directly, instead of fetching it from state.
//
// This is to avoid unnecessary state reads during swaps for performance reasons.
func (k Keeper) updateGivenPoolUptimeAccumulatorsToNow(ctx sdk.Context, pool types.ConcentratedPoolExtension, uptimeAccums []*accum.AccumulatorObject) error {
	if pool == nil {
		return types.ErrPoolNil
	}

	// Since our base unit of time is nanoseconds, we divide with truncation by 10^9 to get
	// time elapsed in seconds
	timeElapsedNanoSec := osmomath.NewDec(int64(ctx.BlockTime().Sub(pool.GetLastLiquidityUpdate())))
	timeElapsedSec := timeElapsedNanoSec.QuoMut(dec1e9)

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

	incentiveScalingFactorForPool, err := k.getIncentiveScalingFactorForPool(ctx, poolId)
	if err != nil {
		return err
	}

	// We optimistically assume that all liquidity on the active tick qualifies and handle
	// uptime-related checks in forfeiting logic.

	// If there is no share to be incentivized for the current uptime accumulator, we leave it unchanged
	qualifyingLiquidity := pool.GetLiquidity()
	if !qualifyingLiquidity.LT(oneDec) {
		for uptimeIndex := range uptimeAccums {
			// Get relevant uptime-level values
			curUptimeDuration := types.SupportedUptimes[uptimeIndex]
			incentivesToAddToCurAccum, updatedPoolRecords, err := calcAccruedIncentivesForAccum(
				ctx, curUptimeDuration, qualifyingLiquidity, timeElapsedSec, poolIncentiveRecords, poolId, incentiveScalingFactorForPool)
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
func calcAccruedIncentivesForAccum(
	ctx sdk.Context,
	accumUptime time.Duration,
	liquidityInAccum osmomath.Dec,
	timeElapsed osmomath.Dec,
	poolIncentiveRecords []types.IncentiveRecord,
	poolID uint64,
	incentiveScalingFactorForPool osmomath.Dec,
) (sdk.DecCoins, []types.IncentiveRecord, error) {
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
		totalEmittedAmount, err := computeTotalIncentivesToEmit(timeElapsed, incentiveRecordBody.EmissionRate)
		if err != nil {
			emitIncentiveOverflowTelemetry(poolID, incentiveRecord.IncentiveId, timeElapsed, incentiveRecordBody.EmissionRate, err)
			// Silently ignore the truncated incentive record to avoid halting the entire accumulator update.
			// Continue to the next incentive record.
			continue
		}

		// We scale up the remaining rewards to avoid truncation to zero
		// when dividing by the liquidity in the accumulator.
		scaledTotalEmittedAmount, err := scaleUpTotalEmittedAmount(totalEmittedAmount, incentiveScalingFactorForPool)
		if err != nil {
			emitIncentiveOverflowTelemetry(poolID, incentiveRecord.IncentiveId, timeElapsed, incentiveRecordBody.EmissionRate, err)
			// Silently ignore the truncated incentive record to avoid halting the entire accumulator update.
			// Continue to the next incentive record.
			continue
		}

		// Incentives to emit per unit of qualifying liquidity = total emitted / liquidityInAccum
		// Note that we truncate to ensure we do not overdistribute incentives
		incentivesPerLiquidity := scaledTotalEmittedAmount.QuoTruncate(liquidityInAccum)

		emittedIncentivesPerLiquidity := sdk.NewDecCoinFromDec(incentiveRecordBody.RemainingCoin.Denom, incentivesPerLiquidity)

		// Ensure that we only emit if there are enough incentives remaining to be emitted
		remainingRewards := poolIncentiveRecords[incentiveIndex].IncentiveRecordBody.RemainingCoin.Amount

		// if total amount emitted does not exceed remaining rewards,
		if totalEmittedAmount.LTE(remainingRewards) {
			// Emit telemetry for accumulator updates
			emitAccumulatorUpdateTelemetry(types.IncentiveTruncationTelemetryName, incentivesPerLiquidity, totalEmittedAmount, poolID, liquidityInAccum)

			incentivesToAddToCurAccum = incentivesToAddToCurAccum.Add(emittedIncentivesPerLiquidity)

			// Update incentive record to reflect the incentives that were emitted
			remainingRewards = remainingRewards.Sub(totalEmittedAmount)

			// Each incentive record should only be modified once
			copyPoolIncentiveRecords[incentiveIndex].IncentiveRecordBody.RemainingCoin.Amount = remainingRewards
		} else {
			// If there are not enough incentives remaining to be emitted, we emit the remaining rewards.
			// When the returned records are set in state, all records with remaining rewards of zero will be cleared.

			// We scale up the remaining rewards to avoid truncation to zero
			// when dividing by the liquidity in the accumulator.
			remainingRewardsScaled, err := scaleUpTotalEmittedAmount(remainingRewards, incentiveScalingFactorForPool)
			if err != nil {
				ctx.Logger().Info(types.IncentiveOverflowTelemetryName, "pool_id", poolID, "incentive_id", incentiveRecord.IncentiveId, "time_elapsed", timeElapsed, "emission_rate", incentiveRecordBody.EmissionRate, "error", err.Error())
				// Silently ignore the truncated incentive record to avoid halting the entire accumulator update.
				// Continue to the next incentive record.
				continue
			}
			remainingIncentivesPerLiquidity := remainingRewardsScaled.QuoTruncateMut(liquidityInAccum)

			emittedIncentivesPerLiquidity = sdk.NewDecCoinFromDec(incentiveRecordBody.RemainingCoin.Denom, remainingIncentivesPerLiquidity)

			// Emit telemetry for accumulator updates
			emitAccumulatorUpdateTelemetry(types.IncentiveTruncationTelemetryName, remainingIncentivesPerLiquidity, remainingRewards, poolID, liquidityInAccum)

			incentivesToAddToCurAccum = incentivesToAddToCurAccum.Add(emittedIncentivesPerLiquidity)

			copyPoolIncentiveRecords[incentiveIndex].IncentiveRecordBody.RemainingCoin.Amount = osmomath.ZeroDec()
		}
	}

	return incentivesToAddToCurAccum, copyPoolIncentiveRecords, nil
}

// scaleUpTotalEmittedAmount scales up the total emitted amount to avoid truncation to zero.
// Returns error if the total emitted amount is too high and causes overflow when applying scaling factor.
func scaleUpTotalEmittedAmount(totalEmittedAmount osmomath.Dec, scalingFactor osmomath.Dec) (scaledTotalEmittedAmount osmomath.Dec, err error) {
	defer func() {
		r := recover()

		if r != nil {
			telemetry.IncrCounter(1, types.IncentiveOverflowTelemetryName)
			err = types.IncentiveScalingFactorOverflowError{
				PanicMessage: fmt.Sprintf("%v", r),
			}
		}
	}()

	return totalEmittedAmount.MulTruncate(scalingFactor), nil
}

// scaleDownIncentiveAmount scales down the incentive amount by the scaling factor.
func scaleDownIncentiveAmount(incentiveAmount osmomath.Int, scalingFactor osmomath.Dec) (scaledTotalEmittedAmount osmomath.Int) {
	return incentiveAmount.ToLegacyDec().QuoTruncateMut(scalingFactor).TruncateInt()
}

// computeTotalIncentivesToEmit computes the total incentives to emit based on the time elapsed and emission rate.
// Returns error if timeElapsed or emissionRate are too high, causing overflow during multiplication.
func computeTotalIncentivesToEmit(timeElapsedSeconds osmomath.Dec, emissionRate osmomath.Dec) (totalEmittedAmount osmomath.Dec, err error) {
	defer func() {
		r := recover()

		if r != nil {
			telemetry.IncrCounter(1, types.IncentiveOverflowTelemetryName)
			err = types.IncentiveEmissionOvrflowError{
				PanicMessage: fmt.Sprintf("%v", r),
			}
		}
	}()

	// This may panic if emissionRate is too high and causes overflow during multiplication
	// We are unlikely to see an overflow due to too much time elapsing since
	// 100 years in seconds is roughly
	// 3.15576e9 * 100 = 3.15576e11
	// 60 * 60 * 24 * 365 * 100 = 3153600000 seconds
	// The bit decimal bit length is 2^256 which is arond 10^77
	// However, it is possible for an attacker to try and create incentives with a very high emission rate
	// consisting of cheap token in the USD denomination. This is why we have the panic recovery above.
	return timeElapsedSeconds.MulTruncate(emissionRate), nil
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
// UptimeGrowthInside tracks the incentives accrued by a specific LP within a pool. It keeps track of the cumulative amount of incentives
// collected by a specific LP within a pool. This function also measures the growth of incentives accrued by a particular LP since the last
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
// UptimeGrowthOutside tracks the incentive accrued by the entire pool. It keeps track of the cumulative amount of incentives collected
// by a specific pool since the last time incentives were accrued.
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
func (k Keeper) prepareClaimAllIncentivesForPosition(ctx sdk.Context, positionId uint64) (sdk.Coins, sdk.Coins, []sdk.Coins, error) {
	// Retrieve the position with the given ID.
	position, err := k.GetPosition(ctx, positionId)
	if err != nil {
		return sdk.Coins{}, sdk.Coins{}, nil, err
	}

	err = k.UpdatePoolUptimeAccumulatorsToNow(ctx, position.PoolId)
	if err != nil {
		return sdk.Coins{}, sdk.Coins{}, nil, err
	}

	// Compute the age of the position.
	positionAge := ctx.BlockTime().Sub(position.JoinTime)

	// Should never happen, defense in depth.
	if positionAge < 0 {
		return sdk.Coins{}, sdk.Coins{}, nil, types.NegativeDurationError{Duration: positionAge}
	}

	// Retrieve the uptime accumulators for the position's pool.
	uptimeAccumulators, err := k.GetUptimeAccumulators(ctx, position.PoolId)
	if err != nil {
		return sdk.Coins{}, sdk.Coins{}, nil, err
	}

	// Compute uptime growth outside of the range between lower tick and upper tick
	uptimeGrowthOutside, err := k.GetUptimeGrowthOutsideRange(ctx, position.PoolId, position.LowerTick, position.UpperTick)
	if err != nil {
		return sdk.Coins{}, sdk.Coins{}, nil, err
	}

	// Create a variable to hold the name of the position.
	positionName := string(types.KeyPositionId(positionId))

	// Create variables to hold the total collected and forfeited incentives for the position.
	collectedIncentivesForPosition := sdk.Coins{}
	forfeitedIncentivesForPosition := sdk.Coins{}

	supportedUptimes := types.SupportedUptimes

	incentiveScalingFactor, err := k.getIncentiveScalingFactorForPool(ctx, position.PoolId)
	if err != nil {
		return sdk.Coins{}, sdk.Coins{}, nil, err
	}

	// Loop through each uptime accumulator for the pool.
	scaledForfeitedIncentivesByUptime := make([]sdk.Coins, len(types.SupportedUptimes))
	for uptimeIndex, uptimeAccum := range uptimeAccumulators {
		// Check if the accumulator contains the position.
		// There should never be a case where you can have a position for 1 accumulator, and not the rest.
		hasPosition := uptimeAccum.HasPosition(positionName)

		// If the accumulator contains the position, claim the position's incentives.
		if hasPosition {
			collectedIncentivesForUptimeScaled, _, err := updateAccumAndClaimRewards(uptimeAccum, positionName, uptimeGrowthOutside[uptimeIndex])
			if err != nil {
				return sdk.Coins{}, sdk.Coins{}, nil, err
			}

			// We scale the uptime per-unit of liquidity accumulator up to avoid truncation to zero.
			// However, once we compute the total for the liquidity entitlement, we must scale it back down.
			// We always truncate down in the pool's favor.
			collectedIncentivesForUptime := sdk.NewCoins()
			for _, incentiveCoin := range collectedIncentivesForUptimeScaled {
				incentiveCoin.Amount = scaleDownIncentiveAmount(incentiveCoin.Amount, incentiveScalingFactor)
				if incentiveCoin.Amount.IsPositive() {
					collectedIncentivesForUptime = append(collectedIncentivesForUptime, incentiveCoin)
				}
			}

			if positionAge < supportedUptimes[uptimeIndex] {
				// We track forfeited incentives by uptime accumulator to allow for efficient redepositing.
				// To avoid descaling and rescaling, we keep the forfeited incentives in scaled form.
				// This is slightly unwieldy as it means we return a slice of scaled coins, but doing it this way
				// allows us to efficiently handle all cases related to forfeited incentives.
				scaledForfeitedIncentivesByUptime[uptimeIndex] = collectedIncentivesForUptimeScaled

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

	return collectedIncentivesForPosition, forfeitedIncentivesForPosition, scaledForfeitedIncentivesByUptime, nil
}

// redepositForfeitedIncentives handles logic for redepositing forfeited incentives for a given pool.
// Specifically, it implements the following flows:
//   - If there is no remaining active liquidity, the forfeited incentives are sent back to the sender.
//   - If there is active liquidity, the forfeited incentives are redeposited into the uptime accumulators.
//     Since forfeits are already being tracked in "scaled form", we do not need to do any additional scaling
//     and simply deposit amount / activeLiquidity into the uptime accumulators.
//
// Returns error if:
// * Pool with the given ID does not exist
// * Uptime accumulators for the pool cannot be retrieved
// * Forfeited incentives length does not match the supported uptimes (defense in depth, should never happen)
// * Bank send fails
func (k Keeper) redepositForfeitedIncentives(ctx sdk.Context, poolId uint64, sender sdk.AccAddress, scaledForfeitedIncentivesByUptime []sdk.Coins, totalForefeitedIncentives sdk.Coins) error {
	if len(scaledForfeitedIncentivesByUptime) != len(types.SupportedUptimes) {
		return types.InvalidForfeitedIncentivesLengthError{ForfeitedIncentivesLength: len(scaledForfeitedIncentivesByUptime), ExpectedLength: len(types.SupportedUptimes)}
	}

	// Fetch pool from state to check active liquidity.
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return err
	}
	activeLiquidity := pool.GetLiquidity()

	// If no active liquidity, give the forfeited incentives to the sender.
	if activeLiquidity.LT(osmomath.OneDec()) {
		err := k.bankKeeper.SendCoins(ctx, pool.GetIncentivesAddress(), sender, totalForefeitedIncentives)
		if err != nil {
			return err
		}
		return nil
	}

	// If pool has active liquidity on current tick, redeposit forfeited incentives into uptime accumulators.
	uptimeAccums, err := k.GetUptimeAccumulators(ctx, poolId)
	if err != nil {
		return err
	}

	// Loop through each uptime accumulator for the pool and redeposit forfeited incentives.
	for uptimeIndex := range uptimeAccums {
		curUptimeForfeited := scaledForfeitedIncentivesByUptime[uptimeIndex]
		if curUptimeForfeited.IsZero() {
			continue
		}

		// Note that this logic is a simplified version of the regular incentive distribution logic.
		// It leans on the fact that the tracked forfeited incentives are already scaled appropriately
		// so we do not need to run any additional computations beyond dividing by the active liquidity.
		incentivesToAddToCurAccum := sdk.NewDecCoins()
		for _, forfeitedCoin := range curUptimeForfeited {
			// Calculate the amount to add to the accumulator by dividing the forfeited coin amount by the current uptime duration
			forfeitedAmountPerLiquidity := forfeitedCoin.Amount.ToLegacyDec().QuoTruncate(activeLiquidity)

			// Create a DecCoin from the calculated amount
			decCoinToAdd := sdk.NewDecCoinFromDec(forfeitedCoin.Denom, forfeitedAmountPerLiquidity)

			// Add the calculated DecCoin to the incentives to add to current accumulator
			incentivesToAddToCurAccum = incentivesToAddToCurAccum.Add(decCoinToAdd)
		}

		// Emit incentives to current uptime accumulator
		uptimeAccums[uptimeIndex].AddToAccumulator(incentivesToAddToCurAccum)
	}

	return nil
}

func (k Keeper) GetClaimableIncentives(ctx sdk.Context, positionId uint64) (sdk.Coins, sdk.Coins, error) {
	// Since this is a query, we don't want to modify the state and therefore use a cache context.
	cacheCtx, _ := ctx.CacheContext()
	// We omit the by-uptime forfeited incentives slice as it is not needed for this query.
	collectedIncentives, forfeitedIncentives, _, err := k.prepareClaimAllIncentivesForPosition(cacheCtx, positionId)
	return collectedIncentives, forfeitedIncentives, err
}

// collectIncentives collects incentives for all uptime accumulators for the specified position id.
//
// Upon successful collection, it bank sends the incentives from the pool address to the owner and returns the collected coins.
// Returns error if:
// - position with the given id does not exist
// - other internal database or math errors.
func (k Keeper) collectIncentives(ctx sdk.Context, sender sdk.AccAddress, positionId uint64) (sdk.Coins, sdk.Coins, []sdk.Coins, error) {
	// Retrieve the position with the given ID.
	position, err := k.GetPosition(ctx, positionId)
	if err != nil {
		return sdk.Coins{}, sdk.Coins{}, nil, err
	}

	if sender.String() != position.Address {
		return sdk.Coins{}, sdk.Coins{}, nil, types.NotPositionOwnerError{
			PositionId: positionId,
			Address:    sender.String(),
		}
	}

	// Claim all incentives for the position.
	collectedIncentivesForPosition, totalForfeitedIncentivesForPosition, scaledAmountForfeitedByUptime, err := k.prepareClaimAllIncentivesForPosition(ctx, position.PositionId)
	if err != nil {
		return sdk.Coins{}, sdk.Coins{}, nil, err
	}

	// If no incentives were collected, return an empty coin set.
	if collectedIncentivesForPosition.IsZero() && totalForfeitedIncentivesForPosition.IsZero() {
		return collectedIncentivesForPosition, totalForfeitedIncentivesForPosition, scaledAmountForfeitedByUptime, nil
	}

	// Send the collected incentives to the position's owner.
	pool, err := k.getPoolById(ctx, position.PoolId)
	if err != nil {
		return sdk.Coins{}, sdk.Coins{}, nil, err
	}

	// Send the collected incentives to the position's owner from the pool's address.
	if !collectedIncentivesForPosition.IsZero() {
		if err := k.bankKeeper.SendCoins(ctx, pool.GetIncentivesAddress(), sender, collectedIncentivesForPosition); err != nil {
			return sdk.Coins{}, sdk.Coins{}, nil, err
		}
	}

	// Emit an event indicating that incentives were collected.
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtCollectIncentives,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(pool.GetId(), 10)),
			sdk.NewAttribute(types.AttributeKeyPositionId, strconv.FormatUint(positionId, 10)),
			sdk.NewAttribute(types.AttributeKeyTokensOut, collectedIncentivesForPosition.String()),
			sdk.NewAttribute(types.AttributeKeyForfeitedTokens, totalForfeitedIncentivesForPosition.String()),
		),
	})

	return collectedIncentivesForPosition, totalForfeitedIncentivesForPosition, scaledAmountForfeitedByUptime, nil
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

// getIncentiveScalingFactorForPool returns the scaling factor for the given pool.
// It returns perUnitLiqScalingFactor if the pool is migrated or if the pool ID is greater than the migration threshold.
// It returns oneDecScalingFactor otherwise.
func (k Keeper) getIncentiveScalingFactorForPool(ctx sdk.Context, poolID uint64) (osmomath.Dec, error) {
	migrationThreshold, err := k.GetIncentivePoolIDMigrationThreshold(ctx)
	if err != nil {
		return osmomath.Dec{}, err
	}

	// If the given pool ID is greater than the migration threshold, we return the perUnitLiqScalingFactor.
	if poolID > migrationThreshold {
		return perUnitLiqScalingFactor, nil
	}

	// If the given pool ID is in the migrated incentive accumulator pool IDs, we return the perUnitLiqScalingFactor.
	_, isMigrated := types.MigratedIncentiveAccumulatorPoolIDs[poolID]
	if isMigrated {
		return perUnitLiqScalingFactor, nil
	}

	// If the given pool ID is in the migrated incentive accumulator pool IDs (v24), we return the perUnitLiqScalingFactor.
	_, isMigrated = types.MigratedIncentiveAccumulatorPoolIDsV24[poolID]
	if isMigrated {
		return perUnitLiqScalingFactor, nil
	}

	// Otherwise, we return the oneDecScalingFactor.
	return oneDecScalingFactor, nil
}

// nolint: unused
// getLargestSupportedUptimeDuration retrieves the largest supported uptime duration from the preset constant slice.
func (k Keeper) getLargestSupportedUptimeDuration() time.Duration {
	return getLargestDuration(types.SupportedUptimes)
}

// SetIncentivePoolIDMigrationThreshold sets the pool ID migration threshold to the last pool ID.
func (k Keeper) SetIncentivePoolIDMigrationThreshold(ctx sdk.Context, poolIDThreshold uint64) {
	// Set the pool ID migration threshold to the last pool ID
	store := ctx.KVStore(k.storeKey)

	store.Set(types.KeyIncentiveAccumulatorMigrationThreshold, sdk.Uint64ToBigEndian(poolIDThreshold))
}

// GetIncentivePoolIDMigrationThreshold returns the pool ID migration threshold for incentive accumulators.
func (k Keeper) GetIncentivePoolIDMigrationThreshold(ctx sdk.Context) (uint64, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyIncentiveAccumulatorMigrationThreshold)

	if bz == nil {
		return 0, fmt.Errorf("incentive accumulator migration threshold not found")
	}

	threshold := sdk.BigEndianToUint64(bz)

	return threshold, nil
}
