package twap

import sdk "github.com/cosmos/cosmos-sdk/types"

func (k twapkeeper) endBlockLogic(ctx sdk.Context) {
	// TODO: Update TWAP entries
	// step 1: Get all altered pool ids
	// 'altered pool ids' should be automatically cleared
}
