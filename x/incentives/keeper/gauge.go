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

	"github.com/osmosis-labs/osmosis/v17/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v17/x/lockup/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v17/x/poolmanager/types"
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

// TODO implement getGetGroupGaugeFromIterator function
// func (k Keeper) getGroupGaugesFromIterator(ctx sdk.Context, iterator db.Iterator) []types.GroupGauge {

// }

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
// Note: this function adds gauge reference to state to identify if this gauge is (upcoming, active, finished) with a certain KV.
// Note: We can probably reuse this function for GroupGauge.
func (k Keeper) CreateGaugeRefKeys(ctx sdk.Context, gauge *types.Gauge, combinedKeys []byte, activeOrUpcomingGauge bool) error {
	if err := k.addGaugeRefByKey(ctx, combinedKeys, gauge.Id); err != nil {
		return err
	}
	// Note: i dont think we need this for GroupGauge since we donot need GroupGauge <> Denoms connection.
	// Note: we use denom here so that we can getAllGaugesByDenom. look into getAllGaugeIDsByDenom for more details.
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

// TODO: refactor CreateGauge to use this as well
// type GaugeParams struct {
// 	IsPerpetual       bool
// 	DistributeTo      lockuptypes.QueryCondition
// 	Coins             sdk.Coins
// 	StartTime         time.Time
// 	NumEpochsPaidOver uint64
// }

// TODO: consider doing a general gauge interface?
// type GroupGauge struct {
//  Id                uint64
//  GaugeParams 	  GaugeParams
//  InternalGaugeIDs  []uint64 // consider just using pool IDs
//  SplittingPolicy   SplittingPolicy
// }

// CreateGroupGauge creates a group gauge (need better name) that allocates rewards dynamically across its internal gauges based on the given splitting policy.
// The only supported splitting policy for now is VolumeSplitting, which allocates rewards based on the volume of each internal gauge.
// Note: we should expect that the internal gauges consist of the gauges that are automatically created for each pool upon pool creation, as even non-perpetual
// external incentives would likely want to route through these.
//
// Observation: in that case, can we just make people pass in a list pool IDs instead of internalGaugeIDs? This would allow us to abstract the notion of internal gauges
// from people who just want to incentivize a list of pools.
//
// Alternatively, we can have a pre-approved list of "Groups" which are just lists of pool IDs that are allowed to be used in group gauges. Then, people can just pass in the IDs for these.
// These can be a module param set by governance.
func (k Keeper) CreateGroupGauge(ctx sdk.Context, internalGaugeIDs []uint64, gaugeParams GaugeParams, splittingPolicy SplittingPolicy) (newGroupGauge GroupGauge, err error) {
	// Require all internal gauges:
	// - to be perpetual
	// - to share the same base asset (let's just require an OSMO pair). Get pool using `GetPoolIdFromGaugeIdStoreKey`
	// - TODO: allow for non OSMO pair using `GetPoolForDenomPair`
	// - TODO: determine if there should be a governance-approved list of pool IDs that are allowed (see observation above as well)
	// - TODO: if not governance-approved groups, add cap on length. This all becomes much easier if we just have governance-approved groups.

	// Validate gauge params (usual validation logic)

	// Validate splitting policy:
	// If != enum 0 (VolumeSplitting, defined in types like e.g. SuperfluidUnbonding), then error

	// TODO: add initial one in upgrade handler
	nextGroupGaugeId := k.GetLastGroupGaugeID(ctx) + 1

	newGroupGauge := GroupGauge{
		Id:               nextGroupGaugeId,
		GaugeParams:      gaugeParams,
		InternalGaugeIDs: internalGaugeIDs,
		SplittingPolicy:  splittingPolicy,
	}

	// TODO: implement this
	err = k.setGroupGauge(ctx, &newGroupGauge)
	if err != nil {
		return 0, err
	}

	// TODO: implement this
	k.SetLastGroupGaugeID(ctx, gauge.Id)

	//Figure out gauge ref key stuff here

	combinedKeys := combineKeys(types.keyPrefixUpcomingGroupGauges, getTimeKey(groupGauge.StartTime))
	activeOrUpcomingGroupGauge := true

	err = k.CreateGaugeRefKeys(ctx, &gauge, combinedKeys, activeOrUpcomingGroupGauge)
	if err != nil {
		return 0, err
	}

	//Scale gas by number of internal gauges (linearly)

	return newGroupGauge, nil
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

// TODO: implement GetUpcomingGroupGauges
func (k Keeper) GetUpcomingGroupGauges(ctx sdk.Context) []types.Gauge {
	// Todo: call getGroupGaugesFromIterator with upcomingGroupGaugeIterator
	return k.getGaugesFromIterator(ctx, k.UpcomingGaugesIterator(ctx))
}

// TODO: implement GetActiveGroupGauges
func (k Keeper) GetActiveGroupGauges(ctx sdk.Context) []types.Gauge {
	// Todo: call getGroupGaugesFromIterator with activeGroupGaugeIterator
	return k.getGaugesFromIterator(ctx, k.ActiveGaugesIterator(ctx))
}

// TODO: implement GetFinishedGroupGauges
func (k Keeper) GetFinishedGroupGauges(ctx sdk.Context) []types.Gauge {
	// Todo: call getGroupGaugesFromIterator with finishedGroupGaugeIterator
	return k.getGaugesFromIterator(ctx, k.FinishedGaugesIterator(ctx))
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
func (k Keeper) chargeFeeIfSufficientFeeDenomBalance(ctx sdk.Context, address sdk.AccAddress, fee sdk.Int, gaugeCoins sdk.Coins) (err error) {
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
