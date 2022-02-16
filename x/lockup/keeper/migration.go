package keeper

import (
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/lockup/types"
)

var (
	HourDuration, _    = time.ParseDuration("1h")
	DayDuration, _     = time.ParseDuration("24h")
	WeekDuration, _    = time.ParseDuration("168h")
	TwoWeekDuration, _ = time.ParseDuration("336h")
	BaselineDurations  = []time.Duration{DayDuration, WeekDuration, TwoWeekDuration}
)

func normalizeDuration(baselineDurations []time.Duration, allowedDiff time.Duration, duration time.Duration) (time.Duration, bool) {
	for _, base := range baselineDurations {
		if base <= duration && duration < base+allowedDiff {
			return base, true
		}
	}
	return duration, false
}

// between duration windows of baselineDurations.map((duration) => (duration, duration+durationDiff)).
func MergeLockupsForSimilarDurations(
	ctx sdk.Context,
	k Keeper,
	ak types.AccountKeeper,
	baselineDurations []time.Duration,
	durationDiff time.Duration,
) {
	for _, acc := range ak.GetAllAccounts(ctx) {
		addr := acc.GetAddress()
		normals := make(map[string]uint64)
		for _, lock := range k.GetAccountPeriodLocks(ctx, addr) {
			// ignore multilocks
			if len(lock.Coins) > 1 {
				continue
			}
			// ignore unlocking locks; they will be removed from the state anyway
			if lock.IsUnlocking() {
				continue
			}
			coin := lock.Coins[0]
			normalizedDuration, ok := normalizeDuration(baselineDurations, durationDiff, lock.Duration)
			if !ok {
				continue
			}

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
				normalLock = types.NewPeriodLock(normalID, owner, normalizedDuration, time.Time{}, lock.Coins)
				k.addLockRefs(ctx, types.KeyPrefixNotUnlocking, normalLock)
				k.SetLastLockID(ctx, normalID)
				normals[key] = normalID
			} else {
				normalLockPtr, err := k.GetLockByID(ctx, normalID)
				if err != nil {
					panic(err)
				}
				normalLock = *normalLockPtr
				normalLock.Coins[0].Amount = normalLock.Coins[0].Amount.Add(coin.Amount)
			}

			k.accumulationStore(ctx, coin.Denom).Decrease(accumulationKey(lock.Duration), coin.Amount)
			k.accumulationStore(ctx, coin.Denom).Increase(accumulationKey(normalizedDuration), coin.Amount)

			k.setLock(ctx, normalLock)

			k.deleteLock(ctx, lock.ID)
			k.deleteLockRefs(ctx, types.KeyPrefixNotUnlocking, lock)

			// don't call hooks, tokens are just moved from a lock to another
		}
	}
}
