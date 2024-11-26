package keeper

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	db "github.com/cosmos/cosmos-db"
	"github.com/hashicorp/go-metrics"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/coinutil"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"

	sdkmath "cosmossdk.io/math"
)

var (
	millisecondsInSecDec = osmomath.NewDec(1000)
	zeroInt              = osmomath.ZeroInt()
)

// DistributionValueCache is a cache for when we calculate the minimum value
// an underlying token must be to be distributed.
type DistributionValueCache struct {
	minDistrValue      sdk.Coin
	denomToMinValueMap map[string]osmomath.Int
}

// AllocateAcrossGauges for every gauge in the input, it updates the weights according to the splitting
// policy and allocates the coins to the underlying gauges per the updated weights.
// Note, every group is associated with a group gauge. The distribution to regular gauges
// happens from the group gauge.
// Returns error if:
// - fails to retrieve a group gauge
// - fails to add gauge rewards to any of the underlying gauges in the group.
//
// Silently skips the following errors:
// - syncing group gauge weights fails - this is acceptable. We don't want to fail all distributions
// due to an oddity with one group. The rewards should stay in the group gauge until the next distribution.
// Any issues can be fixed with a major upgrade without any loss of funds.
// - group gauge is inactive - similar reason, don't want to fail all distributions due to non-fatal issue.
//
// ASSUMPTIONS:
// - group cannot outlive the gauges associated with it.
// CONTRACT:
// - every group in the input is active. If inactive group is passed in, it will be skipped.
func (k Keeper) AllocateAcrossGauges(ctx sdk.Context, activeGroups []types.Group) error {
	for _, group := range activeGroups {
		err := k.syncGroupWeights(ctx, group)
		if err != nil {
			telemetry.IncrCounterWithLabels([]string{types.SyncGroupGaugeFailureMetricName}, 1, []metrics.Label{
				{
					Name:  "group_gauge_id",
					Value: strconv.FormatUint(group.GroupGaugeId, 10),
				},
				{
					Name:  "err",
					Value: err.Error(),
				},
			})
			continue
		}

		// Refetch group
		// TODO: consider mutating receiver of syncGroupWeights instead of refetching.
		// https://github.com/osmosis-labs/osmosis/issues/6556
		group, err := k.GetGroupByGaugeID(ctx, group.GroupGaugeId)
		if err != nil {
			return err
		}

		// Get the groupGauge corresponding to the group.
		groupGauge, err := k.GetGaugeByID(ctx, group.GroupGaugeId)
		if err != nil {
			return err
		}

		// If upcoming, skip.
		if !groupGauge.IsActiveGauge(ctx.BlockTime()) {
			ctx.Logger().Debug(fmt.Sprintf("Group %d is not active, skipping", group.GroupGaugeId), "height", ctx.BlockHeight())
			continue
		}

		// Get amount to distribute in coins (based on perpetual or non perpetual group gauge)
		coinsToDistribute := groupGauge.Coins.Sub(groupGauge.DistributedCoins...)
		if !groupGauge.IsPerpetual {
			remainingEpochs := int64(groupGauge.NumEpochsPaidOver - groupGauge.FilledEpochs)

			// Divide each coin by remainingEpochs
			coinutil.QuoRawMut(coinsToDistribute, remainingEpochs)
		}

		// Exit early if nothing to distribute.
		if coinsToDistribute.IsZero() {
			// Update the group gauge despite zero coins distributed to ensure the filled epochs are updated.
			// Either updates the group or deletes it from state if it is finished.
			if err = k.handleGroupPostDistribute(ctx, *groupGauge, coinsToDistribute); err != nil {
				return err
			}
			ctx.Logger().Debug(fmt.Sprintf("Group %d has no coins to distribute, skipping", group.GroupGaugeId), "height", ctx.BlockHeight())
			continue
		}

		ctx.Logger().Debug(fmt.Sprintf("Distributing total amount %s from group %d", coinsToDistribute, group.GroupGaugeId), "height", ctx.BlockHeight())

		// We track this for distributing all the remaining amount in the last
		// gauge without leaving truncation dust.
		amountDistributed := sdk.NewCoins()

		// Define variables for brevity
		totalGroupWeight := group.InternalGaugeInfo.TotalWeight
		gaugeCount := len(group.InternalGaugeInfo.GaugeRecords)

		// Note that if total weight is zero, we expect an error to be returned
		// during syncing and the group silently skipped.
		// However, we return the error here to be safe and to avoid
		// panicking due to dividing by zero below.
		if totalGroupWeight.IsZero() {
			return types.GroupTotalWeightZeroError{GroupID: group.GroupGaugeId}
		}

		// Iterate over underlying gauge records in the group.
		for gaugeIndex, distrRecord := range group.InternalGaugeInfo.GaugeRecords {
			// Between 0 and 1. to determine the pro-rata share of the total amount to distribute
			// TODO: handle division by zero gracefully and update test
			// https://github.com/osmosis-labs/osmosis/issues/6558
			gaugeDistributionRatio := distrRecord.CurrentWeight.ToLegacyDec().Quo(totalGroupWeight.ToLegacyDec())

			// Loop through `coinsToDistribute` and get the amount to distribute to the current gauge
			// based on the distribution ratio.
			currentGaugeCoins := coinutil.MulDec(coinsToDistribute, gaugeDistributionRatio)

			// For the last gauge, distribute all remaining amounts.
			// Special case the last gauge to avoid leaving truncation dust in the group gauge
			// and consume the amounts in-full.
			if gaugeIndex == gaugeCount-1 {
				err = k.addToGaugeRewards(ctx, coinsToDistribute.Sub(amountDistributed...), distrRecord.GaugeId)
				if err != nil {
					// We error in this case instead of silently skipping because AddToGaugeRewards should never fail
					// unless something fundamental has gone wrong.
					//
					// Assumption we are making: no gauge in the group outlives the group.
					return err
				}
				ctx.Logger().Debug(fmt.Sprintf("Distributing %s from group %d to gauge %d", coinsToDistribute, group.GroupGaugeId, distrRecord.GaugeId), "height", ctx.BlockHeight())
				break
			}

			ctx.Logger().Debug(fmt.Sprintf("Distributing %s from group %d to gauge %d", coinsToDistribute, group.GroupGaugeId, distrRecord.GaugeId), "height", ctx.BlockHeight())
			err = k.addToGaugeRewards(ctx, currentGaugeCoins, distrRecord.GaugeId)
			if err != nil {
				// We error in this case instead of silently skipping because AddToGaugeRewards should never fail
				// unless something fundamental has gone wrong.
				//
				// Assumption we are making: no gauge in the group outlives the group.
				return err
			}

			// Update total distribute amount.
			amountDistributed = amountDistributed.Add(currentGaugeCoins...)
		}

		// Either updates the group or deletes it from state if it is finished.
		if err = k.handleGroupPostDistribute(ctx, *groupGauge, coinsToDistribute); err != nil {
			return err
		}

		ctx.Logger().Debug(fmt.Sprintf("Finished distributing from group %d and updated it", group.GroupGaugeId), "height", ctx.BlockHeight())
	}

	return nil
}

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
	return coins.Sub(distributed...)
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

	remainCoins := gauge.Coins.Sub(gauge.DistributedCoins...)
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
	ctx sdk.Context, gauge types.Gauge, locks []*lockuptypes.PeriodLock, distrInfo *distributionInfo, minDistrValueCache *DistributionValueCache,
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

	sortedAndTrimmedQualifiedLocks := make([]*lockuptypes.PeriodLock, curIndex)
	// This is not an issue because we directly
	// use v.index and &v.locks. However, we must be careful not to
	// take the address of &v.
	// nolint: exportloopref
	for _, v := range qualifiedLocksMap {
		v := v
		if v.index < 0 {
			continue
		}
		sortedAndTrimmedQualifiedLocks[v.index] = &v.lock
	}

	return k.distributeInternal(ctx, gauge, sortedAndTrimmedQualifiedLocks, distrInfo, minDistrValueCache)
}

