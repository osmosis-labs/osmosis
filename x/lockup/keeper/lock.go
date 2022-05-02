package keeper

import (
	"fmt"
	"sort"
	"time"

	"github.com/gogo/protobuf/proto"

	"github.com/osmosis-labs/osmosis/v7/store"
	"github.com/osmosis-labs/osmosis/v7/x/lockup/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: Reorganize functions in this file

// WithdrawAllMaturedLocks withdraws every lock thats in the process of unlocking, and has finished unlocking by
// the current block time.
func (k Keeper) WithdrawAllMaturedLocks(ctx sdk.Context) {
	k.unlockFromIterator(ctx, k.LockIteratorBeforeTime(ctx, ctx.BlockTime()))
}

func (k Keeper) getCoinsFromLocks(locks []types.PeriodLock) sdk.Coins {
	coins := sdk.Coins{}
	for _, lock := range locks {
		coins = coins.Add(lock.Coins...)
	}
	return coins
}

func (k Keeper) accumulationStore(ctx sdk.Context, denom string) store.Tree {
	return store.NewTree(prefix.NewStore(ctx.KVStore(k.storeKey), accumulationStorePrefix(denom)), 10)
}

// GetModuleBalance Returns full balance of the module.
func (k Keeper) GetModuleBalance(ctx sdk.Context) sdk.Coins {
	// TODO: should add invariant test for module balance and lock items
	acc := k.ak.GetModuleAccount(ctx, types.ModuleName)
	return k.bk.GetAllBalances(ctx, acc.GetAddress())
}

// GetModuleLockedCoins Returns locked balance of the module.
func (k Keeper) GetModuleLockedCoins(ctx sdk.Context) sdk.Coins {
	// all not unlocking + not finished unlocking
	notUnlockingCoins := k.getCoinsFromIterator(ctx, k.LockIterator(ctx, false))
	unlockingCoins := k.getCoinsFromIterator(ctx, k.LockIteratorAfterTime(ctx, ctx.BlockTime()))
	return notUnlockingCoins.Add(unlockingCoins...)
}

// GetPeriodLocksByDuration returns the total amount of query.Denom tokens locked for longer than
// query.Duration.
func (k Keeper) GetPeriodLocksAccumulation(ctx sdk.Context, query types.QueryCondition) sdk.Int {
	beginKey := accumulationKey(query.Duration)
	return k.accumulationStore(ctx, query.Denom).SubsetAccumulation(beginKey, nil)
}

// BeginUnlockAllNotUnlockings begins unlock for all not unlocking coins.
func (k Keeper) BeginUnlockAllNotUnlockings(ctx sdk.Context, account sdk.AccAddress) ([]types.PeriodLock, error) {
	locks, err := k.beginUnlockFromIterator(ctx, k.AccountLockIterator(ctx, false, account))
	return locks, err
}

func (k Keeper) addTokensToLock(ctx sdk.Context, lock *types.PeriodLock, coins sdk.Coins) error {
	lock.Coins = lock.Coins.Add(coins...)

	err := k.setLock(ctx, *lock)
	if err != nil {
		return err
	}

	// modifications to accumulation store
	for _, coin := range coins {
		k.accumulationStore(ctx, coin.Denom).Increase(accumulationKey(lock.Duration), coin.Amount)
	}

	// modifications to accumulation store by synthlocks
	// CONTRACT: lock will have synthetic lock only if it has a single coin
	lockedCoin, err := lock.SingleCoin()
	if err == nil {
		for _, synthlock := range k.GetAllSyntheticLockupsByLockup(ctx, lock.ID) {
			k.accumulationStore(ctx, synthlock.SynthDenom).Increase(accumulationKey(synthlock.Duration), coins.AmountOf(lockedCoin.Denom))
		}
	}

	k.hooks.AfterAddTokensToLock(ctx, lock.OwnerAddress(), lock.GetID(), coins)

	return nil
}

// removeTokensFromLock is called by lockup slash function - called by superfluid module only.
func (k Keeper) removeTokensFromLock(ctx sdk.Context, lock *types.PeriodLock, coins sdk.Coins) error {
	// TODO: Handle 100% slash eventually, not needed for osmosis codebase atm.
	lock.Coins = lock.Coins.Sub(coins)

	err := k.setLock(ctx, *lock)
	if err != nil {
		return err
	}

	// modifications to accumulation store
	for _, coin := range coins {
		k.accumulationStore(ctx, coin.Denom).Decrease(accumulationKey(lock.Duration), coin.Amount)
	}

	// increase synthetic lockup's accumulation store
	synthLocks := k.GetAllSyntheticLockupsByLockup(ctx, lock.ID)

	// Note: since synthetic lockup deletion is using native lockup's coins to reduce accumulation store
	// all the synthetic lockups' accumulation should be decreased
	for _, synthlock := range synthLocks {
		k.accumulationStore(ctx, synthlock.SynthDenom).Decrease(accumulationKey(synthlock.Duration), coins[0].Amount)
	}
	return nil
}

// AddTokensToLock locks more tokens into a lockup
// This also saves the lock to the store.
func (k Keeper) AddTokensToLockByID(ctx sdk.Context, lockID uint64, coins sdk.Coins) (*types.PeriodLock, error) {
	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return nil, err
	}
	if err := k.bk.SendCoinsFromAccountToModule(ctx, lock.OwnerAddress(), types.ModuleName, coins); err != nil {
		return nil, err
	}

	err = k.addTokensToLock(ctx, lock, coins)
	if err != nil {
		return nil, err
	}

	if k.hooks == nil {
		return lock, nil
	}

	k.hooks.OnTokenLocked(ctx, lock.OwnerAddress(), lock.ID, coins, lock.Duration, lock.EndTime)
	return lock, nil
}

