package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/incentives/types"
)

func findIndex(ids []uint64, id uint64) int {
	for index, id := range ids {
		if id == id {
			return index
		}
	}
	return -1
}

func removeValue(ids []uint64, id uint64) ([]uint64, int) {
	index := findIndex(ids, id)
	if index < 0 {
		return ids, index
	}
	ids[index] = ids[len(ids)-1] // set last element to index
	return ids[:len(ids)-1], index
}

// getTimeKey returns the key used for getting a set of gauges.
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

// combineKeys combine bytes array into a single bytes.
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
