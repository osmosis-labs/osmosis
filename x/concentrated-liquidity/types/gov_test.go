package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
)

func TestCreateConcentratedLiquidityPoolsProposalMarshalUnmarshal(t *testing.T) {
	records := []types.PoolRecord{
		{
			Denom0:             "uion",
			Denom1:             "uosmo",
			TickSpacing:        100,
			ExponentAtPriceOne: sdk.NewInt(-1),
			SpreadFactor:       sdk.MustNewDecFromStr("0.01"),
		},
		{
			Denom0:             "stake",
			Denom1:             "uosmo",
			TickSpacing:        1000,
			ExponentAtPriceOne: sdk.NewInt(-5),
			SpreadFactor:       sdk.MustNewDecFromStr("0.02"),
		},
		{
			Denom0:             "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
			Denom1:             "uosmo",
			TickSpacing:        10,
			ExponentAtPriceOne: sdk.NewInt(-3),
			SpreadFactor:       sdk.MustNewDecFromStr("0.05"),
		},
	}

	tests := []struct {
		proposal *types.CreateConcentratedLiquidityPoolsProposal
	}{
		{ // empty title
			proposal: &types.CreateConcentratedLiquidityPoolsProposal{
				Title:       "",
				Description: "proposal to update migration records",
			},
		},
		{ // empty description
			proposal: &types.CreateConcentratedLiquidityPoolsProposal{
				Title:       "title",
				Description: "",
			},
		},
		{ // happy path
			proposal: &types.CreateConcentratedLiquidityPoolsProposal{
				Title:       "title",
				Description: "proposal to update migration records",
				PoolRecords: records,
			},
		},
	}

	for _, test := range tests {
		bz, err := proto.Marshal(test.proposal)
		require.NoError(t, err)
		decoded := types.CreateConcentratedLiquidityPoolsProposal{}
		err = proto.Unmarshal(bz, &decoded)
		require.NoError(t, err)
		require.Equal(t, *test.proposal, decoded)
	}
}

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