// SlashTokensFromLockByID send slashed tokens to community pool - called by superfluid module only.
func (k Keeper) SlashTokensFromLockByID(ctx sdk.Context, lockID uint64, coins sdk.Coins) (*types.PeriodLock, error) {
	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return nil, err
	}

	modAddr := k.ak.GetModuleAddress(types.ModuleName)
	err = k.dk.FundCommunityPool(ctx, coins, modAddr)
	if err != nil {
		return nil, err
	}

	err = k.removeTokensFromLock(ctx, lock, coins)
	if err != nil {
		return nil, err
	}

	if k.hooks == nil {
		return lock, nil
	}

	k.hooks.OnTokenSlashed(ctx, lock.ID, coins)
	return lock, nil
}

// LockTokens lock tokens from an account for specified duration.
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

func (k Keeper) ClearAccumulationStores(ctx sdk.Context) {
	k.clearKeysByPrefix(ctx, types.KeyPrefixLockAccumulation)
}

// ResetAllLocks takes a set of locks, and initializes state to be storing
// them all correctly. This utilizes batch optimizations to improve efficiency,
// as this becomes a bottleneck at chain initialization & upgrades.
func (k Keeper) ResetAllLocks(ctx sdk.Context, locks []types.PeriodLock) error {
	// index by coin.Denom, them duration -> amt
	// We accumulate the accumulation store entries separately,
	// to avoid hitting the myriad of slowdowns in the SDK iterator creation process.
	// We then save these once to the accumulation store at the end.
	accumulationStoreEntries := make(map[string]map[time.Duration]sdk.Int)
	denoms := []string{}
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
		for _, coin := range lock.Coins {
			// update or create the new map from duration -> Int for this denom.
			var curDurationMap map[time.Duration]sdk.Int
			if durationMap, ok := accumulationStoreEntries[coin.Denom]; ok {
				curDurationMap = durationMap
				// update or create new amount in the duration map
				newAmt := coin.Amount
				if curAmt, ok := durationMap[lock.Duration]; ok {
					newAmt = newAmt.Add(curAmt)
				}
				curDurationMap[lock.Duration] = newAmt
			} else {
				denoms = append(denoms, coin.Denom)
				curDurationMap = map[time.Duration]sdk.Int{lock.Duration: coin.Amount}
			}
			accumulationStoreEntries[coin.Denom] = curDurationMap
		}
	}

	// deterministically iterate over durationMap cache.
	sort.Strings(denoms)
	for _, denom := range denoms {
		curDurationMap := accumulationStoreEntries[denom]
		durations := make([]time.Duration, 0, len(curDurationMap))
		for duration := range curDurationMap {
			durations = append(durations, duration)
		}
		sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
		// now that we have a sorted list of durations for this denom,
		// add them all to accumulation store
		msg := fmt.Sprintf("Setting accumulation entries for locks for %s, there are %d distinct durations",
			denom, len(durations))
		ctx.Logger().Info(msg)
		for _, d := range durations {
			amt := curDurationMap[d]
			k.accumulationStore(ctx, denom).Increase(accumulationKey(d), amt)
		}
	}

	return nil
}

