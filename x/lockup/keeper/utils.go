package keeper

import (
	"bytes"
	"time"

	"github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// combineKeys combine bytes array into a single bytes.
func combineKeys(keys ...[]byte) []byte {
	return bytes.Join(keys, types.KeyIndexSeparator)
}

// getTimeKey returns the key used for getting a set of period locks
// where unlockTime is after a specific time.
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
// where duration is longer than a specific duration.
func getDurationKey(duration time.Duration) []byte {
	if duration < 0 {
		duration = 0
	}
	key := sdk.Uint64ToBigEndian(uint64(duration))
	return combineKeys(types.KeyPrefixDuration, key)
}

func durationLockRefKeys(lock types.PeriodLock) ([][]byte, error) {
	refKeys := [][]byte{}
	durationKey := getDurationKey(lock.Duration)
	owner, err := sdk.AccAddressFromBech32(lock.Owner)
	if err != nil {
		return nil, err
	}

	refKeys = append(refKeys, combineKeys(types.KeyPrefixLockDuration, durationKey))
	refKeys = append(refKeys, combineKeys(types.KeyPrefixAccountLockDuration, owner, durationKey))

	for _, coin := range lock.Coins {
		denomBz := []byte(coin.Denom)
		refKeys = append(refKeys, combineKeys(types.KeyPrefixDenomLockDuration, denomBz, durationKey))
		refKeys = append(refKeys, combineKeys(types.KeyPrefixAccountDenomLockDuration, owner, denomBz, durationKey))
	}
	return refKeys, nil
}

func lockRefKeys(lock types.PeriodLock) ([][]byte, error) {
	refKeys, _ := durationLockRefKeys(lock)
	timeKey := getTimeKey(lock.EndTime)

	owner, err := sdk.AccAddressFromBech32(lock.Owner)
	if err != nil {
		return nil, err
	}

	refKeys = append(refKeys, combineKeys(types.KeyPrefixLockTimestamp, timeKey))
	refKeys = append(refKeys, combineKeys(types.KeyPrefixAccountLockTimestamp, owner, timeKey))

	for _, coin := range lock.Coins {
		denomBz := []byte(coin.Denom)
		refKeys = append(refKeys, combineKeys(types.KeyPrefixDenomLockTimestamp, denomBz, timeKey))
		refKeys = append(refKeys, combineKeys(types.KeyPrefixAccountDenomLockTimestamp, owner, denomBz, timeKey))
	}
	return refKeys, nil
}

// syntheticLockRefKeys are different from native lockRefKeys to avoid conflicts
// They differ by using the synth denom rather than the native denom.
// All the values at each lockref key points to the underlying lock ID of the synth lock though.
func syntheticLockRefKeys(lock types.PeriodLock, synthLock types.SyntheticLock) ([][]byte, error) {
	// Note: syntheticLockRefKeys should be only used for querying and should not be combined with native lockup operations
	// synthetic suffix denom should not conflict with native denom
	refKeys := [][]byte{}
	timeKey := getTimeKey(synthLock.EndTime)
	durationKey := getDurationKey(synthLock.Duration)

	owner, err := sdk.AccAddressFromBech32(lock.Owner)
	if err != nil {
		return nil, err
	}

	denomBz := []byte(synthLock.SynthDenom)
	refKeys = append(refKeys, combineKeys(types.KeyPrefixDenomLockTimestamp, denomBz, timeKey))
	refKeys = append(refKeys, combineKeys(types.KeyPrefixDenomLockDuration, denomBz, durationKey))
	refKeys = append(refKeys, combineKeys(types.KeyPrefixAccountDenomLockTimestamp, owner, denomBz, timeKey))
	refKeys = append(refKeys, combineKeys(types.KeyPrefixAccountDenomLockDuration, owner, denomBz, durationKey))

	return refKeys, nil
}

func combineLocks(pl1 []types.PeriodLock, pl2 []types.PeriodLock) []types.PeriodLock {
	return append(pl1, pl2...)
}
