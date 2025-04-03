package keeper

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	db "github.com/cosmos/cosmos-db"
	"github.com/cosmos/gogoproto/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	epochtypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

var byGroupQueryCondition = lockuptypes.QueryCondition{LockQueryType: lockuptypes.ByGroup}

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
func (k Keeper) CreateGaugeRefKeys(ctx sdk.Context, gauge *types.Gauge, combinedKeys []byte) error {
	if err := k.addGaugeRefByKey(ctx, combinedKeys, gauge.Id); err != nil {
		return err
	}

	activeOrUpcomingGauge := gauge.IsActiveGauge(ctx.BlockTime()) || gauge.IsUpcomingGauge(ctx.BlockTime())

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

	if gauge.IsUpcomingGauge(curTime) {
		combinedKeys := combineKeys(types.KeyPrefixUpcomingGauges, timeKey)
		return k.CreateGaugeRefKeys(ctx, gauge, combinedKeys)
	} else if gauge.IsActiveGauge(curTime) {
		combinedKeys := combineKeys(types.KeyPrefixActiveGauges, timeKey)
		return k.CreateGaugeRefKeys(ctx, gauge, combinedKeys)
	} else {
		combinedKeys := combineKeys(types.KeyPrefixFinishedGauges, timeKey)
		return k.CreateGaugeRefKeys(ctx, gauge, combinedKeys)
	}
}

