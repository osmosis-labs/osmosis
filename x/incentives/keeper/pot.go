package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/c-osmosis/osmosis/x/incentives/types"
	lockuptypes "github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	db "github.com/tendermint/tm-db"
)

// Iterate over everything in a pots iterator, until it reaches the end. Return all pots iterated over.
func (k Keeper) getPotsFromIterator(ctx sdk.Context, iterator db.Iterator) []types.Pot {
	pots := []types.Pot{}
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		potIDs := []uint64{}
		err := json.Unmarshal(iterator.Value(), &potIDs)
		if err != nil {
			panic(err)
		}
		for _, potID := range potIDs {
			pot, err := k.GetPotByID(ctx, potID)
			if err != nil {
				panic(err)
			}
			pots = append(pots, *pot)
		}
	}
	return pots
}

// Compute the total amount of coins in all the pots
func (k Keeper) getCoinsFromPots(pots []types.Pot) sdk.Coins {
	coins := sdk.Coins{}
	for _, pot := range pots {
		coins = coins.Add(pot.Coins...)
	}
	return coins
}

func (k Keeper) getDistributedCoinsFromPots(pots []types.Pot) sdk.Coins {
	coins := sdk.Coins{}
	for _, pot := range pots {
		coins = coins.Add(pot.DistributedCoins...)
	}
	return coins
}

func (k Keeper) getToDistributeCoinsFromPots(pots []types.Pot) sdk.Coins {
	// TODO: Consider optimizing this in the future to only require one iteration over all pots.
	coins := k.getCoinsFromPots(pots)
	distributed := k.getDistributedCoinsFromPots(pots)
	return coins.Sub(distributed)
}

func (k Keeper) getCoinsFromIterator(ctx sdk.Context, iterator db.Iterator) sdk.Coins {
	return k.getCoinsFromPots(k.getPotsFromIterator(ctx, iterator))
}

func (k Keeper) getToDistributeCoinsFromIterator(ctx sdk.Context, iterator db.Iterator) sdk.Coins {
	return k.getToDistributeCoinsFromPots(k.getPotsFromIterator(ctx, iterator))
}

func (k Keeper) getDistributedCoinsFromIterator(ctx sdk.Context, iterator db.Iterator) sdk.Coins {
	return k.getDistributedCoinsFromPots(k.getPotsFromIterator(ctx, iterator))
}

// setPot modify pot into different one
func (k Keeper) setPot(ctx sdk.Context, pot *types.Pot) error {
	store := ctx.KVStore(k.storeKey)
	store.Set(potStoreKey(pot.Id), k.cdc.MustMarshalJSON(pot))
	return nil
}

// CreatePot create a pot and send coins to the pot
func (k Keeper) CreatePot(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, distrTo lockuptypes.QueryCondition, startTime time.Time, numEpochsPaidOver uint64) (uint64, error) {
	pot := types.Pot{
		Id:                k.getLastPotID(ctx) + 1,
		DistributeTo:      distrTo,
		Coins:             coins,
		StartTime:         startTime,
		NumEpochsPaidOver: numEpochsPaidOver,
	}

	if err := k.bk.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, pot.Coins); err != nil {
		return 0, err
	}

	k.setPot(ctx, &pot)
	k.setLastPotID(ctx, pot.Id)

	if err := k.addPotRefByKey(ctx, combineKeys(types.KeyPrefixUpcomingPots, getTimeKey(pot.StartTime)), pot.Id); err != nil {
		return 0, err
	}
	k.hooks.AfterCreatePot(ctx, pot.Id)
	return pot.Id, nil
}

// AddToPot add coins to pot
func (k Keeper) AddToPotRewards(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, potID uint64) error {
	pot, err := k.GetPotByID(ctx, potID)
	if err != nil {
		return err
	}
	if err := k.bk.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, coins); err != nil {
		return err
	}

	pot.Coins = pot.Coins.Add(coins...)
	k.setPot(ctx, pot)
	k.hooks.AfterAddToPot(ctx, pot.Id)
	return nil
}

// BeginDistribution is a utility to begin distribution for a specific pot
func (k Keeper) BeginDistribution(ctx sdk.Context, pot types.Pot) error {
	// validation for current time and distribution start time
	curTime := ctx.BlockTime()
	if curTime.Before(pot.StartTime) {
		return fmt.Errorf("pot is not able to start distribution yet: %s >= %s", curTime.String(), pot.StartTime.String())
	}

	timeKey := getTimeKey(pot.StartTime)
	k.deletePotRefByKey(ctx, combineKeys(types.KeyPrefixUpcomingPots, timeKey), pot.Id)
	k.addPotRefByKey(ctx, combineKeys(types.KeyPrefixActivePots, timeKey), pot.Id)
	k.hooks.AfterFinishDistribution(ctx, pot.Id)
	return nil
}

