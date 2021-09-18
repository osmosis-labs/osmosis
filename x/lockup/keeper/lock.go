package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/osmosis-labs/osmosis/store"
	"github.com/osmosis-labs/osmosis/x/lockup/types"
	db "github.com/tendermint/tm-db"
)

func (k Keeper) getLocksFromIterator(ctx sdk.Context, iterator db.Iterator) []types.PeriodLock {
	locks := []types.PeriodLock{}
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		lockID := sdk.BigEndianToUint64(iterator.Value())
		lock, err := k.GetLockByID(ctx, lockID)
		if err != nil {
			panic(err)
		}
		locks = append(locks, *lock)
	}
	return locks
}

func (k Keeper) beginUnlockFromIterator(ctx sdk.Context, iterator db.Iterator) ([]types.PeriodLock, sdk.Coins, error) {
	coins := sdk.Coins{}
	locks := k.getLocksFromIterator(ctx, iterator)
	for _, lock := range locks {
		err := k.BeginUnlock(ctx, lock)
		if err != nil {
			return locks, coins, err
		}
		// sum up all coins begin unlocking
		coins = coins.Add(lock.Coins...)
	}
	return locks, coins, nil
}

// WithdrawAllMaturedLocks withdraws every lock thats in the process of unlocking, and has finished unlocking by
// the current block time.
func (k Keeper) WithdrawAllMaturedLocks(ctx sdk.Context) {
	k.unlockFromIterator(ctx, k.LockIteratorBeforeTime(ctx, true, ctx.BlockTime()))
}

func (k Keeper) addLockRefs(ctx sdk.Context, lockRefPrefix []byte, lock types.PeriodLock) error {
	refKeys, err := lockRefKeys(lock)
	if err != nil {
		return err
	}
	for _, refKey := range refKeys {
		if err := k.addLockRefByKey(ctx, combineKeys(lockRefPrefix, refKey), lock.ID); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) deleteLockRefs(ctx sdk.Context, lockRefPrefix []byte, lock types.PeriodLock) error {
	refKeys, err := lockRefKeys(lock)
	if err != nil {
		return err
	}
	for _, refKey := range refKeys {
		if err := k.deleteLockRefByKey(ctx, combineKeys(lockRefPrefix, refKey), lock.ID); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) unlockFromIterator(ctx sdk.Context, iterator db.Iterator) ([]types.PeriodLock, sdk.Coins) {
	coins := sdk.Coins{}
	locks := k.getLocksFromIterator(ctx, iterator)
	for _, lock := range locks {
		err := k.Unlock(ctx, lock)
		if err != nil {
			panic(err)
		}
		// sum up all coins unlocked
		coins = coins.Add(lock.Coins...)
	}
	return locks, coins
}

func (k Keeper) getCoinsFromLocks(locks []types.PeriodLock) sdk.Coins {
	coins := sdk.Coins{}
	for _, lock := range locks {
		coins = coins.Add(lock.Coins...)
	}
	return coins
}

func (k Keeper) getCoinsFromIterator(ctx sdk.Context, iterator db.Iterator) sdk.Coins {
	return k.getCoinsFromLocks(k.getLocksFromIterator(ctx, iterator))
}

func (k Keeper) accumulationStore(ctx sdk.Context, denom string) store.Tree {
	return store.NewTree(prefix.NewStore(ctx.KVStore(k.storeKey), accumulationStorePrefix(denom)), 10)
}

// GetModuleBalance Returns full balance of the module
func (k Keeper) GetModuleBalance(ctx sdk.Context) sdk.Coins {
	// TODO: should add invariant test for module balance and lock items
	acc := k.ak.GetModuleAccount(ctx, types.ModuleName)
	return k.bk.GetAllBalances(ctx, acc.GetAddress())
}

// GetModuleLockedCoins Returns locked balance of the module
func (k Keeper) GetModuleLockedCoins(ctx sdk.Context) sdk.Coins {
	// all not unlocking + not finished unlocking
	notUnlockingCoins := k.getCoinsFromIterator(ctx, k.LockIterator(ctx, false))
	unlockingCoins := k.getCoinsFromIterator(ctx, k.LockIteratorAfterTime(ctx, true, ctx.BlockTime()))
	return notUnlockingCoins.Add(unlockingCoins...)
}

// GetAccountUnlockableCoins Returns whole unlockable coins which are not withdrawn yet
func (k Keeper) GetAccountUnlockableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return k.getCoinsFromIterator(ctx, k.AccountLockIteratorBeforeTime(ctx, true, addr, ctx.BlockTime()))
}

// GetAccountUnlockingCoins Returns whole unlocking coins
func (k Keeper) GetAccountUnlockingCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return k.getCoinsFromIterator(ctx, k.AccountLockIteratorAfterTime(ctx, true, addr, ctx.BlockTime()))
}

