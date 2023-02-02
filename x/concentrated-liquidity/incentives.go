package concentrated_liquidity

import (
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
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
	poolIncentiveRecords := pool.GetPoolIncentives()
	timeElapsed := sdk.NewDec(int64(time.Since(pool.GetLastLiquidityUpdate())))
	uptimeAccums, err := k.getUptimeAccumulators(ctx, poolId)

	if err != nil {
		return err
	}
	
	for uptimeIndex, uptimeAccum := range uptimeAccums {
		// Get relevant uptime-level values
		curUptimeDuration := types.SupportedUptimes[uptimeIndex]

		// Qualifying liquidity is the amount of liquidity that satisfies uptime requirements
		qualifyingLiquidity := uptimeAccum.GetTotalShares()

		rewardsToAddToCurAccum := sdk.NewDecCoins()
		for incentiveIndex, incentiveRecord := range poolIncentiveRecords {
			// We consider all incentives matching the current uptime that began emitting before the current blocktime
			if incentiveRecord.StartTime.Before(ctx.BlockTime()) && incentiveRecord.MinUptime == curUptimeDuration {
				// Total amount emitted = time elapsed * emission
				totalEmittedAmount := timeElapsed.Mul(incentiveRecord.EmissionRate)

				// Incentives to emit per unit of qualifying liquidity = total emitted / qualifying liquidity
				// Note that we truncate to ensure we do not overdistribute incentives
				incentivesPerLiquidity := totalEmittedAmount.Quo(qualifyingLiquidity)
				emittedIncentivesPerLiquidity := sdk.NewDecCoinFromDec(incentiveRecord.IncentiveDenom, incentivesPerLiquidity)

				// Ensure that we only emit if there are enough incentives remaining to be emitted
				remainingRewards := poolIncentiveRecords[incentiveIndex].RemainingAmount
				if totalEmittedAmount.LTE(remainingRewards) {
					// Add incentives to accumulator
					rewardsToAddToCurAccum = rewardsToAddToCurAccum.Add(emittedIncentivesPerLiquidity)

					// Update incentive record to reflect the incentives that were emitted
					remainingRewards = remainingRewards.Sub(totalEmittedAmount)

					// Each incentive record should only be modified once
					poolIncentiveRecords[incentiveIndex].RemainingAmount = remainingRewards
				}
			}
			// Emit incentives to current uptime accumulator
			uptimeAccum.AddToAccumulator(rewardsToAddToCurAccum)
		}

		// Update pool incentive records and LastLiquidityUpdate time in state to reflect emitted incentives
		pool.SetPoolIncentives(poolIncentiveRecords)
		pool.SetLastLiquidityUpdate(ctx.BlockTime())
	}

	return nil
}