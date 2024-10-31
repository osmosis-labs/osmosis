package types_test

import (
	"testing"

	proto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

func TestCreateConcentratedLiquidityPoolsProposalMarshalUnmarshal(t *testing.T) {
	records := []types.PoolRecord{
		{
			Denom0:       "uion",
			Denom1:       appparams.BaseCoinUnit,
			TickSpacing:  100,
			SpreadFactor: osmomath.MustNewDecFromStr("0.01"),
		},
		{
			Denom0:       "stake",
			Denom1:       appparams.BaseCoinUnit,
			TickSpacing:  1000,
			SpreadFactor: osmomath.MustNewDecFromStr("0.02"),
		},
		{
			Denom0:      "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
			Denom1:      appparams.BaseCoinUnit,
			TickSpacing: 10,

			SpreadFactor: osmomath.MustNewDecFromStr("0.05"),
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
	baseRecord := types.PoolRecord{
		Denom0:       "uion",
		Denom1:       appparams.BaseCoinUnit,
		TickSpacing:  100,
		SpreadFactor: osmomath.MustNewDecFromStr("0.01"),
	}

	withInvalidTickSpacing := func(record types.PoolRecord) types.PoolRecord {
		record.TickSpacing = 0
		return record
	}

	withSameDenom := func(record types.PoolRecord) types.PoolRecord {
		record.Denom1 = record.Denom0
		return record
	}

	withInvalidDenom0 := func(record types.PoolRecord) types.PoolRecord {
		record.Denom0 = "0"
		return record
	}

	withInvalidDenom1 := func(record types.PoolRecord) types.PoolRecord {
		record.Denom1 = "1"
		return record
	}

	withInvalidSpreadFactor := func(record types.PoolRecord) types.PoolRecord {
		record.SpreadFactor = osmomath.MustNewDecFromStr("1.01")
		return record
	}

	tests := []struct {
		name       string
		modifyFunc func(types.PoolRecord) types.PoolRecord
		expectPass bool
	}{
		{
			name:       "proper msg",
			modifyFunc: func(record types.PoolRecord) types.PoolRecord { return record },
			expectPass: true,
		},
		{
			name:       "invalid tick spacing",
			modifyFunc: withInvalidTickSpacing,
			expectPass: false,
		},
		{
			name:       "invalid denom pair",
			modifyFunc: withSameDenom,
			expectPass: false,
		},
		{
			name:       "invalid denom0",
			modifyFunc: withInvalidDenom0,
			expectPass: false,
		},
		{
			name:       "invalid denom1",
			modifyFunc: withInvalidDenom1,
			expectPass: false,
		},
		{
			name:       "invalid spread factor",
			modifyFunc: withInvalidSpreadFactor,
			expectPass: false,
		},
	}

	for _, test := range tests {
		records := []types.PoolRecord{test.modifyFunc(baseRecord)}

		createClPoolsProposal := types.NewCreateConcentratedLiquidityPoolsProposal("title", "description", records)

		if test.expectPass {
			require.NoError(t, createClPoolsProposal.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, createClPoolsProposal.ValidateBasic(), "test: %v", test.name)
		}
	}
}