// FinishDistribution is a utility to finish distribution for a specific pot
func (k Keeper) FinishDistribution(ctx sdk.Context, pot types.Pot) error {
	timeKey := getTimeKey(pot.StartTime)
	k.deletePotRefByKey(ctx, combineKeys(types.KeyPrefixActivePots, timeKey), pot.Id)
	k.addPotRefByKey(ctx, combineKeys(types.KeyPrefixFinishedPots, timeKey), pot.Id)
	k.hooks.AfterFinishDistribution(ctx, pot.Id)
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

// FilteredLocksDistributionEst estimate distribution amount coins from pot for fitting conditions
func (k Keeper) FilteredLocksDistributionEst(ctx sdk.Context, pot types.Pot, filteredLocks []lockuptypes.PeriodLock) (types.Pot, sdk.Coins, error) {
	filteredLockIDs := make(map[uint64]bool)
	for _, lock := range filteredLocks {
		filteredLockIDs[lock.ID] = true
	}

	totalDistrCoins := sdk.NewCoins()
	filteredDistrCoins := sdk.NewCoins()
	locks := k.GetLocksToDistribution(ctx, pot.DistributeTo)
	lockSum := lockuptypes.SumLocksByDenom(locks, pot.DistributeTo.Denom)

	if lockSum.IsZero() {
		return types.Pot{}, nil, nil
	}

	remainCoins := pot.Coins.Sub(pot.DistributedCoins)
	remainEpochs := pot.NumEpochsPaidOver - pot.FilledEpochs
	for _, lock := range locks {
		distrCoins := sdk.Coins{}
		for _, coin := range remainCoins {
			// distribution amount = pot_size * denom_lock_amount / (total_denom_lock_amount * remain_epochs)
			denomLockAmt := lock.Coins.AmountOf(pot.DistributeTo.Denom)
			amt := coin.Amount.Mul(denomLockAmt).Quo(lockSum.Mul(sdk.NewInt(int64(remainEpochs))))
			if amt.IsPositive() {
				distrCoins = distrCoins.Add(sdk.NewCoin(coin.Denom, amt))
			}
		}
		distrCoins = distrCoins.Sort()
		if !distrCoins.Empty() && (len(filteredLocks) == 0 || filteredLockIDs[lock.ID]) {
			filteredDistrCoins = filteredDistrCoins.Add(distrCoins...)
		}
		totalDistrCoins = totalDistrCoins.Add(distrCoins...)
	}

	// increase filled epochs after distribution
	pot.FilledEpochs += 1
	pot.DistributedCoins = pot.DistributedCoins.Add(totalDistrCoins...)

	return pot, filteredDistrCoins, nil
}

// Distribute coins from pot according to its conditions
func (k Keeper) Distribute(ctx sdk.Context, pot types.Pot) (sdk.Coins, error) {
	totalDistrCoins := sdk.NewCoins()
	locks := k.GetLocksToDistribution(ctx, pot.DistributeTo)
	lockSum := lockuptypes.SumLocksByDenom(locks, pot.DistributeTo.Denom)

	if lockSum.IsZero() {
		return nil, nil
	}

	remainCoins := pot.Coins.Sub(pot.DistributedCoins)
	remainEpochs := pot.NumEpochsPaidOver - pot.FilledEpochs
	for _, lock := range locks {
		distrCoins := sdk.Coins{}
		for _, coin := range remainCoins {
			// distribution amount = pot_size * denom_lock_amount / (total_denom_lock_amount * remain_epochs)
			denomLockAmt := lock.Coins.AmountOf(pot.DistributeTo.Denom)
			amt := coin.Amount.Mul(denomLockAmt).Quo(lockSum.Mul(sdk.NewInt(int64(remainEpochs))))
			if amt.IsPositive() {
				distrCoins = distrCoins.Add(sdk.NewCoin(coin.Denom, amt))
			}
		}
		distrCoins = distrCoins.Sort()
		if distrCoins.Empty() {
			continue
		}
		if err := k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lock.Owner, distrCoins); err != nil {
			return nil, err
		}
		totalDistrCoins = totalDistrCoins.Add(distrCoins...)
	}

	// increase filled epochs after distribution
	pot.FilledEpochs += 1
	pot.DistributedCoins = pot.DistributedCoins.Add(totalDistrCoins...)
	k.setPot(ctx, &pot)

	k.hooks.AfterDistribute(ctx, pot.Id)
	return totalDistrCoins, nil
}

