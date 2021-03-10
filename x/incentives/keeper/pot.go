package keeper

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/c-osmosis/osmosis/x/incentives/types"
	lockuptypes "github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	db "github.com/tendermint/tm-db"
)

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
	coins := k.getCoinsFromPots(pots)
	distributed := k.getDistributedCoinsFromPots(pots)
	return coins.Sub(distributed)
}

func (k Keeper) getCoinsFromIterator(ctx sdk.Context, iterator db.Iterator) sdk.Coins {
	return k.getCoinsFromPots(k.getPotsFromIterator(ctx, iterator))
}

func (k Keeper) getToDistributeCoinsFromIterator(ctx sdk.Context, iterator db.Iterator) sdk.Coins {
	return k.getCoinsFromPots(k.getPotsFromIterator(ctx, iterator))
}

func (k Keeper) getDistributedCoinsFromIterator(ctx sdk.Context, iterator db.Iterator) sdk.Coins {
	return k.getCoinsFromPots(k.getPotsFromIterator(ctx, iterator))
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

// CreatePot create a pot and send coins to the pot
func (k Keeper) CreatePot(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, distrTo *types.DistrCondition, startTime time.Time, numEpochs uint64) (uint64, error) {
	pot := types.Pot{
		Id:           k.getLastPotID(ctx) + 1,
		DistributeTo: distrTo,
		Coins:        coins,
		StartTime:    startTime,
		NumEpochs:    numEpochs,
	}

	if err := k.bk.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, pot.Coins); err != nil {
		return 0, err
	}

	potID := pot.Id
	store := ctx.KVStore(k.storeKey)
	store.Set(potStoreKey(potID), k.cdc.MustMarshalJSON(&pot))
	k.setLastPotID(ctx, potID)

	if err := k.addPotRefByKey(ctx, combineKeys(types.KeyPrefixUncomingPots, getTimeKey(pot.StartTime)), potID); err != nil {
		return 0, err
	}
	return potID, nil
}

// AddToPot add coins to pot
func (k Keeper) AddToPot(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, potID uint64) error {
	pot, err := k.GetPotByID(ctx, potID)
	if err != nil {
		return err
	}
	if err := k.bk.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, coins); err != nil {
		return err
	}

	pot.Coins = pot.Coins.Add(coins...)

	store := ctx.KVStore(k.storeKey)
	store.Set(potStoreKey(potID), k.cdc.MustMarshalJSON(pot))
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
	k.deletePotRefByKey(ctx, combineKeys(types.KeyPrefixUncomingPots, timeKey), pot.Id)
	k.addPotRefByKey(ctx, combineKeys(types.KeyPrefixActivePots, timeKey), pot.Id)
	return nil
}

// FinishDistribution is a utility to finish distribution for a specific pot
func (k Keeper) FinishDistribution(ctx sdk.Context, pot types.Pot) error {
	timeKey := getTimeKey(pot.StartTime)
	k.deletePotRefByKey(ctx, combineKeys(types.KeyPrefixActivePots, timeKey), pot.Id)
	k.addPotRefByKey(ctx, combineKeys(types.KeyPrefixFinishedPots, timeKey), pot.Id)
	return nil
}

// Distribute distriute coins from pot for fitting conditions
func (k Keeper) Distribute(ctx sdk.Context, pot types.Pot) error {
	locks := []lockuptypes.PeriodLock{}
	switch pot.DistributeTo.LockQueryType {
	case types.ByDuration:
		locks = k.lk.GetLocksLongerThanDurationDenom(ctx, pot.DistributeTo.Denom, pot.DistributeTo.Duration)
	case types.ByTime:
		locks = k.lk.GetLocksPastTimeDenom(ctx, pot.DistributeTo.Denom, pot.DistributeTo.Timestamp)
	default:
	}

	lockSum := sdk.NewInt(0)
	for _, lock := range locks {
		lockSum = lockSum.Add(lock.Coins.AmountOf(pot.DistributeTo.Denom))
	}
	if lockSum.IsZero() {
		return nil
	}
	for _, lock := range locks {
		distrCoins := pot.Coins
		for i, coin := range distrCoins {
			bi := big.NewInt(0).Div(coin.Amount.BigInt(), big.NewInt(int64(pot.NumEpochs)))
			bi = bi.Mul(bi, lock.Coins.AmountOf(pot.DistributeTo.Denom).BigInt())
			bi = bi.Div(bi, lockSum.BigInt())
			distrCoins[i].Amount = sdk.NewIntFromBigInt(bi)
		}
		distrCoins = distrCoins.Sort()
		if distrCoins.Empty() {
			continue
		}
		if err := k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lock.Owner, pot.Coins); err != nil {
			return err
		}
	}

	return nil
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
func (k Keeper) GetRewardsEst(ctx sdk.Context) sdk.Coins {
	// TODO: how params should look like and how to calculate this?
	return nil
}
