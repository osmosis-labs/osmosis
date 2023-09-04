package keeper

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/gogo/protobuf/proto"
	db "github.com/tendermint/tm-db"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v19/x/lockup/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
	epochtypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

// getGaugesFromIterator iterates over everything in a gauge's iterator, until it reaches the end. Return all gauges iterated over.
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

// setGauge set the gauge inside store.
func (k Keeper) setGauge(ctx sdk.Context, gauge *types.Gauge) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(gauge)
	if err != nil {
		return err
	}
	store.Set(gaugeStoreKey(gauge.Id), bz)
	return nil
}

// CreateGaugeRefKeys takes combinedKey (the keyPrefix for upcoming, active, or finished gauges combined with gauge start time) and adds a reference to the respective gauge ID.
// If gauge is active or upcoming, creates reference between the denom and gauge ID.
// Used to consolidate codepaths for InitGenesis and CreateGauge.
func (k Keeper) CreateGaugeRefKeys(ctx sdk.Context, gauge *types.Gauge, combinedKeys []byte, activeOrUpcomingGauge bool) error {
	if err := k.addGaugeRefByKey(ctx, combinedKeys, gauge.Id); err != nil {
		return err
	}
	if activeOrUpcomingGauge {
		if err := k.addGaugeIDForDenom(ctx, gauge.Id, gauge.DistributeTo.Denom); err != nil {
			return err
		}
	}
	return nil
}

// SetGaugeWithRefKey takes a single gauge and assigns a key.
// Takes combinedKey (the keyPrefix for upcoming, active, or finished gauges combined with gauge start time) and adds a reference to the respective gauge ID.
// If this gauge is active or upcoming, creates reference between the denom and gauge ID.
func (k Keeper) SetGaugeWithRefKey(ctx sdk.Context, gauge *types.Gauge) error {
	err := k.setGauge(ctx, gauge)
	if err != nil {
		return err
	}

	curTime := ctx.BlockTime()
	timeKey := getTimeKey(gauge.StartTime)
	activeOrUpcomingGauge := gauge.IsActiveGauge(curTime) || gauge.IsUpcomingGauge(curTime)

	if gauge.IsUpcomingGauge(curTime) {
		combinedKeys := combineKeys(types.KeyPrefixUpcomingGauges, timeKey)
		return k.CreateGaugeRefKeys(ctx, gauge, combinedKeys, activeOrUpcomingGauge)
	} else if gauge.IsActiveGauge(curTime) {
		combinedKeys := combineKeys(types.KeyPrefixActiveGauges, timeKey)
		return k.CreateGaugeRefKeys(ctx, gauge, combinedKeys, activeOrUpcomingGauge)
	} else {
		combinedKeys := combineKeys(types.KeyPrefixFinishedGauges, timeKey)
		return k.CreateGaugeRefKeys(ctx, gauge, combinedKeys, activeOrUpcomingGauge)
	}
}

