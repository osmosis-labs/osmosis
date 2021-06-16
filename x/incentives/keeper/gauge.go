package keeper

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
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
	k.setGauge(ctx, gauge)

	curTime := ctx.BlockTime()
	timeKey := getTimeKey(gauge.StartTime)
	if gauge.IsUpcomingGauge(curTime) {
		k.addGaugeRefByKey(ctx, combineKeys(types.KeyPrefixUpcomingGauges, timeKey), gauge.Id)
	} else if gauge.IsActiveGauge(curTime) {
		k.addGaugeRefByKey(ctx, combineKeys(types.KeyPrefixActiveGauges, timeKey), gauge.Id)
	} else {
		k.addGaugeRefByKey(ctx, combineKeys(types.KeyPrefixFinishedGauges, timeKey), gauge.Id)
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

	k.setGauge(ctx, &gauge)
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
	k.setGauge(ctx, gauge)
	k.hooks.AfterAddToGauge(ctx, gauge.Id)
	return nil
}

// BeginDistribution is a utility to begin distribution for a specific gauge
func (k Keeper) BeginDistribution(ctx sdk.Context, gauge types.Gauge) error {
	// validation for current time and distribution start time
	curTime := ctx.BlockTime()
	if curTime.Before(gauge.StartTime) {
		return fmt.Errorf("gauge is not able to start distribution yet: %s >= %s", curTime.String(), gauge.StartTime.String())
	}

	// addGaugeIDForDenom is already called in CreateGauge
	timeKey := getTimeKey(gauge.StartTime)
	k.deleteGaugeRefByKey(ctx, combineKeys(types.KeyPrefixUpcomingGauges, timeKey), gauge.Id)
	k.addGaugeRefByKey(ctx, combineKeys(types.KeyPrefixActiveGauges, timeKey), gauge.Id)
	k.hooks.AfterStartDistribution(ctx, gauge.Id)
	return nil
}

// FinishDistribution is a utility to finish distribution for a specific gauge
func (k Keeper) FinishDistribution(ctx sdk.Context, gauge types.Gauge) error {
	timeKey := getTimeKey(gauge.StartTime)
	k.deleteGaugeRefByKey(ctx, combineKeys(types.KeyPrefixActiveGauges, timeKey), gauge.Id)
	k.addGaugeRefByKey(ctx, combineKeys(types.KeyPrefixFinishedGauges, timeKey), gauge.Id)
	k.deleteGaugeIDForDenom(ctx, gauge.Id, gauge.DistributeTo.Denom)
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
func (k Keeper) DistributeAllGauges(ctx sdk.Context) (sdk.Coins, error) {
	totalDistributedCoins := sdk.Coins{}
	// Plan:
	// We need to get a map for denom -> active gauges
	// and a map for denom -> locks
	// Both exist.
	// Now we need to get list
	// k.getAllGaugeIDsByDenom()
	// We use total supply to get every denom
	// totalCoins := k.bk.GetSupply(ctx).GetTotal()
	for _, denom := range k.getDenomList(ctx) {
		// fmt.Printf("test %v\n", denom)
		// denom := coin.Denom
		partialCoins, err := k.distributeAllGaugesForDenom(ctx, denom)
		if err != nil {
			return sdk.Coins{}, err
		}
		totalDistributedCoins = totalDistributedCoins.Add(partialCoins...)
	}
	return totalDistributedCoins, nil
}

type durationRewardPerUnitPair struct {
	duration time.Duration
	dec      sdk.DecCoins
}

func (dr durationRewardPerUnitPair) equal(s durationRewardPerUnitPair) bool {
	return dr.duration == s.duration
}

// modified from
// https://stackoverflow.com/questions/42746972/golang-insert-to-a-sorted-slice
// but edited to use our struct, and search by duration
func sortedListInsert(ss []durationRewardPerUnitPair, s durationRewardPerUnitPair) []durationRewardPerUnitPair {
	if len(ss) == 0 {
		return []durationRewardPerUnitPair{s}
	}
	i := sort.Search(len(ss), func(i int) bool {
		return ss[i].duration >= s.duration
	})
	// If the duration is already in there, then we have a second gauge with the same duration.
	// We add the reward contributions together.
	if i != len(ss) && ss[i].equal(s) {
		ss[i].dec = ss[i].dec.Add(s.dec...)
		return ss
	}
	ss = append(ss, s)
	if i == len(ss) {
		return ss
	}
	// shift over everything by 1, and insert new entry accordingly
	copy(ss[i+1:], ss[i:])
	ss[i] = s
	return ss
}

func (k Keeper) payRewardToLock(ctx sdk.Context, lock lockuptypes.PeriodLock, reward sdk.Coins) error {
	owner, err := sdk.AccAddressFromBech32(lock.Owner)
	if err != nil {
		return err
	}
	if reward.Empty() {
		return nil
	}
	err = k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, owner, reward)
	return err
}

