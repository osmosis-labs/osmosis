package keeper

import (
	"fmt"
	"sort"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"

	"github.com/osmosis-labs/osmosis/v7/store"
	"github.com/osmosis-labs/osmosis/v7/x/lockup/types"
)

// WithdrawAllMaturedLocks withdraws every lock thats in the process of unlocking, and has finished unlocking by
// the current block time.
func (k Keeper) WithdrawAllMaturedLocks(ctx sdk.Context) {
	k.unlockFromIterator(ctx, k.LockIteratorBeforeTime(ctx, ctx.BlockTime()))
}

// GetModuleBalance returns full balance of the module.
func (k Keeper) GetModuleBalance(ctx sdk.Context) sdk.Coins {
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

// BeginUnlockAllNotUnlockings begins unlock for all not unlocking locks of the given account.
func (k Keeper) BeginUnlockAllNotUnlockings(ctx sdk.Context, account sdk.AccAddress) ([]types.PeriodLock, error) {
	locks, err := k.beginUnlockFromIterator(ctx, k.AccountLockIterator(ctx, false, account))
	return locks, err
}

// AddToExistingLock adds the given coin to the existing lock with the same owner and duration.
// Returns an empty array of period lock when a lock with the given condition does not exist.
func (k Keeper) AddToExistingLock(ctx sdk.Context, owner sdk.AccAddress, coin sdk.Coin, duration time.Duration) ([]types.PeriodLock, error) {
	locks := k.GetAccountLockedDurationNotUnlockingOnly(ctx, owner, coin.Denom, duration)
	// if existing lock with same duration and denom exists, just add there
	if len(locks) > 0 {
		lock := locks[0]
		_, err := k.AddTokensToLockByID(ctx, lock.ID, owner, coin)
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
		}
	}
	return locks, nil
}

// AddTokensToLock locks additional tokens into an existing lock with the given ID.
// Tokens locked are sent and kept in the module account.
// This method alters the lock state in store, thus we do a sanity check to ensure
// lock owner matches the given owner.
func (k Keeper) AddTokensToLockByID(ctx sdk.Context, lockID uint64, owner sdk.AccAddress, tokensToAdd sdk.Coin) (*types.PeriodLock, error) {
	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return nil, err
	}

	if lock.GetOwner() != owner.String() {
		return nil, types.ErrNotLockOwner
	}

	lock.Coins = lock.Coins.Add(tokensToAdd)
	err = k.lock(ctx, *lock, sdk.NewCoins(tokensToAdd))
	if err != nil {
		return nil, err
	}

	for _, synthlock := range k.GetAllSyntheticLockupsByLockup(ctx, lock.ID) {
		k.accumulationStore(ctx, synthlock.SynthDenom).Increase(accumulationKey(synthlock.Duration), tokensToAdd.Amount)
	}

	if k.hooks == nil {
		return lock, nil
	}

	k.hooks.AfterAddTokensToLock(ctx, lock.OwnerAddress(), lock.GetID(), sdk.NewCoins(tokensToAdd))

	return lock, nil
}

// CreateLock creates a new lock with the specified duration for the owner.
// Returns an error in the following conditions:
//  - account does not have enough balance
func (k Keeper) CreateLock(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, duration time.Duration) (types.PeriodLock, error) {
	ID := k.GetLastLockID(ctx) + 1
	// unlock time is initially set without a value, gets set as unlock start time + duration
	// when unlocking starts.
	lock := types.NewPeriodLock(ID, owner, duration, time.Time{}, coins)
	err := k.lock(ctx, lock, lock.Coins)
	if err != nil {
		return lock, err
	}

	// add lock refs into not unlocking queue
	err = k.addLockRefs(ctx, lock)
	if err != nil {
		return lock, err
	}

	k.SetLastLockID(ctx, lock.ID)
	return lock, nil
}

