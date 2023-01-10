package types

import (
	"strings"
	"testing"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"
)

func TestFormatMostRecentTWAPKey(t *testing.T) {
	tests := map[string]struct {
		poolId uint64
		denom1 string
		denom2 string
		want   string
	}{
		"standard":       {poolId: 1, denom1: "B", denom2: "A", want: "recent_twap|00000000000000000001|B|A"},
		"standard2digit": {poolId: 10, denom1: "B", denom2: "A", want: "recent_twap|00000000000000000010|B|A"},
		"maxPoolId":      {poolId: ^uint64(0), denom1: "B", denom2: "A", want: "recent_twap|18446744073709551615|B|A"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := FormatMostRecentTWAPKey(tt.poolId, tt.denom1, tt.denom2)
			require.Equal(t, tt.want, string(got))
		})
	}
}

func TestFormatHistoricalTwapKeys(t *testing.T) {
	// go playground default time
	// 2009-11-10 23:00:00 +0000 UTC m=+0.000000001
	baseTime := time.Unix(1257894000, 0).UTC()
	tests := map[string]struct {
		poolId        uint64
		time          time.Time
		denom1        string
		denom2        string
		wantPoolIndex string
		wantTimeIndex string
	}{
		"standard": {poolId: 1, time: baseTime, denom1: "B", denom2: "A", wantTimeIndex: "historical_time_index|2009-11-10T23:00:00.000000000|1|B|A", wantPoolIndex: "historical_pool_index|1|B|A|2009-11-10T23:00:00.000000000"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotTimeKey := FormatHistoricalTimeIndexTWAPKey(tt.time, tt.poolId, tt.denom1, tt.denom2)
			gotPoolKey := FormatHistoricalPoolIndexTWAPKey(tt.poolId, tt.denom1, tt.denom2, tt.time)
			require.Equal(t, tt.wantTimeIndex, string(gotTimeKey))
			require.Equal(t, tt.wantPoolIndex, string(gotPoolKey))

			poolIndexPrefix := FormatHistoricalPoolIndexTimePrefix(tt.poolId, tt.denom1, tt.denom2)
			require.True(t, strings.HasPrefix(string(gotPoolKey), string(poolIndexPrefix)), string(gotPoolKey), string(poolIndexPrefix))

			poolIndexSuffix := FormatHistoricalPoolIndexTimeSuffix(tt.poolId, tt.denom1, tt.denom2, tt.time)
			require.True(t, strings.HasPrefix(string(poolIndexSuffix), string(gotPoolKey)))
		})
	}
}

func TestParseTwapFromBz(t *testing.T) {
	baseTime := time.Unix(1257894000, 0).UTC()
	baseParseRecord := TwapRecord{
		PoolId:                      123,
		Asset0Denom:                 "B",
		Asset1Denom:                 "A",
		Height:                      1,
		Time:                        baseTime,
		P0LastSpotPrice:             sdk.NewDecWithPrec(1, 5),
		P1LastSpotPrice:             sdk.NewDecWithPrec(2, 5), // inconsistent value
		P0ArithmeticTwapAccumulator: sdk.ZeroDec(),
		P1ArithmeticTwapAccumulator: sdk.ZeroDec(),
	}

	withGeomAcc := func(r TwapRecord, acc sdk.Dec) TwapRecord {
		r.GeometricTwapAccumulator = acc
		return r
	}

	tests := map[string]struct {
		record              TwapRecord
		isGeometricAccumNil bool
	}{
		"standard": {
			baseParseRecord,
			false,
		},
		"with nil geometric twap accumulator -> set to zero": {
			withGeomAcc(baseParseRecord, sdk.Dec{}),
			true,
		},
		"with non-nil geometric twap accumulator -> not overwritten": {
			withGeomAcc(baseParseRecord, sdk.OneDec()),
			false,
		},
	}
	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			bz, err := proto.Marshal(&tt.record)
			require.NoError(t, err)
			record, err := ParseTwapFromBz(bz)
			require.NoError(t, err)

			if tt.isGeometricAccumNil {
				tt.record.GeometricTwapAccumulator = sdk.ZeroDec()
			}

			require.Equal(t, tt.record, record)
		})
	}
}
