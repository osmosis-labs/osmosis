package types

import "sort"

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
