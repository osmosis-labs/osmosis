package types

import (
	"fmt"
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var MaxSpotPrice = sdk.NewDec(2).Power(128).Sub(sdk.OneDec())

// GetAllUniqueDenomPairs returns all unique pairs of denoms, where for every pair
// (X, Y), X < Y.
// The pair (X,Y) should only appear once in the list. Denoms are lexicographically sorted.
// Panics if finds duplicate pairs.
//
// NOTE: Sorts the input denoms slice.
func GetAllUniqueDenomPairs(denoms []string) []DenomPair {
	// get denoms in ascending order
	sort.Strings(denoms)

	denomPairs := []DenomPair{}

	for i := 0; i < len(denoms); i++ {
		for j := i + 1; j < len(denoms); j++ {
			if denoms[i] == denoms[j] {
				panic("input had duplicated denom")
			}
			denomPairs = append(denomPairs, DenomPair{Denom0: denoms[i], Denom1: denoms[j]})
		}
	}

	return denomPairs
}

// SpotPriceMulDuration returns the spot price multiplied by the time delta,
// that is the spot price between the current and last TWAP record.
// A single second accounts for 1_000_000_000 when converted to int64.
func SpotPriceMulDuration(sp sdk.Dec, timeDelta time.Duration) sdk.Dec {
	deltaMS := timeDelta.Milliseconds()
	return sp.MulInt64(deltaMS)
}

// AccumDiffDivDuration returns the accumulated difference divided by the the
// time delta, that is the spot price between the current and last TWAP record.
func AccumDiffDivDuration(accumDiff sdk.Dec, timeDelta time.Duration) sdk.Dec {
	deltaMS := timeDelta.Milliseconds()
	return accumDiff.QuoInt64(deltaMS)
}

// LexicographicalOrderDenoms takes two denoms and returns them to be in lexicographically ascending order.
// In other words, the first returned denom string will be the lexicographically smaller of the two denoms.
// If the denoms are equal, an error will be returned.
func LexicographicalOrderDenoms(denom0, denom1 string) (string, string, error) {
	if denom0 == denom1 {
		return "", "", fmt.Errorf("both assets cannot be of the same denom: assetA: %s, assetB: %s", denom0, denom1)
	}
	if denom0 > denom1 {
		denom0, denom1 = denom1, denom0
	}
	return denom0, denom1, nil
}

// DenomPair contains pair of assetA and assetB denoms which belong to a pool.
type DenomPair struct {
	Denom0 string
	Denom1 string
}
