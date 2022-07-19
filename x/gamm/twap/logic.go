package twap

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap/types"
)

func (k twapkeeper) afterCreatePool(ctx sdk.Context, poolId uint64) error {
	denoms, err := k.gammkeeper.GetPoolDenoms(ctx, poolId)
	denomPairs0, denomPairs1 := types.GetAllUniqueDenomPairs(denoms)
	// for every denom pair do create twap
	_, _ = denomPairs0, denomPairs1
	return err
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

func (k twapkeeper) endBlockLogic(ctx sdk.Context) {
	// TODO: Update TWAP entries
	// Step 1: Get all altered pool ids
	changedPoolIds := k.getChangedPools(ctx)
	if len(changedPoolIds) == 0 {
		return
	}
	// Step 2:
}
