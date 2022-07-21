package twap_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap"
	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap/types"
)

var zeroDec = sdk.ZeroDec()
var oneDec = sdk.OneDec()
var twoDec = oneDec.Add(oneDec)
var OneSec = sdk.NewDec(1e9)

func newRecord(t time.Time, sp0, accum0, accum1 sdk.Dec) types.TwapRecord {
	return types.TwapRecord{
		Asset0Denom:     defaultUniV2Coins[0].Denom,
		Asset1Denom:     defaultUniV2Coins[1].Denom,
		Time:            t,
		P0LastSpotPrice: sp0,
		P1LastSpotPrice: sdk.OneDec().Quo(sp0),
		// make new copies
		P0ArithmeticTwapAccumulator: accum0.Add(sdk.ZeroDec()),
		P1ArithmeticTwapAccumulator: accum1.Add(sdk.ZeroDec()),
	}
}

// make an expected record for math tests, we adjust other values in the test runner.
func newExpRecord(accum0, accum1 sdk.Dec) types.TwapRecord {
	return types.TwapRecord{
		Asset0Denom: defaultUniV2Coins[0].Denom,
		Asset1Denom: defaultUniV2Coins[1].Denom,
		// make new copies
		P0ArithmeticTwapAccumulator: accum0.Add(sdk.ZeroDec()),
		P1ArithmeticTwapAccumulator: accum1.Add(sdk.ZeroDec()),
	}
}

func TestInterpolateRecord(t *testing.T) {
	tests := map[string]struct {
		record          types.TwapRecord
		interpolateTime time.Time
		expRecord       types.TwapRecord
	}{
		"0accum": {
			record:          newRecord(time.Unix(1, 0), sdk.NewDec(10), zeroDec, zeroDec),
			interpolateTime: time.Unix(2, 0),
			expRecord:       newExpRecord(OneSec.MulInt64(10), OneSec.QuoInt64(10)),
		},
		"small starting accumulators": {
			record:          newRecord(time.Unix(1, 0), sdk.NewDec(10), oneDec, twoDec),
			interpolateTime: time.Unix(2, 0),
			expRecord:       newExpRecord(oneDec.Add(OneSec.MulInt64(10)), twoDec.Add(OneSec.QuoInt64(10))),
		},
		"larger time interval": {
			record:          newRecord(time.Unix(11, 0), sdk.NewDec(10), oneDec, twoDec),
			interpolateTime: time.Unix(55, 0),
			expRecord:       newExpRecord(oneDec.Add(OneSec.MulInt64(44*10)), twoDec.Add(OneSec.MulInt64(44).QuoInt64(10))),
		},
		"same time": {
			record:          newRecord(time.Unix(1, 0), sdk.NewDec(10), oneDec, twoDec),
			interpolateTime: time.Unix(1, 0),
			expRecord:       newExpRecord(oneDec, twoDec),
		},
		// TODO: Overflow tests
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// correct expected record based off copy/paste values
			test.expRecord.Time = test.interpolateTime
			test.expRecord.P0LastSpotPrice = test.record.P0LastSpotPrice
			test.expRecord.P1LastSpotPrice = test.record.P1LastSpotPrice

			gotRecord := twap.InterpolateRecord(test.record, test.interpolateTime)
			require.Equal(t, test.expRecord, gotRecord)
		})
	}
}

func (s *TestSuite) TestUpdateTwap() {
	poolId := s.PrepareUni2PoolWithAssets(defaultUniV2Coins[0], defaultUniV2Coins[1])
	newSp := sdk.OneDec()

	tests := map[string]struct {
		record     types.TwapRecord
		updateTime time.Time
		expRecord  types.TwapRecord
	}{
		"0 accum start": {
			record:     newRecord(time.Unix(1, 0), sdk.NewDec(10), zeroDec, zeroDec),
			updateTime: time.Unix(2, 0),
			expRecord:  newExpRecord(OneSec.MulInt64(10), OneSec.QuoInt64(10)),
		},
	}
	for name, test := range tests {
		s.Run(name, func() {
			// setup common, block time, pool Id, expected spot prices
			s.Ctx = s.Ctx.WithBlockTime(test.updateTime.UTC())
			test.record.PoolId = poolId
			test.expRecord.PoolId = poolId
			test.expRecord.P0LastSpotPrice = newSp
			test.expRecord.P1LastSpotPrice = newSp
			test.expRecord.Height = s.Ctx.BlockHeight()
			test.expRecord.Time = s.Ctx.BlockTime()

			newRecord := s.twapkeeper.UpdateRecord(s.Ctx, test.record)
			s.Require().Equal(test.expRecord, newRecord)
		})
	}
}
