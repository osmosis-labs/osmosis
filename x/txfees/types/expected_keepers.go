package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// this line is used by starport scaffolding # genesis/types/import
// this line is used by starport scaffolding # ibc/genesistype/import

// SpotPriceCalculator defines the contract that must be fulfilled by a spot price calculator
// The x/gamm keeper is expected to satisfy this interface
type SpotPriceCalculator interface {
	CalculateSpotPrice(ctx sdk.Context, poolId uint64, tokenInDenom, tokenOutDenom string) (sdk.Dec, error)
}
