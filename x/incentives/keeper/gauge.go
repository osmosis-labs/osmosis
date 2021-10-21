package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	epochtypes "github.com/osmosis-labs/osmosis/x/epochs/types"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
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

// GetLocksToF1Distribution Get locks that are eligible to get gauge incentives
func (k Keeper) GetLocksToF1Distribution(ctx sdk.Context, denom string, duration time.Duration) []lockuptypes.PeriodLock {
	epochInfo := k.GetEpochInfo(ctx)
	timestamp := epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration).Add(duration)
	return k.lk.GetLocksValidAfterTimeDenomDuration(ctx, denom, timestamp, duration)
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

// distributionInfo stores all of the information for pent up sends for rewards distributions.
// This enables us to lower the number of events and calls to back
type distributionInfo struct {
	nextID            int
	lockOwnerAddrToID map[string]int
	idToBech32Addr    []string
	idToDecodedAddr   []sdk.AccAddress
	idToDistrCoins    []sdk.Coins
}

func newDistributionInfo() distributionInfo {
	return distributionInfo{
		nextID:            0,
		lockOwnerAddrToID: make(map[string]int),
		idToBech32Addr:    []string{},
		idToDecodedAddr:   []sdk.AccAddress{},
		idToDistrCoins:    []sdk.Coins{},
	}
}

func (d *distributionInfo) addLockRewards(lock lockuptypes.PeriodLock, rewards sdk.Coins) error {
	if id, ok := d.lockOwnerAddrToID[lock.Owner]; ok {
		oldDistrCoins := d.idToDistrCoins[id]
		d.idToDistrCoins[id] = rewards.Add(oldDistrCoins...)
	} else {
		id := d.nextID
		d.nextID += 1
		d.lockOwnerAddrToID[lock.Owner] = id
		decodedOwnerAddr, err := sdk.AccAddressFromBech32(lock.Owner)
		if err != nil {
			return err
		}
		d.idToBech32Addr = append(d.idToBech32Addr, lock.Owner)
		d.idToDecodedAddr = append(d.idToDecodedAddr, decodedOwnerAddr)
		d.idToDistrCoins = append(d.idToDistrCoins, rewards)
	}
	return nil
}

func (k Keeper) doDistributionSends(ctx sdk.Context, distrs *distributionInfo) error {
	numIDs := len(distrs.idToDecodedAddr)
	ctx.Logger().Debug(fmt.Sprintf("Beginning distribution to %d users", numIDs))
	err := k.bk.SendCoinsFromModuleToManyAccounts(
		ctx,
		types.ModuleName,
		distrs.idToDecodedAddr,
		distrs.idToDistrCoins)
	if err != nil {
		return err
	}
	ctx.Logger().Debug("Finished sending, now creating liquidity add events")
	for id := 0; id < numIDs; id++ {
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.TypeEvtDistribution,
				sdk.NewAttribute(types.AttributeReceiver, distrs.idToBech32Addr[id]),
				sdk.NewAttribute(types.AttributeAmount, distrs.idToDistrCoins[id].String()),
			),
		})
	}
	ctx.Logger().Debug(fmt.Sprintf("Finished Distributing to %d users", numIDs))
	return nil
}

// distributeInternal runs the distribution logic for a gauge, and adds the sends to
// the distrInfo computed. It also updates the gauge for the distribution.
func (k Keeper) distributeInternal(
	ctx sdk.Context, gauge types.Gauge, distrInfo *distributionInfo) (sdk.Coins, error) {
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
			denomLockAmt := lock.Coins.AmountOfNoDenomValidation(gauge.DistributeTo.Denom)
			amt := coin.Amount.Mul(denomLockAmt).Quo(lockSum.Mul(sdk.NewInt(int64(remainEpochs))))
			if amt.IsPositive() {
				newlyDistributedCoin := sdk.Coin{Denom: coin.Denom, Amount: amt}
				distrCoins = distrCoins.Add(newlyDistributedCoin)
			}
		}
		distrCoins = distrCoins.Sort()
		if distrCoins.Empty() {
			continue
		}
		// Update the amount for that address
		err := distrInfo.addLockRewards(lock, distrCoins)
		if err != nil {
			return nil, err
		}

		totalDistrCoins = totalDistrCoins.Add(distrCoins...)
	}

	// increase filled epochs after distribution
	gauge.FilledEpochs += 1
	gauge.DistributedCoins = gauge.DistributedCoins.Add(totalDistrCoins...)
	if err := k.setGauge(ctx, &gauge); err != nil {
		return nil, err
	}

	return totalDistrCoins, nil
}

