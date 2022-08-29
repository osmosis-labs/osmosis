package twap

import (
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/x/twap/types"
)

func newTwapRecord(k types.AmmInterface, ctx sdk.Context, poolId uint64, denom0, denom1 string) (types.TwapRecord, error) {
	denom0, denom1, err := types.LexicographicalOrderDenoms(denom0, denom1)
	if err != nil {
		return types.TwapRecord{}, err
	}
	lastErrorTime := time.Time{}
	sp0, sp1, lastErrorTime := getSpotPrices(ctx, k, poolId, denom0, denom1, lastErrorTime)
	return types.TwapRecord{
		PoolId:                      poolId,
		Asset0Denom:                 denom0,
		Asset1Denom:                 denom1,
		Height:                      ctx.BlockHeight(),
		Time:                        ctx.BlockTime(),
		P0LastSpotPrice:             sp0,
		P1LastSpotPrice:             sp1,
		P0ArithmeticTwapAccumulator: sdk.ZeroDec(),
		P1ArithmeticTwapAccumulator: sdk.ZeroDec(),
		LastErrorTime:               lastErrorTime,
	}, nil
}

// getSpotPricecs gets the spot prices for the pool,
func getSpotPrices(ctx sdk.Context, k types.AmmInterface, poolId uint64, denom0, denom1 string, lastErrorTime time.Time) (
	sp0 sdk.Dec, sp1 sdk.Dec, lastErrTime time.Time) {
	lastErrTime = lastErrorTime
	sp0, err0 := k.CalculateSpotPrice(ctx, poolId, denom0, denom1)
	sp1, err1 := k.CalculateSpotPrice(ctx, poolId, denom1, denom0)
	if err0 != nil || err1 != nil {
		lastErrTime = ctx.BlockTime()
		// In the event of an error, we just sanity replace empty values with zero values
		// so that the numbers can be still be calculated within TWAPs over error values
		if (sp0 == sdk.Dec{}) {
			sp0 = sdk.ZeroDec()
		}
		if (sp1 == sdk.Dec{}) {
			sp1 = sdk.ZeroDec()
		}
	}
	return sp0, sp1, lastErrTime
}

// afterCreatePool creates new twap records of all the unique pairs of denoms within a pool.
func (k Keeper) afterCreatePool(ctx sdk.Context, poolId uint64) error {
	denoms, err := k.ammkeeper.GetPoolDenoms(ctx, poolId)
	denomPairs0, denomPairs1 := types.GetAllUniqueDenomPairs(denoms)
	for i := 0; i < len(denomPairs0); i++ {
		record, err := newTwapRecord(k.ammkeeper, ctx, poolId, denomPairs0[i], denomPairs1[i])
		// err should be impossible given GetAllUniqueDenomPairs guarantees
		if err != nil {
			return err
		}
		// we create a record here, because we need the record to exist in the event
		// that there is a swap against this pool in this same block.
		// furthermore, this protects against an edge case where a pool is created
		// during EndBlock, after twapkeeper's endblock.
		k.storeNewRecord(ctx, record)
	}
	k.trackChangedPool(ctx, poolId)
	return err
}

func (k Keeper) EndBlock(ctx sdk.Context) {
	// get changed pools grabs all altered pool ids from the transient store.
	// 'altered pool ids' gets automatically cleared on commit by being a transient store
	changedPoolIds := k.getChangedPools(ctx)
	for _, id := range changedPoolIds {
		err := k.updateRecords(ctx, id)
		if err != nil {
			panic(err)
		}
	}
}

func (k Keeper) updateRecords(ctx sdk.Context, poolId uint64) error {
	// Will only err if pool doesn't have most recent entry set
	records, err := k.getAllMostRecentRecordsForPool(ctx, poolId)
	if err != nil {
		return err
	}

	for _, record := range records {
		newRecord := k.updateRecord(ctx, record)
		k.storeNewRecord(ctx, newRecord)
	}
	return nil
}

// updateRecord returns a new record with updated accumulators and block time
// for the current block time.
func (k Keeper) updateRecord(ctx sdk.Context, record types.TwapRecord) types.TwapRecord {
	newRecord := recordWithUpdatedAccumulators(record, ctx.BlockTime())
	newRecord.Height = ctx.BlockHeight()

	newSp0, newSp1, lastErrorTime := getSpotPrices(
		ctx, k.ammkeeper, record.PoolId, record.Asset0Denom, record.Asset1Denom, record.LastErrorTime)

	// set last spot price to be last price of this block. This is what will get used in interpolation.
	newRecord.P0LastSpotPrice = newSp0
	newRecord.P1LastSpotPrice = newSp1
	newRecord.LastErrorTime = lastErrorTime

	return newRecord
}

