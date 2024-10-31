package types_test

import (
	"testing"
	"time"

	proto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"
)

func TestParamsMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		params *types.Params
	}{
		{ // empty denom
			params: &types.Params{
				MintedDenom: "",
			},
		},
		{ // filled
			params: &types.Params{
				MintedDenom: "stake",
			},
		},
	}

	for _, test := range tests {
		bz, err := proto.Marshal(test.params)
		require.NoError(t, err)
		decoded := types.Params{}
		err = proto.Unmarshal(bz, &decoded)
		require.NoError(t, err)
		require.Equal(t, *test.params, decoded)
	}
}

func TestLockableDurationsInfoMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		durations *types.LockableDurationsInfo
	}{
		{ // empty struct
			durations: &types.LockableDurationsInfo{},
		},
		{ // empty lockable durations
			durations: &types.LockableDurationsInfo{
				LockableDurations: []time.Duration(nil),
			},
		},
		{ // filled
			durations: &types.LockableDurationsInfo{
				LockableDurations: []time.Duration{time.Second, time.Hour},
			},
		},
	}

	for _, test := range tests {
		bz, err := proto.Marshal(test.durations)
		require.NoError(t, err)
		decoded := types.LockableDurationsInfo{}
		err = proto.Unmarshal(bz, &decoded)
		require.NoError(t, err)
		require.Equal(t, *test.durations, decoded)
	}
}

func TestDistrInfoMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		info *types.DistrInfo
	}{
		{ // empty records
			info: &types.DistrInfo{
				TotalWeight: osmomath.NewInt(0),
				Records:     []types.DistrRecord(nil),
			},
		},
		{ // one record
			info: &types.DistrInfo{
				TotalWeight: osmomath.NewInt(1),
				Records: []types.DistrRecord{
					{
						GaugeId: 1,
						Weight:  osmomath.NewInt(1),
					},
				},
			},
		},
		{ // two records
			info: &types.DistrInfo{
				TotalWeight: osmomath.NewInt(2),
				Records: []types.DistrRecord{
					{
						GaugeId: 1,
						Weight:  osmomath.NewInt(1),
					},
					{
						GaugeId: 2,
						Weight:  osmomath.NewInt(1),
					},
				},
			},
		},
	}

	for _, test := range tests {
		bz, err := proto.Marshal(test.info)
		require.NoError(t, err)
		decoded := types.DistrInfo{}
		err = proto.Unmarshal(bz, &decoded)
		require.NoError(t, err)
		require.Equal(t, *test.info, decoded)
	}
}

func TestDistrRecordMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		info *types.DistrRecord
	}{
		{ // empty struct
			info: &types.DistrRecord{},
		},
		{ // filled struct
			info: &types.DistrRecord{
				GaugeId: 1,
				Weight:  osmomath.NewInt(1),
			},
		},
	}

	for _, test := range tests {
		bz, err := proto.Marshal(test.info)
		require.NoError(t, err)
		decoded := types.DistrRecord{}
		err = proto.Unmarshal(bz, &decoded)
		require.NoError(t, err)
		require.Equal(t, *test.info, decoded)
	}
}
