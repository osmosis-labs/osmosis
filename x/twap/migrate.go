package twap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MigrateExistingPools iterates through all pools and creates state entry for the twap module.
func (k Keeper) MigrateExistingPools(ctx sdk.Context, latestPoolId uint64) error {
	for i := uint64(1); i <= latestPoolId; i++ {
		err := k.afterCreatePool(ctx, i)
		if err != nil {
			return err
		}
	}
	return nil
}