// CreateGauge creates a gauge with the given parameters and sends coins to the gauge.
// There can be 3 kinds of gauges for a given set of parameters:
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
// * lockuptypes.Group - a gauge that incentivizes a group of internal pool gauges based on the splitting
// policy created by a group data structure. It is expected to be created via CreateGroup keeper method.
// This gauge is the only gauge type that does not have ref keys (active/upcoming/finished) created and
// associated with it.
// For this gauge, the pool id must be 0. Fails if not.
//
// Returns error if:
// - attempts to create non-perpetual gauge with numEpochsPaidOver of 0
//
// On success, returns the gauge ID.
func (k Keeper) CreateGauge(ctx sdk.Context, isPerpetual bool, owner sdk.AccAddress, coins sdk.Coins, distrTo lockuptypes.QueryCondition, startTime time.Time, numEpochsPaidOver uint64, poolId uint64) (uint64, error) {
	if numEpochsPaidOver == types.PerpetualNumEpochsPaidOver && !isPerpetual {
		return 0, types.ErrZeroNumEpochsPaidOver
	}

	// Check that the coins being sent to the gauge exist as a skip hot route
	// This is used to determine the underlying value of the rewards per user at epoch,
	// since we don't distribute tokens values under a certain threshold.
	// If the denom doesn't exist in the skip hot route, we would never distribute rewards
	// from this gauge.
	if err := k.checkIfDenomsAreDistributable(ctx, coins); err != nil {
		return 0, err
	}

	// If the gauge has no lock, then we currently assume it is a concentrated pool
	// and ensure the gauge "lock" duration is an authorized uptime.
	isNoLockGauge := distrTo.LockQueryType == lockuptypes.NoLock

	// If the gauge has an internal gauge denom, it is an internal gauge
	// and should be run through different validation logic (see below).
	//
	// Two important reminders/assumptions:
	// 1. `NoLock` gauges are required to have empty denoms in `ValidateBasic`, so this
	// check cannot be controlled by user input.
	// 2. The safety of this leans on the special-casing of internal gauge logic during
	// distributions, which should be using the internal incentive duration gov param instead of the duration value.
	isInternalNoLockGauge := isNoLockGauge && distrTo.Denom == types.NoLockInternalGaugeDenom(poolId)
	isExternalNoLockGauge := isNoLockGauge && !isInternalNoLockGauge

	// We check durations if the gauge is a regular duration based gauge or if it is a
	// CL gauge. Note that this excludes time-based gauges and group gauges.
	if isNoLockGauge || distrTo.LockQueryType == lockuptypes.ByDuration {
		// Ensure that this gauge's duration is one of the allowed durations on chain
		// Concentrated pool gauges check against authorized uptimes (if external) or
		// epoch duration (if internal).
		//
		// All other gauges check against the default set of lockable durations.
		var durations []time.Duration
		if isExternalNoLockGauge {
			durations = k.clk.GetParams(ctx).AuthorizedUptimes
		} else if isInternalNoLockGauge {
			// Internal CL gauges use epoch time as their duration. This is a legacy
			// property that does not affect the uptime on created records, which is
			// determined by the gov param for internal incentive uptimes.
			durations = []time.Duration{k.GetEpochInfo(ctx).Duration}
		} else {
			// This branch is applicable to CFMM pool types such as balancer and stableswap.
			durations = k.GetLockableDurations(ctx)
		}

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
	if isNoLockGauge {
		if poolId == 0 {
			return 0, fmt.Errorf("'no lock' type gauges must have a pool id")
		}

		// If not internal gauge denom, then must be set to ""
		// and get overwritten with the external prefix + pool id
		// for internal query purposes.
		distrToDenom := distrTo.Denom
		if !isInternalNoLockGauge {
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
		k.pik.SetPoolGaugeIdNoLock(ctx, poolId, nextGaugeId, distrTo.Duration)
	} else {
		// For all other gauges, pool id must be 0.
		if poolId != 0 {
			return 0, fmt.Errorf("pool id must be 0 for gauges with lock")
		}

		// Group gauges do not distribute to a denom. skip this check for group gauges.
		if distrTo.LockQueryType != lockuptypes.ByGroup {
			// check if denom this gauge pays out to exists on-chain
			// N.B.: The reason we check for osmovaloper is to account for gauges that pay out to
			// superfluid synthetic locks. These locks have the following format:
			// "cl/pool/1/superbonding/osmovaloper1wcfyglfgjs2xtsyqu7pl60d0mpw5g7f4wh7pnm"
			// See x/superfluid module README for details.
			if !k.bk.HasSupply(ctx, distrTo.Denom) && !strings.Contains(distrTo.Denom, "osmovaloper") {
				return 0, fmt.Errorf("denom does not exist: %s", distrTo.Denom)
			}
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

	// Only create ref keys (upcoming/active/finished) if gauge is not a group gauge
	// Group gauges do not follow a similar lifecycle as other gauges.
	if gauge.DistributeTo.LockQueryType != lockuptypes.ByGroup {
		err = k.CreateGaugeRefKeys(ctx, &gauge, combinedKeys)
		if err != nil {
			return 0, err
		}
	}

	// TODO: We comment out AfterCreateGauge hook for two reasons:
	// 1. It is not used anywhere in the codebase.
	// 2. There is a bug where we initHooks after we init gov routes. Therefore,
	// if we attempt to call a method that calls a hook via a gov prop, it will panic.
	// https://github.com/osmosis-labs/osmosis/issues/6580
	// k.hooks.AfterCreateGauge(ctx, gauge.Id)
	return gauge.Id, nil
}

// GetGaugeByID returns gauge from gauge ID.
func (k Keeper) GetGaugeByID(ctx sdk.Context, gaugeID uint64) (*types.Gauge, error) {
	gauge := types.Gauge{}
	store := ctx.KVStore(k.storeKey)
	gaugeKey := gaugeStoreKey(gaugeID)
	if !store.Has(gaugeKey) {
		return nil, types.GaugeNotFoundError{GaugeID: gaugeID}
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

// AddToGaugeRewards adds coins to gauge.
func (k Keeper) AddToGaugeRewards(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, gaugeID uint64) error {
	if err := k.addToGaugeRewards(ctx, coins, gaugeID); err != nil {
		return err
	}

	if err := k.bk.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, coins); err != nil {
		return err
	}
	return nil
}

// addToGaugeRewards adds coins to gauge with the given ID.
//
// Returns error if:
// - fails to retrieve gauge from state
// - gauge is finished.
// - fails to store an updated gauge to state
//
// Notes: does not do token transfers since it is used internally for token transferring value within the
// incentives module or by higher level functions that do transfer.
func (k Keeper) addToGaugeRewards(ctx sdk.Context, coins sdk.Coins, gaugeID uint64) error {
	if err := k.checkIfDenomsAreDistributable(ctx, coins); err != nil {
		return err
	}

	gauge, err := k.GetGaugeByID(ctx, gaugeID)
	if err != nil {
		return err
	}
	if gauge.IsFinishedGauge(ctx.BlockTime()) {
		return types.UnexpectedFinishedGaugeError{GaugeId: gaugeID}
	}

	gauge.Coins = gauge.Coins.Add(coins...)
	err = k.setGauge(ctx, gauge)
	if err != nil {
		return err
	}

	// Fixed gas consumption adding reward to gauges based on the number of coins to add
	ctx.GasMeter().ConsumeGas(uint64(types.BaseGasFeeForAddRewardToGauge*(len(coins)+len(gauge.Coins))), "scaling gas cost for adding to gauge rewards")

	k.hooks.AfterAddToGauge(ctx, gauge.Id)
	return nil
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

// checkIfDenomsAreDistributable checks if the denoms in the provided coins are registered in protorev.
// It iterates over the coins and for each coin, it tries to get the pool for the denom pair with "uosmo".
// If the pool does not exist, it returns an error indicating that the denom does not exist as a protorev hot route.
// If all denoms are valid, it returns nil.
func (k Keeper) checkIfDenomsAreDistributable(ctx sdk.Context, coins sdk.Coins) error {
	for _, coin := range coins {
		if coin.Denom != appparams.BaseCoinUnit {
			_, err := k.prk.GetPoolForDenomPairNoOrder(ctx, coin.Denom, appparams.BaseCoinUnit)
			if err != nil {
				return types.NoRouteForDenomError{Denom: coin.Denom}
			}
		}
	}
	return nil
}
