package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

func TestCreateConcentratedLiquidityPoolProposalMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		proposal *types.CreateConcentratedLiquidityPoolProposal
	}{
		{ // empty title
			proposal: &types.CreateConcentratedLiquidityPoolProposal{
				Title:       "",
				Description: "proposal to update migration records",
			},
		},
		{ // empty description
			proposal: &types.CreateConcentratedLiquidityPoolProposal{
				Title:       "title",
				Description: "",
			},
		},
		{ // happy path
			proposal: &types.CreateConcentratedLiquidityPoolProposal{
				Title:              "title",
				Description:        "proposal to update migration records",
				Denom0:             "denom0",
				Denom1:             "denom1",
				TickSpacing:        uint64(1),
				ExponentAtPriceOne: sdk.NewInt(-1),
				SwapFee:            sdk.MustNewDecFromStr("0.01"),
			},
		},
	}

	for _, test := range tests {
		bz, err := proto.Marshal(test.proposal)
		require.NoError(t, err)
		decoded := types.CreateConcentratedLiquidityPoolProposal{}
		err = proto.Unmarshal(bz, &decoded)
		require.NoError(t, err)
		require.Equal(t, *test.proposal, decoded)
	}
}
