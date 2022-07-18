package twap

import sdk "github.com/cosmos/cosmos-sdk/types"

func (k twapkeeper) endBlockLogic(ctx sdk.Context) {
	// TODO: Update TWAP entries
	// Step 1: Get all altered pool ids
	changedPoolIds := k.getChangedPools(ctx)
	if len(changedPoolIds) == 0 {
		return
	}
	// Step 2:
}
