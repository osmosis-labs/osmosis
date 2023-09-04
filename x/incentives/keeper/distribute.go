package keeper

import (
	"fmt"
	"time"

	db "github.com/tendermint/tm-db"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v19/x/lockup/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
)

// getDistributedCoinsFromGauges returns coins that have been distributed already from the provided gauges
func (k Keeper) getDistributedCoinsFromGauges(gauges []types.Gauge) sdk.Coins {
	coins := sdk.Coins{}
	for _, gauge := range gauges {
		coins = coins.Add(gauge.DistributedCoins...)
	}
	return coins
}

// getToDistributeCoinsFromGauges returns coins that have not been distributed yet from the provided gauges
func (k Keeper) getToDistributeCoinsFromGauges(gauges []types.Gauge) sdk.Coins {
	coins := sdk.Coins{}
	distributed := sdk.Coins{}

	for _, gauge := range gauges {
		coins = coins.Add(gauge.Coins...)
		distributed = distributed.Add(gauge.DistributedCoins...)
	}
	return coins.Sub(distributed)
}

// getToDistributeCoinsFromIterator utilizes iterator to return a list of gauges.
// From these gauges, coins that have not yet been distributed are returned
func (k Keeper) getToDistributeCoinsFromIterator(ctx sdk.Context, iterator db.Iterator) sdk.Coins {
	return k.getToDistributeCoinsFromGauges(k.getGaugesFromIterator(ctx, iterator))
}

// getDistributedCoinsFromIterator utilizes iterator to return a list of gauges.
// From these gauges, coins that have already been distributed are returned
func (k Keeper) getDistributedCoinsFromIterator(ctx sdk.Context, iterator db.Iterator) sdk.Coins {
	return k.getDistributedCoinsFromGauges(k.getGaugesFromIterator(ctx, iterator))
}

