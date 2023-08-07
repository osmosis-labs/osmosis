package types

import (
	fmt "fmt"
)

// GetConcentratedLockupDenomFromPoolId returns the concentrated lockup denom for a given pool id.
func GetConcentratedLockupDenomFromPoolId(poolId uint64) string {
	return fmt.Sprintf("%s/%d", ConcentratedLiquidityTokenPrefix, poolId)
}
