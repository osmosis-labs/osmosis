package types

import (
	"testing"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

const (
	denom0            = "token/B"
	denom1            = "token/A"
	denom2            = "token/C"
	basePoolId uint64 = 1
)

var (
	baseTime   = time.Unix(1257894000, 0).UTC()
	tPlusOne   = baseTime.Add(time.Second)
	baseRecord = TwapRecord{
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
)

func TestGenesisState_Validate(t *testing.T) {
	var (
		basicParams = NewParams("week")

		basicCustomGenesis = NewGenesisState(
			basicParams,
			[]TwapRecord{
				baseRecord,
			})

		multiRecordGenesis = NewGenesisState(
			basicParams,
			[]TwapRecord{
				baseRecord,
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
				{
					PoolId:                      basePoolId,
					Asset0Denom:                 denom0,
					Asset1Denom:                 denom1,
					Height:                      3,
					Time:                        tPlusOne.Add(time.Second),
					P0LastSpotPrice:             sdk.OneDec(),
					P1LastSpotPrice:             sdk.OneDec(),
					P0ArithmeticTwapAccumulator: sdk.OneDec(),
					P1ArithmeticTwapAccumulator: sdk.OneDec(),
				},
			})
	)

	testCases := map[string]struct {
		twapGenesis *GenesisState

		expectedErr bool
	}{
		"valid default genesis": {
			twapGenesis: DefaultGenesis(),
		},
		"valid basic": {
			twapGenesis: basicCustomGenesis,
		},
		"valid multi record": {
			twapGenesis: multiRecordGenesis,
		},
		"valid empty records": {
			twapGenesis: NewGenesisState(basicParams, []TwapRecord{}),
		},
		"invalid genesis - error": {
			twapGenesis: NewGenesisState(
				NewParams("week"),
				[]TwapRecord{
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

			expectedErr: true,
		},
		"invalid param - error": {
			twapGenesis: NewGenesisState(
				NewParams(""), // invalid empty string
				[]TwapRecord{
					baseRecord,
				}),

			expectedErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Setup.

			// System under test.
			err := tc.twapGenesis.Validate()

			// Assertions.
			if tc.expectedErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestTWAPRecord_Validate(t *testing.T) {
	type testcase struct {
		twapRecord  TwapRecord
		expectedErr bool
	}
	testCases := map[string]testcase{
		"valid base record": {
			twapRecord: baseRecord,
		},
		"invalid pool id": {
			twapRecord: func() TwapRecord {
				r := baseRecord
				r.PoolId = 0
				return r
			}(),

			expectedErr: true,
		},
		"invalid denom": {
			twapRecord: func() TwapRecord {
				r := baseRecord
				r.Asset0Denom = ""
				return r
			}(),

			expectedErr: true,
		},
		"invalid height": {
			twapRecord: func() TwapRecord {
				r := baseRecord
				r.Height = 0
				return r
			}(),

			expectedErr: true,
		},
		"invalid time": {
			twapRecord: func() TwapRecord {
				r := baseRecord
				r.Time = time.Time{}
				return r
			}(),

			expectedErr: true,
		},
		"invalid last spot price": {
			twapRecord: func() TwapRecord {
				r := baseRecord
				r.P0LastSpotPrice = sdk.ZeroDec()
				return r
			}(),

			expectedErr: true,
		},
		"invalid last spot price with error": {
			twapRecord: func() TwapRecord {
				r := baseRecord
				r.LastErrorTime = r.Time
				r.P0LastSpotPrice = sdk.NewDec(5) // has to be distinct to be symmetric
				r.P1LastSpotPrice = sdk.ZeroDec() // need this to be 0, to test the other case on error
				return r
			}(),

			expectedErr: true,
		},
		"invalid arithmetic accum": {
			twapRecord: func() TwapRecord {
				r := baseRecord
				r.P0ArithmeticTwapAccumulator = sdk.OneDec().Neg()
				return r
			}(),

			expectedErr: true,
		},
	}

	// make test cases symmetric
	testCasesSym := map[string]testcase{}
	for k, tc := range testCases {
		if tc.twapRecord.Asset0Denom != baseRecord.Asset0Denom ||
			!tc.twapRecord.P0LastSpotPrice.Equal(baseRecord.P0LastSpotPrice) ||
			!tc.twapRecord.P0ArithmeticTwapAccumulator.Equal(baseRecord.P0ArithmeticTwapAccumulator) {
			testCasesSym[k+": asset 0"] = tc
			tSym := tc.twapRecord
			tSym.Asset0Denom, tSym.Asset1Denom = tSym.Asset1Denom, tSym.Asset0Denom
			tSym.P0LastSpotPrice, tSym.P1LastSpotPrice = tSym.P1LastSpotPrice, tSym.P0LastSpotPrice
			tSym.P0ArithmeticTwapAccumulator, tSym.P1ArithmeticTwapAccumulator = tSym.P1ArithmeticTwapAccumulator, tSym.P0ArithmeticTwapAccumulator
			testCasesSym[k+": asset 1"] = testcase{tSym, tc.expectedErr}
		} else {
			testCasesSym[k] = tc
		}
	}

	for name, tc := range testCasesSym {
		t.Run(name, func(t *testing.T) {
			// System under test.
			err := tc.twapRecord.Validate()

			// Assertions.
			if tc.expectedErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
