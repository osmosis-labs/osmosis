package keeper

import (
	"time"

	"github.com/osmosis-labs/osmosis/v12/x/incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// findIndex takes an array of IDs. Then return the index of a specific ID.
func findIndex(iDs []uint64, iD uint64) int {
	for index, id := range iDs {
		if id == iD {
			return index
		}
	}
	return -1
}

// removeValue takes an array of IDs. Then finds the index of the IDs and remove those IDs from the array.
func removeValue(iDs []uint64, iD uint64) ([]uint64, int) {
	index := findIndex(iDs, iD)
	if index < 0 {
		return iDs, index
	}
	iDs[index] = iDs[len(iDs)-1] // set last element to index
	return iDs[:len(iDs)-1], index
}

// getTimeKey returns the time key used when getting a set of gauges.
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

// combineKeys combines the byte arrays of multiple keys into a single byte array.
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
