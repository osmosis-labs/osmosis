package keeper

import (
	"time"

	"github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func findIndex(IDs []uint64, ID uint64) int {
	for index, id := range IDs {
		if id == ID {
			return index
		}
	}
	return -1
}

func removeValue(IDs []uint64, ID uint64) ([]uint64, int) {
	index := findIndex(IDs, ID)
	if index < 0 {
		return IDs, index
	}
	IDs[index] = IDs[len(IDs)-1] // set last element to index
	return IDs[:len(IDs)-1], index
}

// combineKeys combine bytes array into a single bytes
func combineKeys(keys ...[]byte) []byte {
	combined := []byte{}
	for i, key := range keys {
		combined = append(combined, key...)
		if i < len(keys)-1 { // not last item
			combined = append(combined, types.KeyIndexSeparator...)
		}
	}
	return combined
}

// getTimeKey returns the key used for getting a set of period locks
// where unlockTime is after a specific time
func getTimeKey(timestamp time.Time) []byte {
	timeBz := sdk.FormatTimeBytes(timestamp)
	timeBzL := len(timeBz)
	prefixL := len(types.KeyPrefixTimestamp)

	bz := make([]byte, prefixL+8+timeBzL)

	// copy the prefix
	copy(bz[:prefixL], types.KeyPrefixTimestamp)

	// copy the encoded time bytes length
	copy(bz[prefixL:prefixL+8], sdk.Uint64ToBigEndian(uint64(timeBzL)))

	// copy the encoded time bytes
	copy(bz[prefixL+8:prefixL+8+timeBzL], timeBz)
	return bz
}

// getDurationKey returns the key used for getting a set of period locks
// where duration is longer than a specific duration
func getDurationKey(duration time.Duration) []byte {
	if duration < 0 {
		duration = 0
	}
	key := sdk.Uint64ToBigEndian(uint64(duration))
	return combineKeys(types.KeyPrefixDuration, key)
}

func lockRefKeys(lock types.PeriodLock) [][]byte {
	refKeys := [][]byte{}
	timeKey := getTimeKey(lock.EndTime)
	durationKey := getDurationKey(lock.Duration)
	refKeys = append(refKeys, combineKeys(types.KeyPrefixLockTimestamp, timeKey))
	refKeys = append(refKeys, combineKeys(types.KeyPrefixLockDuration, durationKey))
	refKeys = append(refKeys, combineKeys(types.KeyPrefixAccountLockTimestamp, []byte(lock.Owner), timeKey))
	refKeys = append(refKeys, combineKeys(types.KeyPrefixAccountLockDuration, []byte(lock.Owner), durationKey))

	for _, coin := range lock.Coins {
		denomBz := []byte(coin.Denom)
		refKeys = append(refKeys, combineKeys(types.KeyPrefixDenomLockTimestamp, denomBz, timeKey))
		refKeys = append(refKeys, combineKeys(types.KeyPrefixDenomLockDuration, denomBz, durationKey))
		refKeys = append(refKeys, combineKeys(types.KeyPrefixAccountDenomLockTimestamp, []byte(lock.Owner), denomBz, timeKey))
		refKeys = append(refKeys, combineKeys(types.KeyPrefixAccountDenomLockDuration, []byte(lock.Owner), denomBz, durationKey))
	}
	return refKeys
}

func combineLocks(pl1 []types.PeriodLock, pl2 []types.PeriodLock) []types.PeriodLock {
	return append(pl1, pl2...)
}
