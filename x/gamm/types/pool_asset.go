package types

import (
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"gopkg.in/yaml.v2"
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

type poolAssetPretty struct {
	Token  sdk.Coin `json:"token" yaml:"token"`
	Weight sdk.Dec  `json:"weight" yaml:"weight"`
}

func (asset PoolAsset) prettify() poolAssetPretty {
	return poolAssetPretty{
		Weight: sdk.NewDecFromInt(asset.Weight).QuoInt64(GuaranteedWeightPrecision),
		Token:  asset.Token,
	}
}

// D: at name
// func (asset poolAssetPretty) uglify() PoolAsset {
// 	return PoolAsset{
// 		Weight: asset.Weight.MulInt64(GuaranteedWeightPrecision).RoundInt(),
// 		Token:  asset.Token,
// 	}
// }

// MarshalYAML returns the YAML representation of a PoolAsset.
// This is assumed to not be called on a stand-alone instance, so it removes the first marshalled line.
func (pa PoolAsset) MarshalYAML() (interface{}, error) {
	bz, err := yaml.Marshal(pa.prettify())
	if err != nil {
		return nil, err
	}
	s := string(bz)
	return s, nil
}

// SortPoolAssetsOutOfPlaceByDenom sorts pool assets in place, by weight
// Doesn't deep copy the underlying weights, but it does place the assets
// into a new slice.
func SortPoolAssetsOutOfPlaceByDenom(assets []PoolAsset) []PoolAsset {
	assets_copy := make([]PoolAsset, len(assets))
	copy(assets_copy, assets)
	SortPoolAssetsByDenom(assets_copy)
	return assets_copy
}

// SortPoolAssetsByDenom sorts pool assets in place, by weight
func SortPoolAssetsByDenom(assets []PoolAsset) {
	sort.Slice(assets, func(i, j int) bool {
		PoolAssetA := assets[i]
		PoolAssetB := assets[j]

		return strings.Compare(PoolAssetA.Token.Denom, PoolAssetB.Token.Denom) == -1
	})
}

func ValidateUserSpecifiedPoolAssets(assets []PoolAsset) error {
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

// PoolAssetsCoins returns all the coins corresponding to a slice of pool assets
func PoolAssetsCoins(assets []PoolAsset) sdk.Coins {
	coins := sdk.Coins{}
	for _, asset := range assets {
		coins = coins.Add(asset.Token)
	}
	return coins
}
