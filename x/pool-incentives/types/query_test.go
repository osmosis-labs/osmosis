package types_test

import (
	"testing"
	"time"

	proto "github.com/gogo/protobuf/proto"
	"github.com/osmosis-labs/osmosis/x/pool-incentives/types"
	"github.com/stretchr/testify/require"
)

func TestQueryPotIdsResponseMarshalUnmarshal(t *testing.T) {

	tests := []struct {
		response *types.QueryPotIdsResponse
	}{
		{ // empty struct
			response: &types.QueryPotIdsResponse{},
		},
		{ // length one value
			response: &types.QueryPotIdsResponse{
				PotIdsWithDuration: []*types.QueryPotIdsResponse_PotIdWithDuration{
					{
						PotId:    1,
						Duration: time.Second,
					},
				},
			},
		},
	}

	for _, test := range tests {
		bz, err := proto.Marshal(test.response)
		require.NoError(t, err)
		decoded := types.QueryPotIdsResponse{}
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
						PotId:            1,
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