// GetAccountLockedCoins Returns a locked coins that can't be withdrawn
func (k Keeper) GetAccountLockedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	// all account unlocking + not finished unlocking
	notUnlockingCoins := k.getCoinsFromIterator(ctx, k.AccountLockIterator(ctx, false, addr))
	unlockingCoins := k.getCoinsFromIterator(ctx, k.AccountLockIteratorAfterTime(ctx, true, addr, ctx.BlockTime()))
	return notUnlockingCoins.Add(unlockingCoins...)
}

// GetAccountLockedPastTime Returns the total locks of an account whose unlock time is beyond timestamp
func (k Keeper) GetAccountLockedPastTime(ctx sdk.Context, addr sdk.AccAddress, timestamp time.Time) []types.PeriodLock {
	// unlockings finish after specific time + not started locks that will finish after the time even though it start now
	unlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorAfterTime(ctx, true, addr, timestamp))
	duration := time.Duration(0)
	if timestamp.After(ctx.BlockTime()) {
		duration = timestamp.Sub(ctx.BlockTime())
	}
	notUnlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDuration(ctx, false, addr, duration))
	return combineLocks(notUnlockings, unlockings)
}

// GetAccountLockedPastTimeNotUnlockingOnly Returns the total locks of an account whose unlock time is beyond timestamp
func (k Keeper) GetAccountLockedPastTimeNotUnlockingOnly(ctx sdk.Context, addr sdk.AccAddress, timestamp time.Time) []types.PeriodLock {
	duration := time.Duration(0)
	if timestamp.After(ctx.BlockTime()) {
		duration = timestamp.Sub(ctx.BlockTime())
	}
	return k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDuration(ctx, false, addr, duration))
}

// GetAccountUnlockedBeforeTime Returns the total unlocks of an account whose unlock time is before timestamp
func (k Keeper) GetAccountUnlockedBeforeTime(ctx sdk.Context, addr sdk.AccAddress, timestamp time.Time) []types.PeriodLock {
	// unlockings finish before specific time + not started locks that can finish before the time if start now
	unlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorBeforeTime(ctx, true, addr, timestamp))
	if timestamp.Before(ctx.BlockTime()) {
		return unlockings
	}
	duration := timestamp.Sub(ctx.BlockTime())
	notUnlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorShorterThanDuration(ctx, false, addr, duration))
	return combineLocks(notUnlockings, unlockings)
}

// GetAccountLockedPastTimeDenom is equal to GetAccountLockedPastTime but denom specific
func (k Keeper) GetAccountLockedPastTimeDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, timestamp time.Time) []types.PeriodLock {
	// unlockings finish after specific time + not started locks that will finish after the time even though it start now
	unlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorAfterTimeDenom(ctx, true, addr, denom, timestamp))
	duration := time.Duration(0)
	if timestamp.After(ctx.BlockTime()) {
		duration = timestamp.Sub(ctx.BlockTime())
	}
	notUnlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDurationDenom(ctx, false, addr, denom, duration))
	return combineLocks(notUnlockings, unlockings)
}

// GetAccountLockedLongerDuration Returns account locked with duration longer than specified
func (k Keeper) GetAccountLockedLongerDuration(ctx sdk.Context, addr sdk.AccAddress, duration time.Duration) []types.PeriodLock {
	// it does not matter started unlocking or not for duration query
	unlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDuration(ctx, true, addr, duration))
	notUnlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDuration(ctx, false, addr, duration))
	return combineLocks(notUnlockings, unlockings)
}

// GetAccountLockedLongerDurationNotUnlockingOnly Returns account locked with duration longer than specified
func (k Keeper) GetAccountLockedLongerDurationNotUnlockingOnly(ctx sdk.Context, addr sdk.AccAddress, duration time.Duration) []types.PeriodLock {
	return k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDuration(ctx, false, addr, duration))
}

// GetAccountLockedLongerDurationDenom Returns account locked with duration longer than specified with specific denom
func (k Keeper) GetAccountLockedLongerDurationDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, duration time.Duration) []types.PeriodLock {
	// it does not matter started unlocking or not for duration query
	unlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDurationDenom(ctx, true, addr, denom, duration))
	notUnlockings := k.getLocksFromIterator(ctx, k.AccountLockIteratorLongerDurationDenom(ctx, false, addr, denom, duration))
	return combineLocks(notUnlockings, unlockings)
}

