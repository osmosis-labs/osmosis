package types

import (
	"fmt"
	"sort"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/osmoutils"
)

// mustGetSpotPrice returns the spot price for the given pool id, and denom0 in terms of denom1.
// Panics if the pool state is misconfigured, which will halt any tx that interacts with this.
func MustGetSpotPrice(k AmmInterface, ctx sdk.Context, poolId uint64, baseAssetDenom string, quoteAssetDenom string) sdk.Dec {
	sp, err := k.CalculateSpotPrice(ctx, poolId, baseAssetDenom, quoteAssetDenom)
	if err != nil {
		panic(err)
	}
	return sp
}

// GetAllUniqueDenomPairs returns all unique pairs of denoms, where for every pair
// (X, Y), X >= Y.
// The pair (X,Y) should only appear once in the list
//
// NOTE: Sorts the input denoms slice.
func GetAllUniqueDenomPairs(denoms []string) ([]string, []string) {
	// get denoms in descending order
	sort.Strings(denoms)
	reverseDenoms := osmoutils.ReverseSlice(denoms)

	numPairs := len(denoms) * (len(denoms) - 1) / 2
	pairGT := make([]string, 0, numPairs)
	pairLT := make([]string, 0, numPairs)

	for i := 0; i < len(reverseDenoms); i++ {
		for j := i + 1; j < len(reverseDenoms); j++ {
			pairGT = append(pairGT, reverseDenoms[i])
			pairLT = append(pairLT, reverseDenoms[j])
		}
	}
	// sanity check
	for i := 0; i < numPairs; i++ {
		if pairGT[i] == pairLT[i] {
			panic("input had duplicated denom")
		}
	}
	return pairGT, pairLT
}

func SpotPriceTimesDuration(sp sdk.Dec, timeDelta time.Duration) sdk.Dec {
	return sp.MulInt64(int64(timeDelta))
}

func AccumDiffDivDuration(accumDiff sdk.Dec, timeDelta time.Duration) sdk.Dec {
	return accumDiff.QuoInt64(int64(timeDelta))
}

// LexicographicalOrderDenoms takes two denoms and returns them to be in lexicographically ascending order.
// In other words, the first returned denom string will be the lexicographically smaller of the two denoms.
// If the denoms are equal, an error will be returned.
func LexicographicalOrderDenoms(denom0, denom1 string) (string, string, error) {
	if denom0 >= denom1 {
		if denom0 == denom1 {
			return "", "", fmt.Errorf("both assets cannot be of the same denom: assetA: %s, assetB: %s", denom0, denom1)
		} else {
			denom0, denom1 = denom1, denom0
		}
	}
	return denom0, denom1, nil
}