// pruneRecords prunes twap records that happened earlier than recordHistoryKeepPeriod
// before current block time.
func (k Keeper) pruneRecords(ctx sdk.Context) error {
	recordHistoryKeepPeriod := k.RecordHistoryKeepPeriod(ctx)

	lastKeptTime := ctx.BlockTime().Add(-recordHistoryKeepPeriod)
	return k.pruneRecordsBeforeTime(ctx, lastKeptTime)
}

// recordWithUpdatedAccumulators returns a record, with updated accumulator values and time for provided newTime.
// otherwise referred to as "interpolating the record" to the target time.
//
// pre-condition: newTime >= record.Time
func recordWithUpdatedAccumulators(record types.TwapRecord, newTime time.Time) types.TwapRecord {
	if record.Time.Equal(newTime) {
		return record
	}
	newRecord := record
	timeDelta := newTime.Sub(record.Time)
	newRecord.Time = newTime

	// record.LastSpotPrice is the last spot price from the block the record was created in,
	// thus it is treated as the effective spot price until the new time.
	// (As there was no change until at or after this time)
	// TODO: Think about overflow
	newRecord.P0ArithmeticTwapAccumulator = newRecord.P0ArithmeticTwapAccumulator.Add(
		types.SpotPriceTimesDuration(record.P0LastSpotPrice, timeDelta))
	newRecord.P1ArithmeticTwapAccumulator = newRecord.P1ArithmeticTwapAccumulator.Add(
		types.SpotPriceTimesDuration(record.P1LastSpotPrice, timeDelta))
	return newRecord
}

// getInterpolatedRecord returns a record for this pool, representing its accumulator state at time `t`.
// This is achieved by getting the record `r` that is at, or immediately preceding in state time `t`.
// To be clear: the record r s.t. `t - r.Time` is minimized AND `t >= r.Time`
func (k Keeper) getInterpolatedRecord(ctx sdk.Context, poolId uint64, t time.Time, assetA, assetB string) (types.TwapRecord, error) {
	assetA, assetB, err := types.LexicographicalOrderDenoms(assetA, assetB)
	if err != nil {
		return types.TwapRecord{}, err
	}
	record, err := k.getRecordAtOrBeforeTime(ctx, poolId, t, assetA, assetB)
	if err != nil {
		return types.TwapRecord{}, err
	}
	record = recordWithUpdatedAccumulators(record, t)
	return record, nil
}

func (k Keeper) getMostRecentRecord(ctx sdk.Context, poolId uint64, assetA, assetB string) (types.TwapRecord, error) {
	assetA, assetB, err := types.LexicographicalOrderDenoms(assetA, assetB)
	if err != nil {
		return types.TwapRecord{}, err
	}
	record, err := k.getMostRecentRecordStoreRepresentation(ctx, poolId, assetA, assetB)
	if err != nil {
		return types.TwapRecord{}, err
	}
	record = recordWithUpdatedAccumulators(record, ctx.BlockTime())
	return record, nil
}

// computeArithmeticTwap computes and returns an arithmetic TWAP between
// two records given the quote asset.
// precondition: endRecord.Time >= startRecord.Time
// if endRecord.LastErrorTime is after startRecord.Time, return an error at end + result
// if (endRecord.Time == startRecord.Time) returns endRecord.LastSpotPrice
// else returns
// (endRecord.Accumulator - startRecord.Accumulator) / (endRecord.Time - startRecord.Time)
func computeArithmeticTwap(startRecord types.TwapRecord, endRecord types.TwapRecord, quoteAsset string) (sdk.Dec, error) {
	// see if we need to return an error, due to spot price issues
	var err error = nil
	if endRecord.LastErrorTime.After(startRecord.Time) || endRecord.LastErrorTime.Equal(startRecord.Time) {
		err = errors.New("twap: error in pool spot price occurred between start and end time, twap result may be faulty")
	}
	timeDelta := endRecord.Time.Sub(startRecord.Time)
	// if time difference is 0, then return the last spot price based off of start.
	if timeDelta == time.Duration(0) {
		if quoteAsset == startRecord.Asset0Denom {
			return endRecord.P0LastSpotPrice, err
		}
		return endRecord.P1LastSpotPrice, err
	}
	var accumDiff sdk.Dec
	if quoteAsset == startRecord.Asset0Denom {
		accumDiff = endRecord.P0ArithmeticTwapAccumulator.Sub(startRecord.P0ArithmeticTwapAccumulator)
	} else {
		accumDiff = endRecord.P1ArithmeticTwapAccumulator.Sub(startRecord.P1ArithmeticTwapAccumulator)
	}
	return types.AccumDiffDivDuration(accumDiff, timeDelta), err
}
