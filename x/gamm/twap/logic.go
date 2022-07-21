package twap

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap/types"
)

func (k Keeper) afterCreatePool(ctx sdk.Context, poolId uint64) error {
	denoms, err := k.ammkeeper.GetPoolDenoms(ctx, poolId)
	denomPairs0, denomPairs1 := types.GetAllUniqueDenomPairs(denoms)
	for i := 0; i < len(denomPairs0); i++ {
		record := types.NewTwapRecord(k.ammkeeper, ctx, poolId, denomPairs0[i], denomPairs1[i])
		k.storeMostRecentTWAP(ctx, record)
		k.storeHistoricalTWAP(ctx, record)
	}
	return err
}

func (k Keeper) updateTwapIfNotRedundant(ctx sdk.Context, poolId uint64) error {
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

func (k Keeper) updateTWAPs(ctx sdk.Context, poolId uint64) error {
	// Will only err if pool doesn't have most recent entry set
	twaps, err := k.getAllMostRecentTWAPsForPool(ctx, poolId)
	if err != nil {
		return err
	}

	for _, record := range twaps {
		timeDelta := ctx.BlockTime().Sub(record.Time)

		// no update if were in the same block.
		// should be caught earlier, but secondary check.
		if int(timeDelta) <= 0 {
			return nil
		}

		record.Height = ctx.BlockHeight()
		record.Time = ctx.BlockTime()

		// TODO: Ensure order is correct
		sp0 := types.MustGetSpotPrice(k.ammkeeper, ctx, poolId, record.Asset0Denom, record.Asset1Denom)
		sp1 := types.MustGetSpotPrice(k.ammkeeper, ctx, poolId, record.Asset1Denom, record.Asset0Denom)

		// TODO: Think about overflow
		record.P0ArithmeticTwapAccumulator.AddMut(types.SpotPriceTimesDuration(sp0, timeDelta))
		record.P1ArithmeticTwapAccumulator.AddMut(types.SpotPriceTimesDuration(sp1, timeDelta))
		k.storeMostRecentTWAP(ctx, record)
	}
	return nil
}

func (k Keeper) endBlock(ctx sdk.Context) {
	// TODO: Update all LastSpotPrice's
}

func (k Keeper) pruneOldTwaps(ctx sdk.Context) {
	// TODO: Read this from parameter
	lastAllowedTime := ctx.BlockTime().Add(-48 * time.Hour)
	k.pruneRecordsBeforeTime(ctx, lastAllowedTime)
}

func (k Keeper) getStartRecord(ctx sdk.Context, poolId uint64, time time.Time, assetA string, assetB string) (types.TwapRecord, error) {
	if !(assetA > assetB) {
		assetA, assetB = assetB, assetA
	}
	record, err := k.getRecordAtOrBeforeTime(ctx, poolId, time, assetA, assetB)
	if err != nil {
		return types.TwapRecord{}, err
	}
	record = interpolateRecord(record, time)
	return record, nil
}

func (k Keeper) getMostRecentRecord(ctx sdk.Context, poolId uint64, assetA string, assetB string) (types.TwapRecord, error) {
	if !(assetA > assetB) {
		assetA, assetB = assetB, assetA
	}
	record, err := k.getMostRecentRecordStoreRepresentation(ctx, poolId, assetA, assetB)
	if err != nil {
		return types.TwapRecord{}, err
	}
	record = interpolateRecord(record, ctx.BlockTime())
	return record, nil
}

// pre-condition: interpolateTime >= record.Time
func interpolateRecord(record types.TwapRecord, interpolateTime time.Time) types.TwapRecord {
	if record.Time.Equal(interpolateTime) {
		return record
	}
	interpolatedRecord := record
	timeDelta := interpolateTime.Sub(record.Time)
	interpolatedRecord.Time = interpolateTime

	// TODO: Were currently using the wrong LastSpotPrice, we need to get it from EndBlock for changed pools.

	interpolatedRecord.P0ArithmeticTwapAccumulator = interpolatedRecord.P0ArithmeticTwapAccumulator.Add(
		types.SpotPriceTimesDuration(record.P0LastSpotPrice, timeDelta))
	interpolatedRecord.P1ArithmeticTwapAccumulator = interpolatedRecord.P1ArithmeticTwapAccumulator.Add(
		types.SpotPriceTimesDuration(record.P1LastSpotPrice, timeDelta))

	return interpolatedRecord
}

// For now just assuming p0 price, TODO switch between the two
// precondition: endRecord.Time > startRecord.Time
func (k Keeper) getArithmeticTwap(startRecord types.TwapRecord, endRecord types.TwapRecord) sdk.Dec {
	timeDelta := endRecord.Time.Sub(startRecord.Time)
	accumDiff := endRecord.P0ArithmeticTwapAccumulator.Sub(startRecord.P0ArithmeticTwapAccumulator)
	return types.AccumDiffDivDuration(accumDiff, timeDelta)
}