func (k Keeper) ResetAllSyntheticLocks(ctx sdk.Context, syntheticLocks []types.SyntheticLock) error {
	// index by coin.Denom, them duration -> amt
	// We accumulate the accumulation store entries separately,
	// to avoid hitting the myriad of slowdowns in the SDK iterator creation process.
	// We then save these once to the accumulation store at the end.
	accumulationStoreEntries := make(map[string]map[time.Duration]sdk.Int)
	denoms := []string{}
	for i, synthLock := range syntheticLocks {
		if i%25000 == 0 {
			msg := fmt.Sprintf("Reset %d synthetic lock refs", i)
			ctx.Logger().Info(msg)
		}

		// Add to the accumlation store cache
		lock, err := k.GetLockByID(ctx, synthLock.UnderlyingLockId)
		if err != nil {
			return err
		}

		err = k.setSyntheticLockAndResetRefs(ctx, *lock, synthLock)
		if err != nil {
			return err
		}

		coin, err := lock.SingleCoin()
		if err != nil {
			return err
		}

		var curDurationMap map[time.Duration]sdk.Int
		if durationMap, ok := accumulationStoreEntries[synthLock.SynthDenom]; ok {
			curDurationMap = durationMap
			newAmt := coin.Amount
			if curAmt, ok := durationMap[synthLock.Duration]; ok {
				newAmt = newAmt.Add(curAmt)
			}
			curDurationMap[synthLock.Duration] = newAmt
		} else {
			denoms = append(denoms, synthLock.SynthDenom)
			curDurationMap = map[time.Duration]sdk.Int{synthLock.Duration: coin.Amount}
		}
		accumulationStoreEntries[synthLock.SynthDenom] = curDurationMap
	}

	// deterministically iterate over durationMap cache.
	sort.Strings(denoms)
	for _, denom := range denoms {
		curDurationMap := accumulationStoreEntries[denom]
		durations := make([]time.Duration, 0, len(curDurationMap))
		for duration := range curDurationMap {
			durations = append(durations, duration)
		}
		sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
		// now that we have a sorted list of durations for this denom,
		// add them all to accumulation store
		msg := fmt.Sprintf("Setting accumulation entries for locks for %s, there are %d distinct durations",
			denom, len(durations))
		ctx.Logger().Info(msg)
		for _, d := range durations {
			amt := curDurationMap[d]
			k.accumulationStore(ctx, denom).Increase(accumulationKey(d), amt)
		}
	}

	return nil
}

func (k Keeper) setSyntheticLockAndResetRefs(ctx sdk.Context, lock types.PeriodLock, synthLock types.SyntheticLock) error {
	err := k.setSyntheticLockupObject(ctx, &synthLock)
	if err != nil {
		return err
	}

	// store synth lock refs
	return k.addSyntheticLockRefs(ctx, lock, synthLock)
}

