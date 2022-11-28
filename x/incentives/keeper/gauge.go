package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	epochtypes "github.com/osmosis-labs/osmosis/v3/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v3/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v3/x/lockup/types"
	db "github.com/tendermint/tm-db"
)

// Iterate over everything in a gauges iterator, until it reaches the end. Return all gauges iterated over.
func (k Keeper) getGaugesFromIterator(ctx sdk.Context, iterator db.Iterator) []types.Gauge {
	gauges := []types.Gauge{}
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		gaugeIDs := []uint64{}
		err := json.Unmarshal(iterator.Value(), &gaugeIDs)
		if err != nil {
			panic(err)
		}
		for _, gaugeID := range gaugeIDs {
			gauge, err := k.GetGaugeByID(ctx, gaugeID)
			if err != nil {
				panic(err)
			}
			gauges = append(gauges, *gauge)
		}
	}
	return gauges
}

// Compute the total amount of coins in all the gauges
func (k Keeper) getCoinsFromGauges(gauges []types.Gauge) sdk.Coins {
	coins := sdk.Coins{}
	for _, gauge := range gauges {
		coins = coins.Add(gauge.Coins...)
	}
	return coins
}

func (k Keeper) getDistributedCoinsFromGauges(gauges []types.Gauge) sdk.Coins {
	coins := sdk.Coins{}
	for _, gauge := range gauges {
		coins = coins.Add(gauge.DistributedCoins...)
	}
	return coins
}

func (k Keeper) getToDistributeCoinsFromGauges(gauges []types.Gauge) sdk.Coins {
	// TODO: Consider optimizing this in the future to only require one iteration over all gauges.
	coins := k.getCoinsFromGauges(gauges)
	distributed := k.getDistributedCoinsFromGauges(gauges)
	return coins.Sub(distributed)
}

func (k Keeper) getCoinsFromIterator(ctx sdk.Context, iterator db.Iterator) sdk.Coins {
	return k.getCoinsFromGauges(k.getGaugesFromIterator(ctx, iterator))
}

func (k Keeper) getToDistributeCoinsFromIterator(ctx sdk.Context, iterator db.Iterator) sdk.Coins {
	return k.getToDistributeCoinsFromGauges(k.getGaugesFromIterator(ctx, iterator))
}

func (k Keeper) getDistributedCoinsFromIterator(ctx sdk.Context, iterator db.Iterator) sdk.Coins {
	return k.getDistributedCoinsFromGauges(k.getGaugesFromIterator(ctx, iterator))
}

// setGauge set the gauge inside store
func (k Keeper) setGauge(ctx sdk.Context, gauge *types.Gauge) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(gauge)
	if err != nil {
		return err
	}
	store.Set(gaugeStoreKey(gauge.Id), bz)
	return nil
}

func (k Keeper) SetGaugeWithRefKey(ctx sdk.Context, gauge *types.Gauge) error {
	err := k.setGauge(ctx, gauge)
	if err != nil {
		return err
	}

	curTime := ctx.BlockTime()
	timeKey := getTimeKey(gauge.StartTime)
	if gauge.IsUpcomingGauge(curTime) {
		if err := k.addGaugeRefByKey(ctx, combineKeys(types.KeyPrefixUpcomingGauges, timeKey), gauge.Id); err != nil {
			return err
		}
	} else if gauge.IsActiveGauge(curTime) {
		if err := k.addGaugeRefByKey(ctx, combineKeys(types.KeyPrefixActiveGauges, timeKey), gauge.Id); err != nil {
			return err
		}
	} else {
		if err := k.addGaugeRefByKey(ctx, combineKeys(types.KeyPrefixFinishedGauges, timeKey), gauge.Id); err != nil {
			return err
		}
	}
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(gauge)
	if err != nil {
		return err
	}
	store.Set(gaugeStoreKey(gauge.Id), bz)
	return nil
}

