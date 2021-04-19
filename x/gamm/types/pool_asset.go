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