// lock is an internal utility to lock coins and set corresponding states.
// This is only called by either of the two possible entry points to lock tokens.
// 1. CreateLock
// 2. AddTokensToLockByID
func (k Keeper) lock(ctx sdk.Context, lock types.PeriodLock, tokensToLock sdk.Coins) error {
	owner, err := sdk.AccAddressFromBech32(lock.Owner)
	if err != nil {
		return err
	}
	if err := k.bk.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, tokensToLock); err != nil {
		return err
	}

	// store lock object into the store
	err = k.setLock(ctx, lock)
	if err != nil {
		return err
	}

	// add to accumulation store
	for _, coin := range tokensToLock {
		k.accumulationStore(ctx, coin.Denom).Increase(accumulationKey(lock.Duration), coin.Amount)
	}

	k.hooks.OnTokenLocked(ctx, owner, lock.ID, lock.Coins, lock.Duration, lock.EndTime)
	return nil
}

// BeginUnlock is a utility to start unlocking coins from NotUnlocking queue.
// Returns an error if the lock has a synthetic lock.
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

	return k.beginUnlock(ctx, *lock, coins)
}

// BeginForceUnlock begins force unlock of the given lock.
// This method should be called by the superfluid module ONLY, as it does not check whether
// the lock has a synthetic lock or not before unlocking.
func (k Keeper) BeginForceUnlock(ctx sdk.Context, lockID uint64, coins sdk.Coins) error {
	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}
	return k.beginUnlock(ctx, *lock, coins)
}

// beginUnlock unlocks specified tokens from the given lock. Existing lock refs
// of not unlocking queue are deleted and new lock refs are then added.
// EndTime of the lock is set within this method.
// Coins provided as the parameter does not require to have all the tokens in the lock,
// as we allow partial unlockings of a lock.
func (k Keeper) beginUnlock(ctx sdk.Context, lock types.PeriodLock, coins sdk.Coins) error {
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

	// remove existing lock refs from not unlocking queue
	err := k.deleteLockRefs(ctx, types.KeyPrefixNotUnlocking, lock)
	if err != nil {
		return err
	}

	// store lock with the end time set to current block time + duration
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

func (k Keeper) BeginForceUnlockWithEndTime(ctx sdk.Context, lockID uint64, endTime time.Time) error {
	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}
	return k.beginForceUnlockWithEndTime(ctx, *lock, endTime)
}