// CreateGauge create a gauge and send coins to the gauge
func (k Keeper) CreateGauge(ctx sdk.Context, isPerpetual bool, owner sdk.AccAddress, coins sdk.Coins, distrTo lockuptypes.QueryCondition, startTime time.Time, numEpochsPaidOver uint64) (uint64, error) {
	durations := k.GetLockableDurations(ctx)
	if distrTo.LockQueryType == lockuptypes.ByDuration {
		durationOk := false
		for _, duration := range durations {
			if duration == distrTo.Duration {
				durationOk = true
				break
			}
		}
		if !durationOk {
			return 0, fmt.Errorf("invalid duration: %d", distrTo.Duration)
		}
	}

	gauge := types.Gauge{
		Id:                k.getLastGaugeID(ctx) + 1,
		IsPerpetual:       isPerpetual,
		DistributeTo:      distrTo,
		Coins:             coins,
		StartTime:         startTime,
		NumEpochsPaidOver: numEpochsPaidOver,
	}

	if err := k.bk.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, gauge.Coins); err != nil {
		return 0, err
	}

	err := k.setGauge(ctx, &gauge)
	if err != nil {
		return 0, err
	}
	k.setLastGaugeID(ctx, gauge.Id)

	// TODO: Do we need to be concerned with case where this should be ActiveGauges?
	if err := k.addGaugeRefByKey(ctx, combineKeys(types.KeyPrefixUpcomingGauges, getTimeKey(gauge.StartTime)), gauge.Id); err != nil {
		return 0, err
	}
	if err := k.addGaugeIDForDenom(ctx, gauge.Id, gauge.DistributeTo.Denom); err != nil {
		return 0, err
	}
	k.hooks.AfterCreateGauge(ctx, gauge.Id)
	return gauge.Id, nil
}

// AddToGauge add coins to gauge
func (k Keeper) AddToGaugeRewards(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, gaugeID uint64) error {
	gauge, err := k.GetGaugeByID(ctx, gaugeID)
	if err != nil {
		return err
	}
	if err := k.bk.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, coins); err != nil {
		return err
	}

	gauge.Coins = gauge.Coins.Add(coins...)
	err = k.setGauge(ctx, gauge)
	if err != nil {
		return err
	}
	k.hooks.AfterAddToGauge(ctx, gauge.Id)
	return nil
}

// BeginDistribution is a utility to begin distribution for a specific gauge
func (k Keeper) BeginDistribution(ctx sdk.Context, gauge types.Gauge) error {
	// validation for current time and distribution start time
	if ctx.BlockTime().Before(gauge.StartTime) {
		return fmt.Errorf("gauge is not able to start distribution yet: %s >= %s", ctx.BlockTime().String(), gauge.StartTime.String())
	}

	timeKey := getTimeKey(gauge.StartTime)
	if err := k.deleteGaugeRefByKey(ctx, combineKeys(types.KeyPrefixUpcomingGauges, timeKey), gauge.Id); err != nil {
		return err
	}
	if err := k.addGaugeRefByKey(ctx, combineKeys(types.KeyPrefixActiveGauges, timeKey), gauge.Id); err != nil {
		return err
	}
	k.hooks.AfterStartDistribution(ctx, gauge.Id)
	return nil
}

// FinishDistribution is a utility to finish distribution for a specific gauge
func (k Keeper) FinishDistribution(ctx sdk.Context, gauge types.Gauge) error {
	timeKey := getTimeKey(gauge.StartTime)
	if err := k.deleteGaugeRefByKey(ctx, combineKeys(types.KeyPrefixActiveGauges, timeKey), gauge.Id); err != nil {
		return err
	}
	if err := k.addGaugeRefByKey(ctx, combineKeys(types.KeyPrefixFinishedGauges, timeKey), gauge.Id); err != nil {
		return err
	}
	if err := k.deleteGaugeIDForDenom(ctx, gauge.Id, gauge.DistributeTo.Denom); err != nil {
		return err
	}
	k.hooks.AfterFinishDistribution(ctx, gauge.Id)
	return nil
}

