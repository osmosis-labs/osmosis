package twap_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v10/app/apptesting"
	"github.com/osmosis-labs/osmosis/v10/x/twap"
	"github.com/osmosis-labs/osmosis/v10/x/twap/types"
)

// TODO: Consider switching this everywhere
var (
	denom0                   = "token/A"
	denom1                   = "token/B"
	denom2                   = "token/C"
	defaultUniV2Coins        = sdk.NewCoins(sdk.NewInt64Coin(denom1, 1_000_000_000), sdk.NewInt64Coin(denom0, 1_000_000_000))
	baseTime                 = time.Unix(1257894000, 0).UTC()
	tPlusOne                 = baseTime.Add(time.Second)
	basePoolId        uint64 = 1
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

// sets up a new two asset pool, with spot price 1
func (s *TestSuite) setupDefaultPool() (poolId uint64, denomA, denomB string) {
	poolId = s.PrepareUni2PoolWithAssets(defaultUniV2Coins[0], defaultUniV2Coins[1])
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

// createTestRecordsFromTime creates and returns 2 test records in the following order:
// - at time t - 2
// - at time t - 1
// - at time t
// - at time t + 1
func (s *TestSuite) createTestRecordsFromTime(t time.Time) (types.TwapRecord, types.TwapRecord, types.TwapRecord, types.TwapRecord) {
	baseRecord := newEmptyPriceRecord(basePoolId, t, denom1, denom0)

	tMin1 := t.Add(-time.Second)
	tMin1Record := newEmptyPriceRecord(basePoolId+1, tMin1, denom1, denom0)

	tMin2 := t.Add(-time.Second * 2)
	tMin2Record := newEmptyPriceRecord(basePoolId+2, tMin2, denom1, denom0)

	tPlus1 := t.Add(time.Second)
	tPlus1Record := newEmptyPriceRecord(basePoolId+3, tPlus1, denom1, denom0)

	return tMin2Record, tMin1Record, baseRecord, tPlus1Record
}

func newTwapRecordWithDefaults(t time.Time, sp0, accum0, accum1 sdk.Dec) types.TwapRecord {
	return types.TwapRecord{
		PoolId:      1,
		Time:        t,
		Asset0Denom: denom1,
		Asset1Denom: denom0,

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
