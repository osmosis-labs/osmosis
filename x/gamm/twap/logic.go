package twap

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k twapkeeper) endBlockLogic(ctx sdk.Context) {
	// TODO: Update TWAP entries
	// Step 1: Get all altered pool ids
	changedPoolIds := k.getChangedPools(ctx)
	if len(changedPoolIds) == 0 {
		return
	}
	// Step 2:
}

func (k twapkeeper) updateTWAPs(ctx sdk.Context, poolId uint64) error {
	twaps, err := k.getAllMostRecentTWAPsForPool(ctx, poolId)
	if err != nil {
		return err
	}
	for _, twap := range twaps {
		// TODO: Update logic
		_ = twap
	}
	return errors.New("Not yet implemented")
}
