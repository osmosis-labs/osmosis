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
		GeometricTwapAccumulator:    sdk.OneDec(),
	}
)

func TestGenesisState_Validate(t *testing.T) {
	var (
		basicParams = NewParams("week", 48*time.Hour)

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
					GeometricTwapAccumulator:    sdk.OneDec(),
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
					GeometricTwapAccumulator:    sdk.OneDec(),
				},
			})
	)

	withGeometricAcc := func(record TwapRecord, geometricAcc sdk.Dec) TwapRecord {
		record.GeometricTwapAccumulator = geometricAcc
		return record
	}

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
		"valid geometric twap acc is negative": {
			twapGenesis: NewGenesisState(basicParams, []TwapRecord{withGeometricAcc(baseRecord, sdk.NewDec(-1))}),
		},
		"invalid geometric twap acc is nil": {
			twapGenesis: NewGenesisState(basicParams, []TwapRecord{withGeometricAcc(baseRecord, sdk.Dec{})}),
			expectedErr: true,
		},
		"invalid genesis - pool ID doesn't exist": {
			twapGenesis: NewGenesisState(
				NewParams("week", 48*time.Hour),
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
		"invalid pruneEpochIdentifier - error": {
			twapGenesis: NewGenesisState(
				NewParams("", 48*time.Hour), // invalid empty string
				[]TwapRecord{
					baseRecord,
				}),

			expectedErr: true,
		},
		"invalid recordHistoryKeepPeriod - error": {
			twapGenesis: NewGenesisState(
				NewParams("week", -1*time.Hour), // invalid duration
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
		"invalid p0 last spot price: zero": {
			twapRecord: func() TwapRecord {
				r := baseRecord
				r.P0LastSpotPrice = sdk.ZeroDec()
				return r
			}(),

			expectedErr: true,
		},
		"invalid p0 last spot price: negative": {
			twapRecord: func() TwapRecord {
				r := baseRecord
				r.P0LastSpotPrice = sdk.OneDec().Neg()
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
		"invalid p0 last spot price: nil": {
			twapRecord: func() TwapRecord {
				r := TwapRecord{
					PoolId:                      basePoolId,
					Asset0Denom:                 denom0,
					Asset1Denom:                 denom1,
					Height:                      3,
					Time:                        tPlusOne.Add(time.Second),
					P1LastSpotPrice:             sdk.OneDec(),
					P0ArithmeticTwapAccumulator: sdk.OneDec(),
					P1ArithmeticTwapAccumulator: sdk.OneDec(),
				}
				return r
			}(),

			expectedErr: true,
		},
		"invalid p1 last spot price: nil": {
			twapRecord: func() TwapRecord {
				r := TwapRecord{
					PoolId:                      basePoolId,
					Asset0Denom:                 denom0,
					Asset1Denom:                 denom1,
					Height:                      3,
					Time:                        tPlusOne.Add(time.Second),
					P0LastSpotPrice:             sdk.OneDec(),
					P0ArithmeticTwapAccumulator: sdk.OneDec(),
					P1ArithmeticTwapAccumulator: sdk.OneDec(),
				}
				return r
			}(),

			expectedErr: true,
		},
		"invalid p0 arithmetic accum: negative": {
			twapRecord: func() TwapRecord {
				r := baseRecord
				r.P0ArithmeticTwapAccumulator = sdk.OneDec().Neg()
				return r
			}(),

			expectedErr: true,
		},
		"invalid p0 arithmetic accum: nil": {
			twapRecord: func() TwapRecord {
				r := TwapRecord{
					PoolId:                      basePoolId,
					Asset0Denom:                 denom0,
					Asset1Denom:                 denom1,
					Height:                      3,
					Time:                        tPlusOne.Add(time.Second),
					P0LastSpotPrice:             sdk.OneDec(),
					P1LastSpotPrice:             sdk.OneDec(),
					P1ArithmeticTwapAccumulator: sdk.OneDec(),
				}
				return r
			}(),

			expectedErr: true,
		},
		"invalid p1 arithmetic accum: nil": {
			twapRecord: func() TwapRecord {
				r := TwapRecord{
					PoolId:                      basePoolId,
					Asset0Denom:                 denom0,
					Asset1Denom:                 denom1,
					Height:                      3,
					Time:                        tPlusOne.Add(time.Second),
					P0LastSpotPrice:             sdk.OneDec(),
					P1LastSpotPrice:             sdk.OneDec(),
					P0ArithmeticTwapAccumulator: sdk.OneDec(),
				}
				return r
			}(),
			expectedErr: true,
		},
		"invalid geometric accum: nil": {
			twapRecord: func() TwapRecord {
				r := TwapRecord{
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
				return r
			}(),
			expectedErr: true,
		},
	}
	// make test cases symmetric
	testCasesSym := map[string]testcase{}
	for k, tc := range testCases {
		if tc.twapRecord.P0LastSpotPrice.IsNil() ||
			tc.twapRecord.P1LastSpotPrice.IsNil() ||
			tc.twapRecord.P0ArithmeticTwapAccumulator.IsNil() ||
			tc.twapRecord.P1ArithmeticTwapAccumulator.IsNil() {
			testCasesSym[k] = tc
		} else if tc.twapRecord.Asset0Denom != baseRecord.Asset0Denom ||
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