// GetLocksToDistribution get locks that are associated to a condition
func (k Keeper) GetLocksToDistribution(ctx sdk.Context, distrTo lockuptypes.QueryCondition) []lockuptypes.PeriodLock {
	switch distrTo.LockQueryType {
	case lockuptypes.ByDuration:
		return k.lk.GetLocksLongerThanDurationDenom(ctx, distrTo.Denom, distrTo.Duration)
	case lockuptypes.ByTime:
		return k.lk.GetLocksPastTimeDenom(ctx, distrTo.Denom, distrTo.Timestamp)
	default:
	}
	return []lockuptypes.PeriodLock{}
}

// FilteredLocksDistributionEst estimate distribution amount coins from gauge for fitting conditions
// Expectation: gauge is a valid gauge
// filteredLocks are all locks that are valid for gauge
// It also applies an update for the gauge, handling the sending of the rewards.
// (Note this update is in-memory, it does not change state.)
func (k Keeper) FilteredLocksDistributionEst(ctx sdk.Context, gauge types.Gauge, filteredLocks []lockuptypes.PeriodLock) (types.Gauge, sdk.Coins, error) {
	TotalAmtLocked := k.lk.GetPeriodLocksAccumulation(ctx, gauge.DistributeTo)
	if TotalAmtLocked.IsZero() {
		return types.Gauge{}, nil, nil
	}

	remainCoins := gauge.Coins.Sub(gauge.DistributedCoins)
	// Remaining epochs is the number of remaining epochs that the gauge will pay out its rewards
	// For a perpetual gauge, it will pay out everything in the next epoch, and we don't make
	// an assumption for what rate it will get refilled at.
	remainEpochs := uint64(1)
	if !gauge.IsPerpetual {
		remainEpochs = gauge.NumEpochsPaidOver - gauge.FilledEpochs
	}
	// TODO: Should this return err
	if remainEpochs == 0 {
		return gauge, sdk.Coins{}, nil
	}

	remainCoinsPerEpoch := sdk.Coins{}
	for _, coin := range remainCoins {
		// distribution amount per epoch = gauge_size / (remain_epochs)
		amt := coin.Amount.QuoRaw(int64(remainEpochs))
		remainCoinsPerEpoch = remainCoinsPerEpoch.Add(sdk.NewCoin(coin.Denom, amt))
	}

	// Now we compute the filtered coins
	filteredDistrCoins := sdk.Coins{}
	if len(filteredLocks) == 0 {
		// If were doing no filtering, we want to calculate the total amount to distributed in
		// the next epoch.
		// distribution in next epoch = gauge_size  / (remain_epochs)
		filteredDistrCoins = remainCoinsPerEpoch
	}
	for _, lock := range filteredLocks {
		denomLockAmt := lock.Coins.AmountOf(gauge.DistributeTo.Denom)

		for _, coin := range remainCoinsPerEpoch {
			// distribution amount = gauge_size * denom_lock_amount / (total_denom_lock_amount * remain_epochs)
			// distribution amount = gauge_size_per_epoch * denom_lock_amount / total_denom_lock_amount
			amt := coin.Amount.Mul(denomLockAmt).Quo(TotalAmtLocked)
			filteredDistrCoins = filteredDistrCoins.Add(sdk.NewCoin(coin.Denom, amt))
		}
	}

	// increase filled epochs after distribution
	gauge.FilledEpochs += 1
	gauge.DistributedCoins = gauge.DistributedCoins.Add(remainCoinsPerEpoch...)

	return gauge, filteredDistrCoins, nil
}