// GetLocksPastTimeDenom Returns the locks whose unlock time is beyond timestamp
func (k Keeper) GetLocksPastTimeDenom(ctx sdk.Context, denom string, timestamp time.Time) []types.PeriodLock {
	// returns both unlocking started and not started assuming it started unlocking current time
	unlockings := k.getLocksFromIterator(ctx, k.LockIteratorAfterTimeDenom(ctx, true, denom, timestamp))
	duration := time.Duration(0)
	if timestamp.After(ctx.BlockTime()) {
		duration = timestamp.Sub(ctx.BlockTime())
	}
	notUnlockings := k.getLocksFromIterator(ctx, k.LockIteratorLongerThanDurationDenom(ctx, false, denom, duration))
	return combineLocks(notUnlockings, unlockings)
}

// GetLockedDenom Returns the total amount of denom that are locked
func (k Keeper) GetLockedDenom(ctx sdk.Context, denom string, duration time.Duration) sdk.Int {
	totalAmtLocked := k.GetPeriodLocksAccumulation(ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         denom,
		Duration:      duration,
	})
	return totalAmtLocked
}

// GetLocksLongerThanDurationDenom Returns the locks whose unlock duration is longer than duration
func (k Keeper) GetLocksLongerThanDurationDenom(ctx sdk.Context, denom string, duration time.Duration) []types.PeriodLock {
	// returns both unlocking started and not started
	unlockings := k.getLocksFromIterator(ctx, k.LockIteratorLongerThanDurationDenom(ctx, true, denom, duration))
	notUnlockings := k.getLocksFromIterator(ctx, k.LockIteratorLongerThanDurationDenom(ctx, false, denom, duration))
	return combineLocks(notUnlockings, unlockings)
}

// GetLockByID Returns lock from lockID
func (k Keeper) GetLockByID(ctx sdk.Context, lockID uint64) (*types.PeriodLock, error) {
	lock := types.PeriodLock{}
	store := ctx.KVStore(k.storeKey)
	lockKey := lockStoreKey(lockID)
	if !store.Has(lockKey) {
		return nil, fmt.Errorf("lock with ID %d does not exist", lockID)
	}
	bz := store.Get(lockKey)
	err := proto.Unmarshal(bz, &lock)
	return &lock, err
}

// GetPeriodLocks Returns the period locks on pool
func (k Keeper) GetPeriodLocks(ctx sdk.Context) ([]types.PeriodLock, error) {
	unlockings := k.getLocksFromIterator(ctx, k.LockIterator(ctx, true))
	notUnlockings := k.getLocksFromIterator(ctx, k.LockIterator(ctx, false))
	return combineLocks(notUnlockings, unlockings), nil
}

// GetAccountPeriodLocks Returns the period locks associated to an account
func (k Keeper) GetAccountPeriodLocks(ctx sdk.Context, addr sdk.AccAddress) []types.PeriodLock {
	unlockings := k.getLocksFromIterator(ctx, k.AccountLockIterator(ctx, true, addr))
	notUnlockings := k.getLocksFromIterator(ctx, k.AccountLockIterator(ctx, false, addr))
	return combineLocks(notUnlockings, unlockings)
}

// GetPeriodLocksByDuration returns the total amount of query.Denom tokens locked for longer than
// query.Duration
func (k Keeper) GetPeriodLocksAccumulation(ctx sdk.Context, query types.QueryCondition) sdk.Int {
	beginKey := accumulationKey(query.Duration)
	return k.accumulationStore(ctx, query.Denom).SubsetAccumulation(beginKey, nil)
}

// BeginUnlockAllNotUnlockings begins unlock for all not unlocking coins
func (k Keeper) BeginUnlockAllNotUnlockings(ctx sdk.Context, account sdk.AccAddress) ([]types.PeriodLock, sdk.Coins, error) {
	locks, coins, err := k.beginUnlockFromIterator(ctx, k.AccountLockIterator(ctx, false, account))
	return locks, coins, err
}

// BeginUnlockPeriodLockByID begin unlock by period lock ID
func (k Keeper) BeginUnlockPeriodLockByID(ctx sdk.Context, LockID uint64) (*types.PeriodLock, error) {
	lock, err := k.GetLockByID(ctx, LockID)
	if err != nil {
		return lock, err
	}
	err = k.BeginUnlock(ctx, *lock)
	return lock, err
}

// UnlockPeriodLockByID unlock by period lock ID
func (k Keeper) UnlockPeriodLockByID(ctx sdk.Context, LockID uint64) (*types.PeriodLock, error) {
	lock, err := k.GetLockByID(ctx, LockID)
	if err != nil {
		return lock, err
	}
	err = k.Unlock(ctx, *lock)
	return lock, err
}

