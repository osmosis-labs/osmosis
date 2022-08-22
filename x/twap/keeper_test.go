package twap_test

import (
	"sort"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v11/app/apptesting"
	"github.com/osmosis-labs/osmosis/v11/app/apptesting/osmoassert"
	"github.com/osmosis-labs/osmosis/v11/x/twap"
	"github.com/osmosis-labs/osmosis/v11/x/twap/types"
)

// TODO: Consider switching this everywhere
var (
	denom0                        = "token/B"
	denom1                        = "token/A"
	denom2                        = "token/C"
	defaultUniV2Coins             = sdk.NewCoins(sdk.NewInt64Coin(denom0, 1_000_000_000), sdk.NewInt64Coin(denom1, 1_000_000_000))
	defaultThreeAssetCoins        = sdk.NewCoins(sdk.NewInt64Coin(denom0, 1_000_000_000), sdk.NewInt64Coin(denom1, 1_000_000_000), sdk.NewInt64Coin(denom2, 1_000_000_000))
	baseTime                      = time.Unix(1257894000, 0).UTC()
	tPlusOne                      = baseTime.Add(time.Second)
	basePoolId             uint64 = 1
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
}

var (
	basicParams = types.NewParams("week", 48*time.Hour)

	mostRecentRecordPoolOne = types.TwapRecord{
		PoolId:                      basePoolId,
		Asset0Denom:                 denom0,
		Asset1Denom:                 denom1,
		Height:                      3,
		Time:                        tPlusOne.Add(time.Second),
		P0LastSpotPrice:             sdk.OneDec(),
		P1LastSpotPrice:             sdk.OneDec(),
		P0ArithmeticTwapAccumulator: sdk.OneDec(),
		P1ArithmeticTwapAccumulator: sdk.OneDec(),
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
				P0LastSpotPrice:             sdk.OneDec(),
				P1LastSpotPrice:             sdk.OneDec(),
				P0ArithmeticTwapAccumulator: sdk.OneDec(),
				P1ArithmeticTwapAccumulator: sdk.OneDec(),
			},
			{
				PoolId:                      basePoolId,
				Asset0Denom:                 denom0,
				Asset1Denom:                 denom1,
				Height:                      2,
				Time:                        tPlusOne,
				P0LastSpotPrice:             sdk.OneDec(),
				P1LastSpotPrice:             sdk.OneDec(),
				P0ArithmeticTwapAccumulator: sdk.OneDec(),
				P1ArithmeticTwapAccumulator: sdk.OneDec(),
			},
			mostRecentRecordPoolOne,
		})

	mostRecentRecordPoolTwo = types.TwapRecord{
		PoolId:                      basePoolId,
		Asset0Denom:                 denom0,
		Asset1Denom:                 denom2,
		Height:                      1,
		Time:                        tPlusOne.Add(time.Second),
		P0LastSpotPrice:             sdk.OneDec(),
		P1LastSpotPrice:             sdk.OneDec(),
		P0ArithmeticTwapAccumulator: sdk.OneDec(),
		P1ArithmeticTwapAccumulator: sdk.OneDec(),
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
				P0LastSpotPrice:             sdk.OneDec(),
				P1LastSpotPrice:             sdk.OneDec(),
				P0ArithmeticTwapAccumulator: sdk.OneDec(),
				P1ArithmeticTwapAccumulator: sdk.OneDec(),
			},
			{
				PoolId:                      basePoolId,
				Asset0Denom:                 denom0,
				Asset1Denom:                 denom2,
				Height:                      3,
				Time:                        baseTime,
				P0LastSpotPrice:             sdk.OneDec(),
				P1LastSpotPrice:             sdk.OneDec(),
				P0ArithmeticTwapAccumulator: sdk.OneDec(),
				P1ArithmeticTwapAccumulator: sdk.OneDec(),
			},
		})

	bothPoolsGenesis = types.NewGenesisState(
		basicParams,
		append(increasingOrderByTimeRecordsPoolOne.Twaps, decreasingOrderByTimeRecordsPoolTwo.Twaps...),
	)
)

// TestTWAPInitGenesis tests that genesis is initialized correctly
// with different parameters and state.
// Asserts that the most recent records are set correctly.
func (suite *TestSuite) TestTwapInitGenesis() {
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
						P0LastSpotPrice:             sdk.OneDec(),
						P1LastSpotPrice:             sdk.OneDec(),
						P0ArithmeticTwapAccumulator: sdk.OneDec(),
						P1ArithmeticTwapAccumulator: sdk.OneDec(),
					},
				}),

			expectPanic: true,
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			suite.Setup()
			// Setup.
			ctx := suite.Ctx
			twapKeeper := suite.App.TwapKeeper

			// Test.
			osmoassert.ConditionalPanic(suite.T(), tc.expectPanic, func() { twapKeeper.InitGenesis(ctx, tc.twapGenesis) })
			if tc.expectPanic {
				return
			}

			// Assertions.

			// Parameters were set.
			suite.Require().Equal(tc.twapGenesis.Params, twapKeeper.GetParams(ctx))

			for _, expectedMostRecentRecord := range tc.expectedMostRecentRecord {
				record, err := twapKeeper.GetMostRecentRecordStoreRepresentation(ctx, expectedMostRecentRecord.PoolId, expectedMostRecentRecord.Asset0Denom, expectedMostRecentRecord.Asset1Denom)
				suite.Require().NoError(err)
				suite.Require().Equal(expectedMostRecentRecord, record)
			}
		})
	}
}

