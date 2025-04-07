package twap_test

import (
	"sort"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/x/twap"
	"github.com/osmosis-labs/osmosis/v27/x/twap/types"
)

// TODO: Consider switching this everywhere
var (
	denom0                        = "token/A"
	denom1                        = "token/B"
	denom2                        = "token/C"
	defaultTwoAssetCoins          = sdk.NewCoins(sdk.NewInt64Coin(denom0, 1_000_000_000), sdk.NewInt64Coin(denom1, 1_000_000_000))
	defaultThreeAssetCoins        = sdk.NewCoins(sdk.NewInt64Coin(denom0, 1_000_000_000), sdk.NewInt64Coin(denom1, 1_000_000_000), sdk.NewInt64Coin(denom2, 1_000_000_000))
	baseTime                      = time.Unix(1257894000, 0).UTC()
	tPlusOne                      = baseTime.Add(time.Second)
	tMinOne                       = baseTime.Add(-time.Second)
	tPlusOneMin                   = baseTime.Add(time.Minute)
	basePoolId             uint64 = 1
	oneHundredNanoseconds         = 100 * time.Nanosecond
)

type TestSuite struct {
	apptesting.KeeperTestHelper
	twapkeeper *twap.Keeper
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupTest() {
	s.Setup()
	s.twapkeeper = s.App.TwapKeeper
	s.Ctx = s.Ctx.WithBlockTime(baseTime)
	// add x/twap test specific denoms
	poolManagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)
	poolManagerParams.AuthorizedQuoteDenoms = append(poolManagerParams.AuthorizedQuoteDenoms, denom0, denom1, denom2)
	s.App.PoolManagerKeeper.SetParams(s.Ctx, poolManagerParams)
}

var (
	basicParams = types.NewParams("week", 48*time.Hour)

	mostRecentRecordPoolOne = types.TwapRecord{
		PoolId:                      basePoolId,
		Asset0Denom:                 denom0,
		Asset1Denom:                 denom1,
		Height:                      3,
		Time:                        tPlusOne.Add(time.Second),
		P0LastSpotPrice:             osmomath.OneDec(),
		P1LastSpotPrice:             osmomath.OneDec(),
		P0ArithmeticTwapAccumulator: osmomath.OneDec(),
		P1ArithmeticTwapAccumulator: osmomath.OneDec(),
		GeometricTwapAccumulator:    osmomath.OneDec(),
	}

	basicCustomGenesis = types.NewGenesisState(
		basicParams,
		[]types.TwapRecord{
			mostRecentRecordPoolOne,
		})

	increasingOrderByTimeRecordsPoolOne = types.NewGenesisState(
		basicParams,
		[]types.TwapRecord{
			{
				PoolId:                      basePoolId,
				Asset0Denom:                 denom0,
				Asset1Denom:                 denom1,
				Height:                      1,
				Time:                        baseTime,
				P0LastSpotPrice:             osmomath.OneDec(),
				P1LastSpotPrice:             osmomath.OneDec(),
				P0ArithmeticTwapAccumulator: osmomath.OneDec(),
				P1ArithmeticTwapAccumulator: osmomath.OneDec(),
				GeometricTwapAccumulator:    osmomath.OneDec(),
			},
			{
				PoolId:                      basePoolId,
				Asset0Denom:                 denom0,
				Asset1Denom:                 denom1,
				Height:                      2,
				Time:                        tPlusOne,
				P0LastSpotPrice:             osmomath.OneDec(),
				P1LastSpotPrice:             osmomath.OneDec(),
				P0ArithmeticTwapAccumulator: osmomath.OneDec(),
				P1ArithmeticTwapAccumulator: osmomath.OneDec(),
				GeometricTwapAccumulator:    osmomath.OneDec(),
			},
			mostRecentRecordPoolOne,
		})

	mostRecentRecordPoolTwo = types.TwapRecord{
		PoolId:                      basePoolId,
		Asset0Denom:                 denom0,
		Asset1Denom:                 denom2,
		Height:                      1,
		Time:                        tPlusOne.Add(time.Second),
		P0LastSpotPrice:             osmomath.OneDec(),
		P1LastSpotPrice:             osmomath.OneDec(),
		P0ArithmeticTwapAccumulator: osmomath.OneDec(),
		P1ArithmeticTwapAccumulator: osmomath.OneDec(),
		GeometricTwapAccumulator:    osmomath.OneDec(),
	}

	decreasingOrderByTimeRecordsPoolTwo = types.NewGenesisState(
		basicParams,
		[]types.TwapRecord{
			mostRecentRecordPoolTwo,
			{
				PoolId:                      basePoolId,
				Asset0Denom:                 denom0,
				Asset1Denom:                 denom2,
				Height:                      2,
				Time:                        tPlusOne,
				P0LastSpotPrice:             osmomath.OneDec(),
				P1LastSpotPrice:             osmomath.OneDec(),
				P0ArithmeticTwapAccumulator: osmomath.OneDec(),
				P1ArithmeticTwapAccumulator: osmomath.OneDec(),
				GeometricTwapAccumulator:    osmomath.OneDec(),
			},
			{
				PoolId:                      basePoolId,
				Asset0Denom:                 denom0,
				Asset1Denom:                 denom2,
				Height:                      3,
				Time:                        baseTime,
				P0LastSpotPrice:             osmomath.OneDec(),
				P1LastSpotPrice:             osmomath.OneDec(),
				P0ArithmeticTwapAccumulator: osmomath.OneDec(),
				P1ArithmeticTwapAccumulator: osmomath.OneDec(),
				GeometricTwapAccumulator:    osmomath.OneDec(),
			},
		})

	bothPoolsGenesis = types.NewGenesisState(
		basicParams,
		append(increasingOrderByTimeRecordsPoolOne.Twaps, decreasingOrderByTimeRecordsPoolTwo.Twaps...),
	)
)