// setLockAndResetLockRefs sets the lock, and resets all of its lock references
// This puts the lock into a 'clean' state, aside from the AccumulationStore.
func (k Keeper) setLockAndResetLockRefs(ctx sdk.Context, lock types.PeriodLock) error {
	err := k.setLock(ctx, lock)
	if err != nil {
		return err
	}

	return k.addLockRefs(ctx, lock)
}

// setLock is a utility to store lock object into the store.
func (k Keeper) setLock(ctx sdk.Context, lock types.PeriodLock) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(&lock)
	if err != nil {
		return err
	}
	store.Set(lockStoreKey(lock.ID), bz)
	return nil
}

// deleteLock removes the lock object from the state.
func (k Keeper) deleteLock(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(lockStoreKey(id))
}

// Lock is a utility to lock coins into module account.
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
	err = k.addLockRefs(ctx, lock)
	if err != nil {
		return err
	}

	// add to accumulation store
	for _, coin := range lock.Coins {
		k.accumulationStore(ctx, coin.Denom).Increase(accumulationKey(lock.Duration), coin.Amount)
	}

	k.hooks.OnTokenLocked(ctx, owner, lock.ID, lock.Coins, lock.Duration, lock.EndTime)
	return nil
}

// splitLock splits a lock with the given amount, and stores split new lock to the state.
func (k Keeper) splitLock(ctx sdk.Context, lock types.PeriodLock, coins sdk.Coins) (types.PeriodLock, error) {
	if lock.IsUnlocking() {
		return types.PeriodLock{}, fmt.Errorf("cannot split unlocking lock")
	}
	lock.Coins = lock.Coins.Sub(coins)
	err := k.setLock(ctx, lock)
	if err != nil {
		return types.PeriodLock{}, err
	}

	splitLockID := k.GetLastLockID(ctx) + 1
	k.SetLastLockID(ctx, splitLockID)

	splitLock := types.NewPeriodLock(splitLockID, lock.OwnerAddress(), lock.Duration, lock.EndTime, coins)
	err = k.setLock(ctx, splitLock)
	return splitLock, err
}

// BeginUnlock is a utility to start unlocking coins from NotUnlocking queue.
func (k Keeper) BeginUnlock(ctx sdk.Context, lockID uint64, coins sdk.Coins) error {
	// prohibit BeginUnlock if synthetic locks are referring to this
	// TODO: In the future, make synthetic locks only get partial restrictions on the main lock.
	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}
	if k.HasAnySyntheticLockups(ctx, lock.ID) {
		return fmt.Errorf("cannot BeginUnlocking a lock with synthetic lockup")
	}

	return k.beginForceUnlock(ctx, *lock, coins)
}

func (k Keeper) BeginForceUnlock(ctx sdk.Context, lockID uint64, coins sdk.Coins) error {
	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}
	return k.beginForceUnlock(ctx, *lock, coins)
}

func (k Keeper) beginForceUnlock(ctx sdk.Context, lock types.PeriodLock, coins sdk.Coins) error {
	// sanity check
	if !coins.IsAllLTE(lock.Coins) {
		return fmt.Errorf("requested amount to unlock exceeds locked tokens")
	}

	// If the amount were unlocking is empty, or the entire coins amount, unlock the entire lock.
	// Otherwise, split the lock into two locks, and fully unlock the newly created lock.
	// (By virtue, the newly created lock we split into should have the unlock amount)
	if len(coins) != 0 && !coins.IsEqual(lock.Coins) {
		splitLock, err := k.splitLock(ctx, lock, coins)
		if err != nil {
			return err
		}
		lock = splitLock
	}

	// remove lock refs from not unlocking queue if exists
	err := k.deleteLockRefs(ctx, types.KeyPrefixNotUnlocking, lock)
	if err != nil {
		return err
	}

	// store lock with end time set
	lock.EndTime = ctx.BlockTime().Add(lock.Duration)
	err = k.setLock(ctx, lock)
	if err != nil {
		return err
	}

	// add lock refs into unlocking queue
	err = k.addLockRefs(ctx, lock)
	if err != nil {
		return err
	}

	if k.hooks != nil {
		k.hooks.OnStartUnlock(ctx, lock.OwnerAddress(), lock.ID, lock.Coins, lock.Duration, lock.EndTime)
	}

	return nil
}