// CreateGauge creates a gauge with the given parameters and sends coins to the gauge.
// There can be 2 kinds of gauges for a given set of parameters:
// * lockuptypes.ByDuration - a gauge that incentivizes one of the lockable durations.
// For this gauge, the pool id must be 0. Fails if not.
//
// * lockuptypes.NoLock - a gauge that incentivizes pools without locking. Initially,
// this is meant specifically for the concentrated liquidity pools. As a result,
// if NoLock gauge is being created, the given pool id must be non-zero, the pool
// at this id must exist and be of a concentrated liquidity type. Fails if not.
// Additionally, lockuptypes.Denom must be either an empty string, signifying that
// this is an external gauge, or be equal to types.NoLockInternalGaugeDenom(poolId).
// If the denom is empty, it will get overwritten to types.NoLockExternalGaugeDenom(poolId).
// This denom formatting is useful for querying internal vs external gauges associated with a pool.
func (k Keeper) CreateGauge(ctx sdk.Context, isPerpetual bool, owner sdk.AccAddress, coins sdk.Coins, distrTo lockuptypes.QueryCondition, startTime time.Time, numEpochsPaidOver uint64, poolId uint64) (uint64, error) {
	// Ensure that this gauge's duration is one of the allowed durations on chain
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

	nextGaugeId := k.GetLastGaugeID(ctx) + 1

	// For no lock gauges, a pool id must be set.
	// A pool with such id must exist and be a concentrated pool.
	if distrTo.LockQueryType == lockuptypes.NoLock {
		if poolId == 0 {
			return 0, fmt.Errorf("'no lock' type gauges must have a pool id")
		}

		// If not internal gauge denom, then must be set to ""
		// and get overwritten with the external prefix + pool id
		// for internal query purposes.
		distrToDenom := distrTo.Denom
		if distrToDenom != types.NoLockInternalGaugeDenom(poolId) {
			// If denom is set, then fails.
			if distrToDenom != "" {
				return 0, fmt.Errorf("'no lock' type external gauges must have an empty denom set, was %s", distrToDenom)
			}
			distrTo.Denom = types.NoLockExternalGaugeDenom(poolId)
		}

		pool, err := k.pmk.GetPool(ctx, poolId)
		if err != nil {
			return 0, err
		}

		if pool.GetType() != poolmanagertypes.Concentrated {
			return 0, fmt.Errorf("'no lock' type gauges must be created for concentrated pools only")
		}

		// Note that this is a general linking between the gauge and the pool
		// for "NoLock" gauges. It occurs for both external and internal gauges.
		// That being said, internal gauges have an additional linking
		// by duration where duration is the incentives epoch duration.
		// The internal incentive linking is set in x/pool-incentives CreateConcentratedLiquidityPoolGauge.
		k.pik.SetPoolGaugeIdNoLock(ctx, poolId, nextGaugeId)
	} else {
		// For all other gauges, pool id must be 0.
		if poolId != 0 {
			return 0, fmt.Errorf("pool id must be 0 for gauges with lock")
		}

		// check if denom this gauge pays out to exists on-chain
		// N.B.: The reason we check for osmovaloper is to account for gauges that pay out to
		// superfluid synthetic locks. These locks have the following format:
		// "cl/pool/1/superbonding/osmovaloper1wcfyglfgjs2xtsyqu7pl60d0mpw5g7f4wh7pnm"
		// See x/superfluid module README for details.
		if !k.bk.HasSupply(ctx, distrTo.Denom) && !strings.Contains(distrTo.Denom, "osmovaloper") {
			return 0, fmt.Errorf("denom does not exist: %s", distrTo.Denom)
		}
	}

	gauge := types.Gauge{
		Id:                nextGaugeId,
		IsPerpetual:       isPerpetual,
		DistributeTo:      distrTo,
		Coins:             coins,
		StartTime:         startTime,
		NumEpochsPaidOver: numEpochsPaidOver,
	}

	// Fixed gas consumption create gauge based on the number of coins to add
	ctx.GasMeter().ConsumeGas(uint64(types.BaseGasFeeForCreateGauge*len(gauge.Coins)), "scaling gas cost for creating gauge rewards")

	if err := k.bk.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, gauge.Coins); err != nil {
		return 0, err
	}

	err := k.setGauge(ctx, &gauge)
	if err != nil {
		return 0, err
	}
	k.SetLastGaugeID(ctx, gauge.Id)

	combinedKeys := combineKeys(types.KeyPrefixUpcomingGauges, getTimeKey(gauge.StartTime))
	activeOrUpcomingGauge := true

	err = k.CreateGaugeRefKeys(ctx, &gauge, combinedKeys, activeOrUpcomingGauge)
	if err != nil {
		return 0, err
	}
	k.hooks.AfterCreateGauge(ctx, gauge.Id)
	return gauge.Id, nil
}

// CreateGroupGauge creates a new gauge, that allocates rewards dynamically across its internal gauges based on the given splitting policy.
// Note: we should expect that the internal gauges consist of the gauges that are automatically created for each pool upon pool creation, as even non-perpetual
// external incentives would likely want to route through these.
func (k Keeper) CreateGroupGauge(ctx sdk.Context, coins sdk.Coins, numEpochPaidOver uint64, owner sdk.AccAddress, internalGaugeIds []uint64, gaugetype lockuptypes.LockQueryType, splittingPolicy types.SplittingPolicy) (uint64, error) {
	if len(internalGaugeIds) == 0 {
		return 0, fmt.Errorf("No internalGauge provided.")
	}

	if gaugetype != lockuptypes.ByGroup {
		return 0, fmt.Errorf("Invalid gauge type needs to be ByGroup, got %s.", gaugetype)
	}

	// TODO: remove this check once volume splitting is implemented
	if splittingPolicy != types.Evenly {
		return 0, fmt.Errorf("Invalid splitting policy, needs to be Evenly got %s", splittingPolicy)
	}

	// check that all the internalGaugeIds exist
	internalGauges, err := k.GetGaugeFromIDs(ctx, internalGaugeIds)
	if err != nil {
		return 0, fmt.Errorf("Invalid internalGaugeIds, please make sure all the internalGauge have been created.")
	}

	// check that all internalGauges are perp
	for _, gauge := range internalGauges {
		if !gauge.IsPerpetual {
			return 0, fmt.Errorf("Internal Gauge id %d is non-perp, all internalGauge must be perpetual Gauge.", gauge.Id)
		}
	}

	nextGaugeId := k.GetLastGaugeID(ctx) + 1

	gauge := types.Gauge{
		Id:          nextGaugeId,
		IsPerpetual: numEpochPaidOver == 1,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: gaugetype,
		},
		Coins:             coins,
		StartTime:         ctx.BlockTime(),
		NumEpochsPaidOver: numEpochPaidOver,
	}

	if err := k.bk.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, gauge.Coins); err != nil {
		return 0, err
	}

	if err := k.setGauge(ctx, &gauge); err != nil {
		return 0, err
	}

	newGroupGauge := types.GroupGauge{
		GroupGaugeId:    nextGaugeId,
		InternalIds:     internalGaugeIds,
		SplittingPolicy: splittingPolicy,
	}

	k.SetGroupGauge(ctx, newGroupGauge)
	k.SetLastGaugeID(ctx, gauge.Id)

	// TODO: check if this is necessary, will investigate this in following PR.
	combinedKeys := combineKeys(types.KeyPrefixUpcomingGauges, getTimeKey(gauge.StartTime))
	activeOrUpcomingGauge := true

	if err := k.CreateGaugeRefKeys(ctx, &gauge, combinedKeys, activeOrUpcomingGauge); err != nil {
		return 0, err
	}
	k.hooks.AfterCreateGauge(ctx, gauge.Id)

	return nextGaugeId, nil
}

