package keeper

import (
	"strconv"
	"time"

	"github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	HourDuration, _    = time.ParseDuration("1h")
	DayDuration, _     = time.ParseDuration("24h")
	WeekDuration, _    = time.ParseDuration("168h")
	TwoWeekDuration, _ = time.ParseDuration("336h")
	BaselineDurations  = []time.Duration{DayDuration, WeekDuration, TwoWeekDuration}
)

// baselineDurations is expected to be sorted by the caller.
func normalizeDuration(baselineDurations []time.Duration, allowedDiff time.Duration, duration time.Duration) (time.Duration, bool) {
	for _, base := range baselineDurations {
		// if base > duration, continue to next base size.
		// if base <= duration, then we are in a duration that is greater than or equal to base size.
		// If its within base + allowed diff, we set it to base.
		if base <= duration && duration < base+allowedDiff {
			return base, true
		}
	}
	return duration, false
}

// MergeLockupsForSimilarDurations iterates through every account. For each account,
// it combines all lockups it has at a similar duration (to be defined in a bit).
// It will delete every existing lockup for that account, and make at most, a single new lockup per
// "base duration", denom pair.
// If a lockup is far from any base duration, we don't change anything about it.
// We define a lockup length as a "Similar duration to base duration D", if:
// D <= lockup length <= D + durationDiff.
// Any locks with different reward receiver and lock owner would not be merged.
// NOTE: This method has been outdated, please make new migration code if there is a need to merge locks with similar duration.
func MergeLockupsForSimilarDurations(
	ctx sdk.Context,
	k Keeper,
	ak types.AccountKeeper,
	baselineDurations []time.Duration,
	durationDiff time.Duration,
) {
	numLocksCreated := 0
	for _, acc := range ak.GetAllAccounts(ctx) {
		addr := acc.GetAddress()
		// We make at most one lock per (addr, denom, base duration) triplet, which we keep adding coins to.
		// We call this the new "normalized lock", and the value in the map is the new lock ID.
		normals := make(map[string]uint64)
		locksToNormalize := []types.PeriodLock{}
		for _, lock := range k.GetAccountPeriodLocks(ctx, addr) {
			// ignore multilocks
			if len(lock.Coins) > 1 {
				continue
			}
			// ignore unlocking locks; they will be removed from the state anyway
			if lock.IsUnlocking() {
				continue
			}
			normalizedDuration, ok := normalizeDuration(baselineDurations, durationDiff, lock.Duration)
			if !ok {
				continue
			}
			coin := lock.Coins[0]

			// check if the reward receiver is the lock owner.
			// if not, we do not normalize the lock.
			if lock.RewardReceiverAddress != "" {
				continue
			}

			// In this pass, add lock to normals if exact match
			key := addr.String() + "/" + coin.Denom + "/" + strconv.FormatInt(int64(normalizedDuration), 10)
			_, normalExists := normals[key]
			if normalizedDuration == lock.Duration && !normalExists {
				normals[key] = lock.ID
				continue
			}
			locksToNormalize = append(locksToNormalize, lock)
		}

		for _, lock := range locksToNormalize {
			coin := lock.Coins[0]
			normalizedDuration, ok := normalizeDuration(baselineDurations, durationDiff, lock.Duration)
			if !ok {
				continue
			}

			// serialize (addr, denom, duration) into a unique triplet for use in normals map.
			key := addr.String() + "/" + coin.Denom + "/" + strconv.FormatInt(int64(normalizedDuration), 10)
			normalID, ok := normals[key]

			var normalLock types.PeriodLock
			if !ok {
				owner, err := sdk.AccAddressFromBech32(lock.Owner)
				if err != nil {
					panic(err)
				}
				// create a normalized lock that will absorb the locks in the duration window
				normalID = k.GetLastLockID(ctx) + 1
				normalLock = types.NewPeriodLock(normalID, owner, lock.RewardReceiverAddress, normalizedDuration, time.Time{}, lock.Coins)
				err = k.addLockRefs(ctx, normalLock)
				if err != nil {
					panic(err)
				}
				k.SetLastLockID(ctx, normalID)
				normals[key] = normalID
				numLocksCreated += 1
			} else {
				normalLockPtr, err := k.GetLockByID(ctx, normalID)
				if err != nil {
					panic(err)
				}
				normalLock = *normalLockPtr
				normalLock.Coins[0].Amount = normalLock.Coins[0].Amount.Add(coin.Amount)
			}

			// k.accumulationStore(ctx, coin.Denom).Decrease(accumulationKey(lock.Duration), coin.Amount)
			// k.accumulationStore(ctx, coin.Denom).Increase(accumulationKey(normalizedDuration), coin.Amount)

			err := k.setLock(ctx, normalLock)
			if err != nil {
				panic(err)
			}

			k.deleteLock(ctx, lock.ID)
			err = k.deleteLockRefs(ctx, types.KeyPrefixNotUnlocking, lock)
			if err != nil {
				panic(err)
			}

			// don't call hooks, tokens are just moved from a lock to another
		}
		// numLocksNormalized := len(locksToNormalize)
		// ctx.Logger().Info("normalized %d locks. This involved creating %d new locks.",
		// numLocksNormalized, numLocksCreated)
	}
}
