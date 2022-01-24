package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
)

// GetCurrentReward gets total rewards to be distributed in the next epoch per denom + lock
func (k Keeper) GetCurrentReward(ctx sdk.Context, denom string, lockDuration time.Duration) (types.CurrentReward, error) {
	currentReward := types.CurrentReward{}
	currentReward.TotalShares.Denom = denom
	store := ctx.KVStore(k.storeKey)
	rewardKey := combineKeys(types.KeyCurrentReward, []byte(denom+"/"+lockDuration.String()))

	bz := store.Get(rewardKey)
	if bz == nil {
		currentReward.TotalShares = sdk.NewCoin(denom, sdk.NewInt(0))
		currentReward.LastProcessedEpoch = -1
		return currentReward, nil
	}

	err := proto.Unmarshal(bz, &currentReward)
	if err != nil {
		return currentReward, err
	}
	return currentReward, nil
}

// GetAllCurrentReward gets all current reward
func (k Keeper) GetAllCurrentReward(ctx sdk.Context) []types.CurrentReward {
	iterator := k.iterator(ctx, types.KeyCurrentReward)
	currentRewards := []types.CurrentReward{}
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		currentReward := types.CurrentReward{}
		err := proto.Unmarshal(iterator.Value(), &currentReward)
		if err != nil {
			panic(err)
		}
		currentRewards = append(currentRewards, currentReward)
	}
	return currentRewards
}

func (k Keeper) SetCurrentReward(ctx sdk.Context, currentReward types.CurrentReward, denom string, lockDuration time.Duration) error {
	store := ctx.KVStore(k.storeKey)
	rewardKey := combineKeys(types.KeyCurrentReward, []byte(denom+"/"+lockDuration.String()))

	currentReward.Denom = denom
	currentReward.LockDuration = lockDuration
	bz, err := proto.Marshal(&currentReward)
	if err != nil {
		return err
	}

	store.Set(rewardKey, bz)

	return nil
}

func (k Keeper) GetHistoricalReward(ctx sdk.Context, denom string, lockDuration time.Duration, epochNumber int64) (types.HistoricalReward, error) {
	historicalReward := types.HistoricalReward{}
	store := ctx.KVStore(k.storeKey)
	rewardKey := combineKeys(types.KeyHistoricalReward, []byte(denom+"/"+lockDuration.String()), sdk.Uint64ToBigEndian(uint64(epochNumber)))

	if epochNumber == -1 {
		historicalReward.CumulativeRewardRatio = sdk.DecCoins{}
		return historicalReward, nil
	}

	bz := store.Get(rewardKey)
	if bz == nil {
		return historicalReward, fmt.Errorf("historical rewards is not present = %d", epochNumber)
	}

	err := proto.Unmarshal(bz, &historicalReward)
	if err != nil {
		return historicalReward, err
	}
	return historicalReward, nil
}

func (k Keeper) SetHistoricalReward(ctx sdk.Context, cumulativeRewardRatio sdk.DecCoins, denom string, lockDuration time.Duration, epochNumber int64) error {
	store := ctx.KVStore(k.storeKey)
	historicalRewardKey := combineKeys(types.KeyHistoricalReward, []byte(denom+"/"+lockDuration.String()), sdk.Uint64ToBigEndian(uint64(epochNumber)))
	historicalReward := types.HistoricalReward{
		CumulativeRewardRatio: cumulativeRewardRatio,
		Epoch:                 epochNumber,
	}

	bz, err := proto.Marshal(&historicalReward)
	if err != nil {
		return err
	}
	store.Set(historicalRewardKey, bz)

	return nil
}

func (k Keeper) SetPeriodLockReward(ctx sdk.Context, periodLockReward types.PeriodLockReward) error {
	store := ctx.KVStore(k.storeKey)
	rewardKey := combineKeys(types.KeyPeriodLockReward, sdk.Uint64ToBigEndian(periodLockReward.LockId))

	bz, err := proto.Marshal(&periodLockReward)
	if err != nil {
		return err
	}

	store.Set(rewardKey, bz)

	return nil
}

func (k Keeper) deletePeriodLockReward(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	rewardKey := combineKeys(types.KeyPeriodLockReward, sdk.Uint64ToBigEndian(id))
	store.Delete(rewardKey)
}

func (k Keeper) GetAllPeriodLockReward(ctx sdk.Context) []types.PeriodLockReward {
	iterator := k.iterator(ctx, types.KeyPeriodLockReward)
	periodLockRewards := []types.PeriodLockReward{}
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		periodLockReward := types.PeriodLockReward{}
		err := proto.Unmarshal(iterator.Value(), &periodLockReward)
		if err != nil {
			panic(err)
		}
		periodLockRewards = append(periodLockRewards, periodLockReward)
	}
	return periodLockRewards
}

func (k Keeper) GetPeriodLockReward(ctx sdk.Context, lockID uint64) (types.PeriodLockReward, error) {
	store := ctx.KVStore(k.storeKey)
	rewardKey := combineKeys(types.KeyPeriodLockReward, sdk.Uint64ToBigEndian(lockID))

	bz := store.Get(rewardKey)
	if bz == nil {
		return types.PeriodLockReward{
			LockId: lockID,
		}, nil
	}

	periodLockReward := types.PeriodLockReward{}
	err := proto.Unmarshal(bz, &periodLockReward)
	if err != nil {
		return periodLockReward, err
	}
	return periodLockReward, nil
}

// updateReward adds currentReward to historicalReward and resets currentReward
func (k Keeper) updateReward(ctx sdk.Context, denom string, lockableDuration time.Duration, epochNumber int64) error {
	currentReward, err := k.GetCurrentReward(ctx, denom, lockableDuration)
	if err != nil {
		return err
	}

	if currentReward.LastProcessedEpoch != epochNumber {
		// update currentReward if it's a new epoch that is being stored
		cumulativeRewardRatio, err := k.CalculateCumulativeRewardRatio(ctx, currentReward, denom, lockableDuration, epochNumber)
		if err != nil {
			return err
		}
		err = k.SetHistoricalReward(ctx, cumulativeRewardRatio, denom, lockableDuration, epochNumber)
		if err != nil {
			return err
		}
		currentReward.LastProcessedEpoch = epochNumber
		currentReward.Rewards = sdk.Coins{}
		err = k.SetCurrentReward(ctx, currentReward, denom, lockableDuration)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateRewardForAllLockDuration updates historical and currentReward for all lockDurations
func (k Keeper) UpdateRewardForAllLockDuration(ctx sdk.Context, lockedCoins sdk.Coins, lockDuration time.Duration) error {
	lockableDurations := k.GetLockableDurations(ctx)
	epochInfo := k.GetEpochInfo(ctx)
	for _, lockableDuration := range lockableDurations {
		if lockDuration < lockableDuration {
			continue
		}
		for _, coin := range lockedCoins {
			if err := k.updateReward(ctx, coin.Denom, lockableDuration, epochInfo.CurrentEpoch); err != nil {
				return err
			}
		}
	}
	return nil
}

func (k Keeper) UpdatePeriodLockReward(ctx sdk.Context, lock lockuptypes.PeriodLock, lockReward types.PeriodLockReward) error {
	lockReward, err := k.GetRewardForLock(ctx, lock, lockReward)
	if err != nil {
		return err
	}

	err = k.SetPeriodLockReward(ctx, lockReward)
	if err != nil {
		return err
	}

	return nil
}
