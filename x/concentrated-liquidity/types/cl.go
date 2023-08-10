package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetConcentratedLockupDenomFromPoolId returns the concentrated lockup denom for a given pool id.
func GetConcentratedLockupDenomFromPoolId(poolId uint64) string {
	return fmt.Sprintf("%s/%d", ConcentratedLiquidityTokenPrefix, poolId)
}

// CreateFullRangePositionData represents the return data from any method
// that creates a full range position. We have multipl variants to
// account for varying locking scenarios.
type CreateFullRangePositionData struct {
	ID        uint64
	Amount0   sdk.Int
	Amount1   sdk.Int
	Liquidity sdk.Dec
}
