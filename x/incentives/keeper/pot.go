package keeper

import (
	"fmt"
	"math/big"
	"time"

	"github.com/c-osmosis/osmosis/x/incentives/types"
	lockuptypes "github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
func (k Keeper) CreatePot(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, distrTo []*types.DistrCondition, startTime time.Time, numEpochs uint64) (uint64, error) {
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

	if err := k.addPotRefByKey(ctx, getTimeKey(pot.StartTime), potID); err != nil {
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

// Distribute distriute coins from pot for fitting conditions
func (k Keeper) Distribute(ctx sdk.Context, pot types.Pot) error {
	for _, distrCondition := range pot.DistributeTo {
		locks := []lockuptypes.PeriodLock{}
		switch distrCondition.LockType {
		case types.ByDuration:
			locks = k.lk.GetLocksLongerThanDurationDenom(ctx, distrCondition.Denom, distrCondition.Duration)
		case types.ByTime:
			locks = k.lk.GetLocksPastTimeDenom(ctx, distrCondition.Denom, distrCondition.Timestamp)
		default:
		}

		lockSum := sdk.NewInt(0)
		for _, lock := range locks {
			lockSum = lockSum.Add(lock.Coins.AmountOf(distrCondition.Denom))
		}
		if lockSum.IsZero() {
			continue
		}
		for _, lock := range locks {
			distrCoins := pot.Coins
			for i, coin := range distrCoins {
				bi := big.NewInt(0).Div(coin.Amount.BigInt(), big.NewInt(int64(pot.NumEpochs)))
				bi = bi.Mul(bi, lock.Coins.AmountOf(distrCondition.Denom).BigInt())
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
	}

	return nil
}
