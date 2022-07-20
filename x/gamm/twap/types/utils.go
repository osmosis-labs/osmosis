package types

import (
	fmt "fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/osmoutils"
)

func NewTwapRecord(k AmmInterface, ctx sdk.Context, poolId uint64, denom0 string, denom1 string) TwapRecord {
	if !(denom0 > denom1) {
		panic(fmt.Sprintf("precondition denom0 > denom1 not satisfied. denom0 %s | denom1 %s", denom0, denom1))
	}
	sp0 := MustGetSpotPrice(k, ctx, poolId, denom0, denom1)
	sp1 := MustGetSpotPrice(k, ctx, poolId, denom1, denom0)
	return TwapRecord{
		PoolId:                      poolId,
		Asset0Denom:                 denom0,
		Asset1Denom:                 denom1,
		Height:                      ctx.BlockHeight(),
		Time:                        ctx.BlockTime(),
		P0LastSpotPrice:             sp0,
		P1LastSpotPrice:             sp1,
		P0ArithmeticTwapAccumulator: sdk.ZeroDec(),
		P1ArithmeticTwapAccumulator: sdk.ZeroDec(),
	}
}

// mustGetSpotPrice returns the spot price for the given pool id, and denom0 in terms of denom1.
// Panics if the pool state is misconfigured, which will halt any tx that interacts with this.
func MustGetSpotPrice(k AmmInterface, ctx sdk.Context, poolId uint64, denom0 string, denom1 string) sdk.Dec {
	sp, err := k.CalculateSpotPrice(ctx, poolId, denom0, denom1)
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
// (Should not be a problem, as this should come from coins.Denoms(), which returns a sorted order)
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

// TODO
func (g *GenesisState) Validate() error {
	return nil
}
