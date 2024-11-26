package types_test

import (
	"testing"

	proto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"
)

func TestUpdatePoolIncentivesProposalMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		proposal *types.UpdatePoolIncentivesProposal
	}{
		{ // empty title
			proposal: &types.UpdatePoolIncentivesProposal{
				Title:       "",
				Description: "proposal to update pool incentives",
				Records:     []types.DistrRecord(nil),
			},
		},
		{ // empty description
			proposal: &types.UpdatePoolIncentivesProposal{
				Title:       "title",
				Description: "",
				Records:     []types.DistrRecord(nil),
			},
		},
		{ // empty records
			proposal: &types.UpdatePoolIncentivesProposal{
				Title:       "title",
				Description: "proposal to update pool incentives",
				Records:     []types.DistrRecord(nil),
			},
		},
		{ // one record
			proposal: &types.UpdatePoolIncentivesProposal{
				Title:       "title",
				Description: "proposal to update pool incentives",
				Records: []types.DistrRecord{
					{
						GaugeId: 1,
						Weight:  osmomath.NewInt(1),
					},
				},
			},
		},
		{ // zero-weight record
			proposal: &types.UpdatePoolIncentivesProposal{
				Title:       "title",
				Description: "proposal to update pool incentives",
				Records: []types.DistrRecord{
					{
						GaugeId: 1,
						Weight:  osmomath.NewInt(0),
					},
				},
			},
		},
		{ // two records
			proposal: &types.UpdatePoolIncentivesProposal{
				Title:       "title",
				Description: "proposal to update pool incentives",
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
		bz, err := proto.Marshal(test.proposal)
		require.NoError(t, err)
		decoded := types.UpdatePoolIncentivesProposal{}
		err = proto.Unmarshal(bz, &decoded)
		require.NoError(t, err)
		require.Equal(t, *test.proposal, decoded)
	}
}
