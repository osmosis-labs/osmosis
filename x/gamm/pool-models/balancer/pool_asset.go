package balancer

import (
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type poolAssetPretty struct {
	Token  sdk.Coin `json:"token" yaml:"token"`
	Weight sdk.Dec  `json:"weight" yaml:"weight"`
}

// validates a pool asset, to check if it has a valid weight.
func (pa PoolAsset) validateWeight() error {
	if pa.Weight.LTE(sdk.ZeroInt()) {
		return fmt.Errorf("a token's weight in the pool must be greater than 0")
	}

	// TODO: add validation for asset weight overflow:
	// https://github.com/osmosis-labs/osmosis/issues/1958

	return nil
}

func (pa PoolAsset) prettify() poolAssetPretty {
	return poolAssetPretty{
		Weight: sdk.NewDecFromInt(pa.Weight).QuoInt64(GuaranteedWeightPrecision),
		Token:  pa.Token,
	}
}

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

// sortPoolAssetsOutOfPlaceByDenom sorts pool assets in place, by weight
// Doesn't deep copy the underlying weights, but it does place the assets
// into a new slice.
func sortPoolAssetsOutOfPlaceByDenom(assets []PoolAsset) []PoolAsset {
	assets_copy := make([]PoolAsset, len(assets))
	copy(assets_copy, assets)
	sortPoolAssetsByDenom(assets_copy)
	return assets_copy
}

// sortPoolAssetsByDenom sorts pool assets in place, by weight.
func sortPoolAssetsByDenom(assets []PoolAsset) {
	sort.Slice(assets, func(i, j int) bool {
		PoolAssetA := assets[i]
		PoolAssetB := assets[j]

		return strings.Compare(PoolAssetA.Token.Denom, PoolAssetB.Token.Denom) == -1
	})
}

func validateUserSpecifiedPoolAssets(assets []PoolAsset) error {
	// The pool must be swapping between at least two assets
	if len(assets) < types.MinNumOfAssetsInPool {
		return types.ErrTooFewPoolAssets
	}

	if len(assets) > types.MaxNumOfAssetsInPool {
		return errorsmod.Wrapf(types.ErrTooManyPoolAssets, "%d", len(assets))
	}

	assetExistsMap := map[string]bool{}
	for _, asset := range assets {
		err := ValidateUserSpecifiedWeight(asset.Weight)
		if err != nil {
			return err
		}

		if !asset.Token.IsValid() || !asset.Token.IsPositive() {
			return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, asset.Token.String())
		}
		if _, exists := assetExistsMap[asset.Token.Denom]; exists {
			return errorsmod.Wrapf(types.ErrTooFewPoolAssets, "pool asset %s already exists", asset.Token.Denom)
		}
		assetExistsMap[asset.Token.Denom] = true
	}
	return nil
}

// poolAssetsCoins returns all the coins corresponding to a slice of pool assets.
func poolAssetsCoins(assets []PoolAsset) sdk.Coins {
	coins := sdk.Coins{}
	for _, asset := range assets {
		coins = coins.Add(asset.Token)
	}
	return coins
}

func getPoolAssetByDenom(assets []PoolAsset, denom string) (PoolAsset, bool) {
	for _, asset := range assets {
		if asset.Token.Denom == denom {
			return asset, true
		}
	}
	return PoolAsset{}, false
}
