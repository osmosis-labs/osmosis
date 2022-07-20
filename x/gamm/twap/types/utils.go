package types

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewTwapRecord(ctx sdk.Context, poolId uint64, denom0 string, denom1 string) TwapRecord {
	if !(denom0 > denom1) {
		panic("precondition denom0 > denom1 not satisfied")
	}
	return TwapRecord{
		PoolId:                      poolId,
		Asset0Denom:                 denom0,
		Asset1Denom:                 denom1,
		Height:                      ctx.BlockHeight(),
		Time:                        ctx.BlockTime(),
		P0ArithmeticTwapAccumulator: sdk.ZeroDec(),
		P1ArithmeticTwapAccumulator: sdk.ZeroDec(),
	}
}

// GetAllUniqueDenomPairs returns all unique pairs of denoms, where for every pair
// (X, Y), X >= Y.
// The pair (X,Y) should only appear once in the list
//
// NOTE: Sorts the input denoms slice.
// (Should not be a problem, as this should come from coins.Denoms(), which returns a sorted order)
func GetAllUniqueDenomPairs(denoms []string) ([]string, []string) {
	sort.Strings(denoms)
	numPairs := len(denoms) * (len(denoms) - 1) / 2
	pairGT := make([]string, 0, numPairs)
	pairLT := make([]string, 0, numPairs)
	for i := 0; i < len(denoms); i++ {
		for j := i + 1; j < len(denoms); j++ {
			pairGT = append(pairGT, denoms[i])
			pairLT = append(pairLT, denoms[j])
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