func withPoolId(twap types.TwapRecord, poolId uint64) types.TwapRecord {
	twap.PoolId = poolId
	return twap
}

func withLastErrTime(twap types.TwapRecord, lastErrorTime time.Time) types.TwapRecord {
	twap.LastErrorTime = lastErrorTime
	return twap
}

func withSp0(twap types.TwapRecord, sp osmomath.Dec) types.TwapRecord {
	twap.P0LastSpotPrice = sp
	return twap
}

func withSp1(twap types.TwapRecord, sp osmomath.Dec) types.TwapRecord {
	twap.P1LastSpotPrice = sp
	return twap
}

// TestTWAPInitGenesis tests that genesis is initialized correctly
// with different parameters and state.
// Asserts that the most recent records are set correctly.
func (s *TestSuite) TestTwapInitGenesis() {
	testCases := map[string]struct {
		twapGenesis *types.GenesisState

		expectPanic bool

		expectedMostRecentRecord []types.TwapRecord
	}{
		"default genesis - success": {
			twapGenesis: types.DefaultGenesis(),
		},
		"custom valid genesis; success": {
			twapGenesis: basicCustomGenesis,

			expectedMostRecentRecord: []types.TwapRecord{
				mostRecentRecordPoolOne,
			},
		},
		"custom valid multi record; increasing; success": {
			twapGenesis: increasingOrderByTimeRecordsPoolOne,

			expectedMostRecentRecord: []types.TwapRecord{
				mostRecentRecordPoolOne,
			},
		},
		"custom valid multi record; decreasing (sorted internally); success": {
			twapGenesis: decreasingOrderByTimeRecordsPoolTwo,

			expectedMostRecentRecord: []types.TwapRecord{
				mostRecentRecordPoolTwo,
			},
		},
		"custom valid multi record and multi pool; success": {
			twapGenesis: bothPoolsGenesis,

			expectedMostRecentRecord: []types.TwapRecord{
				mostRecentRecordPoolTwo, mostRecentRecordPoolOne,
			},
		},
		"custom invalid genesis - error": {
			twapGenesis: types.NewGenesisState(
				types.NewParams("week", 48*time.Hour),
				[]types.TwapRecord{
					{
						PoolId:                      0, // invalid
						Asset0Denom:                 "test1",
						Asset1Denom:                 "test2",
						Height:                      1,
						Time:                        baseTime,
						P0LastSpotPrice:             osmomath.OneDec(),
						P1LastSpotPrice:             osmomath.OneDec(),
						P0ArithmeticTwapAccumulator: osmomath.OneDec(),
						P1ArithmeticTwapAccumulator: osmomath.OneDec(),
						GeometricTwapAccumulator:    osmomath.OneDec(),
					},
				}),

			expectPanic: true,
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			s.Setup()
			// Setup.
			ctx := s.Ctx
			twapKeeper := s.App.TwapKeeper

			// Test.
			osmoassert.ConditionalPanic(s.T(), tc.expectPanic, func() { twapKeeper.InitGenesis(ctx, tc.twapGenesis) })
			if tc.expectPanic {
				return
			}

			// Assertions.

			// Parameters were set.
			s.Require().Equal(tc.twapGenesis.Params, twapKeeper.GetParams(ctx))

			for _, expectedMostRecentRecord := range tc.expectedMostRecentRecord {
				record, err := twapKeeper.GetMostRecentRecordStoreRepresentation(ctx, expectedMostRecentRecord.PoolId, expectedMostRecentRecord.Asset0Denom, expectedMostRecentRecord.Asset1Denom)
				s.Require().NoError(err)
				s.Require().Equal(expectedMostRecentRecord, record)
			}
		})
	}
}