// distributeAllGaugesForDenom distributes tokens for all gauges of denom `d`,
// to all lockups of denom `d`.
// It returns the total amount tokens distributed.
// The performance of this function is `O(#lockups_d log_2(#gauges_d) + #gauges_d)`
func (k Keeper) distributeAllGaugesForDenom(
	ctx sdk.Context, denom string) (coins sdk.Coins, err error) {
	totalDistrCoins := sdk.Coins{}
	// We will use the following two properties to compute the number of rewards
	// with the desired efficiency.
	// 1) Linearity of rewards w.r.t. amount locked at duration D.
	//	  Namely, if 1 token locked w/ duration D gets `R` rewards,
	//	  Then k tokens locked w/ duration D would get `kR` rewards.
	// 2) Efficiency of getting rewards per unit locked for duration > {lockup_time}
	//    Let `R_G` be the rewards for gauge G, let `V_G` be the total locked for
	//    a duration greater than G.Duration.
	//	  Then the amount of rewards per unit locked for duration = {lockup time} is:
	//	  `sum_{gauges with duration <= lockup time} R_G / V_G`
	//
	// These imply an algorithm with the stated efficiency goals.
	// In time O(#gauges_denom) we build the list of rewards per unit lockup of duration
	// equal to a gauges duration.
	// Then in time O(#lockups_d log_2(#gauges_d)) we get the closest gauge w/ duration
	// less than that lockups time. Call this R_{L time}. (We do this via a standard
	// binary search true)
	// Then in time  O(#lockups_d) we compute the rewards for that lockup as
	// lockup.Amt * R_{lockup.Time}

	gaugeIDs := k.getAllGaugeIDsByDenom(ctx, denom)
	// all gauges corresponding to the above gaugeIDs, for gauges that are active
	filteredGauges, err := k.activeGaugesFromIDs(ctx, gaugeIDs)
	if err != nil {
		return sdk.Coins{}, err
	}
	if len(filteredGauges) == 0 {
		return sdk.Coins{}, nil
	}
	// fmt.Println("filteredGauges")
	// fmt.Println(filteredGauges)

	// List of (duration, R_G / V_G) pairs, sorted by duration
	rewardSumsPerUnitDenom := []durationRewardPerUnitPair{}
	for _, gauge := range filteredGauges {
		rewardsPerUnit := k.rewardsPerUnitForGauge(ctx, gauge)
		pair := durationRewardPerUnitPair{
			duration: gauge.DistributeTo.GetDuration(),
			dec:      rewardsPerUnit,
		}
		// fmt.Println("Rewards per unit pre-insert ", len(rewardSumsPerUnitDenom))
		// This combines the rewards per unit if the duration is the same
		rewardSumsPerUnitDenom = sortedListInsert(rewardSumsPerUnitDenom, pair)
		// fmt.Println("Rewards per unit ", len(rewardSumsPerUnitDenom))
		// fmt.Println("Rewards sums per unit ", rewardSumsPerUnitDenom,
		// 	rewardSumsPerUnitDenom[0].dec)
	}

	// List of prefix-sums of (duration, R_G / V_G) pairs, sorted by duration
	// So the ith entry (d_i, R_i) represents the following:
	// If I have T tokens locked with duration d, s.t. d_i <= d < d_{i + 1},
	// then my rewards are T * R_i.
	accumRewardsPerUnit := []durationRewardPerUnitPair{}
	// We make a map to track if theres been a lock to a given duration yet
	durationSeenYet := map[time.Duration]bool{}
	for i, v := range rewardSumsPerUnitDenom {
		durationSeenYet[v.duration] = false
		if i == 0 {
			accumRewardsPerUnit = append(accumRewardsPerUnit, v)
			continue
		}
		prev := accumRewardsPerUnit[i-1]
		cur := prev
		cur.dec = cur.dec.Add(v.dec...)
		accumRewardsPerUnit = append(accumRewardsPerUnit, cur)
	}
	// fmt.Println("_sdf", accumRewardsPerUnit, accumRewardsPerUnit[0].dec)

	// Get all relevant locks to these gauges
	minDuration := rewardSumsPerUnitDenom[0].duration - time.Second
	locks := k.lk.GetLocksLongerThanDurationDenom(ctx, denom, minDuration)
	// fmt.Println("asdf", minDuration, locks)

	for _, lock := range locks {
		index := sort.Search(len(accumRewardsPerUnit), func(i int) bool {
			return lock.Duration <= accumRewardsPerUnit[i].duration
		})
		// index is the first entry where lock.duration <= accumDuration.
		// We want to shift it to be first entry where  its >=
		shiftedIndex := index
		if index == len(accumRewardsPerUnit) ||
			lock.Duration < accumRewardsPerUnit[index].duration {
			shiftedIndex -= 1
		}
		if shiftedIndex < 0 {
			continue
		}
		// Mark things as seen
		if !durationSeenYet[accumRewardsPerUnit[shiftedIndex].duration] {
			for i := shiftedIndex; i >= 0; i-- {
				if durationSeenYet[accumRewardsPerUnit[i].duration] {
					break
				}
				durationSeenYet[accumRewardsPerUnit[i].duration] = true
			}
		}
		rewardPerUnit := accumRewardsPerUnit[shiftedIndex].dec
		amt := lock.Coins.AmountOf(denom)

		distrCoins := sdk.Coins{}
		for i, rewardCoinPerUnit := range rewardPerUnit {
			distrCoins = distrCoins.Add(
				sdk.NewCoin(
					rewardPerUnit[i].Denom,
					rewardCoinPerUnit.Amount.MulInt(amt).RoundInt(),
				),
			)
			// fmt.Println("jsdf", rewardPerUnit, "amt", amt, "distr", distrCoins)
		}
		// Payout reward to lock
		err = k.payRewardToLock(ctx, lock, distrCoins)
		if err != nil {
			return totalDistrCoins, err
		}

		totalDistrCoins = totalDistrCoins.Add(distrCoins...)
	}

	sumGaugeRewards := sdk.Coins{}
	// Handle "cleanup" for gauges
	for _, gauge := range filteredGauges {
		if durationSeenYet[gauge.DistributeTo.Duration] {
			gaugeRewards := k.rewardsForGauge(ctx, gauge)
			gauge.FilledEpochs += 1
			gauge.DistributedCoins = gauge.DistributedCoins.Add(gaugeRewards...)
			// fmt.Println("distr coins update", i, gauge.DistributedCoins)
			k.setGauge(ctx, &gauge)
			sumGaugeRewards = sumGaugeRewards.Add(gaugeRewards...)
		}

		k.hooks.AfterDistribute(ctx, gauge.Id)

		// if !gauge.IsPerpetual && gauge.NumEpochsPaidOver <= gauge.FilledEpochs {
		// 	k.FinishDistribution(ctx, gauge)
		// }
	}
	// if !sumGaugeRewards.IsEqual(totalDistrCoins) {
	// 	panic(fmt.Sprintf("basic %v, %v, %v, \n filtered gauges %v,\n locks %v\n", sumGaugeRewards, totalDistrCoins, totalDistrCoins.Sub(sumGaugeRewards),
	// 		filteredGauges, locks))
	// } else {
	// 	fmt.Println("Working correctly")
	// }
	return totalDistrCoins, nil
}

