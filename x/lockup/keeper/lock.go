package keeper

import (
	"fmt"
	"sort"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/gogoproto/proto"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/sumtree"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v27/x/lockup/types"
)

// WithdrawAllMaturedLocks withdraws every lock that's in the process of unlocking, and has finished unlocking by
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
func (k Keeper) GetPeriodLocksAccumulation(ctx sdk.Context, query types.QueryCondition) osmomath.Int {
	beginKey := accumulationKey(query.Duration)
	return k.accumulationStore(ctx, query.Denom).SubsetAccumulation(beginKey, nil)
}

// BeginUnlockAllNotUnlockings begins unlock for all not unlocking locks of the given account.
func (k Keeper) BeginUnlockAllNotUnlockings(ctx sdk.Context, account sdk.AccAddress) ([]types.PeriodLock, error) {
	locks, err := k.beginUnlockFromIterator(ctx, k.AccountLockIterator(ctx, false, account))
	return locks, err
}

// AddToExistingLock adds the given coin to the existing lock with the same owner and duration.
// Returns the updated lock ID if successfully added coin, returns 0 and error when a lock with
// given condition does not exist, or if fails to add to lock.
func (k Keeper) AddToExistingLock(ctx sdk.Context, owner sdk.AccAddress, coin sdk.Coin, duration time.Duration) (uint64, error) {
	locks := k.GetAccountLockedDurationNotUnlockingOnly(ctx, owner, coin.Denom, duration)

	// if no lock exists for the given owner + denom + duration, return an error
	if len(locks) < 1 {
		return 0, errorsmod.Wrapf(types.ErrLockupNotFound, "lock with denom %s before duration %s does not exist", coin.Denom, duration.String())
	}

	// if existing lock with same duration and denom exists, add to the existing lock
	// there should only be a single lock with the same duration + token, thus we take the first lock
	lock := locks[0]
	_, err := k.AddTokensToLockByID(ctx, lock.ID, owner, coin)
	if err != nil {
		return 0, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
	}

	return lock.ID, nil
}

// HasLock returns true if lock with the given condition exists
func (k Keeper) HasLock(ctx sdk.Context, owner sdk.AccAddress, denom string, duration time.Duration) bool {
	locks := k.GetAccountLockedDurationNotUnlockingOnly(ctx, owner, denom, duration)
	return len(locks) > 0
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

	// Send the tokens we are about to add to lock to the lockup module account.
	if err := k.bk.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, sdk.NewCoins(tokensToAdd)); err != nil {
		return nil, err
	}

	err = k.lock(ctx, *lock, sdk.NewCoins(tokensToAdd))
	if err != nil {
		return nil, err
	}

	// TODO: Handle found case in a better way, with state breaking update
	synthlock, _, err := k.GetSyntheticLockupByUnderlyingLockId(ctx, lock.ID)
	if err != nil {
		return nil, err
	}
	k.accumulationStore(ctx, synthlock.SynthDenom).Increase(accumulationKey(synthlock.Duration), tokensToAdd.Amount)

	if k.hooks == nil {
		return lock, nil
	}

	k.hooks.AfterAddTokensToLock(ctx, lock.OwnerAddress(), lock.GetID(), sdk.NewCoins(tokensToAdd))

	return lock, nil
}

// CreateLock creates a new lock with the specified duration for the owner.
// Returns an error in the following conditions:
//   - account does not have enough balance
func (k Keeper) CreateLock(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, duration time.Duration) (types.PeriodLock, error) {
	// Send the coins we are about to lock to the lockup module account.
	if err := k.bk.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, coins); err != nil {
		return types.PeriodLock{}, err
	}

	// Run the createLock logic without the send since we sent the coins above.
	lock, err := k.CreateLockNoSend(ctx, owner, coins, duration)
	if err != nil {
		return types.PeriodLock{}, err
	}
	return lock, nil
}

