package types

import (
	fmt "fmt"

	"github.com/osmosis-labs/osmosis/osmomath"
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
	Amount0   osmomath.Int
	Amount1   osmomath.Int
	Liquidity osmomath.Dec
}

// UpdatePositionData represents the return data from updating a position.
// Tick flags are used to signal if tick is not referenced by any liquidity after the update
// for removal purposes.
type UpdatePositionData struct {
	Amount0          osmomath.Int
	Amount1          osmomath.Int
	LowerTickIsEmpty bool
	UpperTickIsEmpty bool
}
