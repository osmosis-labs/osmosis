package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
	"github.com/osmosis-labs/osmosis/v3/x/pool-incentives/types"
	"github.com/stretchr/testify/require"
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
				DistrInfo:         nil,
			},
		},
		{ // empty array distribution info
			state: &types.GenesisState{
				Params:            types.DefaultParams(),
				LockableDurations: []time.Duration(nil),
				DistrInfo:         &types.DistrInfo{},
			},
		},
		{ // one record distribution info
			state: &types.GenesisState{
				Params:            types.DefaultParams(),
				LockableDurations: []time.Duration(nil),
				DistrInfo: &types.DistrInfo{
					TotalWeight: sdk.NewInt(1),
					Records: []types.DistrRecord{
						{
							GaugeId: 1,
							Weight:  sdk.NewInt(1),
						},
					},
				},
			},
		},
		{ // empty params
			state: &types.GenesisState{
				Params:            types.Params{},
				LockableDurations: []time.Duration(nil),
				DistrInfo:         nil,
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
