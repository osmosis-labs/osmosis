package keeper

import (
	"fmt"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/lockup/types"
)

var (
	HourDuration, _    = time.ParseDuration("1h")
	DayDuration, _     = time.ParseDuration("24h")
	WeekDuration, _    = time.ParseDuration("168h")
	TwoWeekDuration, _ = time.ParseDuration("336h")
	BaselineDurations  = []time.Duration{DayDuration, WeekDuration, TwoWeekDuration}
	AllowedDiff        = HourDuration
)

func MigrateLockups(
	ctx sdk.Context,
	k Keeper,
) {
	// reset accumulation store, and re-set it
	k.ClearAccumulationStores(ctx)

	// normals stores lockup id for each (owner addr, denom, normalized duration) triplet
	normals := make(map[string]uint64)
	key := func(addr sdk.AccAddress, denom string, duration time.Duration) string {
		return fmt.Sprintf("%s/%s/%s", addr.String(), denom, strconv.FormatInt(int64(duration), 10))
	}

	// getNormalLock create or get normalized lock for given triplet
	getNormalLock := func(addr sdk.AccAddress, denom string, normalizedDuration time.Duration) types.PeriodLock {
		normalID, ok := normals[key(addr, denom, normalizedDuration)]
		if ok {
			normalLock, err := k.GetLockByID(ctx, normalID)
			if err != nil {
				panic(err)
			}
			return *normalLock
		}
		// crease a normalized lock if not exists
		normalLock, err := k.createLock(ctx, addr, normalizedDuration, sdk.Coins{sdk.NewInt64Coin(denom, 0)})
		if err != nil {
			panic(err)
		}
		normals[key(addr, denom, normalizedDuration)] = normalLock.ID
		return normalLock
	}

	tryNormalizeDuration := func(lock types.PeriodLock) (res time.Duration, ok bool) {
		// multilocks and unlocking locks are not normalizable
		if len(lock.Coins) != 1 || lock.IsUnlocking() {
			return
		}

		// find out the normalizing base duration, if exists
		// if base > duration, continue to next base size.
		// if base <= duration, then we are in a duration that is greater than or equal to base size.
		// If its within base + allowed diff, we set it to base.
		for _, base := range BaselineDurations {
			if base <= lock.Duration && lock.Duration < base+AllowedDiff {
				return base, true
			}
		}
		return
	}

	mergeLockup := func(lock, normalLock types.PeriodLock) {
		// increase normal lock Coins
		normalLock.Coins = normalLock.Coins.Add(lock.Coins[0])
		err := k.setLock(ctx, normalLock)
		if err != nil {
			panic(err)
		}

		// delete lock
		err = k.deleteLock(ctx, lock)
		if err != nil {
			panic(err)
		}
	}

	locks, err := k.GetPeriodLocks(ctx)
	if err != nil {
		panic(err)
	}

	for _, lock := range locks {
		// if qualified for merging, do merge
		if normalizedDuration, ok := tryNormalizeDuration(lock); ok {
			normalLock := getNormalLock(lock.OwnerAddress(), lock.Coins[0].Denom, normalizedDuration)
			mergeLockup(lock, normalLock)
			lock = normalLock
		}

		// increase accumulationstore
		for _, coin := range lock.Coins {
			k.accumulationStore(ctx, coin.Denom).Increase(accumulationKey(lock.Duration), coin.Amount)
		}
	}
}

// MergeLockupsForSimilarDurations iterates through every account. For each account,
// it combines all lockups it has at a similar duration (to be defined in a bit).
// It will delete every existing lockup for that account, and make at most, a single new lockup per
// "base duration", denom pair.
// If a lockup is far from any base duration, we don't change anything about it.
// We define a lockup length as a "Similar duration to base duration D", if:
// D <= lockup length <= D + durationDiff.
