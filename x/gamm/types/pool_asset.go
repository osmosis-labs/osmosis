package types

import (
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func ValidateUserSpecifiedWeight(weight sdk.Int) error {
	if !weight.IsPositive() {
		return sdkerrors.Wrap(ErrNotPositiveWeight, weight.String())
	}

	if weight.GTE(MaxUserSpecifiedWeight) {
		return sdkerrors.Wrap(ErrWeightTooLarge, weight.String())
	}
	return nil
}

func ValidateUserSpecifiedPoolAssets(assets []BalancerPoolAsset) error {
	// The pool must be swapping between at least two assets
	if len(assets) < 2 {
		return ErrTooFewPoolAssets
	}

	// TODO: Add the limit of binding token to the pool params?
	if len(assets) > 8 {
		return sdkerrors.Wrapf(ErrTooManyPoolAssets, "%d", len(assets))
	}

	for _, asset := range assets {
		err := ValidateUserSpecifiedWeight(asset.Weight)
		if err != nil {
			return err
		}

		if !asset.Token.IsValid() || !asset.Token.IsPositive() {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, asset.Token.String())
		}
	}
	return nil
}

// SortPoolAssetsOutOfPlaceByDenom sorts pool assets in place, by weight
// Doesn't deep copy the underlying weights, but it does place the assets
// into a new slice.
func SortPoolAssetsOutOfPlaceByDenom(assets []BalancerPoolAsset) []BalancerPoolAsset {
	assets_copy := make([]BalancerPoolAsset, len(assets))
	copy(assets_copy, assets)
	SortPoolAssetsByDenom(assets_copy)
	return assets_copy
}

// SortPoolAssetsByDenom sorts pool assets in place, by weight
func SortPoolAssetsByDenom(assets []BalancerPoolAsset) {
	sort.Slice(assets, func(i, j int) bool {
		PoolAssetA := assets[i]
		PoolAssetB := assets[j]

		return strings.Compare(PoolAssetA.Token.Denom, PoolAssetB.Token.Denom) == -1
	})
}

// Validates a pool asset, to check if it has a valid weight.
func (asset BalancerPoolAsset) ValidateWeight() error {
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
func subPoolAssetWeights(base []BalancerPoolAsset, other []BalancerPoolAsset) []BalancerPoolAsset {
	weightDifference := make([]BalancerPoolAsset, len(base))
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
		weightDifference[i] = BalancerPoolAsset{Token: asset.Token, Weight: curWeightDiff}
	}
	return weightDifference
}

// addPoolAssetWeights adds the weights of two different pool asset slices.
// It assumes that both pool assets have the same token denominations,
// with the denominations in the same order.
// Returned weights can be negative.
func addPoolAssetWeights(base []BalancerPoolAsset, other []BalancerPoolAsset) []BalancerPoolAsset {
	weightSum := make([]BalancerPoolAsset, len(base))
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
		weightSum[i] = BalancerPoolAsset{Token: asset.Token, Weight: curWeightSum}
	}
	return weightSum
}

// assumes 0 < d < 1
func poolAssetsMulDec(base []BalancerPoolAsset, d sdk.Dec) []BalancerPoolAsset {
	newWeights := make([]BalancerPoolAsset, len(base))
	for i, asset := range base {
		// TODO: This can adversarially panic at the moment! (as can Pool.TotalWeight)
		// Ensure this won't be able to panic in the future PR where we bound
		// each assets weight, and add precision
		newWeight := d.MulInt(asset.Weight).RoundInt()
		newWeights[i] = BalancerPoolAsset{Token: asset.Token, Weight: newWeight}
	}
	return newWeights
}

// PoolAssetsCoins returns all the coins corresponding to a slice of pool assets
func PoolAssetsCoins(assets []BalancerPoolAsset) sdk.Coins {
	coins := sdk.Coins{}
	for _, asset := range assets {
		coins = coins.Add(asset.Token)
	}
	return coins
}
