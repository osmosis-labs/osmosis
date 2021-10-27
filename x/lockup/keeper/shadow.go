package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/osmosis-labs/osmosis/x/lockup/types"
)

// Shadow lockup spec
// - Shadow lockup uses same denom as prefix ({origin_denom}/staked_{validator_id})
// - Shadow lockup addition, deletion, state transition to unbonding should be called by external modules
// - Shadow lockup should follow the changes of native lockups
// - Shadow lockup has reference to native lockup ID
// - AccumulationStore should be managed for shadow lockups as another denom

// Scenario
// - Distribution module distribute rewards to shadow lockups using accumulation store I guess
// - If a user begin unlock the lockup, shadow lockup automatically move to unlocking lockup if exist.
// (Staking module or superfluid module should make following actions for this for voting power change etc.)
// - If unlock of the lockup finishes and lockup is deleted, shadow lockup should be deleted together. (Do it via hooks? or do directly?)
//// - Superfluid module create shadow lockup if a user want to use his lockup for superfluid staking
//// - Superfluid module start unbonding of shadow lockup if a user don't want to do superfluid staking
//// - Superfluid module add unbonding shadow lockup if the user redelegate to another validator
//// Shadow lockup could exist more than one per denom, and if suffix is same, only one could exist.
//// - Should be able to get native lockup ID from shadow and from native to shadows

func (k Keeper) setShadowLockupObject(ctx sdk.Context, lockID uint64, shadow string, endTime time.Time) error {
	shadowLock := &types.ShadowLock{
		LockId:  lockID,
		Shadow:  shadow,
		EndTime: endTime,
	}
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(shadowLock)
	if err != nil {
		return err
	}
	store.Set(shadowLockStoreKey(lockID, shadow), bz)
	if !endTime.Equal(time.Time{}) {
		store.Set(shadowLockTimeStoreKey(lockID, shadow, endTime), bz)
	}
	return nil
}

func (k Keeper) deleteShadowLockupObject(ctx sdk.Context, lockID uint64, shadow string) {
	store := ctx.KVStore(k.storeKey)
	shadowLock, _ := k.GetShadowLockup(ctx, lockID, shadow)
	if shadowLock != nil && !shadowLock.EndTime.Equal(time.Time{}) {
		store.Delete(shadowLockTimeStoreKey(lockID, shadow, shadowLock.EndTime))
	}
	store.Delete(shadowLockStoreKey(lockID, shadow))
}

func (k Keeper) GetShadowLockup(ctx sdk.Context, lockID uint64, shadow string) (*types.ShadowLock, error) {
	shadowLock := types.ShadowLock{}
	store := ctx.KVStore(k.storeKey)
	shadowLockKey := shadowLockStoreKey(lockID, shadow)
	if !store.Has(shadowLockKey) {
		return nil, fmt.Errorf("shadow lock with ID %d and shadow %s does not exist", lockID, shadow)
	}
	bz := store.Get(shadowLockKey)
	err := proto.Unmarshal(bz, &shadowLock)
	return &shadowLock, err
}

func (k Keeper) GetAllShadowsByLockup(ctx sdk.Context, lockID uint64) []types.ShadowLock {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, combineKeys(types.KeyPrefixShadowLockup, sdk.Uint64ToBigEndian(lockID)))

	shadowLocks := []types.ShadowLock{}
	for ; iterator.Valid(); iterator.Next() {
		shadowLock := types.ShadowLock{}
		err := proto.Unmarshal(iterator.Value(), &shadowLock)
		if err != nil {
			panic(err)
		}
		shadowLocks = append(shadowLocks, shadowLock)
	}
	return shadowLocks
}

func (k Keeper) GetAllShadows(ctx sdk.Context) []types.ShadowLock {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixShadowLockup)

	shadowLocks := []types.ShadowLock{}
	for ; iterator.Valid(); iterator.Next() {
		shadowLock := types.ShadowLock{}
		err := proto.Unmarshal(iterator.Value(), &shadowLock)
		if err != nil {
			panic(err)
		}
		shadowLocks = append(shadowLocks, shadowLock)
	}
	return shadowLocks
}

// CreateShadowLockup create shadow of lockup with lock id and shadow(denom suffix)
func (k Keeper) CreateShadowLockup(ctx sdk.Context, lockID uint64, shadow string, isUnlocking bool) error {
	// Note: shadow lock up is doing everything same as lockup except coin movement
	// There is no relationship between unbonding and bonding shadow lockup, it's managed separately
	// Accumulation store works without caring about unlocking shadow or not
	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}

	lock.Coins = shadowCoins(lock.Coins, shadow)
	if isUnlocking { // end time is set automatically if it's unlocking lockup
		lock.EndTime = ctx.BlockTime().Add(lock.Duration)
	} else {
		lock.EndTime = time.Time{}
	}

	// set shadow lockup object
	err = k.setShadowLockupObject(ctx, lockID, shadow, lock.EndTime)
	if err != nil {
		return err
	}

	unlockingPrefix := unlockingPrefix(isUnlocking)

	// add lock refs into not unlocking queue
	err = k.addShadowLockRefs(ctx, unlockingPrefix, *lock)
	if err != nil {
		return err
	}

	// add to accumulation store
	for _, coin := range lock.Coins {
		k.accumulationStore(ctx, coin.Denom).Increase(accumulationKey(lock.Duration), coin.Amount)
	}
	return nil
}

// DeleteShadowLockup delete shadow of lockup with lock id and shadow(denom suffix)
func (k Keeper) DeleteShadowLockup(ctx sdk.Context, lockID uint64, shadow string) error {
	shadowLock, err := k.GetShadowLockup(ctx, lockID, shadow)
	if err != nil {
		return err
	}

	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}

	// update lock for shadow lock
	lock.Coins = shadowCoins(lock.Coins, shadow)
	lock.EndTime = shadowLock.EndTime

	k.deleteShadowLockupObject(ctx, lockID, shadow)

	// delete lock refs from the unlocking queue
	err = k.deleteShadowLockRefs(ctx, unlockingPrefix(lock.IsUnlocking()), *lock)
	if err != nil {
		return err
	}

	// remove from accumulation store
	for _, coin := range lock.Coins {
		k.accumulationStore(ctx, coin.Denom).Decrease(accumulationKey(lock.Duration), coin.Amount)
	}
	return nil
}

// DeleteAllShadowByLockup delete all the shadows of lockup by id
func (k Keeper) DeleteAllShadowsByLockup(ctx sdk.Context, lockID uint64) error {
	shadowLocks := k.GetAllShadowsByLockup(ctx, lockID)
	for _, shadowLock := range shadowLocks {
		err := k.DeleteShadowLockup(ctx, lockID, shadowLock.Shadow)
		if err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) DeleteAllMaturedShadowLocks(ctx sdk.Context) {
	iterator := k.iteratorBeforeTime(ctx, combineKeys(types.KeyPrefixShadowLockTimestamp), ctx.BlockTime())
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		shadowLock := types.ShadowLock{}
		err := proto.Unmarshal(iterator.Value(), &shadowLock)
		if err != nil {
			panic(err)
		}
		k.DeleteShadowLockup(ctx, shadowLock.LockId, shadowLock.Shadow)
	}
}
