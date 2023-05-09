package keeper

import (
	"errors"
	"fmt"
	"time"

	db "github.com/tendermint/tm-db"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v15/x/pool-incentives/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
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
	nextID            int
	lockOwnerAddrToID map[string]int
	idToBech32Addr    []string
	idToDecodedAddr   []sdk.AccAddress
	idToDistrCoins    []sdk.Coins
}

// newDistributionInfo creates a new distributionInfo struct
func newDistributionInfo() distributionInfo {
	return distributionInfo{
		nextID:            0,
		lockOwnerAddrToID: make(map[string]int),
		idToBech32Addr:    []string{},
		idToDecodedAddr:   []sdk.AccAddress{},
		idToDistrCoins:    []sdk.Coins{},
	}
}

// addLockRewards adds the provided rewards to the lockID mapped to the provided owner address.
func (d *distributionInfo) addLockRewards(owner string, rewards sdk.Coins) error {
	if id, ok := d.lockOwnerAddrToID[owner]; ok {
		oldDistrCoins := d.idToDistrCoins[id]
		d.idToDistrCoins[id] = rewards.Add(oldDistrCoins...)
	} else {
		id := d.nextID
		d.nextID += 1
		d.lockOwnerAddrToID[owner] = id
		decodedOwnerAddr, err := sdk.AccAddressFromBech32(owner)
		if err != nil {
			return err
		}
		d.idToBech32Addr = append(d.idToBech32Addr, owner)
		d.idToDecodedAddr = append(d.idToDecodedAddr, decodedOwnerAddr)
		d.idToDistrCoins = append(d.idToDistrCoins, rewards)
	}
	return nil
}