// TestTWAPExportGenesis tests that genesis is exported correctly.
// It first initializes genesis to the expected value. Then, attempts
// to export it. Lastly, compares exported to the expected.
func (s *TestSuite) TestTWAPExportGenesis() {
	testCases := map[string]struct {
		expectedGenesis *types.GenesisState
	}{
		"default genesis": {
			expectedGenesis: types.DefaultGenesis(),
		},
		"custom genesis": {
			expectedGenesis: basicCustomGenesis,
		},
		"custom multi-record; increasing": {
			expectedGenesis: increasingOrderByTimeRecordsPoolOne,
		},
		"custom multi-record; decreasing": {
			expectedGenesis: decreasingOrderByTimeRecordsPoolTwo,
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			s.Setup()
			// Setup.
			app := s.App
			ctx := s.Ctx
			twapKeeper := app.TwapKeeper

			twapKeeper.InitGenesis(ctx, tc.expectedGenesis)

			// Test.
			actualGenesis := twapKeeper.ExportGenesis(ctx)

			// Assertions.
			s.Require().Equal(tc.expectedGenesis.Params, actualGenesis.Params)

			// Sort expected by time. This is done because the exported genesis returns
			// recors in ascending order by time.
			sort.Slice(tc.expectedGenesis.Twaps, func(i, j int) bool {
				return tc.expectedGenesis.Twaps[i].Time.Before(tc.expectedGenesis.Twaps[j].Time)
			})

			s.Require().Equal(tc.expectedGenesis.Twaps, actualGenesis.Twaps)
		})
	}
}

// sets up a new two asset pool, with spot price 1
func (s *TestSuite) setupDefaultPool() (poolId uint64, denomA, denomB string) {
	poolId = s.PrepareBalancerPoolWithCoins(defaultTwoAssetCoins[0], defaultTwoAssetCoins[1])
	denomA, denomB = defaultTwoAssetCoins[1].Denom, defaultTwoAssetCoins[0].Denom
	return
}

// preSetRecords pre sets records on the twap keeper to the
// given records.
func (s *TestSuite) preSetRecords(records []types.TwapRecord) {
	for _, record := range records {
		s.twapkeeper.StoreNewRecord(s.Ctx, record)
	}
}

// preSetRecords pre sets records on the twap keeper to the
// given records. The records are updated to use the provided pool ID
func (s *TestSuite) preSetRecordsWithPoolId(poolId uint64, records []types.TwapRecord) {
	for _, record := range records {
		record.PoolId = poolId
		s.twapkeeper.StoreNewRecord(s.Ctx, record)
	}
}

