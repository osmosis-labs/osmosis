package twap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap/types"
)

func (k twapkeeper) afterCreatePool(ctx sdk.Context, poolId uint64) error {
	denoms, err := k.gammkeeper.GetPoolDenoms(ctx, poolId)
	denomPairs0, denomPairs1 := types.GetAllUniqueDenomPairs(denoms)
	for i := 0; i < len(denomPairs0); i++ {
		record := types.NewTwapRecord(ctx, poolId, denomPairs0[i], denomPairs1[i])
		k.storeMostRecentTWAP(ctx, record)
	}
	return err
}

func (k twapkeeper) updateTwapIfNotRedundant(ctx sdk.Context, poolId uint64) error {
	if k.hasPoolChangedThisBlock(ctx, poolId) {
		return nil
	}
	err := k.updateTWAPs(ctx, poolId)
	if err != nil {
		return err
	}
	k.trackChangedPool(ctx, poolId)
	return nil
}

func (k twapkeeper) updateTWAPs(ctx sdk.Context, poolId uint64) error {
	// Will only err if pool doesn't have most recent entry set
	twaps, err := k.getAllMostRecentTWAPsForPool(ctx, poolId)
	if err != nil {
		return err
	}

	for _, record := range twaps {
		k.storeHistoricalTWAP(ctx, record)
		timeDelta := ctx.BlockTime().Sub(record.Time)

		// no update if were in the same block.
		// should be caught earlier, but secondary check.
		if int(timeDelta) <= 0 {
			return nil
		}

		record.Height = ctx.BlockHeight()
		record.Time = ctx.BlockTime()

		// TODO: Think about order
		sp0, err := k.gammkeeper.GetSpotPrice(ctx, poolId, record.Asset0Denom, record.Asset1Denom)
		// TODO: Document in what situations it can error
		if err != nil {
			return err
		}
		sp1, err := k.gammkeeper.GetSpotPrice(ctx, poolId, record.Asset0Denom, record.Asset1Denom)
		if err != nil {
			return err
		}

		// TODO: Think about overflow
		record.P0ArithmeticTwapAccumulator.AddMut(sp0.MulInt64(int64(timeDelta)))
		record.P1ArithmeticTwapAccumulator.AddMut(sp1.MulInt64(int64(timeDelta)))
		k.storeMostRecentTWAP(ctx, record)
	}
	return nil
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