// syncGroupWeights updates the individual and total weights of the group records based on the splitting policy.
// It mutates the passed in object and sets the updated value in state.
// If there is an error, the passed in object is not mutated.
//
// It returns an error if:
// - the splitting policy is not supported
// - a lower level issue arises when syncing weights (e.g. the volume for a linked pool cannot be found under volume-splitting policy)
func (k Keeper) syncGroupWeights(ctx sdk.Context, group types.Group) error {
	if group.SplittingPolicy == types.ByVolume {
		err := k.syncVolumeSplitGroup(ctx, group)
		// This error implies that there was volume initialized at some point
		// but has not been updated since the last epoch.
		// For this case, we accept to fallback to the previous weights.
		if err != nil && !errors.As(err, &types.NoVolumeSinceLastSyncError{}) {
			return err
		}
	} else {
		return types.UnsupportedSplittingPolicyError{GroupGaugeId: group.GroupGaugeId, SplittingPolicy: group.SplittingPolicy}
	}

	return nil
}

// calculateGroupWeights calculates the updated weights of the group records based on the pool volumes.
// It returns the updated group and an error if any. It does not mutate the passed in object.
func (k Keeper) calculateGroupWeights(ctx sdk.Context, group types.Group) (types.Group, error) {
	totalWeight := zeroInt

	// We operate on a deep copy of the given group because we expect to handle specific errors quietly
	// and want to avoid the scenario where the original group gauge is partially mutated in such cases.
	updatedGroup := types.Group{
		GroupGaugeId: group.GroupGaugeId,
		InternalGaugeInfo: types.InternalGaugeInfo{
			TotalWeight:  group.InternalGaugeInfo.TotalWeight,
			GaugeRecords: make([]types.InternalGaugeRecord, len(group.InternalGaugeInfo.GaugeRecords)),
		},
		SplittingPolicy: group.SplittingPolicy,
	}

	// Loop through gauge records and update their state to reflect new pool volumes
	for i, gaugeRecord := range group.InternalGaugeInfo.GaugeRecords {
		gauge, err := k.GetGaugeByID(ctx, gaugeRecord.GaugeId)
		if err != nil {
			return types.Group{}, err
		}

		gaugeType := gauge.DistributeTo.LockQueryType
		gaugeDuration := time.Duration(0)

		if gaugeType == lockuptypes.NoLock {
			// If NoLock, it's a CL pool, so we set the "lockableDuration" to epoch duration
			gaugeDuration = k.GetEpochInfo(ctx).Duration
		} else {
			// Otherwise, it's a balancer pool so we set it to longest lockable duration
			// TODO: add support for CW pools once there's clarity around default gauge type.
			// Tracked in issue https://github.com/osmosis-labs/osmosis/issues/6403
			gaugeDuration, err = k.pik.GetLongestLockableDuration(ctx)
			if err != nil {
				return types.Group{}, err
			}
		}

		// Retrieve pool ID using GetPoolIdFromGaugeId(gaugeId, lockableDuration)
		poolId, err := k.pik.GetPoolIdFromGaugeId(ctx, gaugeRecord.GaugeId, gaugeDuration)
		if err != nil {
			return types.Group{}, err
		}

		// Get new volume for pool. Assert GTE gauge's weight
		cumulativePoolVolume := k.pmk.GetOsmoVolumeForPool(ctx, poolId)

		// If new volume is 0, there was an issue with volume tracking. Return error.
		// We expect this to be handled quietly in update logic but not in init logic.
		// By returning an error, we let the caller decide whether to handle it quietly or not.
		if !cumulativePoolVolume.IsPositive() {
			return types.Group{}, types.NoPoolVolumeError{PoolId: poolId}
		}

		// Update gauge record's weight to new volume - last volume snapshot
		volumeDelta := cumulativePoolVolume.Sub(gaugeRecord.CumulativeWeight)
		if volumeDelta.IsNegative() {
			return types.Group{}, types.CumulativeVolumeDecreasedError{PoolId: poolId, PreviousVolume: gaugeRecord.CumulativeWeight, NewVolume: cumulativePoolVolume}
		}

		// This check implies that there was volume initialized at some point
		// but has not been updated since the last epoch.
		// We expect to handle this in the caller (syncGroupWeights) and
		// fallback to the previous weights in that case.
		if volumeDelta.IsZero() {
			return types.Group{}, types.NoVolumeSinceLastSyncError{PoolID: poolId}
		}

		gaugeRecord.CurrentWeight = volumeDelta

		// Snapshot cumulative volume
		gaugeRecord.CumulativeWeight = cumulativePoolVolume

		// Add new this diff to total weight
		totalWeight = totalWeight.Add(volumeDelta)

		// Mutate original group to ensure changes are tracked
		updatedGroup.InternalGaugeInfo.GaugeRecords[i] = gaugeRecord
	}

	// Update group's total weight
	updatedGroup.InternalGaugeInfo.TotalWeight = totalWeight
	return updatedGroup, nil
}