// Distribute coins from gauge according to its conditions
func (k Keeper) Distribute(ctx sdk.Context, gauge types.Gauge) (sdk.Coins, error) {
	totalDistrCoins := sdk.NewCoins()
	locks := k.GetLocksToDistribution(ctx, gauge.DistributeTo)
	lockSum := lockuptypes.SumLocksByDenom(locks, gauge.DistributeTo.Denom)

	if lockSum.IsZero() {
		return nil, nil
	}

	remainCoins := gauge.Coins.Sub(gauge.DistributedCoins)
	remainEpochs := uint64(1)
	if !gauge.IsPerpetual { // set remain epochs when it's not perpetual gauge
		remainEpochs = gauge.NumEpochsPaidOver - gauge.FilledEpochs
	}
	for _, lock := range locks {
		distrCoins := sdk.Coins{}
		for _, coin := range remainCoins {
			// distribution amount = gauge_size * denom_lock_amount / (total_denom_lock_amount * remain_epochs)
			denomLockAmt := lock.Coins.AmountOf(gauge.DistributeTo.Denom)
			amt := coin.Amount.Mul(denomLockAmt).Quo(lockSum.Mul(sdk.NewInt(int64(remainEpochs))))
			if amt.IsPositive() {
				distrCoins = distrCoins.Add(sdk.NewCoin(coin.Denom, amt))
			}
		}
		distrCoins = distrCoins.Sort()
		if distrCoins.Empty() {
			continue
		}
		owner, err := sdk.AccAddressFromBech32(lock.Owner)
		if err != nil {
			return nil, err
		}
		if err := k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, owner, distrCoins); err != nil {
			return nil, err
		}
		totalDistrCoins = totalDistrCoins.Add(distrCoins...)
	}

	// increase filled epochs after distribution
	gauge.FilledEpochs += 1
	gauge.DistributedCoins = gauge.DistributedCoins.Add(totalDistrCoins...)
	err := k.setGauge(ctx, &gauge)
	if err != nil {
		return nil, err
	}

	k.hooks.AfterDistribute(ctx, gauge.Id)
	return totalDistrCoins, nil
}

// GetModuleToDistributeCoins returns sum of to distribute coins for all of the module
func (k Keeper) GetModuleToDistributeCoins(ctx sdk.Context) sdk.Coins {
	activeGaugesDistr := k.getToDistributeCoinsFromIterator(ctx, k.ActiveGaugesIterator(ctx))
	upcomingGaugesDistr := k.getToDistributeCoinsFromIterator(ctx, k.UpcomingGaugesIteratorAfterTime(ctx, ctx.BlockTime()))
	return activeGaugesDistr.Add(upcomingGaugesDistr...)
}

// GetModuleDistributedCoins returns sum of distributed coins so far
func (k Keeper) GetModuleDistributedCoins(ctx sdk.Context) sdk.Coins {
	activeGaugesDistr := k.getDistributedCoinsFromIterator(ctx, k.ActiveGaugesIterator(ctx))
	finishedGaugesDistr := k.getDistributedCoinsFromIterator(ctx, k.FinishedGaugesIterator(ctx))
	return activeGaugesDistr.Add(finishedGaugesDistr...)
}

// GetGaugeByID Returns gauge from gauge ID
func (k Keeper) GetGaugeByID(ctx sdk.Context, gaugeID uint64) (*types.Gauge, error) {
	gauge := types.Gauge{}
	store := ctx.KVStore(k.storeKey)
	gaugeKey := gaugeStoreKey(gaugeID)
	if !store.Has(gaugeKey) {
		return nil, fmt.Errorf("gauge with ID %d does not exist", gaugeID)
	}
	bz := store.Get(gaugeKey)
	err := proto.Unmarshal(bz, &gauge)
	if err != nil {
		return nil, err
	}
	return &gauge, nil
}

// GetGaugeFromIDs returns gauges from gauge ids reference
func (k Keeper) GetGaugeFromIDs(ctx sdk.Context, refValue []byte) ([]types.Gauge, error) {
	gauges := []types.Gauge{}
	gaugeIDs := []uint64{}
	err := json.Unmarshal(refValue, &gaugeIDs)
	if err != nil {
		return gauges, err
	}
	for _, gaugeID := range gaugeIDs {
		gauge, err := k.GetGaugeByID(ctx, gaugeID)
		if err != nil {
			panic(err)
		}
		gauges = append(gauges, *gauge)
	}
	return gauges, nil
}

