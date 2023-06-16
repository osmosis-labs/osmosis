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

func TestCreateConcentratedLiquidityPoolsProposal_ValidateBasic(t *testing.T) {
	defaultRecords := []types.PoolRecord{
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
		name                string
		records             []types.PoolRecord
		invalidTickSpacing  bool
		invalidSameDenom    bool
		invalidDenom0       bool
		invalidDenom1       bool
		invalidSpreadFactor bool
		expectPass          bool
	}{
		{
			name:       "proper msg",
			expectPass: true,
		},
		{
			name:               "invalid tick spacing",
			invalidTickSpacing: true,
			expectPass:         false,
		},
		{
			name:             "invalid denom pair",
			invalidSameDenom: true,
			expectPass:       false,
		},
		{
			name:          "invalid denom0",
			invalidDenom0: true,
			expectPass:    false,
		},
		{
			name:          "invalid denom1",
			invalidDenom1: true,
			expectPass:    false,
		},
		{
			name:                "invalid spread factor",
			invalidSpreadFactor: true,
			expectPass:          false,
		},
	}

	for _, test := range tests {

		records := defaultRecords

		if test.invalidTickSpacing {
			records[0].TickSpacing = 0
		}

		if test.invalidSameDenom {
			records[0].Denom1 = records[0].Denom0
		}

		if test.invalidDenom0 {
			records[0].Denom0 = "invalidDenom0"
		}

		if test.invalidDenom1 {
			records[0].Denom1 = "invalidDenom1"
		}

		if test.invalidSpreadFactor {
			records[0].SpreadFactor = sdk.MustNewDecFromStr("1.01")
		}

		createClPoolsProposal := types.NewCreateConcentratedLiquidityPoolsProposal("title", "description", records)

		if test.expectPass {
			require.NoError(t, createClPoolsProposal.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, createClPoolsProposal.ValidateBasic(), "test: %v", test.name)
		}
	}
}
