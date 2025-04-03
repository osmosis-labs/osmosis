package types_test

import (
	"testing"
	"time"

	proto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"
)

func TestQueryGaugeIdsResponseMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		response *types.QueryGaugeIdsResponse
	}{
		{ // empty struct
			response: &types.QueryGaugeIdsResponse{},
		},
		{ // length one value
			response: &types.QueryGaugeIdsResponse{
				GaugeIdsWithDuration: []*types.QueryGaugeIdsResponse_GaugeIdWithDuration{
					{
						GaugeId:  1,
						Duration: time.Second,
					},
				},
			},
		},
	}

	for _, test := range tests {
		bz, err := proto.Marshal(test.response)
		require.NoError(t, err)
		decoded := types.QueryGaugeIdsResponse{}
		err = proto.Unmarshal(bz, &decoded)
		require.NoError(t, err)
		require.Equal(t, *test.response, decoded)
	}
}

func TestQueryIncentivizedPoolsResponseMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		response *types.QueryIncentivizedPoolsResponse
	}{
		{ // empty struct
			response: &types.QueryIncentivizedPoolsResponse{},
		},
		{ // length one value
			response: &types.QueryIncentivizedPoolsResponse{
				IncentivizedPools: []types.IncentivizedPool{
					{
						PoolId:           1,
						LockableDuration: time.Second,
						GaugeId:          1,
					},
				},
			},
		},
	}

	for _, test := range tests {
		bz, err := proto.Marshal(test.response)
		require.NoError(t, err)
		decoded := types.QueryIncentivizedPoolsResponse{}
		err = proto.Unmarshal(bz, &decoded)
		require.NoError(t, err)
		require.Equal(t, *test.response, decoded)
	}
}
