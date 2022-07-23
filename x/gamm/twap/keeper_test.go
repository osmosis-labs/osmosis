package twap_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v10/app/apptesting"
	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap"
	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap/types"
)

// TODO: Consider switching this everywhere
var denom0 = "token/B"
var denom1 = "token/A"
var defaultUniV2Coins = sdk.NewCoins(sdk.NewInt64Coin(denom0, 1_000_000_000), sdk.NewInt64Coin(denom1, 1_000_000_000))
var baseTime = time.Unix(1257894000, 0).UTC()

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
}

// sets up a new two asset pool, with spot price 1
func (s *TestSuite) setupDefaultPool() (poolId uint64, denomA, denomB string) {
	poolId = s.PrepareUni2PoolWithAssets(defaultUniV2Coins[0], defaultUniV2Coins[1])
	denomA, denomB = defaultUniV2Coins[1].Denom, defaultUniV2Coins[0].Denom
	return
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