// GetModuleToDistributeCoins returns sum of to distribute coins for all of the module
func (k Keeper) GetModuleToDistributeCoins(ctx sdk.Context) sdk.Coins {
	activePotsDistr := k.getToDistributeCoinsFromIterator(ctx, k.ActivePotsIterator(ctx))
	upcomingPotsDistr := k.getToDistributeCoinsFromIterator(ctx, k.UpcomingPotsIteratorAfterTime(ctx, ctx.BlockTime()))
	return activePotsDistr.Add(upcomingPotsDistr...)
}

// GetModuleDistributedCoins returns sum of distributed coins so far
func (k Keeper) GetModuleDistributedCoins(ctx sdk.Context) sdk.Coins {
	activePotsDistr := k.getDistributedCoinsFromIterator(ctx, k.ActivePotsIterator(ctx))
	finishedPotsDistr := k.getDistributedCoinsFromIterator(ctx, k.FinishedPotsIterator(ctx))
	return activePotsDistr.Add(finishedPotsDistr...)
}

// GetPotByID Returns pot from pot ID
func (k Keeper) GetPotByID(ctx sdk.Context, potID uint64) (*types.Pot, error) {
	pot := types.Pot{}
	store := ctx.KVStore(k.storeKey)
	potKey := potStoreKey(potID)
	if !store.Has(potKey) {
		return nil, fmt.Errorf("pot with ID %d does not exist", potID)
	}
	bz := store.Get(potKey)
	k.cdc.MustUnmarshalJSON(bz, &pot)
	return &pot, nil
}

// GetPotFromIDs returns pots from pot ids reference
func (k Keeper) GetPotFromIDs(ctx sdk.Context, refValue []byte) ([]types.Pot, error) {
	pots := []types.Pot{}
	potIDs := []uint64{}
	err := json.Unmarshal(refValue, &potIDs)
	if err != nil {
		return pots, err
	}
	for _, potID := range potIDs {
		pot, err := k.GetPotByID(ctx, potID)
		if err != nil {
			panic(err)
		}
		pots = append(pots, *pot)
	}
	return pots, nil
}

// GetPots returns pots both upcoming and active
func (k Keeper) GetPots(ctx sdk.Context) []types.Pot {
	return append(k.GetActivePots(ctx), k.GetUpcomingPots(ctx)...)
}

// GetActivePots returns active pots
func (k Keeper) GetActivePots(ctx sdk.Context) []types.Pot {
	return k.getPotsFromIterator(ctx, k.ActivePotsIterator(ctx))
}

// GetUpcomingPots returns scheduled pots
func (k Keeper) GetUpcomingPots(ctx sdk.Context) []types.Pot {
	return k.getPotsFromIterator(ctx, k.UpcomingPotsIterator(ctx))
}

// GetFinishedPots returns finished pots
func (k Keeper) GetFinishedPots(ctx sdk.Context) []types.Pot {
	return k.getPotsFromIterator(ctx, k.FinishedPotsIterator(ctx))
}

// GetRewardsEst returns rewards estimation at a future specific time
func (k Keeper) GetRewardsEst(ctx sdk.Context, addr sdk.AccAddress, locks []lockuptypes.PeriodLock, pots []types.Pot, endEpoch int64) sdk.Coins {
	// initialize pots to active and upcomings if not set
	if len(pots) == 0 {
		pots = k.GetPots(ctx)
	}

	// estimate rewards
	estimatedRewards := sdk.Coins{}
	params := k.GetParams(ctx)
	currentEpoch, _ := k.GetCurrentEpochInfo(ctx)

	// no need to change storage while doing estimation and we use cached context
	cacheCtx, _ := ctx.CacheContext()
	for _, pot := range pots {
		distrBeginEpoch := currentEpoch
		blockTime := ctx.BlockTime()
		if pot.StartTime.After(blockTime) {
			avgBlockTime := time.Second * 5
			epochDuration := params.BlocksPerEpoch * int64(avgBlockTime)
			distrBeginEpoch = currentEpoch + 1 + int64(pot.StartTime.Sub(blockTime))/epochDuration
		}

		for epoch := distrBeginEpoch; epoch <= endEpoch; epoch++ {
			newPot, distrCoins, err := k.FilteredLocksDistributionEst(cacheCtx, pot, locks)
			if err != nil {
				continue
			}
			estimatedRewards = estimatedRewards.Add(distrCoins...)
			pot = newPot
		}
	}

	return estimatedRewards
}
