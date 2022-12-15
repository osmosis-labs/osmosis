package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GammKeeper defines the expected interface needed for swaprouter module
type GammKeeper interface {
	SwapI
	// TODO: Migrate params in subsequent PR
	GetPoolCreationFee(ctx sdk.Context) sdk.Coins

	GetNextPoolId(ctx sdk.Context) uint64
}
