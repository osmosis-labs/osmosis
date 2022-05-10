package stableswap

import (
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	errMsgFmtNonPositiveTokenAmount   = "token amount for denom %s must be positive, was %d"
	errMsgFmtNonPositiveScalingFactor = "scaling factor for denom %s must be positive, was %d"
)

// Validates a pool asset, returns nil if valid, error otherwise.
func (asset PoolAsset) Validate() error {
	if asset.Token.Amount.LTE(sdk.ZeroInt()) {
		return fmt.Errorf(errMsgFmtNonPositiveTokenAmount, asset.Token.Denom, asset.Token.Amount.Int64())
	}

	if asset.ScalingFactor.LTE(sdk.ZeroInt()) {
		return fmt.Errorf(errMsgFmtNonPositiveScalingFactor, asset.Token.Denom, asset.ScalingFactor.Int64())
	}
	return nil
}

// SortPoolAssetsByDenom sorts pool assets in place, by denom.
func SortPoolAssetsByDenom(assets []PoolAsset) {
	sort.Slice(assets, func(i, j int) bool {
		PoolAssetA := assets[i]
		PoolAssetB := assets[j]

		return strings.Compare(PoolAssetA.Token.Denom, PoolAssetB.Token.Denom) == -1
	})
}
