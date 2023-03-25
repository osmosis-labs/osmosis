package concentrated_liquidity

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/exp/slices"

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
			} else {
				// If there are not enough incentives remaining to be emitted, we emit the remaining rewards.
				// When the returned records are set in state, all records with remaining rewards of zero will be cleared.
				remainingIncentivesPerLiquidity := remainingRewards.QuoTruncate(qualifyingLiquidity)
				emittedIncentivesPerLiquidity = sdk.NewDecCoinFromDec(incentiveRecord.IncentiveDenom, remainingIncentivesPerLiquidity)
				incentivesToAddToCurAccum = incentivesToAddToCurAccum.Add(emittedIncentivesPerLiquidity)

				poolIncentiveRecords[incentiveIndex].RemainingAmount = sdk.ZeroDec()
			}
		}
	}

	return incentivesToAddToCurAccum, poolIncentiveRecords, nil
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

// nolint: unused
// setIncentiveRecords sets the passed in incentive records in state
// Errors if the incentive record has an unsupported min uptime.
func (k Keeper) setIncentiveRecord(ctx sdk.Context, incentiveRecord types.IncentiveRecord) error {
	store := ctx.KVStore(k.storeKey)

	uptimeIndex, err := findUptimeIndex(incentiveRecord.MinUptime)
	if err != nil {
		return err
	}

	key := types.KeyIncentiveRecord(incentiveRecord.PoolId, uptimeIndex, incentiveRecord.IncentiveDenom, incentiveRecord.IncentiveCreator)
	incentiveRecordBody := types.IncentiveRecordBody{
		RemainingAmount: incentiveRecord.RemainingAmount,
		EmissionRate:    incentiveRecord.EmissionRate,
		StartTime:       incentiveRecord.StartTime,
	}

	// If the remaining amount is zero and the record already exists in state, we delete the record from state.
	// If it's zero and the record doesn't exist in state, we do a no-op.
	// In all other cases, we update the record in state
	if store.Has(key) && incentiveRecordBody.RemainingAmount.IsZero() {
		store.Delete(key)
	} else if incentiveRecordBody.RemainingAmount.GT(sdk.ZeroDec()) {
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
func (k Keeper) GetIncentiveRecord(ctx sdk.Context, poolId uint64, denom string, minUptime time.Duration, incentiveCreator sdk.AccAddress) (types.IncentiveRecord, error) {
	store := ctx.KVStore(k.storeKey)
	incentiveBodyStruct := types.IncentiveRecordBody{}

	uptimeIndex, err := findUptimeIndex(minUptime)
	if err != nil {
		return types.IncentiveRecord{}, err
	}

	key := types.KeyIncentiveRecord(poolId, uptimeIndex, denom, incentiveCreator)

	found, err := osmoutils.Get(store, key, &incentiveBodyStruct)
	if err != nil {
		return types.IncentiveRecord{}, err
	}

	if !found {
		return types.IncentiveRecord{}, types.IncentiveRecordNotFoundError{PoolId: poolId, IncentiveDenom: denom, MinUptime: minUptime, IncentiveCreatorStr: incentiveCreator.String()}
	}

	return types.IncentiveRecord{
		PoolId:           poolId,
		IncentiveDenom:   denom,
		IncentiveCreator: incentiveCreator,
		MinUptime:        minUptime,
		RemainingAmount:  incentiveBodyStruct.RemainingAmount,
		EmissionRate:     incentiveBodyStruct.EmissionRate,
		StartTime:        incentiveBodyStruct.StartTime,
	}, nil
}

// GetAllIncentiveRecordsForPool gets all the incentive records for poolId
// Returns error if it is unable to retrieve records.
func (k Keeper) GetAllIncentiveRecordsForPool(ctx sdk.Context, poolId uint64) ([]types.IncentiveRecord, error) {
	return osmoutils.GatherValuesFromStorePrefixWithKeyParser(ctx.KVStore(k.storeKey), types.KeyPoolIncentiveRecords(poolId), ParseFullIncentiveRecordFromBz)
}

// GetAllIncentiveRecordsForUptime gets all the incentive records for the given poolId and minUptime
// Returns error if the passed in uptime is not supported or it is unable to retrieve records.
func (k Keeper) GetAllIncentiveRecordsForUptime(ctx sdk.Context, poolId uint64, minUptime time.Duration) ([]types.IncentiveRecord, error) {
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

// initPositionUptime adds for all uptime accumulators `position` qualifies for.
func (k Keeper) initPositionUptime(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, liquidityDelta sdk.Dec, joinTime time.Time, freezeDuration time.Duration, positionId uint64) error {
	// Create records for relevant uptime accumulators here.
	uptimeAccumulators, err := k.getUptimeAccumulators(ctx, poolId)
	if err != nil {
		return err
	}

	globalUptimeGrowthInsideRange, err := k.GetUptimeGrowthInsideRange(ctx, poolId, lowerTick, upperTick)
	if err != nil {
		return err
	}

	// Loop through uptime accums for all supported uptimes on the pool and init or update position's records
	positionName := string(types.KeyPositionId(positionId))
	for uptimeIndex, uptime := range types.SupportedUptimes {
		// We assume every position update requires the position to be frozen for the
		// min uptime again. Thus, the difference between the position's `freezeDuration`
		// and the blocktime when the update happens should be greater than or equal
		// to the required uptime.
		if freezeDuration >= uptime {
			curUptimeAccum := uptimeAccumulators[uptimeIndex]

			// If a record does not exist for this uptime accumulator, create a new position.
			// Otherwise, throw an error
			recordExists, err := curUptimeAccum.HasPosition(positionName)
			if err != nil {
				return err
			}
			if recordExists {
				return types.PositionAlreadyExistsError{PoolId: poolId, LowerTick: lowerTick, UpperTick: upperTick, JoinTime: joinTime, FreezeDuration: freezeDuration}
			}

			// Since the position should only be entitled to uptime growth within its range, we checkpoint globalUptimeGrowthInsideRange as
			// its accumulator's init value. During the claiming (or, equivalently, position updating) process, we ensure that incentives are
			// not overpaid.
			err = curUptimeAccum.NewPositionCustomAcc(positionName, liquidityDelta, globalUptimeGrowthInsideRange[uptimeIndex], emptyOptions)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// prepareAccumAndClaimRewards claims and returns the rewards that `positionKey` is entitled to, updating the accumulator's value before
// and after claiming to ensure that rewards are never overdistributed.
func prepareAccumAndClaimRewards(accum accum.AccumulatorObject, positionKey string, growthOutside sdk.DecCoins) (sdk.Coins, error) {
	err := preparePositionAccumulator(accum, positionKey, growthOutside)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Claim incentives
	incentivesClaimedCurrAccum, err := accum.ClaimRewards(positionKey)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Check if position record was deleted after claiming rewards. If not, we update the custom accumulator value.
	hasPosition, err := accum.HasPosition(positionKey)
	if err != nil {
		return sdk.Coins{}, err
	}

	if hasPosition {
		customAccumulatorValue := accum.GetValue().Sub(growthOutside)
		err := accum.SetPositionCustomAcc(positionKey, customAccumulatorValue)
		if err != nil {
			return sdk.Coins{}, err
		}
	}

	return incentivesClaimedCurrAccum, nil
}

// claimAllIncentivesForPosition claims and returns all the incentives for a given position.
// It takes in a `forfeitIncentives` boolean to indicate whether the accrued incentives should be forfeited, in which case it
// redeposits the accrued rewards back into the accumulator as additional rewards for other participants.
func (k Keeper) claimAllIncentivesForPosition(ctx sdk.Context, positionId uint64, forfeitIncentives bool) (sdk.Coins, error) {
	// Retrieve the position with the given ID.
	position, err := k.GetPosition(ctx, positionId)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Retrieve the uptime accumulators for the position's pool.
	uptimeAccumulators, err := k.getUptimeAccumulators(ctx, position.PoolId)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Compute uptime growth outside of the range between lower tick and upper tick
	uptimeGrowthOutside, err := k.GetUptimeGrowthOutsideRange(ctx, position.PoolId, position.LowerTick, position.UpperTick)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Create a variable to hold the name of the position.
	positionName := string(types.KeyPositionId(positionId))

	// Create a variable to hold the total collected incentives for the position.
	collectedIncentivesForPosition := sdk.Coins{}

	// Loop through each uptime accumulator for the pool.
	for uptimeIndex, uptimeAccum := range uptimeAccumulators {
		// Check if the accumulator contains the position.
		hasPosition, err := uptimeAccum.HasPosition(positionName)
		if err != nil {
			return sdk.Coins{}, err
		}

		// If the accumulator contains the position, claim the position's incentives.
		if hasPosition {
			collectedIncentivesForUptime, err := prepareAccumAndClaimRewards(uptimeAccum, positionName, uptimeGrowthOutside[uptimeIndex])
			if err != nil {
				return sdk.Coins{}, err
			}

			// If the claimed incentives are forfeited, deposit them back into the accumulator to be distributed
			// to other qualifying positions.
			if forfeitIncentives {
				uptimeAccum.AddToAccumulator(sdk.NewDecCoinsFromCoins(collectedIncentivesForUptime...))
			}

			collectedIncentivesForPosition = collectedIncentivesForPosition.Add(collectedIncentivesForUptime...)
		}
	}

	return collectedIncentivesForPosition, nil
}

// collectIncentives collects incentives for all uptime accumulators for the specified position id.
//
// Upon successful collection, it bank sends the incentives from the pool address to the owner and returns the collected coins.
// Returns error if:
// - position with the given id does not exist
// - other internal database or math errors.
func (k Keeper) collectIncentives(ctx sdk.Context, owner sdk.AccAddress, positionId uint64) (sdk.Coins, error) {
	// Retrieve the position with the given ID.
	position, err := k.GetPosition(ctx, positionId)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Claim all incentives for the position.
	collectedIncentivesForPosition, err := k.claimAllIncentivesForPosition(ctx, position.PositionId, false)
	if err != nil {
		return sdk.Coins{}, err
	}

	// If no incentives were collected, return an empty coin set.
	if collectedIncentivesForPosition.IsZero() {
		return collectedIncentivesForPosition, nil
	}

	// Send the collected incentives to the position's owner.
	pool, err := k.getPoolById(ctx, position.PoolId)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Send the collected incentives to the position's owner from the pool's address.
	if err := k.bankKeeper.SendCoins(ctx, pool.GetAddress(), owner, collectedIncentivesForPosition); err != nil {
		return sdk.Coins{}, err
	}

	return collectedIncentivesForPosition, nil
}

// createIncentive creates an incentive record in state for the given pool
func (k Keeper) createIncentive(ctx sdk.Context, poolId uint64, sender sdk.AccAddress, incentiveDenom string, incentiveAmount sdk.Int, emissionRate sdk.Dec, startTime time.Time, minUptime time.Duration) (types.IncentiveRecord, error) {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return types.IncentiveRecord{}, err
	}

	// Ensure incentive amount is nonzero and nonnegative
	if !incentiveAmount.IsPositive() {
		return types.IncentiveRecord{}, types.NonPositiveIncentiveAmountError{PoolId: poolId, IncentiveAmount: incentiveAmount.ToDec()}
	}

	// Ensure start time is >= current blocktime
	if startTime.Before(ctx.BlockTime()) {
		return types.IncentiveRecord{}, types.StartTimeTooEarlyError{PoolId: poolId, CurrentBlockTime: ctx.BlockTime(), StartTime: startTime}
	}

	// Ensure emission rate is nonzero and nonnegative
	if !emissionRate.IsPositive() {
		return types.IncentiveRecord{}, types.NonPositiveEmissionRateError{PoolId: poolId, EmissionRate: emissionRate}
	}

	// Ensure min uptime is one of the supported periods
	validUptime := false
	for _, supportedUptime := range types.SupportedUptimes {
		if minUptime == supportedUptime {
			validUptime = true
		}
	}
	if !validUptime {
		return types.IncentiveRecord{}, types.InvalidMinUptimeError{PoolId: poolId, MinUptime: minUptime, SupportedUptimes: types.SupportedUptimes}
	}

	// Ensure sender has balance for incentive denom
	incentiveCoin := sdk.NewCoin(incentiveDenom, incentiveAmount)
	senderHasBalance := k.bankKeeper.HasBalance(ctx, sender, incentiveCoin)
	if !senderHasBalance {
		return types.IncentiveRecord{}, types.IncentiveInsufficientBalanceError{PoolId: poolId, IncentiveDenom: incentiveDenom, IncentiveAmount: incentiveAmount}
	}

	// Sync global uptime accumulators to current blocktime to ensure consistency in reward emissions
	err = k.updateUptimeAccumulatorsToNow(ctx, poolId)
	if err != nil {
		return types.IncentiveRecord{}, err
	}

	// Set up incentive record to put in state
	incentiveRecord := types.IncentiveRecord{
		PoolId:           poolId,
		IncentiveDenom:   incentiveDenom,
		IncentiveCreator: sender,
		RemainingAmount:  incentiveAmount.ToDec(),
		EmissionRate:     emissionRate,
		StartTime:        startTime,
		MinUptime:        minUptime,
	}

	// Get all incentive records for uptime
	existingRecordsForUptime, err := k.GetAllIncentiveRecordsForUptime(ctx, poolId, minUptime)
	if err != nil {
		return types.IncentiveRecord{}, err
	}

	// Fixed gas consumption per incentive creation to prevent spam
	ctx.GasMeter().ConsumeGas(uint64(types.BaseGasFeeForNewIncentive*len(existingRecordsForUptime)), "cl incentive creation")

	// Set incentive record in state
	err = k.setIncentiveRecord(ctx, incentiveRecord)
	if err != nil {
		return types.IncentiveRecord{}, err
	}

	// Transfer tokens from sender to pool balance
	if err := k.bankKeeper.SendCoins(ctx, sender, pool.GetAddress(), sdk.NewCoins(incentiveCoin)); err != nil {
		return types.IncentiveRecord{}, err
	}

	return incentiveRecord, nil
}
