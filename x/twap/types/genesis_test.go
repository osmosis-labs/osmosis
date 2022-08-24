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
				NewParams("week", 48*time.Hour), // invalid empty string
				[]TwapRecord{
					baseRecord,
				}),

			expectedErr: true,
		},
		"invalid recordHistoryKeepPeriod - error": {
			twapGenesis: NewGenesisState(
				NewParams("", -1*time.Hour), // invalid duration
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
	testCases := map[string]struct {
		twapRecord TwapRecord

		expectedErr bool
	}{
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
		"invalid asset0 denom": {
			twapRecord: func() TwapRecord {
				r := baseRecord
				r.Asset0Denom = ""
				return r
			}(),

			expectedErr: true,
		},
		"invalid asset1 denom": {
			twapRecord: func() TwapRecord {
				r := baseRecord
				r.Asset1Denom = ""
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
		"invalid p0 last spot price": {
			twapRecord: func() TwapRecord {
				r := baseRecord
				r.P0LastSpotPrice = sdk.ZeroDec()
				return r
			}(),

			expectedErr: true,
		},
		"invalid p1 last spot price": {
			twapRecord: func() TwapRecord {
				r := baseRecord
				r.P1LastSpotPrice = sdk.ZeroDec()
				return r
			}(),

			expectedErr: true,
		},
		"invalid p0 arithmetic accum": {
			twapRecord: func() TwapRecord {
				r := baseRecord
				r.P0ArithmeticTwapAccumulator = sdk.OneDec().Neg()
				return r
			}(),

			expectedErr: true,
		},
		"invalid p1 arithmetic accum": {
			twapRecord: func() TwapRecord {
				r := baseRecord
				r.P1ArithmeticTwapAccumulator = sdk.OneDec().Neg()
				return r
			}(),

			expectedErr: true,
		},
	}

	for name, tc := range testCases {
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