func (k Keeper) beginForceUnlockWithEndTime(ctx sdk.Context, lock types.PeriodLock, endTime time.Time) error {
	// remove lock refs from not unlocking queue if exists
	err := k.deleteLockRefs(ctx, types.KeyPrefixNotUnlocking, lock)
	if err != nil {
		return err
	}

	// store lock with end time set
	lock.EndTime = endTime
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

// UnlockMaturedLock finishes unlocking by sending back the locked tokens from the module accounts
// to the owner. This method requires lock to be matured, having passed the endtime of the lock.
func (k Keeper) UnlockMaturedLock(ctx sdk.Context, lockID uint64) error {
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

	return k.unlockMaturedLockInternalLogic(ctx, *lock)
}

// ForceUnlock ignores unlock duration and immediately unlocks the lock and refunds tokens to lock owner.
func (k Keeper) ForceUnlock(ctx sdk.Context, lock types.PeriodLock) error {
	// Steps:
	// 1) Break associated synthetic locks. (Superfluid data)
	// 2) If lock is bonded, move it to unlocking
	// 3) Run logic to delete unlocking metadata, and send tokens to owner.

	synthLocks := k.GetAllSyntheticLockupsByLockup(ctx, lock.ID)
	err := k.DeleteAllSyntheticLocks(ctx, lock, synthLocks)
	if err != nil {
		return err
	}

	if !lock.IsUnlocking() {
		err := k.BeginUnlock(ctx, lock.ID, nil)
		if err != nil {
			return err
		}
	}
	// NOTE: This caused a bug! BeginUnlock changes the owner the lock.EndTime
	// This shows the bad API design of not using lock.ID in every public function.
	lockPtr, err := k.GetLockByID(ctx, lock.ID)
	if err != nil {
		return err
	}
	return k.unlockMaturedLockInternalLogic(ctx, *lockPtr)
}

// unlockMaturedLockInternalLogic handles internal logic for finishing unlocking matured locks.
func (k Keeper) unlockMaturedLockInternalLogic(ctx sdk.Context, lock types.PeriodLock) error {
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

// ExtendLockup changes the existing lock duration to the given lock duration.
// Updating lock duration would fail on either of the following conditions.
// 1. Only lock owner is able to change the duration of the lock.
// 2. Locks that are unlokcing are not allowed to change duration.
// 3. Locks that have synthetic lockup are not allowed to change.
// 4. Provided duration should be greater than the original duration.
func (k Keeper) ExtendLockup(ctx sdk.Context, lockID uint64, owner sdk.AccAddress, newDuration time.Duration) error {
	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}

	if lock.GetOwner() != owner.String() {
		return types.ErrNotLockOwner
	}

	if lock.IsUnlocking() {
		return fmt.Errorf("cannot edit unlocking lockup for lock %d", lock.ID)
	}

	// check synthetic lockup exists
	if k.HasAnySyntheticLockups(ctx, lock.ID) {
		return fmt.Errorf("cannot edit lockup with synthetic lock %d", lock.ID)
	}

	// completely delete existing lock refs
	err = k.deleteLockRefs(ctx, unlockingPrefix(lock.IsUnlocking()), *lock)
	if err != nil {
		return err
	}

	oldDuration := lock.GetDuration()
	if newDuration != 0 {
		if newDuration <= oldDuration {
			return fmt.Errorf("new duration should be greater than the original")
		}

		// update accumulation store
		for _, coin := range lock.Coins {
			k.accumulationStore(ctx, coin.Denom).Decrease(accumulationKey(lock.Duration), coin.Amount)
			k.accumulationStore(ctx, coin.Denom).Increase(accumulationKey(newDuration), coin.Amount)
		}

		lock.Duration = newDuration
	}

	// add lock refs with the new duration
	err = k.addLockRefs(ctx, *lock)
	if err != nil {
		return err
	}

	err = k.setLock(ctx, *lock)
	if err != nil {
		return err
	}

	k.hooks.OnLockupExtend(ctx,
		lock.GetID(),
		oldDuration,
		lock.GetDuration(),
	)

	return nil
}

// InitializeAllLocks takes a set of locks, and initializes state to be storing
// them all correctly. This utilizes batch optimizations to improve efficiency,
// as this becomes a bottleneck at chain initialization & upgrades.
func (k Keeper) InitializeAllLocks(ctx sdk.Context, locks []types.PeriodLock) error {
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
		err := k.setLockAndAddLockRefs(ctx, lock)
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

func (k Keeper) InitializeAllSyntheticLocks(ctx sdk.Context, syntheticLocks []types.SyntheticLock) error {
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

// SlashTokensFromLockByID sends slashed tokens directly from the lock to the community pool.
// Called by the superfluid module ONLY.
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

func (k Keeper) accumulationStore(ctx sdk.Context, denom string) store.Tree {
	return store.NewTree(prefix.NewStore(ctx.KVStore(k.storeKey), accumulationStorePrefix(denom)), 10)
}

// removeTokensFromLock is called by lockup slash function.
// Called by the superfluid module ONLY.
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

// setLockAndAddLockRefs sets the lock, and resets all of its lock references
// This puts the lock into a 'clean' state, aside from the AccumulationStore.
func (k Keeper) setLockAndAddLockRefs(ctx sdk.Context, lock types.PeriodLock) error {
	err := k.setLock(ctx, lock)
	if err != nil {
		return err
	}

	return k.addLockRefs(ctx, lock)
}

// setSyntheticLockAndResetRefs sets the synthetic lock object, and resets all of its lock references
func (k Keeper) setSyntheticLockAndResetRefs(ctx sdk.Context, lock types.PeriodLock, synthLock types.SyntheticLock) error {
	err := k.setSyntheticLockupObject(ctx, &synthLock)
	if err != nil {
		return err
	}

	// store synth lock refs
	return k.addSyntheticLockRefs(ctx, lock, synthLock)
}

// deleteLock removes the lock object from the state.
func (k Keeper) deleteLock(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(lockStoreKey(id))
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

func (k Keeper) getCoinsFromLocks(locks []types.PeriodLock) sdk.Coins {
	coins := sdk.Coins{}
	for _, lock := range locks {
		coins = coins.Add(lock.Coins...)
	}
	return coins
}
