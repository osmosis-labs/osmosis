package keeper

import (
	"time"

	"github.com/osmosis-labs/osmosis/v10/x/incentives/types"

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