// Distribute coins from gauge according to its conditions
func (k Keeper) Distribute(ctx sdk.Context, gauges []types.Gauge) (sdk.Coins, error) {
	distrInfo := newDistributionInfo()

	totalDistributedCoins := sdk.Coins{}
	for _, gauge := range gauges {
		gaugeDistributedCoins, err := k.distributeInternal(ctx, gauge, &distrInfo)
		if err != nil {
			return nil, err
		}
		totalDistributedCoins = totalDistributedCoins.Add(gaugeDistributedCoins...)
	}

	err := k.doDistributionSends(ctx, &distrInfo)
	if err != nil {
		return nil, err
	}

	k.hooks.AfterEpochDistribution(ctx)
	return totalDistributedCoins, nil
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
	if err := proto.Unmarshal(bz, &gauge); err != nil {
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
			return []types.Gauge{}, err
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

// GetRewards returns current estimate of accumulated rewards
func (k Keeper) GetRewards(ctx sdk.Context, addr sdk.AccAddress, locks []lockuptypes.PeriodLock) (rewards sdk.Coins) {
	if len(locks) == 0 {
		locks = k.lk.GetAccountPeriodLocks(ctx, addr)
	}
	rewards = sdk.Coins{}
	for _, lock := range locks {
		estLockReward, err := k.EstimateLockReward(ctx, lock)
		if err != nil {
			continue
		}
		rewards = rewards.Add(estLockReward.Rewards...)
	}
	return
}

func (k Keeper) GetEpochInfo(ctx sdk.Context) epochtypes.EpochInfo {
	params := k.GetParams(ctx)
	return k.ek.GetEpochInfo(ctx, params.DistrEpochIdentifier)
}

func (k Keeper) SetCurrentReward(ctx sdk.Context, currentReward types.CurrentReward, denom string, lockDuration time.Duration) error {
	store := ctx.KVStore(k.storeKey)
	rewardKey := combineKeys(types.KeyCurrentReward, []byte(denom), []byte(lockDuration.String()))

	bz, err := proto.Marshal(&currentReward)
	if err != nil {
		return err
	}

	store.Set(rewardKey, bz)

	return nil
}

func (k Keeper) GetCurrentReward(ctx sdk.Context, denom string, lockDuration time.Duration) (types.CurrentReward, error) {
	currentReward := types.CurrentReward{}
	currentReward.Coin.Denom = denom
	store := ctx.KVStore(k.storeKey)
	rewardKey := combineKeys(types.KeyCurrentReward, []byte(denom), []byte(lockDuration.String()))

	bz := store.Get(rewardKey)
	if bz == nil {
		currentReward.Period = 1 // starting period is 1
		currentReward.Coin = sdk.NewCoin(denom, sdk.NewInt(0))
		currentReward.LastProcessedEpoch = 0
		return currentReward, nil
	}

	err := proto.Unmarshal(bz, &currentReward)
	if err != nil {
		return currentReward, err
	}
	return currentReward, nil
}

func (k Keeper) addHistoricalRewardRefs(ctx sdk.Context, prefix []byte, period uint64, epochNumber int64) error {
	store := ctx.KVStore(k.storeKey)
	periodBz := sdk.Uint64ToBigEndian(period)
	endKey := combineKeys(prefix, sdk.Uint64ToBigEndian(uint64(epochNumber)))

	if store.Has(endKey) {
		ctx.Logger().Debug(fmt.Sprintf("HistoricalReward with period exist: %d", period))
		return fmt.Errorf("HistoricalReward with period exist: %d", period)
	}

	store.Set(endKey, periodBz)

	return nil
}

func (k Keeper) getHistoricalRewardPeriodByEpoch(ctx sdk.Context, denom string, lockDuration time.Duration, epochNumber int64) (uint64, error) {
	store := ctx.KVStore(k.storeKey)
	period := uint64(0)
	rewardKey := combineKeys(types.KeyHistoricalReward, []byte(denom+"/"+lockDuration.String()), sdk.Uint64ToBigEndian(uint64(epochNumber)))

	bz := store.Get(rewardKey)
	if bz == nil {
		ctx.Logger().Debug(fmt.Sprintf("historical rewards is not present = %d", epochNumber))
		return period, ErrHistoricalRewardNotExists
	}

	period = sdk.BigEndianToUint64(bz)
	return period, nil
}

// func (k Keeper) deleteHistoricalRewardRefs(ctx sdk.Context, prefix []byte, epochNumber uint64, period uint64) error {
// 	store := ctx.KVStore(k.storeKey)
// 	endKey := combineKeys(prefix, sdk.Uint64ToBigEndian(epochNumber))

// 	store.Delete(endKey)

// 	return nil
// }

func (k Keeper) AddHistoricalReward(ctx sdk.Context, historicalReward types.HistoricalReward, denom string, lockDuration time.Duration, period uint64, epochNumber int64) error {
	store := ctx.KVStore(k.storeKey)
	prefix := combineKeys(types.KeyHistoricalReward, []byte(denom+"/"+lockDuration.String()))
	rewardKey := combineKeys(prefix, sdk.Uint64ToBigEndian(period))

	if store.Has(rewardKey) {
		ctx.Logger().Info(fmt.Sprintf("historical reward is already exist. Denom/Duration/Period = %s/%ds/%d", denom, lockDuration, period))
	}

	bz, err := proto.Marshal(&historicalReward)
	if err != nil {
		return err
	}

	store.Set(rewardKey, bz)

	err = k.addHistoricalRewardRefs(ctx, prefix, period, epochNumber)
	return err
}

func (k Keeper) GetHistoricalReward(ctx sdk.Context, denom string, lockDuration time.Duration, period uint64) (types.HistoricalReward, error) {
	historicalReward := types.HistoricalReward{}
	store := ctx.KVStore(k.storeKey)
	rewardKey := combineKeys(types.KeyHistoricalReward, []byte(denom+"/"+lockDuration.String()), sdk.Uint64ToBigEndian(period))

	if period == 0 {
		historicalReward.CummulativeRewardRatio = sdk.DecCoins{}
		return historicalReward, nil
	}

	bz := store.Get(rewardKey)
	if bz == nil {
		return historicalReward, fmt.Errorf("historical rewards is not present = %d", period)
	}

	err := proto.Unmarshal(bz, &historicalReward)
	if err != nil {
		return historicalReward, err
	}
	return historicalReward, nil
}

func (k Keeper) SetPeriodLockReward(ctx sdk.Context, periodLockReward types.PeriodLockReward) error {
	store := ctx.KVStore(k.storeKey)
	rewardKey := combineKeys(types.KeyHistoricalReward, sdk.Uint64ToBigEndian(periodLockReward.ID))

	bz, err := proto.Marshal(&periodLockReward)
	if err != nil {
		return err
	}

	store.Set(rewardKey, bz)

	return nil
}

func (k Keeper) clearPeriodLockReward(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	rewardKey := combineKeys(types.KeyHistoricalReward, sdk.Uint64ToBigEndian(id))
	store.Delete(rewardKey)
}

func (k Keeper) GetPeriodLockReward(ctx sdk.Context, id uint64) (types.PeriodLockReward, error) {
	store := ctx.KVStore(k.storeKey)
	rewardKey := combineKeys(types.KeyHistoricalReward, sdk.Uint64ToBigEndian(id))

	bz := store.Get(rewardKey)
	if bz == nil {
		return types.PeriodLockReward{
			ID:     id,
			Period: make(map[string]uint64),
		}, nil
	}

	periodLockReward := types.PeriodLockReward{}
	err := proto.Unmarshal(bz, &periodLockReward)
	if err != nil {
		return periodLockReward, err
	}
	return periodLockReward, nil
}

// GetLocksToDistribution get locks that are associated to a condition
func (k Keeper) GetUnlockingsToDistribution(ctx sdk.Context, denom string, epochTime time.Time, duration time.Duration) []lockuptypes.PeriodLock {
	startTime := epochTime
	endTime := epochTime.Add(duration)
	// case lockuptypes.ByTime:
	// 	return k.lk.GetLocksPastTimeDenom(ctx, distrTo.Denom, distrTo.Timestamp)
	return k.lk.GetUnlockingsBetweenTimeDenom(ctx, denom, startTime, endTime)
}

func (k Keeper) F1Distribute(ctx sdk.Context, gauge *types.Gauge) error {
	remainCoins := gauge.Coins.Sub(gauge.DistributedCoins)
	remainEpochs := uint64(1)
	if !gauge.IsPerpetual { // set remain epochs when it's not perpetual gauge
		remainEpochs = gauge.NumEpochsPaidOver - gauge.FilledEpochs
		if remainEpochs == 0 { // prevents divide by 0
			k.Logger(ctx).Debug(fmt.Sprintf("remainEpochs is 0. Gauge ID = %d", gauge.Id))
			return nil
		}
	}

	denom := gauge.DistributeTo.Denom
	duration := gauge.DistributeTo.Duration
	currentReward, err := k.GetCurrentReward(ctx, denom, duration)
	if err != nil {
		return err
	}

	epochInfo := k.GetEpochInfo(ctx)
	epochStartTime := epochInfo.CurrentEpochStartTime

	// checking to see if staking ratio has been changed
	locks := k.GetLocksToF1Distribution(ctx, denom, duration)
	lockSum := lockuptypes.SumLocksByDenom(locks, denom)

	searchStart := epochStartTime.Add(duration)
	unlockings := k.GetUnlockingsToDistribution(ctx, denom, searchStart, epochInfo.Duration)

	if (!currentReward.Coin.Amount.Equal(lockSum)) || (len(unlockings) > 0) {
		historicalReward, err := k.CalculateHistoricalRewards(ctx, currentReward, denom, duration, epochInfo)
		if err != nil {
			return err
		}
		if historicalReward != nil { // new historical reward
			err = k.AddHistoricalReward(ctx, *historicalReward, denom, duration, currentReward.Period, epochInfo.CurrentEpoch)
			if err != nil {
				return err
			}
		}

		currentReward.LastProcessedEpoch = epochInfo.CurrentEpoch
	}
	if epochInfo.CurrentEpoch == currentReward.LastProcessedEpoch {
		// Move to Next Period
		currentReward.Period++
		currentReward.Coin = sdk.NewCoin(denom, lockSum)
		currentReward.Rewards = sdk.Coins{}
	}

	// Skip gauge process if locked amount is 0
	if currentReward.Coin.Amount.GT(sdk.NewInt(0)) {
		for _, coin := range remainCoins {
			amt := coin.Amount.Quo(sdk.NewInt(int64(remainEpochs)))
			if amt.IsPositive() {
				currentReward.Rewards = currentReward.Rewards.Add(sdk.NewCoin(coin.Denom, amt))
				gauge.DistributedCoins = gauge.DistributedCoins.Add(sdk.NewCoin(coin.Denom, amt))
			}
		}
		gauge.FilledEpochs += 1
		err := k.setGauge(ctx, gauge)
		if err != nil {
			return err
		}
	}

	err = k.SetCurrentReward(ctx, currentReward, denom, duration)

	return err
}

func (k Keeper) CalculateHistoricalRewards(ctx sdk.Context, currentReward types.CurrentReward, denom string, duration time.Duration, epochInfo epochtypes.EpochInfo) (*types.HistoricalReward, error) {
	currentEpoch := epochInfo.CurrentEpoch

	if currentReward.LastProcessedEpoch != currentEpoch {
		totalStakes := currentReward.Coin.Amount
		prevPeriod := currentReward.Period - 1
		prevHistoricalReward, err := k.GetHistoricalReward(ctx, denom, duration, prevPeriod)
		if err != nil {
			return nil, err
		}
		newHistoricalReward := types.HistoricalReward{
			CummulativeRewardRatio: prevHistoricalReward.CummulativeRewardRatio,
		}

		for _, coin := range currentReward.Rewards {
			totalReward := coin.Amount.ToDec()

			if totalReward.IsNegative() {
				return nil, fmt.Errorf("current rewards is negative. denom: %s, duration: %s, reward amount = %d", denom, duration.String(), totalReward)
			}
			currRewardPerShare := sdk.NewDec(0)
			if !totalStakes.IsZero() {
				currRewardPerShare = totalReward.Quo(totalStakes.ToDec())
			}

			newHistoricalReward.CummulativeRewardRatio = newHistoricalReward.CummulativeRewardRatio.Add(sdk.NewDecCoinFromDec(coin.Denom, currRewardPerShare))
		}

		return &newHistoricalReward, nil
	}

	return nil, nil
}

func (k Keeper) CalculateRewardBetweenPeriod(ctx sdk.Context, denom string, duration time.Duration, amount sdk.Int, currPeriod uint64, prevPeriod uint64) (sdk.Coins, error) {
	totalReward := sdk.Coins{}
	prevHistoricalReward, err := k.GetHistoricalReward(ctx, denom, duration, prevPeriod)
	if err != nil {
		return totalReward, err
	}
	targetHistoricalReward, err := k.GetHistoricalReward(ctx, denom, duration, currPeriod)
	if err != nil {
		return totalReward, err
	}
	accumReward := targetHistoricalReward.CummulativeRewardRatio.Sub(prevHistoricalReward.CummulativeRewardRatio)
	for _, decCoin := range accumReward {
		if decCoin.IsPositive() {
			reward := decCoin.Amount.Mul(amount.ToDec()).TruncateInt()
			totalReward = totalReward.Add(sdk.NewCoin(decCoin.Denom, reward))
		}
	}
	return totalReward, nil
}

func (k Keeper) CalculateRewardForLock(ctx sdk.Context, lock lockuptypes.PeriodLock, lockReward *types.PeriodLockReward, epochInfo epochtypes.EpochInfo, lockableDuration time.Duration, findLastPeriod bool) error {
	for _, coin := range lock.Coins {
		denom := coin.Denom
		currentReward, err := k.GetCurrentReward(ctx, denom, lockableDuration)
		if err != nil {
			return err
		}
		lockRewardKey := denom + "/" + lockableDuration.String()

		targetPeriod := uint64(0)
		if findLastPeriod {
			remainEpoch := lock.EndTime.Sub(epochInfo.CurrentEpochStartTime).Nanoseconds() / epochInfo.Duration.Nanoseconds()
			durationInEpoch := lockableDuration.Nanoseconds() / epochInfo.Duration.Nanoseconds()
			epochNumber := (epochInfo.CurrentEpoch + remainEpoch - durationInEpoch)
			targetPeriod, err = k.getHistoricalRewardPeriodByEpoch(ctx, denom, lockableDuration, epochNumber)
			if err != nil {
				if err == ErrHistoricalRewardNotExists {
					continue
				}
				return err
			}
		} else {
			targetPeriod = currentReward.Period // last updated historical reward
			if epochInfo.CurrentEpoch != currentReward.LastProcessedEpoch {
				targetPeriod--
			}
		}
		period, ok := lockReward.Period[lockRewardKey]
		if ok {
			reward, err := k.CalculateRewardBetweenPeriod(ctx, denom, lockableDuration, coin.Amount, targetPeriod, period)
			if err != nil {
				return err
			}
			lockReward.Rewards = lockReward.Rewards.Add(reward...)
		}
		lockReward.Period[lockRewardKey] = targetPeriod
	}
	return nil
}

func (k Keeper) GetRewardForLock(ctx sdk.Context, lock lockuptypes.PeriodLock, lockReward types.PeriodLockReward, epochInfo epochtypes.EpochInfo, lockableDurations []time.Duration) (types.PeriodLockReward, error) {
	for _, lockableDuration := range lockableDurations {
		if lockableDuration > lock.Duration {
			continue
		}
		// check current reward can be applied to this lock
		if lock.Coins.Empty() {
			return types.PeriodLockReward{}, fmt.Errorf("getLockRewards failed: there are no coins for lock=%v", lock)
		}
		findLastPeriod := lock.IsUnlocking() &&
			(lock.EndTime.Before(epochInfo.CurrentEpochStartTime.Add(lockableDuration)) || !lock.EndTime.After(ctx.BlockTime()))
		err := k.CalculateRewardForLock(ctx, lock, &lockReward, epochInfo, lockableDuration, findLastPeriod)
		if err != nil {
			return lockReward, err
		}

	}
	return lockReward, nil
}

func (k Keeper) UpdateHistoricalReward(ctx sdk.Context, lockedCoins sdk.Coins, lockDuration time.Duration, epochInfo epochtypes.EpochInfo, lockableDurations []time.Duration) error {
	for _, lockableDuration := range lockableDurations {
		if lockDuration < lockableDuration {
			continue
		}
		for _, coin := range lockedCoins {
			denom := coin.Denom
			currentReward, err := k.GetCurrentReward(ctx, denom, lockableDuration)
			if err != nil {
				return err
			}
			historicalReward, err := k.CalculateHistoricalRewards(ctx, currentReward, denom, lockableDuration, epochInfo)
			if err != nil {
				return err
			}
			if historicalReward != nil {
				err = k.AddHistoricalReward(ctx, *historicalReward, denom, lockableDuration, currentReward.Period, epochInfo.CurrentEpoch)
				if err != nil {
					return err
				}
				currentReward.LastProcessedEpoch = epochInfo.CurrentEpoch
				err = k.SetCurrentReward(ctx, currentReward, denom, lockableDuration)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (k Keeper) UpdatePeriodLockReward(ctx sdk.Context, lock lockuptypes.PeriodLock, lockReward types.PeriodLockReward, epochInfo epochtypes.EpochInfo, lockableDurations []time.Duration) error {
	estLockReward, err := k.GetRewardForLock(ctx, lock, lockReward, epochInfo, lockableDurations)
	if err != nil {
		return err
	}

	err = k.SetPeriodLockReward(ctx, estLockReward)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) SendPeriodLockRewardToOwner(ctx sdk.Context, lock lockuptypes.PeriodLock, lockReward types.PeriodLockReward, duration []time.Duration) (sdk.Coins, error) {
	reward := sdk.Coins{}
	owner, err := sdk.AccAddressFromBech32(lock.Owner)
	if err != nil {
		return reward, err
	}

	err = k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, owner, lockReward.Rewards)
	if err != nil {
		return reward, err
	}
	reward = reward.Add(lockReward.Rewards...)
	lockReward.Rewards = sdk.NewCoins()
	err = k.SetPeriodLockReward(ctx, lockReward)
	if err != nil {
		return reward, err
	}
	return reward, nil
}

func (k Keeper) ClaimLockReward(ctx sdk.Context, lock lockuptypes.PeriodLock, owner sdk.AccAddress) (sdk.Coins, error) {
	lockID := lock.ID
	lockReward, err := k.GetPeriodLockReward(ctx, lockID)
	if err != nil {
		return nil, err
	}
	epochInfo := k.GetEpochInfo(ctx)
	lockableDurations := k.GetLockableDurations(ctx)
	err = k.UpdateHistoricalReward(ctx, lock.Coins, lock.Duration, epochInfo, lockableDurations)
	if err != nil {
		return nil, err
	}
	lockReward, err = k.GetRewardForLock(ctx, lock, lockReward, epochInfo, lockableDurations)
	if err != nil {
		return nil, err
	}
	sentRewards, err := k.SendPeriodLockRewardToOwner(ctx, lock, lockReward, lockableDurations)
	if err != nil {
		return nil, err
	}

	return sentRewards, nil
}

func (k Keeper) EstimateLockReward(ctx sdk.Context, lock lockuptypes.PeriodLock) (types.PeriodLockReward, error) {
	lockID := lock.ID
	lockReward, err := k.GetPeriodLockReward(ctx, lockID)
	if err != nil {
		return types.PeriodLockReward{}, err
	}
	epochInfo := k.GetEpochInfo(ctx)
	lockableDurations := k.GetLockableDurations(ctx)
	lockReward, err = k.GetRewardForLock(ctx, lock, lockReward, epochInfo, lockableDurations)
	if err != nil {
		return types.PeriodLockReward{}, err
	}

	for _, lockableDuration := range lockableDurations {
		if lock.Duration < lockableDuration {
			continue
		}
		// check current reward can be applied to this lock
		if lock.IsUnlocking() && lock.EndTime.Before(epochInfo.CurrentEpochStartTime.Add(lockableDuration)) {
			continue
		}
		for _, coin := range lock.Coins {
			denom := coin.Denom
			currentReward, err := k.GetCurrentReward(ctx, denom, lockableDuration)
			if err != nil {
				return types.PeriodLockReward{}, err
			}
			historicalReward, err := k.CalculateHistoricalRewards(ctx, currentReward, denom, lockableDuration, epochInfo)
			if err != nil {
				return types.PeriodLockReward{}, err
			}
			if historicalReward != nil {
				prevHistoricalReward, err := k.GetHistoricalReward(ctx, denom, lockableDuration, currentReward.Period-1)
				if err != nil {
					return types.PeriodLockReward{}, err
				}

				estDecCoins := historicalReward.CummulativeRewardRatio.Sub(prevHistoricalReward.CummulativeRewardRatio).MulDec(coin.Amount.ToDec())
				for _, decCoin := range estDecCoins {
					estCoin := sdk.NewCoin(decCoin.Denom, decCoin.Amount.TruncateInt())
					lockReward.Rewards = lockReward.Rewards.Add(estCoin)
				}
			}
		}
	}

	return lockReward, nil
}
