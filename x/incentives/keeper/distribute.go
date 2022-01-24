package keeper

import (
	"encoding/binary"
	"fmt"
	"time"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochtypes "github.com/osmosis-labs/osmosis/x/epochs/types"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	db "github.com/tendermint/tm-db"
)

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

func (k Keeper) getToDistributeCoinsFromIterator(ctx sdk.Context, iterator db.Iterator) sdk.Coins {
	return k.getToDistributeCoinsFromGauges(k.getGaugesFromIterator(ctx, iterator))
}

func (k Keeper) getDistributedCoinsFromIterator(ctx sdk.Context, iterator db.Iterator) sdk.Coins {
	return k.getDistributedCoinsFromGauges(k.getGaugesFromIterator(ctx, iterator))
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

// GetLocksElgibleForF1Distribution gets locks that aren't unlocking before the next epoch, thus eligible for gauge incentive
func (k Keeper) GetLocksElgibleForF1Distribution(ctx sdk.Context, denom string, gaugeDuration time.Duration) []lockuptypes.PeriodLock {
	epochInfo := k.GetEpochInfo(ctx)
	timestamp := epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration).Add(gaugeDuration)
	return k.lk.GetLocksValidAfterTimeDenomDuration(ctx, denom, timestamp, gaugeDuration)
}