// doDistributionSends utilizes provided distributionInfo to send coins from the module account to various recipients.
func (k Keeper) doDistributionSends(ctx sdk.Context, distrs *distributionInfo) error {
	numIDs := len(distrs.idToDecodedAddr)
	if numIDs > 0 {
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

// distributeConcentratedLiquidity runs the distribution logic for CL pools only. It creates new incentive record with osmo incentives
// and distributes all the tokens to the dedicated pool
func (k Keeper) distributeConcentratedLiquidity(ctx sdk.Context, poolId uint64, sender sdk.AccAddress, incentiveCoin sdk.Coin, emissionRate sdk.Dec, startTime time.Time, minUptime time.Duration, gauge types.Gauge) error {
	_, err := k.clk.CreateIncentive(ctx,
		poolId,
		sender,
		incentiveCoin,
		emissionRate,
		startTime,
		minUptime,
	)
	if err != nil {
		return err
	}

	// updateGaugePostDistribute adds the coins that were just distributed to the gauge's distributed coins field.
	err = k.updateGaugePostDistribute(ctx, gauge, sdk.NewCoins(incentiveCoin))
	if err != nil {
		return err
	}
	return nil
}

// distributeInternal runs the distribution logic for a gauge, and adds the sends to
// the distrInfo struct. It also updates the gauge for the distribution.
// Locks is expected to be the correct set of lock recipients for this gauge.
func (k Keeper) distributeInternal(
	ctx sdk.Context, gauge types.Gauge, locks []lockuptypes.PeriodLock, distrInfo *distributionInfo,
) (sdk.Coins, error) {
	totalDistrCoins := sdk.NewCoins()
	denom := lockuptypes.NativeDenom(gauge.DistributeTo.Denom)
	lockSum := lockuptypes.SumLocksByDenom(locks, denom)

	if lockSum.IsZero() {
		return nil, nil
	}

	remainCoins := gauge.Coins.Sub(gauge.DistributedCoins)
	// if its a perpetual gauge, we set remaining epochs to 1.
	// otherwise is is a non perpetual gauge and we determine how many epoch payouts are left
	remainEpochs := uint64(1)
	if !gauge.IsPerpetual {
		remainEpochs = gauge.NumEpochsPaidOver - gauge.FilledEpochs
	}

	for _, lock := range locks {
		distrCoins := sdk.Coins{}
		for _, coin := range remainCoins {
			// distribution amount = gauge_size * denom_lock_amount / (total_denom_lock_amount * remain_epochs)
			denomLockAmt := lock.Coins.AmountOfNoDenomValidation(denom)
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
		// update the amount for that address
		err := distrInfo.addLockRewards(lock.Owner, distrCoins)
		if err != nil {
			return nil, err
		}

		totalDistrCoins = totalDistrCoins.Add(distrCoins...)
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

// Distribute distributes coins from an array of gauges to all eligible locks.
func (k Keeper) Distribute(ctx sdk.Context, gauges []types.Gauge) (sdk.Coins, error) {
	distrInfo := newDistributionInfo()

	locksByDenomCache := make(map[string][]lockuptypes.PeriodLock)
	totalDistributedCoins := sdk.NewCoins()

	// get pool Id corresponding to the gaugeId
	incentiveParams := k.GetParams(ctx).DistrEpochIdentifier
	currentEpoch := k.ek.GetEpochInfo(ctx, incentiveParams)

	for _, gauge := range gauges {
		var gaugeDistributedCoins sdk.Coins
		pool, err := k.GetPoolFromGaugeId(ctx, gauge.Id, currentEpoch.Duration)
		// Note: getting NoPoolAssociatedWithGaugeError implies that there is no pool associated with the gauge but we still want to distribute.
		// This happens with superfluid gauges which are not connected to any specific pool directly but, instead,
		// via an intermediary account.
		if err != nil && !errors.Is(err, poolincentivestypes.NoPoolAssociatedWithGaugeError{GaugeId: gauge.Id, Duration: currentEpoch.Duration}) {
			// TODO: add test case to cover this
			return nil, err
		} else if pool != nil && pool.GetType() == poolmanagertypes.Concentrated {
			// only want to run this logic if the gaugeId is associated with CL PoolId
			for _, coin := range gauge.Coins {
				// emissionRate calculates amount of tokens to emit per second
				// for ex: 10000tokens to be distributed over 1day epoch will be 1000 tokens ÷ 86,400 seconds ≈ 0.01157 tokens per second (truncated)
				// Note: reason why we do millisecond conversion is because floats are non-deterministic so if someone refactors this and accidentally
				// uses the return of currEpoch.Duration.Seconds() in math operations, this will lead to an app hash.
				emissionRate := sdk.NewDecFromInt(coin.Amount).QuoTruncate(sdk.NewDec(currentEpoch.Duration.Milliseconds()).QuoInt(sdk.NewInt(1000)))
				err := k.distributeConcentratedLiquidity(ctx,
					pool.GetId(),
					k.ak.GetModuleAddress(types.ModuleName),
					coin,
					emissionRate,
					gauge.GetStartTime(),
					// Note that the minimum uptime does not affect the distribution of incentives from the gauge and
					// thus can be any value authorized by the CL module.
					types.DefaultConcentratedUptime,
					gauge,
				)
				if err != nil {
					return nil, err
				}
				gaugeDistributedCoins = gaugeDistributedCoins.Add(coin)
			}
		} else {
			// Assume that there is no pool associated with the gauge and attempt to distribute to base locks
			filteredLocks := k.getDistributeToBaseLocks(ctx, gauge, locksByDenomCache)
			// send based on synthetic lockup coins if it's distributing to synthetic lockups
			var err error
			if lockuptypes.IsSyntheticDenom(gauge.DistributeTo.Denom) {
				// TODO: add test case to cover this
				gaugeDistributedCoins, err = k.distributeSyntheticInternal(ctx, gauge, filteredLocks, &distrInfo)
			} else {
				gaugeDistributedCoins, err = k.distributeInternal(ctx, gauge, filteredLocks, &distrInfo)
			}
			if err != nil {
				// TODO: add test case to cover this
				return nil, err
			}
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