func (k Keeper) addTokensToLock(ctx sdk.Context, lock *types.PeriodLock, coins sdk.Coins) error {
	lock.Coins = lock.Coins.Add(coins...)

	err := k.setLock(ctx, *lock)
	if err != nil {
		return err
	}

	// modifications to accumulation store
	// for _, coin := range coins {
	// 	k.accumulationStore(ctx, coin.Denom).Increase(accumulationKey(lock.Duration), coin.Amount)
	// }

	return nil
}

// AddTokensToLock locks more tokens into a lockup
// This also saves the lock to the store.
func (k Keeper) AddTokensToLockByID(ctx sdk.Context, owner sdk.AccAddress, lockID uint64, coins sdk.Coins) (*types.PeriodLock, error) {
	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return nil, err
	}
	if lock.Owner != owner.String() {
		return nil, types.ErrNotLockOwner
	}
	if err := k.bk.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, coins); err != nil {
		return nil, err
	}

	err = k.addTokensToLock(ctx, lock, coins)
	if err != nil {
		return nil, err
	}

	if k.hooks == nil {
		return lock, nil
	}

	k.hooks.OnTokenLocked(ctx, owner, lock.ID, coins, lock.Duration, lock.EndTime)
	return lock, nil
}

// LockTokens lock tokens from an account for specified duration
func (k Keeper) LockTokens(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, duration time.Duration) (types.PeriodLock, error) {
	ID := k.GetLastLockID(ctx) + 1
	// unlock time is set at the beginning of unlocking time
	lock := types.NewPeriodLock(ID, owner, duration, time.Time{}, coins)
	err := k.Lock(ctx, lock)
	if err != nil {
		return lock, err
	}
	k.SetLastLockID(ctx, lock.ID)
	return lock, nil
}

func (k Keeper) clearKeysByPrefix(ctx sdk.Context, prefix []byte) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
}

func (k Keeper) ClearAllLockRefKeys(ctx sdk.Context) {
	k.clearKeysByPrefix(ctx, types.KeyPrefixNotUnlocking)
	k.clearKeysByPrefix(ctx, types.KeyPrefixUnlocking)
}

func (k Keeper) ClearAccumulationStores(ctx sdk.Context) {
	k.clearKeysByPrefix(ctx, types.KeyPrefixLockAccumulation)
}

// ResetAllLocks takes a set of locks, and initializes state to be storing
// them all correctly. This utilizes batch optimizations to improve efficiency,
// as this becomes a bottleneck at chain initialization & upgrades.
// UPDATE: accumulation store is disabled in the v4 upgrade.
func (k Keeper) ResetAllLocks(ctx sdk.Context, locks []types.PeriodLock) error {
	// index by coin.Denom, them duration -> amt
	// We accumulate the accumulation store entries separately,
	// to avoid hitting the myriad of slowdowns in the SDK iterator creation process.
	// We then save these once to the accumulation store at the end.
	// accumulationStoreEntries := make(map[string]map[time.Duration]sdk.Int)
	// denoms := []string{}
	for i, lock := range locks {
		if i%25000 == 0 {
			msg := fmt.Sprintf("Reset %d lock refs, cur lock ID %d", i, lock.ID)
			ctx.Logger().Info(msg)
		}
		err := k.setLockAndResetLockRefs(ctx, lock)
		if err != nil {
			return err
		}

		// Add to the accumlation store cache
		// for _, coin := range lock.Coins {
		// 	// update or create the new map from duration -> Int for this denom.
		// 	var curDurationMap map[time.Duration]sdk.Int
		// 	if durationMap, ok := accumulationStoreEntries[coin.Denom]; ok {
		// 		curDurationMap = durationMap
		// 		// update or create new amount in the duration map
		// 		newAmt := coin.Amount
		// 		if curAmt, ok := durationMap[lock.Duration]; ok {
		// 			newAmt = newAmt.Add(curAmt)
		// 		}
		// 		curDurationMap[lock.Duration] = newAmt
		// 	} else {
		// 		denoms = append(denoms, coin.Denom)
		// 		curDurationMap = map[time.Duration]sdk.Int{lock.Duration: coin.Amount}
		// 	}
		// 	accumulationStoreEntries[coin.Denom] = curDurationMap
		// }
	}

	// deterministically iterate over durationMap cache.
	// sort.Strings(denoms)
	// for _, denom := range denoms {
	// 	curDurationMap := accumulationStoreEntries[denom]
	// 	durations := make([]time.Duration, 0, len(curDurationMap))
	// 	for duration, _ := range curDurationMap {
	// 		durations = append(durations, duration)
	// 	}
	// 	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	// 	// now that we have a sorted list of durations for this denom,
	// 	// add them all to accumulation store
	// 	msg := fmt.Sprintf("Setting accumulation entries for locks for %s, there are %d distinct durations",
	// 		denom, len(durations))
	// 	ctx.Logger().Info(msg)
	// 	for _, d := range durations {
	// 		amt := curDurationMap[d]
	// 		k.accumulationStore(ctx, denom).Increase(accumulationKey(d), amt)
	// 	}
	// }

	return nil
}

