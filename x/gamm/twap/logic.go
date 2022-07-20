package twap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap/types"
)

func (k twapkeeper) afterCreatePool(ctx sdk.Context, poolId uint64) error {
	denoms, err := k.gammkeeper.GetPoolDenoms(ctx, poolId)
	denomPairs0, denomPairs1 := types.GetAllUniqueDenomPairs(denoms)
	for i := 0; i < len(denomPairs0); i++ {
		record := types.NewTwapRecord(k.gammkeeper, ctx, poolId, denomPairs0[i], denomPairs1[i])
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

		// TODO: Ensure order is correct
		sp0 := types.MustGetSpotPrice(k.gammkeeper, ctx, poolId, record.Asset0Denom, record.Asset1Denom)
		sp1 := types.MustGetSpotPrice(k.gammkeeper, ctx, poolId, record.Asset1Denom, record.Asset0Denom)

		// TODO: Think about overflow
		record.P0ArithmeticTwapAccumulator.AddMut(sp0.MulInt64(int64(timeDelta)))
		record.P1ArithmeticTwapAccumulator.AddMut(sp1.MulInt64(int64(timeDelta)))
		k.storeMostRecentTWAP(ctx, record)
	}
	return nil
}