// AddToGaugeRewardsFromGauge transfer coins from groupGaugeId to InternalGaugeId.
// Prior to calling this function, we make sure that the internalGaugeId is linked with the associated groupGaugeId.
// Note: we donot have to bankSend for this gauge transfer because all the available incentive has already been bank sent
// when we create Group Gauge. Now we are just allocating funds from groupGauge to internalGauge.
func (k Keeper) AddToGaugeRewardsFromGauge(ctx sdk.Context, groupGaugeId uint64, coins sdk.Coins, internalGaugeId uint64) error {
	// check if the internalGaugeId is present in groupGauge
	groupGaugeObj, err := k.GetGroupGaugeById(ctx, groupGaugeId)
	if err != nil {
		return err
	}
	found := false
	for _, val := range groupGaugeObj.InternalIds {
		if val == internalGaugeId {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("InternalGaugeId %d is not present in groupGauge: %v", internalGaugeId, groupGaugeObj.InternalIds)
	}

	groupGauge, err := k.GetGaugeByID(ctx, groupGaugeId)
	if err != nil {
		return err
	}

	internalGauge, err := k.GetGaugeByID(ctx, internalGaugeId)
	if err != nil {
		return err
	}

	if internalGauge.IsFinishedGauge(ctx.BlockTime()) {
		return errors.New("gauge is already completed")
	}

	// check if there is sufficient funds in groupGauge to make the transfer
	remainingCoins := groupGauge.Coins.Sub(groupGauge.DistributedCoins)
	if remainingCoins.IsAllLT(coins) {
		return fmt.Errorf("group gauge id: %d doesnot have enough tokens to transfer", groupGaugeId)
	}

	internalGauge.Coins = internalGauge.Coins.Add(coins...)
	err = k.setGauge(ctx, internalGauge)
	if err != nil {
		return err
	}

	return nil
}

// AddToGaugeRewards adds coins to gauge.
func (k Keeper) AddToGaugeRewards(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, gaugeID uint64) error {
	gauge, err := k.GetGaugeByID(ctx, gaugeID)
	if err != nil {
		return err
	}
	if gauge.IsFinishedGauge(ctx.BlockTime()) {
		return errors.New("gauge is already completed")
	}

	// Fixed gas consumption adding reward to gauges based on the number of coins to add
	ctx.GasMeter().ConsumeGas(uint64(types.BaseGasFeeForAddRewardToGauge*(len(coins)+len(gauge.Coins))), "scaling gas cost for adding to gauge rewards")

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

// GetGaugeByID returns gauge from gauge ID.
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

// GetGaugeFromIDs returns multiple gauges from a gaugeIDs array.
func (k Keeper) GetGaugeFromIDs(ctx sdk.Context, gaugeIDs []uint64) ([]types.Gauge, error) {
	gauges := []types.Gauge{}
	for _, gaugeID := range gaugeIDs {
		gauge, err := k.GetGaugeByID(ctx, gaugeID)
		if err != nil {
			return []types.Gauge{}, err
		}
		gauges = append(gauges, *gauge)
	}
	return gauges, nil
}

// GetGauges returns upcoming, active, and finished gauges.
func (k Keeper) GetGauges(ctx sdk.Context) []types.Gauge {
	return k.getGaugesFromIterator(ctx, k.GaugesIterator(ctx))
}

// GetNotFinishedGauges returns both upcoming and active gauges.
func (k Keeper) GetNotFinishedGauges(ctx sdk.Context) []types.Gauge {
	return append(k.GetActiveGauges(ctx), k.GetUpcomingGauges(ctx)...)
}

// GetActiveGauges returns active gauges.
func (k Keeper) GetActiveGauges(ctx sdk.Context) []types.Gauge {
	return k.getGaugesFromIterator(ctx, k.ActiveGaugesIterator(ctx))
}

// GetUpcomingGauges returns upcoming gauges.
func (k Keeper) GetUpcomingGauges(ctx sdk.Context) []types.Gauge {
	return k.getGaugesFromIterator(ctx, k.UpcomingGaugesIterator(ctx))
}

// GetFinishedGauges returns finished gauges.
func (k Keeper) GetFinishedGauges(ctx sdk.Context) []types.Gauge {
	return k.getGaugesFromIterator(ctx, k.FinishedGaugesIterator(ctx))
}

// GetRewardsEst returns rewards estimation at a future specific time (by epoch)
// If locks are nil, it returns the rewards between now and the end epoch associated with address.
// If locks are not nil, it returns all the rewards for the given locks between now and end epoch.
func (k Keeper) GetRewardsEst(ctx sdk.Context, addr sdk.AccAddress, locks []lockuptypes.PeriodLock, endEpoch int64) sdk.Coins {
	// if locks are nil, populate with all locks associated with the address
	if len(locks) == 0 {
		locks = k.lk.GetAccountPeriodLocks(ctx, addr)
	}
	// get all gauges that reward to these locks
	// first get all the denominations being locked up
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
		// each gauge only rewards locks to one denom, so no duplicates
		for _, id := range gaugeIDs {
			gauge, err := k.GetGaugeByID(ctx, id)
			// shouldn't happen
			if err != nil {
				return sdk.Coins{}
			}
			gauges = append(gauges, *gauge)
		}
	}

	// estimate rewards
	estimatedRewards := sdk.Coins{}
	epochInfo := k.GetEpochInfo(ctx)

	// no need to change storage while doing estimation as we use cached context
	cacheCtx, _ := ctx.CacheContext()
	for _, gauge := range gauges {
		distrBeginEpoch := epochInfo.CurrentEpoch
		blockTime := ctx.BlockTime()
		if gauge.StartTime.After(blockTime) {
			distrBeginEpoch = epochInfo.CurrentEpoch + 1 + int64(gauge.StartTime.Sub(blockTime)/epochInfo.Duration)
		}

		for epoch := distrBeginEpoch; epoch <= endEpoch; epoch++ {
			newGauge, distrCoins, isBuggedGauge, err := k.FilteredLocksDistributionEst(cacheCtx, gauge, locks)
			if err != nil {
				continue
			}
			if isBuggedGauge {
				ctx.Logger().Error("Reward estimation does not include gauge " + strconv.Itoa(int(gauge.Id)) + " due to accumulation store bug")
			}
			estimatedRewards = estimatedRewards.Add(distrCoins...)
			gauge = newGauge
		}
	}

	return estimatedRewards
}

// GetEpochInfo returns EpochInfo struct given context.
func (k Keeper) GetEpochInfo(ctx sdk.Context) epochtypes.EpochInfo {
	params := k.GetParams(ctx)
	return k.ek.GetEpochInfo(ctx, params.DistrEpochIdentifier)
}

// chargeFeeIfSufficientFeeDenomBalance charges fee in the base denom on the address if the address has
// balance that is less than fee + amount of the coin from gaugeCoins that is of base denom.
// gaugeCoins might not have a coin of tx base denom. In that case, fee is only compared to balance.
// The fee is sent to the community pool.
// Returns nil on success, error otherwise.
func (k Keeper) chargeFeeIfSufficientFeeDenomBalance(ctx sdk.Context, address sdk.AccAddress, fee osmomath.Int, gaugeCoins sdk.Coins) (err error) {
	feeDenom, err := k.tk.GetBaseDenom(ctx)
	if err != nil {
		return err
	}

	totalCost := gaugeCoins.AmountOf(feeDenom).Add(fee)
	accountBalance := k.bk.GetBalance(ctx, address, feeDenom).Amount

	if accountBalance.LT(totalCost) {
		return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, "account's balance of %s (%s) is less than the total cost of the message (%s)", feeDenom, accountBalance, totalCost)
	}

	if err := k.ck.FundCommunityPool(ctx, sdk.NewCoins(sdk.NewCoin(feeDenom, fee)), address); err != nil {
		return err
	}
	return nil
}