// Unlock is a utility to unlock coins from module account.
func (k Keeper) Unlock(ctx sdk.Context, lockID uint64) error {
	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}

	// validation for current time and unlock time
	curTime := ctx.BlockTime()
	if !lock.IsUnlocking() {
		return fmt.Errorf("lock hasn't started unlocking yet")
	}
	if curTime.Before(lock.EndTime) {
		return fmt.Errorf("lock is not unlockable yet: %s >= %s", curTime.String(), lock.EndTime.String())
	}

	return k.unlockInternalLogic(ctx, *lock)
}

// ForceUnlock ignores unlock duration and immediately unlock and refund.
// CONTRACT: should be used only at the chain upgrade script
// TODO: Revisit for Superfluid Staking.
func (k Keeper) ForceUnlock(ctx sdk.Context, lock types.PeriodLock) error {
	if !lock.IsUnlocking() {
		err := k.BeginUnlock(ctx, lock.ID, nil)
		if err != nil {
			return err
		}
	}
	return k.unlockInternalLogic(ctx, lock)
}

func (k Keeper) unlockInternalLogic(ctx sdk.Context, lock types.PeriodLock) error {
	owner, err := sdk.AccAddressFromBech32(lock.Owner)
	if err != nil {
		return err
	}

	// send coins back to owner
	if err := k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, owner, lock.Coins); err != nil {
		return err
	}

	k.deleteLock(ctx, lock.ID)

	// delete lock refs from the unlocking queue
	err = k.deleteLockRefs(ctx, types.KeyPrefixUnlocking, lock)
	if err != nil {
		return err
	}

	// remove from accumulation store
	for _, coin := range lock.Coins {
		k.accumulationStore(ctx, coin.Denom).Decrease(accumulationKey(lock.Duration), coin.Amount)
	}

	k.hooks.OnTokenUnlocked(ctx, owner, lock.ID, lock.Coins, lock.Duration, lock.EndTime)
	return nil
}

func (k Keeper) ExtendLockup(ctx sdk.Context, lock types.PeriodLock, newDuration time.Duration) error {
	if lock.IsUnlocking() {
		return fmt.Errorf("cannot edit unlocking lockup for lock %d", lock.ID)
	}

	// check synthetic lockup exists
	if k.HasAnySyntheticLockups(ctx, lock.ID) {
		return fmt.Errorf("cannot edit lockup with synthetic lock %d", lock.ID)
	}

	oldLock := lock

	if newDuration != 0 {
		if newDuration <= lock.Duration {
			return fmt.Errorf("new duration should be greater than the original")
		}

		// update accumulation store
		for _, coin := range lock.Coins {
			k.accumulationStore(ctx, coin.Denom).Decrease(accumulationKey(lock.Duration), coin.Amount)
			k.accumulationStore(ctx, coin.Denom).Increase(accumulationKey(newDuration), coin.Amount)
		}

		lock.Duration = newDuration
	}

	// update lockup
	err := k.deleteLockRefs(ctx, unlockingPrefix(oldLock.IsUnlocking()), oldLock)
	if err != nil {
		return err
	}

	err = k.addLockRefs(ctx, lock)
	if err != nil {
		return err
	}

	err = k.setLock(ctx, lock)
	if err != nil {
		return err
	}

	k.hooks.OnTokenUnlocked(ctx, lock.OwnerAddress(), lock.ID, lock.Coins, oldLock.Duration, lock.EndTime)
	k.hooks.OnTokenLocked(ctx, lock.OwnerAddress(), lock.ID, lock.Coins, lock.Duration, lock.EndTime)

	return nil
}
