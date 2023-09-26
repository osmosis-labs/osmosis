package types_test

import (
	"testing"

	proto "github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v19/x/incentives/types"
)

var (
	defaultGroups = []types.CreateGroup{
		{PoolIds: []uint64{1, 2, 3}},
		{PoolIds: []uint64{4, 5, 6}},
	}
)

func TestCreateGaugeGroupsProposal_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		proposal *types.CreateGaugeGroupsProposal
	}{
		{ // empty title
			proposal: &types.CreateGaugeGroupsProposal{
				Title:       "",
				Description: "proposal to add gauge groups",
			},
		},
		{ // empty description
			proposal: &types.CreateGaugeGroupsProposal{
				Title:       "title",
				Description: "",
			},
		},
		{ // happy path
			proposal: &types.CreateGaugeGroupsProposal{
				Title:        "title",
				Description:  "proposal to add gauge groups",
				CreateGroups: defaultGroups,
			},
		},
	}

	for _, test := range tests {
		bz, err := proto.Marshal(test.proposal)
		require.NoError(t, err)
		decoded := types.CreateGaugeGroupsProposal{}
		err = proto.Unmarshal(bz, &decoded)
		require.NoError(t, err)
		require.Equal(t, *test.proposal, decoded)
	}
}

func TestCreateGaugeGroupsProposal_ValidateBasic(t *testing.T) {
	notEnoughPoolIdsInGroup := []types.CreateGroup{
		{PoolIds: []uint64{1, 2, 3}},
		{PoolIds: []uint64{4}},
	}

	emptyCreateGroup := []types.CreateGroup{}

	tests := []struct {
		name        string
		createGroup []types.CreateGroup
		expectPass  bool
	}{
		{
			name:        "proper msg",
			createGroup: defaultGroups,
			expectPass:  true,
		},
		{
			name:        "not enough PoolIds in second group",
			createGroup: notEnoughPoolIdsInGroup,
			expectPass:  false,
		},
		{
			name:        "empty create group",
			createGroup: emptyCreateGroup,
			expectPass:  false,
		},
	}

	for _, test := range tests {
		denomPairTakerFeeProposal := types.NewCreateGaugeGroupsProposal("title", "description", test.createGroup)

		if test.expectPass {
			require.NoError(t, denomPairTakerFeeProposal.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, denomPairTakerFeeProposal.ValidateBasic(), "test: %v", test.name)
		}
	}
}
