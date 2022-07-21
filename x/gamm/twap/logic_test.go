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

func newRecord(t time.Time, sp0, accum0, accum1 sdk.Dec) types.TwapRecord {
	return types.TwapRecord{
		Time:                        t,
		P0LastSpotPrice:             sp0,
		P1LastSpotPrice:             sdk.OneDec().Quo(sp0),
		P0ArithmeticTwapAccumulator: accum0,
		P1ArithmeticTwapAccumulator: accum1,
	}
}

func TestInterpolateRecord(t *testing.T) {
	// make an expected record, we adjust other values in the test case.
	newExpRecord := func(accum0, accum1 sdk.Dec) types.TwapRecord {
		return types.TwapRecord{P0ArithmeticTwapAccumulator: accum0,
			P1ArithmeticTwapAccumulator: accum1}
	}

	OneSec := sdk.NewDec(1e9)
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