// moveUpcomingGaugeToActiveGauge moves a gauge that has reached it's start time from an upcoming to an active status.
func (k Keeper) moveUpcomingGaugeToActiveGauge(ctx sdk.Context, gauge types.Gauge) error {
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

// moveActiveGaugeToFinishedGauge moves a gauge that has completed its distribution from an active to a finished status.
func (k Keeper) moveActiveGaugeToFinishedGauge(ctx sdk.Context, gauge types.Gauge) error {
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

// getLocksToDistributionWithMaxDuration returns locks that match the provided lockuptypes QueryCondition,
// are greater than the provided minDuration, AND have yet to be distributed to.
func (k Keeper) getLocksToDistributionWithMaxDuration(ctx sdk.Context, distrTo lockuptypes.QueryCondition, minDuration time.Duration) []lockuptypes.PeriodLock {
	switch distrTo.LockQueryType {
	case lockuptypes.ByDuration:
		denom := lockuptypes.NativeDenom(distrTo.Denom)
		if distrTo.Duration > minDuration {
			return k.lk.GetLocksLongerThanDurationDenom(ctx, denom, minDuration)
		}
		return k.lk.GetLocksLongerThanDurationDenom(ctx, distrTo.Denom, distrTo.Duration)
	case lockuptypes.ByTime:
		panic("Gauge by time is present, however is no longer supported. This should have been blocked in ValidateBasic")
	default:
	}
	return []lockuptypes.PeriodLock{}
}

// FilteredLocksDistributionEst estimates distribution amount of coins from gauge.
// It also applies an update for the gauge, handling the sending of the rewards.
// (Note this update is in-memory, it does not change state.)
func (k Keeper) FilteredLocksDistributionEst(ctx sdk.Context, gauge types.Gauge, filteredLocks []lockuptypes.PeriodLock) (types.Gauge, sdk.Coins, bool, error) {
	TotalAmtLocked := k.lk.GetPeriodLocksAccumulation(ctx, gauge.DistributeTo)
	if TotalAmtLocked.IsZero() {
		return types.Gauge{}, nil, false, nil
	}
	if TotalAmtLocked.IsNegative() {
		return types.Gauge{}, nil, true, nil
	}

	remainCoins := gauge.Coins.Sub(gauge.DistributedCoins)
	// remainEpochs is the number of remaining epochs that the gauge will pay out its rewards.
	// for a perpetual gauge, it will pay out everything in the next epoch, and we don't make
	// an assumption of the rate at which it will get refilled at.
	remainEpochs := uint64(1)
	if !gauge.IsPerpetual {
		remainEpochs = gauge.NumEpochsPaidOver - gauge.FilledEpochs
	}
	if remainEpochs == 0 {
		return gauge, sdk.Coins{}, false, nil
	}

	remainCoinsPerEpoch := sdk.Coins{}
	for _, coin := range remainCoins {
		// distribution amount per epoch = gauge_size / (remain_epochs)
		amt := coin.Amount.QuoRaw(int64(remainEpochs))
		remainCoinsPerEpoch = remainCoinsPerEpoch.Add(sdk.NewCoin(coin.Denom, amt))
	}

	// now we compute the filtered coins
	filteredDistrCoins := sdk.Coins{}
	if len(filteredLocks) == 0 {
		// if were doing no filtering, we want to calculate the total amount to distributed in
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

	return gauge, filteredDistrCoins, false, nil
}

// distributionInfo stores all of the information for pent up sends for rewards distributions.
// This enables us to lower the number of events and calls to back.
type distributionInfo struct {
	nextID                        int
	lockOwnerAddrToID             map[string]int
	lockOwnerAddrToRewardReceiver map[string]string
	idToBech32Addr                []string
	idToDecodedRewardReceiverAddr []sdk.AccAddress
	idToDistrCoins                []sdk.Coins
}

// newDistributionInfo creates a new distributionInfo struct
func newDistributionInfo() distributionInfo {
	return distributionInfo{
		nextID:                        0,
		lockOwnerAddrToID:             make(map[string]int),
		lockOwnerAddrToRewardReceiver: make(map[string]string),
		idToBech32Addr:                []string{},
		idToDecodedRewardReceiverAddr: []sdk.AccAddress{},
		idToDistrCoins:                []sdk.Coins{},
	}
}

// addLockRewards adds the provided rewards to the lockID mapped to the provided owner address.
func (d *distributionInfo) addLockRewards(owner, rewardReceiver string, rewards sdk.Coins) error {
	// if we have already added current lock owner's info to distribution Info, simply add reward.
	if id, ok := d.lockOwnerAddrToID[owner]; ok {
		oldDistrCoins := d.idToDistrCoins[id]
		d.idToDistrCoins[id] = rewards.Add(oldDistrCoins...)
	} else { // if this is a new owner that we have not added to distributionInfo yet,
		// add according information to the distributionInfo maps.
		id := d.nextID
		d.nextID += 1
		d.lockOwnerAddrToID[owner] = id
		decodedRewardReceiverAddr, err := sdk.AccAddressFromBech32(rewardReceiver)
		if err != nil {
			return err
		}
		d.idToBech32Addr = append(d.idToBech32Addr, rewardReceiver)
		d.idToDecodedRewardReceiverAddr = append(d.idToDecodedRewardReceiverAddr, decodedRewardReceiverAddr)
		d.idToDistrCoins = append(d.idToDistrCoins, rewards)
	}
	return nil
}

// doDistributionSends utilizes provided distributionInfo to send coins from the module account to various recipients.
func (k Keeper) doDistributionSends(ctx sdk.Context, distrs *distributionInfo) error {
	numIDs := len(distrs.idToDecodedRewardReceiverAddr)
	if numIDs > 0 {
		ctx.Logger().Debug(fmt.Sprintf("Beginning distribution to %d users", numIDs))
		// send rewards from the gauge to the reward receiver address
		err := k.bk.SendCoinsFromModuleToManyAccounts(
			ctx,
			types.ModuleName,
			distrs.idToDecodedRewardReceiverAddr,
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
	}
	return nil
}

// distributeSyntheticInternal runs the distribution logic for a synthetic rewards distribution gauge, and adds the sends to
// the distrInfo struct. It also updates the gauge for the distribution.
// locks is expected to be the correct set of lock recipients for this gauge.
func (k Keeper) distributeSyntheticInternal(
	ctx sdk.Context, gauge types.Gauge, locks []lockuptypes.PeriodLock, distrInfo *distributionInfo,
) (sdk.Coins, error) {
	qualifiedLocks := k.lk.GetLocksLongerThanDurationDenom(ctx, gauge.DistributeTo.Denom, gauge.DistributeTo.Duration)

	// map from lockID to present index in resultant list
	// to be state compatible with what we had before, we iterate over locks, to get qualified locks
	// to be in the same order as what is present in locks.
	// in a future release, we can just use qualified locks directly.
	type lockIndexPair struct {
		lock  lockuptypes.PeriodLock
		index int
	}
	qualifiedLocksMap := make(map[uint64]lockIndexPair, len(qualifiedLocks))
	for _, lock := range qualifiedLocks {
		qualifiedLocksMap[lock.ID] = lockIndexPair{lock, -1}
	}
	curIndex := 0
	for _, lock := range locks {
		if v, ok := qualifiedLocksMap[lock.ID]; ok {
			qualifiedLocksMap[lock.ID] = lockIndexPair{v.lock, curIndex}
			curIndex += 1
		}
	}

	sortedAndTrimmedQualifiedLocks := make([]lockuptypes.PeriodLock, curIndex)
	for _, v := range qualifiedLocksMap {
		if v.index < 0 {
			continue
		}
		sortedAndTrimmedQualifiedLocks[v.index] = v.lock
	}

	return k.distributeInternal(ctx, gauge, sortedAndTrimmedQualifiedLocks, distrInfo)
}

// AllocateAcrossGauges gets all the active groupGauges and distributes tokens evenly based on the internalGauges set for that
// groupGauge. After each iteration we update the groupGauge by modifying filledEpoch and distributed coins.
func (k Keeper) AllocateAcrossGauges(ctx sdk.Context) error {
	currTime := ctx.BlockTime()

	groupGauges, err := k.GetAllGroupGauges(ctx)
	if err != nil {
		return err
	}

	for _, groupGauge := range groupGauges {
		gauge, err := k.GetGaugeByID(ctx, groupGauge.GroupGaugeId)
		if err != nil {
			return err
		}

		// only allow distribution if the GroupGauge is Active
		if gauge.IsActiveGauge(currTime) {
			coinsToDistributePerInternalGauge, coinsToDistributeThisEpoch, err := k.calcSplitPolicyCoins(groupGauge.SplittingPolicy, gauge, groupGauge)
			if err != nil {
				return err
			}

			for _, internalGaugeId := range groupGauge.InternalIds {
				err = k.AddToGaugeRewardsFromGauge(ctx, groupGauge.GroupGaugeId, coinsToDistributePerInternalGauge, internalGaugeId)
				if err != nil {
					return err
				}
			}

			// we distribute tokens from groupGauge to internal gauge therefore update groupGauge fields
			// updates filledEpoch and distributedCoins
			if err := k.updateGaugePostDistribute(ctx, *gauge, coinsToDistributeThisEpoch); err != nil {
				return err
			}
		}
	}

	return nil
}

// calcSplitPolicyCoins calculates tokens to split given a policy and groupGauge.
// TODO: add volume split policy
// nolint: unused
func (k Keeper) calcSplitPolicyCoins(policy types.SplittingPolicy, groupGauge *types.Gauge, groupGaugeObj types.GroupGauge) (sdk.Coins, sdk.Coins, error) {
	if policy == types.Evenly {
		remainCoins := groupGauge.Coins.Sub(groupGauge.DistributedCoins)

		var coinsDistPerInternalGauge, coinsDistThisEpoch sdk.Coins
		for _, coin := range remainCoins {
			epochDiff := groupGauge.NumEpochsPaidOver - groupGauge.FilledEpochs
			internalGaugeLen := len(groupGaugeObj.InternalIds)

			distPerEpoch := coin.Amount.Quo(osmomath.NewIntFromUint64(epochDiff))
			distPerGauge := distPerEpoch.Quo(osmomath.NewInt(int64(internalGaugeLen)))

			coinsDistThisEpoch = coinsDistThisEpoch.Add(sdk.NewCoin(coin.Denom, distPerEpoch))
			coinsDistPerInternalGauge = coinsDistPerInternalGauge.Add(sdk.NewCoin(coin.Denom, distPerGauge))
		}

		return coinsDistPerInternalGauge, coinsDistThisEpoch, nil
	} else {
		return nil, nil, fmt.Errorf("GroupGauge id %d doesnot have enought coins to distribute.", &groupGauge.Id)
	}
}

// distributeInternal runs the distribution logic for a gauge, and adds the sends to
// the distrInfo struct. It also updates the gauge for the distribution.
// It handles any kind of gauges:
// - distributing to locks
//   - Locks is expected to be the correct set of lock recipients for this gauge.
//   - perpetual
//   - non-perpetual
//
// - distributing to pools
//   - perpetual
//   - non-perpetual
//
// CONTRACT: gauge passed in as argument must be an active gauge.
func (k Keeper) distributeInternal(
	ctx sdk.Context, gauge types.Gauge, locks []lockuptypes.PeriodLock, distrInfo *distributionInfo,
) (sdk.Coins, error) {
	totalDistrCoins := sdk.NewCoins()

	remainCoins := gauge.Coins.Sub(gauge.DistributedCoins)

	// if its a perpetual gauge, we set remaining epochs to 1.
	// otherwise is is a non perpetual gauge and we determine how many epoch payouts are left
	remainEpochs := uint64(1)
	if !gauge.IsPerpetual {
		remainEpochs = gauge.NumEpochsPaidOver - gauge.FilledEpochs
	}

	// defense in depth
	// this should never happen in practice since gauge passed in should always be an active gauge.
	if remainEpochs == uint64(0) {
		return nil, fmt.Errorf("gauge with id of %d is not active", gauge.Id)
	}

	// This is a no lock distribution flow that assumes that we have a pool associated with the gauge.
	// Currently, this flow is only used for CL pools. Fails if the pool is not found.
	// Fails if the pool found is not a CL pool.
	if gauge.DistributeTo.LockQueryType == lockuptypes.NoLock {
		ctx.Logger().Debug("distributeInternal NoLock gauge", "module", types.ModuleName, "gaugeId", gauge.Id, "height", ctx.BlockHeight())
		pool, err := k.GetPoolFromGaugeId(ctx, gauge.Id, gauge.DistributeTo.Duration)

		if err != nil {
			return nil, err
		}

		poolType := pool.GetType()
		if poolType != poolmanagertypes.Concentrated {
			return nil, fmt.Errorf("pool type %s is not supported for no lock distribution", poolType)
		}

		// Get distribution epoch duration. This is used to calculate the emission rate.
		currentEpoch := k.GetEpochInfo(ctx)

		// For every coin in the gauge, calculate the remaining reward per epoch
		// and create a concentrated liquidity incentive record for it that
		// is supposed to distribute over that epoch.
		for _, remainCoin := range remainCoins {
			// remaining coin amount per epoch.
			remainAmountPerEpoch := remainCoin.Amount.Quo(osmomath.NewIntFromUint64(remainEpochs))
			remainCoinPerEpoch := sdk.NewCoin(remainCoin.Denom, remainAmountPerEpoch)

			// emissionRate calculates amount of tokens to emit per second
			// for ex: 10000uosmo to be distributed over 1day epoch will be 1000 tokens ÷ 86,400 seconds ≈ 0.01157 tokens per second (truncated)
			// Note: reason why we do millisecond conversion is because floats are non-deterministic.
			emissionRate := osmomath.NewDecFromInt(remainAmountPerEpoch).QuoTruncate(osmomath.NewDec(currentEpoch.Duration.Milliseconds()).QuoInt(osmomath.NewInt(1000)))

			ctx.Logger().Info("distributeInternal, CreateIncentiveRecord NoLock gauge", "module", types.ModuleName, "gaugeId", gauge.Id, "poolId", pool.GetId(), "remainCoinPerEpoch", remainCoinPerEpoch, "height", ctx.BlockHeight())
			_, err := k.clk.CreateIncentive(ctx,
				pool.GetId(),
				k.ak.GetModuleAddress(types.ModuleName),
				remainCoinPerEpoch,
				emissionRate,
				// Use current block time as start time, NOT the gauge start time.
				// Gauge start time should be checked whenever moving between active
				// and inactive gauges. By the time we get here, the gauge should be active.
				ctx.BlockTime(),
				// Only default uptime is supported at launch.
				types.DefaultConcentratedUptime,
			)

			ctx.Logger().Info(fmt.Sprintf("distributeInternal CL for pool id %d finished", pool.GetId()))
			if err != nil {
				return nil, err
			}
			totalDistrCoins = totalDistrCoins.Add(remainCoinPerEpoch)
		}
	} else {
		// This is a standard lock distribution flow that assumes that we have locks associated with the gauge.
		if len(locks) == 0 {
			return nil, nil
		}

		// In this case, remove redundant cases.
		// Namely: gauge empty OR gauge coins undistributable.
		if remainCoins.Empty() {
			ctx.Logger().Debug(fmt.Sprintf("gauge debug, this gauge is empty, why is it being ran %d. Balancer code", gauge.Id))
			err := k.updateGaugePostDistribute(ctx, gauge, totalDistrCoins)
			return totalDistrCoins, err
		}

		// Remove some spam gauges, is state compatible.
		// If they're to pool 1 they can't distr at this small of a quantity.
		if remainCoins.Len() == 1 && remainCoins[0].Amount.LTE(osmomath.NewInt(10)) && gauge.DistributeTo.Denom == "gamm/pool/1" && remainCoins[0].Denom != "uosmo" {
			ctx.Logger().Debug(fmt.Sprintf("gauge debug, this gauge is perceived spam, skipping %d", gauge.Id))
			err := k.updateGaugePostDistribute(ctx, gauge, totalDistrCoins)
			return totalDistrCoins, err
		}

		// This is a standard lock distribution flow that assumes that we have locks associated with the gauge.
		denom := lockuptypes.NativeDenom(gauge.DistributeTo.Denom)
		lockSum := lockuptypes.SumLocksByDenom(locks, denom)

		if lockSum.IsZero() {
			return nil, nil
		}

		for _, lock := range locks {
			distrCoins := sdk.Coins{}
			ctx.Logger().Debug("distributeInternal, distribute to lock", "module", types.ModuleName, "gaugeId", gauge.Id, "lockId", lock.ID, "remainCons", remainCoins, "height", ctx.BlockHeight())
			for _, coin := range remainCoins {
				// distribution amount = gauge_size * denom_lock_amount / (total_denom_lock_amount * remain_epochs)
				denomLockAmt := lock.Coins.AmountOfNoDenomValidation(denom)
				amt := coin.Amount.Mul(denomLockAmt).Quo(lockSum.Mul(osmomath.NewInt(int64(remainEpochs))))
				if amt.IsPositive() {
					newlyDistributedCoin := sdk.Coin{Denom: coin.Denom, Amount: amt}
					distrCoins = distrCoins.Add(newlyDistributedCoin)
				}
			}
			distrCoins = distrCoins.Sort()
			if distrCoins.Empty() {
				continue
			}
			// update the amount for that address
			rewardReceiver := lock.RewardReceiverAddress

			// if the reward receiver stored in state is an empty string, it indicates that the owner is the reward receiver.
			if rewardReceiver == "" {
				rewardReceiver = lock.Owner
			}
			err := distrInfo.addLockRewards(lock.Owner, rewardReceiver, distrCoins)
			if err != nil {
				return nil, err
			}

			totalDistrCoins = totalDistrCoins.Add(distrCoins...)
		}
	}

	err := k.updateGaugePostDistribute(ctx, gauge, totalDistrCoins)
	return totalDistrCoins, err
}

// updateGaugePostDistribute increments the gauge's filled epochs field.
// Also adds the coins that were just distributed to the gauge's distributed coins field.
func (k Keeper) updateGaugePostDistribute(ctx sdk.Context, gauge types.Gauge, newlyDistributedCoins sdk.Coins) error {
	gauge.FilledEpochs += 1
	gauge.DistributedCoins = gauge.DistributedCoins.Add(newlyDistributedCoins...)
	if err := k.setGauge(ctx, &gauge); err != nil {
		return err
	}
	return nil
}

// getDistributeToBaseLocks takes a gauge along with cached period locks by denom and returns locks that must be distributed to
func (k Keeper) getDistributeToBaseLocks(ctx sdk.Context, gauge types.Gauge, cache map[string][]lockuptypes.PeriodLock) []lockuptypes.PeriodLock {
	// if gauge is empty, don't get the locks
	if gauge.Coins.Empty() {
		return []lockuptypes.PeriodLock{}
	}
	// Confusingly, there is no way to get all synthetic lockups. Thus we use a separate method `distributeSyntheticInternal` to separately get lockSum for synthetic lockups.
	// All gauges have a precondition of being ByDuration.
	distributeBaseDenom := lockuptypes.NativeDenom(gauge.DistributeTo.Denom)
	if _, ok := cache[distributeBaseDenom]; !ok {
		cache[distributeBaseDenom] = k.getLocksToDistributionWithMaxDuration(
			ctx, gauge.DistributeTo, time.Millisecond)
	}
	// get this from memory instead of hitting iterators / underlying stores.
	// due to many details of cacheKVStore, iteration will still cause expensive IAVL reads.
	allLocks := cache[distributeBaseDenom]
	return FilterLocksByMinDuration(allLocks, gauge.DistributeTo.Duration)
}

// Distribute distributes coins from an array of gauges to all eligible locks and pools in the case of "NoLock" gauges.
// CONTRACT: gauges must be active.
func (k Keeper) Distribute(ctx sdk.Context, gauges []types.Gauge) (sdk.Coins, error) {
	distrInfo := newDistributionInfo()

	locksByDenomCache := make(map[string][]lockuptypes.PeriodLock)
	totalDistributedCoins := sdk.NewCoins()

	for _, gauge := range gauges {
		var gaugeDistributedCoins sdk.Coins
		filteredLocks := k.getDistributeToBaseLocks(ctx, gauge, locksByDenomCache)
		// send based on synthetic lockup coins if it's distributing to synthetic lockups
		var err error
		if lockuptypes.IsSyntheticDenom(gauge.DistributeTo.Denom) {
			ctx.Logger().Debug("distributeSyntheticInternal, gauge id %d, %d", "module", types.ModuleName, "gaugeId", gauge.Id, "height", ctx.BlockHeight())
			gaugeDistributedCoins, err = k.distributeSyntheticInternal(ctx, gauge, filteredLocks, &distrInfo)
		} else {
			// Do not distribute if LockQueryType = Group, because if we distribute here we will be double distributing.
			if gauge.DistributeTo.LockQueryType == lockuptypes.ByGroup {
				continue
			}

			gaugeDistributedCoins, err = k.distributeInternal(ctx, gauge, filteredLocks, &distrInfo)
		}
		if err != nil {
			return nil, err
		}

		totalDistributedCoins = totalDistributedCoins.Add(gaugeDistributedCoins...)
	}

	err := k.doDistributionSends(ctx, &distrInfo)
	if err != nil {
		// TODO: add test case to cover this
		return nil, err
	}

	k.hooks.AfterEpochDistribution(ctx)

	k.checkFinishDistribution(ctx, gauges)
	return totalDistributedCoins, nil
}

// GetPoolFromGaugeId returns a pool associated with the given gauge id.
// Returns error if there is no link between pool id and gauge id.
// Returns error if pool is not saved in state.
func (k Keeper) GetPoolFromGaugeId(ctx sdk.Context, gaugeId uint64, duration time.Duration) (poolmanagertypes.PoolI, error) {
	poolId, err := k.pik.GetPoolIdFromGaugeId(ctx, gaugeId, duration)
	if err != nil {
		return nil, err
	}

	pool, err := k.pmk.GetPool(ctx, poolId)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

// checkFinishDistribution checks if all non perpetual gauges provided have completed their required distributions.
// If complete, move the gauge from an active to a finished status.
func (k Keeper) checkFinishDistribution(ctx sdk.Context, gauges []types.Gauge) {
	for _, gauge := range gauges {
		// filled epoch is increased in this step and we compare with +1
		if !gauge.IsPerpetual && gauge.NumEpochsPaidOver <= gauge.FilledEpochs+1 {
			if err := k.moveActiveGaugeToFinishedGauge(ctx, gauge); err != nil {
				panic(err)
			}
		}
	}
}

// GetModuleToDistributeCoins returns sum of coins yet to be distributed for all of the module.
func (k Keeper) GetModuleToDistributeCoins(ctx sdk.Context) sdk.Coins {
	activeGaugesDistr := k.getToDistributeCoinsFromIterator(ctx, k.ActiveGaugesIterator(ctx))
	upcomingGaugesDistr := k.getToDistributeCoinsFromIterator(ctx, k.UpcomingGaugesIterator(ctx))
	return activeGaugesDistr.Add(upcomingGaugesDistr...)
}

// GetModuleDistributedCoins returns sum of coins that have been distributed so far for all of the module.
func (k Keeper) GetModuleDistributedCoins(ctx sdk.Context) sdk.Coins {
	activeGaugesDistr := k.getDistributedCoinsFromIterator(ctx, k.ActiveGaugesIterator(ctx))
	finishedGaugesDistr := k.getDistributedCoinsFromIterator(ctx, k.FinishedGaugesIterator(ctx))
	return activeGaugesDistr.Add(finishedGaugesDistr...)
}