func (k Keeper) activeGaugesFromIDs(ctx sdk.Context, ids []uint64) ([]types.Gauge, error) {
	activeGauges := []types.Gauge{}
	for _, id := range ids {
		gauge, err := k.GetGaugeByID(ctx, id)
		if err != nil {
			return nil, err
		}
		// fmt.Println(gauge)
		if gauge.IsActiveGauge(ctx.BlockTime()) {
			activeGauges = append(activeGauges, *gauge)
			// fmt.Println(activeGauges)
		}
	}
	return activeGauges, nil
}

func (k Keeper) rewardsForGauge(ctx sdk.Context, gauge types.Gauge) sdk.Coins {
	if gauge.DistributedCoins.IsAnyGTE(gauge.Coins) {
		return sdk.Coins{}
	}
	remainCoins := gauge.Coins.Sub(gauge.DistributedCoins)
	remainEpochs := uint64(1)
	if !gauge.IsPerpetual { // set remain epochs when it's not perpetual gauge
		remainEpochs = gauge.NumEpochsPaidOver - gauge.FilledEpochs
	}
	for i := 0; i < len(remainCoins); i++ {
		remainCoins[i].Amount = remainCoins[i].Amount.QuoRaw(int64(remainEpochs))
		if remainCoins[i].Amount.IsNegative() {
			fmt.Println("wat")
		}
	}
	return remainCoins
}