// GetGauges returns gauges both upcoming and active
func (k Keeper) GetGauges(ctx sdk.Context) []types.Gauge {
	return k.getGaugesFromIterator(ctx, k.GaugesIterator(ctx))
}

func (k Keeper) GetNotFinishedGauges(ctx sdk.Context) []types.Gauge {
	return append(k.GetActiveGauges(ctx), k.GetUpcomingGauges(ctx)...)
}

// GetActiveGauges returns active gauges
func (k Keeper) GetActiveGauges(ctx sdk.Context) []types.Gauge {
	return k.getGaugesFromIterator(ctx, k.ActiveGaugesIterator(ctx))
}

// GetUpcomingGauges returns scheduled gauges
func (k Keeper) GetUpcomingGauges(ctx sdk.Context) []types.Gauge {
	return k.getGaugesFromIterator(ctx, k.UpcomingGaugesIterator(ctx))
}

// GetFinishedGauges returns finished gauges
func (k Keeper) GetFinishedGauges(ctx sdk.Context) []types.Gauge {
	return k.getGaugesFromIterator(ctx, k.FinishedGaugesIterator(ctx))
}

// GetRewardsEst returns rewards estimation at a future specific time
// If locks are nil, it returns the rewards between now and the end epoch associated with address.
// If locks are not nil, it returns all the rewards for the given locks between now and end epoch.
func (k Keeper) GetRewardsEst(ctx sdk.Context, addr sdk.AccAddress, locks []lockuptypes.PeriodLock, endEpoch int64) sdk.Coins {
	// If locks are nil, populate with all locks associated with the address
	if len(locks) == 0 {
		locks = k.lk.GetAccountPeriodLocks(ctx, addr)
	}
	// Get all gauges that reward to these locks
	// First get all the denominations being locked up
	denomSet := map[string]bool{}
	for _, l := range locks {
		for _, c := range l.Coins {
			denomSet[c.Denom] = true
		}
	}
	gauges := []types.Gauge{}
	// initialize gauges to active and upcomings if not set
	for s := range denomSet {
		gaugeIDs := k.getAllGaugeIDsByDenom(ctx, s)
		// Each gauge only rewards locks to one denom, so no duplicates
		for _, id := range gaugeIDs {
			gauge, err := k.GetGaugeByID(ctx, id)
			// Shouldn't happen
			if err != nil {
				return sdk.Coins{}
			}
			gauges = append(gauges, *gauge)
		}
	}

	// estimate rewards
	estimatedRewards := sdk.Coins{}
	epochInfo := k.GetEpochInfo(ctx)

	// no need to change storage while doing estimation and we use cached context
	cacheCtx, _ := ctx.CacheContext()
	for _, gauge := range gauges {
		distrBeginEpoch := epochInfo.CurrentEpoch
		blockTime := ctx.BlockTime()
		if gauge.StartTime.After(blockTime) {
			distrBeginEpoch = epochInfo.CurrentEpoch + 1 + int64(gauge.StartTime.Sub(blockTime)/epochInfo.Duration)
		}

		for epoch := distrBeginEpoch; epoch <= endEpoch; epoch++ {
			newGauge, distrCoins, err := k.FilteredLocksDistributionEst(cacheCtx, gauge, locks)
			if err != nil {
				continue
			}
			estimatedRewards = estimatedRewards.Add(distrCoins...)
			gauge = newGauge
		}
	}

	return estimatedRewards
}

func (k Keeper) GetEpochInfo(ctx sdk.Context) epochtypes.EpochInfo {
	params := k.GetParams(ctx)
	return k.ek.GetEpochInfo(ctx, params.DistrEpochIdentifier)
}
