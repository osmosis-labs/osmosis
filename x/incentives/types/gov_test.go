package types_test

import (
	"testing"

	proto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
)

var (
	defaultGroups = []types.CreateGroup{
		{PoolIds: []uint64{1, 2, 3}},
		{PoolIds: []uint64{4, 5, 6}},
	}
)

func TestCreateGroupsProposal_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		proposal *types.CreateGroupsProposal
	}{
		{ // empty title
			proposal: &types.CreateGroupsProposal{
				Title:       "",
				Description: "proposal to add groups",
			},
		},
		{ // empty description
			proposal: &types.CreateGroupsProposal{
				Title:       "title",
				Description: "",
			},
		},
		{ // happy path
			proposal: &types.CreateGroupsProposal{
				Title:        "title",
				Description:  "proposal to add groups",
				CreateGroups: defaultGroups,
			},
		},
	}

	for _, test := range tests {
		bz, err := proto.Marshal(test.proposal)
		require.NoError(t, err)
		decoded := types.CreateGroupsProposal{}
		err = proto.Unmarshal(bz, &decoded)
		require.NoError(t, err)
		require.Equal(t, *test.proposal, decoded)
	}
}

func TestCreateGroupsProposal_ValidateBasic(t *testing.T) {
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
		denomPairTakerFeeProposal := types.NewCreateGroupsProposal("title", "description", test.createGroup)

		if test.expectPass {
			require.NoError(t, denomPairTakerFeeProposal.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, denomPairTakerFeeProposal.ValidateBasic(), "test: %v", test.name)
		}
	}
}