func (k Keeper) rewardsPerUnitForGauge(ctx sdk.Context, gauge types.Gauge) sdk.DecCoins {
	if gauge.DistributedCoins.IsAnyGTE(gauge.Coins) {
		return sdk.DecCoins{}
	}
	remainCoins := gauge.Coins.Sub(gauge.DistributedCoins)
	remainEpochs := uint64(1)
	if !gauge.IsPerpetual { // set remain epochs when it's not perpetual gauge
		remainEpochs = gauge.NumEpochsPaidOver - gauge.FilledEpochs
	}
	TotalAmtLocked := k.lk.GetPeriodLocksAccumulation(ctx, gauge.DistributeTo)
	if TotalAmtLocked.IsZero() {
		return sdk.DecCoins{}
	}
	coinsPerUnit := make(sdk.DecCoins, len(remainCoins))
	divisor := TotalAmtLocked.MulRaw(int64(remainEpochs))
	for i := 0; i < len(remainCoins); i++ {
		coinsPerUnit[i].Amount = remainCoins[i].Amount.ToDec().QuoInt(divisor)
		coinsPerUnit[i].Denom = remainCoins[i].Denom
	}
	return coinsPerUnit
}

// Distribute coins from gauge according to its conditions
// func (k Keeper) Distribute(ctx sdk.Context, gauge types.Gauge) (sdk.Coins, error) {
// 	totalDistrCoins := sdk.NewCoins()
// 	locks := k.GetLocksToDistribution(ctx, gauge.DistributeTo)
// 	rewardsPerUnit := k.rewardsPerUnitForGauge(ctx, gauge)
// 	if rewardsPerUnit.Empty() {
// 		return nil, nil
// 	}

// 	for _, lock := range locks {
// 		distrCoins := sdk.Coins{}
// 		for _, rewardPerUnit := range rewardsPerUnit {
// 			// distribution amount = gauge_size * denom_lock_amount / (total_denom_lock_amount * remain_epochs)
// 			denomLockAmt := lock.Coins.AmountOf(gauge.DistributeTo.Denom)
// 			amt := rewardPerUnit.Amount.MulInt(denomLockAmt)
// 			if amt.IsPositive() {
// 				distrCoins = distrCoins.Add(sdk.NewCoin(rewardPerUnit.Denom, amt.RoundInt()))
// 			}
// 		}
// 		distrCoins = distrCoins.Sort()
// 		if distrCoins.Empty() {
// 			continue
// 		}

// 		// Payout reward to lock
// 		err := k.payRewardToLock(ctx, lock, distrCoins)
// 		if err != nil {
// 			return totalDistrCoins, err
// 		}

// 		totalDistrCoins = totalDistrCoins.Add(distrCoins...)
// 	}

// 	// increase filled epochs after distribution
// 	gauge.FilledEpochs += 1
// 	gauge.DistributedCoins = gauge.DistributedCoins.Add(totalDistrCoins...)
// 	k.setGauge(ctx, &gauge)

// 	k.hooks.AfterDistribute(ctx, gauge.Id)
// 	return totalDistrCoins, nil
// }

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
	proto.Unmarshal(bz, &gauge)
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
	for s, _ := range denomSet {
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
	params := k.GetParams(ctx)
	epochInfo := k.ek.GetEpochInfo(ctx, params.DistrEpochIdentifier)

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
