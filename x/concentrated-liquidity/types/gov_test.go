package types_test

import (
	"testing"

	proto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v31/x/concentrated-liquidity/types"
)

func TestTickSpacingDecreaseProposalMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		proposal *types.TickSpacingDecreaseProposal
	}{
		{ // empty title
			proposal: &types.TickSpacingDecreaseProposal{
				Title:       "",
				Description: "proposal to update migration records",
			},
		},
		{ // empty description
			proposal: &types.TickSpacingDecreaseProposal{
				Title:       "title",
				Description: "",
			},
		},
		{ // happy path
			proposal: &types.TickSpacingDecreaseProposal{
				Title:       "title",
				Description: "proposal to update migration records",
				PoolIdToTickSpacingRecords: []types.PoolIdToTickSpacingRecord{
					{
						PoolId:         1,
						NewTickSpacing: uint64(1),
					},
				},
			},
		},
	}

	for _, test := range tests {
		bz, err := proto.Marshal(test.proposal)
		require.NoError(t, err)
		decoded := types.TickSpacingDecreaseProposal{}
		err = proto.Unmarshal(bz, &decoded)
		require.NoError(t, err)
		require.Equal(t, *test.proposal, decoded)
	}
}
