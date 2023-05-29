package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
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
			TickSpacing:        100,
			ExponentAtPriceOne: sdk.NewInt(-5),
			SpreadFactor:       sdk.MustNewDecFromStr("0.02"),
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