// setLockAndResetLockRefs sets the lock, and resets all of its lock references
// This puts the lock into a 'clean' state, aside from the AccumulationStore.
func (k Keeper) setLockAndResetLockRefs(ctx sdk.Context, lock types.PeriodLock) error {
	err := k.setLock(ctx, lock)
	if err != nil {
		return err
	}

	// store refs by the status of unlock
	if lock.IsUnlocking() {
		return k.addLockRefs(ctx, types.KeyPrefixUnlocking, lock)
	}

	return k.addLockRefs(ctx, types.KeyPrefixNotUnlocking, lock)
}

// setLock is a utility to store lock object into the store
func (k Keeper) setLock(ctx sdk.Context, lock types.PeriodLock) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(&lock)
	if err != nil {
		return err
	}
	store.Set(lockStoreKey(lock.ID), bz)
	return nil
}

// Lock is a utility to lock coins into module account
func (k Keeper) Lock(ctx sdk.Context, lock types.PeriodLock) error {
	owner, err := sdk.AccAddressFromBech32(lock.Owner)
	if err != nil {
		return err
	}
	if err := k.bk.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, lock.Coins); err != nil {
		return err
	}

	// store lock object into the store
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(&lock)
	if err != nil {
		return err
	}
	store.Set(lockStoreKey(lock.ID), bz)

	// add lock refs into not unlocking queue
	err = k.addLockRefs(ctx, types.KeyPrefixNotUnlocking, lock)
	if err != nil {
		return err
	}

	// add to accumulation store
	// disabled in v4
	// for _, coin := range lock.Coins {
	// 	k.accumulationStore(ctx, coin.Denom).Increase(accumulationKey(lock.Duration), coin.Amount)
	// }

	k.hooks.OnTokenLocked(ctx, owner, lock.ID, lock.Coins, lock.Duration, lock.EndTime)
	return nil
}

// BeginUnlock is a utility to start unlocking coins from NotUnlocking queue
func (k Keeper) BeginUnlock(ctx sdk.Context, lock types.PeriodLock) error {
	// remove lock refs from not unlocking queue
	err := k.deleteLockRefs(ctx, types.KeyPrefixNotUnlocking, lock)
	if err != nil {
		return err
	}

	// store lock with end time set
	lock.EndTime = ctx.BlockTime().Add(lock.Duration)
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(&lock)
	if err != nil {
		return err
	}
	store.Set(lockStoreKey(lock.ID), bz)

	// add lock refs into unlocking queue
	err = k.addLockRefs(ctx, types.KeyPrefixUnlocking, lock)
	if err != nil {
		return err
	}

	return nil
}

// Unlock is a utility to unlock coins from module account
func (k Keeper) Unlock(ctx sdk.Context, lock types.PeriodLock) error {
	// validation for current time and unlock time
	curTime := ctx.BlockTime()
	if !lock.IsUnlocking() {
		return fmt.Errorf("lock hasn't started unlocking yet")
	}
	if curTime.Before(lock.EndTime) {
		return fmt.Errorf("lock is not unlockable yet: %s >= %s", curTime.String(), lock.EndTime.String())
	}

	owner, err := sdk.AccAddressFromBech32(lock.Owner)
	if err != nil {
		return err
	}

	// send coins back to owner
	if err := k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, owner, lock.Coins); err != nil {
		return err
	}

	// remove lock from store object
	store := ctx.KVStore(k.storeKey)
	store.Delete(lockStoreKey(lock.ID))

	// delete lock refs from the unlocking queue
	err = k.deleteLockRefs(ctx, types.KeyPrefixUnlocking, lock)
	if err != nil {
		return err
	}

	// remove from accumulation store
	// for _, coin := range lock.Coins {
	// 	k.accumulationStore(ctx, coin.Denom).Decrease(accumulationKey(lock.Duration), coin.Amount)
	// }

	k.hooks.OnTokenUnlocked(ctx, owner, lock.ID, lock.Coins, lock.Duration, lock.EndTime)
	return nil
}
