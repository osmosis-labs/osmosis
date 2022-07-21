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
		k.storeNewRecord(ctx, record)
	}
	return err
}

func (k Keeper) updateTWAPs(ctx sdk.Context, poolId uint64) error {
	// Will only err if pool doesn't have most recent entry set
	records, err := k.getAllMostRecentRecordsForPool(ctx, poolId)
	if err != nil {
		return err
	}

	for _, record := range records {
		timeDelta := ctx.BlockTime().Sub(record.Time)

		// no update if were in the same block.
		// should be caught earlier, but secondary check.
		if int(timeDelta) <= 0 {
			return nil
		}

		record.Height = ctx.BlockHeight()
		record.Time = ctx.BlockTime()

		// TODO: Ensure order is correct
		newSp0 := types.MustGetSpotPrice(k.ammkeeper, ctx, poolId, record.Asset0Denom, record.Asset1Denom)
		newSp1 := types.MustGetSpotPrice(k.ammkeeper, ctx, poolId, record.Asset1Denom, record.Asset0Denom)

		// TODO: Think about overflow
		// Update accumulators based on last block / update's spot price
		record.P0ArithmeticTwapAccumulator.AddMut(types.SpotPriceTimesDuration(record.P0LastSpotPrice, timeDelta))
		record.P1ArithmeticTwapAccumulator.AddMut(types.SpotPriceTimesDuration(record.P1LastSpotPrice, timeDelta))

		// set last spot price to be last price of this block. This is what will get used in interpolation.
		record.P0LastSpotPrice = newSp0
		record.P1LastSpotPrice = newSp1

		k.storeNewRecord(ctx, record)
	}
	return nil
}

func (k Keeper) endBlock(ctx sdk.Context) {
	changedPoolIds := k.getChangedPools(ctx)
	for _, id := range changedPoolIds {
		err := k.updateTWAPs(ctx, id)
		if err != nil {
			panic(err)
		}
	}
	// 'altered pool ids' gets automatically cleared by being a transient store
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

// interpolate record updates the record's accumulator values and time to the interpolate time.
//
// pre-condition: interpolateTime >= record.Time
func interpolateRecord(record types.TwapRecord, interpolateTime time.Time) types.TwapRecord {
	if record.Time.Equal(interpolateTime) {
		return record
	}
	interpolatedRecord := record
	timeDelta := interpolateTime.Sub(record.Time)
	interpolatedRecord.Time = interpolateTime

	// record.LastSpotPrice is the last spot price from the block the record was created in,
	// thus it is treated as the effective spot price at the interpolation time.
	// (As there was no change until the next block began)

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
