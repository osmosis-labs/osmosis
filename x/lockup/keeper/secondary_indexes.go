package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/lockup/types"
)

func SecondaryIndexAccumulationStoreKey(denom string, secondaryIndex string) string {
	return string(combineKeys([]byte(denom), []byte(secondaryIndex)))
}

func (k Keeper) AddSecondaryIndex(ctx sdk.Context, lockID uint64, newSecondaryIndex string) error {
	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}

	// enforce that lock can only have one denom
	coin, err := lock.SingleCoin()
	if err != nil {
		return err
	}

	if lock.HasSecondaryIndex(ctx, newSecondaryIndex) {
		return types.ErrSecondaryIndexAlreadyAdded
	}

	lock.SecondaryIndexes = append(lock.SecondaryIndexes, newSecondaryIndex)
	err = k.setLock(ctx, *lock)
	if err != nil {
		return err
	}

	// Add to accumulation store key
	accumulationStoreKey := SecondaryIndexAccumulationStoreKey(coin.Denom, newSecondaryIndex)
	k.accumulationStore(ctx, accumulationStoreKey).Increase(accumulationKey(lock.Duration), coin.Amount)

	return err
}

func (k Keeper) RemoveSecondaryIndex(ctx sdk.Context, lockID uint64, removeSecondaryIndex string) error {
	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}

	for i, existingSecondaryIndexes := range lock.SecondaryIndexes {
		if existingSecondaryIndexes == removeSecondaryIndex {
			lock.SecondaryIndexes = append(lock.SecondaryIndexes[:i], lock.SecondaryIndexes[i+1:]...)
			err = k.setLock(ctx, *lock)

			// remove from accumulation store
			coin, err := lock.SingleCoin()
			if err != nil {
				return err
			}

			accumulationStoreKey := SecondaryIndexAccumulationStoreKey(coin.Denom, removeSecondaryIndex)
			k.accumulationStore(ctx, accumulationStoreKey).Decrease(accumulationKey(lock.Duration), coin.Amount)
			return nil

			return err
		}
	}

	return nil
}

// // syntheticLockRefKeys are different from native lockRefKeys to avoid conflicts
// func secondaryIndexRefKeys(lock types.PeriodLock, secondaryIndex string) ([][]byte, error) {
// 	refKeys := [][]byte{}
// 	timeKey := getTimeKey(lock.EndTime)
// 	durationKey := getDurationKey(lock.Duration)

// 	owner, err := sdk.AccAddressFromBech32(lock.Owner)
// 	if err != nil {
// 		return nil, err
// 	}

// 	refKeys = append(refKeys, combineKeys(types.KeyPrefixLockTimestamp, timeKey))
// 	refKeys = append(refKeys, combineKeys(types.KeyPrefixLockDuration, durationKey))
// 	refKeys = append(refKeys, combineKeys(types.KeyPrefixAccountLockTimestamp, owner, timeKey))
// 	refKeys = append(refKeys, combineKeys(types.KeyPrefixAccountLockDuration, owner, durationKey))

// 	for _, coin := range lock.Coins {
// 		denomBz := []byte(coin.Denom)
// 		refKeys = append(refKeys, combineKeys(types.KeyPrefixDenomLockTimestamp, denomBz, timeKey))
// 		refKeys = append(refKeys, combineKeys(types.KeyPrefixDenomLockDuration, denomBz, durationKey))
// 		refKeys = append(refKeys, combineKeys(types.KeyPrefixAccountDenomLockTimestamp, owner, denomBz, timeKey))
// 		refKeys = append(refKeys, combineKeys(types.KeyPrefixAccountDenomLockDuration, owner, denomBz, durationKey))
// 	}
// 	return refKeys, nil
// }
