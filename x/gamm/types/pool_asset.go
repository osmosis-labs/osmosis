package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validates a pool asset, to check if it has a valid weight.
func (asset PoolAsset) ValidateWeight() error {
	if asset.Weight.LTE(sdk.ZeroInt()) {
		return fmt.Errorf("a token's weight in the pool must be greater than 0")
	}

	// TODO: Choose a value that is too large for weights
	// if asset.Weight >= (1 << 32) {
	// 	return fmt.Errorf("a token's weight in the pool must be less than 2^32")
	// }

	return nil
}

// subPoolAssetWeights subtracts the weights of two different pool asset slices.
// It assumes that both pool assets have the same token denominations,
// with the denominations in the same order.
// Returned weights can (and probably will have some) be negative.
func subPoolAssetWeights(base []PoolAsset, other []PoolAsset) []PoolAsset {
	weightDifference := make([]PoolAsset, len(base))
	// TODO: Consider deleting these panics for performance
	if len(base) != len(other) {
		panic("subPoolAssetWeights called with invalid input, len(base) != len(other)")
	}
	for i, asset := range base {
		if asset.Token.Denom != other[i].Token.Denom {
			panic(fmt.Sprintf("subPoolAssetWeights called with invalid input, "+
				"expected other's %vth asset to be %v, got %v",
				i, asset.Token.Denom, other[i].Token.Denom))
		}
		curWeightDiff := asset.Weight.Sub(other[i].Weight)
		weightDifference[i] = PoolAsset{Token: asset.Token, Weight: curWeightDiff}
	}
	return weightDifference
}

// addPoolAssetWeights adds the weights of two different pool asset slices.
// It assumes that both pool assets have the same token denominations,
// with the denominations in the same order.
// Returned weights can be negative.
func addPoolAssetWeights(base []PoolAsset, other []PoolAsset) []PoolAsset {
	weightSum := make([]PoolAsset, len(base))
	// TODO: Consider deleting these panics for performance
	if len(base) != len(other) {
		panic("addPoolAssetWeights called with invalid input, len(base) != len(other)")
	}
	for i, asset := range base {
		if asset.Token.Denom != other[i].Token.Denom {
			panic(fmt.Sprintf("addPoolAssetWeights called with invalid input, "+
				"expected other's %vth asset to be %v, got %v",
				i, asset.Token.Denom, other[i].Token.Denom))
		}
		curWeightSum := asset.Weight.Add(other[i].Weight)
		weightSum[i] = PoolAsset{Token: asset.Token, Weight: curWeightSum}
	}
	return weightSum
}

// assumes 0 < d < 1
func poolAssetsMulDec(base []PoolAsset, d sdk.Dec) []PoolAsset {
	newWeights := make([]PoolAsset, len(base))
	for i, asset := range base {
		// TODO: This can adversarially panic at the moment! (as can PoolAccount.TotalWeight)
		// Ensure this won't be able to panic in the future PR where we bound
		// each assets weight, and add precision
		newWeight := d.MulInt(asset.Weight).RoundInt()
		newWeights[i] = PoolAsset{Token: asset.Token, Weight: newWeight}
	}
	return newWeights
}