// CreateLockNoSend behaves the same as CreateLock, but does not send the coins to the lockup module account.
// This method is used in the concentrated liquidity module since we mint coins directly to the lockup module account.
// We do not want to mint the coins to send to the user just to send them back to the lockup module account for two reasons:
//   - it is gas inefficient
//   - users should not be able to have cl shares in their account, so this is an extra safety measure
func (k Keeper) CreateLockNoSend(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, duration time.Duration) (types.PeriodLock, error) {
	ID := k.GetLastLockID(ctx) + 1
	// unlock time is initially set without a value, gets set as unlock start time + duration
	// when unlocking starts.
	// the reward receiver is set as the owner by default when creating a lock, and we indicate this by using an empty string.
	lock := types.NewPeriodLock(ID, owner, "", duration, time.Time{}, coins)

	// lock the coins without sending them to the lockup module account
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
// WARNING: this method does not send the underlying coins to the lockup module account.
// This must be done by the caller.
func (k Keeper) lock(ctx sdk.Context, lock types.PeriodLock, tokensToLock sdk.Coins) error {
	owner, err := sdk.AccAddressFromBech32(lock.Owner)
	if err != nil {
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
func (k Keeper) BeginUnlock(ctx sdk.Context, lockID uint64, coins sdk.Coins) (uint64, error) {
	// prohibit BeginUnlock if synthetic locks are referring to this
	// TODO: In the future, make synthetic locks only get partial restrictions on the main lock.
	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return 0, err
	}
	if k.HasAnySyntheticLockups(ctx, lock.ID) {
		return 0, fmt.Errorf("cannot BeginUnlocking a lock with synthetic lockup")
	}

	unlockingLock, err := k.beginUnlock(ctx, *lock, coins)
	return unlockingLock, err
}

// BeginForceUnlock begins force unlock of the given lock.
// This method should be called by the superfluid module ONLY, as it does not check whether
// the lock has a synthetic lock or not before unlocking.
// Returns lock id, new lock id if the lock was split, else same lock id.
func (k Keeper) BeginForceUnlock(ctx sdk.Context, lockID uint64, coins sdk.Coins) (uint64, error) {
	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return 0, err
	}

	lockID, err = k.beginUnlock(ctx, *lock, coins)
	if err != nil {
		return 0, err
	}

	return lockID, nil
}

// beginUnlock unlocks specified tokens from the given lock. Existing lock refs
// of not unlocking queue are deleted and new lock refs are then added.
// EndTime of the lock is set within this method.
// Coins provided as the parameter does not require to have all the tokens in the lock,
// as we allow partial unlockings of a lock.
// Returns lock id, new lock id if the lock was split, else same lock id.
func (k Keeper) beginUnlock(ctx sdk.Context, lock types.PeriodLock, coins sdk.Coins) (uint64, error) {
	// sanity check
	if !coins.IsAllLTE(lock.Coins) {
		return 0, fmt.Errorf("requested amount to unlock exceeds locked tokens")
	}

	if lock.IsUnlocking() {
		return 0, fmt.Errorf("trying to unlock a lock that is already unlocking")
	}

	// If the amount were unlocking is empty, or the entire coins amount, unlock the entire lock.
	// Otherwise, split the lock into two locks, and fully unlock the newly created lock.
	// (By virtue, the newly created lock we split into should have the unlock amount)
	if len(coins) != 0 && !coins.Equal(lock.Coins) {
		splitLock, err := k.SplitLock(ctx, lock, coins, false)
		if err != nil {
			return 0, err
		}
		lock = splitLock
	}

	// remove existing lock refs from not unlocking queue
	err := k.deleteLockRefs(ctx, types.KeyPrefixNotUnlocking, lock)
	if err != nil {
		return 0, err
	}

	// store lock with the end time set to current block time + duration
	lock.EndTime = ctx.BlockTime().Add(lock.Duration)
	err = k.setLock(ctx, lock)
	if err != nil {
		return 0, err
	}

	// add lock refs into unlocking queue
	err = k.addLockRefs(ctx, lock)
	if err != nil {
		return 0, err
	}

	if k.hooks != nil {
		k.hooks.OnStartUnlock(ctx, lock.OwnerAddress(), lock.ID, lock.Coins, lock.Duration, lock.EndTime)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		createBeginUnlockEvent(&lock),
	})

	return lock.ID, nil
}

func (k Keeper) clearKeysByPrefix(ctx sdk.Context, prefix []byte) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
}