// getLocksToDistributionWithMaxDuration get locks that are associated to a condition
// and if its by duration, then use the min Duration
func (k Keeper) getLocksToDistributionWithMaxDuration(ctx sdk.Context, distrTo lockuptypes.QueryCondition, minDuration time.Duration) []lockuptypes.PeriodLock {
	switch distrTo.LockQueryType {
	case lockuptypes.ByDuration:
		if distrTo.Duration > minDuration {
			return k.lk.GetLocksLongerThanDurationDenom(ctx, distrTo.Denom, minDuration)
		}
		return k.lk.GetLocksLongerThanDurationDenom(ctx, distrTo.Denom, distrTo.Duration)
	case lockuptypes.ByTime:
		panic("Gauge by time is present!?!? Should have been blocked in ValidateBasic")
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
// locks is expected to be the correct set of lock recipients for this gauge.
func (k Keeper) distributeInternal(
	ctx sdk.Context, gauge types.Gauge, locks []lockuptypes.PeriodLock, distrInfo *distributionInfo) (sdk.Coins, error) {
	totalDistrCoins := sdk.NewCoins()
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

	locksByDenomCache := make(map[string][]lockuptypes.PeriodLock)

	totalDistributedCoins := sdk.Coins{}
	for _, gauge := range gauges {
		// All gauges have a precondition of being ByDuration
		if _, ok := locksByDenomCache[gauge.DistributeTo.Denom]; !ok {
			locksByDenomCache[gauge.DistributeTo.Denom] = k.getLocksToDistributionWithMaxDuration(
				ctx, gauge.DistributeTo, time.Millisecond)
		}
		// get this from memory instead of hitting iterators / underlying stores.
		// due to many details of cacheKVStore, iteration will still cause expensive IAVL reads.
		allLocks := locksByDenomCache[gauge.DistributeTo.Denom]
		filteredLocks := FilterLocksByMinDuration(allLocks, gauge.DistributeTo.Duration)
		gaugeDistributedCoins, err := k.distributeInternal(ctx, gauge, filteredLocks, &distrInfo)
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

// GetRewards returns current estimate of accumulated rewards for specified locks
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

// GetUnlockingLocksBeforeNextEpoch gets locks that are unlocking before the next epoch
func (k Keeper) GetUnlockingLocksBeforeNextEpoch(ctx sdk.Context, denom string, epochTime time.Time, duration time.Duration) []lockuptypes.PeriodLock {
	startTime := epochTime
	endTime := epochTime.Add(duration)
	// case lockuptypes.ByTime:
	// 	return k.lk.GetLocksPastTimeDenom(ctx, distrTo.Denom, distrTo.Timestamp)
	return k.lk.GetUnlockingsBetweenTimeDenom(ctx, denom, startTime, endTime)
}

// F1Distribute updates currentReward in accordance to a gauge for distribution
func (k Keeper) F1Distribute(ctx sdk.Context, gauge *types.Gauge) error {
	remainDistributionCoins := gauge.Coins.Sub(gauge.DistributedCoins)
	remainEpochs := uint64(1)
	if !gauge.IsPerpetual { // set remain epochs when it's not perpetual gauge
		remainEpochs = gauge.NumEpochsPaidOver - gauge.FilledEpochs
		if remainEpochs == 0 { // prevents division by 0
			k.Logger(ctx).Debug(fmt.Sprintf("remainEpochs is 0. Gauge ID = %d", gauge.Id))
			return nil
		}
	}

	denom := gauge.DistributeTo.Denom
	gaugeDuration := gauge.DistributeTo.Duration

	epochInfo := k.GetEpochInfo(ctx)
	epochNumber := epochInfo.CurrentEpoch
	epochStartTime := epochInfo.CurrentEpochStartTime

	locks := k.GetLocksElgibleForF1Distribution(ctx, denom, gaugeDuration)
	lockSum := lockuptypes.SumLocksByDenom(locks, denom)

	searchStart := epochStartTime.Add(gaugeDuration)
	unlockings := k.GetUnlockingLocksBeforeNextEpoch(ctx, denom, searchStart, epochInfo.Duration)

	currentReward, err := k.GetCurrentReward(ctx, denom, gaugeDuration)
	if err != nil {
		return err
	}

	// check if total stake has changed && last processed epoch is different from current epoch
	// if so, add currentReward to historicalReward and reset currentReward
	if !currentReward.TotalShares.Amount.Equal(lockSum) || len(unlockings) > 0 {
		if currentReward.LastProcessedEpoch != epochNumber {
			cumulativeRewardRatio, err := k.CalculateCumulativeRewardRatio(ctx, currentReward, denom, gaugeDuration, epochNumber)
			if err != nil {
				return err
			}

			err = k.SetHistoricalReward(ctx, cumulativeRewardRatio, denom, gaugeDuration, epochNumber)
			if err != nil {
				return err
			}
			currentReward.LastProcessedEpoch = epochNumber
			currentReward.Rewards = sdk.Coins{}
		}
		currentReward.TotalShares = sdk.NewCoin(denom, lockSum)
	}

	// skip gauge process if locked amount is 0
	if currentReward.TotalShares.Amount.GT(sdk.ZeroInt()) {
		for _, coin := range remainDistributionCoins {
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
	if currentReward.LastProcessedEpoch != -1 {
		err = k.SetCurrentReward(ctx, currentReward, denom, gaugeDuration)
		return err
	}
	return nil
}

// CalculateCumulativeRewardRatio calulates the cumulativeRewardRatio given currentReward.
// cumulativeRewardRatio represents reward per share for each denom
func (k Keeper) CalculateCumulativeRewardRatio(ctx sdk.Context, currentReward types.CurrentReward, denom string, duration time.Duration, epochNumber int64) (sdk.DecCoins, error) {
	totalStakes := currentReward.TotalShares.Amount
	prevHistoricalReward, err := k.GetHistoricalReward(ctx, denom, duration, currentReward.LastProcessedEpoch)
	if err != nil {
		return nil, err
	}
	cumulataiveRewardRatioCoins := prevHistoricalReward.CumulativeRewardRatio

	for _, coin := range currentReward.Rewards {
		totalReward := coin.Amount.ToDec()
		if totalReward.IsNegative() {
			return nil, fmt.Errorf("current rewards is negative. denom: %s, duration: %s, reward amount = %d", denom, duration.String(), totalReward)
		}
		currRewardPerShare := sdk.NewDec(0)
		if !totalStakes.IsZero() {
			currRewardPerShare = totalReward.Quo(totalStakes.ToDec())
		}
		cumulataiveRewardRatioCoins = cumulataiveRewardRatioCoins.Add(sdk.NewDecCoinFromDec(coin.Denom, currRewardPerShare))
	}
	return cumulataiveRewardRatioCoins, nil
}

// CalculateRewardForLock gets the most recent lockReward for a periodLock
func (k Keeper) CalculateRewardForLock(ctx sdk.Context, lock lockuptypes.PeriodLock, lockReward types.PeriodLockReward, epochInfo epochtypes.EpochInfo, lockableDuration time.Duration, finishedLock bool) (types.PeriodLockReward, error) {
	for _, coin := range lock.Coins {
		denom := coin.Denom
		currentReward, err := k.GetCurrentReward(ctx, denom, lockableDuration)
		if err != nil {
			return types.PeriodLockReward{}, err
		}

		var latestEpoch int64

		if finishedLock {
			remainEpoch := lock.EndTime.Sub(epochInfo.CurrentEpochStartTime).Nanoseconds() / epochInfo.Duration.Nanoseconds()
			durationInEpoch := lockableDuration.Nanoseconds() / epochInfo.Duration.Nanoseconds()
			epochNumber := epochInfo.CurrentEpoch + remainEpoch - durationInEpoch

			latestEpoch, err = k.GetLatestEpochForHistoricalReward(ctx, denom, lockableDuration, epochNumber)
			if err != nil {
				return types.PeriodLockReward{}, err
			}
		} else {
			// last updated historical reward
			latestEpoch = currentReward.LastProcessedEpoch
		}

		index, exists := k.findEpochForLockReward(lockReward, denom, lockableDuration)
		if exists {
			epoch := lockReward.LastEligibleEpochs[index].Epoch
			reward, err := k.CalculateRewardBetweenEpoch(ctx, denom, lockableDuration, coin.Amount, epoch, latestEpoch)
			if err != nil {
				return types.PeriodLockReward{}, err
			}
			lockReward.Rewards = lockReward.Rewards.Add(reward...)
			lockReward.LastEligibleEpochs[index].Epoch = latestEpoch
		} else {
			lastEligibleEpoch := types.LastEligibleEpochByDurationAndDenom{
				LockDuration: lockableDuration,
				Denom:        denom,
				Epoch:        latestEpoch,
			}
			lockReward.LastEligibleEpochs = append(lockReward.LastEligibleEpochs, &lastEligibleEpoch)
		}
	}
	return lockReward, nil
}

// findEpochForLockReward finds the most recent epoch that was stored for period lock reward
func (k Keeper) findEpochForLockReward(lockReward types.PeriodLockReward, denom string, locakbleDuration time.Duration) (index int, exists bool) {
	for i, lastEligibleEpoch := range lockReward.LastEligibleEpochs {
		if lastEligibleEpoch.Denom == denom && lastEligibleEpoch.LockDuration == locakbleDuration {
			return i, true
		}
	}
	return -1, false
}

// GetRecentEpoch gets the most recent epoch that was stored for historicalReward
func (k Keeper) GetLatestEpochForHistoricalReward(ctx sdk.Context, denom string, lockDuration time.Duration, epochNumber int64) (int64, error) {
	rewardKey := combineKeys(types.KeyHistoricalReward, []byte(denom+"/"+lockDuration.String()))
	endbyte := combineKeys(rewardKey, sdk.Uint64ToBigEndian(uint64(epochNumber)))
	iter := ctx.KVStore(k.storeKey).ReverseIterator(rewardKey, storetypes.InclusiveEndBytes(endbyte))

	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		key = key[len(key)-8:]
		epochNumber := int64(binary.BigEndian.Uint64(key))
		return epochNumber, nil
	}
	return 0, nil
}

// CalculateRewardBetweenEpoch calculates reward eligible to claim given amount of coins locked
func (k Keeper) CalculateRewardBetweenEpoch(ctx sdk.Context, denom string, duration time.Duration, amountLocked sdk.Int, prevEpoch int64, latestEpoch int64) (sdk.Coins, error) {
	totalReward := sdk.Coins{}
	prevHistoricalReward, err := k.GetHistoricalReward(ctx, denom, duration, prevEpoch)
	if err != nil {
		return totalReward, err
	}
	latestHistoricalReward, err := k.GetHistoricalReward(ctx, denom, duration, latestEpoch)
	if err != nil {
		return totalReward, err
	}
	accumReward := latestHistoricalReward.CumulativeRewardRatio.Sub(prevHistoricalReward.CumulativeRewardRatio)
	for _, decCoin := range accumReward {
		if decCoin.IsPositive() {
			reward := decCoin.Amount.Mul(amountLocked.ToDec()).TruncateInt()
			totalReward = totalReward.Add(sdk.NewCoin(decCoin.Denom, reward))
		}
	}
	return totalReward, nil
}

// GetRewardForLock gets all rewards claimable for a lock
func (k Keeper) GetRewardForLock(ctx sdk.Context, lock lockuptypes.PeriodLock, lockReward types.PeriodLockReward) (types.PeriodLockReward, error) {
	epochInfo := k.GetEpochInfo(ctx)
	lockableDurations := k.GetLockableDurations(ctx)

	for _, lockableDuration := range lockableDurations {
		if lockableDuration > lock.Duration {
			continue
		}
		if lock.Coins.Empty() {
			return types.PeriodLockReward{}, fmt.Errorf("getLockRewards failed: there are no coins for lock=%v", lock)
		}

		// find if lock has been finished. if finishedLock=true, we find what the last epoch was for the lock.
		// if finishedLock=false, we use the most recent epoch to calculate reward for lock.
		finishedLock := lock.IsUnlocking() &&
			(lock.EndTime.Before(epochInfo.CurrentEpochStartTime.Add(lockableDuration)) || !lock.EndTime.After(ctx.BlockTime()))

		// if short assignments are used for err, lockReward gets de-referenced
		var err error
		lockReward, err = k.CalculateRewardForLock(ctx, lock, lockReward, epochInfo, lockableDuration, finishedLock)
		if err != nil {
			return lockReward, err
		}
	}

	return lockReward, nil
}

// ClaimLockReward claims all reward for a specific lock
func (k Keeper) ClaimLockReward(ctx sdk.Context, lock lockuptypes.PeriodLock, owner sdk.AccAddress) (sdk.Coins, error) {
	lockID := lock.ID
	lockReward, err := k.GetPeriodLockReward(ctx, lockID)
	if err != nil {
		return nil, err
	}

	err = k.UpdateRewardForAllLockDuration(ctx, lock.Coins, lock.Duration)
	if err != nil {
		return nil, err
	}
	lockReward, err = k.GetRewardForLock(ctx, lock, lockReward)
	if err != nil {
		return nil, err
	}
	sentRewards, err := k.SendPeriodLockRewardToOwner(ctx, lock, lockReward)
	if err != nil {
		return nil, err
	}

	return sentRewards, nil
}

// SendPeriodLockRewardToOwner sends lock reward to owner and initializes lockReward
func (k Keeper) SendPeriodLockRewardToOwner(ctx sdk.Context, lock lockuptypes.PeriodLock, lockReward types.PeriodLockReward) (sdk.Coins, error) {
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

// EstimateLockReward returns an approximate reward claimable for a specific lock
func (k Keeper) EstimateLockReward(ctx sdk.Context, lock lockuptypes.PeriodLock) (types.PeriodLockReward, error) {
	lockID := lock.ID
	lockReward, err := k.GetPeriodLockReward(ctx, lockID)
	if err != nil {
		return types.PeriodLockReward{}, err
	}
	epochInfo := k.GetEpochInfo(ctx)
	lockableDurations := k.GetLockableDurations(ctx)
	lockReward, err = k.GetRewardForLock(ctx, lock, lockReward)
	if err != nil {
		return types.PeriodLockReward{}, err
	}

	for _, lockableDuration := range lockableDurations {
		if lock.Duration < lockableDuration {
			continue
		}

		// find if lock has been finished. if finishedLock=true, we find what the last epoch was for the lock.
		// if finishedLock=false, we use the most recent epoch to calculate reward for lock.
		finishedLock := lock.IsUnlocking() &&
			(lock.EndTime.Before(epochInfo.CurrentEpochStartTime.Add(lockableDuration)) || !lock.EndTime.After(ctx.BlockTime()))

		for _, coin := range lock.Coins {
			denom := coin.Denom
			currentReward, err := k.GetCurrentReward(ctx, denom, lockableDuration)
			if err != nil {
				return types.PeriodLockReward{}, err
			}

			historicalReward := (*types.HistoricalReward)(nil)
			cumulativeRewardRatioCoins := sdk.DecCoins{}
			if finishedLock {
				remainEpoch := lock.EndTime.Sub(epochInfo.CurrentEpochStartTime).Nanoseconds() / epochInfo.Duration.Nanoseconds()
				durationInEpoch := lockableDuration.Nanoseconds() / epochInfo.Duration.Nanoseconds()
				epochNumber := epochInfo.CurrentEpoch + remainEpoch - durationInEpoch

				latestEpoch, err := k.GetLatestEpochForHistoricalReward(ctx, denom, lockableDuration, epochNumber)
				if err != nil {
					return types.PeriodLockReward{}, err
				}

				latestHistoricalReward, err := k.GetHistoricalReward(ctx, denom, lockableDuration, latestEpoch)
				if err != nil {
					panic(err)
				}
				historicalReward = &latestHistoricalReward
			} else if currentReward.LastProcessedEpoch != epochInfo.CurrentEpoch {
				cumulativeRewardRatioCoins, err = k.CalculateCumulativeRewardRatio(ctx, currentReward, denom, lockableDuration, epochInfo.CurrentEpoch)
				if err != nil {
					return types.PeriodLockReward{}, err
				}
				historicalReward = &types.HistoricalReward{
					CumulativeRewardRatio: cumulativeRewardRatioCoins,
				}
			} else {

			}

			if historicalReward != nil {
				prevHistoricalReward, err := k.GetHistoricalReward(ctx, denom, lockableDuration, currentReward.LastProcessedEpoch)
				if err != nil {
					return types.PeriodLockReward{}, err
				}

				estDecCoins := historicalReward.CumulativeRewardRatio.Sub(prevHistoricalReward.CumulativeRewardRatio).MulDec(coin.Amount.ToDec())
				for _, decCoin := range estDecCoins {
					estCoin := sdk.NewCoin(decCoin.Denom, decCoin.Amount.TruncateInt())
					lockReward.Rewards = lockReward.Rewards.Add(estCoin)
				}
			}
		}
	}

	return lockReward, nil
}