// syncVolumeSplitGroup syncs a group according to volume splitting policy.
// It mutates the passed in object and sets the updated value in state.
// If there is an error, the passed in object is not mutated.
//
// It returns an error if:
// - the volume for any linked pool is zero or cannot be found
// - the cumulative volume for any linked pool has decreased (should never happen)
func (k Keeper) syncVolumeSplitGroup(ctx sdk.Context, group types.Group) error {
	updatedGroup, err := k.calculateGroupWeights(ctx, group)
	if err != nil {
		return err
	}

	k.SetGroup(ctx, updatedGroup)

	// We return zero here so that the Group with zero total weight is silently skipped in the
	// caller distribution logic.
	if updatedGroup.InternalGaugeInfo.TotalWeight.IsZero() {
		return types.GroupTotalWeightZeroError{GroupID: group.GroupGaugeId}
	}

	return nil
}

// getNoLockGaugeUptime retrieves the uptime corresponding to the passed in gauge.
// For external gauges, it returns the uptime specified in the gauge.
// For internal gauges, it returns the module param for internal gauge uptime.
//
// In either case, if the fetched uptime is invalid or unauthorized, it falls back to a default uptime.
func (k Keeper) getNoLockGaugeUptime(ctx sdk.Context, gauge types.Gauge, poolId uint64) time.Duration {
	// If internal gauge, use InternalUptime param as the gauge's uptime.
	// Otherwise, use the gauge's duration.
	gaugeUptime := gauge.DistributeTo.Duration
	if gauge.DistributeTo.Denom == types.NoLockInternalGaugeDenom(poolId) {
		gaugeUptime = k.GetParams(ctx).InternalUptime
	}

	// Validate that the gauge's corresponding uptime is authorized.
	authorizedUptimes := k.clk.GetParams(ctx).AuthorizedUptimes
	isUptimeAuthorized := false
	for _, authorizedUptime := range authorizedUptimes {
		if gaugeUptime == authorizedUptime {
			isUptimeAuthorized = true
		}
	}

	// If the gauge's uptime is not authorized, we fall back to a default instead of erroring.
	//
	// This is for two reasons:
	// 1. To allow uptimes to be unauthorized without entirely freezing existing gauges
	// 2. To avoid having to do a state migration on existing gauges at time of adding
	// this change, since prior to this, CL gauges were not required to associate with
	// an uptime that was authorized.
	if !isUptimeAuthorized {
		gaugeUptime = types.DefaultConcentratedUptime
	}

	return gaugeUptime
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
	ctx sdk.Context, gauge types.Gauge, locks []*lockuptypes.PeriodLock, distrInfo *distributionInfo, minDistrValueCache *DistributionValueCache,
) (sdk.Coins, error) {
	totalDistrCoins := sdk.NewCoins()

	// Retrieve the min value for distribution.
	// If any distribution amount is valued less than what the param is set, it will be skipped.
	minValueForDistr := minDistrValueCache.minDistrValue

	remainCoins := gauge.Coins.Sub(gauge.DistributedCoins...)

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

		// Get the uptime for the gauge. Note that if the gauge's uptime is not authorized,
		// this falls back to a default value of 1ns.
		gaugeUptime := k.getNoLockGaugeUptime(ctx, gauge, pool.GetId())

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
			emissionRate := osmomath.NewDecFromInt(remainAmountPerEpoch).QuoTruncateMut(osmomath.NewDec(currentEpoch.Duration.Milliseconds()).QuoMut(millisecondsInSecDec))

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
				// The uptime for each distribution is determined by the gauge's duration field.
				// If it is unauthorized, we fall back to a default above.
				gaugeUptime,
			)

			ctx.Logger().Info(fmt.Sprintf("distributeInternal CL for pool id %d finished", pool.GetId()))
			if err != nil {
				return nil, err
			}
			totalDistrCoins = totalDistrCoins.Add(remainCoinPerEpoch)
		}
	} else {
		// This is a standard lock distribution flow that assumes that we have locks associated with the gauge.
		isSpam, totaltotalDistrCoins, err := k.skipSpamGaugeDistribute(ctx, locks, gauge, totalDistrCoins, remainCoins)
		if isSpam {
			return totaltotalDistrCoins, err
		}

		// This is a standard lock distribution flow that assumes that we have locks associated with the gauge.
		denom := lockuptypes.NativeDenom(gauge.DistributeTo.Denom)
		lockSum, err := lockuptypes.SumLocksByDenom(locks, denom)
		if lockSum.IsZero() || err != nil {
			return nil, nil
		}

		// total_denom_lock_amount * remain_epochs
		lockSumTimesRemainingEpochs := lockSum.MulRaw(int64(remainEpochs))
		lockSumTimesRemainingEpochsBi := lockSumTimesRemainingEpochs.BigIntMut()

		for _, lock := range locks {
			distrCoins := sdk.Coins{}
			// too expensive + verbose even in debug mode.
			// ctx.Logger().Debug("distributeInternal, distribute to lock", "module", types.ModuleName, "gaugeId", gauge.Id, "lockId", lock.ID, "remainCons", remainCoins, "height", ctx.BlockHeight())

			denomLockAmt := guaranteedNonzeroCoinAmountOf(lock.Coins, denom).BigIntMut()
			for _, coin := range remainCoins {
				amtInt := sdkmath.NewIntFromBigInt(denomLockAmt)
				amtIntBi := amtInt.BigIntMut()
				// distribution amount = gauge_size * denom_lock_amount / (total_denom_lock_amount * remain_epochs)
				amtIntBi = amtIntBi.Mul(amtIntBi, coin.Amount.BigIntMut())

				// We should check overflow as we switch back to sdk.Int representation.
				// However we know the final result will not be larger than the gauge size,
				// which is bounded to an Int. So we can safely skip this.

				amtIntBi.Quo(amtIntBi, lockSumTimesRemainingEpochsBi)

				// Determine if the value to distribute is worth enough in minValueForDistr denom to be distributed.
				if coin.Denom == minValueForDistr.Denom {
					// If the denom is the same as the minValueForDistr param, no transformation is needed.
					if amtInt.LT(minValueForDistr.Amount) {
						continue
					}
				} else {
					// If the denom is not the minValueForDistr denom, we need to transform the underlying to it.
					// Check if the denom exists in the cached values
					value, ok := minDistrValueCache.denomToMinValueMap[coin.Denom]
					if !ok {
						// Cache miss, figure out the value and add it to the cache
						poolId, err := k.prk.GetPoolForDenomPairNoOrder(ctx, minValueForDistr.Denom, coin.Denom)
						if err != nil {
							// If the pool denom pair pool route does not exist in protorev, we add a zero value to cache to avoid
							// querying the pool again.
							minDistrValueCache.denomToMinValueMap[coin.Denom] = zeroInt
							continue
						}
						swapModule, pool, err := k.pmk.GetPoolModuleAndPool(ctx, poolId)
						if err != nil {
							return nil, err
						}

						minTokenRequiredForDistr, err := swapModule.CalcOutAmtGivenIn(ctx, pool, minValueForDistr, coin.Denom, osmomath.ZeroDec())
						if err != nil {
							return nil, err
						}

						// Add min token required for distribution to the cache
						minDistrValueCache.denomToMinValueMap[coin.Denom] = minTokenRequiredForDistr.Amount

						// Check if the value is worth enough in the token to be distributed.
						if amtInt.LT(minTokenRequiredForDistr.Amount) {
							// The value is not worth enough, continue
							continue
						}
					} else {
						// Cache hit, use the value

						// This route does not exist in protorev so a zero value has been added when a cache miss occurred
						if value.IsZero() {
							continue
						}
						// Check if the underlying is worth enough in the token to be distributed.
						if amtInt.LT(value) {
							continue
						}
					}
				}

				if amtInt.Sign() == 1 {
					newlyDistributedCoin := sdk.Coin{Denom: coin.Denom, Amount: amtInt}
					distrCoins = distrCoins.Add(newlyDistributedCoin)
				}
			}
			if distrCoins.Len() > 1 {
				// Sort makes a runtime copy, due to some interesting golang details.
				distrCoins = distrCoins.Sort()
			}
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

func (k Keeper) skipSpamGaugeDistribute(ctx sdk.Context, locks []*lockuptypes.PeriodLock, gauge types.Gauge, totalDistrCoins sdk.Coins, remainCoins sdk.Coins) (bool, sdk.Coins, error) {
	if len(locks) == 0 {
		return true, nil, nil
	}

	// In this case, remove redundant cases.
	// Namely: gauge empty OR gauge coins undistributable.
	if remainCoins.Empty() {
		ctx.Logger().Debug(fmt.Sprintf("gauge debug, this gauge is empty, why is it being ran %d. Balancer code", gauge.Id))
		err := k.updateGaugePostDistribute(ctx, gauge, totalDistrCoins)
		return true, totalDistrCoins, err
	}

	// Remove some spam gauges that are not worth distributing. (We ignore the denom stake because of tests.)
	if remainCoins.Len() == 1 && remainCoins[0].Amount.LTE(osmomath.NewInt(100)) && remainCoins[0].Denom != "stake" {
		ctx.Logger().Debug(fmt.Sprintf("gauge debug, this gauge is perceived spam, skipping %d", gauge.Id))
		err := k.updateGaugePostDistribute(ctx, gauge, totalDistrCoins)
		return true, totalDistrCoins, err
	}
	return false, totalDistrCoins, nil
}

// faster coins.AmountOf if we know that coins must contain the denom.
// returns a new big int that can be mutated.
func guaranteedNonzeroCoinAmountOf(coins sdk.Coins, denom string) osmomath.Int {
	if coins.Len() == 1 {
		return coins[0].Amount
	}
	return coins.AmountOfNoDenomValidation(denom)
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

// handleGroupPostDistribute handles the post distribution logic for groups and group gauges.
// If group gauge is perpetual or non-perpetual at the last distribution epoch, it will update the gauge in state by increasing filled epochs 1 and updating
// the distributed coins field to add the coins distributed.
// If group gauge is non-perpetual and at the last distribution epoch, it will delete the group and group gauge from state.
// Before deleting the non-perpetual gauge, any difference between distributed coins and gauge's coins is sent to
// community pool.
// CONTRACT: non-perpetual gauge at the last distribution epoch must have distributed all of its coins already
func (k Keeper) handleGroupPostDistribute(ctx sdk.Context, groupGauge types.Gauge, coinsDistributed sdk.Coins) error {
	// Prune expired non-perpetual gauges.
	if groupGauge.IsLastNonPerpetualDistribution() {
		// Send truncation dust to community pool.
		truncationDust, anyNegative := groupGauge.Coins.SafeSub(groupGauge.DistributedCoins.Add(coinsDistributed...)...)
		if !anyNegative && !truncationDust.IsZero() {
			err := k.ck.FundCommunityPool(ctx, truncationDust, k.ak.GetModuleAddress(types.ModuleName))
			if err != nil {
				return err
			}
		}

		// Delete the group.
		store := ctx.KVStore(k.storeKey)
		store.Delete(types.KeyGroupByGaugeID(groupGauge.Id))
		// Delete the group gauge.
		store.Delete(gaugeStoreKey(groupGauge.Id))
	} else {
		// Update total coins distributed and filled epoch of the group gauge.
		if err := k.updateGaugePostDistribute(ctx, groupGauge, coinsDistributed); err != nil {
			return err
		}
	}
	return nil
}

// getDistributeToBaseLocks takes a gauge along with cached period locks by denom and returns locks that must be distributed to
func (k Keeper) getDistributeToBaseLocks(ctx sdk.Context, gauge types.Gauge, cache map[string][]lockuptypes.PeriodLock, scratchSlice *[]*lockuptypes.PeriodLock) []*lockuptypes.PeriodLock {
	// if gauge is empty, don't get the locks
	if gauge.Coins.Empty() {
		return []*lockuptypes.PeriodLock{}
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
	return FilterLocksByMinDuration(allLocks, gauge.DistributeTo.Duration, scratchSlice)
}

// Distribute distributes coins from an array of gauges to all eligible locks and pools in the case of "NoLock" gauges.
// Skips any group gauges as they are handled separately in AllocateAcrossGauges()
// CONTRACT: gauges must be active.
func (k Keeper) Distribute(ctx sdk.Context, gauges []types.Gauge) (sdk.Coins, error) {
	distrInfo := newDistributionInfo()

	locksByDenomCache := make(map[string][]lockuptypes.PeriodLock)
	totalDistributedCoins := sdk.NewCoins()
	scratchSlice := make([]*lockuptypes.PeriodLock, 0, 50000)

	// Instead of re-fetching the minimum value an underlying token must be to meet the minimum
	// requirement for distribution, we cache the values here.
	// While this isn't precise as it doesn't account for price impact, it is good enough for the sole
	// purpose of determining if we should distribute the token or not.
	minDistrValueCache := &DistributionValueCache{
		minDistrValue:      k.GetParams(ctx).MinValueForDistribution,
		denomToMinValueMap: make(map[string]osmomath.Int),
	}

	for _, gauge := range gauges {
		var gaugeDistributedCoins sdk.Coins
		filteredLocks := k.getDistributeToBaseLocks(ctx, gauge, locksByDenomCache, &scratchSlice)
		// send based on synthetic lockup coins if it's distributing to synthetic lockups
		var err error
		if lockuptypes.IsSyntheticDenom(gauge.DistributeTo.Denom) {
			ctx.Logger().Debug("distributeSyntheticInternal, gauge id %d, %d", "module", types.ModuleName, "gaugeId", gauge.Id, "height", ctx.BlockHeight())
			gaugeDistributedCoins, err = k.distributeSyntheticInternal(ctx, gauge, filteredLocks, &distrInfo, minDistrValueCache)
		} else {
			// Do not distribute if LockQueryType = Group, because if we distribute here we will be double distributing.
			if gauge.DistributeTo.LockQueryType == lockuptypes.ByGroup {
				continue
			}

			gaugeDistributedCoins, err = k.distributeInternal(ctx, gauge, filteredLocks, &distrInfo, minDistrValueCache)
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