func (k Keeper) RebuildAccumulationStoreForDenom(ctx sdk.Context, denom string) {
	prefix := accumulationStorePrefix(denom)
	k.clearKeysByPrefix(ctx, prefix)
	locks := k.GetLocksDenom(ctx, denom)
	mapDurationToAmount := make(map[time.Duration]osmomath.Int)
	for _, lock := range locks {
		if v, ok := mapDurationToAmount[lock.Duration]; ok {
			mapDurationToAmount[lock.Duration] = v.Add(lock.Coins.AmountOf(denom))
		} else {
			mapDurationToAmount[lock.Duration] = lock.Coins.AmountOf(denom)
		}
	}

	k.writeDurationValuesToAccumTree(ctx, denom, mapDurationToAmount)
}

func (k Keeper) RebuildSuperfluidAccumulationStoresForDenom(ctx sdk.Context, denom string) {
	superfluidPrefix := denom + "/super"
	superfluidStorePrefix := accumulationStorePrefix(superfluidPrefix)
	// remove trailing slash
	superfluidStorePrefix = superfluidStorePrefix[0 : len(superfluidStorePrefix)-1]
	k.clearKeysByPrefix(ctx, superfluidStorePrefix)

	accumulationStoreEntries := make(map[string]map[time.Duration]osmomath.Int)
	locks := k.GetLocksDenom(ctx, denom)
	for _, lock := range locks {
		synthLock, found, err := k.GetSyntheticLockupByUnderlyingLockId(ctx, lock.ID)
		if err != nil || !found {
			continue
		}

		var curDurationMap map[time.Duration]osmomath.Int
		if durationMap, ok := accumulationStoreEntries[synthLock.SynthDenom]; ok {
			curDurationMap = durationMap
		} else {
			curDurationMap = make(map[time.Duration]osmomath.Int)
		}
		newAmt := lock.Coins.AmountOf(denom)
		if curAmt, ok := curDurationMap[synthLock.Duration]; ok {
			newAmt = newAmt.Add(curAmt)
		}
		curDurationMap[synthLock.Duration] = newAmt
		accumulationStoreEntries[synthLock.SynthDenom] = curDurationMap
	}

	for synthDenom, durationMap := range accumulationStoreEntries {
		k.writeDurationValuesToAccumTree(ctx, synthDenom, durationMap)
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

// PartialForceUnlock begins partial ForceUnlock of given lock for the given amount of coins.
// ForceUnlocks the lock as a whole when provided coins are empty, or coin provided equals amount of coins in the lock.
// This also supports the case of lock in an unbonding status.
func (k Keeper) PartialForceUnlock(ctx sdk.Context, lock types.PeriodLock, coins sdk.Coins) error {
	// sanity check
	if !coins.IsAllLTE(lock.Coins) {
		return fmt.Errorf("requested amount to unlock exceeds locked tokens")
	}

	// split lock to support partial force unlock.
	// (By virtue, the newly created lock we split into should have the unlock amount)
	if len(coins) != 0 && !coins.Equal(lock.Coins) {
		splitLock, err := k.SplitLock(ctx, lock, coins, true)
		if err != nil {
			return err
		}
		lock = splitLock
	}

	return k.ForceUnlock(ctx, lock)
}

// ForceUnlock ignores unlock duration and immediately unlocks the lock and refunds tokens to lock owner.
func (k Keeper) ForceUnlock(ctx sdk.Context, lock types.PeriodLock) error {
	// Steps:
	// 1) Break associated synthetic lock. (Superfluid data)
	// 2) If lock is bonded, move it to unlocking
	// 3) Run logic to delete unlocking metadata, and send tokens to owner.

	// TODO: Use found instead of !synthLock.IsNil() later on.
	synthLock, _, err := k.GetSyntheticLockupByUnderlyingLockId(ctx, lock.ID)
	if err != nil {
		return err
	}
	if !synthLock.IsNil() {
		err = k.DeleteSyntheticLockup(ctx, lock.ID, synthLock.SynthDenom)
		if err != nil {
			return err
		}
	}

	if !lock.IsUnlocking() {
		_, err := k.BeginUnlock(ctx, lock.ID, nil)
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

	// If the lock contains CL liquidity tokens, we do not send them back to the owner.
	coins := lock.Coins
	finalCoinsToSendBackToUser := sdk.NewCoins()
	for _, coin := range coins {
		if strings.HasPrefix(coin.Denom, cltypes.ConcentratedLiquidityTokenPrefix) {
			// If the coin is a CL liquidity token, we do not add it to the finalCoinsToSendBackToUser and instead burn it
			err := k.bk.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(coin))
			if err != nil {
				return err
			}
		} else {
			// Otherwise, we add it to the finalCoinsToSendBackToUser
			finalCoinsToSendBackToUser = finalCoinsToSendBackToUser.Add(coin)
		}
	}

	// send coins back to owner
	// if the lock was made completely of CL liquidity tokens, this will be a no-op
	if !finalCoinsToSendBackToUser.Empty() {
		if err := k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, owner, finalCoinsToSendBackToUser); err != nil {
			return err
		}
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

// SetLockRewardReceiverAddress changes the reward recipient address to the given address.
// Storing an empty string for reward receiver would indicate the owner being reward receiver.
func (k Keeper) SetLockRewardReceiverAddress(ctx sdk.Context, lockID uint64, owner sdk.AccAddress, newReceiverAddress string) error {
	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}
	// check if the lock owner is the method caller.
	if lock.GetOwner() != owner.String() {
		return types.ErrNotLockOwner
	}

	// if the given receiver address is same as the lock owner, we store an empty string instead.
	if lock.Owner == newReceiverAddress {
		newReceiverAddress = types.DefaultOwnerReceiverPlaceholder
	}

	if lock.RewardReceiverAddress == newReceiverAddress {
		return types.ErrRewardReceiverIsSame
	}

	lock.RewardReceiverAddress = newReceiverAddress

	err = k.setLock(ctx, *lock)
	if err != nil {
		return err
	}

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
	accumulationStoreEntries := make(map[string]map[time.Duration]osmomath.Int)
	denoms := []string{}
	for i, lock := range locks {
		if i%25000 == 0 {
			msg := fmt.Sprintf("Reset %d lock refs, cur lock ID %d", i, lock.ID)
			ctx.Logger().Debug(msg)
		}
		err := k.setLockAndAddLockRefs(ctx, lock)
		if err != nil {
			return err
		}

		// Add to the accumulation store cache
		for _, coin := range lock.Coins {
			// update or create the new map from duration -> Int for this denom.
			var curDurationMap map[time.Duration]osmomath.Int
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
				curDurationMap = map[time.Duration]osmomath.Int{lock.Duration: coin.Amount}
			}
			accumulationStoreEntries[coin.Denom] = curDurationMap
		}
	}

	// deterministically iterate over durationMap cache.
	sort.Strings(denoms)
	for _, denom := range denoms {
		curDurationMap := accumulationStoreEntries[denom]
		k.writeDurationValuesToAccumTree(ctx, denom, curDurationMap)
	}

	return nil
}

func (k Keeper) writeDurationValuesToAccumTree(ctx sdk.Context, denom string, durationValueMap map[time.Duration]osmomath.Int) {
	durations := make([]time.Duration, 0, len(durationValueMap))
	for duration := range durationValueMap {
		durations = append(durations, duration)
	}
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	// now that we have a sorted list of durations for this denom,
	// add them all to accumulation store
	msg := fmt.Sprintf("Setting accumulation entries for locks for %s, there are %d distinct durations",
		denom, len(durations))
	ctx.Logger().Debug(msg)
	for _, d := range durations {
		amt := durationValueMap[d]
		k.accumulationStore(ctx, denom).Increase(accumulationKey(d), amt)
	}
}

func (k Keeper) InitializeAllSyntheticLocks(ctx sdk.Context, syntheticLocks []types.SyntheticLock) error {
	// index by coin.Denom, them duration -> amt
	// We accumulate the accumulation store entries separately,
	// to avoid hitting the myriad of slowdowns in the SDK iterator creation process.
	// We then save these once to the accumulation store at the end.
	accumulationStoreEntries := make(map[string]map[time.Duration]osmomath.Int)
	denoms := []string{}
	for i, synthLock := range syntheticLocks {
		if i%25000 == 0 {
			msg := fmt.Sprintf("Reset %d synthetic lock refs", i)
			ctx.Logger().Debug(msg)
		}

		// Add to the accumulation store cache
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

		var curDurationMap map[time.Duration]osmomath.Int
		if durationMap, ok := accumulationStoreEntries[synthLock.SynthDenom]; ok {
			curDurationMap = durationMap
			newAmt := coin.Amount
			if curAmt, ok := durationMap[synthLock.Duration]; ok {
				newAmt = newAmt.Add(curAmt)
			}
			curDurationMap[synthLock.Duration] = newAmt
		} else {
			denoms = append(denoms, synthLock.SynthDenom)
			curDurationMap = map[time.Duration]osmomath.Int{synthLock.Duration: coin.Amount}
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
		ctx.Logger().Debug(msg)
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
	err = k.ck.FundCommunityPool(ctx, coins, modAddr)
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

// SlashTokensFromLockByIDSendUnderlyingAndBurn performs the same logic as SlashTokensFromLockByID, but
// 1. Sends the underlying tokens from the pool address to the community pool (instead of sending the liquidity shares from the module account to the community pool)
// 2. Burns the liquidity shares from the module account (instead of sending them to the community pool)
func (k Keeper) SlashTokensFromLockByIDSendUnderlyingAndBurn(ctx sdk.Context, lockID uint64, liquiditySharesToSlash, underlyingPositionAssets sdk.Coins, poolAddress sdk.AccAddress) (*types.PeriodLock, error) {
	lock, err := k.GetLockByID(ctx, lockID)
	if err != nil {
		return nil, err
	}

	// Send the underlying assets of the concentrated liquidity position that the liquidity shares represent from the concentrated pool address to the community pool.
	err = k.ck.FundCommunityPool(ctx, underlyingPositionAssets, poolAddress)
	if err != nil {
		return nil, err
	}

	// Burn the liquidity shares of the concentrated liquidity position residing in the lockup module account.
	err = k.bk.BurnCoins(ctx, types.ModuleName, liquiditySharesToSlash)
	if err != nil {
		return nil, err
	}

	// Also, remove these liquidity shares from the lock.
	err = k.removeTokensFromLock(ctx, lock, liquiditySharesToSlash)
	if err != nil {
		return nil, err
	}

	if k.hooks == nil {
		return lock, nil
	}

	k.hooks.OnTokenSlashed(ctx, lock.ID, liquiditySharesToSlash)
	return lock, nil
}

func (k Keeper) accumulationStore(ctx sdk.Context, denom string) sumtree.Tree {
	return sumtree.NewTree(prefix.NewStore(ctx.KVStore(k.storeKey), accumulationStorePrefix(denom)), 10)
}

// removeTokensFromLock is called by lockup slash function.
// Called by the superfluid module ONLY.
func (k Keeper) removeTokensFromLock(ctx sdk.Context, lock *types.PeriodLock, coins sdk.Coins) error {
	// TODO: Handle 100% slash eventually, not needed for osmosis codebase atm.
	lock.Coins = lock.Coins.Sub(coins...)

	err := k.setLock(ctx, *lock)
	if err != nil {
		return err
	}

	// modifications to accumulation store
	for _, coin := range coins {
		k.accumulationStore(ctx, coin.Denom).Decrease(accumulationKey(lock.Duration), coin.Amount)
	}

	// increase synthetic lockup's accumulation store
	// TODO: In next state break, do err != nil || found == false
	synthLock, _, err := k.GetSyntheticLockupByUnderlyingLockId(ctx, lock.ID)
	if err != nil {
		return err
	}

	// Note: since synthetic lockup deletion is using native lockup's coins to reduce accumulation store
	// all the synthetic lockups' accumulation should be decreased
	k.accumulationStore(ctx, synthLock.SynthDenom).Decrease(accumulationKey(synthLock.Duration), coins[0].Amount)
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

// SplitLock splits a lock with the given amount, and stores split new lock to the state.
// Returns the new lock after modifying the state of the old lock.
func (k Keeper) SplitLock(ctx sdk.Context, lock types.PeriodLock, coins sdk.Coins, forceUnlock bool) (types.PeriodLock, error) {
	if !forceUnlock && lock.IsUnlocking() {
		return types.PeriodLock{}, fmt.Errorf("cannot split unlocking lock")
	}

	lock.Coins = lock.Coins.Sub(coins...)
	err := k.setLock(ctx, lock)
	if err != nil {
		return types.PeriodLock{}, err
	}

	// create a new lock
	splitLockID := k.GetLastLockID(ctx) + 1
	k.SetLastLockID(ctx, splitLockID)

	splitLock := types.NewPeriodLock(splitLockID, lock.OwnerAddress(), lock.RewardReceiverAddress, lock.Duration, lock.EndTime, coins)

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