// TestTWAPExportGenesis tests that genesis is exported correctly.
// It first initializes genesis to the expected value. Then, attempts
// to export it. Lastly, compares exported to the expected.
func (suite *TestSuite) TestTWAPExportGenesis() {
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
		suite.Run(name, func() {
			suite.Setup()
			// Setup.
			app := suite.App
			ctx := suite.Ctx
			twapKeeper := app.TwapKeeper

			twapKeeper.InitGenesis(ctx, tc.expectedGenesis)

			// Test.
			actualGenesis := twapKeeper.ExportGenesis(ctx)

			// Assertions.
			suite.Require().Equal(tc.expectedGenesis.Params, actualGenesis.Params)

			// Sort expected by time. This is done because the exported genesis returns
			// recors in ascending order by time.
			sort.Slice(tc.expectedGenesis.Twaps, func(i, j int) bool {
				return tc.expectedGenesis.Twaps[i].Time.Before(tc.expectedGenesis.Twaps[j].Time)
			})

			suite.Require().Equal(tc.expectedGenesis.Twaps, actualGenesis.Twaps)
		})
	}
}

// sets up a new two asset pool, with spot price 1
func (s *TestSuite) setupDefaultPool() (poolId uint64, denomA, denomB string) {
	poolId = s.PrepareBalancerPoolWithCoins(defaultUniV2Coins[0], defaultUniV2Coins[1])
	denomA, denomB = defaultUniV2Coins[1].Denom, defaultUniV2Coins[0].Denom
	return
}

// preSetRecords pre sets records on the twap keeper to the
// given records.
func (s *TestSuite) preSetRecords(records []types.TwapRecord) {
	for _, record := range records {
		s.twapkeeper.StoreNewRecord(s.Ctx, record)
	}
}

// validateExpectedRecords validates that the twap keeper has the expected records.
func (s *TestSuite) validateExpectedRecords(expectedRecords []types.TwapRecord) {
	twapKeeper := s.twapkeeper

	// validate that the time indexed TWAPs are cleared.
	timeIndexedTwaps, err := twapKeeper.GetAllHistoricalTimeIndexedTWAPs(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(timeIndexedTwaps, len(expectedRecords))
	s.Require().Equal(timeIndexedTwaps, expectedRecords)

	// validate that the pool indexed TWAPs are cleared.
	poolIndexedTwaps, err := twapKeeper.GetAllHistoricalPoolIndexedTWAPs(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(poolIndexedTwaps, len(expectedRecords))
	// N.B.: ElementsMatch is used here because order might differ from expected due to
	// diverging indexing structure.
	s.Require().ElementsMatch(poolIndexedTwaps, expectedRecords)
}

// createTestRecordsFromTime creates and returns 4 test records in the following order:
// - at time t - 2 seconds
// - at time t - 1 seconds
// - at time t
// - at time t + 1 seconds
func (s *TestSuite) createTestRecordsFromTime(t time.Time) (types.TwapRecord, types.TwapRecord, types.TwapRecord, types.TwapRecord) {
	baseRecord := newEmptyPriceRecord(basePoolId, t, denom0, denom1)

	tMin1 := t.Add(-time.Second)
	tMin1Record := newEmptyPriceRecord(basePoolId+1, tMin1, denom0, denom1)

	tMin2 := t.Add(-time.Second * 2)
	tMin2Record := newEmptyPriceRecord(basePoolId+2, tMin2, denom0, denom1)

	tPlus1 := t.Add(time.Second)
	tPlus1Record := newEmptyPriceRecord(basePoolId+3, tPlus1, denom0, denom1)

	return tMin2Record, tMin1Record, baseRecord, tPlus1Record
}

func newTwapRecordWithDefaults(t time.Time, sp0, accum0, accum1 sdk.Dec) types.TwapRecord {
	return types.TwapRecord{
		PoolId:      1,
		Time:        t,
		Asset0Denom: denom0,
		Asset1Denom: denom1,

		P0LastSpotPrice:             sp0,
		P1LastSpotPrice:             sdk.OneDec().Quo(sp0),
		P0ArithmeticTwapAccumulator: accum0,
		P1ArithmeticTwapAccumulator: accum1,
	}
}

func newEmptyPriceRecord(poolId uint64, t time.Time, asset0 string, asset1 string) types.TwapRecord {
	return types.TwapRecord{
		PoolId:      poolId,
		Time:        t,
		Asset0Denom: asset0,
		Asset1Denom: asset1,

		P0LastSpotPrice:             sdk.ZeroDec(),
		P1LastSpotPrice:             sdk.ZeroDec(),
		P0ArithmeticTwapAccumulator: sdk.ZeroDec(),
		P1ArithmeticTwapAccumulator: sdk.ZeroDec(),
	}
}

func recordWithUpdatedAccum(record types.TwapRecord, accum0 sdk.Dec, accum1 sdk.Dec) types.TwapRecord {
	record.P0ArithmeticTwapAccumulator = accum0
	record.P1ArithmeticTwapAccumulator = accum1
	return record
}

func recordWithUpdatedSpotPrice(record types.TwapRecord, sp0 sdk.Dec, sp1 sdk.Dec) types.TwapRecord {
	record.P0LastSpotPrice = sp0
	record.P1LastSpotPrice = sp1
	return record
}
