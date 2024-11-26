package types_test

import (
	"testing"
	"time"

	proto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"
)

func TestGenesisStateMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		state *types.GenesisState
	}{
		{ // default genesis state
			state: types.DefaultGenesisState(),
		},
		{ // empty lock durations
			state: &types.GenesisState{
				Params:            types.DefaultParams(),
				LockableDurations: []time.Duration(nil),
				DistrInfo: &types.DistrInfo{
					TotalWeight: osmomath.ZeroInt(),
					Records:     nil,
				},
			},
		},
		{ // empty array distribution info
			state: &types.GenesisState{
				Params:            types.DefaultParams(),
				LockableDurations: []time.Duration(nil),
				DistrInfo: &types.DistrInfo{
					TotalWeight: osmomath.ZeroInt(),
					Records:     nil,
				},
			},
		},
		{ // one record distribution info
			state: &types.GenesisState{
				Params:            types.DefaultParams(),
				LockableDurations: []time.Duration(nil),
				DistrInfo: &types.DistrInfo{
					TotalWeight: osmomath.NewInt(1),
					Records: []types.DistrRecord{
						{
							GaugeId: 1,
							Weight:  osmomath.NewInt(1),
						},
					},
				},
			},
		},
		{ // empty params
			state: &types.GenesisState{
				Params:            types.Params{},
				LockableDurations: []time.Duration(nil),
				DistrInfo: &types.DistrInfo{
					TotalWeight: osmomath.ZeroInt(),
					Records:     nil,
				},
			},
		},
	}

	for _, test := range tests {
		bz, err := proto.Marshal(test.state)
		require.NoError(t, err)
		decoded := types.GenesisState{}
		err = proto.Unmarshal(bz, &decoded)
		require.NoError(t, err)
		require.Equal(t, *test.state, decoded)
	}
}
