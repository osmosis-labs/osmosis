package twap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MigrateExistingPools iterates through all pools and creates state entry for the twap module.
func (k Keeper) MigrateExistingPools(ctx sdk.Context, latestPoolId uint64) error {
	for i := 1; i <= int(latestPoolId); i++ {
		err := k.afterCreatePool(ctx, uint64(i))
		if err != nil {
			return err
		}
	}
	return nil
}