// getAllHistoricalRecordsForPool returns all historical records for a given pool.
func (s *TestSuite) getAllHistoricalRecordsForPool(poolId uint64) []types.TwapRecord {
	allRecords, err := s.twapkeeper.GetAllHistoricalPoolIndexedTWAPs(s.Ctx)
	s.Require().NoError(err)
	filteredRecords := make([]types.TwapRecord, 0)
	for _, record := range allRecords {
		if record.PoolId == poolId {
			filteredRecords = append(filteredRecords, record)
		}
	}
	return filteredRecords
}

// validateExpectedRecords validates that the twap keeper has the expected records.
func (s *TestSuite) validateExpectedRecords(expectedRecords []types.TwapRecord) {
	twapKeeper := s.twapkeeper

	// validate that the pool indexed TWAPs are cleared.
	poolIndexedTwaps, err := twapKeeper.GetAllHistoricalPoolIndexedTWAPs(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(poolIndexedTwaps, len(expectedRecords))
	// N.B.: ElementsMatch is used here because order might differ from expected due to
	// diverging indexing structure.
	s.Require().ElementsMatch(poolIndexedTwaps, expectedRecords)
}

// createTestRecordsFromTime creates and returns 6 test records in the following order:
// - 1 record at time t - 2 seconds, with pool id 1
// - 3 records at time t - 1 seconds, with pool id 2 (3 asset pool)
// - 1 record at time t, with pool id 3
// - 1 record at time t + 1 seconds, with pool id 3
func (s *TestSuite) createTestRecordsFromTime(t time.Time) (types.TwapRecord, types.TwapRecord, types.TwapRecord, types.TwapRecord, types.TwapRecord, types.TwapRecord) { //nolint:revive // this function breaks the function result limit of 4 return results.  TODO: fixme
	baseRecord := newEmptyPriceRecord(basePoolId, t, denom0, denom1)

	tMin1 := t.Add(-time.Second)
	tMin1RecordAB := newEmptyPriceRecord(basePoolId+1, tMin1, denom0, denom1)
	tMin1RecordAC := newEmptyPriceRecord(basePoolId+1, tMin1, denom0, denom2)
	tMin1RecordBC := newEmptyPriceRecord(basePoolId+1, tMin1, denom1, denom2)

	tMin2 := t.Add(-time.Second * 2)
	tMin2Record := newEmptyPriceRecord(basePoolId+2, tMin2, denom0, denom1)

	tPlus1 := t.Add(time.Second)
	tPlus1Record := newEmptyPriceRecord(basePoolId+3, tPlus1, denom0, denom1)

	return tMin2Record, tMin1RecordAB, tMin1RecordAC, tMin1RecordBC, baseRecord, tPlus1Record
}

// createTestRecordsFromTimeInPool creates and returns 12 test records in the following order:
// - 3 records at time t - 2 seconds
// - 3 records at time t - 1 seconds
// - 3 records at time t
// - 3 records t time t + 1 seconds
// all returned records belong to the same pool with poolId
func (s *TestSuite) CreateTestRecordsFromTimeInPool(t time.Time, poolId uint64) (types.TwapRecord, types.TwapRecord, types.TwapRecord, types.TwapRecord, types.TwapRecord, types.TwapRecord, //nolint:revive // this function breaks the function result limit of 4 return results.  TODO: fixme
	types.TwapRecord, types.TwapRecord, types.TwapRecord, types.TwapRecord, types.TwapRecord, types.TwapRecord,
) {
	baseRecordAB := newEmptyPriceRecord(poolId, t, denom0, denom1)
	baseRecordAC := newEmptyPriceRecord(poolId, t, denom0, denom2)
	baseRecordBC := newEmptyPriceRecord(poolId, t, denom1, denom2)

	tMin1 := t.Add(-time.Second)
	tMin1RecordAB := newEmptyPriceRecord(poolId, tMin1, denom0, denom1)
	tMin1RecordAC := newEmptyPriceRecord(poolId, tMin1, denom0, denom2)
	tMin1RecordBC := newEmptyPriceRecord(poolId, tMin1, denom1, denom2)

	tMin2 := t.Add(-time.Second * 2)
	tMin2RecordAB := newEmptyPriceRecord(poolId, tMin2, denom0, denom1)
	tMin2RecordAC := newEmptyPriceRecord(poolId, tMin2, denom0, denom2)
	tMin2RecordBC := newEmptyPriceRecord(poolId, tMin2, denom1, denom2)

	tPlus1 := t.Add(time.Second)
	tPlus1RecordAB := newEmptyPriceRecord(poolId, tPlus1, denom0, denom1)
	tPlus1RecordAC := newEmptyPriceRecord(poolId, tPlus1, denom0, denom2)
	tPlus1RecordBC := newEmptyPriceRecord(poolId, tPlus1, denom1, denom2)

	return tMin2RecordAB, tMin2RecordAC, tMin2RecordBC, tMin1RecordAB, tMin1RecordAC, tMin1RecordBC, baseRecordAB, baseRecordAC, baseRecordBC, tPlus1RecordAB, tPlus1RecordAC, tPlus1RecordBC
}

// newTwoAssetPoolTwapRecordWithDefaults creates a single twap records, mimicking what one would expect from a two asset pool.
// given a spot price 0 (sp0), this spot price is assigned to denomA and sp0 is then created and assigned to denomB by
// calculating (1 / spA).osmomath.Dec
func newTwoAssetPoolTwapRecordWithDefaults(t time.Time, sp0, accum0, accum1, geomAccum osmomath.Dec) types.TwapRecord {
	return types.TwapRecord{
		PoolId:      1,
		Time:        t,
		Asset0Denom: denom0,
		Asset1Denom: denom1,

		P0LastSpotPrice:             sp0,
		P1LastSpotPrice:             osmomath.OneDec().Quo(sp0),
		P0ArithmeticTwapAccumulator: accum0,
		P1ArithmeticTwapAccumulator: accum1,
		GeometricTwapAccumulator:    geomAccum,
	}
}

// newThreeAssetPoolTwapRecordWithDefaults creates three twap records, mimicking what one would expect from a three asset pool.
// given a spot price 0 (sp0), this spot price is assigned to denomA and referred to as spA. spB is then created and assigned by
// calculating (1 / spA). Finally spC is created and assigned by calculating (2 * spA).osmomath.Dec
func newThreeAssetPoolTwapRecordWithDefaults(t time.Time, sp0, accumA, accumB, accumC, geomAccumAB, geomAccumAC, geomAccumBC osmomath.Dec) (types.TwapRecord, types.TwapRecord, types.TwapRecord) {
	spA := sp0
	spB := osmomath.OneDec().Quo(sp0)
	spC := sp0.Mul(osmomath.NewDec(2))
	twapAB := types.TwapRecord{
		PoolId:      2,
		Time:        t,
		Asset0Denom: denom0,
		Asset1Denom: denom1,

		P0LastSpotPrice:             spA,
		P1LastSpotPrice:             spB,
		P0ArithmeticTwapAccumulator: accumA,
		P1ArithmeticTwapAccumulator: accumB,
		GeometricTwapAccumulator:    geomAccumAB,
	}
	twapAC := twapAB
	twapAC.Asset1Denom = denom2
	twapAC.P1LastSpotPrice = spC
	twapAC.P1ArithmeticTwapAccumulator = accumC
	twapAC.GeometricTwapAccumulator = geomAccumAC
	twapBC := twapAC
	twapBC.Asset0Denom = denom1
	twapBC.P0LastSpotPrice = spB
	twapBC.P0ArithmeticTwapAccumulator = accumB
	twapBC.GeometricTwapAccumulator = geomAccumBC

	return twapAB, twapAC, twapBC
}

func newEmptyPriceRecord(poolId uint64, t time.Time, asset0 string, asset1 string) types.TwapRecord {
	return types.TwapRecord{
		PoolId:      poolId,
		Time:        t,
		Asset0Denom: asset0,
		Asset1Denom: asset1,

		P0LastSpotPrice:             osmomath.ZeroDec(),
		P1LastSpotPrice:             osmomath.ZeroDec(),
		P0ArithmeticTwapAccumulator: osmomath.ZeroDec(),
		P1ArithmeticTwapAccumulator: osmomath.ZeroDec(),
		GeometricTwapAccumulator:    osmomath.ZeroDec(),
	}
}

func withPrice0Set(twapRecord types.TwapRecord, price0ToSet osmomath.Dec) types.TwapRecord {
	twapRecord.P0LastSpotPrice = price0ToSet
	return twapRecord
}

func withPrice1Set(twapRecord types.TwapRecord, price1ToSet osmomath.Dec) types.TwapRecord {
	twapRecord.P1LastSpotPrice = price1ToSet
	return twapRecord
}

func withTime(twapRecord types.TwapRecord, time time.Time) types.TwapRecord {
	twapRecord.Time = time
	return twapRecord
}

func newRecord(poolId uint64, t time.Time, sp0, accum0, accum1, geomAccum osmomath.Dec) types.TwapRecord {
	return types.TwapRecord{
		PoolId:          poolId,
		Asset0Denom:     defaultTwoAssetCoins[0].Denom,
		Asset1Denom:     defaultTwoAssetCoins[1].Denom,
		Time:            t,
		P0LastSpotPrice: sp0,
		P1LastSpotPrice: osmomath.OneDec().Quo(sp0),
		// make new copies
		P0ArithmeticTwapAccumulator: accum0.Add(osmomath.ZeroDec()),
		P1ArithmeticTwapAccumulator: accum1.Add(osmomath.ZeroDec()),
		GeometricTwapAccumulator:    geomAccum.Add(osmomath.ZeroDec()),
	}
}

// make an expected record for math tests, wosmomath.Dect other values in the test runner.
func newExpRecord(accum0, accum1, geomAccum osmomath.Dec) types.TwapRecord {
	return types.TwapRecord{
		Asset0Denom: defaultTwoAssetCoins[0].Denom,
		Asset1Denom: defaultTwoAssetCoins[1].Denom,
		// make new copies
		P0ArithmeticTwapAccumulator: accum0.Add(osmomath.ZeroDec()),
		P1ArithmeticTwapAccumulator: accum1.Add(osmomath.ZeroDec()),
		GeometricTwapAccumulator:    geomAccum.Add(osmomath.ZeroDec()),
	}
}

func newThreeAssetRecord(poolId uint64, t time.Time, sp0, accumA, accumB, accumC, geomAccumAB, geomAccumAC, geomAccumBC osmomath.Dec) []types.TwapRecord { //nolint:unparam // poolID always receives 2 but this could change later
	spA := sp0
	spB := osmomath.OneDec().Quo(sp0)
	spC := sp0.Mul(osmomath.NewDec(2))
	twapAB := types.TwapRecord{
		PoolId:          poolId,
		Asset0Denom:     defaultThreeAssetCoins[0].Denom,
		Asset1Denom:     defaultThreeAssetCoins[1].Denom,
		Time:            t,
		P0LastSpotPrice: spA,
		P1LastSpotPrice: spB,
		// make new copies
		P0ArithmeticTwapAccumulator: accumA.Add(osmomath.ZeroDec()),
		P1ArithmeticTwapAccumulator: accumB.Add(osmomath.ZeroDec()),
		GeometricTwapAccumulator:    geomAccumAB.Add(osmomath.ZeroDec()),
	}
	twapAC := twapAB
	twapAC.Asset1Denom = denom2
	twapAC.P1LastSpotPrice = spC
	twapAC.P1ArithmeticTwapAccumulator = accumC
	twapAC.GeometricTwapAccumulator = geomAccumAC.Add(osmomath.ZeroDec())
	twapBC := twapAC
	twapBC.Asset0Denom = denom1
	twapBC.P0LastSpotPrice = spB
	twapBC.P0ArithmeticTwapAccumulator = accumB
	twapBC.GeometricTwapAccumulator = geomAccumBC.Add(osmomath.ZeroDec())
	return []types.TwapRecord{twapAB, twapAC, twapBC}
}

// make an expected record for math tests, we adjust other values in the test runner.osmomath.Dec
func newThreeAssetExpRecord(poolId uint64, accumA, accumB, accumC, geomAccumAB, geomAccumAC, geomAccumBC osmomath.Dec) []types.TwapRecord {
	twapAB := types.TwapRecord{
		PoolId:      poolId,
		Asset0Denom: defaultThreeAssetCoins[0].Denom,
		Asset1Denom: defaultThreeAssetCoins[1].Denom,
		// make new copies
		P0ArithmeticTwapAccumulator: accumA.Add(osmomath.ZeroDec()),
		P1ArithmeticTwapAccumulator: accumB.Add(osmomath.ZeroDec()),
		GeometricTwapAccumulator:    geomAccumAB.Add(osmomath.ZeroDec()),
	}
	twapAC := twapAB
	twapAC.Asset1Denom = denom2
	twapAC.P1ArithmeticTwapAccumulator = accumC
	twapAC.GeometricTwapAccumulator = geomAccumAC.Add(osmomath.ZeroDec())
	twapBC := twapAC
	twapBC.Asset0Denom = denom1
	twapBC.P0ArithmeticTwapAccumulator = accumB
	twapBC.GeometricTwapAccumulator = geomAccumBC.Add(osmomath.ZeroDec())
	return []types.TwapRecord{twapAB, twapAC, twapBC}
}

func newOneSidedRecord(time time.Time, accum osmomath.Dec, useP0 bool) types.TwapRecord {
	record := types.TwapRecord{Time: time, Asset0Denom: denom0, Asset1Denom: denom1}
	if useP0 {
		record.P0ArithmeticTwapAccumulator = accum
	} else {
		record.P1ArithmeticTwapAccumulator = accum
	}
	record.P0LastSpotPrice = osmomath.ZeroDec()
	record.P1LastSpotPrice = osmomath.OneDec()
	return record
}

func newOneSidedGeometricRecord(time time.Time, accum osmomath.Dec) types.TwapRecord {
	record := types.TwapRecord{Time: time, Asset0Denom: denom0, Asset1Denom: denom1}
	record.GeometricTwapAccumulator = accum
	record.P0LastSpotPrice = osmomath.NewDec(10)
	return record
}

func newThreeAssetOneSidedRecord(time time.Time, accum osmomath.Dec, useP0 bool) []types.TwapRecord { //nolint:unparam // useP0 always true, but this could change later.
	record := types.TwapRecord{Time: time, Asset0Denom: denom0, Asset1Denom: denom1}
	if useP0 {
		record.P0ArithmeticTwapAccumulator = accum
	} else {
		record.P1ArithmeticTwapAccumulator = accum
	}
	record.GeometricTwapAccumulator = accum
	record.P0LastSpotPrice = osmomath.ZeroDec()
	record.P1LastSpotPrice = osmomath.OneDec()
	records := []types.TwapRecord{record, record, record}
	records[1].Asset1Denom = denom2
	records[2].Asset0Denom = denom1
	records[2].Asset1Denom = denom2
	return records
}

func recordWithUpdatedAccum(record types.TwapRecord, accum0 osmomath.Dec, accum1, geomAccum osmomath.Dec) types.TwapRecord {
	record.P0ArithmeticTwapAccumulator = accum0
	record.P1ArithmeticTwapAccumulator = accum1
	record.GeometricTwapAccumulator = geomAccum
	return record
}

func recordWithUpdatedSpotPrice(record types.TwapRecord, sp0 osmomath.Dec, sp1 osmomath.Dec) types.TwapRecord {
	record.P0LastSpotPrice = sp0
	record.P1LastSpotPrice = sp1
	return record
}
